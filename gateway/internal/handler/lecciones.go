package handler

import (
	"net/http"
	"strconv"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	leccionespb "Prueba-Go/gen/lecciones"
	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/gin-gonic/gin"
)

type LeccionesHandler struct{ c *clients.Clients }

func NewLeccionesHandler(c *clients.Clients) *LeccionesHandler { return &LeccionesHandler{c: c} }

// ── Árbol del curso ───────────────────────────────────────────────────────────

// GET /api/capacitaciones/:id/tree
func (h *LeccionesHandler) GetCursoTree(ctx *gin.Context) {
	resp, err := h.c.Lecciones.GetCursoTree(ctx.Request.Context(), &leccionespb.CursoUserRequest{
		CursoId: ctx.Param("id"),
		UserId:  ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GET /api/instructor/capacitaciones/:id/tree
func (h *LeccionesHandler) InstructorGetCursoTree(ctx *gin.Context) {
	resp, err := h.c.Lecciones.InstructorGetCursoTree(ctx.Request.Context(), &leccionespb.CursoRequest{
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ── Módulos ───────────────────────────────────────────────────────────────────

// POST /api/instructor/capacitaciones/:id/modulos
func (h *LeccionesHandler) InstructorCreateModulo(ctx *gin.Context) {
	var body struct {
		Title       string `json:"title"       binding:"required"`
		Description string `json:"description"`
		Orden       int32  `json:"orden"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorCreateModulo(ctx.Request.Context(), &leccionespb.CreateModuloRequest{
		CursoId:     ctx.Param("id"),
		Title:       body.Title,
		Description: body.Description,
		Orden:       body.Orden,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// PUT /api/instructor/capacitaciones/:id/modulos/:modulo_id
func (h *LeccionesHandler) InstructorUpdateModulo(ctx *gin.Context) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Orden       int32  `json:"orden"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorUpdateModulo(ctx.Request.Context(), &leccionespb.UpdateModuloRequest{
		ModuloId:    ctx.Param("modulo_id"),
		CursoId:     ctx.Param("id"),
		Title:       body.Title,
		Description: body.Description,
		Orden:       body.Orden,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DELETE /api/instructor/capacitaciones/:id/modulos/:modulo_id
func (h *LeccionesHandler) InstructorDeleteModulo(ctx *gin.Context) {
	_, err := h.c.Lecciones.InstructorDeleteModulo(ctx.Request.Context(), &leccionespb.ModuloIDRequest{
		ModuloId: ctx.Param("modulo_id"),
		CursoId:  ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// PUT /api/instructor/capacitaciones/:id/modulos/reorder
func (h *LeccionesHandler) InstructorReorderModulos(ctx *gin.Context) {
	var body struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Lecciones.InstructorReorderModulos(ctx.Request.Context(), &leccionespb.ReorderModulosRequest{
		CursoId:   ctx.Param("id"),
		ModuloIds: body.IDs,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "orden de módulos actualizado"})
}

// ── Submódulos ────────────────────────────────────────────────────────────────

// POST /api/instructor/capacitaciones/:id/modulos/:modulo_id/submodulos
func (h *LeccionesHandler) InstructorCreateSubmodulo(ctx *gin.Context) {
	var body struct {
		Title       string `json:"title"       binding:"required"`
		Description string `json:"description"`
		Orden       int32  `json:"orden"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorCreateSubmodulo(ctx.Request.Context(), &leccionespb.CreateSubmoduloRequest{
		ModuloId:    ctx.Param("modulo_id"),
		CursoId:     ctx.Param("id"),
		Title:       body.Title,
		Description: body.Description,
		Orden:       body.Orden,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// PUT /api/instructor/capacitaciones/:id/modulos/:modulo_id/submodulos/:submodulo_id
func (h *LeccionesHandler) InstructorUpdateSubmodulo(ctx *gin.Context) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Orden       int32  `json:"orden"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorUpdateSubmodulo(ctx.Request.Context(), &leccionespb.UpdateSubmoduloRequest{
		SubmoduloId: ctx.Param("submodulo_id"),
		ModuloId:    ctx.Param("modulo_id"),
		Title:       body.Title,
		Description: body.Description,
		Orden:       body.Orden,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DELETE /api/instructor/capacitaciones/:id/modulos/:modulo_id/submodulos/:submodulo_id
func (h *LeccionesHandler) InstructorDeleteSubmodulo(ctx *gin.Context) {
	_, err := h.c.Lecciones.InstructorDeleteSubmodulo(ctx.Request.Context(), &leccionespb.SubmoduloIDRequest{
		SubmoduloId: ctx.Param("submodulo_id"),
		ModuloId:    ctx.Param("modulo_id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// PUT /api/instructor/capacitaciones/:id/modulos/:modulo_id/submodulos/reorder
func (h *LeccionesHandler) InstructorReorderSubmodulos(ctx *gin.Context) {
	var body struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.c.Lecciones.InstructorReorderSubmodulos(ctx.Request.Context(), &leccionespb.ReorderSubmodulosRequest{
		ModuloId:     ctx.Param("modulo_id"),
		SubmoduloIds: body.IDs,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "orden de submódulos actualizado"})
}

// ── Lecciones ─────────────────────────────────────────────────────────────────

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
	resp, err := h.c.Lecciones.MarcarLeccionCompleta(ctx.Request.Context(), &leccionespb.MarcarRequest{
		LeccionId: ctx.Param("leccion_id"),
		UserId:    ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	// Devuelve puntos ganados e insignia desbloqueada (si la hay)
	ctx.JSON(http.StatusOK, gin.H{
		"message":       "lección marcada como completa",
		"points_earned": resp.PointsEarned,
		"total_points":  resp.TotalPoints,
		"badge_unlocked": resp.BadgeUnlocked,
	})
}

// POST /api/lecciones/:leccion_id/progreso-video
func (h *LeccionesHandler) GuardarProgresoVideo(ctx *gin.Context) {
	var body struct {
		SegundosVistos int32 `json:"segundos_vistos"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "segundos_vistos requerido"})
		return
	}
	_, err := h.c.Lecciones.GuardarProgresoVideo(ctx.Request.Context(), &leccionespb.GuardarProgresoVideoRequest{
		LeccionId:      ctx.Param("leccion_id"),
		UserId:         ctx.GetString(middleware.CtxUserID),
		SegundosVistos: body.SegundosVistos,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "progreso guardado"})
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

// ── Gamificación — Minijuegos ─────────────────────────────────────────────────

// POST /api/lecciones/:leccion_id/game-score
// Body: { "curso_id": "...", "points": 80, "time_secs": 45 }
func (h *LeccionesHandler) SubmitGameScore(ctx *gin.Context) {
	var body struct {
		CursoID  string `json:"curso_id"  binding:"required"`
		Points   int32  `json:"points"`
		TimeSecs int32  `json:"time_secs"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.SubmitGameScore(ctx.Request.Context(), &leccionespb.SubmitGameScoreRequest{
		UserId:    ctx.GetString(middleware.CtxUserID),
		LeccionId: ctx.Param("leccion_id"),
		CursoId:   body.CursoID,
		Points:    body.Points,
		TimeSecs:  body.TimeSecs,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GET /api/capacitaciones/:id/leaderboard?top=5
func (h *LeccionesHandler) GetLeaderboard(ctx *gin.Context) {
	topN := int32(5)
	if t := ctx.Query("top"); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			topN = int32(n)
		}
	}
	resp, err := h.c.Lecciones.GetCursoLeaderboard(ctx.Request.Context(), &leccionespb.LeaderboardRequest{
		CursoId: ctx.Param("id"),
		TopN:    topN,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	for _, entry := range resp.Entries {
		if (entry.UserName == "" || entry.UserName == "Anónimo") && entry.UserId != "" && h.c.Usuarios != nil {
			uResp, err := h.c.Usuarios.GetPublicPerfil(ctx.Request.Context(), &usuariospb.UserIDRequest{UserId: entry.UserId})
			if err == nil && uResp != nil {
				entry.UserName = uResp.Name
				if entry.UserName == "" {
					entry.UserName = uResp.Email
				}
				entry.AvatarUrl = uResp.AvatarUrl
			}
		}
	}
	ctx.JSON(http.StatusOK, resp)
}

// GET /api/gamificacion/mis-puntos
func (h *LeccionesHandler) GetMisPuntos(ctx *gin.Context) {
	resp, err := h.c.Lecciones.GetUserPoints(ctx.Request.Context(), &leccionespb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GET /api/gamificacion/insignias
// Devuelve el catálogo completo de insignias con estado desbloqueado/bloqueado del usuario actual.
func (h *LeccionesHandler) GetMisInsignias(ctx *gin.Context) {
	// El catálogo en Go lo servimos directamente sin ir al servicio —
	// solo necesitamos los slugs desbloqueados del usuario.
	// En el futuro esto podría moverse a un endpoint del servicio de usuarios.
	resp, err := h.c.Lecciones.GetUserPoints(ctx.Request.Context(), &leccionespb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
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
		Title          string `json:"title"           binding:"required"`
		Description    string `json:"description"`
		LessonType     int32  `json:"lesson_type"`     // valor del enum LessonType
		FilePath       string `json:"file_path"`
		Content        string `json:"content"`
		Orden          int32  `json:"orden"`
		DuracionMin    int32  `json:"duracion_min"`
		ModuloID       string `json:"modulo_id"`
		SubmoduloID    string `json:"submodulo_id"`
		GameConfigJSON string `json:"game_config_json"` // JSON libre, validado en el frontend
		PointsReward   int32  `json:"points_reward"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorCreateLeccion(ctx.Request.Context(), &leccionespb.CreateLeccionRequest{
		CursoId:        ctx.Param("id"),
		Title:          body.Title,
		Description:    body.Description,
		LessonType:     leccionespb.LessonType(body.LessonType),
		FilePath:       body.FilePath,
		Content:        body.Content,
		Orden:          body.Orden,
		DuracionMin:    body.DuracionMin,
		ModuloId:       body.ModuloID,
		SubmoduloId:    body.SubmoduloID,
		GameConfigJson: body.GameConfigJSON,
		PointsReward:   body.PointsReward,
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
		Title          string `json:"title"`
		Description    string `json:"description"`
		LessonType     int32  `json:"lesson_type"`
		FilePath       string `json:"file_path"`
		Content        string `json:"content"`
		Orden          int32  `json:"orden"`
		DuracionMin    int32  `json:"duracion_min"`
		ModuloID       string `json:"modulo_id"`
		SubmoduloID    string `json:"submodulo_id"`
		GameConfigJSON string `json:"game_config_json"`
		PointsReward   int32  `json:"points_reward"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.c.Lecciones.InstructorUpdateLeccion(ctx.Request.Context(), &leccionespb.UpdateLeccionRequest{
		LeccionId:      ctx.Param("leccion_id"),
		CursoId:        ctx.Param("id"),
		Title:          body.Title,
		Description:    body.Description,
		LessonType:     leccionespb.LessonType(body.LessonType),
		FilePath:       body.FilePath,
		Content:        body.Content,
		Orden:          body.Orden,
		DuracionMin:    body.DuracionMin,
		ModuloId:       body.ModuloID,
		SubmoduloId:    body.SubmoduloID,
		GameConfigJson: body.GameConfigJSON,
		PointsReward:   body.PointsReward,
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

// GET /api/instructor/capacitaciones/:id/intermedias
func (h *LeccionesHandler) InstructorListPreguntasIntermedias(ctx *gin.Context) {
	resp, err := h.c.Lecciones.InstructorListPreguntasIntermedias(ctx.Request.Context(), &leccionespb.CursoRequest{
		CursoId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Preguntas)
}

// POST /api/instructor/capacitaciones/:id/intermedias
func (h *LeccionesHandler) InstructorCreatePreguntaIntermedia(ctx *gin.Context) {
	var body struct {
		DespuesDeLeccionID string `json:"despues_de_leccion_id"`
		Texto              string `json:"texto"  binding:"required"`
		Tipo               string `json:"tipo"   binding:"required"`
		Orden              int32  `json:"orden"`
		Opciones           []struct {
			Texto      string `json:"texto"`
			EsCorrecta bool   `json:"es_correcta"`
		} `json:"opciones"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var opciones []*leccionespb.OpcionInput
	for _, o := range body.Opciones {
		opciones = append(opciones, &leccionespb.OpcionInput{Texto: o.Texto, EsCorrecta: o.EsCorrecta})
	}
	resp, err := h.c.Lecciones.InstructorCreatePreguntaIntermedia(ctx.Request.Context(), &leccionespb.CreateIntermediaRequest{
		CursoId:            ctx.Param("id"),
		DespuesDeLeccionId: body.DespuesDeLeccionID,
		Texto:              body.Texto,
		Tipo:               body.Tipo,
		Orden:              body.Orden,
		Opciones:           opciones,
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// DELETE /api/instructor/capacitaciones/:id/intermedias/:pregunta_id
func (h *LeccionesHandler) InstructorDeletePreguntaIntermedia(ctx *gin.Context) {
	_, err := h.c.Lecciones.InstructorDeletePreguntaIntermedia(ctx.Request.Context(), &leccionespb.IntermediaIDRequest{
		PreguntaId: ctx.Param("pregunta_id"),
		CursoId:    ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
