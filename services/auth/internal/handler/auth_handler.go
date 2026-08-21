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
	return toAuthResponse(result), nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.AuthResponse, error) {
	result, err := h.svc.Login(ctx, req.Email, req.Password, req.RecaptchaToken)
	if err != nil {
		return nil, mapError(err, "Login")
	}
	return toAuthResponse(result), nil
}

// VerifyEmail confirma el código de 6 dígitos y devuelve el JWT definitivo.
func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authpb.VerifyEmailRequest) (*authpb.AuthResponse, error) {
	result, err := h.svc.VerifyEmail(ctx, req.Email, req.Code)
	if err != nil {
		return nil, mapError(err, "VerifyEmail")
	}
	return toAuthResponse(result), nil
}

// ResendVerificationCode emite un código nuevo respetando el cooldown.
func (h *AuthHandler) ResendVerificationCode(ctx context.Context, req *authpb.ResendVerificationRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.ResendVerificationCode(ctx, req.Email); err != nil {
		return nil, mapError(err, "ResendVerificationCode")
	}
	return &authpb.EmptyResponse{}, nil
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

func (h *AuthHandler) SolicitarCambioPassword(ctx context.Context, req *authpb.SolicitarCambioPasswordRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.SolicitarCambioPassword(ctx, req.UserId); err != nil {
		return nil, mapError(err, "SolicitarCambioPassword")
	}
	return &authpb.EmptyResponse{}, nil
}

func (h *AuthHandler) CambiarPasswordConOTP(ctx context.Context, req *authpb.CambiarPasswordConOTPRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.CambiarPasswordConOTP(ctx, req.UserId, req.Codigo, req.NuevaPassword); err != nil {
		return nil, mapError(err, "CambiarPasswordConOTP")
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
	// FailedPrecondition (no Unauthenticated): las credenciales son correctas,
	// falta un paso previo. El Gateway lo traduce a 403 + código accionable en
	// lugar de 401, que dispararía el logout automático del frontend.
	case errors.Is(err, service.ErrEmailNotVerified):
		return status.Error(codes.FailedPrecondition, "email_not_verified")
	case errors.Is(err, service.ErrCodeInvalid):
		return status.Error(codes.InvalidArgument, "el código es incorrecto")
	case errors.Is(err, service.ErrCodeExpired):
		return status.Error(codes.DeadlineExceeded, "el código expiró, solicita uno nuevo")
	case errors.Is(err, service.ErrTooManyAttempts):
		return status.Error(codes.ResourceExhausted, "demasiados intentos, solicita un código nuevo")
	case errors.Is(err, service.ErrResendTooSoon):
		return status.Error(codes.Unavailable, "espera unos segundos antes de solicitar otro código")
	// InvalidArgument y no PermissionDenied: no es que le falten permisos, es
	// que la petición llegó sin un dato obligatorio.
	case errors.Is(err, service.ErrAvisoNoAceptado):
		return status.Error(codes.InvalidArgument, "debes aceptar el aviso de privacidad")
	default:
		slog.Error("unhandled error", "op", op, "error", err)
		return status.Error(codes.Internal, "error interno del servidor")
	}
}

// userToProto convierte el modelo de dominio al mensaje proto.
func userToProto(u *model.User) *authpb.UserProfile {
	if u == nil {
		return nil
	}
	return &authpb.UserProfile{
		Id:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// toAuthResponse arma la respuesta común de Register/Login/VerifyEmail.
func toAuthResponse(r *service.LoginResult) *authpb.AuthResponse {
	if r == nil {
		return &authpb.AuthResponse{}
	}
	return &authpb.AuthResponse{
		Token:                r.Token,
		User:                 userToProto(r.User),
		RequiresVerification: r.RequiresVerification,
	}
}

func (h *AuthHandler) AceptarAviso(ctx context.Context, req *authpb.AceptarAvisoRequest) (*authpb.EmptyResponse, error) {
	if err := h.svc.AceptarAviso(ctx, req.UserId, req.Version); err != nil {
		return nil, mapError(err, "AceptarAviso")
	}
	return &authpb.EmptyResponse{}, nil
}
