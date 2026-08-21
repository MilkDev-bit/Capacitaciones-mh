package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/middleware"
	authpb "Prueba-Go/gen/auth"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
		// Versión del aviso de privacidad que aceptó. Vacía hace fallar el alta
		// en el servicio: el registro es donde empieza el tratamiento de datos.
		AvisoVersion string `json:"aviso_version"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		// El error crudo del validador ("Key: 'Email' Error:Field validation
		// for 'Email' failed on the 'email' tag") se filtraba tal cual al toast
		// del frontend. Se traduce a un mensaje que el usuario pueda accionar.
		ctx.JSON(http.StatusBadRequest, gin.H{"error": registerValidationMessage(err)})
		return
	}

	resp, err := h.c.Auth.Register(ctx.Request.Context(), &authpb.RegisterRequest{
		Name:           body.Name,
		Email:          body.Email,
		Password:       body.Password,
		Role:           body.Role,
		RecaptchaToken: body.RecaptchaToken,
		AvisoVersion:   body.AvisoVersion,
	})
	if err != nil {
		h.handleGRPCError(ctx, err)
		return
	}

	// El alta ya no inicia sesión: sin correo verificado no hay cookie ni JWT.
	if resp.RequiresVerification {
		ctx.JSON(http.StatusCreated, gin.H{
			"user":                  resp.User,
			"requires_verification": true,
			"email":                 body.Email,
			"message":               "te enviamos un código de 6 dígitos para confirmar tu correo",
		})
		return
	}

	h.setAuthCookie(ctx, resp.Token)
	ctx.JSON(http.StatusCreated, gin.H{"user": resp.User})
}

// POST /api/verify-email
func (h *AuthHandler) VerifyEmail(ctx *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code"  binding:"required,len=6,numeric"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ingresa el código de 6 dígitos"})
		return
	}

	resp, err := h.c.Auth.VerifyEmail(ctx.Request.Context(), &authpb.VerifyEmailRequest{
		Email: body.Email,
		Code:  body.Code,
	})
	if err != nil {
		h.handleGRPCError(ctx, err)
		return
	}

	// Verificar equivale a iniciar sesión: el usuario entra directo.
	h.setAuthCookie(ctx, resp.Token)
	ctx.JSON(http.StatusOK, gin.H{"user": resp.User})
}

// POST /api/resend-verification
func (h *AuthHandler) ResendVerification(ctx *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.c.Auth.ResendVerificationCode(ctx.Request.Context(), &authpb.ResendVerificationRequest{
		Email: body.Email,
	})
	if err != nil {
		h.handleGRPCError(ctx, err)
		return
	}
	// Respuesta genérica: no revelamos si el correo existe o ya está verificado.
	ctx.JSON(http.StatusOK, gin.H{"message": "si la cuenta existe y está pendiente, enviamos un código nuevo"})
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

// registerValidationMessage traduce los errores del validador de Gin a un
// texto en español apto para mostrarse directamente al usuario.
func registerValidationMessage(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return "revisa los datos del formulario"
	}
	for _, fe := range ve {
		switch fe.Field() {
		case "Email":
			return "el correo electrónico no es válido"
		case "Password":
			return "la contraseña debe tener al menos 8 caracteres"
		case "Name":
			return "escribe tu nombre completo"
		case "Code":
			return "ingresa el código de 6 dígitos"
		}
	}
	return "revisa los datos del formulario"
}

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

	// 403 y NO 401: el interceptor de axios trata cualquier 401 como sesión
	// expirada y expulsa al usuario. Aquí queremos redirigirlo a verificación.
	case codes.FailedPrecondition:
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "debes verificar tu correo antes de iniciar sesión",
			"code":  "email_not_verified",
		})
	case codes.DeadlineExceeded:
		ctx.JSON(http.StatusGone, gin.H{"error": st.Message(), "code": "code_expired"})
	case codes.ResourceExhausted:
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": st.Message(), "code": "too_many_attempts"})
	case codes.Unavailable:
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": st.Message(), "code": "resend_cooldown"})
	default:
		slog.Error("gRPC error", "code", st.Code(), "message", st.Message(), "path", ctx.FullPath())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
	}
}

// POST /api/perfil/password/codigo
//
// Envía un código de un solo uso al correo de quien tiene la sesión abierta.
func (h *AuthHandler) SolicitarCambioPassword(ctx *gin.Context) {
	if _, err := h.c.Auth.SolicitarCambioPassword(ctx.Request.Context(), &authpb.SolicitarCambioPasswordRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	// No se devuelve el correo de destino, ni siquiera enmascarado: el frontend
	// ya lo tiene del perfil, y repetirlo aquí lo pondría en una respuesta más
	// de las que registrar o interceptar.
	ctx.JSON(http.StatusOK, gin.H{"enviado": true})
}

// POST /api/perfil/password
//
// Cambia la contraseña. Exige el código enviado al correo: sin él, una sesión
// robada bastaría para tomar la cuenta, y hasta ahora el perfil ni siquiera
// comprobaba nada —de hecho descartaba el campo y decía que había guardado—.
func (h *AuthHandler) CambiarPassword(ctx *gin.Context) {
	var body struct {
		Codigo   string `json:"codigo"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "revisa los datos"})
		return
	}
	if strings.TrimSpace(body.Codigo) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ingresa el código que enviamos a tu correo"})
		return
	}

	if _, err := h.c.Auth.CambiarPasswordConOTP(ctx.Request.Context(), &authpb.CambiarPasswordConOTPRequest{
		UserId:        ctx.GetString(middleware.CtxUserID),
		Codigo:        body.Codigo,
		NuevaPassword: body.Password,
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	// La sesión actual también queda revocada: es lo que hace que el cambio
	// sirva para expulsar a quien hubiera entrado con la contraseña anterior.
	// Se limpia la cookie para que el frontend no siga creyendo que hay sesión.
	h.clearAuthCookie(ctx)
	ctx.JSON(http.StatusOK, gin.H{"cambiada": true})
}

// POST /api/perfil/aviso
//
// Deja constancia de que este usuario aceptó una versión del aviso. Lo usan las
// cuentas creadas antes de que el consentimiento se registrara y las que ven
// una versión nueva del texto.
func (h *AuthHandler) AceptarAviso(ctx *gin.Context) {
	var body struct {
		Version string `json:"version"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "revisa los datos"})
		return
	}
	if _, err := h.c.Auth.AceptarAviso(ctx.Request.Context(), &authpb.AceptarAvisoRequest{
		UserId:  ctx.GetString(middleware.CtxUserID),
		Version: body.Version,
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"aceptado": true})
}
