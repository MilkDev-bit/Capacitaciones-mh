package handler

import (
	"net/http"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	examenespb "Prueba-Go/gen/examenes"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type ExamenesHandler struct{ c *clients.Clients }

func NewExamenesHandler(c *clients.Clients) *ExamenesHandler { return &ExamenesHandler{c: c} }

// GET /api/mis-examenes
func (h *ExamenesHandler) ListMisExamenes(ctx *gin.Context) {
	resp, err := h.c.Examenes.ListMisExamenes(ctx.Request.Context(), &examenespb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Examenes)
}

// GET /api/examenes/:id
func (h *ExamenesHandler) GetExamen(ctx *gin.Context) {
	resp, err := h.c.Examenes.GetExamen(ctx.Request.Context(), &examenespb.ExamenUserRequest{
		ExamenId: ctx.Param("id"),
		UserId:   ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// POST /api/examenes/:id/submit
func (h *ExamenesHandler) SubmitExamen(ctx *gin.Context) {
	var body struct {
		Respuestas []struct {
			PreguntaID     string `json:"pregunta_id"`
			OpcionID       string `json:"opcion_id"`
			RespuestaTexto string `json:"respuesta_texto"`
		} `json:"respuestas"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var respuestas []*examenespb.RespuestaInput
	for _, r := range body.Respuestas {
		respuestas = append(respuestas, &examenespb.RespuestaInput{
			PreguntaId:     r.PreguntaID,
			OpcionId:       r.OpcionID,
			RespuestaTexto: r.RespuestaTexto,
		})
	}
	md := metadata.Pairs("x-user-name", ctx.GetString(middleware.CtxUserName))
	grpcCtx := metadata.NewOutgoingContext(ctx.Request.Context(), md)
	resp, err := h.c.Examenes.SubmitExamen(grpcCtx, &examenespb.SubmitRequest{
		ExamenId:   ctx.Param("id"),
		UserId:     ctx.GetString(middleware.CtxUserID),
		Respuestas: respuestas,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ── Instructor ────────────────────────────────────────────────────────────────

// GET /api/instructor/examenes
func (h *ExamenesHandler) InstructorListExamenes(ctx *gin.Context) {
	resp, err := h.c.Examenes.InstructorListExamenes(ctx.Request.Context(), &examenespb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Examenes)
}

// DELETE /api/instructor/examenes/:id
func (h *ExamenesHandler) InstructorDeleteExamen(ctx *gin.Context) {
	_, err := h.c.Examenes.InstructorDeleteExamen(ctx.Request.Context(), &examenespb.ExamenRequest{
		ExamenId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// GET /api/instructor/examenes/:id/resultados
func (h *ExamenesHandler) InstructorGetResultados(ctx *gin.Context) {
	resp, err := h.c.Examenes.InstructorGetResultados(ctx.Request.Context(), &examenespb.ExamenRequest{
		ExamenId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Resultados)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

// GET /api/admin/examenes
func (h *ExamenesHandler) AdminListExamenes(ctx *gin.Context) {
	resp, err := h.c.Examenes.AdminListExamenes(ctx.Request.Context(), &examenespb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Examenes)
}

// DELETE /api/admin/examenes/:id
func (h *ExamenesHandler) AdminDeleteExamen(ctx *gin.Context) {
	_, err := h.c.Examenes.AdminDeleteExamen(ctx.Request.Context(), &examenespb.ExamenRequest{
		ExamenId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
