package service

import (
	"context"
	"errors"

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

func (s *LeccionesService) MarcarCompleta(ctx context.Context, leccionID, userID string) error {
	return s.repo.MarcarCompleta(ctx, leccionID, userID)
}

func (s *LeccionesService) GetPreguntasIntermedias(ctx context.Context, cursoID, userID string) ([]*leccionespb.IntermediaResponse, error) {
	return s.buildIntermediasResponse(ctx, cursoID, false)
}

func (s *LeccionesService) SubmitRespuestas(ctx context.Context, req *leccionespb.SubmitIntermediasRequest) (int32, int32, error) {
	return s.repo.SubmitRespuestas(ctx, req.CursoId, req.UserId, req.Respuestas)
}

// ── Instructor ────────────────────────────────────────────────────────────────

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

func (s *LeccionesService) InstructorListPreguntas(ctx context.Context, cursoID string) ([]*leccionespb.IntermediaResponse, error) {
	return s.buildIntermediasResponse(ctx, cursoID, true) // incluye es_correcta
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

// ── helpers ───────────────────────────────────────────────────────────────────

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
