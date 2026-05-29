// Package service contiene la lógica de negocio de la aplicación.
// Los servicios no saben nada de HTTP (gin, cookies, status codes).
// Reciben datos limpios y devuelven resultados o errores de dominio.
package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"Prueba-Go/internal/config"
	"Prueba-Go/internal/models"
	"Prueba-Go/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// ── Errores de dominio exportados ────────────────────────────────────────────

// ErrInvalidCredentials se devuelve cuando el email no existe o la contraseña no coincide.
var ErrInvalidCredentials = errors.New("credenciales incorrectas")

// ErrEmailTaken se devuelve cuando ya existe un usuario con ese email (clave única).
var ErrEmailTaken = errors.New("el email ya está registrado")

// dummyHash se usa durante el login para equiparar el tiempo de respuesta cuando
// el usuario no existe, previniendo ataques de enumeración por timing.
const dummyHash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK/6xVHMa"

// ── AuthService ───────────────────────────────────────────────────────────────

// AuthService encapsula las reglas de negocio de autenticación.
// Recibe un UserRepository como dependencia (inyección), lo que permite
// reemplazarlo por un mock en los tests sin necesitar una base de datos.
type AuthService struct {
	users repository.UserRepository
}

// NewAuthService construye el servicio con el repositorio de usuarios inyectado.
func NewAuthService(users repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

// ── Register ─────────────────────────────────────────────────────────────────

// RegisterInput contiene los datos validados necesarios para registrar un usuario.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Role     string // "user" | "instructor" — cualquier otro valor resulta en "user"
}

// Register crea un nuevo usuario. Devuelve el UUID generado o un error de dominio.
// Errores posibles: ErrEmailTaken, errores internos.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Register: bcrypt failed", "error", err)
		return "", err
	}

	role := "user"
	if in.Role == "instructor" {
		role = "instructor"
	}

	id, err := s.users.Create(ctx, in.Name, in.Email, string(hash), role)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", ErrEmailTaken
		}
		slog.Error("Register: error creando usuario", "error", err)
		return "", err
	}
	return id, nil
}

// ── Login ─────────────────────────────────────────────────────────────────────

// LoginResult contiene el JWT firmado y el perfil del usuario autenticado.
type LoginResult struct {
	Token string
	User  *models.User
}

// Login autentica al usuario con email + contraseña y devuelve un JWT firmado.
// Errores posibles: ErrInvalidCredentials, errores internos.
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Ejecutar bcrypt sobre hash dummy para equiparar tiempo de respuesta
			bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password)) //nolint:errcheck
			slog.Warn("login fallido: usuario no encontrado", "email", email)
			return nil, ErrInvalidCredentials
		}
		slog.Error("Login: error consultando usuario", "error", err)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		slog.Warn("login fallido: contraseña incorrecta", "email", email)
		return nil, ErrInvalidCredentials
	}

	token, err := s.signJWT(user)
	if err != nil {
		slog.Error("Login: error firmando JWT", "error", err)
		return nil, err
	}

	return &LoginResult{Token: token, User: user}, nil
}

// signJWT genera y firma el JWT para el usuario dado.
func (s *AuthService) signJWT(u *models.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"role": u.Role,
		"ver":  u.TokenVersion,
		"iat":  now.Unix(),
		"exp":  now.Add(config.C.JWTExpiry).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.C.JWTSecret)
}
