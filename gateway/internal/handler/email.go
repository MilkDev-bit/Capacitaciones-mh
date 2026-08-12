package handler

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"Prueba-Go/pkg/mailer"
)

// DC3Notifier envía el aviso de constancias DC-3 al representante de una
// licencia corporativa.
//
// Vive aparte de los handlers porque el disparador cambió de dueño: antes lo
// invocaba CursosHandler al cerrar una videollamada; ahora lo hace
// LeccionesHandler cuando un participante termina el contenido.
type DC3Notifier struct {
	mail *mailer.Client
}

func NewDC3Notifier(mail *mailer.Client) *DC3Notifier { return &DC3Notifier{mail: mail} }

// prefijoExamen limpia encabezados tipo "Examen Final - " del título del curso:
// en el formato DC-3 va el nombre de la capacitación, no el del examen.
var prefijoExamen = regexp.MustCompile(`(?i)^Exám?en(\s+Final)?\s*[-–:]*\s*`)

// EnviarAvisoDC3 manda el correo con el enlace al formulario de constancias.
// Envío síncrono: el llamador decide si lo lanza en una goroutine.
func (d *DC3Notifier) EnviarAvisoDC3(to, name, nombreCurso string, duracionMinutos int) error {
	if !d.mail.Enabled() {
		return fmt.Errorf("RESEND_API_KEY no configurada en gateway")
	}

	nombreLimpio := strings.TrimSpace(prefijoExamen.ReplaceAllString(nombreCurso, ""))
	if nombreLimpio == "" {
		nombreLimpio = "Capacitación"
	}

	// La STPS razona en horas completas: 45 minutos siguen siendo 1 hora.
	duracionHoras := int(math.Ceil(float64(duracionMinutos) / 60.0))
	if duracionHoras < 1 {
		duracionHoras = 1
	}

	formURL := fmt.Sprintf(
		"https://dc3.mhsolucionesempresariales.com/formulario-dc3-8f9d3a2b?nombre_curso=%s&duracion_horas=%d&area_tematica=6000",
		url.QueryEscape(nombreLimpio), duracionHoras,
	)

	msg := d.mail.DC3Representative(name, nombreLimpio, duracionHoras, formURL)
	msg.To = []string{to}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return d.mail.Send(ctx, msg)
}
