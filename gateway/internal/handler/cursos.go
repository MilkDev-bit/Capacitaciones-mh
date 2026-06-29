package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	cursospb "Prueba-Go/gen/cursos"

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
	var b strings.Builder
	for _, r := range s {
		if r >= 0x20 && r <= 0x7E {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CursosHandler traduce peticiones HTTP ↔ RPC del cursos service.
type CursosHandler struct {
	c *clients.Clients
}

func NewCursosHandler(c *clients.Clients) *CursosHandler {
	return &CursosHandler{c: c}
}

// func genMetadata(ctx *gin.Context) context.Context
func genMetadata(ctx *gin.Context) context.Context {
	md := metadata.Pairs(
		"x-user-name", toASCII(ctx.GetString(middleware.CtxUserName)),
		"x-user-email", toASCII(ctx.GetString(middleware.CtxUserEmail)),
	)
	return metadata.NewOutgoingContext(ctx.Request.Context(), md)
}

// GET /api/cursos-publicos
func (h *CursosHandler) ListCursosPublicos(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListCursosPublicos(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Cursos)
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
	ctx.JSON(http.StatusOK, resp)
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
	resp, err := h.c.Cursos.ListMisCapacitaciones(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Cursos)
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
	ctx.JSON(http.StatusOK, resp)
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
	ctx.JSON(http.StatusCreated, gin.H{"message": "inscripción exitosa"})
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
		ScheduleID string `json:"schedule_id"`
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
		ScheduleId: req.ScheduleID,
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
		ScheduleID string `json:"schedule_id"`
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
		ScheduleId: req.ScheduleID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.Url})
}

// POST /api/webhooks/stripe
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

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		ref := session.ClientReferenceID
		parts := strings.Split(ref, "||")
		if len(parts) >= 3 && parts[0] == "curso" {
			// Es una compra de curso individual (B2C)
			// Formato: curso||userID||cursoID[||scheduleID]
			userID := parts[1]
			capID := parts[2]
			scheduleID := ""
			if len(parts) == 4 {
				scheduleID = parts[3]
			}
			_, _ = h.c.Cursos.WebhookEnroll(c.Request.Context(), &cursospb.WebhookEnrollRequest{
				UserId:         userID,
				CapacitacionId: capID,
				LicenciaId:     "", // no hay licencia, es directo
				ScheduleId:     scheduleID,
			})
		} else if len(parts) == 3 && parts[0] == "licencia" {
			// Es una compra de licencia corporativa (B2B)
			// Formato: licencia||userID||licenciaID
			userID := parts[1]
			licID := parts[2]
			_, _ = h.c.Cursos.WebhookComprarLicencia(c.Request.Context(), &cursospb.WebhookComprarLicenciaRequest{
				UserId:     userID,
				LicenciaId: licID,
			})
		} else if len(parts) >= 4 && parts[0] == "b2b_direct" {
			// Es una compra de licencia corporativa en autoservicio (B2B Direct)
			// Formato: b2b_direct||userID||cursoID||cantidad[||scheduleID]
			userID := parts[1]
			cursoID := parts[2]
			cantidadStr := parts[3]
			cantidad, _ := strconv.Atoi(cantidadStr)
			scheduleID := ""
			if len(parts) == 5 {
				scheduleID = parts[4]
			}
			_, _ = h.c.Cursos.WebhookComprarB2BDirect(c.Request.Context(), &cursospb.WebhookComprarB2BDirectRequest{
				UserId:     userID,
				CursoId:    cursoID,
				Cantidad:   int32(cantidad),
				ScheduleId: scheduleID,
			})
		}
	}
	c.Status(http.StatusOK)
}

// POST /api/verify-checkout-session  ← llamado desde el frontend al volver de Stripe
// Permite crear la licencia sin depender del webhook cuando el usuario regresa del pago.
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

	ref := sess.ClientReferenceID
	parts := strings.Split(ref, "||")

	userID := ctx.GetString(middleware.CtxUserID)

	// Pasar el session_id via metadata para que el servicio lo pueda guardar en la BD
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), metadata.Pairs(
		"x-stripe-session-id", req.SessionID,
		"x-user-email", toASCII(ctx.GetString(middleware.CtxUserEmail)),
		"x-user-name", toASCII(ctx.GetString(middleware.CtxUserName)),
	))

	if len(parts) == 3 && parts[0] == "curso" {
		capID := parts[2]
		_, _ = h.c.Cursos.WebhookEnroll(grpcCtx, &cursospb.WebhookEnrollRequest{
			UserId:         userID,
			CapacitacionId: capID,
			LicenciaId:     "",
		})
	} else if len(parts) == 3 && parts[0] == "licencia" {
		licID := parts[2]
		_, _ = h.c.Cursos.WebhookComprarLicencia(grpcCtx, &cursospb.WebhookComprarLicenciaRequest{
			UserId:     userID,
			LicenciaId: licID,
		})
	} else if len(parts) == 4 && parts[0] == "b2b_direct" {
		cursoID := parts[2]
		cantidad, _ := strconv.Atoi(parts[3])
		_, _ = h.c.Cursos.WebhookComprarB2BDirect(grpcCtx, &cursospb.WebhookComprarB2BDirectRequest{
			UserId:   userID,
			CursoId:  cursoID,
			Cantidad: int32(cantidad),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
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
	s, err := stripeSession.Get(sessionID, nil)
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

// GET /api/licencias/:id/tickets
func (h *CursosHandler) GetLicenciaTickets(ctx *gin.Context) {
	licenciaID := ctx.Param("id")

	resp, err := h.c.Cursos.ListLicenciaTickets(genMetadata(ctx), &cursospb.LicenciaIDRequest{Id: licenciaID})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Tickets)
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
	ctx.JSON(http.StatusOK, resp.Cursos)
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
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
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
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
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
		"x-user-name", body.UserName,
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
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		"x-user-name", body.UserName,
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

// ── Videocalls ────────────────────────────────────────────────────────────────

func (h *CursosHandler) JoinVideocall(ctx *gin.Context) {
	var body struct {
		Codigo string `json:"codigo" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.c.Cursos.JoinVideocall(ctx.Request.Context(), &cursospb.JoinVideocallRequest{
		Codigo: body.Codigo,
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *CursosHandler) LeaveVideocall(ctx *gin.Context) {
	var body struct {
		Codigo string `json:"codigo" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Cursos.LeaveVideocall(ctx.Request.Context(), &cursospb.LeaveVideocallRequest{
		Codigo: body.Codigo,
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Has salido de la videollamada"})
}

func (h *CursosHandler) EndVideocall(ctx *gin.Context) {
	_, err := h.c.Cursos.EndVideocall(ctx.Request.Context(), &cursospb.CursoIDRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Videollamada finalizada para todos"})
}

// ── Horarios Instructores (Admin) ─────────────────────────────────────────────

func (h *CursosHandler) AdminListSchedules(ctx *gin.Context) {
	// Se puede pasar instructor_id opcional por query parameter
	instructorID := ctx.Query("instructor_id")
	
	res, err := h.c.Cursos.AdminListSchedules(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: instructorID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res.Schedules)
}

func (h *CursosHandler) AdminCreateSchedule(ctx *gin.Context) {
	var body struct {
		InstructorId string `json:"instructor_id" binding:"required"`
		StartTime    string `json:"start_time" binding:"required"`
		EndTime      string `json:"end_time" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.c.Cursos.AdminCreateSchedule(ctx.Request.Context(), &cursospb.CreateScheduleRequest{
		InstructorId: body.InstructorId,
		StartTime:    body.StartTime,
		EndTime:      body.EndTime,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (h *CursosHandler) AdminUpdateSchedule(ctx *gin.Context) {
	var body struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Status    string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.c.Cursos.AdminUpdateSchedule(ctx.Request.Context(), &cursospb.UpdateScheduleRequest{
		ScheduleId: ctx.Param("id"),
		StartTime:  body.StartTime,
		EndTime:    body.EndTime,
		Status:     body.Status,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *CursosHandler) AdminDeleteSchedule(ctx *gin.Context) {
	_, err := h.c.Cursos.AdminDeleteSchedule(ctx.Request.Context(), &cursospb.ScheduleIDRequest{
		ScheduleId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// GET /api/schedules/public/:instructor_id
func (h *CursosHandler) GetPublicSchedules(ctx *gin.Context) {
	instructorID := ctx.Param("instructor_id")
	resp, err := h.c.Cursos.ListPublicSchedules(ctx.Request.Context(), &cursospb.ListPublicSchedulesRequest{
		InstructorId: instructorID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Schedules)
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
	default:
		slog.Error("grpc error", "code", st.Code(), "msg", st.Message(), "path", ctx.FullPath())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}
