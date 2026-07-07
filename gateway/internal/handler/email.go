package handler

import (
	"fmt"
	"math"
	"net/smtp"
	"net/url"
	"regexp"
	"strings"

	"Prueba-Go/gateway/internal/config"
)

// sendDC3RepresentativeEmail envía el correo al representante con el link al formulario de trámite DC-3.
func sendDC3RepresentativeEmail(to, name, nombreCurso string, duracionMinutos int) error {
	host := config.C.SMTPHost
	if host == "" {
		return fmt.Errorf("SMTP_HOST no configurado en gateway")
	}

	user := config.C.SMTPUser
	pass := config.C.SMTPPass
	from := config.C.SMTPFrom
	if from == "" {
		from = user
	}
	appName := config.C.AppName
	appURL := config.C.AppURL

	subject := fmt.Sprintf("Constancias DC-3 disponibles — %s — %s", nombreCurso, appName)
	body := buildDC3EmailHTML(appURL, appName, name, nombreCurso, duracionMinutos)

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + body

	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(host+":"+config.C.SMTPPort, auth, from, []string{to}, []byte(msg))
}

func buildDC3EmailHTML(appURL, appName, name, nombreCurso string, duracionMinutos int) string {
	logoURL := appURL + "/logo-capacitaciones.png"
	nombreCursoClean := regexp.MustCompile(`(?i)^Exám?en(\s+Final)?\s*[-–:]*\s*`).ReplaceAllString(nombreCurso, "")
	nombreCursoClean = strings.TrimSpace(nombreCursoClean)
	if nombreCursoClean == "" {
		nombreCursoClean = "Capacitación"
	}

	duracionHoras := int(math.Ceil(float64(duracionMinutos) / 60.0))
	if duracionHoras < 1 {
		duracionHoras = 1
	}

	formURL := fmt.Sprintf("https://dc3.mhsolucionesempresariales.com/formulario-dc3-8f9d3a2b?nombre_curso=%s&duracion_horas=%d&area_tematica=6000", url.QueryEscape(nombreCursoClean), duracionHoras)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:48px 0">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%">

        <!-- Header -->
        <tr><td style="background:#1c1d1f;border-radius:16px 16px 0 0;padding:32px 40px;text-align:center">
          <img src="%s" width="60" height="60" alt="%s" style="display:block;margin:0 auto 14px;border-radius:12px" />
          <h1 style="margin:0;color:#ffffff;font-size:22px;font-weight:800;letter-spacing:-0.4px">%s</h1>
        </td></tr>

        <!-- Body -->
        <tr><td style="background:#ffffff;padding:40px 40px 32px">
          <h2 style="margin:0 0 14px;font-size:20px;font-weight:800;color:#111827">
            📋 Trámite de Constancias DC-3
          </h2>
          <p style="margin:0 0 18px;color:#374151;font-size:15px;line-height:1.65">
            Hola <strong>%s</strong>,
          </p>
          <p style="margin:0 0 24px;color:#4b5563;font-size:15px;line-height:1.65">
            La videollamada de la capacitación <strong>%s</strong> ha concluido satisfactoriamente. Como representante o responsable del grupo, puedes proceder a la generación y descarga de las constancias DC-3 para tus trabajadores y participantes.
          </p>

          <!-- CTA Box -->
          <table width="100%%" cellpadding="0" cellspacing="0">
            <tr><td align="center" style="padding:12px 0 32px">
              <a href="%s" target="_blank" style="display:inline-block;background:#10b981;color:#ffffff;text-decoration:none;font-weight:700;font-size:16px;padding:14px 28px;border-radius:10px;box-shadow:0 4px 12px rgba(16,185,129,0.3)">
                👉 Formulario para Constancias DC-3
              </a>
            </td></tr>
          </table>

          <div style="background:#f9fafb;border-left:4px solid #3b82f6;padding:16px 20px;border-radius:0 8px 8px 0;margin-bottom:24px">
            <p style="margin:0;font-size:13px;color:#4b5563;line-height:1.5">
              <strong>Nota importante:</strong> El enlace ya contiene precargado el nombre de la capacitación y la duración en horas (%d hrs). Por favor, asegúrate de ingresar los datos correctos del centro de trabajo y los nombres oficiales de los participantes.
            </p>
          </div>

          <p style="margin:0;font-size:14px;color:#6b7280;line-height:1.6">
            Agradecemos tu preferencia.<br />
            <strong>El equipo de %s</strong>
          </p>
        </td></tr>

        <!-- Footer -->
        <tr><td style="background:#f9fafb;border-radius:0 0 16px 16px;padding:20px 40px;text-align:center;border-top:1px solid #f3f4f6">
          <p style="margin:0;font-size:12px;color:#9ca3af">
            © %s &nbsp;·&nbsp; Plataforma de Capacitación Empresarial
          </p>
        </td></tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`, logoURL, appName, appName, name, nombreCurso, formURL, duracionHoras, appName, appName)
}
