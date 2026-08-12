package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config centraliza todas las variables de entorno del auth service.
// config.Load() llama a os.Exit(1) si falta alguna variable crítica (Fail Fast).
type Config struct {
	// gRPC
	GRPCPort string

	// Base de datos
	DatabaseURL string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// reCAPTCHA (opcional — se omite si está vacío)
	RecaptchaSecretKey string

	// Resend — proveedor de correo transaccional (sustituye a SMTP).
	ResendAPIKey  string
	ResendFrom    string
	ResendReplyTo string

	// Verificación de correo en el registro
	EmailVerificationTTLMinutes  int // vigencia del código de 6 dígitos
	EmailVerificationCooldownSec int // espera mínima entre reenvíos

	// SMTP — DEPRECADO. Se conserva solo para no romper despliegues que aún
	// inyectan estas variables; el envío real pasa por Resend.
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// URL base de la app (para construir enlaces en emails)
	AppURL  string
	AppName string

	// Observabilidad
	LogLevel string
}

// C es la instancia global cargada por Load().
var C Config

// Load lee las variables de entorno, valida las críticas y rellena C.
func Load() *Config {
	C = Config{
		GRPCPort:           getEnvOr("GRPC_PORT", "50051"),
		DatabaseURL:        requireEnv("DATABASE_URL"),
		JWTSecret:          requireEnv("JWT_SECRET"),
		JWTExpiryHours:     getEnvInt("JWT_EXPIRY_HOURS", 720),
		RecaptchaSecretKey: os.Getenv("RECAPTCHA_SECRET_KEY"),

		ResendAPIKey:  os.Getenv("RESEND_API_KEY"),
		ResendFrom:    getEnvAny("RESEND_FROM", "SMTP_FROM"),
		ResendReplyTo: os.Getenv("RESEND_REPLY_TO"),

		EmailVerificationTTLMinutes:  getEnvInt("EMAIL_VERIFICATION_TTL_MINUTES", 15),
		EmailVerificationCooldownSec: getEnvInt("EMAIL_VERIFICATION_COOLDOWN_SECONDS", 60),

		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: getEnvOr("SMTP_PORT", "587"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),
		AppURL:   normalizeOrigin(getEnvOr("APP_URL", "http://localhost:5173")),
		AppName:  getEnvOr("APP_NAME", "Capacitaciones"),
		LogLevel: getEnvOr("LOG_LEVEL", "info"),
	}
	return &C
}

// normalizeOrigin devuelve solo scheme://host del URL, descartando path, query y fragment.
// Esto garantiza que APP_URL siempre sea la raíz del dominio aunque venga con path extra.
func normalizeOrigin(raw string) string {
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}
	return u.Scheme + "://" + u.Host
}

// ── helpers ───────────────────────────────────────────────────────────────────

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "[FATAL] variable de entorno requerida no encontrada: %s\n", key)
		os.Exit(1)
	}
	return v
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvAny devuelve el primer valor no vacío entre varias claves.
// Permite migrar de SMTP_FROM a RESEND_FROM sin romper despliegues existentes.
func getEnvAny(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
