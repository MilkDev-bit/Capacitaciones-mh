package service

import (
	"context"
	"database/sql"

	examenespb "Prueba-Go/gen/examenes"
	"Prueba-Go/services/examenes/internal/repository"
)

type ExamenesService struct {
	repo repository.ExamenesRepository
}

func NewExamenesService(repo repository.ExamenesRepository) *ExamenesService {
	return &ExamenesService{repo: repo}
}

// buildExamenResponse construye un ExamenResponse con o sin es_correcta.
func (s *ExamenesService) buildExamenResponse(ctx context.Context, e *repository.Examen, showCorrect bool) (*examenespb.ExamenResponse, error) {
	r := &examenespb.ExamenResponse{
		Id: e.ID, Title: e.Title, Description: e.Description,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if e.InstructorID != nil {
		r.InstructorId = *e.InstructorID
	}
	if e.CapacitacionID != nil {
		r.CapacitacionId = *e.CapacitacionID
	}
	preguntas, err := s.repo.GetPreguntas(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	for _, p := range preguntas {
		pq := &examenespb.Pregunta{
			Id: p.ID, Texto: p.Texto, Tipo: p.Tipo,
			Valor: p.Valor, Orden: p.Orden,
		}
		opts, _ := s.repo.GetOpciones(ctx, p.ID)
		for _, o := range opts {
			op := &examenespb.Opcion{Id: o.ID, Texto: o.Texto}
			if showCorrect {
				op.EsCorrecta = o.EsCorrecta
			}
			pq.Opciones = append(pq.Opciones, op)
		}
		r.Preguntas = append(r.Preguntas, pq)
	}
	return r, nil
}

// ── Usuario ────────────────────────────────────────────────────────────────────

func (s *ExamenesService) ListMisExamenes(ctx context.Context, userID string) ([]*examenespb.ExamenResponse, error) {
	exams, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*examenespb.ExamenResponse, 0, len(exams))
	for _, e := range exams {
		r, err := s.buildExamenResponse(ctx, e, false)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *ExamenesService) GetExamen(ctx context.Context, examenID, userID string, showCorrect bool) (*examenespb.ExamenResponse, error) {
	e, err := s.repo.FindByID(ctx, examenID)
	if err != nil {
		return nil, err
	}
	return s.buildExamenResponse(ctx, e, showCorrect)
}

func (s *ExamenesService) SubmitExamen(ctx context.Context, req *examenespb.SubmitRequest) (*examenespb.ResultadoResponse, error) {
	return s.repo.SubmitRespuestas(ctx, req.ExamenId, req.UserId, req.Respuestas)
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (s *ExamenesService) InstructorListExamenes(ctx context.Context, instructorID string) ([]*examenespb.ExamenResponse, error) {
	exams, err := s.repo.ListByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	result := make([]*examenespb.ExamenResponse, 0, len(exams))
	for _, e := range exams {
		r, err := s.buildExamenResponse(ctx, e, true)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *ExamenesService) InstructorCreate(ctx context.Context, req *examenespb.CreateExamenRequest) (*examenespb.ExamenResponse, error) {
	e, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.buildExamenResponse(ctx, e, true)
}

// InstructorGetExamen trae el examen para editarlo.
//
// Distinto de GetExamen, que es el del alumno: aquí `es_correcta` sí viaja
// —sin ella no se puede editar la respuesta correcta— y cada pregunta lleva
// cuántas personas ya la respondieron, que es lo que permite avisar antes de
// borrar algo con historial.
func (s *ExamenesService) InstructorGetExamen(ctx context.Context, examenID, instructorID string) (*examenespb.ExamenResponse, error) {
	e, err := s.repo.FindByID(ctx, examenID)
	if err != nil {
		return nil, err
	}
	// instructorID vacío = admin, que puede abrir cualquiera.
	if instructorID != "" && (e.InstructorID == nil || *e.InstructorID != instructorID) {
		// sql.ErrNoRows y no un "no puedes": el handler lo traduce a NotFound, y
		// así probar identificadores en la URL no revela qué exámenes existen.
		return nil, sql.ErrNoRows
	}

	resp, err := s.buildExamenResponse(ctx, e, true)
	if err != nil {
		return nil, err
	}

	// El conteo se pide UNA vez para todo el examen y se reparte, en lugar de
	// una consulta por pregunta.
	conteo, err := s.repo.RespuestasPorPregunta(ctx, examenID)
	if err != nil {
		return nil, err
	}
	for _, p := range resp.Preguntas {
		p.Respuestas = conteo[p.Id]
	}
	return resp, nil
}

func (s *ExamenesService) InstructorUpdate(ctx context.Context, req *examenespb.UpdateExamenRequest) (*examenespb.ExamenResponse, error) {
	e, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.buildExamenResponse(ctx, e, true)
}

func (s *ExamenesService) InstructorDelete(ctx context.Context, examenID string) error {
	return s.repo.Delete(ctx, examenID)
}

func (s *ExamenesService) InstructorGetResultados(ctx context.Context, examenID string) ([]*examenespb.ResultadoUsuario, error) {
	rows, err := s.repo.GetResultados(ctx, examenID)
	if err != nil {
		return nil, err
	}
	result := make([]*examenespb.ResultadoUsuario, 0, len(rows))
	for _, r := range rows {
		result = append(result, &examenespb.ResultadoUsuario{
			UserId: r.UserID, UserName: r.UserName,
			Puntaje: r.Puntaje, Porcentaje: r.Porcentaje,
		})
	}
	return result, nil
}

func (s *ExamenesService) InstructorGetRespuestasUsuario(ctx context.Context, examenID, userID string) (*examenespb.RespuestasResponse, error) {
	return s.repo.GetRespuestasUsuario(ctx, examenID, userID)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

func (s *ExamenesService) AdminListExamenes(ctx context.Context) ([]*examenespb.ExamenResponse, error) {
	exams, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*examenespb.ExamenResponse, 0, len(exams))
	for _, e := range exams {
		r, err := s.buildExamenResponse(ctx, e, true)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *ExamenesService) AdminCreate(ctx context.Context, req *examenespb.CreateExamenRequest) (*examenespb.ExamenResponse, error) {
	e, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.buildExamenResponse(ctx, e, true)
}

func (s *ExamenesService) AdminDelete(ctx context.Context, examenID string) error {
	return s.repo.Delete(ctx, examenID)
}
