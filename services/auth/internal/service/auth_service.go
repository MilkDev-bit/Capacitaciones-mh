package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"Prueba-Go/pkg/mailer"
	"Prueba-Go/services/auth/internal/config"
	"Prueba-Go/services/auth/internal/model"
	"Prueba-Go/services/auth/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// uniqueViolation detecta el código SQLSTATE 23505 (unique_violation).
//
// El auth service abre la conexión con el driver `pgx` (stdlib), que devuelve
// *pgconn.PgError — NO *pq.Error. Comprobar solo *pq.Error hacía que el choque
// de correo duplicado se colara como error genérico y el usuario recibiera
// "error interno del servidor" en lugar de "el email ya está registrado".
// Se cubren ambos tipos para que la función siga siendo válida si algún
// servicio del monolito legado (que sí usa lib/pq) reutiliza este paquete.
func uniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// ── Errores de dominio ────────────────────────────────────────────────────────

var (
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrEmailTaken         = errors.New("email ya registrado")
	ErrInvalidRecaptcha   = errors.New("verificación de reCAPTCHA fallida")
	ErrTokenInvalid       = errors.New("token inválido o expirado")
	ErrTokenRevoked       = errors.New("sesión revocada")

	// Verificación de correo
	ErrEmailNotVerified = errors.New("correo no verificado")
	ErrCodeInvalid      = errors.New("código de verificación incorrecto")
	ErrCodeExpired      = errors.New("código de verificación expirado")
	ErrTooManyAttempts  = errors.New("demasiados intentos fallidos")
	ErrResendTooSoon    = errors.New("espera antes de solicitar otro código")
	ErrAvisoNoAceptado  = errors.New("debes aceptar el aviso de privacidad")
)

// maxVerificationAttempts limita el fuerza-bruta sobre un código de 6 dígitos
// (10^6 combinaciones). Al agotarse hay que pedir un código nuevo.
const maxVerificationAttempts = 5

type tvCacheItem struct {
	version   int
	expiresAt time.Time
}

var tvCache sync.Map

// ── DTOs ─────────────────────────────────────────────────────────────────────

type RegisterInput struct {
	Name           string
	Email          string
	Password       string
	Role           string
	RecaptchaToken string
	// AvisoVersion es la versión del aviso de privacidad aceptada. Vacía hace
	// fallar el alta.
	AvisoVersion string
}

type LoginResult struct {
	Token string
	User  *model.User
	// RequiresVerification indica que la cuenta existe pero aún no confirmó su
	// correo. Cuando es true, Token viene vacío a propósito.
	RequiresVerification bool
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
	mail  *mailer.Client
}

func NewAuthService(users repository.UserRepository, cfg *config.Config, mail *mailer.Client) *AuthService {
	return &AuthService{users: users, cfg: cfg, mail: mail}
}

// Register valida reCAPTCHA, hashea la contraseña, persiste el usuario y devuelve JWT.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*LoginResult, error) {
	if s.cfg.RecaptchaSecretKey != "" {
		if err := verifyRecaptcha(in.RecaptchaToken, s.cfg.RecaptchaSecretKey); err != nil {
			return nil, ErrInvalidRecaptcha
		}
	}

	// Sin aceptación no hay alta: el registro es donde empieza el tratamiento
	// de los datos, y guardarlos antes de que la persona haya visto para qué es
	// justo lo que el aviso debe impedir.
	if strings.TrimSpace(in.AvisoVersion) == "" {
		return nil, ErrAvisoNoAceptado
	}

	email := normalizeEmail(in.Email)

	// Pre-chequeo explícito: da un mensaje claro sin depender del nombre del
	// índice único y evita gastar un bcrypt de coste 12 (~250 ms) en un alta
	// que ya sabemos que va a fallar. El INSERT sigue siendo la fuente de
	// verdad ante carreras concurrentes.
	if existing, err := s.users.FindByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrEmailTaken
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find user: %w", err)
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
		ID:            uuid.New().String(),
		Name:          in.Name,
		Email:         email,
		PasswordHash:  string(hash),
		Role:          role,
		TokenVersion:  1,
		EmailVerified: false,
	}

	if err := s.users.Create(ctx, u); err != nil {
		if uniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	// La constancia se guarda en cuanto existe la fila. Si esto fallara, el alta
	// no se revierte —la cuenta ya está creada y el usuario espera su código—,
	// pero queda registrado: al entrar se le volverá a pedir, porque su versión
	// aparecerá vacía.
	if err := s.users.RegistrarAceptacionAviso(ctx, u.ID, strings.TrimSpace(in.AvisoVersion)); err != nil {
		slog.Error("Register: no se pudo registrar la aceptación del aviso",
			"user_id", u.ID, "error", err)
	}

	// No se emite JWT todavía: la cuenta no vale nada hasta confirmar el buzón.
	// Un fallo al enviar el correo no revierte el alta — el usuario puede pedir
	// un reenvío desde la pantalla de verificación.
	if err := s.issueVerificationCode(ctx, u); err != nil {
		slog.Error("Register: no se pudo enviar el código de verificación",
			"user_id", u.ID, "error", err)
	}

	return &LoginResult{User: u, RequiresVerification: true}, nil
}

// issueVerificationCode genera un código de 6 dígitos, guarda su hash y lo envía.
// El código nunca se persiste en claro: si se filtra la BD no sirve para activar
// cuentas ajenas.
func (s *AuthService) issueVerificationCode(ctx context.Context, u *model.User) error {
	code, err := generateNumericCode(6)
	if err != nil {
		return fmt.Errorf("generar código: %w", err)
	}

	ttl := time.Duration(s.cfg.EmailVerificationTTLMinutes) * time.Minute
	if err := s.users.StoreEmailVerification(ctx, u.ID, hashCode(u.ID, code), time.Now().Add(ttl)); err != nil {
		return fmt.Errorf("guardar código: %w", err)
	}

	msg := s.mail.VerificationCode(u.Name, code, s.cfg.EmailVerificationTTLMinutes)
	msg.To = []string{u.Email}
	s.mail.SendAsync(msg)
	return nil
}

// VerifyEmail valida el código de 6 dígitos y, si es correcto, activa la cuenta
// y devuelve el JWT — es el único punto donde una cuenta nueva obtiene sesión.
func (s *AuthService) VerifyEmail(ctx context.Context, email, code string) (*LoginResult, error) {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCodeInvalid // no revelamos si el correo existe
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	// Idempotencia: si ya estaba verificado devolvemos sesión en lugar de error,
	// así un doble clic en "Verificar" no bloquea al usuario.
	if u.EmailVerified {
		token, err := s.generateToken(u)
		if err != nil {
			return nil, fmt.Errorf("generate token: %w", err)
		}
		return &LoginResult{Token: token, User: u}, nil
	}

	v, err := s.users.GetEmailVerification(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("get verification: %w", err)
	}
	if v.Hash == nil || v.Expires == nil {
		return nil, ErrCodeExpired
	}
	if v.Attempts >= maxVerificationAttempts {
		return nil, ErrTooManyAttempts
	}
	if time.Now().After(*v.Expires) {
		return nil, ErrCodeExpired
	}

	// Comparación en tiempo constante: evita distinguir códigos por latencia.
	expected := hashCode(u.ID, strings.TrimSpace(code))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(*v.Hash)) != 1 {
		if err := s.users.IncrementVerificationAttempts(ctx, u.ID); err != nil {
			slog.Error("VerifyEmail: no se pudo incrementar intentos", "user_id", u.ID, "error", err)
		}
		return nil, ErrCodeInvalid
	}

	if err := s.users.MarkEmailVerified(ctx, u.ID); err != nil {
		return nil, fmt.Errorf("marcar verificado: %w", err)
	}
	u.EmailVerified = true

	token, err := s.generateToken(u)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	return &LoginResult{Token: token, User: u}, nil
}

// ResendVerificationCode emite un código nuevo respetando un cooldown.
// Devuelve nil silenciosamente si el correo no existe o ya está verificado,
// para no convertir el endpoint en un oráculo de cuentas registradas.
func (s *AuthService) ResendVerificationCode(ctx context.Context, email string) error {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("find user: %w", err)
	}
	if u.EmailVerified {
		return nil
	}

	v, err := s.users.GetEmailVerification(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("get verification: %w", err)
	}
	cooldown := time.Duration(s.cfg.EmailVerificationCooldownSec) * time.Second
	if v.SentAt != nil && time.Since(*v.SentAt) < cooldown {
		return ErrResendTooSoon
	}

	return s.issueVerificationCode(ctx, u)
}

// Login verifica credenciales y devuelve JWT + perfil.
func (s *AuthService) Login(ctx context.Context, email, password, recaptchaToken string) (*LoginResult, error) {
	if s.cfg.RecaptchaSecretKey != "" {
		// No revelamos si reCAPTCHA falló para evitar enumeración de cuentas.
		if err := verifyRecaptcha(recaptchaToken, s.cfg.RecaptchaSecretKey); err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// La verificación se comprueba DESPUÉS de la contraseña: de lo contrario el
	// endpoint revelaría qué correos están registrados sin conocer credenciales.
	if !u.EmailVerified {
		// Reenvío best-effort para que el usuario tenga un código fresco al
		// aterrizar en la pantalla de verificación. El cooldown evita el abuso.
		if err := s.ResendVerificationCode(ctx, u.Email); err != nil && !errors.Is(err, ErrResendTooSoon) {
			slog.Error("Login: fallo al reenviar código", "user_id", u.ID, "error", err)
		}
		return nil, ErrEmailNotVerified
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

	// Comprobamos token_version en caché (15s) o BD para detectar revocaciones.
	var tv int
	if item, ok := tvCache.Load(claims.UserID); ok {
		cached := item.(tvCacheItem)
		if time.Now().Before(cached.expiresAt) {
			tv = cached.version
		}
	}
	if tv == 0 {
		u, err := s.users.FindByID(ctx, claims.UserID)
		if err != nil {
			return nil, fmt.Errorf("find user: %w", err)
		}
		tv = u.TokenVersion
		tvCache.Store(claims.UserID, tvCacheItem{version: tv, expiresAt: time.Now().Add(15 * time.Second)})
	}
	if tv != claims.TokenVersion {
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
	tvCache.Delete(userID)
	return s.users.UpdateTokenVersion(ctx, userID)
}

// ForgotPassword genera un token de reset y lo envía por email.
// No revela si el email existe o no (seguridad).
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// normalizeEmail: el alta guarda el correo en minúsculas, así que buscarlo
	// tal cual lo teclea el usuario dejaba sin recuperación a quien escribiera
	// "Juan@Empresa.com".
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
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

	link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.AppURL, url.QueryEscape(token))
	msg := s.mail.PasswordResetLink(u.Name, link)
	msg.To = []string{u.Email}
	s.mail.SendAsync(msg) // fire-and-forget
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
	tvCache.Delete(u.ID)
	return s.users.UpdateTokenVersion(ctx, u.ID)
}

// RevokeUserSessions invalida todos los JWT activos de un usuario (acción de admin).
func (s *AuthService) RevokeUserSessions(ctx context.Context, userID string) error {
	tvCache.Delete(userID)
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

// ── Helpers de verificación ───────────────────────────────────────────────────

// normalizeEmail evita cuentas duplicadas por mayúsculas o espacios pegados.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// generateNumericCode produce un código decimal criptográficamente aleatorio.
// Se usa crypto/rand (no math/rand) porque el código es una credencial.
func generateNumericCode(digits int) (string, error) {
	maxVal := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", digits, n), nil
}

// hashCode deriva el hash almacenado. Se mezcla el userID para que dos usuarios
// con el mismo código no compartan hash (y una tabla precalculada no sirva).
func hashCode(userID, code string) string {
	sum := sha256.Sum256([]byte(userID + ":" + code))
	return hex.EncodeToString(sum[:])
}

// ── Cambio de contraseña desde el perfil ─────────────────────────────────────

// ttlPasswordOTP es lo que vive el código. Corto a propósito: es la ventana en
// la que alguien con acceso al buzón podría usarlo.
const ttlPasswordOTP = 10 * time.Minute

// SolicitarCambioPassword emite un código de un solo uso al correo del usuario.
//
// Este flujo NO es el de "olvidé mi contraseña": aquí hay sesión iniciada y se
// sabe quién pide el cambio. El código sirve para probar que quien está frente
// a la pantalla controla además el buzón, de modo que una sesión robada —una
// cookie filtrada, un equipo desatendido— no baste para tomar la cuenta.
//
// Antes de esto, el perfil aceptaba una contraseña nueva sin comprobar nada. En
// realidad ni siquiera la guardaba: el gateway descartaba el campo, así que el
// usuario veía "actualizado" y su contraseña seguía siendo la misma.
func (s *AuthService) SolicitarCambioPassword(ctx context.Context, userID string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	code, err := generateNumericCode(6)
	if err != nil {
		return fmt.Errorf("generar código: %w", err)
	}
	if err := s.users.StorePasswordOTP(ctx, u.ID, hashCode(u.ID, code), time.Now().Add(ttlPasswordOTP)); err != nil {
		return fmt.Errorf("guardar código: %w", err)
	}

	msg := s.mail.PasswordResetCode(code)
	msg.To = []string{u.Email}
	s.mail.SendAsync(msg)
	return nil
}

// CambiarPasswordConOTP valida el código y actualiza la contraseña.
//
// El código se comprueba SIEMPRE contra el userID de la sesión. Por eso puede
// ser de seis dígitos sin ser adivinable en la práctica: probar combinaciones
// solo ataca a esa cuenta, y a los cinco fallos el código se invalida.
func (s *AuthService) CambiarPasswordConOTP(ctx context.Context, userID, code, nueva string) error {
	if len(nueva) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}

	v, err := s.users.GetPasswordOTP(ctx, userID)
	if err != nil {
		return fmt.Errorf("get otp: %w", err)
	}
	if v.Hash == nil || v.Expira == nil {
		return ErrCodeExpired
	}
	if v.Intentos >= maxVerificationAttempts {
		return ErrTooManyAttempts
	}
	if time.Now().After(*v.Expira) {
		return ErrCodeExpired
	}

	// Comparación en tiempo constante: evita distinguir códigos por latencia.
	esperado := hashCode(userID, strings.TrimSpace(code))
	if subtle.ConstantTimeCompare([]byte(esperado), []byte(*v.Hash)) != 1 {
		if err := s.users.IncrementPasswordOTPAttempts(ctx, userID); err != nil {
			slog.Error("CambiarPasswordConOTP: no se pudo incrementar intentos", "user_id", userID, "error", err)
		}
		// Al agotar los intentos se quema el código: sin esto, seguiría siendo
		// válido y bastaría con esperar a pedir otro para reanudar las pruebas.
		if v.Intentos+1 >= maxVerificationAttempts {
			if err := s.users.ClearPasswordOTP(ctx, userID); err != nil {
				slog.Error("CambiarPasswordConOTP: no se pudo invalidar el código", "user_id", userID, "error", err)
			}
		}
		return ErrCodeInvalid
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nueva), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := s.users.ClearPasswordOTP(ctx, userID); err != nil {
		slog.Error("CambiarPasswordConOTP: no se pudo limpiar el código", "user_id", userID, "error", err)
	}

	// Se cierran TODAS las sesiones, incluida la que hizo el cambio.
	//
	// Es lo que convierte esto en una vía de recuperación: si alguien había
	// entrado con la contraseña vieja, cambiarla lo expulsa. Dejar viva la
	// sesión del atacante haría el cambio casi inútil.
	if err := s.users.UpdateTokenVersion(ctx, userID); err != nil {
		return fmt.Errorf("revocar sesiones: %w", err)
	}
	tvCache.Delete(userID)
	return nil
}

// AceptarAviso deja constancia de que el usuario aceptó una versión del aviso.
//
// Sirve para las cuentas creadas antes de que esto existiera y para cuando se
// publica una versión nueva. No comprueba cuál es la vigente a propósito: esa
// decisión vive en el frontend, que es quien muestra el texto, y validarla aquí
// obligaría a mantener el número en dos sitios que se desincronizarían.
func (s *AuthService) AceptarAviso(ctx context.Context, userID, version string) error {
	v := strings.TrimSpace(version)
	if v == "" {
		return ErrAvisoNoAceptado
	}
	return s.users.RegistrarAceptacionAviso(ctx, userID, v)
}
