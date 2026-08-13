// Package service contiene la lógica de negocio del usuarios service.
// No conoce ni HTTP ni gRPC — solo opera con tipos de dominio.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	usuariospb "Prueba-Go/gen/usuarios"
	"Prueba-Go/services/usuarios/internal/repository"
)

// Errores de validación de notificaciones. Se exportan para que el handler gRPC
// pueda traducirlos a InvalidArgument en vez de a un Internal genérico.
var (
	ErrDatosNotificacion = errors.New("faltan datos obligatorios de la notificación")
	ErrTipoNotificacion  = errors.New("tipo de notificación no reconocido")
)

// UsuariosService encapsula la lógica de gestión de perfiles.
type UsuariosService struct {
	repo repository.UsuarioRepository
}

func NewUsuariosService(repo repository.UsuarioRepository) *UsuariosService {
	return &UsuariosService{repo: repo}
}

func (s *UsuariosService) GetPerfil(ctx context.Context, userID string) (*usuariospb.PerfilResponse, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.ToProto(), nil
}

func (s *UsuariosService) UpdatePerfil(ctx context.Context, req *usuariospb.UpdatePerfilRequest) (*usuariospb.PerfilResponse, error) {
	if err := s.repo.UpdatePerfil(ctx, req); err != nil {
		return nil, fmt.Errorf("update perfil: %w", err)
	}
	return s.GetPerfil(ctx, req.UserId)
}

func (s *UsuariosService) UpdateAvatarURL(ctx context.Context, userID, url string) (*usuariospb.PerfilResponse, error) {
	if err := s.repo.UpdateField(ctx, userID, "avatar_url", url); err != nil {
		return nil, err
	}
	return s.GetPerfil(ctx, userID)
}

func (s *UsuariosService) UpdateCoverURL(ctx context.Context, userID, url string) (*usuariospb.PerfilResponse, error) {
	if err := s.repo.UpdateField(ctx, userID, "cover_url", url); err != nil {
		return nil, err
	}
	return s.GetPerfil(ctx, userID)
}

func (s *UsuariosService) BecomeInstructor(ctx context.Context, userID string) (*usuariospb.PerfilResponse, error) {
	if err := s.repo.UpdateField(ctx, userID, "role", "instructor"); err != nil {
		return nil, err
	}
	return s.GetPerfil(ctx, userID)
}

func (s *UsuariosService) AdminUpdateRole(ctx context.Context, req *usuariospb.AdminUpdateRoleRequest) (*usuariospb.PerfilResponse, error) {
	admin, err := s.repo.FindByID(ctx, req.AdminId)
	if err != nil {
		return nil, err
	}
	if admin.Role != "admin" {
		return nil, fmt.Errorf("forbidden: only admins can update roles")
	}

	if req.NewRole != "admin" && req.NewRole != "instructor" && req.NewRole != "user" {
		return nil, fmt.Errorf("invalid role: %s", req.NewRole)
	}

	if err := s.repo.UpdateField(ctx, req.TargetUserId, "role", req.NewRole); err != nil {
		return nil, err
	}
	return s.GetPerfil(ctx, req.TargetUserId)
}

func (s *UsuariosService) ListUsers(ctx context.Context, role string) (*usuariospb.ListUsersResponse, error) {
	users, err := s.repo.List(ctx, role)
	if err != nil {
		return nil, err
	}
	var summaries []*usuariospb.UserSummary
	for _, u := range users {
		summaries = append(summaries, u.ToSummaryProto())
	}
	return &usuariospb.ListUsersResponse{Users: summaries, Total: int32(len(summaries))}, nil
}

func (s *UsuariosService) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}

func (s *UsuariosService) Search(ctx context.Context, query string, limit int, requesterID string) ([]*usuariospb.UserSummary, error) {
	users, err := s.repo.Search(ctx, query, limit, requesterID)
	if err != nil {
		return nil, err
	}
	var summaries []*usuariospb.UserSummary
	for _, u := range users {
		summaries = append(summaries, u.ToSummaryProto())
	}
	return summaries, nil
}

func (s *UsuariosService) ListNotificaciones(ctx context.Context, userID string) (*usuariospb.ListNotificacionesResponse, error) {
	notifs, err := s.repo.ListNotificaciones(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &usuariospb.ListNotificacionesResponse{Notificaciones: notifs}, nil
}

func (s *UsuariosService) MarkNotificacionesRead(ctx context.Context, userID string, ids []string) error {
	return s.repo.MarkNotificacionesRead(ctx, userID, ids)
}

// tiposNotificacion es la lista cerrada de tipos aceptados.
//
// La validación vive aquí y no en el Gateway porque este servicio es el dueño
// de la tabla: si mañana otro emisor escribe un tipo inventado, el frontend lo
// descartaría en silencio al filtrar por perfil y el fallo sería invisible.
var tiposNotificacion = map[string]bool{
	"compra":          true,
	"inscripcion":     true,
	"nuevo_alumno":    true,
	"mensaje":         true,
	"llamada_perdida": true,
	"foro_respuesta":  true,
}

func (s *UsuariosService) CreateNotificacion(ctx context.Context, req *usuariospb.CreateNotificacionRequest) (*usuariospb.CreateNotificacionResponse, error) {
	if req.UserId == "" {
		return nil, ErrDatosNotificacion
	}
	if !tiposNotificacion[req.Tipo] {
		return nil, fmt.Errorf("%w: %q", ErrTipoNotificacion, req.Tipo)
	}
	if strings.TrimSpace(req.Titulo) == "" {
		return nil, ErrDatosNotificacion
	}

	id, creada, err := s.repo.CreateNotificacion(ctx, req)
	if err != nil {
		return nil, err
	}
	return &usuariospb.CreateNotificacionResponse{Id: id, Creada: creada}, nil
}
