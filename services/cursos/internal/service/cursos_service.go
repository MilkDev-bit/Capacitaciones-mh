package service

import (
	"context"
	"errors"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/repository"
)

// Errores de dominio.
var (
	ErrNotFound  = errors.New("no encontrado")
	ErrForbidden = errors.New("sin permisos")
	ErrConflict  = errors.New("ya inscrito")
)

// CursosService contiene la lógica de negocio del servicio de cursos.
type CursosService struct {
	repo repository.CursosRepository
}

func NewCursosService(repo repository.CursosRepository) *CursosService {
	return &CursosService{repo: repo}
}

func (s *CursosService) ListPublicos(ctx context.Context) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListPublicos(ctx)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) PreviewCurso(ctx context.Context, codigo string) (*cursospb.CursoResponse, error) {
	c, err := s.repo.FindByCodigo(ctx, codigo)
	if err != nil {
		return nil, ErrNotFound
	}
	return c.ToProto(), nil
}

func (s *CursosService) ListMisCapacitaciones(ctx context.Context, userID string) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) GetCurso(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	enrolled, err := s.repo.IsEnrolled(ctx, userID, cursoID)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	// Si no está inscrito, solo puede ver si es público.
	if !enrolled && !c.IsPublic {
		return nil, ErrForbidden
	}
	return c.ToProto(), nil
}

func (s *CursosService) Inscribirse(ctx context.Context, userID, cursoID string) error {
	enrolled, _ := s.repo.IsEnrolled(ctx, userID, cursoID)
	if enrolled {
		return ErrConflict
	}
	return s.repo.Inscribirse(ctx, userID, cursoID)
}

func (s *CursosService) UnirseConCodigo(ctx context.Context, userID, codigo string) error {
	_, err := s.repo.UnirseConCodigo(ctx, userID, codigo)
	return err
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (s *CursosService) InstructorListCapacitaciones(ctx context.Context, instructorID string) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) InstructorCreate(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorUpdate(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	// Verificar que el instructor es dueño del curso.
	existing, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != req.UserId {
		return nil, ErrForbidden
	}
	c, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorDelete(ctx context.Context, cursoID, userID string) error {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, cursoID)
}

func (s *CursosService) InstructorTogglePublic(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return nil, ErrForbidden
	}
	c, err := s.repo.TogglePublic(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorResetCodigo(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return nil, ErrForbidden
	}
	c, err := s.repo.ResetCodigo(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorListEstudiantes(ctx context.Context, instructorID, cursoID string) ([]*cursospb.EstudianteResponse, error) {
	rows, err := s.repo.ListEstudiantes(ctx, instructorID, cursoID)
	if err != nil {
		return nil, err
	}
	result := make([]*cursospb.EstudianteResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, &cursospb.EstudianteResponse{
			Id: r.ID, Name: r.Name, Email: r.Email,
			AssignedAt: r.AssignedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

func (s *CursosService) InstructorAsignar(ctx context.Context, instructorID, userID, cursoID string) error {
	return s.repo.InstructorAsignar(ctx, instructorID, userID, cursoID)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

func (s *CursosService) AdminList(ctx context.Context) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) AdminCreate(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) AdminUpdate(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) AdminDelete(ctx context.Context, cursoID string) error {
	return s.repo.Delete(ctx, cursoID)
}

func (s *CursosService) AdminListAsignaciones(ctx context.Context) ([]*cursospb.AsignacionResponse, error) {
	asigs, err := s.repo.ListAsignaciones(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*cursospb.AsignacionResponse, 0, len(asigs))
	for _, a := range asigs {
		result = append(result, a.ToProto())
	}
	return result, nil
}

func (s *CursosService) AdminAsignar(ctx context.Context, userID, cursoID string) error {
	return s.repo.AdminAsignar(ctx, userID, cursoID)
}

func (s *CursosService) AdminDesAsignar(ctx context.Context, asignacionID string) error {
	return s.repo.DesAsignar(ctx, asignacionID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toProtoSlice(cursos []*repository.Curso) []*cursospb.CursoResponse {
	result := make([]*cursospb.CursoResponse, 0, len(cursos))
	for _, c := range cursos {
		result = append(result, c.ToProto())
	}
	return result
}
