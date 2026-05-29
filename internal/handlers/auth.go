package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"Prueba-Go/internal/config"
	"Prueba-Go/internal/db"
	"Prueba-Go/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	RecaptchaToken string `json:"recaptcha_token"`
}

type registerRequest struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=8"`
	Role           string `json:"role"`
	RecaptchaToken string `json:"recaptcha_token"`
}

// AuthHandler agrupa los endpoints de autenticación.
// Recibe el servicio como dependencia inyectada — nunca accede a la BD directamente.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler construye el handler con el servicio de autenticación inyectado.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register POST /register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}
	if !verifyRecaptcha(req.RecaptchaToken) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verificación de seguridad fallida"})
		return
	}

	id, err := h.authSvc.Register(c.Request.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "el email ya está registrado"})
			return
		}
		slog.Error("Register: error interno", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// Login POST /login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}
	if !verifyRecaptcha(req.RecaptchaToken) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verificación de seguridad fallida"})
		return
	}

	result, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			slog.Warn("login fallido", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales incorrectas"})
			return
		}
		slog.Error("Login: error interno", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	// La cookie viaja como HttpOnly — el JWT no es accesible desde JS.
	secure := config.C.RailwayEnvironment != ""
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", result.Token, int(config.C.JWTExpiry.Seconds()), "/", "", secure, true)
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    result.User.ID,
			"name":  result.User.Name,
			"email": result.User.Email,
			"role":  result.User.Role,
		},
	})
}

// Logout POST /logout
func Logout(c *gin.Context) {
	secure := config.C.RailwayEnvironment != ""
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", "", -1, "/", "", secure, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ForgotPassword POST /forgot-password
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "correo inválido"})
		return
	}

	var userID string
	err := db.DB.QueryRowContext(c.Request.Context(), `SELECT id FROM users WHERE email=$1`, req.Email).Scan(&userID)
	if err != nil {
		// Respuesta genérica — no revelar si el email existe o no
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	code := fmt.Sprintf("%s-%s", generateCode(3), generateCode(3))
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	db.DB.ExecContext(c.Request.Context(), `DELETE FROM password_resets WHERE email=$1`, req.Email) //nolint:errcheck
	_, err = db.DB.ExecContext(
		c.Request.Context(),
		`INSERT INTO password_resets(email, code_hash, expires_at) VALUES($1,$2,$3)`,
		req.Email, string(hash), time.Now().Add(15*time.Minute),
	)
	if err != nil {
		slog.Error("ForgotPassword: error guardando token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	go func(email, token string) {
		if err := sendPasswordResetEmail(email, token); err != nil {
			slog.Error("ForgotPassword: error enviando email", "email", email, "error", err)
		}
	}(req.Email, code)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ResetPassword POST /reset-password
func ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	var resetID, codeHash string
	err := db.DB.QueryRowContext(
		c.Request.Context(),
		`SELECT id, code_hash FROM password_resets
		 WHERE email=$1 AND used=false AND expires_at > NOW()
		 ORDER BY created_at DESC LIMIT 1`,
		req.Email,
	).Scan(&resetID, &codeHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "código inválido o expirado"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(codeHash), []byte(req.Code)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "código incorrecto"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	tx, err := db.DB.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err = tx.Exec(`UPDATE users SET password_hash=$1, token_version=token_version+1 WHERE email=$2`, string(hash), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar la contraseña"})
		return
	}
	if _, err = tx.Exec(`UPDATE password_resets SET used=true WHERE id=$1`, resetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
