package handler

import (
	"context"
	"database/sql"
	"errors"
	"strings"

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

// ── Árbol del curso ───────────────────────────────────────────────────────────

func (h *LeccionesHandler) GetCursoTree(ctx context.Context, req *leccionespb.CursoUserRequest) (*leccionespb.CursoTreeResponse, error) {
	tree, err := h.svc.GetCursoTree(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return tree, nil
}

func (h *LeccionesHandler) InstructorGetCursoTree(ctx context.Context, req *leccionespb.CursoRequest) (*leccionespb.CursoTreeResponse, error) {
	tree, err := h.svc.GetCursoTree(ctx, req.CursoId, "") // sin userID → completada=false
	if err != nil {
		return nil, toGRPC(err)
	}
	return tree, nil
}

// ── Módulos ───────────────────────────────────────────────────────────────────

func (h *LeccionesHandler) InstructorCreateModulo(ctx context.Context, req *leccionespb.CreateModuloRequest) (*leccionespb.ModuloResponse, error) {
	m, err := h.svc.CreateModulo(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return m, nil
}

func (h *LeccionesHandler) InstructorUpdateModulo(ctx context.Context, req *leccionespb.UpdateModuloRequest) (*leccionespb.ModuloResponse, error) {
	m, err := h.svc.UpdateModulo(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return m, nil
}

func (h *LeccionesHandler) InstructorDeleteModulo(ctx context.Context, req *leccionespb.ModuloIDRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.DeleteModulo(ctx, req.ModuloId); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

func (h *LeccionesHandler) InstructorReorderModulos(ctx context.Context, req *leccionespb.ReorderModulosRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.ReorderModulos(ctx, req.CursoId, req.ModuloIds); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

// ── Submódulos ────────────────────────────────────────────────────────────────

func (h *LeccionesHandler) InstructorCreateSubmodulo(ctx context.Context, req *leccionespb.CreateSubmoduloRequest) (*leccionespb.SubmoduloResponse, error) {
	sub, err := h.svc.CreateSubmodulo(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return sub, nil
}

func (h *LeccionesHandler) InstructorUpdateSubmodulo(ctx context.Context, req *leccionespb.UpdateSubmoduloRequest) (*leccionespb.SubmoduloResponse, error) {
	sub, err := h.svc.UpdateSubmodulo(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return sub, nil
}

func (h *LeccionesHandler) InstructorDeleteSubmodulo(ctx context.Context, req *leccionespb.SubmoduloIDRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.DeleteSubmodulo(ctx, req.SubmoduloId); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

func (h *LeccionesHandler) InstructorReorderSubmodulos(ctx context.Context, req *leccionespb.ReorderSubmodulosRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.ReorderSubmodulos(ctx, req.ModuloId, req.SubmoduloIds); err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.EmptyResponse{}, nil
}

// ── Lecciones ─────────────────────────────────────────────────────────────────

func (h *LeccionesHandler) GetLeccionesConProgreso(ctx context.Context, req *leccionespb.CursoUserRequest) (*leccionespb.ListLeccionesResponse, error) {
	list, err := h.svc.GetLeccionesConProgreso(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &leccionespb.ListLeccionesResponse{Lecciones: list}, nil
}

func (h *LeccionesHandler) MarcarLeccionCompleta(ctx context.Context, req *leccionespb.MarcarRequest) (*leccionespb.MarcarLeccionResponse, error) {
	resp, err := h.svc.MarcarCompleta(ctx, req.LeccionId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return resp, nil
}

func (h *LeccionesHandler) GuardarProgresoVideo(ctx context.Context, req *leccionespb.GuardarProgresoVideoRequest) (*leccionespb.EmptyResponse, error) {
	if err := h.svc.GuardarProgresoVideo(ctx, req.LeccionId, req.UserId, req.SegundosVistos); err != nil {
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

// ── Gamificación ──────────────────────────────────────────────────────────────

func (h *LeccionesHandler) SubmitGameScore(ctx context.Context, req *leccionespb.SubmitGameScoreRequest) (*leccionespb.SubmitGameScoreResponse, error) {
	resp, err := h.svc.SubmitGameScore(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return resp, nil
}

func (h *LeccionesHandler) GetCursoLeaderboard(ctx context.Context, req *leccionespb.LeaderboardRequest) (*leccionespb.LeaderboardResponse, error) {
	resp, err := h.svc.GetCursoLeaderboard(ctx, req.CursoId, req.TopN)
	if err != nil {
		return nil, toGRPC(err)
	}
	return resp, nil
}

func (h *LeccionesHandler) GetUserPoints(ctx context.Context, req *leccionespb.UserRequest) (*leccionespb.UserPointsResponse, error) {
	resp, err := h.svc.GetUserPoints(ctx, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return resp, nil
}

// ── Error helper ──────────────────────────────────────────────────────────────

func toGRPC(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return status.Error(codes.NotFound, "recurso no encontrado")
	}
	if strings.Contains(err.Error(), "invalid input syntax for type uuid") || strings.Contains(err.Error(), "SQLSTATE 22P02") {
		return status.Error(codes.InvalidArgument, "ID de recurso inválido")
	}
	return status.Error(codes.Internal, err.Error())
}
