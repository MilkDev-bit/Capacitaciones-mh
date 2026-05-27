package handlers

import (
	"log"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	// A01: el rol admin nunca puede asignarse por la API pública
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !verifyRecaptcha(req.RecaptchaToken) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verificación de seguridad fallida"})
		return
	}
	var user models.User
	err := db.DB.QueryRow(
		`SELECT id, name, email, password_hash, role FROM users WHERE email=$1`, req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		// A09: registrar intento fallido sin revelar si el email existe
		log.Printf("[AUTH] login fallido (usuario no encontrado): ip=%s", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales incorrectas"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// A09: registrar intento fallido
		log.Printf("[AUTH] login fallido (contraseña incorrecta): ip=%s", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales incorrectas"})
		return
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(24 * time.Hour).Unix(),
	})
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme_secret_key_32chars_long!!"
	}
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generando token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": signed,
		"user":  gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
	})
}

// ForgotPassword genera un código de 6 chars y lo envía al correo del usuario.
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "correo inválido"})
		return
	}

	// Verificar que el usuario existe (sin revelar si no existe)
	var userID string
	err := db.DB.QueryRow(`SELECT id FROM users WHERE email=$1`, req.Email).Scan(&userID)
	if err != nil {
		// No revelar si el email está registrado (anti-enumeración)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// Generar código alfanumérico de 6 chars (misma entropía que generateCode)
	code := fmt.Sprintf("%s-%s", generateCode(3), generateCode(3))
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	// Eliminar tokens previos del mismo email y guardar el nuevo
	db.DB.Exec(`DELETE FROM password_resets WHERE email=$1`, req.Email)
	_, err = db.DB.Exec(
		`INSERT INTO password_resets(email, code_hash, expires_at) VALUES($1,$2,$3)`,
		req.Email, string(hash), time.Now().Add(15*time.Minute),
	)
	if err != nil {
		log.Printf("[FORGOT] error guardando token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	if err := sendPasswordResetEmail(req.Email, code); err != nil {
		log.Printf("[FORGOT] error enviando email a %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo enviar el correo. Verifica tu dirección o contacta soporte."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ResetPassword verifica el código y actualiza la contraseña.
func ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	_, err = db.DB.Exec(`UPDATE users SET password_hash=$1 WHERE email=$2`, string(hash), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar la contraseña"})
		return
	}

	// Marcar el token como usado
	db.DB.Exec(`UPDATE password_resets SET used=true WHERE id=$1`, resetID)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
