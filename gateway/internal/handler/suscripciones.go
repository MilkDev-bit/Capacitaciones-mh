package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/middleware"
	cursospb "Prueba-Go/gen/cursos"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v86"
	billingportal "github.com/stripe/stripe-go/v86/billingportal/session"
)

// ─────────────────────────────────────────────────────────────────────────────
// Suscripciones — capa HTTP
//
// La gestión (cambiar tarjeta, cancelar, ver recibos) se delega al Billing
// Portal de Stripe en lugar de reimplementarla: mantiene la integración en
// SAQ A porque ningún dato de tarjeta toca nuestras páginas.
// ─────────────────────────────────────────────────────────────────────────────

// GET /api/planes — público, alimenta la página de precios.
func (h *CursosHandler) ListPlanes(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListPlanes(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Planes)
}

// GET /api/mi-suscripcion
func (h *CursosHandler) GetMiSuscripcion(ctx *gin.Context) {
	resp, err := h.c.Cursos.GetMiSuscripcion(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	// stripe_customer_id se omite a propósito: es un identificador del PSP que
	// el navegador no necesita y que no conviene exponer.
	ctx.JSON(http.StatusOK, gin.H{
		"id":                   resp.Id,
		"plan_id":              resp.PlanId,
		"plan_nombre":          resp.PlanNombre,
		"modalidad":            resp.Modalidad,
		"intervalo":            resp.Intervalo,
		"precio_centavos":      resp.PrecioCentavos,
		"estado":               resp.Estado,
		"asientos":             resp.Asientos,
		"asientos_ocupados":    resp.AsientosOcupados,
		"periodo_inicio":       resp.PeriodoInicio,
		"periodo_fin":          resp.PeriodoFin,
		"prueba_fin":           resp.PruebaFin,
		"cancelar_al_terminar": resp.CancelarAlTerminar,
		"acceso_vigente":       resp.AccesoVigente,
	})
}

// POST /api/suscripcion/checkout
func (h *CursosHandler) CrearCheckoutSuscripcion(ctx *gin.Context) {
	var body struct {
		PlanCodigo string `json:"plan_codigo" binding:"required"`
		Asientos   int32  `json:"asientos"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	base := strings.TrimRight(h.cfg.AppURL, "/")
	if body.SuccessURL == "" {
		body.SuccessURL = base + "/usuario/suscripcion?alta=1"
	}
	if body.CancelURL == "" {
		body.CancelURL = base + "/planes"
	}

	resp, err := h.c.Cursos.CrearCheckoutSuscripcion(genMetadata(ctx), &cursospb.CheckoutSuscripcionRequest{
		UserId:     ctx.GetString(middleware.CtxUserID),
		PlanCodigo: body.PlanCodigo,
		Asientos:   body.Asientos,
		SuccessUrl: body.SuccessURL,
		CancelUrl:  body.CancelURL,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.Url})
}

// POST /api/suscripcion/portal
//
// Devuelve una URL efímera al portal de Stripe. No se cachea: cada sesión
// caduca y va firmada para un cliente concreto.
func (h *CursosHandler) PortalFacturacion(ctx *gin.Context) {
	sus, err := h.c.Cursos.GetMiSuscripcion(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	if sus.Id == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no tienes una suscripción activa"})
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Stripe no está configurado"})
		return
	}

	// Sin customer en Stripe la suscripción no llegó a crearse del todo.
	if sus.StripeCustomerId == "" {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "tu suscripción aún se está activando, intenta en un momento",
		})
		return
	}

	portal, err := billingportal.New(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(sus.StripeCustomerId),
		ReturnURL: stripe.String(strings.TrimRight(h.cfg.AppURL, "/") + "/usuario/suscripcion"),
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "no se pudo abrir el portal de facturación"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": portal.URL})
}

// ── Asientos ─────────────────────────────────────────────────────────────────

// POST /api/suscripcion/:id/asientos
func (h *CursosHandler) AsignarAsientos(ctx *gin.Context) {
	var body struct {
		Participantes []struct {
			Nombre string `json:"nombre"`
			Email  string `json:"email" binding:"required,email"`
		} `json:"participantes" binding:"required,min=1,max=500,dive"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "revisa los correos capturados: " + err.Error()})
		return
	}

	// Se deduplica aquí para no gastar asientos por un copy/paste repetido.
	vistos := map[string]bool{}
	req := &cursospb.AsignarAsientosRequest{
		SuscripcionId: ctx.Param("id"),
		TitularId:     ctx.GetString(middleware.CtxUserID),
	}
	for _, p := range body.Participantes {
		email := strings.ToLower(strings.TrimSpace(p.Email))
		if email == "" || vistos[email] {
			continue
		}
		vistos[email] = true
		req.Participantes = append(req.Participantes, &cursospb.ParticipanteAsiento{
			Nombre: strings.TrimSpace(p.Nombre),
			Email:  email,
		})
	}
	if len(req.Participantes) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no hay correos válidos"})
		return
	}

	resp, err := h.c.Cursos.AsignarAsientos(genMetadata(ctx), req)
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GET /api/suscripcion/:id/asientos
func (h *CursosHandler) ListAsientos(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListAsientos(ctx.Request.Context(), &cursospb.SuscripcionIDRequest{
		SuscripcionId: ctx.Param("id"),
		TitularId:     ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DELETE /api/suscripcion/:id/asientos
func (h *CursosHandler) RevocarAsiento(ctx *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Cursos.RevocarAsiento(ctx.Request.Context(), &cursospb.RevocarAsientoRequest{
		SuscripcionId: ctx.Param("id"),
		TitularId:     ctx.GetString(middleware.CtxUserID),
		Email:         body.Email,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ── Webhooks de suscripción ──────────────────────────────────────────────────

// sincronizarDesdeStripe traduce el objeto Subscription de Stripe al dominio.
//
// Se toma SIEMPRE el estado que reporta Stripe en lugar de deducirlo: si el
// cobro de un 'past_due' se recupera, Stripe manda 'active' y nosotros lo
// aplicamos tal cual. Deducirlo localmente dejaría al usuario fuera.
func (h *CursosHandler) sincronizarDesdeStripe(ctx context.Context, sub *stripe.Subscription) {
	if sub == nil || sub.ID == "" {
		return
	}

	inicio, fin := periodoDeSuscripcion(sub)
	req := &cursospb.SincronizarSuscripcionRequest{
		StripeSubscriptionId: sub.ID,
		Estado:               estadoSuscripcionDesdeStripe(string(sub.Status)),
		CancelarAlTerminar:   sub.CancelAtPeriodEnd,
		PeriodoInicio:        deUnix(inicio),
		PeriodoFin:           deUnix(fin),
		PruebaFin:            deUnix(sub.TrialEnd),
	}
	if sub.Customer != nil {
		req.StripeCustomerId = sub.Customer.ID
	}
	// La metadata solo viene completa en el alta; en las actualizaciones puede
	// faltar y el upsert la ignora.
	if sub.Metadata != nil {
		req.UserId = sub.Metadata["user_id"]
		req.PlanCodigo = sub.Metadata["plan_codigo"]
	}
	// La cantidad de asientos vive en el primer item de la suscripción.
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		req.Asientos = int32(sub.Items.Data[0].Quantity)
	}

	if _, err := h.c.Cursos.SincronizarSuscripcion(ctx, req); err != nil {
		slog.Error("no se pudo sincronizar la suscripción", "error", err, "stripe_id", sub.ID)
	}
}

// periodoDeSuscripcion devuelve el periodo de facturación vigente.
//
// Hasta la versión 2025-03-31.basil de la API estos campos vivían en la
// suscripción (`sub.CurrentPeriodStart/End`). Stripe los movió al ítem porque
// una suscripción puede facturar ítems con periodos distintos.
//
// Se toma el del primer ítem: este dominio vende un plan con N asientos, o sea
// un único ítem — el mismo del que ya se leía `Quantity` más abajo. Si algún día
// se venden planes combinados, aquí habrá que decidir cuál manda.
func periodoDeSuscripcion(sub *stripe.Subscription) (inicio, fin int64) {
	if sub == nil || sub.Items == nil || len(sub.Items.Data) == 0 {
		return 0, 0
	}
	item := sub.Items.Data[0]
	return item.CurrentPeriodStart, item.CurrentPeriodEnd
}

// suscripcionDeFactura extrae el ID de la suscripción que originó la factura.
//
// Hasta 2025-03-31.basil esto era `inv.Subscription`. Ahora la factura describe
// su origen en `Parent`, que no siempre es una suscripción (una factura suelta
// no lo es), de ahí el recorrido con guardas en cada salto.
func suscripcionDeFactura(inv *stripe.Invoice) string {
	if inv == nil || inv.Parent == nil || inv.Parent.SubscriptionDetails == nil {
		return ""
	}
	if sub := inv.Parent.SubscriptionDetails.Subscription; sub != nil {
		return sub.ID
	}
	return ""
}

// registrarFactura guarda el resultado de un cobro recurrente.
func (h *CursosHandler) registrarFactura(ctx context.Context, inv *stripe.Invoice, estado string) {
	subID := suscripcionDeFactura(inv)
	if inv == nil || inv.ID == "" || subID == "" {
		return
	}
	moneda := strings.ToUpper(string(inv.Currency))
	if _, err := h.c.Cursos.RegistrarFacturaSuscripcion(ctx, &cursospb.FacturaSuscripcionRequest{
		StripeSubscriptionId: subID,
		StripeInvoiceId:      inv.ID,
		Estado:               estado,
		TotalCentavos:        inv.Total,
		Moneda:               moneda,
		IntentoCobro:         int32(inv.AttemptCount),
		UrlPdf:               inv.InvoicePDF,
		PeriodoInicio:        deUnix(inv.PeriodStart),
		PeriodoFin:           deUnix(inv.PeriodEnd),
	}); err != nil {
		slog.Error("no se pudo registrar la factura", "error", err, "invoice", inv.ID)
	}
}

// estadoSuscripcionDesdeStripe traduce el vocabulario del PSP al del dominio.
func estadoSuscripcionDesdeStripe(s string) string {
	switch s {
	case "trialing":
		return "en_prueba"
	case "active":
		return "activa"
	case "past_due":
		return "vencida"
	case "unpaid":
		return "impagada"
	case "canceled", "incomplete_expired":
		return "cancelada"
	default:
		return "incompleta"
	}
}

// deUnix convierte un timestamp de Stripe a RFC3339. Devuelve "" en el cero
// para que el servicio lo guarde como NULL y no como 1970.
func deUnix(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}
