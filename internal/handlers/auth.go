package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}
	if !verifyRecaptcha(req.RecaptchaToken) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verificación de seguridad fallida"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	role := "user"
	if req.Role == "instructor" {
		role = "instructor"
	}
	var id string
	err = db.DB.QueryRow(
		`INSERT INTO users(name,email,password_hash,role) VALUES($1,$2,$3,$4) RETURNING id`,
		req.Name, req.Email, string(hash), role,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "el email ya está registrado"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}
	if !verifyRecaptcha(req.RecaptchaToken) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verificación de seguridad fallida"})
		return
	}
	// dummyHash: hash bcrypt estático usado para normalizar el tiempo de respuesta
	// cuando el usuario no existe, previniendo enumeración por timing.
	const dummyHash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK/6xVHMa"

	var user models.User
	err := db.DB.QueryRow(
		`SELECT id, name, email, password_hash, role, token_version FROM users WHERE email=$1`, req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.TokenVersion)
	if err != nil {
		// Ejecutar bcrypt sobre hash dummy para equiparar tiempo con un login fallido por contraseña
		bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(req.Password)) //nolint:errcheck
		slog.Warn("login fallido: usuario no encontrado", "ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales incorrectas"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		slog.Warn("login fallido: contraseña incorrecta", "ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales incorrectas"})
		return
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"ver":  user.TokenVersion,
		"iat":  now.Unix(),
		"exp":  now.Add(24 * time.Hour).Unix(),
	})
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		slog.Error("JWT_SECRET no configurado")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error de configuración del servidor"})
		return
	}
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generando token"})
		return
	}
	// Determinar si estamos en producción (Railway) para usar Secure cookie
	secure := os.Getenv("RAILWAY_ENVIRONMENT") != ""
	c.SetCookie("auth_token", signed, 24*60*60, "/", "", secure, true)
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
	})
}

func Logout(c *gin.Context) {
	secure := os.Getenv("RAILWAY_ENVIRONMENT") != ""
	// MaxAge=-1 hace que el navegador elimine la cookie inmediatamente
	c.SetCookie("auth_token", "", -1, "/", "", secure, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "correo inválido"})
		return
	}

	var userID string
	err := db.DB.QueryRow(`SELECT id FROM users WHERE email=$1`, req.Email).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	code := fmt.Sprintf("%s-%s", generateCode(3), generateCode(3))
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	db.DB.Exec(`DELETE FROM password_resets WHERE email=$1`, req.Email)
	_, err = db.DB.Exec(
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
	err := db.DB.QueryRow(
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

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	defer tx.Rollback()

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
