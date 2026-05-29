package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"Prueba-Go/services/auth/internal/config"
	"Prueba-Go/services/auth/internal/model"
	"Prueba-Go/services/auth/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// ── Errores de dominio ────────────────────────────────────────────────────────

var (
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrEmailTaken         = errors.New("email ya registrado")
	ErrInvalidRecaptcha   = errors.New("verificación de reCAPTCHA fallida")
	ErrTokenInvalid       = errors.New("token inválido o expirado")
	ErrTokenRevoked       = errors.New("sesión revocada")
)

// ── DTOs ─────────────────────────────────────────────────────────────────────

type RegisterInput struct {
	Name           string
	Email          string
	Password       string
	Role           string
	RecaptchaToken string
}

type LoginResult struct {
	Token string
	User  *model.User
}

// Claims son los datos que el auth service extrae de un JWT válido.
type Claims struct {
	UserID       string
	Email        string
	Role         string
	TokenVersion int
}

// ── Service ───────────────────────────────────────────────────────────────────

// AuthService contiene la lógica de negocio de autenticación.
// No conoce nada de HTTP ni de gRPC — solo entradas/salidas de dominio.
type AuthService struct {
	users repository.UserRepository
	cfg   *config.Config
}

func NewAuthService(users repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

// Register valida reCAPTCHA, hashea la contraseña, persiste el usuario y devuelve JWT.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*LoginResult, error) {
	if s.cfg.RecaptchaSecretKey != "" {
		if err := verifyRecaptcha(in.RecaptchaToken, s.cfg.RecaptchaSecretKey); err != nil {
			return nil, ErrInvalidRecaptcha
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	role := in.Role
	if role == "" {
		role = "user"
	}

	u := &model.User{
		ID:           uuid.New().String(),
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         role,
		TokenVersion: 1,
	}

	if err := s.users.Create(ctx, u); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.generateToken(u)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResult{Token: token, User: u}, nil
}

// Login verifica credenciales y devuelve JWT + perfil.
func (s *AuthService) Login(ctx context.Context, email, password, recaptchaToken string) (*LoginResult, error) {
	if s.cfg.RecaptchaSecretKey != "" {
		// No revelamos si reCAPTCHA falló para evitar enumeración de cuentas.
		if err := verifyRecaptcha(recaptchaToken, s.cfg.RecaptchaSecretKey); err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(u)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResult{Token: token, User: u}, nil
}

// ValidateToken verifica la firma y token_version del JWT.
// El Gateway llama a este método en cada petición autenticada.
func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*Claims, error) {
	claims := &jwtClaims{}
	t, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo inesperado: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !t.Valid {
		return nil, ErrTokenInvalid
	}

	// Comprobamos token_version contra la BD para detectar revocaciones.
	u, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if u.TokenVersion != claims.TokenVersion {
		return nil, ErrTokenRevoked
	}

	return &Claims{
		UserID:       claims.UserID,
		Email:        claims.Email,
		Role:         claims.Role,
		TokenVersion: claims.TokenVersion,
	}, nil
}

// Logout incrementa token_version, invalidando todos los JWT activos del usuario.
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	return s.users.UpdateTokenVersion(ctx, userID)
}

// ForgotPassword genera un token de reset y lo envía por email.
// No revela si el email existe o no (seguridad).
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // silencioso — no revela existencia del email
		}
		return err
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}
	token := hex.EncodeToString(b)
	expiry := time.Now().Add(1 * time.Hour)

	if err := s.users.StorePasswordResetToken(ctx, u.ID, token, expiry); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	go s.sendResetEmail(u.Email, u.Name, token) // fire-and-forget
	return nil
}

// ResetPassword verifica el token de reset y actualiza la contraseña.
func (s *AuthService) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	u, err := s.users.FindByResetToken(ctx, resetToken)
	if err != nil {
		return ErrTokenInvalid
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, u.ID, string(hash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := s.users.ClearPasswordResetToken(ctx, u.ID); err != nil {
		return fmt.Errorf("clear reset token: %w", err)
	}

	// Invalidar todas las sesiones activas tras cambio de contraseña.
	return s.users.UpdateTokenVersion(ctx, u.ID)
}

// RevokeUserSessions invalida todos los JWT activos de un usuario (acción de admin).
func (s *AuthService) RevokeUserSessions(ctx context.Context, userID string) error {
	return s.users.UpdateTokenVersion(ctx, userID)
}

// ── JWT helpers ───────────────────────────────────────────────────────────────

type jwtClaims struct {
	UserID       string `json:"uid"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	TokenVersion int    `json:"tv"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateToken(u *model.User) (string, error) {
	expiry := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	claims := jwtClaims{
		UserID:       u.ID,
		Email:        u.Email,
		Role:         u.Role,
		TokenVersion: u.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

// ── reCAPTCHA v3 ──────────────────────────────────────────────────────────────

type recaptchaResponse struct {
	Success bool    `json:"success"`
	Score   float64 `json:"score"`
}

func verifyRecaptcha(token, secretKey string) error {
	if token == "" {
		return errors.New("recaptcha token vacío")
	}
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {secretKey},
		"response": {token},
	})
	if err != nil {
		return fmt.Errorf("recaptcha request: %w", err)
	}
	defer resp.Body.Close()

	var result recaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("recaptcha decode: %w", err)
	}
	if !result.Success || result.Score < 0.5 {
		return errors.New("recaptcha falló")
	}
	return nil
}

// ── Email helper ──────────────────────────────────────────────────────────────

func (s *AuthService) sendResetEmail(email, name, token string) {
	if s.cfg.SMTPHost == "" {
		return // SMTP no configurado
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.AppURL, token)
	subject := fmt.Sprintf("Restablecer contraseña — %s", s.cfg.AppName)
	body := fmt.Sprintf(
		"Hola %s,\n\nHaz clic en el siguiente enlace para restablecer tu contraseña:\n%s\n\nEste enlace expira en 1 hora.\n\nSi no solicitaste este cambio, ignora este mensaje.",
		name, link,
	)

	msg := strings.Join([]string{
		"From: " + s.cfg.SMTPFrom,
		"To: " + email,
		"Subject: " + subject,
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	// net/smtp.SendMail usa TLS/STARTTLS automáticamente.
	// Para SendGrid/Resend/Mailgun, considera una librería dedicada.
	_ = sendSMTP(addr, s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPFrom, []string{email}, []byte(msg))
}
