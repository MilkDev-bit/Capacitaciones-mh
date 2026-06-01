package handler

import (
	"log/slog"
	"net/http"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	cursospb "Prueba-Go/gen/cursos"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CursosHandler traduce peticiones HTTP ↔ RPC del cursos service.
type CursosHandler struct {
	c *clients.Clients
}

func NewCursosHandler(c *clients.Clients) *CursosHandler {
	return &CursosHandler{c: c}
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
		"x-user-name", ctx.GetString(middleware.CtxUserName),
		"x-user-email", ctx.GetString(middleware.CtxUserEmail),
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
		"x-user-name", ctx.GetString(middleware.CtxUserName),
		"x-user-email", ctx.GetString(middleware.CtxUserEmail),
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

// ── Admin ─────────────────────────────────────────────────────────────────────

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
