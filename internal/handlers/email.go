package handlers

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"Prueba-Go/internal/config"
	"Prueba-Go/pkg/mailer"
)

// Este archivo mantiene el envío de correo del monolito legado apuntando al
// mismo paquete (y por tanto a las mismas plantillas) que usan el gateway y el
// auth service. Antes duplicaba el HTML y hablaba SMTP por su cuenta.

var (
	mailOnce   sync.Once
	mailClient *mailer.Client
)

// mail devuelve el cliente de Resend, construido de forma perezosa porque
// config.C se rellena después de que se cargue este paquete.
func mail() *mailer.Client {
	mailOnce.Do(func() {
		from := config.C.ResendFrom
		if from == "" {
			from = config.C.SMTPFrom // compatibilidad con despliegues antiguos
		}
		mailClient = mailer.New(mailer.Config{
			APIKey:  config.C.ResendAPIKey,
			From:    from,
			ReplyTo: config.C.ResendReplyTo,
			AppName: config.C.AppName,
			AppURL:  config.C.AppURL,
		})
	})
	return mailClient
}

// sendPasswordResetEmail envía el código de recuperación de contraseña.
func sendPasswordResetEmail(to, code string) error {
	m := mail()
	if !m.Enabled() {
		return fmt.Errorf("RESEND_API_KEY no configurada")
	}

	msg := m.PasswordResetCode(code)
	msg.To = []string{to}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return m.Send(ctx, msg)
}

// prefijoExamen limpia encabezados tipo "Examen Final - " del título del curso.
var prefijoExamen = regexp.MustCompile(`(?i)^Exám?en(\s+Final)?\s*[-–:]*\s*`)

// SendDC3RepresentativeEmail avisa al representante que puede tramitar las
// constancias DC-3 del grupo.
func SendDC3RepresentativeEmail(to, name, nombreCurso string, duracionMinutos int) error {
	m := mail()
	if !m.Enabled() {
		return fmt.Errorf("RESEND_API_KEY no configurada")
	}

	nombreLimpio := strings.TrimSpace(prefijoExamen.ReplaceAllString(nombreCurso, ""))
	if nombreLimpio == "" {
		nombreLimpio = "Capacitación"
	}

	// La STPS cuenta horas completas: 45 minutos siguen siendo 1 hora.
	duracionHoras := int(math.Ceil(float64(duracionMinutos) / 60.0))
	if duracionHoras < 1 {
		duracionHoras = 1
	}

	formURL := fmt.Sprintf(
		"https://dc3.mhsolucionesempresariales.com/formulario-dc3-8f9d3a2b?nombre_curso=%s&duracion_horas=%d&area_tematica=6000",
		url.QueryEscape(nombreLimpio), duracionHoras,
	)

	msg := m.DC3Representative(name, nombreLimpio, duracionHoras, formURL)
	msg.To = []string{to}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return m.Send(ctx, msg)
}
