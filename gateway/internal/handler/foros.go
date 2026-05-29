package handler

import (
	"net/http"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	forospb "Prueba-Go/gen/foros"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type ForosHandler struct{ c *clients.Clients }

func NewForosHandler(c *clients.Clients) *ForosHandler { return &ForosHandler{c: c} }

// GET /api/lecciones/:leccion_id/foro
func (h *ForosHandler) ListForoPosts(ctx *gin.Context) {
	resp, err := h.c.Foros.ListForoPosts(ctx.Request.Context(), &forospb.LeccionRequest{
		LeccionId: ctx.Param("leccion_id"),
		UserId:    ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Posts)
}

// POST /api/lecciones/:leccion_id/foro
func (h *ForosHandler) CreateForoPost(ctx *gin.Context) {
	var body struct {
		Titulo    string `json:"titulo"    binding:"required"`
		Contenido string `json:"contenido" binding:"required"`
		MediaURL  string `json:"media_url"`
		MediaType string `json:"media_type"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	md := metadata.Pairs("x-user-name", ctx.GetString(middleware.CtxUserName))
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	resp, err := h.c.Foros.CreateForoPost(grpcCtx, &forospb.CreatePostRequest{
		LeccionId: ctx.Param("leccion_id"),
		UserId:    ctx.GetString(middleware.CtxUserID),
		Titulo:    body.Titulo,
		Contenido: body.Contenido,
		MediaUrl:  body.MediaURL,
		MediaType: body.MediaType,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// DELETE /api/foro/posts/:post_id
func (h *ForosHandler) DeleteForoPost(ctx *gin.Context) {
	_, err := h.c.Foros.DeleteForoPost(ctx.Request.Context(), &forospb.PostUserRequest{
		PostId: ctx.Param("post_id"),
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// GET /api/foro/posts/:post_id/comentarios
func (h *ForosHandler) ListForoComentarios(ctx *gin.Context) {
	resp, err := h.c.Foros.ListForoComentarios(ctx.Request.Context(), &forospb.PostRequest{
		PostId: ctx.Param("post_id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Comentarios)
}

// POST /api/foro/posts/:post_id/comentarios
func (h *ForosHandler) CreateForoComentario(ctx *gin.Context) {
	var body struct {
		Contenido string `json:"contenido" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	md := metadata.Pairs("x-user-name", ctx.GetString(middleware.CtxUserName))
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	resp, err := h.c.Foros.CreateForoComentario(grpcCtx, &forospb.CreateComentarioRequest{
		PostId:    ctx.Param("post_id"),
		UserId:    ctx.GetString(middleware.CtxUserID),
		Contenido: body.Contenido,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// POST /api/foro/posts/:post_id/like
func (h *ForosHandler) ToggleForoPostLike(ctx *gin.Context) {
	resp, err := h.c.Foros.ToggleForoPostLike(ctx.Request.Context(), &forospb.PostUserRequest{
		PostId: ctx.Param("post_id"),
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"like_count": resp.LikeCount, "user_liked": resp.UserLiked})
}
