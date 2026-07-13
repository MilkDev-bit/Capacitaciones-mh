package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	examenespb "Prueba-Go/gen/examenes"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

func examenesASCII(s string) string {
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
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cuerpo inválido"})
		return
	}

	type itemResp struct {
		PreguntaID     string `json:"pregunta_id"`
		OpcionID       string `json:"opcion_id"`
		RespuestaTexto string `json:"respuesta_texto"`
	}

	var items []itemResp
	if err := json.Unmarshal(bodyBytes, &items); err != nil {
		var wrapper struct {
			Respuestas []itemResp `json:"respuestas"`
		}
		if err2 := json.Unmarshal(bodyBytes, &wrapper); err2 != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "formato de respuestas inválido"})
			return
		}
		items = wrapper.Respuestas
	}

	var respuestas []*examenespb.RespuestaInput
	for _, r := range items {
		respuestas = append(respuestas, &examenespb.RespuestaInput{
			PreguntaId:     r.PreguntaID,
			OpcionId:       r.OpcionID,
			RespuestaTexto: r.RespuestaTexto,
		})
	}
	md := metadata.Pairs("x-user-name", examenesASCII(ctx.GetString(middleware.CtxUserName)))
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
	examenID := ctx.Param("id")
	resp, err := h.c.Examenes.InstructorGetResultados(ctx.Request.Context(), &examenespb.ExamenRequest{
		ExamenId: examenID,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	exResp, _ := h.c.Examenes.GetExamen(ctx.Request.Context(), &examenespb.ExamenUserRequest{
		ExamenId: examenID,
	})
	var puntajeMax float64 = 10.0
	if exResp != nil {
		puntajeMax = 0
		for _, p := range exResp.Preguntas {
			puntajeMax += p.Valor
		}
		if puntajeMax == 0 {
			puntajeMax = 10.0
		}
	}

	out := make([]gin.H, 0, len(resp.Resultados))
	for _, r := range resp.Resultados {
		nombre := r.UserName
		if nombre == "" {
			nombre = "Estudiante"
		}
		fecha := r.SubmittedAt
		if fecha == "" {
			fecha = time.Now().Format(time.RFC3339)
		}
		out = append(out, gin.H{
			"user_id":       r.UserId,
			"nombre":        nombre,
			"user_name":     nombre,
			"email":         "",
			"puntaje":       r.Puntaje,
			"puntaje_max":   puntajeMax,
			"porcentaje":    r.Porcentaje,
			"respondido_at": fecha,
			"submitted_at":  fecha,
		})
	}

	ctx.JSON(http.StatusOK, out)
}

// GET /api/instructor/examenes/:id/resultados/:user_id
func (h *ExamenesHandler) InstructorGetRespuestasUsuario(ctx *gin.Context) {
	resp, err := h.c.Examenes.InstructorGetRespuestasUsuario(ctx.Request.Context(), &examenespb.RespuestasUsuarioRequest{
		ExamenId: ctx.Param("id"),
		UserId:   ctx.Param("user_id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Respuestas)
}

// POST /api/instructor/examenes
func (h *ExamenesHandler) InstructorCreateExamen(ctx *gin.Context) {
	var body struct {
		Title          string `json:"title"    binding:"required"`
		Description    string `json:"description"`
		CapacitacionID string `json:"capacitacion_id"`
		Preguntas      []struct {
			Texto    string  `json:"texto"`
			Tipo     string  `json:"tipo"`
			Valor    float64 `json:"valor"`
			Orden    int32   `json:"orden"`
			Opciones []struct {
				Texto      string `json:"texto"`
				EsCorrecta bool   `json:"es_correcta"`
			} `json:"opciones"`
		} `json:"preguntas"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var preguntas []*examenespb.PreguntaInput
	for _, p := range body.Preguntas {
		var opciones []*examenespb.OpcionInput
		for _, o := range p.Opciones {
			opciones = append(opciones, &examenespb.OpcionInput{Texto: o.Texto, EsCorrecta: o.EsCorrecta})
		}
		preguntas = append(preguntas, &examenespb.PreguntaInput{
			Texto: p.Texto, Tipo: p.Tipo, Valor: p.Valor, Orden: p.Orden, Opciones: opciones,
		})
	}
	resp, err := h.c.Examenes.InstructorCreateExamen(ctx.Request.Context(), &examenespb.CreateExamenRequest{
		UserId:         ctx.GetString(middleware.CtxUserID),
		Title:          body.Title,
		Description:    body.Description,
		CapacitacionId: body.CapacitacionID,
		Preguntas:      preguntas,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
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

// POST /api/admin/examenes
func (h *ExamenesHandler) AdminCreateExamen(ctx *gin.Context) {
	var body struct {
		Title          string `json:"title"    binding:"required"`
		Description    string `json:"description"`
		CapacitacionID string `json:"capacitacion_id"`
		Preguntas      []struct {
			Texto    string  `json:"texto"`
			Tipo     string  `json:"tipo"`
			Valor    float64 `json:"valor"`
			Orden    int32   `json:"orden"`
			Opciones []struct {
				Texto      string `json:"texto"`
				EsCorrecta bool   `json:"es_correcta"`
			} `json:"opciones"`
		} `json:"preguntas"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var preguntas []*examenespb.PreguntaInput
	for _, p := range body.Preguntas {
		var opciones []*examenespb.OpcionInput
		for _, o := range p.Opciones {
			opciones = append(opciones, &examenespb.OpcionInput{Texto: o.Texto, EsCorrecta: o.EsCorrecta})
		}
		preguntas = append(preguntas, &examenespb.PreguntaInput{
			Texto: p.Texto, Tipo: p.Tipo, Valor: p.Valor, Orden: p.Orden, Opciones: opciones,
		})
	}
	resp, err := h.c.Examenes.AdminCreateExamen(ctx.Request.Context(), &examenespb.CreateExamenRequest{
		UserId:         ctx.GetString(middleware.CtxUserID),
		Title:          body.Title,
		Description:    body.Description,
		CapacitacionId: body.CapacitacionID,
		Preguntas:      preguntas,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}
