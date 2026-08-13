package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	cursospb "Prueba-Go/gen/cursos"
	mensajespb "Prueba-Go/gen/mensajes"
	"Prueba-Go/pkg/money"
	"Prueba-Go/services/cursos/internal/repository"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/checkout/session"
	"google.golang.org/grpc/metadata"
)

// precioDe devuelve el importe a cobrar priorizando la columna en centavos.
//
// El respaldo a float64 existe solo para filas anteriores a la migración, donde
// precio_centavos todavía vale 0: math.Round dentro de money.FromFloat recupera
// el valor correcto, a diferencia del int64(precio*100) que se usaba antes y
// truncaba (8.20 → 819 centavos).
func precioDe(centavos int64, legacy float64) money.Amount {
	if centavos > 0 {
		return money.MXNAmount(money.Cents(centavos))
	}
	return money.MXNAmount(money.MustFromFloat(legacy))
}

// Errores de dominio.
var (
	ErrNotFound  = errors.New("no encontrado")
	ErrForbidden = errors.New("sin permisos")
	ErrConflict  = errors.New("ya inscrito")
	// ErrRequierePago: curso de pago y el usuario no tiene suscripción vigente.
	// El frontend lo traduce en "cómpralo suelto o suscríbete".
	ErrRequierePago = errors.New("este curso requiere compra individual o una suscripción activa")
)

// CursosService contiene la lógica de negocio del servicio de cursos.
type CursosService struct {
	repo     repository.CursosRepository
	mensajes mensajespb.MensajesServiceClient // nil-safe: called if set
}

func NewCursosService(repo repository.CursosRepository, mensajes mensajespb.MensajesServiceClient) *CursosService {
	return &CursosService{repo: repo, mensajes: mensajes}
}

func (s *CursosService) ListPublicos(ctx context.Context) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListPublicos(ctx)
	if err != nil {
		slog.Error("ListPublicos repo error", "error", err)
		return nil, err
	}
	slog.Info("ListPublicos success", "count", len(cursos))
	return toProtoSlice(cursos), nil
}

func (s *CursosService) PreviewCurso(ctx context.Context, codigo string) (*cursospb.CursoResponse, error) {
	c, err := s.repo.FindByCodigo(ctx, codigo)
	if err != nil {
		return nil, ErrNotFound
	}
	return c.ToProto(), nil
}

func (s *CursosService) GetCursoPublico(ctx context.Context, cursoID string) (*cursospb.CursoResponse, error) {
	c, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if !c.IsPublic {
		return nil, ErrForbidden
	}
	return c.ToProto(), nil
}

func (s *CursosService) ListMisCapacitaciones(ctx context.Context, userID string) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		slog.Error("ListMisCapacitaciones repo error", "error", err, "userId", userID)
		return nil, err
	}

	slog.Info("ListMisCapacitaciones success", "count", len(cursos), "userId", userID)
	return toProtoSlice(cursos), nil
}

func (s *CursosService) GetCurso(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	enrolled, err := s.repo.IsEnrolled(ctx, userID, cursoID)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if enrolled || c.IsPublic {
		return c.ToProto(), nil
	}
	// Una suscripción viva abre todo el catálogo: es justo lo que se vende.
	if conSuscripcion, _, _, _, errS := s.repo.AccesoPorSuscripcion(ctx, userID); errS == nil && conSuscripcion {
		return c.ToProto(), nil
	}
	return nil, ErrForbidden
}

// Inscribirse da acceso sin cobrar. Vale para cursos gratuitos y para
// suscriptores con el plan al corriente.
//
// Se inscribe de verdad (fila en `inscripciones`) en lugar de solo dejar ver el
// contenido: así el progreso, los exámenes y la constancia DC-3 funcionan igual
// que en una compra individual. Si la suscripción caduca, el usuario conserva
// su historial pero deja de poder abrir cursos nuevos.
func (s *CursosService) Inscribirse(ctx context.Context, userID, cursoID string) error {
	enrolled, _ := s.repo.IsEnrolled(ctx, userID, cursoID)
	if enrolled {
		return ErrConflict
	}
	curso, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return err
	}

	esDePago := precioDe(curso.PrecioCentavos, curso.Precio).IsPositive()
	if esDePago {
		conSuscripcion, _, _, estado, errS := s.repo.AccesoPorSuscripcion(ctx, userID)
		if errS != nil {
			return errS
		}
		if !conSuscripcion {
			return ErrRequierePago
		}
		slog.Info("inscripción por suscripción",
			"user_id", userID, "curso_id", cursoID, "estado_suscripcion", estado)
	}
	return s.repo.Inscribirse(ctx, userID, cursoID)
}

func (s *CursosService) UnirseConCodigo(ctx context.Context, userID, codigo string) error {
	_, err := s.repo.UnirseConCodigo(ctx, userID, codigo)
	return err
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (s *CursosService) InstructorListCapacitaciones(ctx context.Context, instructorID string) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	for _, c := range cursos {
		if c.CodigoAcceso == "" {
			if updated, err2 := s.repo.ResetCodigo(ctx, c.ID); err2 == nil && updated != nil {
				c.CodigoAcceso = updated.CodigoAcceso
			}
		}
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) InstructorCreate(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorUpdate(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	// Verificar que el instructor es dueño del curso.
	existing, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != req.UserId {
		return nil, ErrForbidden
	}
	c, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorDelete(ctx context.Context, cursoID, userID string) error {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, cursoID)
}

func (s *CursosService) InstructorTogglePublic(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return nil, ErrForbidden
	}
	c, err := s.repo.TogglePublic(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorResetCodigo(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return nil, ErrForbidden
	}
	c, err := s.repo.ResetCodigo(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorListEstudiantes(ctx context.Context, instructorID, cursoID string) ([]*cursospb.EstudianteInfo, error) {
	rows, err := s.repo.ListEstudiantes(ctx, instructorID, cursoID)
	if err != nil {
		return nil, err
	}
	result := make([]*cursospb.EstudianteInfo, 0, len(rows))
	for _, r := range rows {
		result = append(result, &cursospb.EstudianteInfo{
			UserId: r.ID, Name: r.Name, Email: r.Email,
			AssignedAt: r.AssignedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

func (s *CursosService) InstructorAsignar(ctx context.Context, instructorID, userID, cursoID string) error {
	return s.repo.InstructorAsignar(ctx, instructorID, userID, cursoID)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

func (s *CursosService) AdminList(ctx context.Context) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cursos {
		if c.CodigoAcceso == "" {
			if updated, err2 := s.repo.ResetCodigo(ctx, c.ID); err2 == nil && updated != nil {
				c.CodigoAcceso = updated.CodigoAcceso
			}
		}
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) AdminCreate(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) AdminUpdate(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) AdminDelete(ctx context.Context, cursoID string) error {
	return s.repo.Delete(ctx, cursoID)
}

func (s *CursosService) AdminResetCodigo(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.ResetCodigo(ctx, req.CursoId)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) AdminListAsignaciones(ctx context.Context) ([]*cursospb.AsignacionInfo, error) {
	asigs, err := s.repo.ListAsignaciones(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*cursospb.AsignacionInfo, 0, len(asigs))
	for _, a := range asigs {
		result = append(result, a.ToProto())
	}
	return result, nil
}

func (s *CursosService) AdminAsignar(ctx context.Context, userID, cursoID string) error {
	return s.repo.AdminAsignar(ctx, userID, cursoID)
}

func (s *CursosService) AdminDesAsignar(ctx context.Context, asignacionID string) error {
	return s.repo.DesAsignar(ctx, asignacionID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toProtoSlice(cursos []*repository.Curso) []*cursospb.CursoResponse {
	result := make([]*cursospb.CursoResponse, 0, len(cursos))
	for _, c := range cursos {
		result = append(result, c.ToProto())
	}
	return result
}

// ── Licencias ─────────────────────────────────────────────────────────────────

func (s *CursosService) CreateLicencia(ctx context.Context, req *cursospb.CreateLicenciaRequest) (*cursospb.Licencia, error) {
	lic, err := s.repo.CreateLicencia(ctx, req)
	if err != nil {
		return nil, err
	}
	// Auto-create cohort group in mensajes-service
	if s.mensajes != nil {
		_, _ = s.mensajes.CreateGroupForLicencia(ctx, &mensajespb.CreateGroupForLicenciaRequest{
			LicenciaId: lic.ID,
			Nombre:     lic.Nombre + " — Grupo de Cohorte",
			AdminId:    req.InstructorId,
		})
	}
	return lic.ToProto(), nil
}

func (s *CursosService) UpdateLicencia(ctx context.Context, req *cursospb.UpdateLicenciaRequest) (*cursospb.Licencia, error) {
	lic, err := s.repo.UpdateLicencia(ctx, req)
	if err != nil {
		return nil, err
	}
	return lic.ToProto(), nil
}

func (s *CursosService) DeleteLicencia(ctx context.Context, id string) error {
	return s.repo.DeleteLicencia(ctx, id)
}

func (s *CursosService) ListLicencias(ctx context.Context, cursoID string) ([]*cursospb.Licencia, error) {
	lics, err := s.repo.ListLicencias(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	res := make([]*cursospb.Licencia, len(lics))
	for i, l := range lics {
		res[i] = l.ToProto()
	}
	return res, nil
}

func (s *CursosService) UnirseConLicencia(ctx context.Context, userID, capID, codigoAcceso string) error {
	lic, err := s.repo.FindLicenciaByCodigo(ctx, codigoAcceso)
	if err != nil {
		return errors.New("código de acceso inválido")
	}
	if lic.CapacitacionID != capID {
		return errors.New("el código no corresponde a esta capacitación")
	}
	if lic.CapacidadMaxima > 0 && lic.Usadas >= lic.CapacidadMaxima {
		return errors.New("la licencia ha alcanzado su capacidad máxima")
	}
	err = s.repo.InscribirseConLicencia(ctx, userID, capID, lic.ID)
	if err == nil {
		_ = s.repo.IncrementarUsoLicencia(ctx, lic.ID)
		// Auto-enrol in cohort group
		if s.mensajes != nil {
			_, _ = s.mensajes.EnrollInLicenciaGroup(ctx, &mensajespb.EnrollInLicenciaGroupRequest{
				LicenciaId: lic.ID,
				UserId:     userID,
			})
		}
	}
	return err
}

// WebhookEnroll inscribe al comprador tras un pago B2C confirmado y devuelve
// los datos del curso para que el Gateway pueda redirigir y enviar el acuse.
func (s *CursosService) WebhookEnroll(ctx context.Context, req *cursospb.WebhookEnrollRequest) (*cursospb.EnrollResponse, error) {
	err := s.repo.InscribirseConLicencia(ctx, req.UserId, req.CapacitacionId, req.LicenciaId)
	if err == nil && req.LicenciaId != "" {
		_ = s.repo.IncrementarUsoLicencia(ctx, req.LicenciaId)
	}
	resp := &cursospb.EnrollResponse{CapacitacionId: req.CapacitacionId}
	if err == nil {
		if curso, errC := s.repo.FindByID(ctx, req.CapacitacionId); errC == nil {
			resp.CapacitacionTitulo = curso.Title
			resp.CapacitacionType = curso.Type
			if curso.InstructorID != nil {
				resp.InstructorId = *curso.InstructorID
			}
		}
	}

	return resp, err
}

func (s *CursosService) CreateCheckoutSession(ctx context.Context, req *cursospb.CheckoutSessionRequest) (*cursospb.CheckoutSessionResponse, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	var productName string
	var amount int64
	var clientRef string

	if req.CursoId != "" {
		// B2C Course Purchase
		curso, err := s.repo.FindByID(ctx, req.CursoId)
		if err != nil {
			return nil, err
		}
		// Misma guarda que en el carrito. La inscripción es única por usuario y
		// curso, así que cobrar de nuevo no entrega nada: solo genera un
		// reembolso manual.
		if req.UserId != "" {
			inscrito, errI := s.repo.IsEnrolled(ctx, req.UserId, req.CursoId)
			if errI != nil {
				return nil, errI
			}
			if inscrito {
				return nil, fmt.Errorf("%w: %s", ErrYaInscrito, curso.Title)
			}
		}
		importe := precioDe(curso.PrecioCentavos, curso.Precio)
		if !importe.IsPositive() {
			return nil, errors.New("el curso no tiene precio")
		}
		productName = curso.Title
		amount = importe.StripeAmount()
		clientRef = "curso||" + req.UserId + "||" + curso.ID
	} else {
		// B2B License Purchase
		lic, err := s.repo.FindLicenciaByID(ctx, req.LicenciaId)
		if err != nil {
			return nil, err
		}
		if lic.CapacidadMaxima > 0 && lic.Usadas >= lic.CapacidadMaxima {
			return nil, errors.New("licencia agotada")
		}
		importe := precioDe(lic.PrecioCentavos, lic.Precio)
		productName = lic.Nombre
		amount = importe.StripeAmount()
		clientRef = req.UserId + "||" + lic.CapacitacionID + "||" + lic.ID
	}

	// Crear sesión
	metodos, opcionesPago := metodosDePago(amount)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes:   metodos,
		PaymentMethodOptions: opcionesPago,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("mxn"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(productName),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(req.SuccessUrl),
		CancelURL:         stripe.String(req.CancelUrl),
		ClientReferenceID: stripe.String(clientRef),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled: stripe.Bool(true),
		},
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
		TaxIDCollection: &stripe.CheckoutSessionTaxIDCollectionParams{
			Enabled: stripe.Bool(true),
		},
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-user-email"); len(vals) > 0 && vals[0] != "" {
			params.CustomerEmail = stripe.String(vals[0])
		}
	}

	if stripe.Key == "" {
		return nil, errors.New("STRIPE_SECRET_KEY no está configurada en el servidor (cursos-service)")
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error de Stripe al crear sesión: %v", err)
		return nil, fmt.Errorf("error al conectar con el procesador de pagos: %v", err)
	}
	return &cursospb.CheckoutSessionResponse{Url: sess.URL}, nil
}

func (s *CursosService) WebhookComprarLicencia(ctx context.Context, req *cursospb.WebhookComprarLicenciaRequest) (*cursospb.EmptyResponse, error) {
	err := s.repo.AsignarCompradorLicencia(ctx, req.LicenciaId, req.UserId)
	return &cursospb.EmptyResponse{}, err
}

func (s *CursosService) GetLicenciaPublica(ctx context.Context, req *cursospb.LicenciaIDRequest) (*cursospb.LicenciaPublicaResponse, error) {
	lic, err := s.repo.FindLicenciaByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	curso, err := s.repo.FindByID(ctx, lic.CapacitacionID)
	if err != nil {
		return nil, err
	}
	return &cursospb.LicenciaPublicaResponse{
		Id:                    lic.ID,
		Nombre:                lic.Nombre,
		Precio:                lic.Precio,
		CapacidadMaxima:       lic.CapacidadMaxima,
		CapacitacionId:        curso.ID,
		CapacitacionTitulo:    curso.Title,
		CapacitacionThumbnail: curso.ThumbnailURL,
	}, nil
}

func (s *CursosService) ListLicenciasCompradas(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListLicenciasResponse, error) {
	lics, err := s.repo.ListLicenciasCompradas(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	var res []*cursospb.Licencia
	for _, l := range lics {
		res = append(res, l.ToProto())
	}
	return &cursospb.ListLicenciasResponse{Licencias: res}, nil
}

func (s *CursosService) CreateCheckoutSessionB2BDirect(ctx context.Context, req *cursospb.CreateCheckoutSessionB2BDirectRequest) (*cursospb.CheckoutSessionResponse, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	curso, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return nil, err
	}
	unitario := precioDe(curso.PrecioCentavos, curso.Precio)
	if !unitario.IsPositive() {
		return nil, errors.New("el curso no tiene precio")
	}
	if req.Cantidad < 1 {
		return nil, errors.New("la cantidad debe ser al menos 1")
	}

	productName := "Licencias Corporativas: " + curso.Title
	total, err := unitario.Mul(int64(req.Cantidad))
	if err != nil {
		return nil, fmt.Errorf("total inválido: %w", err)
	}
	amount := total.StripeAmount()
	clientRef := "b2b_direct||" + req.UserId + "||" + curso.ID + "||" + fmt.Sprintf("%d", req.Cantidad)

	// Crear sesión
	metodos, opcionesPago := metodosDePago(amount)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes:   metodos,
		PaymentMethodOptions: opcionesPago,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("mxn"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(productName),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(req.SuccessUrl),
		CancelURL:         stripe.String(req.CancelUrl),
		ClientReferenceID: stripe.String(clientRef),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled: stripe.Bool(true),
		},
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
		TaxIDCollection: &stripe.CheckoutSessionTaxIDCollectionParams{
			Enabled: stripe.Bool(true),
		},
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-user-email"); len(vals) > 0 && vals[0] != "" {
			params.CustomerEmail = stripe.String(vals[0])
		}
	}

	if stripe.Key == "" {
		return nil, errors.New("STRIPE_SECRET_KEY no está configurada en el servidor (cursos-service)")
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error de Stripe al crear sesión B2B directa: %v", err)
		return nil, fmt.Errorf("error al conectar con el procesador de pagos: %v", err)
	}
	return &cursospb.CheckoutSessionResponse{Url: sess.URL}, nil
}

// WebhookComprarB2BDirect crea la licencia corporativa tras el pago y devuelve
// los códigos generados para que el Gateway se los mande por correo al comprador.
func (s *CursosService) WebhookComprarB2BDirect(ctx context.Context, req *cursospb.WebhookComprarB2BDirectRequest) (*cursospb.ComprarB2BDirectResponse, error) {
	// Verificar que el curso existe
	curso, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return nil, err
	}

	// El total se recalcula en centavos y solo al final se convierte a float
	// para la columna NUMERIC legacy.
	unitario := precioDe(curso.PrecioCentavos, curso.Precio)
	totalLic, err := unitario.Mul(int64(req.Cantidad))
	if err != nil {
		return nil, fmt.Errorf("total inválido: %w", err)
	}
	precioTotal := totalLic.Float()

	// Crear licencia
	lic, err := s.repo.CreateLicenciaB2BDirect(ctx, req, precioTotal)
	if err != nil {
		return nil, err
	}

	resp := &cursospb.ComprarB2BDirectResponse{
		LicenciaId:         lic.ID,
		CapacitacionId:     curso.ID,
		CapacitacionTitulo: curso.Title,
		CapacitacionType:   curso.Type,
		Lugares:            req.Cantidad,
		Total:              precioTotal,
	}
	if lic.CodigoAcceso != nil {
		resp.CodigoAcceso = *lic.CodigoAcceso
	}

	return resp, nil
}

// AsignarAccesosLicencia reparte los accesos de una licencia entre los correos
// que capturó el comprador. Solo devuelve los datos: el envío del correo lo hace
// el Gateway, que es quien tiene el cliente de Resend.
func (s *CursosService) AsignarAccesosLicencia(ctx context.Context, req *cursospb.AsignarAccesosLicenciaRequest) (*cursospb.AsignarAccesosLicenciaResponse, error) {
	if len(req.Participantes) == 0 {
		return nil, errors.New("no se recibieron participantes")
	}

	lic, err := s.repo.FindLicenciaByID(ctx, req.LicenciaId)
	if err != nil {
		return nil, err
	}
	// Autorización: solo el comprador reparte los accesos que pagó.
	if lic.CompradorID == nil || *lic.CompradorID != req.CompradorId {
		return nil, ErrForbidden
	}

	curso, err := s.repo.FindByID(ctx, lic.CapacitacionID)
	if err != nil {
		return nil, err
	}

	// Todos los participantes comparten el código de acceso de la licencia.
	codigoCompartido := ""
	if lic.CodigoAcceso != nil {
		codigoCompartido = *lic.CodigoAcceso
	}

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

	invs, err := s.repo.AsignarAccesos(ctx, lic.ID, codigoCompartido, participantes)
	if err != nil {
		return nil, err
	}

	resp := &cursospb.AsignarAccesosLicenciaResponse{}
	for _, inv := range invs {
		resp.Accesos = append(resp.Accesos, &cursospb.AccesoParticipante{
			Nombre:             inv.Nombre,
			Email:              inv.Email,
			Codigo:             inv.Codigo,
			CapacitacionId:     curso.ID,
			CapacitacionTitulo: curso.Title,
			CapacitacionType:   curso.Type,
		})
	}
	return resp, nil
}

// NotificarCursoCompletado decide si hay que avisar al representante de la
// licencia para que tramite las constancias DC-3.
//
// Antes esto se disparaba al terminar una videollamada. Al desaparecer ese
// flujo, el punto natural es que un participante complete el contenido.
func (s *CursosService) NotificarCursoCompletado(ctx context.Context, req *cursospb.CursoCompletadoRequest) (*cursospb.CursoCompletadoResponse, error) {
	resp := &cursospb.CursoCompletadoResponse{}

	curso, err := s.repo.FindByID(ctx, req.CapacitacionId)
	if err != nil {
		return nil, err
	}
	if !curso.DC3Enabled {
		return resp, nil
	}

	lic, err := s.repo.FindLicenciaDeInscripcion(ctx, req.UserId, req.CapacitacionId)
	if err != nil {
		return nil, err
	}
	// Compra individual: el propio participante tramita su constancia desde la
	// plataforma, no hay representante a quien escribir.
	if lic == nil || lic.CompradorID == nil || *lic.CompradorID == "" {
		return resp, nil
	}

	primeraVez, err := s.repo.RegistrarAvisoDC3(ctx, lic.ID, req.CapacitacionId)
	if err != nil {
		return nil, err
	}
	if !primeraVez {
		return resp, nil
	}

	resp.Avisar = true
	resp.RepresentanteId = *lic.CompradorID
	resp.CapacitacionTitulo = curso.Title
	resp.DuracionMinutos = curso.Duration
	return resp, nil
}

// ListInvitacionesLicencia devuelve el estado de entrega de los accesos.
func (s *CursosService) ListInvitacionesLicencia(ctx context.Context, req *cursospb.LicenciaIDRequest) (*cursospb.ListInvitacionesLicenciaResponse, error) {
	invs, err := s.repo.ListInvitacionesLicencia(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	resp := &cursospb.ListInvitacionesLicenciaResponse{}
	for _, i := range invs {
		resp.Invitaciones = append(resp.Invitaciones, i.ToProto())
	}
	return resp, nil
}

func (s *CursosService) GetAdminDashboardStats(ctx context.Context) (*cursospb.AdminDashboardStatsResponse, error) {
	return s.repo.GetAdminDashboardStats(ctx)
}

// ErrYaInscrito: se intenta comprar un curso que el usuario ya tiene.
var ErrYaInscrito = errors.New("ya tienes acceso a este curso")

// normalizarCarrito deja el carrito en un estado que se puede cobrar sin dañar
// a nadie: colapsa renglones repetidos y descarta lo que el usuario ya posee.
//
// Es una validación de SERVIDOR a propósito. El carrito vive en el localStorage
// del navegador, así que su contenido es una propuesta del cliente, no un
// hecho: cualquier arreglo que solo esté en el front se lo salta una pestaña
// vieja, un carrito guardado de antes o una petición a mano.
//
// Las dos reglas nacen de un cobro real de 2×$400 por un mismo curso, donde el
// segundo renglón no compró nada porque la inscripción es única por usuario.
func (s *CursosService) normalizarCarrito(
	ctx context.Context, userID string, items []*cursospb.CartItem,
) ([]*cursospb.CartItem, error) {
	// La clave incluye el tipo: la misma capacitación puede comprarse a la vez
	// como inscripción propia y como licencias para el equipo.
	type clave struct{ cursoID, tipo string }
	posicion := make(map[clave]int, len(items))
	salida := make([]*cursospb.CartItem, 0, len(items))

	for _, item := range items {
		if item == nil || item.CursoId == "" {
			continue
		}

		// Solo aplica a la inscripción propia. Comprar licencias corporativas de
		// un curso que tú ya llevas es legítimo: los lugares son para tu equipo.
		if item.Type == "b2c" && userID != "" {
			inscrito, err := s.repo.IsEnrolled(ctx, userID, item.CursoId)
			if err != nil {
				return nil, err
			}
			if inscrito {
				curso, errC := s.repo.FindByID(ctx, item.CursoId)
				nombre := item.CursoId
				if errC == nil {
					nombre = curso.Title
				}
				return nil, fmt.Errorf("%w: %s", ErrYaInscrito, nombre)
			}
		}

		k := clave{item.CursoId, item.Type}
		if i, visto := posicion[k]; visto {
			// Individual: una inscripción por persona, la cantidad no significa
			// nada. Corporativo: los lugares sí se acumulan.
			if item.Type == "b2b_direct" {
				salida[i].Cantidad += item.Cantidad
			}
			continue
		}
		posicion[k] = len(salida)
		salida = append(salida, item)
	}

	if len(salida) == 0 {
		return nil, errors.New("el carrito está vacío")
	}
	if len(salida) != len(items) {
		slog.Warn("carrito con renglones repetidos, se consolidaron",
			"user_id", userID, "recibidos", len(items), "cobrados", len(salida))
	}
	return salida, nil
}

func (s *CursosService) CreateCheckoutSessionCart(ctx context.Context, req *cursospb.CheckoutCartRequest) (*cursospb.CheckoutSessionResponse, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	if len(req.Items) == 0 {
		return nil, errors.New("el carrito está vacío")
	}

	items, err := s.normalizarCarrito(ctx, req.UserId, req.Items)
	if err != nil {
		return nil, err
	}
	req.Items = items

	var lineItems []*stripe.CheckoutSessionLineItemParams
	metadataMap := make(map[string]string)
	// Los renglones alimentan la orden: el precio se congela al momento de
	// comprar, así que una subida posterior no altera la compra histórica.
	var renglones []repository.OrdenItem
	total := money.MXNAmount(0)

	for i, item := range req.Items {
		curso, err := s.repo.FindByID(ctx, item.CursoId)
		if err != nil {
			return nil, fmt.Errorf("curso no encontrado: %s", item.CursoId)
		}
		unitario := precioDe(curso.PrecioCentavos, curso.Precio)
		if !unitario.IsPositive() {
			return nil, fmt.Errorf("el curso %s no tiene precio", curso.Title)
		}

		var productName string
		var amount int64
		var quantity int64
		var itemMeta string

		if item.Type == "b2c" {
			productName = curso.Title
			amount = unitario.StripeAmount()
			quantity = 1
			itemMeta = fmt.Sprintf("b2c||%s", curso.ID)
		} else if item.Type == "b2b_direct" {
			productName = "Licencia Corporativa: " + curso.Title
			amount = unitario.StripeAmount()
			quantity = int64(item.Cantidad)
			itemMeta = fmt.Sprintf("b2b_direct||%s||%d", curso.ID, item.Cantidad)
		} else {
			return nil, fmt.Errorf("tipo de ítem no válido: %s", item.Type)
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("mxn"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(productName),
				},
				UnitAmount: stripe.Int64(amount),
			},
			Quantity: stripe.Int64(quantity),
		})

		// Stripe metadata limit is 50 keys
		metadataMap[fmt.Sprintf("item_%d", i)] = itemMeta

		subtotal, err := unitario.Mul(quantity)
		if err != nil {
			return nil, fmt.Errorf("subtotal inválido en %s: %w", curso.Title, err)
		}
		if total, err = total.Add(subtotal); err != nil {
			return nil, fmt.Errorf("total del carrito inválido: %w", err)
		}
		renglones = append(renglones, repository.OrdenItem{
			CapacitacionID:         curso.ID,
			Tipo:                   item.Type,
			Titulo:                 curso.Title,
			Cantidad:               int32(quantity),
			PrecioUnitarioCentavos: unitario.StripeAmount(),
			SubtotalCentavos:       subtotal.StripeAmount(),
		})
	}

	// La orden se crea ANTES de hablar con Stripe. Si el usuario abandona el
	// pago, queda el registro del intento; si hace doble clic, cae en la misma
	// orden y se reutiliza su sesión en lugar de abrir una segunda.
	orden, reutilizada, err := s.repo.CrearOAbrirOrden(
		ctx, req.UserId, total.StripeAmount(), string(money.MXN), renglones)
	if err != nil {
		return nil, err
	}
	if reutilizada && orden.StripeSessionID != nil {
		sess, err := session.Get(*orden.StripeSessionID, nil)
		if err == nil && sess.URL != "" {
			slog.Info("checkout: reutilizando sesión existente",
				"orden_id", orden.ID, "session_id", sess.ID)
			return &cursospb.CheckoutSessionResponse{Url: sess.URL, OrdenId: orden.ID}, nil
		}
		// Si Stripe ya no la reconoce, se sigue y se crea una nueva.
		slog.Warn("checkout: la sesión guardada ya no sirve, se crea otra",
			"orden_id", orden.ID, "error", err)
	}

	clientRef := "cart||" + req.UserId

	// El tope de OXXO aplica al total del carrito, no a cada renglón: es lo que
	// Stripe cobra en una sola operación.
	metodos, opcionesPago := metodosDePago(total.StripeAmount())
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes:   metodos,
		PaymentMethodOptions: opcionesPago,
		LineItems:            lineItems,
		Mode:                 stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:           stripe.String(req.SuccessUrl),
		CancelURL:            stripe.String(req.CancelUrl),
		ClientReferenceID:    stripe.String(clientRef),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled: stripe.Bool(true),
		},
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
		TaxIDCollection: &stripe.CheckoutSessionTaxIDCollectionParams{
			Enabled: stripe.Bool(true),
		},
	}

	for k, v := range metadataMap {
		params.AddMetadata(k, v)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-user-email"); len(vals) > 0 && vals[0] != "" {
			params.CustomerEmail = stripe.String(vals[0])
		}
	}

	// La clave de idempotencia deriva de la orden y su intento, no de un UUID
	// por petición: si esta llamada se reintenta (timeout de red, reintento del
	// cliente), Stripe devuelve la MISMA sesión en lugar de crear otra.
	params.IdempotencyKey = stripe.String(fmt.Sprintf("%s-intento-%d", orden.ID, orden.Intento))
	params.AddMetadata("orden_id", orden.ID)

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error de Stripe al crear sesión de carrito: %v", err)
		// La orden queda 'pendiente': el siguiente intento la reabre en lugar
		// de generar una huérfana.
		return nil, fmt.Errorf("error al conectar con Stripe: %v", err)
	}

	if err := s.repo.GuardarSesionStripe(ctx, orden.ID, sess.ID); err != nil {
		// No se aborta: el cobro ya puede ocurrir. Pero sin esta asociación el
		// webhook no sabrá qué orden cerrar, así que se registra como error.
		slog.Error("no se pudo asociar la sesión de Stripe con la orden",
			"orden_id", orden.ID, "session_id", sess.ID, "error", err)
	}

	return &cursospb.CheckoutSessionResponse{Url: sess.URL, OrdenId: orden.ID}, nil
}

// ── Órdenes y webhooks ───────────────────────────────────────────────────────

// RegistrarEventoStripe deduplica los webhooks. Stripe entrega al-menos-una-vez.
func (s *CursosService) RegistrarEventoStripe(ctx context.Context, req *cursospb.EventoStripeRequest) (*cursospb.EventoStripeResponse, error) {
	primeraVez, err := s.repo.RegistrarEventoStripe(ctx, req.EventId, req.Tipo)
	if err != nil {
		return nil, err
	}
	return &cursospb.EventoStripeResponse{PrimeraVez: primeraVez}, nil
}

// ActualizarEstadoOrden aplica la transición tras el resultado del cobro.
func (s *CursosService) ActualizarEstadoOrden(ctx context.Context, req *cursospb.ActualizarEstadoOrdenRequest) (*cursospb.EmptyResponse, error) {
	err := s.repo.ActualizarEstadoOrden(ctx, req.StripeSessionId, req.Estado, req.MotivoFallo, req.StripePaymentIntent)
	return &cursospb.EmptyResponse{}, err
}
