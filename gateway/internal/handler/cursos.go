package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/middleware"
	cursospb "Prueba-Go/gen/cursos"
	leccionespb "Prueba-Go/gen/lecciones"
	"Prueba-Go/pkg/mailer"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	stripeSession "github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// toASCII elimina caracteres no-ASCII para que sean válidos como valores de
// cabecera gRPC (solo se permiten caracteres ASCII imprimibles).
func toASCII(s string) string {
	r := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", "ü", "u",
		"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ñ", "N", "Ü", "U",
	).Replace(s)
	var b strings.Builder
	for _, c := range r {
		if c >= 0x20 && c <= 0x7E {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// CursosHandler traduce peticiones HTTP ↔ RPC del cursos service.
//
// Es también el punto donde se envía el correo transaccional de compras: el
// Gateway es el único componente que conoce la sesión de Stripe completa y el
// perfil del comprador, así que centralizar aquí el envío evita duplicar
// credenciales de Resend en cada microservicio.
type CursosHandler struct {
	c    *clients.Clients
	cfg  *config.Config
	mail *mailer.Client
}

func NewCursosHandler(c *clients.Clients, cfg *config.Config, mail *mailer.Client) *CursosHandler {
	return &CursosHandler{c: c, cfg: cfg, mail: mail}
}

// func genMetadata(ctx *gin.Context) context.Context
func genMetadata(ctx *gin.Context) context.Context {
	md := metadata.Pairs(
		"x-user-name", toASCII(ctx.GetString(middleware.CtxUserName)),
		"x-user-email", toASCII(ctx.GetString(middleware.CtxUserEmail)),
	)
	return metadata.NewOutgoingContext(ctx.Request.Context(), md)
}

func cursoToJSON(resp *cursospb.CursoResponse) gin.H {
	if resp == nil {
		return gin.H{}
	}
	return gin.H{
		"id":                    resp.Id,
		"title":                 resp.Title,
		"description":           resp.Description,
		"type":                  resp.Type,
		"file_path":             resp.FilePath,
		"content":               resp.Content,
		"instructor_id":         resp.InstructorId,
		"is_public":             resp.IsPublic,
		"codigo_acceso":         resp.CodigoAcceso,
		"welcome_message":       resp.WelcomeMessage,
		"thumbnail_url":         resp.ThumbnailUrl,
		"color":                 resp.Color,
		"created_at":            resp.CreatedAt,
		"precio":                resp.Precio,
		"scheduled_at":          resp.ScheduledAt,
		"duration":              resp.Duration,
		"total_lecciones":       resp.TotalLecciones,
		"lecciones_completadas": resp.LeccionesCompletadas,
		"dc3_enabled":           resp.Dc3Enabled,
	}
}

func cursosToJSON(list []*cursospb.CursoResponse) []gin.H {
	out := make([]gin.H, len(list))
	for i, c := range list {
		out[i] = cursoToJSON(c)
	}
	return out
}

// GET /api/cursos-publicos
func (h *CursosHandler) ListCursosPublicos(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListCursosPublicos(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	// Enriquecer con el total de lecciones (sin progreso de usuario)
	for _, curso := range resp.Cursos {
		leccionesResp, err := h.c.Lecciones.InstructorListLecciones(ctx.Request.Context(), &leccionespb.CursoRequest{
			CursoId: curso.Id,
		})
		if err == nil && leccionesResp != nil {
			curso.TotalLecciones = int32(len(leccionesResp.Lecciones))
		}
	}

	ctx.JSON(http.StatusOK, cursosToJSON(resp.Cursos))
}

// GET /api/cursos-publicos/:id
func (h *CursosHandler) GetCursoPublico(ctx *gin.Context) {
	resp, err := h.c.Cursos.GetCursoPublico(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, cursoToJSON(resp))
}

// GET /api/preview-curso/:codigo
func (h *CursosHandler) PreviewCurso(ctx *gin.Context) {
	resp, err := h.c.Cursos.PreviewCurso(ctx.Request.Context(), &cursospb.CodigoRequest{
		Codigo: ctx.Param("codigo"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GET /api/mis-capacitaciones
func (h *CursosHandler) ListMisCapacitaciones(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxUserID)
	resp, err := h.c.Cursos.ListMisCapacitaciones(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: userID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	// Enriquecer cada curso con el progreso real del servicio de lecciones
	for _, curso := range resp.Cursos {
		leccionesResp, err := h.c.Lecciones.GetLeccionesConProgreso(ctx.Request.Context(), &leccionespb.CursoUserRequest{
			CursoId: curso.Id,
			UserId:  userID,
		})
		if err == nil && leccionesResp != nil {
			curso.TotalLecciones = int32(len(leccionesResp.Lecciones))
			completadas := int32(0)
			for _, leccion := range leccionesResp.Lecciones {
				if leccion.Completada {
					completadas++
				}
			}
			curso.LeccionesCompletadas = completadas
		}
	}

	ctx.JSON(http.StatusOK, cursosToJSON(resp.Cursos))
}

// GET /api/capacitaciones/:id
func (h *CursosHandler) GetCurso(ctx *gin.Context) {
	resp, err := h.c.Cursos.GetCurso(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, cursoToJSON(resp))
}

// POST /api/cursos/:id/inscripciones
func (h *CursosHandler) Inscribirse(ctx *gin.Context) {
	md := metadata.Pairs(
		"x-user-name", toASCII(ctx.GetString(middleware.CtxUserName)),
		"x-user-email", toASCII(ctx.GetString(middleware.CtxUserEmail)),
	)
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	_, err := h.c.Cursos.Inscribirse(grpcCtx, &cursospb.InscribirseRequest{
		UserId:  ctx.GetString(middleware.CtxUserID),
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	h.avisarInscripcion(ctx.GetString(middleware.CtxUserID), ctx.Param("id"))
	ctx.JSON(http.StatusCreated, gin.H{"message": "inscripción exitosa"})
}

// avisarInscripcion deja constancia del alta en la campana del alumno.
//
// Las tres vías de inscripción sin pago (directa, por código y por licencia)
// devuelven EmptyResponse, así que no hay título de curso que poner: el enlace
// al detalle es lo útil. Cuando cursos-service devuelva el título, basta con
// pasarlo aquí.
func (h *CursosHandler) avisarInscripcion(userID, cursoID string) {
	enlace := "/usuario/capacitaciones"
	if cursoID != "" {
		enlace += "/" + cursoID
	}
	notificar(h.c, aviso{
		UserID:  userID,
		Tipo:    TipoInscripcion,
		Titulo:  "Tienes una capacitación nueva",
		Mensaje: "Ya puedes empezarla desde tus capacitaciones.",
		Enlace:  enlace,
		Ventana: ventanaCompra,
	})
}

// POST /api/inscripciones  (unirse con código)
func (h *CursosHandler) UnirseConCodigo(ctx *gin.Context) {
	var body struct {
		Codigo string `json:"codigo" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	md := metadata.Pairs(
		"x-user-name", toASCII(ctx.GetString(middleware.CtxUserName)),
		"x-user-email", toASCII(ctx.GetString(middleware.CtxUserEmail)),
	)
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	_, err := h.c.Cursos.UnirseConCodigo(grpcCtx, &cursospb.UnirseRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
		Codigo: body.Codigo,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	// Sin curso_id: el código lo resuelve el servicio y no lo devuelve, así que
	// el enlace va al listado.
	h.avisarInscripcion(ctx.GetString(middleware.CtxUserID), "")
	ctx.JSON(http.StatusCreated, gin.H{"message": "inscripción exitosa"})
}

// POST /api/inscripciones-licencia
func (h *CursosHandler) UnirseConLicencia(ctx *gin.Context) {
	var req struct {
		CapacitacionID string `json:"capacitacion_id" binding:"required"`
		CodigoAcceso   string `json:"codigo_acceso" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Cursos.UnirseConLicencia(genMetadata(ctx), &cursospb.UnirseConLicenciaRequest{
		UserId:         ctx.GetString(middleware.CtxUserID),
		CapacitacionId: req.CapacitacionID,
		CodigoAcceso:   req.CodigoAcceso,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	h.avisarInscripcion(ctx.GetString(middleware.CtxUserID), req.CapacitacionID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Inscrito con licencia correctamente"})
}

// GET /api/usuario/licencias-compradas
func (h *CursosHandler) ListLicenciasCompradas(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListLicenciasCompradas(genMetadata(ctx), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Licencias)
}

// POST /api/checkout-session
func (h *CursosHandler) CreateCheckoutSession(ctx *gin.Context) {
	var req struct {
		CursoID    string `json:"curso_id"`
		LicenciaID string `json:"licencia_id"`
		SuccessUrl string `json:"success_url" binding:"required"`
		CancelUrl  string `json:"cancel_url" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Cursos.CreateCheckoutSession(genMetadata(ctx), &cursospb.CheckoutSessionRequest{
		UserId:     ctx.GetString(middleware.CtxUserID),
		CursoId:    req.CursoID,
		LicenciaId: req.LicenciaID,
		SuccessUrl: req.SuccessUrl,
		CancelUrl:  req.CancelUrl,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.Url})
}

// POST /api/checkout-session-b2b-direct
func (h *CursosHandler) CreateCheckoutSessionB2BDirect(ctx *gin.Context) {
	var req struct {
		CursoID    string `json:"curso_id" binding:"required"`
		Cantidad   int32  `json:"cantidad" binding:"required"`
		SuccessUrl string `json:"success_url" binding:"required"`
		CancelUrl  string `json:"cancel_url" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Cursos.CreateCheckoutSessionB2BDirect(genMetadata(ctx), &cursospb.CreateCheckoutSessionB2BDirectRequest{
		UserId:     ctx.GetString(middleware.CtxUserID),
		CursoId:    req.CursoID,
		Cantidad:   req.Cantidad,
		SuccessUrl: req.SuccessUrl,
		CancelUrl:  req.CancelUrl,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.Url})
}

// POST /api/checkout-session-cart
func (h *CursosHandler) CreateCheckoutSessionCart(ctx *gin.Context) {
	var req struct {
		Items []struct {
			CursoID  string `json:"curso_id"`
			Cantidad int32  `json:"cantidad"`
			Type     string `json:"type"`
		} `json:"items" binding:"required"`
		SuccessUrl string `json:"success_url" binding:"required"`
		CancelUrl  string `json:"cancel_url" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var protoItems []*cursospb.CartItem
	for _, it := range req.Items {
		protoItems = append(protoItems, &cursospb.CartItem{
			CursoId:  it.CursoID,
			Cantidad: it.Cantidad,
			Type:     it.Type,
		})
	}

	resp, err := h.c.Cursos.CreateCheckoutSessionCart(genMetadata(ctx), &cursospb.CheckoutCartRequest{
		UserId:     ctx.GetString(middleware.CtxUserID),
		Items:      protoItems,
		SuccessUrl: req.SuccessUrl,
		CancelUrl:  req.CancelUrl,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.Url})
}

// POST /api/webhooks/stripe
//
// Camino autoritativo: Stripe confirma el pago aunque el usuario cierre la
// pestaña al volver del checkout.
func (h *CursosHandler) StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	signatureHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook signature verification failed"})
		return
	}

	// Deduplicación por event.ID ANTES de procesar. Stripe entrega
	// al-menos-una-vez y reintenta ante cualquier respuesta que no sea 2xx: sin
	// esto, un reenvío repetiría el correo de confirmación.
	dedupe, errDedupe := h.c.Cursos.RegistrarEventoStripe(c.Request.Context(), &cursospb.EventoStripeRequest{
		EventId: event.ID,
		Tipo:    string(event.Type),
	})
	if errDedupe != nil {
		// Si no se puede deduplicar, es preferible fallar y dejar que Stripe
		// reintente antes que procesar a ciegas y arriesgar un doble efecto.
		slog.Error("webhook: fallo al registrar el evento", "error", errDedupe, "event_id", event.ID)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no se pudo registrar el evento"})
		return
	}
	if !dedupe.PrimeraVez {
		slog.Info("webhook: evento ya procesado, se descarta", "event_id", event.ID, "tipo", event.Type)
		c.Status(http.StatusOK)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		h.cerrarOrden(c.Request.Context(), sess.ID, "pagada", "", intentIDDe(&sess))

		// Se responde 200 a Stripe pase lo que pase con el correo: un fallo de
		// Resend no debe provocar reintentos del webhook ni altas duplicadas.
		res := h.procesarSesion(c.Request.Context(), &sess, "")

		h.cerrarOrden(c.Request.Context(), sess.ID, "cumplida", "", "")

		nombre, email := datosComprador(&sess)
		h.notificarCompra(&sess, res, nombre, email)

	case "checkout.session.expired":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}
		// El usuario abandonó el pago. La orden queda 'fallida' para que la
		// conciliación distinga el abandono de un cobro perdido.
		h.cerrarOrden(c.Request.Context(), sess.ID, "fallida", "sesión expirada sin pago", "")

	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}
		h.sincronizarDesdeStripe(c.Request.Context(), &sub)

	case "invoice.paid":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}
		h.registrarFactura(c.Request.Context(), &inv, "pagada")

	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}
		// El acceso NO se corta aquí: Stripe moverá la suscripción a past_due
		// y mandará customer.subscription.updated. El dunning sigue corriendo.
		h.registrarFactura(c.Request.Context(), &inv, "fallida")
		slog.Warn("cobro recurrente fallido",
			"invoice", inv.ID, "intento", inv.AttemptCount, "total", inv.Total)

	case "charge.refunded":
		var ch stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &ch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}
		slog.Info("webhook: reembolso recibido", "charge_id", ch.ID, "payment_intent", ch.PaymentIntent)
	}

	c.Status(http.StatusOK)
}

// cerrarOrden aplica una transición de estado sin abortar el webhook si falla:
// perder el registro de la orden es malo, pero peor sería que Stripe reintentara
// y volviéramos a otorgar los accesos.
func (h *CursosHandler) cerrarOrden(ctx context.Context, sessionID, estado, motivo, paymentIntent string) {
	if sessionID == "" {
		return
	}
	if _, err := h.c.Cursos.ActualizarEstadoOrden(ctx, &cursospb.ActualizarEstadoOrdenRequest{
		StripeSessionId:     sessionID,
		Estado:              estado,
		MotivoFallo:         motivo,
		StripePaymentIntent: paymentIntent,
	}); err != nil {
		slog.Error("no se pudo actualizar el estado de la orden",
			"session_id", sessionID, "estado", estado, "error", err)
	}
}

// intentIDDe extrae el PaymentIntent de la sesión, que es lo que se necesita
// para conciliar contra los cobros de Stripe.
func intentIDDe(sess *stripe.CheckoutSession) string {
	if sess.PaymentIntent != nil {
		return sess.PaymentIntent.ID
	}
	return ""
}

// datosComprador extrae nombre y correo de la sesión de Stripe.
// En el webhook no hay JWT, así que es la única fuente disponible.
func datosComprador(sess *stripe.CheckoutSession) (nombre, email string) {
	if sess.CustomerDetails != nil {
		nombre = sess.CustomerDetails.Name
		email = sess.CustomerDetails.Email
	}
	if email == "" {
		email = sess.CustomerEmail
	}
	return nombre, email
}

// POST /api/verify-checkout-session  ← llamado desde la pantalla de éxito
//
// Respaldo del webhook: permite completar el alta en cuanto el usuario vuelve
// de Stripe, sin esperar a que llegue el evento. Devuelve el resumen de la
// compra para pintar la confirmación y saber a dónde redirigir.
func (h *CursosHandler) VerifyCheckoutSession(ctx *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Stripe no configurado"})
		return
	}

	sess, err := stripeSession.Get(req.SessionID, nil)
	if err != nil {
		slog.Error("VerifyCheckoutSession: error obteniendo sesión de Stripe", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Sesión inválida"})
		return
	}

	if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "El pago no ha sido completado"})
		return
	}

	// El user_id sale del JWT, no de la referencia de Stripe: así una sesión
	// ajena no puede usarse para inscribir a otra cuenta.
	userID := ctx.GetString(middleware.CtxUserID)
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), metadata.Pairs(
		"x-user-email", toASCII(ctx.GetString(middleware.CtxUserEmail)),
		"x-user-name", toASCII(ctx.GetString(middleware.CtxUserName)),
	))

	h.cerrarOrden(ctx.Request.Context(), sess.ID, "pagada", "", intentIDDe(sess))
	res := h.procesarSesion(grpcCtx, sess, userID)
	h.cerrarOrden(ctx.Request.Context(), sess.ID, "cumplida", "", "")

	nombre := ctx.GetString(middleware.CtxUserName)
	email := ctx.GetString(middleware.CtxUserEmail)
	if email == "" {
		nombre, email = datosComprador(sess)
	}
	h.notificarCompra(sess, res, nombre, email)

	ctx.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"items":    res.Items,
		"total":    res.Total,
		"redirect": res.Redirect,
		"etiqueta": res.Etiqueta,
	})
}

// GET /api/licencias/:id/invoice  — devuelve la URL del PDF de la factura de Stripe
func (h *CursosHandler) GetLicenciaInvoicePDF(ctx *gin.Context) {
	licenciaID := ctx.Param("id")
	userID := ctx.GetString(middleware.CtxUserID)

	// Obtener la licencia para saber el stripe_session_id guardado
	resp, err := h.c.Cursos.ListLicenciasCompradas(genMetadata(ctx), &cursospb.UserRequest{UserId: userID})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	var sessionID string
	for _, l := range resp.Licencias {
		if l.Id == licenciaID {
			sessionID = l.StripeProductId // guardamos el session_id aquí
			break
		}
	}

	if sessionID == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No hay factura disponible para esta licencia"})
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	cleanSessionID := sessionID
	if idx := strings.Index(cleanSessionID, "_item_"); idx != -1 {
		cleanSessionID = cleanSessionID[:idx]
	}

	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("invoice")
	s, err := stripeSession.Get(cleanSessionID, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener datos de Stripe", "details": err.Error()})
		return
	}

	if s.Invoice == nil || s.Invoice.InvoicePDF == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "La factura no está disponible en Stripe"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"invoice_pdf": s.Invoice.InvoicePDF,
		"invoice_url": s.Invoice.HostedInvoiceURL,
	})
}

// ── Instructor ────────────────────────────────────────────────────────────────

// GET /api/instructor/capacitaciones
func (h *CursosHandler) InstructorListCapacitaciones(ctx *gin.Context) {
	resp, err := h.c.Cursos.InstructorListCapacitaciones(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, cursosToJSON(resp.Cursos))
}

// POST /api/instructor/capacitaciones
func (h *CursosHandler) InstructorCreateCapacitacion(ctx *gin.Context) {
	var body struct {
		Title          string  `json:"title"           binding:"required"`
		Description    string  `json:"description"`
		Type           string  `json:"type"`
		Content        string  `json:"content"`
		IsPublic       bool    `json:"is_public"`
		WelcomeMessage string  `json:"welcome_message"`
		ThumbnailURL   string  `json:"thumbnail_url"`
		Color          string  `json:"color"`
		Precio         float64 `json:"precio"`
		Duration       int32   `json:"duration"`
		Dc3Enabled     *bool   `json:"dc3_enabled"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dc3Enabled := true
	if body.Dc3Enabled != nil {
		dc3Enabled = *body.Dc3Enabled
	}
	resp, err := h.c.Cursos.InstructorCreateCapacitacion(ctx.Request.Context(), &cursospb.CreateCursoRequest{
		UserId:         ctx.GetString(middleware.CtxUserID),
		Title:          body.Title,
		Description:    body.Description,
		Type:           body.Type,
		Content:        body.Content,
		IsPublic:       body.IsPublic,
		WelcomeMessage: body.WelcomeMessage,
		ThumbnailUrl:   body.ThumbnailURL,
		Color:          body.Color,
		Precio:         body.Precio,
		Duration:       body.Duration,
		Dc3Enabled:     dc3Enabled,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, cursoToJSON(resp))
}

// PUT /api/instructor/capacitaciones/:id
func (h *CursosHandler) InstructorUpdateCapacitacion(ctx *gin.Context) {
	var body struct {
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		Type           string  `json:"type"`
		Content        string  `json:"content"`
		IsPublic       bool    `json:"is_public"`
		WelcomeMessage string  `json:"welcome_message"`
		ThumbnailURL   string  `json:"thumbnail_url"`
		Color          string  `json:"color"`
		Precio         float64 `json:"precio"`
		Duration       int32   `json:"duration"`
		Dc3Enabled     *bool   `json:"dc3_enabled"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dc3Enabled := true
	if body.Dc3Enabled != nil {
		dc3Enabled = *body.Dc3Enabled
	}
	resp, err := h.c.Cursos.InstructorUpdateCapacitacion(ctx.Request.Context(), &cursospb.UpdateCursoRequest{
		CursoId:        ctx.Param("id"),
		UserId:         ctx.GetString(middleware.CtxUserID),
		Title:          body.Title,
		Description:    body.Description,
		Type:           body.Type,
		Content:        body.Content,
		IsPublic:       body.IsPublic,
		WelcomeMessage: body.WelcomeMessage,
		ThumbnailUrl:   body.ThumbnailURL,
		Color:          body.Color,
		Precio:         body.Precio,
		Duration:       body.Duration,
		Dc3Enabled:     dc3Enabled,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, cursoToJSON(resp))
}

// DELETE /api/instructor/capacitaciones/:id
func (h *CursosHandler) InstructorDeleteCapacitacion(ctx *gin.Context) {
	_, err := h.c.Cursos.InstructorDeleteCapacitacion(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// PATCH /api/instructor/capacitaciones/:id/toggle-public
func (h *CursosHandler) InstructorTogglePublic(ctx *gin.Context) {
	resp, err := h.c.Cursos.InstructorTogglePublic(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"is_public": resp.IsPublic})
}

// POST /api/instructor/capacitaciones/:id/reset-codigo
func (h *CursosHandler) InstructorResetCodigo(ctx *gin.Context) {
	resp, err := h.c.Cursos.InstructorResetCodigo(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"codigo_acceso": resp.CodigoAcceso})
}

// GET /api/instructor/estudiantes
func (h *CursosHandler) InstructorListEstudiantes(ctx *gin.Context) {
	resp, err := h.c.Cursos.InstructorListEstudiantes(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Estudiantes)
}

// POST /api/instructor/asignar
func (h *CursosHandler) InstructorAsignar(ctx *gin.Context) {
	var body struct {
		UserID         string `json:"user_id"         binding:"required"`
		UserName       string `json:"user_name"`
		UserEmail      string `json:"user_email"`
		CapacitacionID string `json:"capacitacion_id"`
		ExamenID       string `json:"examen_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	md := metadata.Pairs(
		"x-user-name", toASCII(body.UserName),
		"x-user-email", body.UserEmail,
	)
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	_, err := h.c.Cursos.InstructorAsignar(grpcCtx, &cursospb.AsignarRequest{
		RequesterId:    ctx.GetString(middleware.CtxUserID),
		TargetUserId:   body.UserID,
		CapacitacionId: body.CapacitacionID,
		ExamenId:       body.ExamenID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "asignado"})
}

// POST /api/instructor/licencias
func (h *CursosHandler) InstructorCreateLicencia(ctx *gin.Context) {
	var req cursospb.CreateLicenciaRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Cursos.InstructorCreateLicencia(genMetadata(ctx), &req)
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// PUT /api/instructor/licencias/:id
func (h *CursosHandler) InstructorUpdateLicencia(ctx *gin.Context) {
	var req cursospb.UpdateLicenciaRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = ctx.Param("id")
	resp, err := h.c.Cursos.InstructorUpdateLicencia(genMetadata(ctx), &req)
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DELETE /api/instructor/licencias/:id
func (h *CursosHandler) InstructorDeleteLicencia(ctx *gin.Context) {
	_, err := h.c.Cursos.InstructorDeleteLicencia(genMetadata(ctx), &cursospb.LicenciaIDRequest{Id: ctx.Param("id")})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "eliminada"})
}

// ── Admin ─────────────────────────────────────────────────────────────────────

// GET /api/admin/dashboard/stats
func (h *CursosHandler) GetAdminDashboardStats(ctx *gin.Context) {
	resp, err := h.c.Cursos.GetAdminDashboardStats(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GET /api/admin/capacitaciones
func (h *CursosHandler) AdminListCapacitaciones(ctx *gin.Context) {
	resp, err := h.c.Cursos.AdminListCapacitaciones(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Cursos)
}

// POST /api/admin/capacitaciones
func (h *CursosHandler) AdminCreateCapacitacion(ctx *gin.Context) {
	var body struct {
		Title          string `json:"title"           binding:"required"`
		Description    string `json:"description"`
		Type           string `json:"type"`
		Content        string `json:"content"`
		IsPublic       bool   `json:"is_public"`
		WelcomeMessage string `json:"welcome_message"`
		ThumbnailURL   string `json:"thumbnail_url"`
		Color          string `json:"color"`
		Duration       int32  `json:"duration"`
		Dc3Enabled     *bool  `json:"dc3_enabled"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dc3Enabled := true
	if body.Dc3Enabled != nil {
		dc3Enabled = *body.Dc3Enabled
	}
	resp, err := h.c.Cursos.AdminCreateCapacitacion(ctx.Request.Context(), &cursospb.CreateCursoRequest{
		UserId:         ctx.GetString(middleware.CtxUserID),
		Title:          body.Title,
		Description:    body.Description,
		Type:           body.Type,
		Content:        body.Content,
		IsPublic:       body.IsPublic,
		WelcomeMessage: body.WelcomeMessage,
		ThumbnailUrl:   body.ThumbnailURL,
		Color:          body.Color,
		Duration:       body.Duration,
		Dc3Enabled:     dc3Enabled,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// PUT /api/admin/capacitaciones/:id
func (h *CursosHandler) AdminUpdateCapacitacion(ctx *gin.Context) {
	var body struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		Type           string `json:"type"`
		Content        string `json:"content"`
		IsPublic       bool   `json:"is_public"`
		WelcomeMessage string `json:"welcome_message"`
		ThumbnailURL   string `json:"thumbnail_url"`
		Color          string `json:"color"`
		Duration       int32  `json:"duration"`
		Dc3Enabled     *bool  `json:"dc3_enabled"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dc3Enabled := true
	if body.Dc3Enabled != nil {
		dc3Enabled = *body.Dc3Enabled
	}
	resp, err := h.c.Cursos.AdminUpdateCapacitacion(ctx.Request.Context(), &cursospb.UpdateCursoRequest{
		CursoId:        ctx.Param("id"),
		UserId:         ctx.GetString(middleware.CtxUserID),
		Title:          body.Title,
		Description:    body.Description,
		Type:           body.Type,
		Content:        body.Content,
		IsPublic:       body.IsPublic,
		WelcomeMessage: body.WelcomeMessage,
		ThumbnailUrl:   body.ThumbnailURL,
		Color:          body.Color,
		Duration:       body.Duration,
		Dc3Enabled:     dc3Enabled,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DELETE /api/admin/capacitaciones/:id
func (h *CursosHandler) AdminDeleteCapacitacion(ctx *gin.Context) {
	_, err := h.c.Cursos.AdminDeleteCapacitacion(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// POST /api/admin/capacitaciones/:id/reset-codigo
func (h *CursosHandler) AdminResetCodigo(ctx *gin.Context) {
	resp, err := h.c.Cursos.AdminResetCodigo(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"codigo_acceso": resp.CodigoAcceso})
}

// GET /api/admin/asignaciones
func (h *CursosHandler) AdminListAsignaciones(ctx *gin.Context) {
	resp, err := h.c.Cursos.AdminListAsignaciones(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Asignaciones)
}

// POST /api/admin/asignar
func (h *CursosHandler) AdminAsignar(ctx *gin.Context) {
	var body struct {
		UserID         string `json:"user_id"         binding:"required"`
		UserName       string `json:"user_name"`
		UserEmail      string `json:"user_email"`
		CapacitacionID string `json:"capacitacion_id"`
		ExamenID       string `json:"examen_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	md := metadata.Pairs(
		"x-user-name", toASCII(body.UserName),
		"x-user-email", body.UserEmail,
	)
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	_, err := h.c.Cursos.AdminAsignar(grpcCtx, &cursospb.AsignarRequest{
		RequesterId:    ctx.GetString(middleware.CtxUserID),
		TargetUserId:   body.UserID,
		CapacitacionId: body.CapacitacionID,
		ExamenId:       body.ExamenID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "asignado"})
}

// DELETE /api/admin/asignar/:id
func (h *CursosHandler) AdminDesAsignar(ctx *gin.Context) {
	_, err := h.c.Cursos.AdminDesAsignar(ctx.Request.Context(), &cursospb.AsignacionIDRequest{
		AsignacionId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ── Shared error helper ───────────────────────────────────────────────────────

func grpcToHTTP(ctx *gin.Context, err error) {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.NotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		ctx.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	case codes.PermissionDenied:
		ctx.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	// 422: la petición está bien formada, pero el estado del usuario no permite
	// atenderla todavía — por ejemplo completar una lección saltándose la
	// anterior. No es un 400 (nada que corregir en el cuerpo) ni un 500.
	case codes.FailedPrecondition:
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": st.Message()})
	default:
		slog.Error("grpc error", "code", st.Code(), "msg", st.Message(), "path", ctx.FullPath())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}
