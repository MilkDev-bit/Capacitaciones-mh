package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/pkg/money"
	"Prueba-Go/services/cursos/internal/repository"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/checkout/session"
)

// ─────────────────────────────────────────────────────────────────────────────
// Suscripciones
//
// Dos modalidades conviven:
//   individual → una persona, acceso a todo el catálogo
//   asientos   → una empresa paga N lugares para su equipo
//
// El precio y el ciclo de vida los define Stripe. Aquí solo se refleja lo que
// el PSP reporta por webhook: inferir estados localmente desincroniza el
// sistema en cuanto un reintento de cobro tiene éxito.
// ─────────────────────────────────────────────────────────────────────────────

// ListPlanes devuelve el catálogo de planes vigentes.
func (s *CursosService) ListPlanes(ctx context.Context) (*cursospb.ListPlanesResponse, error) {
	planes, err := s.repo.ListPlanesActivos(ctx)
	if err != nil {
		return nil, err
	}
	resp := &cursospb.ListPlanesResponse{}
	for _, p := range planes {
		resp.Planes = append(resp.Planes, &cursospb.Plan{
			Id:             p.ID,
			Codigo:         p.Codigo,
			Nombre:         p.Nombre,
			Descripcion:    p.Descripcion,
			Modalidad:      p.Modalidad,
			Intervalo:      p.Intervalo,
			PrecioCentavos: p.PrecioCentavos,
			Moneda:         p.Moneda,
			DiasPrueba:     p.DiasPrueba,
			Activo:         p.Activo,
		})
	}
	return resp, nil
}

// GetMiSuscripcion devuelve la suscripción viva del usuario.
// Una respuesta con id vacío significa "no tiene", no es un error.
func (s *CursosService) GetMiSuscripcion(ctx context.Context, userID string) (*cursospb.SuscripcionResponse, error) {
	sus, err := s.repo.FindSuscripcionDeUsuario(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sus == nil {
		return &cursospb.SuscripcionResponse{}, nil
	}

	ocupados, err := s.repo.ContarAsientosOcupados(ctx, sus.ID)
	if err != nil {
		slog.Error("no se pudieron contar los asientos", "suscripcion_id", sus.ID, "error", err)
	}
	return aProtoSuscripcion(sus, ocupados), nil
}

func aProtoSuscripcion(sus *repository.Suscripcion, ocupados int32) *cursospb.SuscripcionResponse {
	r := &cursospb.SuscripcionResponse{
		Id:                 sus.ID,
		PlanId:             sus.PlanID,
		Estado:             sus.Estado,
		Asientos:           sus.Asientos,
		AsientosOcupados:   ocupados,
		CancelarAlTerminar: sus.CancelarAlTerminar,
		AccesoVigente:      repository.EstadoDaAcceso(sus.Estado),
	}
	if sus.PlanNombre != nil {
		r.PlanNombre = *sus.PlanNombre
	}
	if sus.PlanModalidad != nil {
		r.Modalidad = *sus.PlanModalidad
	}
	if sus.PlanIntervalo != nil {
		r.Intervalo = *sus.PlanIntervalo
	}
	if sus.PlanPrecioCentavos != nil {
		r.PrecioCentavos = *sus.PlanPrecioCentavos
	}
	if sus.PeriodoInicio != nil {
		r.PeriodoInicio = sus.PeriodoInicio.Format(time.RFC3339)
	}
	if sus.PeriodoFin != nil {
		r.PeriodoFin = sus.PeriodoFin.Format(time.RFC3339)
	}
	if sus.PruebaFin != nil {
		r.PruebaFin = sus.PruebaFin.Format(time.RFC3339)
	}
	if sus.StripeCustomerID != nil {
		r.StripeCustomerId = *sus.StripeCustomerID
	}
	return r
}

// CrearCheckoutSuscripcion abre una sesión de Stripe en modo suscripción.
func (s *CursosService) CrearCheckoutSuscripcion(ctx context.Context, req *cursospb.CheckoutSuscripcionRequest) (*cursospb.CheckoutSessionResponse, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		return nil, errors.New("STRIPE_SECRET_KEY no está configurada")
	}

	plan, err := s.repo.FindPlanPorCodigo(ctx, req.PlanCodigo)
	if err != nil {
		return nil, err
	}
	if !plan.Activo {
		return nil, errors.New("ese plan ya no está disponible")
	}
	// El price de Stripe es obligatorio: el precio lo define el PSP, no nosotros.
	// El precio_centavos local solo sirve para pintarlo sin llamar a su API.
	if plan.StripePriceID == nil || *plan.StripePriceID == "" {
		return nil, fmt.Errorf("el plan %s no tiene stripe_price_id configurado", plan.Codigo)
	}

	asientos := req.Asientos
	if plan.Modalidad == "individual" {
		asientos = 1 // una membresía individual es siempre un asiento
	}
	if asientos < 1 {
		return nil, errors.New("el número de asientos debe ser al menos 1")
	}

	// Una suscripción viva bloquea el alta de otra: el índice único parcial de
	// la BD lo impediría igual, pero aquí se devuelve un mensaje entendible.
	if actual, err := s.repo.FindSuscripcionDeUsuario(ctx, req.UserId); err == nil && actual != nil {
		return nil, repository.ErrYaTieneSuscripcion
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			Price:    stripe.String(*plan.StripePriceID),
			Quantity: stripe.Int64(int64(asientos)),
		}},
		SuccessURL:        stripe.String(req.SuccessUrl),
		CancelURL:         stripe.String(req.CancelUrl),
		ClientReferenceID: stripe.String("suscripcion||" + req.UserId + "||" + plan.Codigo),
		// La metadata viaja a la Subscription: el webhook la necesita para saber
		// a qué usuario y plan pertenece el alta.
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id":     req.UserId,
				"plan_codigo": plan.Codigo,
			},
		},
	}
	if plan.DiasPrueba > 0 {
		params.SubscriptionData.TrialPeriodDays = stripe.Int64(int64(plan.DiasPrueba))
	}
	params.AddMetadata("user_id", req.UserId)
	params.AddMetadata("plan_codigo", plan.Codigo)

	// Clave derivada del usuario y el plan: un doble clic en "Suscribirme" no
	// abre dos sesiones. Cambia con la hora para permitir reintentar más tarde
	// si la primera sesión expiró.
	params.IdempotencyKey = stripe.String(fmt.Sprintf(
		"susc-%s-%s-%d", req.UserId, plan.Codigo, time.Now().Truncate(time.Hour).Unix()))

	sess, err := session.New(params)
	if err != nil {
		slog.Error("Stripe: fallo al crear sesión de suscripción", "error", err, "plan", plan.Codigo)
		return nil, fmt.Errorf("error al conectar con el procesador de pagos: %w", err)
	}
	return &cursospb.CheckoutSessionResponse{Url: sess.URL}, nil
}

// SincronizarSuscripcion aplica el estado que reporta Stripe.
func (s *CursosService) SincronizarSuscripcion(ctx context.Context, req *cursospb.SincronizarSuscripcionRequest) (*cursospb.EmptyResponse, error) {
	sus := &repository.Suscripcion{
		UserID:               req.UserId,
		Estado:               req.Estado,
		Asientos:             req.Asientos,
		CancelarAlTerminar:   req.CancelarAlTerminar,
		StripeSubscriptionID: &req.StripeSubscriptionId,
	}
	if req.StripeCustomerId != "" {
		sus.StripeCustomerID = &req.StripeCustomerId
	}
	if sus.Asientos < 1 {
		sus.Asientos = 1
	}
	sus.PeriodoInicio = parseTiempo(req.PeriodoInicio)
	sus.PeriodoFin = parseTiempo(req.PeriodoFin)
	sus.PruebaFin = parseTiempo(req.PruebaFin)

	if err := s.repo.UpsertSuscripcion(ctx, sus, req.PlanCodigo); err != nil {
		return nil, err
	}
	slog.Info("suscripción sincronizada",
		"stripe_id", req.StripeSubscriptionId, "estado", req.Estado, "asientos", sus.Asientos)
	return &cursospb.EmptyResponse{}, nil
}

// parseTiempo convierte un RFC3339 vacío en nil en vez de en la fecha cero.
func parseTiempo(v string) *time.Time {
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}

// RegistrarFacturaSuscripcion guarda el resultado de un cobro recurrente.
func (s *CursosService) RegistrarFacturaSuscripcion(ctx context.Context, req *cursospb.FacturaSuscripcionRequest) (*cursospb.EmptyResponse, error) {
	sus, err := s.repo.FindSuscripcionPorStripeID(ctx, req.StripeSubscriptionId)
	if err != nil {
		// Una factura de una suscripción que no conocemos no debe tumbar el
		// webhook: se registra y se sigue.
		slog.Warn("factura de una suscripción desconocida",
			"stripe_subscription_id", req.StripeSubscriptionId, "invoice", req.StripeInvoiceId)
		return &cursospb.EmptyResponse{}, nil
	}

	moneda := req.Moneda
	if moneda == "" {
		moneda = string(money.MXN)
	}
	err = s.repo.RegistrarFactura(ctx, sus.ID, req.StripeInvoiceId, req.Estado,
		req.TotalCentavos, strings.ToUpper(moneda), req.IntentoCobro, req.UrlPdf,
		parseTiempo(req.PeriodoInicio), parseTiempo(req.PeriodoFin))
	return &cursospb.EmptyResponse{}, err
}

// TieneAccesoPorSuscripcion resuelve el acceso del usuario al catálogo.
func (s *CursosService) TieneAccesoPorSuscripcion(ctx context.Context, userID string) (*cursospb.AccesoSuscripcionResponse, error) {
	tiene, origen, susID, estado, err := s.repo.AccesoPorSuscripcion(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &cursospb.AccesoSuscripcionResponse{
		TieneAcceso:   tiene,
		Origen:        origen,
		SuscripcionId: susID,
		Estado:        estado,
	}, nil
}

// ── Asientos ─────────────────────────────────────────────────────────────────

// AsignarAsientos reparte lugares de una suscripción corporativa.
func (s *CursosService) AsignarAsientos(ctx context.Context, req *cursospb.AsignarAsientosRequest) (*cursospb.ListAsientosResponse, error) {
	participantes := make([]repository.Participante, 0, len(req.Participantes))
	for _, p := range req.Participantes {
		email := strings.ToLower(strings.TrimSpace(p.Email))
		if email == "" {
			continue
		}
		participantes = append(participantes, repository.Participante{
			Nombre: strings.TrimSpace(p.Nombre),
			Email:  email,
		})
	}
	if len(participantes) == 0 {
		return nil, errors.New("ningún correo válido en la lista")
	}

	asientos, err := s.repo.AsignarAsientos(ctx, req.SuscripcionId, req.TitularId, participantes)
	if err != nil {
		return nil, err
	}
	return s.armarRespuestaAsientos(ctx, req.SuscripcionId, asientos)
}

// ListAsientos devuelve el reparto actual.
func (s *CursosService) ListAsientos(ctx context.Context, req *cursospb.SuscripcionIDRequest) (*cursospb.ListAsientosResponse, error) {
	asientos, err := s.repo.ListAsientos(ctx, req.SuscripcionId, req.TitularId)
	if err != nil {
		return nil, err
	}
	return s.armarRespuestaAsientos(ctx, req.SuscripcionId, asientos)
}

// RevocarAsiento libera un lugar sin borrar el historial.
func (s *CursosService) RevocarAsiento(ctx context.Context, req *cursospb.RevocarAsientoRequest) (*cursospb.EmptyResponse, error) {
	err := s.repo.RevocarAsiento(ctx, req.SuscripcionId, req.TitularId, req.Email)
	return &cursospb.EmptyResponse{}, err
}

func (s *CursosService) armarRespuestaAsientos(ctx context.Context, suscripcionID string, asientos []*repository.Asiento) (*cursospb.ListAsientosResponse, error) {
	sus, err := s.repo.FindSuscripcionPorID(ctx, suscripcionID)
	if err != nil {
		return nil, err
	}

	resp := &cursospb.ListAsientosResponse{Total: sus.Asientos}
	for _, a := range asientos {
		item := &cursospb.Asiento{
			Id:         a.ID,
			Email:      a.Email,
			Estado:     a.Estado,
			InvitadoAt: a.InvitadoAt.Format(time.RFC3339),
		}
		if a.UserID != nil {
			item.UserId = *a.UserID
		}
		resp.Asientos = append(resp.Asientos, item)
	}
	resp.Ocupados = int32(len(asientos))
	resp.Libres = sus.Asientos - resp.Ocupados
	if resp.Libres < 0 {
		resp.Libres = 0
	}
	return resp, nil
}
