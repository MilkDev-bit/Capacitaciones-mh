package handler

import (
	"context"

	leccionespb "Prueba-Go/gen/lecciones"
	"Prueba-Go/services/lecciones/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LeccionesHandler struct {
	leccionespb.UnimplementedLeccionesServiceServer
	svc *service.LeccionesService
}

func NewLeccionesHandler(svc *service.LeccionesService) *LeccionesHandler {
	return &LeccionesHandler{svc: svc}
}

func (h *LeccionesHandler) GetLeccionesConProgreso(ctx context.Context, req *leccionespb.CursoUserRequest) (*leccionespb.ListLeccionesResponse, error) {
	list, err := h.svc.GetLeccionesConProgreso(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.ListLeccionesResponse{Lecciones: list}, nil
}

func (h *LeccionesHandler) MarcarLeccionCompleta(ctx context.Context, req *leccionespb.MarcarRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.MarcarCompleta(ctx, req.LeccionId, req.UserId); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

func (h *LeccionesHandler) GetPreguntasIntermedias(ctx context.Context, req *leccionespb.CursoUserRequest) (*leccionespb.ListIntermediasResponse, error) {
	list, err := h.svc.GetPreguntasIntermedias(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.ListIntermediasResponse{Preguntas: list}, nil
}

func (h *LeccionesHandler) SubmitPreguntasIntermedias(ctx context.Context, req *leccionespb.SubmitIntermediasRequest) (*leccionespb.SubmitIntermediasResponse, error) {
	correctas, total, err := h.svc.SubmitRespuestas(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.SubmitIntermediasResponse{Correctas: correctas, Total: total}, nil
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (h *LeccionesHandler) InstructorListLecciones(ctx context.Context, req *leccionespb.CursoRequest) (*leccionespb.ListLeccionesResponse, error) {
	list, err := h.svc.InstructorListLecciones(ctx, req.CursoId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.ListLeccionesResponse{Lecciones: list}, nil
}

func (h *LeccionesHandler) InstructorCreateLeccion(ctx context.Context, req *leccionespb.CreateLeccionRequest) (*leccionespb.LeccionResponse, error) {
	l, err := h.svc.InstructorCreateLeccion(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return l, nil
}

func (h *LeccionesHandler) InstructorUpdateLeccion(ctx context.Context, req *leccionespb.UpdateLeccionRequest) (*leccionespb.LeccionResponse, error) {
	l, err := h.svc.InstructorUpdateLeccion(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return l, nil
}

func (h *LeccionesHandler) InstructorDeleteLeccion(ctx context.Context, req *leccionespb.LeccionRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.InstructorDeleteLeccion(ctx, req.LeccionId); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

func (h *LeccionesHandler) InstructorReorderLecciones(ctx context.Context, req *leccionespb.ReorderRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.InstructorReorder(ctx, req.CursoId, req.LeccionIds); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

func (h *LeccionesHandler) InstructorListPreguntasIntermedias(ctx context.Context, req *leccionespb.CursoRequest) (*leccionespb.ListIntermediasResponse, error) {
	list, err := h.svc.InstructorListPreguntas(ctx, req.CursoId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.ListIntermediasResponse{Preguntas: list}, nil
}

func (h *LeccionesHandler) InstructorCreatePreguntaIntermedia(ctx context.Context, req *leccionespb.CreateIntermediaRequest) (*leccionespb.IntermediaResponse, error) {
	r, err := h.svc.InstructorCreatePregunta(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return r, nil
}

func (h *LeccionesHandler) InstructorDeletePreguntaIntermedia(ctx context.Context, req *leccionespb.IntermediaIDRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.InstructorDeletePregunta(ctx, req.PreguntaId); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

func toGRPC(err error) error {
	return status.Error(codes.Internal, err.Error())
}
