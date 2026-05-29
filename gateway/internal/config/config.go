package config

import (
	"fmt"
	"os"
	"strings"
)

// Config centraliza todas las variables de entorno del API Gateway.
type Config struct {
	// HTTP
	Port string

	// gRPC — dirección de cada microservicio
	AuthAddr      string
	UsuariosAddr  string
	CursosAddr    string
	LeccionesAddr string
	ExamenesAddr  string
	ForosAddr     string

	// JWT (validación local de tokens en el Gateway)
	JWTSecret string

	// CORS
	AllowedOrigins []string

	// Cloudflare R2 / S3-compatible (para subida de archivos)
	R2Bucket    string
	R2Endpoint  string
	R2AccessKey string
	R2SecretKey string
	R2PublicURL string

	// Entorno
	GinMode            string
	RailwayEnvironment string
	LogLevel           string
}

var C Config

func Load() *Config {
	C = Config{
		Port:          getEnvOr("PORT", "8080"),
		AuthAddr:      getEnvOr("AUTH_ADDR", "auth-service:50051"),
		UsuariosAddr:  getEnvOr("USUARIOS_ADDR", "usuarios-service:50052"),
		CursosAddr:    getEnvOr("CURSOS_ADDR", "cursos-service:50053"),
		LeccionesAddr: getEnvOr("LECCIONES_ADDR", "lecciones-service:50054"),
		ExamenesAddr:  getEnvOr("EXAMENES_ADDR", "examenes-service:50055"),
		ForosAddr:     getEnvOr("FOROS_ADDR", "foros-service:50056"),

		JWTSecret: requireEnv("JWT_SECRET"),

		AllowedOrigins: parseOrigins(os.Getenv("ALLOWED_ORIGIN")),

		R2Bucket:    os.Getenv("R2_BUCKET"),
		R2Endpoint:  os.Getenv("R2_ENDPOINT"),
		R2AccessKey: os.Getenv("R2_ACCESS_KEY"),
		R2SecretKey: os.Getenv("R2_SECRET_KEY"),
		R2PublicURL: os.Getenv("R2_PUBLIC_URL"),

		GinMode:            os.Getenv("GIN_MODE"),
		RailwayEnvironment: os.Getenv("RAILWAY_ENVIRONMENT"),
		LogLevel:           getEnvOr("LOG_LEVEL", "info"),
	}
	return &C
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "[FATAL] variable de entorno requerida: %s\n", key)
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

func parseOrigins(raw string) []string {
	if raw == "" {
		return []string{"http://localhost:5173"}
	}
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(o); t != "" {
			origins = append(origins, t)
		}
	}
	return origins
}
