// Package handler implementa el servidor gRPC del usuarios service.
// Patrón de capas:
//
//	HTTP (Gateway) → gRPC (handler) → service (lógica) → repository (BD)
package handler

import (
	"context"
	"log/slog"

	usuariospb "Prueba-Go/gen/usuarios"
	"Prueba-Go/services/usuarios/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UsuariosHandler implementa usuariospb.UsuariosServiceServer.
type UsuariosHandler struct {
	usuariospb.UnimplementedUsuariosServiceServer
	svc *service.UsuariosService
}

func NewUsuariosHandler(svc *service.UsuariosService) *UsuariosHandler {
	return &UsuariosHandler{svc: svc}
}

func (h *UsuariosHandler) GetPerfil(ctx context.Context, req *usuariospb.GetPerfilRequest) (*usuariospb.PerfilResponse, error) {
	perfil, err := h.svc.GetPerfil(ctx, req.UserId)
	if err != nil {
		slog.Error("GetPerfil", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.NotFound, "usuario no encontrado")
	}
	return perfil, nil
}

func (h *UsuariosHandler) UpdatePerfil(ctx context.Context, req *usuariospb.UpdatePerfilRequest) (*usuariospb.PerfilResponse, error) {
	perfil, err := h.svc.UpdatePerfil(ctx, req)
	if err != nil {
		slog.Error("UpdatePerfil", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "error actualizando perfil")
	}
	return perfil, nil
}

func (h *UsuariosHandler) UpdateAvatarURL(ctx context.Context, req *usuariospb.UpdateMediaURLRequest) (*usuariospb.PerfilResponse, error) {
	perfil, err := h.svc.UpdateAvatarURL(ctx, req.UserId, req.Url)
	if err != nil {
		return nil, status.Error(codes.Internal, "error actualizando avatar")
	}
	return perfil, nil
}

func (h *UsuariosHandler) UpdateCoverURL(ctx context.Context, req *usuariospb.UpdateMediaURLRequest) (*usuariospb.PerfilResponse, error) {
	perfil, err := h.svc.UpdateCoverURL(ctx, req.UserId, req.Url)
	if err != nil {
		return nil, status.Error(codes.Internal, "error actualizando portada")
	}
	return perfil, nil
}

func (h *UsuariosHandler) BecomeInstructor(ctx context.Context, req *usuariospb.UserIDRequest) (*usuariospb.PerfilResponse, error) {
	perfil, err := h.svc.BecomeInstructor(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "error cambiando rol")
	}
	return perfil, nil
}

func (h *UsuariosHandler) AdminUpdateRole(ctx context.Context, req *usuariospb.AdminUpdateRoleRequest) (*usuariospb.PerfilResponse, error) {
	perfil, err := h.svc.AdminUpdateRole(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return perfil, nil
}

func (h *UsuariosHandler) GetPublicPerfil(ctx context.Context, req *usuariospb.UserIDRequest) (*usuariospb.PerfilResponse, error) {
	perfil, err := h.svc.GetPerfil(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "usuario no encontrado")
	}
	return perfil, nil
}

func (h *UsuariosHandler) ListUsers(ctx context.Context, req *usuariospb.ListUsersRequest) (*usuariospb.ListUsersResponse, error) {
	resp, err := h.svc.ListUsers(ctx, req.Role)
	if err != nil {
		slog.Error("ListUsers", "error", err)
		return nil, status.Error(codes.Internal, "error listando usuarios")
	}
	return resp, nil
}

func (h *UsuariosHandler) DeleteUser(ctx context.Context, req *usuariospb.UserIDRequest) (*usuariospb.EmptyResponse, error) {
	if err := h.svc.DeleteUser(ctx, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, "error eliminando usuario")
	}
	return &usuariospb.EmptyResponse{}, nil
}

func (h *UsuariosHandler) SearchUsers(ctx context.Context, req *usuariospb.SearchUsersRequest) (*usuariospb.SearchUsersResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	resp, err := h.svc.Search(ctx, req.Query, limit, req.RequesterId)
	if err != nil {
		slog.Error("SearchUsers", "error", err)
		return nil, status.Error(codes.Internal, "error buscando usuarios")
	}
	return &usuariospb.SearchUsersResponse{Users: resp}, nil
}

func (h *UsuariosHandler) ListNotificaciones(ctx context.Context, req *usuariospb.UserIDRequest) (*usuariospb.ListNotificacionesResponse, error) {
	resp, err := h.svc.ListNotificaciones(ctx, req.UserId)
	if err != nil {
		slog.Error("ListNotificaciones", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "error listando notificaciones")
	}
	return resp, nil
}

func (h *UsuariosHandler) MarkNotificacionesRead(ctx context.Context, req *usuariospb.MarkNotificacionesReadRequest) (*usuariospb.EmptyResponse, error) {
	if err := h.svc.MarkNotificacionesRead(ctx, req.UserId, req.Ids); err != nil {
		slog.Error("MarkNotificacionesRead", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "error marcando notificaciones leidas")
	}
	return &usuariospb.EmptyResponse{}, nil
}
