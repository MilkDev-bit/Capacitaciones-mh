package handler

import (
	"net/http"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
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
	ctx.JSON(http.StatusOK, resp)
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
	ctx.JSON(http.StatusOK, resp)
}

// POST /api/perfil/avatar  (multipart — el Gateway sube a R2 y pasa la URL al service)
// La lógica de subida a R2 va aquí o en un helper de storage; el service
// solo recibe la URL resultante.
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
	ctx.JSON(http.StatusOK, resp)
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
	ctx.JSON(http.StatusOK, resp)
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
	ctx.JSON(http.StatusOK, resp)
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
	ctx.JSON(http.StatusOK, resp)
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

// uploadFileToR2 es un placeholder — implementación real en storage/r2.go del gateway.
func uploadFileToR2(ctx *gin.Context, folder string) (string, error) {
	// TODO: leer multipart file, subir a R2 con el SDK de AWS S3 compatible,
	// devolver la URL pública. Ver internal/storage/r2.go del monolito original.
	_ = folder
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "upload to R2 pendiente de implementar en gateway"})
	return "", nil
}
