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

	// SMTP para emails de recuperación de contraseña
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
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           getEnvOr("SMTP_PORT", "587"),
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPass:           os.Getenv("SMTP_PASS"),
		SMTPFrom:           os.Getenv("SMTP_FROM"),
		AppURL:             normalizeOrigin(getEnvOr("APP_URL", "http://localhost:5173")),
		AppName:            getEnvOr("APP_NAME", "Capacitaciones"),
		LogLevel:           getEnvOr("LOG_LEVEL", "info"),
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

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
