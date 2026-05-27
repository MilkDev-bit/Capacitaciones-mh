package handlers

import (
	"fmt"
	"net/smtp"
	"os"
)

func sendPasswordResetEmail(to, code string) error {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return fmt.Errorf("SMTP_HOST no configurado")
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = user
	}
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "Capacitaciones MH"
	}

	subject := fmt.Sprintf("Código de recuperación — %s", appName)
	body := buildResetEmailHTML(appName, code)

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + body

	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))
}

func buildResetEmailHTML(appName, code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<body style="margin:0;padding:40px 0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif">
  <div style="max-width:480px;margin:0 auto;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,.08)">
    <div style="background:#f97316;padding:28px 32px">
      <h1 style="margin:0;color:#fff;font-size:20px;font-weight:700">%s</h1>
    </div>
    <div style="padding:32px">
      <h2 style="margin:0 0 8px;font-size:18px;color:#111827">Recuperar contraseña</h2>
      <p style="color:#6b7280;margin:0 0 24px;line-height:1.6">
        Recibimos una solicitud para restablecer tu contraseña.
        Usa el siguiente código en la plataforma. <strong>Expira en 15 minutos.</strong>
      </p>
      <div style="background:#f9fafb;border:2px dashed #e5e7eb;border-radius:10px;padding:24px;text-align:center">
        <span style="letter-spacing:10px;font-size:34px;font-weight:800;color:#111827;font-variant-numeric:tabular-nums">%s</span>
      </div>
      <p style="color:#9ca3af;font-size:13px;margin:24px 0 0;line-height:1.5">
        Si no solicitaste esto, puedes ignorar este mensaje. Tu contraseña no cambiará.
      </p>
    </div>
    <div style="background:#f9fafb;padding:16px 32px;border-top:1px solid #f3f4f6">
      <p style="margin:0;font-size:12px;color:#9ca3af">Este código es de un solo uso y válido por 15 minutos.</p>
    </div>
  </div>
</body>
</html>`, appName, code)
}
