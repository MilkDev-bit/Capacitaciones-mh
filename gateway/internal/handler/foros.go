package handler

import (
	"net/http"
	"strings"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	forospb "Prueba-Go/gen/foros"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

func foroASCII(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x20 && r <= 0x7E {
			b.WriteRune(r)
		}
	}
	return b.String()
}

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
		Titulo    string `json:"titulo"`
		Contenido string `json:"contenido"`
		MediaURL  string `json:"media_url"`
		MediaType string `json:"media_type"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Titulo == "" && body.Contenido == "" && body.MediaURL == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "debes ingresar contenido o archivo"})
		return
	}
	md := metadata.Pairs("x-user-name", foroASCII(ctx.GetString(middleware.CtxUserName)))
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
	resp, err := h.c.Foros.ListForoComentarios(ctx.Request.Context(), &forospb.PostUserRequest{
		PostId: ctx.Param("post_id"),
		UserId: ctx.GetString(middleware.CtxUserID),
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
		ParentID  string `json:"parent_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	md := metadata.Pairs("x-user-name", foroASCII(ctx.GetString(middleware.CtxUserName)))
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	autorID := ctx.GetString(middleware.CtxUserID)
	resp, err := h.c.Foros.CreateForoComentario(grpcCtx, &forospb.CreateComentarioRequest{
		PostId:    ctx.Param("post_id"),
		UserId:    autorID,
		Contenido: body.Contenido,
		ParentId:  body.ParentID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	h.avisarRespuestaForo(ctx.GetString(middleware.CtxUserName), autorID, resp)

	ctx.JSON(http.StatusCreated, resp)
}

// avisarRespuestaForo notifica a quien fue respondido.
//
// Se avisa al autor del comentario padre y, si es distinto, al autor de la
// publicación: en un hilo largo el dueño del post quiere enterarse de la
// actividad aunque la respuesta cuelgue de otro comentario. El `map` evita que
// reciba dos campanas cuando es la misma persona.
func (h *ForosHandler) avisarRespuestaForo(autorNombre, autorID string, c *forospb.ComentarioResponse) {
	if c == nil {
		return
	}
	if autorNombre == "" {
		autorNombre = "Alguien"
	}

	// El foro no tiene pantalla propia: vive dentro de la capacitación. Si el
	// join no encontró el curso (lección huérfana) se manda al listado, que
	// siempre existe, en lugar de a una URL rota.
	enlace := "/usuario/capacitaciones"
	if c.CapacitacionId != "" {
		enlace += "/" + c.CapacitacionId
	}
	extracto := recorta(c.Contenido, 120)

	destinatarios := map[string]string{}
	if c.ParentUserId != "" {
		destinatarios[c.ParentUserId] = "respondió a tu comentario"
	}
	if _, yaAvisado := destinatarios[c.PostUserId]; !yaAvisado && c.PostUserId != "" {
		destinatarios[c.PostUserId] = "comentó en tu publicación"
	}

	avisos := make([]aviso, 0, len(destinatarios))
	for userID, accion := range destinatarios {
		avisos = append(avisos, aviso{
			UserID:  userID,
			Tipo:    TipoForoRespuesta,
			Titulo:  autorNombre + " " + accion,
			Mensaje: extracto,
			Enlace:  enlace,
			Ventana: ventanaForo,
		})
	}
	notificarSalvoA(h.c, autorID, avisos...)
}

// POST /api/foro/posts/:post_id/reactions
func (h *ForosHandler) ToggleForoPostReaction(ctx *gin.Context) {
	var body struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.c.Foros.ToggleForoPostReaction(ctx.Request.Context(), &forospb.PostReactionRequest{
		PostId: ctx.Param("post_id"),
		UserId: ctx.GetString(middleware.CtxUserID),
		Emoji:  body.Emoji,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"reactions": resp.Reactions})
}

// POST /api/foro/comentarios/:comentario_id/reactions
func (h *ForosHandler) ToggleForoComentarioReaction(ctx *gin.Context) {
	var body struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.c.Foros.ToggleForoComentarioReaction(ctx.Request.Context(), &forospb.ComentarioReactionRequest{
		ComentarioId: ctx.Param("comentario_id"),
		UserId:       ctx.GetString(middleware.CtxUserID),
		Emoji:        body.Emoji,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"reactions": resp.Reactions})
}
