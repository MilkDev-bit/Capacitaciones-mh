package handlers

import (
	"fmt"
	"net/smtp"

	"Prueba-Go/internal/config"
)

func sendPasswordResetEmail(to, code string) error {
	host := config.C.SMTPHost
	if host == "" {
		return fmt.Errorf("SMTP_HOST no configurado")
	}

	user := config.C.SMTPUser
	pass := config.C.SMTPPass
	from := config.C.SMTPFrom
	if from == "" {
		from = user
	}
	appName := config.C.AppName
	appURL := config.C.AppURL

	subject := fmt.Sprintf("Código de recuperación — %s", appName)
	body := buildResetEmailHTML(appURL, appName, code)

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + body

	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(host+":"+config.C.SMTPPort, auth, from, []string{to}, []byte(msg))
}

func buildResetEmailHTML(appURL, appName, code string) string {
	logoURL := appURL + "/logo-capacitaciones.png"
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:48px 0">
    <tr><td align="center">
      <table width="560" cellpadding="0" cellspacing="0" style="max-width:560px;width:100%%">

        <!-- Header -->
        <tr><td style="background:#1c1d1f;border-radius:16px 16px 0 0;padding:32px 40px;text-align:center">
          <img src="%s" width="60" height="60" alt="%s" style="display:block;margin:0 auto 14px;border-radius:12px" />
          <h1 style="margin:0;color:#ffffff;font-size:22px;font-weight:800;letter-spacing:-0.4px">%s</h1>
        </td></tr>

        <!-- Body -->
        <tr><td style="background:#ffffff;padding:40px 40px 32px">
          <h2 style="margin:0 0 10px;font-size:20px;font-weight:800;color:#111827">
            🔐 Recuperar contraseña
          </h2>
          <p style="margin:0 0 24px;color:#6b7280;font-size:15px;line-height:1.65">
            Recibimos una solicitud para restablecer tu contraseña. Usa el siguiente
            código de verificación en la plataforma.
            <strong style="color:#374151">Expira en 15 minutos.</strong>
          </p>

          <!-- Code block -->
          <table width="100%%" cellpadding="0" cellspacing="0">
            <tr><td align="center" style="padding:4px 0 28px">
              <div style="display:inline-block;background:#fff7ed;border:2px solid #fdba74;border-radius:14px;padding:24px 40px;text-align:center">
                <p style="margin:0 0 6px;font-size:12px;font-weight:700;color:#f97316;letter-spacing:1.5px;text-transform:uppercase">Tu código de verificación</p>
                <span style="letter-spacing:12px;font-size:38px;font-weight:900;color:#111827;font-family:monospace">%s</span>
              </div>
            </td></tr>
          </table>

          <p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
            Si no solicitaste este cambio, puedes ignorar este mensaje de forma segura.
            Tu contraseña permanecerá sin cambios.
          </p>
        </td></tr>

        <!-- Footer -->
        <tr><td style="background:#f9fafb;border-radius:0 0 16px 16px;padding:20px 40px;text-align:center;border-top:1px solid #f3f4f6">
          <p style="margin:0;font-size:12px;color:#9ca3af">
            © %s &nbsp;·&nbsp; Este código es de un solo uso y válido por <strong>15 minutos</strong>.
          </p>
        </td></tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`, logoURL, appName, appName, code, appName)
}
