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
