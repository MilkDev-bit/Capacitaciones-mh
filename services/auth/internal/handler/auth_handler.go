// Package handler implementa el servidor gRPC del auth service.
// Capa de presentación: traduce mensajes proto ↔ tipos de dominio del service.
// No contiene lógica de negocio.
package handler

import (
	"context"
	"errors"
	"log/slog"

	authpb "Prueba-Go/gen/auth"
	"Prueba-Go/services/auth/internal/model"
	"Prueba-Go/services/auth/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler implementa authpb.AuthServiceServer.
type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// ── RPCs ──────────────────────────────────────────────────────────────────────

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthResponse, error) {
	result, err := h.svc.Register(ctx, service.RegisterInput{
		Name:           req.Name,
		Email:          req.Email,
		Password:       req.Password,
		Role:           req.Role,
		RecaptchaToken: req.RecaptchaToken,
	})
	if err != nil {
		return nil, mapError(err, "Register")
	}
	return &authpb.AuthResponse{Token: result.Token, User: userToProto(result.User)}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.AuthResponse, error) {
	result, err := h.svc.Login(ctx, req.Email, req.Password, req.RecaptchaToken)
	if err != nil {
		return nil, mapError(err, "Login")
	}
	return &authpb.AuthResponse{Token: result.Token, User: userToProto(result.User)}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.UserClaims, error) {
	claims, err := h.svc.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &authpb.UserClaims{
		UserId:       claims.UserID,
		Email:        claims.Email,
		Role:         claims.Role,
		TokenVersion: int32(claims.TokenVersion),
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.Logout(ctx, req.UserId); err != nil {
		slog.Error("Logout", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "error al cerrar sesión")
	}
	return &authpb.EmptyResponse{}, nil
}

func (h *AuthHandler) ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.ForgotPassword(ctx, req.Email); err != nil {
		slog.Error("ForgotPassword", "email", req.Email, "error", err)
		return nil, status.Error(codes.Internal, "error al procesar la solicitud")
	}
	return &authpb.EmptyResponse{}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.ResetPassword(ctx, req.ResetToken, req.NewPassword); err != nil {
		return nil, mapError(err, "ResetPassword")
	}
	return &authpb.EmptyResponse{}, nil
}

func (h *AuthHandler) RevokeUserSessions(ctx context.Context, req *authpb.RevokeRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.RevokeUserSessions(ctx, req.UserId); err != nil {
		slog.Error("RevokeUserSessions", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "error al revocar sesiones")
	}
	return &authpb.EmptyResponse{}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// mapError convierte errores de dominio en códigos gRPC correctos.
func mapError(err error, op string) error {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "credenciales inválidas")
	case errors.Is(err, service.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, "el email ya está registrado")
	case errors.Is(err, service.ErrInvalidRecaptcha):
		return status.Error(codes.InvalidArgument, "verificación de reCAPTCHA fallida")
	case errors.Is(err, service.ErrTokenInvalid):
		return status.Error(codes.Unauthenticated, "token inválido o expirado")
	case errors.Is(err, service.ErrTokenRevoked):
		return status.Error(codes.Unauthenticated, "sesión revocada")
	default:
		slog.Error("unhandled error", "op", op, "error", err)
		return status.Error(codes.Internal, "error interno del servidor")
	}
}

// userToProto convierte el modelo de dominio al mensaje proto.
func userToProto(u *model.User) *authpb.UserProfile {
	return &authpb.UserProfile{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
