package handler

import (
	"context"
	"errors"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CursosHandler implementa cursospb.CursosServiceServer.
type CursosHandler struct {
	cursospb.UnimplementedCursosServiceServer
	svc *service.CursosService
}

func NewCursosHandler(svc *service.CursosService) *CursosHandler {
	return &CursosHandler{svc: svc}
}

// ── Público ────────────────────────────────────────────────────────────────────

func (h *CursosHandler) PreviewCurso(ctx context.Context, req *cursospb.CodigoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.PreviewCurso(ctx, req.Codigo)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) ListCursosPublicos(ctx context.Context, _ *cursospb.EmptyRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.ListPublicos(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

// ── Usuario ────────────────────────────────────────────────────────────────────

func (h *CursosHandler) ListMisCapacitaciones(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.ListMisCapacitaciones(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

func (h *CursosHandler) GetCurso(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.GetCurso(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) Inscribirse(ctx context.Context, req *cursospb.InscribirseRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.Inscribirse(ctx, req.UserId, req.CursoId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) UnirseConCodigo(ctx context.Context, req *cursospb.UnirseRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.UnirseConCodigo(ctx, req.UserId, req.Codigo); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (h *CursosHandler) InstructorListCapacitaciones(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.InstructorListCapacitaciones(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

func (h *CursosHandler) InstructorCreateCapacitacion(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorCreate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorUpdateCapacitacion(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorUpdate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorDeleteCapacitacion(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.InstructorDelete(ctx, req.CursoId, req.UserId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) InstructorTogglePublic(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorTogglePublic(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorResetCodigo(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorResetCodigo(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorListEstudiantes(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListEstudiantesResponse, error) {
	// req.UserId es el instructor; el curso viene como parámetro adicional en
	// el gateway (por convención lo pasamos en user_id para este RPC específico).
	// En la práctica el gateway envía instructor_id y el curso_id en campos separados.
	// Como el proto define solo UserRequest, el gateway lo combina; aquí recibimos
	// el instructorID. Sin curso_id en este proto RPC, listamos todos sus cursos de estudiantes.
	list, err := h.svc.InstructorListEstudiantes(ctx, req.UserId, "")
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListEstudiantesResponse{Estudiantes: list}, nil
}

func (h *CursosHandler) InstructorAsignar(ctx context.Context, req *cursospb.AsignarRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.InstructorAsignar(ctx, req.RequesterId, req.TargetUserId, req.CapacitacionId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

// ── Admin ──────────────────────────────────────────────────────────────────────

func (h *CursosHandler) AdminListCapacitaciones(ctx context.Context, _ *cursospb.EmptyRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.AdminList(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

func (h *CursosHandler) AdminCreateCapacitacion(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.AdminCreate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) AdminUpdateCapacitacion(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.AdminUpdate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) AdminDeleteCapacitacion(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.AdminDelete(ctx, req.CursoId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) AdminListAsignaciones(ctx context.Context, _ *cursospb.EmptyRequest) (*cursospb.ListAsignacionesResponse, error) {
	list, err := h.svc.AdminListAsignaciones(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListAsignacionesResponse{Asignaciones: list}, nil
}

func (h *CursosHandler) AdminAsignar(ctx context.Context, req *cursospb.AsignarRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.AdminAsignar(ctx, req.TargetUserId, req.CapacitacionId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) AdminDesAsignar(ctx context.Context, req *cursospb.AsignacionIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.AdminDesAsignar(ctx, req.AsignacionId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

// ── error mapper ──────────────────────────────────────────────────────────────

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "error interno del servidor")
	}
}
