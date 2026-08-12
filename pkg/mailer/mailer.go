// Package mailer centraliza el envío de correo transaccional del sistema
// usando la API HTTP de Resend (https://resend.com/docs/api-reference).
//
// Se implementa como módulo independiente (sin dependencias externas) para que
// el gateway, el auth service y el monolito legado compartan exactamente las
// mismas plantillas y el mismo comportamiento ante fallos.
//
// Motivo del diseño: la API HTTP de Resend evita mantener conexiones SMTP
// abiertas dentro de contenedores efímeros, entrega feedback de error legible
// (a diferencia de net/smtp, que falla de forma opaca) y permite envíos por
// lote en una sola llamada — necesario para repartir accesos corporativos.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultEndpoint = "https://api.resend.com"
	defaultTimeout  = 15 * time.Second
	maxBatchSize    = 100 // límite duro de la API de Resend
)

// ErrNotConfigured se devuelve cuando falta RESEND_API_KEY. Los llamadores
// pueden tratarlo como "degradado" en desarrollo sin romper el flujo principal.
var ErrNotConfigured = errors.New("mailer: RESEND_API_KEY no configurada")

// Config agrupa la configuración del cliente. Todos los valores provienen de
// variables de entorno inyectadas vía .env — nunca se codifican en el binario.
type Config struct {
	APIKey  string // RESEND_API_KEY
	From    string // RESEND_FROM, ej. "Capacitaciones MH <no-reply@midominio.com>"
	ReplyTo string // RESEND_REPLY_TO (opcional)
	AppName string // APP_NAME
	AppURL  string // APP_URL — usado para construir enlaces y el logo
}

// Client envía correos a través de Resend.
type Client struct {
	cfg      Config
	endpoint string
	http     *http.Client
}

// Message es un correo individual listo para enviarse.
type Message struct {
	To      []string
	Subject string
	HTML    string
}

// New construye un cliente a partir de la configuración dada.
// Nunca devuelve error: si falta la API key, el cliente queda deshabilitado y
// Send registra un aviso en lugar de tumbar el servicio.
func New(cfg Config) *Client {
	cfg.AppURL = strings.TrimRight(cfg.AppURL, "/")
	if cfg.AppName == "" {
		cfg.AppName = "Capacitaciones"
	}
	if cfg.From == "" {
		cfg.From = "Capacitaciones <onboarding@resend.dev>"
	}
	return &Client{
		cfg:      cfg,
		endpoint: defaultEndpoint,
		http:     &http.Client{Timeout: defaultTimeout},
	}
}

// Enabled indica si hay credenciales para enviar correo de verdad.
func (c *Client) Enabled() bool { return c != nil && c.cfg.APIKey != "" }

// AppName expone el nombre de la app para las plantillas.
func (c *Client) AppName() string { return c.cfg.AppName }

// AppURL expone la URL base de la app para las plantillas.
func (c *Client) AppURL() string { return c.cfg.AppURL }

// Send envía un correo individual.
func (c *Client) Send(ctx context.Context, msg Message) error {
	if !c.Enabled() {
		slog.Warn("mailer: envío omitido, RESEND_API_KEY vacía", "subject", msg.Subject, "to", len(msg.To))
		return ErrNotConfigured
	}
	if len(msg.To) == 0 {
		return errors.New("mailer: destinatario vacío")
	}
	return c.post(ctx, "/emails", c.payload(msg))
}

// SendBatch envía hasta 100 correos distintos en una sola llamada.
// Se usa para repartir accesos corporativos sin generar N round-trips.
func (c *Client) SendBatch(ctx context.Context, msgs []Message) error {
	if !c.Enabled() {
		slog.Warn("mailer: lote omitido, RESEND_API_KEY vacía", "mensajes", len(msgs))
		return ErrNotConfigured
	}
	if len(msgs) == 0 {
		return nil
	}

	for start := 0; start < len(msgs); start += maxBatchSize {
		end := min(start+maxBatchSize, len(msgs))

		batch := make([]map[string]any, 0, end-start)
		for _, m := range msgs[start:end] {
			if len(m.To) == 0 {
				continue
			}
			batch = append(batch, c.payload(m))
		}
		if len(batch) == 0 {
			continue
		}
		if err := c.post(ctx, "/emails/batch", batch); err != nil {
			return err
		}
	}
	return nil
}

// SendAsync dispara el envío en segundo plano con su propio contexto.
// Útil en rutas HTTP donde el correo no debe bloquear la respuesta al usuario
// (registro, confirmación de compra). Los fallos se registran, no se propagan.
func (c *Client) SendAsync(msg Message) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		if err := c.Send(ctx, msg); err != nil && !errors.Is(err, ErrNotConfigured) {
			slog.Error("mailer: fallo en envío asíncrono", "error", err, "subject", msg.Subject)
		}
	}()
}

// ── Internos ─────────────────────────────────────────────────────────────────

func (c *Client) payload(msg Message) map[string]any {
	p := map[string]any{
		"from":    c.cfg.From,
		"to":      msg.To,
		"subject": msg.Subject,
		"html":    msg.HTML,
	}
	if c.cfg.ReplyTo != "" {
		p["reply_to"] = c.cfg.ReplyTo
	}
	return p
}

func (c *Client) post(ctx context.Context, path string, body any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("mailer: serializar payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("mailer: construir petición: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: llamada a Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	// Se limita la lectura del cuerpo de error: Resend responde JSON corto y no
	// queremos exponer respuestas gigantes en los logs.
	detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("mailer: Resend respondió %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
}

// safeURL evita inyectar URLs mal formadas en las plantillas HTML.
func safeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	return u.String()
}
