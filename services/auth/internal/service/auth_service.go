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
	Name         string
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
		Name:         claims.Name,
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
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	TokenVersion int    `json:"tv"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateToken(u *model.User) (string, error) {
	expiry := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	claims := jwtClaims{
		UserID:       u.ID,
		Name:         u.Name,
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
	body := buildResetEmailHTML(s.cfg.AppURL, s.cfg.AppName, name, link)

	msg := strings.Join([]string{
		"From: " + s.cfg.SMTPFrom,
		"To: " + email,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	_ = sendSMTP(addr, s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPFrom, []string{email}, []byte(msg))
}

func buildResetEmailHTML(appURL, appName, name, link string) string {
	logoURL := appURL + "/logo-capacitaciones.png"
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:48px 0">
    <tr><td align="center">
      <table width="560" cellpadding="0" cellspacing="0" style="max-width:560px;width:100%%">

        <!-- Header -->
        <tr><td style="background:#1c1d1f;border-radius:16px 16px 0 0;padding:32px 40px;text-align:center">
          <img src="%s" width="60" height="60" alt="%s" style="display:block;margin:0 auto 14px;border-radius:12px" />
          <h1 style="margin:0;color:#ffffff;font-size:22px;font-weight:800;letter-spacing:-0.4px">%s</h1>
        </td></tr>

        <!-- Body -->
        <tr><td style="background:#ffffff;padding:40px 40px 32px">
          <h2 style="margin:0 0 8px;font-size:20px;font-weight:800;color:#111827">
            Restablecer contraseña
          </h2>
          <p style="margin:0 0 6px;color:#6b7280;font-size:15px">
            Hola, <strong style="color:#111827">%s</strong>
          </p>
          <p style="margin:0 0 28px;color:#6b7280;font-size:15px;line-height:1.65">
            Recibimos una solicitud para restablecer la contraseña de tu cuenta.
            Haz clic en el botón de abajo para continuar.
            <strong style="color:#374151">El enlace expira en 1 hora.</strong>
          </p>

          <!-- CTA Button -->
          <table width="100%%" cellpadding="0" cellspacing="0">
            <tr><td align="center" style="padding:4px 0 28px">
              <a href="%s"
                style="display:inline-block;background:#f97316;color:#ffffff;font-size:16px;font-weight:700;text-decoration:none;padding:15px 40px;border-radius:12px;letter-spacing:-0.1px">
                Restablecer mi contraseña →
              </a>
            </td></tr>
          </table>

          <!-- Alt link -->
          <p style="margin:0 0 24px;font-size:13px;color:#9ca3af;line-height:1.6">
            Si el botón no funciona, copia y pega este enlace en tu navegador:<br>
            <a href="%s" style="color:#f97316;word-break:break-all;font-size:12px">%s</a>
          </p>

          <hr style="border:none;border-top:1px solid #f3f4f6;margin:0 0 24px">

          <p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
            Si no solicitaste este cambio, puedes ignorar este mensaje de forma segura.
            Tu contraseña permanecerá sin cambios.
          </p>
        </td></tr>

        <!-- Footer -->
        <tr><td style="background:#f9fafb;border-radius:0 0 16px 16px;padding:20px 40px;text-align:center;border-top:1px solid #f3f4f6">
          <p style="margin:0;font-size:12px;color:#9ca3af">
            © %s &nbsp;·&nbsp; Este correo fue generado automáticamente, no respondas a este mensaje.
          </p>
        </td></tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`, logoURL, appName, appName, name, link, link, link, appName)
}
