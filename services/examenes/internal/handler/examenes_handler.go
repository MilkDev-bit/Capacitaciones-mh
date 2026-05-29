package handler

import (
	"context"

	examenespb "Prueba-Go/gen/examenes"
	"Prueba-Go/services/examenes/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExamenesHandler struct {
	examenespb.UnimplementedExamenesServiceServer
	svc *service.ExamenesService
}

func NewExamenesHandler(svc *service.ExamenesService) *ExamenesHandler {
	return &ExamenesHandler{svc: svc}
}

// ── Usuario ────────────────────────────────────────────────────────────────────

func (h *ExamenesHandler) ListMisExamenes(ctx context.Context, req *examenespb.UserRequest) (*examenespb.ListExamenesResponse, error) {
	list, err := h.svc.ListMisExamenes(ctx, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &examenespb.ListExamenesResponse{Examenes: list}, nil
}

func (h *ExamenesHandler) GetExamen(ctx context.Context, req *examenespb.ExamenUserRequest) (*examenespb.ExamenResponse, error) {
	e, err := h.svc.GetExamen(ctx, req.ExamenId, req.UserId, false)
	if err != nil {
		return nil, toGRPC(err)
	}
	return e, nil
}

func (h *ExamenesHandler) SubmitExamen(ctx context.Context, req *examenespb.SubmitRequest) (*examenespb.ResultadoResponse, error) {
	r, err := h.svc.SubmitExamen(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return r, nil
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (h *ExamenesHandler) InstructorListExamenes(ctx context.Context, req *examenespb.UserRequest) (*examenespb.ListExamenesResponse, error) {
	list, err := h.svc.InstructorListExamenes(ctx, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &examenespb.ListExamenesResponse{Examenes: list}, nil
}

func (h *ExamenesHandler) InstructorCreateExamen(ctx context.Context, req *examenespb.CreateExamenRequest) (*examenespb.ExamenResponse, error) {
	e, err := h.svc.InstructorCreate(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return e, nil
}

func (h *ExamenesHandler) InstructorDeleteExamen(ctx context.Context, req *examenespb.ExamenRequest) (*examenespb.EmptyResponse, error) {
	if err := h.svc.InstructorDelete(ctx, req.ExamenId); err != nil {
		return nil, toGRPC(err)
	}
	return &examenespb.EmptyResponse{}, nil
}

func (h *ExamenesHandler) InstructorGetResultados(ctx context.Context, req *examenespb.ExamenRequest) (*examenespb.ListResultadosResponse, error) {
	list, err := h.svc.InstructorGetResultados(ctx, req.ExamenId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &examenespb.ListResultadosResponse{Resultados: list}, nil
}

func (h *ExamenesHandler) InstructorGetRespuestasUsuario(ctx context.Context, req *examenespb.RespuestasUsuarioRequest) (*examenespb.RespuestasResponse, error) {
	r, err := h.svc.InstructorGetRespuestasUsuario(ctx, req.ExamenId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return r, nil
}

// ── Admin ──────────────────────────────────────────────────────────────────────

func (h *ExamenesHandler) AdminListExamenes(ctx context.Context, _ *examenespb.EmptyRequest) (*examenespb.ListExamenesResponse, error) {
	list, err := h.svc.AdminListExamenes(ctx)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &examenespb.ListExamenesResponse{Examenes: list}, nil
}

func (h *ExamenesHandler) AdminCreateExamen(ctx context.Context, req *examenespb.CreateExamenRequest) (*examenespb.ExamenResponse, error) {
	e, err := h.svc.AdminCreate(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return e, nil
}

func (h *ExamenesHandler) AdminDeleteExamen(ctx context.Context, req *examenespb.ExamenRequest) (*examenespb.EmptyResponse, error) {
	if err := h.svc.AdminDelete(ctx, req.ExamenId); err != nil {
		return nil, toGRPC(err)
	}
	return &examenespb.EmptyResponse{}, nil
}

func toGRPC(err error) error {
	return status.Error(codes.Internal, err.Error())
}
