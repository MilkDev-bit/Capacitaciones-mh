package service

import (
	"context"
	"errors"

	"Prueba-Go/services/lecciones/internal/badges"
	leccionespb "Prueba-Go/gen/lecciones"
	"Prueba-Go/services/lecciones/internal/repository"
)

var (
	ErrNotFound  = errors.New("no encontrado")
	ErrForbidden = errors.New("sin permisos")
)

type LeccionesService struct {
	repo repository.LeccionesRepository
}

func NewLeccionesService(repo repository.LeccionesRepository) *LeccionesService {
	return &LeccionesService{repo: repo}
}

// ── Árbol del curso ───────────────────────────────────────────────────────────

func (s *LeccionesService) GetCursoTree(ctx context.Context, cursoID, userID string) (*leccionespb.CursoTreeResponse, error) {
	return s.repo.BuildCursoTree(ctx, cursoID, userID)
}

// ── Módulos ───────────────────────────────────────────────────────────────────

func (s *LeccionesService) CreateModulo(ctx context.Context, req *leccionespb.CreateModuloRequest) (*leccionespb.ModuloResponse, error) {
	m, err := s.repo.CreateModulo(ctx, req)
	if err != nil {
		return nil, err
	}
	return m.ToProto(), nil
}

func (s *LeccionesService) UpdateModulo(ctx context.Context, req *leccionespb.UpdateModuloRequest) (*leccionespb.ModuloResponse, error) {
	m, err := s.repo.UpdateModulo(ctx, req)
	if err != nil {
		return nil, err
	}
	return m.ToProto(), nil
}

func (s *LeccionesService) DeleteModulo(ctx context.Context, moduloID string) error {
	return s.repo.DeleteModulo(ctx, moduloID)
}

func (s *LeccionesService) ReorderModulos(ctx context.Context, cursoID string, ids []string) error {
	return s.repo.ReorderModulos(ctx, cursoID, ids)
}

// ── Submódulos ────────────────────────────────────────────────────────────────

func (s *LeccionesService) CreateSubmodulo(ctx context.Context, req *leccionespb.CreateSubmoduloRequest) (*leccionespb.SubmoduloResponse, error) {
	sub, err := s.repo.CreateSubmodulo(ctx, req)
	if err != nil {
		return nil, err
	}
	return sub.ToProto(), nil
}

func (s *LeccionesService) UpdateSubmodulo(ctx context.Context, req *leccionespb.UpdateSubmoduloRequest) (*leccionespb.SubmoduloResponse, error) {
	sub, err := s.repo.UpdateSubmodulo(ctx, req)
	if err != nil {
		return nil, err
	}
	return sub.ToProto(), nil
}

func (s *LeccionesService) DeleteSubmodulo(ctx context.Context, submoduloID string) error {
	return s.repo.DeleteSubmodulo(ctx, submoduloID)
}

func (s *LeccionesService) ReorderSubmodulos(ctx context.Context, moduloID string, ids []string) error {
	return s.repo.ReorderSubmodulos(ctx, moduloID, ids)
}

// ── Lecciones ─────────────────────────────────────────────────────────────────

func (s *LeccionesService) GetLeccionesConProgreso(ctx context.Context, cursoID, userID string) ([]*leccionespb.LeccionResponse, error) {
	lecs, err := s.repo.ListByCursoConProgreso(ctx, cursoID, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*leccionespb.LeccionResponse, 0, len(lecs))
	for _, l := range lecs {
		result = append(result, l.ToProto())
	}
	return result, nil
}

// MarcarCompleta marca la lección como completada y, si la lección tiene
// points_reward y es la primera vez que se completa, otorga los puntos.
func (s *LeccionesService) MarcarCompleta(ctx context.Context, leccionID, userID string) (*leccionespb.MarcarLeccionResponse, error) {
	// ¿Ya estaba completada antes?
	yaCompletada, err := s.repo.IsLeccionCompletada(ctx, leccionID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.MarcarCompleta(ctx, leccionID, userID); err != nil {
		return nil, err
	}

	// Sin puntos si ya estaba completada.
	if yaCompletada {
		return &leccionespb.MarcarLeccionResponse{PointsEarned: 0}, nil
	}

	// Obtener datos de la lección para ver si tiene puntos de recompensa.
	lec, err := s.repo.FindByID(ctx, leccionID)
	if err != nil {
		return nil, err
	}

	var earned, total int32
	if lec.PointsReward > 0 {
		score := &repository.GameScore{
			UserID:         userID,
			LeccionID:      leccionID,
			CapacitacionID: lec.CapacitacionID,
			Points:         lec.PointsReward,
		}
		total, err = s.repo.InsertGameScore(ctx, score)
		if err != nil {
			// No falla silenciosamente: el error de gamificación no debe
			// impedir marcar la lección como completada.
			total = 0
		}
		earned = lec.PointsReward
	}

	return &leccionespb.MarcarLeccionResponse{
		PointsEarned: earned,
		TotalPoints:  total,
	}, nil
}

func (s *LeccionesService) GuardarProgresoVideo(ctx context.Context, leccionID, userID string, segundos int32) error {
	return s.repo.GuardarProgresoVideo(ctx, leccionID, userID, segundos)
}

func (s *LeccionesService) InstructorListLecciones(ctx context.Context, cursoID string) ([]*leccionespb.LeccionResponse, error) {
	lecs, err := s.repo.ListByCurso(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	result := make([]*leccionespb.LeccionResponse, 0, len(lecs))
	for _, l := range lecs {
		result = append(result, l.ToProto())
	}
	return result, nil
}

func (s *LeccionesService) InstructorCreateLeccion(ctx context.Context, req *leccionespb.CreateLeccionRequest) (*leccionespb.LeccionResponse, error) {
	l, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return l.ToProto(), nil
}

func (s *LeccionesService) InstructorUpdateLeccion(ctx context.Context, req *leccionespb.UpdateLeccionRequest) (*leccionespb.LeccionResponse, error) {
	l, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return l.ToProto(), nil
}

func (s *LeccionesService) InstructorDeleteLeccion(ctx context.Context, leccionID string) error {
	return s.repo.Delete(ctx, leccionID)
}

func (s *LeccionesService) InstructorReorder(ctx context.Context, cursoID string, ids []string) error {
	return s.repo.Reorder(ctx, cursoID, ids)
}

// ── Preguntas intermedias ─────────────────────────────────────────────────────

func (s *LeccionesService) GetPreguntasIntermedias(ctx context.Context, cursoID, userID string) ([]*leccionespb.IntermediaResponse, error) {
	return s.buildIntermediasResponse(ctx, cursoID, false)
}

func (s *LeccionesService) SubmitRespuestas(ctx context.Context, req *leccionespb.SubmitIntermediasRequest) (int32, int32, error) {
	return s.repo.SubmitRespuestas(ctx, req.CursoId, req.UserId, req.Respuestas)
}

func (s *LeccionesService) InstructorListPreguntas(ctx context.Context, cursoID string) ([]*leccionespb.IntermediaResponse, error) {
	return s.buildIntermediasResponse(ctx, cursoID, true)
}

func (s *LeccionesService) InstructorCreatePregunta(ctx context.Context, req *leccionespb.CreateIntermediaRequest) (*leccionespb.IntermediaResponse, error) {
	pq, err := s.repo.CreatePregunta(ctx, req)
	if err != nil {
		return nil, err
	}
	opts, _ := s.repo.GetOpciones(ctx, pq.ID)
	return buildIntermediaProto(pq, opts, true), nil
}

func (s *LeccionesService) InstructorDeletePregunta(ctx context.Context, preguntaID string) error {
	return s.repo.DeletePregunta(ctx, preguntaID)
}

// ── Gamificación ──────────────────────────────────────────────────────────────

// SubmitGameScore registra el resultado de un minijuego.
// Si el usuario supera su récord previo en esa lección, is_new_record = true.
// SubmitGameScore registra el resultado de un minijuego con validación completa
// de puntos, detección de récord y evaluación de insignias.
func (s *LeccionesService) SubmitGameScore(ctx context.Context, req *leccionespb.SubmitGameScoreRequest) (*leccionespb.SubmitGameScoreResponse, error) {
	// 1. Obtener datos de la lección (tipo y máximo de puntos).
	lec, err := s.repo.FindByID(ctx, req.LeccionId)
	if err != nil {
		return nil, err
	}

	// 2. Sanitizar puntos: 0 ≤ points ≤ points_reward.
	points := req.Points
	if points < 0 {
		points = 0
	}
	if lec.PointsReward > 0 && points > lec.PointsReward {
		points = lec.PointsReward
	}

	// 3. Puntos previos del usuario en el curso (para detectar récord y umbrales de insignia).
	prevTotal, _ := s.repo.GetUserCoursePoints(ctx, req.UserId, req.CursoId)

	// 4. Persistir el intento.
	score := &repository.GameScore{
		UserID:         req.UserId,
		LeccionID:      req.LeccionId,
		CapacitacionID: req.CursoId,
		Points:         points,
		TimeSecs:       req.TimeSecs,
	}
	newTotal, err := s.repo.InsertGameScore(ctx, score)
	if err != nil {
		return nil, err
	}

	// 5. ¿Primer juego del usuario (ever)?
	isFirstGame := prevTotal == 0 && newTotal > 0

	// 6. Evaluar qué insignias se deben desbloquear.
	ev := badges.GameEvent{
		LessonType:        lec.Type,
		Points:            points,
		PointsReward:      lec.PointsReward,
		TimeSecs:          req.TimeSecs,
		TotalCoursePoints: newTotal,
		IsFirstGame:       isFirstGame,
	}
	candidates := badges.EvaluateRules(ev)

	// 7. Filtrar las que el usuario ya tiene (verifica en BD y persiste nuevas).
	var newBadge string
	for _, slug := range candidates {
		gained, err := s.repo.TryAwardBadge(ctx, req.UserId, slug)
		if err == nil && gained && newBadge == "" {
			newBadge = slug // Reportamos solo la primera nueva insignia al frontend
		}
	}

	// 8. Actualizar cache desnormalizado en users.points_total.
	globalTotal, _ := s.repo.GetUserTotalPoints(ctx, req.UserId)
	_ = s.repo.UpdateUserTotalPoints(ctx, req.UserId, globalTotal)

	return &leccionespb.SubmitGameScoreResponse{
		PointsEarned:  points,
		TotalPoints:   newTotal,
		BadgeUnlocked: newBadge,
		IsNewRecord:   newTotal > prevTotal,
	}, nil
}

func (s *LeccionesService) GetCursoLeaderboard(ctx context.Context, cursoID string, topN int32) (*leccionespb.LeaderboardResponse, error) {
	rows, err := s.repo.GetLeaderboard(ctx, cursoID, int(topN))
	if err != nil {
		return nil, err
	}
	entries := make([]*leccionespb.LeaderboardEntry, 0, len(rows))
	for i, row := range rows {
		entries = append(entries, &leccionespb.LeaderboardEntry{
			Rank:      int32(i + 1),
			UserId:    row.UserID,
			UserName:  row.UserName,
			AvatarUrl: row.AvatarURL,
			Points:    row.Points,
		})
	}
	return &leccionespb.LeaderboardResponse{
		CursoId: cursoID,
		Entries: entries,
	}, nil
}

func (s *LeccionesService) GetUserPoints(ctx context.Context, userID string) (*leccionespb.UserPointsResponse, error) {
	total, err := s.repo.GetUserTotalPoints(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &leccionespb.UserPointsResponse{
		UserId:      userID,
		TotalPoints: total,
	}, nil
}

// ── Entregas de Actividades / Tareas ──────────────────────────────────────────

func (s *LeccionesService) SubmitEntregaActividad(ctx context.Context, req *leccionespb.SubmitEntregaRequest) (*leccionespb.EntregaResponse, error) {
	e, err := s.repo.SubmitEntrega(ctx, req)
	if err != nil {
		return nil, err
	}
	return e.ToProto(), nil
}

func (s *LeccionesService) GetEntregaActividadUsuario(ctx context.Context, req *leccionespb.GetEntregaRequest) (*leccionespb.EntregaResponse, error) {
	e, err := s.repo.GetEntregaUsuario(ctx, req.LeccionId, req.UserId)
	if err != nil {
		return nil, err
	}
	return e.ToProto(), nil
}

func (s *LeccionesService) InstructorListEntregas(ctx context.Context, req *leccionespb.InstructorListEntregasRequest) (*leccionespb.ListEntregasResponse, error) {
	list, err := s.repo.InstructorListEntregas(ctx, req.CursoId, req.LeccionId)
	if err != nil {
		return nil, err
	}
	resp := &leccionespb.ListEntregasResponse{
		Entregas: make([]*leccionespb.EntregaResponse, 0, len(list)),
	}
	for _, e := range list {
		resp.Entregas = append(resp.Entregas, e.ToProto())
	}
	return resp, nil
}

// ── Helpers privados ──────────────────────────────────────────────────────────

func (s *LeccionesService) buildIntermediasResponse(ctx context.Context, cursoID string, showCorrect bool) ([]*leccionespb.IntermediaResponse, error) {
	pqs, err := s.repo.ListPreguntas(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	result := make([]*leccionespb.IntermediaResponse, 0, len(pqs))
	for _, pq := range pqs {
		opts, _ := s.repo.GetOpciones(ctx, pq.ID)
		result = append(result, buildIntermediaProto(pq, opts, showCorrect))
	}
	return result, nil
}

func buildIntermediaProto(pq *repository.PreguntaIntermedia, opts []*repository.OpcionIntermedia, showCorrect bool) *leccionespb.IntermediaResponse {
	r := &leccionespb.IntermediaResponse{
		Id: pq.ID, CursoId: pq.CapacitacionID,
		Texto: pq.Texto, Tipo: pq.Tipo, Orden: pq.Orden,
	}
	if pq.DespuesDeLeccionID != nil {
		r.DespuesDeLeccionId = *pq.DespuesDeLeccionID
	}
	for _, o := range opts {
		oi := &leccionespb.OpcionInfo{Id: o.ID, Texto: o.Texto}
		if showCorrect {
			oi.EsCorrecta = o.EsCorrecta
		}
		r.Opciones = append(r.Opciones, oi)
	}
	return r
}

// checkBadges fue reemplazada por badges.EvaluateRules + TryAwardBadge.
