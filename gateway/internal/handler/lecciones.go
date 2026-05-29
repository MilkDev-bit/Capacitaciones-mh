package handler

import (
	"net/http"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	leccionespb "Prueba-Go/gen/lecciones"

	"github.com/gin-gonic/gin"
)

type LeccionesHandler struct{ c *clients.Clients }

func NewLeccionesHandler(c *clients.Clients) *LeccionesHandler { return &LeccionesHandler{c: c} }

// GET /api/capacitaciones/:id/lecciones
func (h *LeccionesHandler) GetLeccionesConProgreso(ctx *gin.Context) {
	resp, err := h.c.Lecciones.GetLeccionesConProgreso(ctx.Request.Context(), &leccionespb.CursoUserRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Lecciones)
}

// POST /api/lecciones/:leccion_id/completar
func (h *LeccionesHandler) MarcarLeccionCompleta(ctx *gin.Context) {
	_, err := h.c.Lecciones.MarcarLeccionCompleta(ctx.Request.Context(), &leccionespb.MarcarRequest{
		LeccionId: ctx.Param("leccion_id"),
		UserId:    ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "lección marcada como completa"})
}

// GET /api/capacitaciones/:id/intermedias
func (h *LeccionesHandler) GetPreguntasIntermedias(ctx *gin.Context) {
	resp, err := h.c.Lecciones.GetPreguntasIntermedias(ctx.Request.Context(), &leccionespb.CursoUserRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Preguntas)
}

// POST /api/capacitaciones/:id/intermedias/submit
func (h *LeccionesHandler) SubmitPreguntasIntermedias(ctx *gin.Context) {
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
	var respuestas []*leccionespb.Respuesta
	for _, r := range body.Respuestas {
		respuestas = append(respuestas, &leccionespb.Respuesta{
			PreguntaId:     r.PreguntaID,
			OpcionId:       r.OpcionID,
			RespuestaTexto: r.RespuestaTexto,
		})
	}
	resp, err := h.c.Lecciones.SubmitPreguntasIntermedias(ctx.Request.Context(), &leccionespb.SubmitIntermediasRequest{
		CursoId:    ctx.Param("id"),
		UserId:     ctx.GetString(middleware.CtxUserID),
		Respuestas: respuestas,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"correctas": resp.Correctas, "total": resp.Total})
}

// ── Instructor ────────────────────────────────────────────────────────────────

// GET /api/instructor/capacitaciones/:id/lecciones
func (h *LeccionesHandler) InstructorListLecciones(ctx *gin.Context) {
	resp, err := h.c.Lecciones.InstructorListLecciones(ctx.Request.Context(), &leccionespb.CursoRequest{
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Lecciones)
}

// POST /api/instructor/capacitaciones/:id/lecciones
func (h *LeccionesHandler) InstructorCreateLeccion(ctx *gin.Context) {
	var body struct {
		Title       string `json:"title"       binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type"`
		FilePath    string `json:"file_path"`
		Content     string `json:"content"`
		Orden       int32  `json:"orden"`
		DuracionMin int32  `json:"duracion_min"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorCreateLeccion(ctx.Request.Context(), &leccionespb.CreateLeccionRequest{
		CursoId:     ctx.Param("id"),
		Title:       body.Title,
		Description: body.Description,
		Type:        body.Type,
		FilePath:    body.FilePath,
		Content:     body.Content,
		Orden:       body.Orden,
		DuracionMin: body.DuracionMin,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// PUT /api/instructor/capacitaciones/:id/lecciones/:leccion_id
func (h *LeccionesHandler) InstructorUpdateLeccion(ctx *gin.Context) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		FilePath    string `json:"file_path"`
		Content     string `json:"content"`
		Orden       int32  `json:"orden"`
		DuracionMin int32  `json:"duracion_min"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorUpdateLeccion(ctx.Request.Context(), &leccionespb.UpdateLeccionRequest{
		LeccionId:   ctx.Param("leccion_id"),
		CursoId:     ctx.Param("id"),
		Title:       body.Title,
		Description: body.Description,
		Type:        body.Type,
		FilePath:    body.FilePath,
		Content:     body.Content,
		Orden:       body.Orden,
		DuracionMin: body.DuracionMin,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DELETE /api/instructor/capacitaciones/:id/lecciones/:leccion_id
func (h *LeccionesHandler) InstructorDeleteLeccion(ctx *gin.Context) {
	_, err := h.c.Lecciones.InstructorDeleteLeccion(ctx.Request.Context(), &leccionespb.LeccionRequest{
		CursoId:   ctx.Param("id"),
		LeccionId: ctx.Param("leccion_id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// PUT /api/instructor/capacitaciones/:id/lecciones/reorder
func (h *LeccionesHandler) InstructorReorderLecciones(ctx *gin.Context) {
	var body struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Lecciones.InstructorReorderLecciones(ctx.Request.Context(), &leccionespb.ReorderRequest{
		CursoId:    ctx.Param("id"),
		LeccionIds: body.IDs,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "orden actualizado"})
}
