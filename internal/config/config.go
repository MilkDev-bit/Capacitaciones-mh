// Package config carga todas las variables de entorno en una única estructura
// al arrancar la aplicación (Fail Fast): si falta una variable crítica, el
// proceso termina de inmediato en lugar de fallar horas después en producción.
package config

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

// Config agrupa toda la configuración de la aplicación leída desde variables de entorno.
type Config struct {
	// Servidor
	Port    string
	GinMode string
	AppName string

	// Logging
	LogLevel string

	// Autenticación JWT
	JWTSecret []byte
	JWTExpiry time.Duration

	// Base de datos
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string

	// Almacenamiento (Cloudflare R2 / S3-compatible)
	R2Bucket          string
	R2PublicURL       string
	R2Endpoint        string
	R2AccessKeyID     string
	R2SecretAccessKey string
	CFAccountID       string
	CFAPIToken        string

	// CORS
	AllowedOrigins []string

	// reCAPTCHA v3
	RecaptchaSecretKey string

	// SMTP (correo)
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// Sentry
	SentryDSN string

	// Entorno (Railway, etc.)
	RailwayEnvironment string
}

// C es la instancia global de configuración. Se inicializa llamando a Load()
// al inicio de main(). El resto del código la lee directamente como config.C.
var C Config

// Load lee todas las variables de entorno y rellena C.
// Llama a os.Exit(1) si alguna variable crítica no está definida.
func Load() {
	// ── Variables críticas — fallo inmediato si faltan ──────────────────────
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET no definido — configura esta variable antes de arrancar")
		os.Exit(1)
	}

	// ── CORS ─────────────────────────────────────────────────────────────────
	allowedOriginsRaw := os.Getenv("ALLOWED_ORIGIN")
	if allowedOriginsRaw == "" {
		allowedOriginsRaw = "http://localhost:5173"
	}
	origins := strings.Split(allowedOriginsRaw, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}

	C = Config{
		// Servidor
		Port:    getEnv("PORT", "8080"),
		GinMode: os.Getenv("GIN_MODE"),
		AppName: getEnv("APP_NAME", "Capacitaciones MH"),

		// Logging
		LogLevel: os.Getenv("LOG_LEVEL"),

		// JWT
		JWTSecret: []byte(jwtSecret),
		JWTExpiry: 24 * time.Hour,

		// Base de datos
		DatabaseURL: os.Getenv("DATABASE_URL"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      getEnv("DB_NAME", "capacitaciones"),

		// Almacenamiento
		R2Bucket:          os.Getenv("R2_BUCKET"),
		R2PublicURL:       strings.TrimRight(os.Getenv("R2_PUBLIC_URL"), "/"),
		R2Endpoint:        os.Getenv("R2_ENDPOINT"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		CFAccountID:       os.Getenv("CF_ACCOUNT_ID"),
		CFAPIToken:        os.Getenv("CF_API_TOKEN"),

		// CORS
		AllowedOrigins: origins,

		// reCAPTCHA
		RecaptchaSecretKey: os.Getenv("RECAPTCHA_SECRET_KEY"),

		// SMTP
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: getEnv("SMTP_PORT", "587"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),

		// Sentry
		SentryDSN: os.Getenv("SENTRY_DSN"),

		// Entorno
		RailwayEnvironment: os.Getenv("RAILWAY_ENVIRONMENT"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
