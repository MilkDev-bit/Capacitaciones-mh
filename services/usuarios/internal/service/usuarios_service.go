// Package service contiene la lógica de negocio del usuarios service.
// No conoce ni HTTP ni gRPC — solo opera con tipos de dominio.
package service

import (
	"context"
	"fmt"

	usuariospb "Prueba-Go/gen/usuarios"
	"Prueba-Go/services/usuarios/internal/repository"
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

func (s *UsuariosService) SearchUsers(ctx context.Context, query string, limit int) (*usuariospb.SearchUsersResponse, error) {
	users, err := s.repo.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	var summaries []*usuariospb.UserSummary
	for _, u := range users {
		summaries = append(summaries, u.ToSummaryProto())
	}
	return &usuariospb.SearchUsersResponse{Users: summaries}, nil
}
