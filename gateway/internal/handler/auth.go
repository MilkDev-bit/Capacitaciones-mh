package handler

import (
	"log/slog"
	"net/http"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/middleware"
	authpb "Prueba-Go/gen/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler traduce peticiones HTTP ↔ RPC del auth service.
type AuthHandler struct {
	c   *clients.Clients
	cfg *config.Config
}

func NewAuthHandler(c *clients.Clients, cfg *config.Config) *AuthHandler {
	return &AuthHandler{c: c, cfg: cfg}
}

// POST /api/register
func (h *AuthHandler) Register(ctx *gin.Context) {
	var body struct {
		Name           string `json:"name"            binding:"required"`
		Email          string `json:"email"           binding:"required,email"`
		Password       string `json:"password"        binding:"required,min=8"`
		Role           string `json:"role"`
		RecaptchaToken string `json:"recaptchaToken"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.c.Auth.Register(ctx.Request.Context(), &authpb.RegisterRequest{
		Name:           body.Name,
		Email:          body.Email,
		Password:       body.Password,
		Role:           body.Role,
		RecaptchaToken: body.RecaptchaToken,
	})
	if err != nil {
		h.handleGRPCError(ctx, err)
		return
	}

	h.setAuthCookie(ctx, resp.Token)
	ctx.JSON(http.StatusCreated, gin.H{"user": resp.User})
}

// POST /api/login
func (h *AuthHandler) Login(ctx *gin.Context) {
	var body struct {
		Email          string `json:"email"    binding:"required,email"`
		Password       string `json:"password" binding:"required"`
		RecaptchaToken string `json:"recaptchaToken"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.c.Auth.Login(ctx.Request.Context(), &authpb.LoginRequest{
		Email:          body.Email,
		Password:       body.Password,
		RecaptchaToken: body.RecaptchaToken,
	})
	if err != nil {
		h.handleGRPCError(ctx, err)
		return
	}

	h.setAuthCookie(ctx, resp.Token)
	ctx.JSON(http.StatusOK, gin.H{"user": resp.User})
}

// POST /api/logout
func (h *AuthHandler) Logout(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxUserID)
	if userID != "" {
		_, _ = h.c.Auth.Logout(ctx.Request.Context(), &authpb.LogoutRequest{UserId: userID})
	}
	h.clearAuthCookie(ctx)
	ctx.JSON(http.StatusOK, gin.H{"message": "sesión cerrada"})
}

// POST /api/forgot-password
func (h *AuthHandler) ForgotPassword(ctx *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, _ = h.c.Auth.ForgotPassword(ctx.Request.Context(), &authpb.ForgotPasswordRequest{Email: body.Email})
	// Respuesta genérica — no revelamos si el email existe.
	ctx.JSON(http.StatusOK, gin.H{"message": "si el email existe, recibirás un correo"})
}

// POST /api/reset-password
func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	var body struct {
		Token    string `json:"token"    binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.c.Auth.ResetPassword(ctx.Request.Context(), &authpb.ResetPasswordRequest{
		ResetToken:  body.Token,
		NewPassword: body.Password,
	})
	if err != nil {
		h.handleGRPCError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "contraseña actualizada"})
}

// ── Cookie helpers ────────────────────────────────────────────────────────────

func (h *AuthHandler) setAuthCookie(ctx *gin.Context, token string) {
	secure := h.cfg.RailwayEnvironment != ""
	maxAge := 30 * 24 * int(time.Hour.Seconds()) // 30 días
	ctx.SetCookie("auth_token", token, maxAge, "/", "", secure, true)
}

func (h *AuthHandler) clearAuthCookie(ctx *gin.Context) {
	secure := h.cfg.RailwayEnvironment != ""
	ctx.SetCookie("auth_token", "", -1, "/", "", secure, true)
}

// ── Error mapping ─────────────────────────────────────────────────────────────

func (h *AuthHandler) handleGRPCError(ctx *gin.Context, err error) {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.AlreadyExists:
		ctx.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.NotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	default:
		slog.Error("gRPC error", "code", st.Code(), "message", st.Message(), "path", ctx.FullPath())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
	}
}
