package handler

import (
	"fmt"
	"net/http"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	"Prueba-Go/gateway/internal/storage"
	authpb "Prueba-Go/gen/auth"
	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/gin-gonic/gin"
)

type UsuariosHandler struct{ c *clients.Clients }

func NewUsuariosHandler(c *clients.Clients) *UsuariosHandler { return &UsuariosHandler{c: c} }

// GET /api/perfil
func (h *UsuariosHandler) GetPerfil(ctx *gin.Context) {
	resp, err := h.c.Usuarios.GetPerfil(ctx.Request.Context(), &usuariospb.GetPerfilRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	stats := gin.H{
		"cursos_inscritos":      resp.CursosInscritos,
		"lecciones_completadas": resp.LeccionesCompletadas,
		"total_lecciones":       resp.TotalLecciones,
		"cursos_creados":        resp.CursosCreados,
		"estudiantes_total":     resp.EstudiantesTotal,
		"examenes_creados":      resp.ExamenesCreados,
	}
	ctx.JSON(http.StatusOK, gin.H{"user": resp, "stats": stats})
}

// PUT /api/perfil
func (h *UsuariosHandler) UpdatePerfil(ctx *gin.Context) {
	var body struct {
		Name      string `json:"name"`
		Bio       string `json:"bio"`
		Phone     string `json:"phone"`
		Specialty string `json:"specialty"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.UpdatePerfil(ctx.Request.Context(), &usuariospb.UpdatePerfilRequest{
		UserId:    ctx.GetString(middleware.CtxUserID),
		Name:      body.Name,
		Bio:       body.Bio,
		Phone:     body.Phone,
		Specialty: body.Specialty,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user": resp})
}

// POST /api/perfil/avatar
func (h *UsuariosHandler) UploadAvatar(ctx *gin.Context) {
	url, err := uploadFileToR2(ctx, "avatars")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.UpdateAvatarURL(ctx.Request.Context(), &usuariospb.UpdateMediaURLRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
		Url:    url,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.AvatarUrl})
}

// POST /api/perfil/cover
func (h *UsuariosHandler) UploadCover(ctx *gin.Context) {
	url, err := uploadFileToR2(ctx, "covers")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.UpdateCoverURL(ctx.Request.Context(), &usuariospb.UpdateMediaURLRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
		Url:    url,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": resp.CoverUrl})
}

// POST /api/perfil/become-instructor
func (h *UsuariosHandler) BecomeInstructor(ctx *gin.Context) {
	resp, err := h.c.Usuarios.BecomeInstructor(ctx.Request.Context(), &usuariospb.UserIDRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"role": resp.Role})
}

// GET /api/usuarios/:id/perfil
func (h *UsuariosHandler) GetPublicPerfil(ctx *gin.Context) {
	resp, err := h.c.Usuarios.GetPublicPerfil(ctx.Request.Context(), &usuariospb.UserIDRequest{
		UserId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user": resp})
}

// GET /api/admin/users
func (h *UsuariosHandler) ListUsers(ctx *gin.Context) {
	resp, err := h.c.Usuarios.ListUsers(ctx.Request.Context(), &usuariospb.ListUsersRequest{
		Role: ctx.Query("role"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Users)
}

// POST /api/admin/users/:id/revoke-sessions
func (h *UsuariosHandler) RevokeUserSessions(ctx *gin.Context) {
	_, err := h.c.Auth.RevokeUserSessions(ctx.Request.Context(), &authpb.RevokeRequest{
		UserId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "sesiones revocadas"})
}

// GET /api/usuarios/search
func (h *UsuariosHandler) SearchUsers(ctx *gin.Context) {
	q := ctx.Query("q")
	if q == "" {
		ctx.JSON(http.StatusOK, []interface{}{})
		return
	}
	resp, err := h.c.Usuarios.SearchUsers(ctx.Request.Context(), &usuariospb.SearchUsersRequest{
		Query:       q,
		Limit:       10,
		RequesterId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Users)
}

// GET /api/notificaciones
func (h *UsuariosHandler) ListNotificaciones(ctx *gin.Context) {
	resp, err := h.c.Usuarios.ListNotificaciones(ctx.Request.Context(), &usuariospb.UserIDRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Notificaciones)
}

// POST /api/notificaciones/marcar-leidas
func (h *UsuariosHandler) MarcarNotificacionesLeidas(ctx *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Usuarios.MarkNotificacionesRead(ctx.Request.Context(), &usuariospb.MarkNotificacionesReadRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
		Ids:    body.IDs,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// uploadFileToR2 lee el multipart del contexto y lo sube a R2.
func uploadFileToR2(ctx *gin.Context, folder string) (string, error) {
	fh, err := ctx.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("archivo requerido")
	}
	return storage.UploadMultipart(ctx.Request.Context(), fh, folder)
}

// PATCH /api/admin/users/:id/role
func (h *UsuariosHandler) AdminUpdateRole(ctx *gin.Context) {
	targetID := ctx.Param("id")
	var body struct {
		Role string `json:"role"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Usuarios.AdminUpdateRole(ctx.Request.Context(), &usuariospb.AdminUpdateRoleRequest{
		AdminId:      ctx.GetString(middleware.CtxUserID),
		TargetUserId: targetID,
		NewRole:      body.Role,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
