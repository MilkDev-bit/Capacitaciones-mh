package config

import (
	"fmt"
	"log/slog"
	"net/url"
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
	MensajesAddr  string

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

	// Resend — proveedor de correo transaccional.
	ResendAPIKey  string
	ResendFrom    string
	ResendReplyTo string

	// SMTP — DEPRECADO, sustituido por Resend. Se conserva para no romper
	// despliegues que aún inyectan estas variables.
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	AppURL  string
	AppName string

	// Jitsi — videollamadas. El servidor se autohospeda con autenticación por
	// JWT: sin token firmado con JitsiAppSecret, Prosody rechaza la entrada a
	// la sala aunque se conozca su nombre.
	JitsiDomain    string // dominio público del servidor Jitsi
	JitsiAppID     string // JWT_APP_ID de Prosody (claim iss/aud)
	JitsiAppSecret string // JWT_APP_SECRET — nunca se codifica en el binario

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
		MensajesAddr:  getEnvOr("MENSAJES_ADDR", "mensajes-service:50057"),

		JWTSecret: requireEnv("JWT_SECRET"),

		AllowedOrigins: parseOrigins(os.Getenv("ALLOWED_ORIGIN")),

		R2Bucket:    os.Getenv("R2_BUCKET"),
		R2Endpoint:  os.Getenv("R2_ENDPOINT"),
		R2AccessKey: getEnvAny("R2_ACCESS_KEY", "R2_ACCESS_KEY_ID"),
		R2SecretKey: getEnvAny("R2_SECRET_KEY", "R2_SECRET_ACCESS_KEY"),
		R2PublicURL: os.Getenv("R2_PUBLIC_URL"),

		ResendAPIKey:  os.Getenv("RESEND_API_KEY"),
		ResendFrom:    getEnvAny("RESEND_FROM", "SMTP_FROM"),
		ResendReplyTo: os.Getenv("RESEND_REPLY_TO"),

		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: getEnvOr("SMTP_PORT", "587"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),
		// APP_URL es la URL PÚBLICA DEL FRONTEND, no la del gateway. De aquí
		// salen los enlaces de los correos ("Activar mi acceso" → /unirse/:codigo,
		// reset de contraseña, retorno de Stripe). Apuntarla al puerto del API
		// (8080) generaba enlaces muertos en producción.
		AppURL:  normalizeOrigin(getEnvOr("APP_URL", "http://localhost:5173")),
		AppName: getEnvOr("APP_NAME", "Capacitaciones MH"),

		JitsiDomain:    strings.TrimSpace(getEnvOr("JITSI_DOMAIN", "localhost:8443")),
		JitsiAppID:     getEnvOr("JITSI_APP_ID", "capacitaciones"),
		JitsiAppSecret: os.Getenv("JITSI_APP_SECRET"),

		GinMode:            os.Getenv("GIN_MODE"),
		RailwayEnvironment: os.Getenv("RAILWAY_ENVIRONMENT"),
		LogLevel:           getEnvOr("LOG_LEVEL", "info"),
	}
	warnIfAppURLLooksLikeAPI(C.AppURL, C.Port)
	if C.JitsiAppSecret == "" {
		slog.Warn("JITSI_APP_SECRET vacía — las videollamadas quedarán deshabilitadas")
	}
	// Se avisa al arrancar y no la primera vez que alguien termina un curso.
	//
	// Sin Gotenberg no hay conversión a PDF y la emisión se aborta —no se cae a
	// .docx a propósito—, así que el síntoma sería un alumno sin constancia
	// varios días después del despliegue, sin relación aparente con la causa.
	if strings.TrimSpace(os.Getenv("GOTENBERG_URL")) == "" {
		slog.Warn("GOTENBERG_URL vacía — no se podrán emitir constancias DC-3 (se entregan en PDF)")
	}
	return &C
}

// JitsiSubject es el claim `sub` que espera el módulo token_verification de
// Prosody: el dominio del servidor Jitsi sin puerto. Con un `sub` distinto al
// que Prosody tiene configurado, el token se rechaza sin explicación visible
// en el navegador, así que se deriva del dominio en vez de pedir otra
// variable de entorno que pueda quedar desincronizada.
func (c *Config) JitsiSubject() string {
	dominio := c.JitsiDomain
	if i := strings.IndexByte(dominio, ':'); i > 0 {
		dominio = dominio[:i]
	}
	return dominio
}

// normalizeOrigin deja APP_URL como scheme://host, descartando path, query y
// fragment. Sin esto, un APP_URL con path (".../login") produce enlaces
// duplicados del tipo ".../login/unirse/<codigo>".
func normalizeOrigin(raw string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}
	return u.Scheme + "://" + u.Host
}

// warnIfAppURLLooksLikeAPI avisa cuando APP_URL apunta al propio gateway.
// Es un error de configuración silencioso y caro: los correos salen con
// enlaces que el usuario no puede abrir y no hay forma de detectarlo desde
// los logs de la petición. No es fatal para no tumbar despliegues en curso.
func warnIfAppURLLooksLikeAPI(appURL, port string) {
	if appURL == "" {
		slog.Warn("APP_URL vacía — los correos saldrán sin enlaces ni logo")
		return
	}
	if strings.HasSuffix(appURL, ":"+port) {
		slog.Warn("APP_URL apunta al puerto del gateway; debe ser la URL pública del FRONTEND",
			"app_url", appURL,
			"efecto", "los botones de los correos (activar acceso, reset de contraseña) abrirán el API en lugar del sitio")
	}
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

// getEnvAny returns the value of the first non-empty env var among the provided keys.
func getEnvAny(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
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
