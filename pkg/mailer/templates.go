package mailer

import (
	"fmt"
	"html"
	"strings"
)

// Este archivo concentra TODAS las plantillas HTML del sistema.
//
// Regla: cualquier dato que provenga del usuario (nombre, título de curso,
// código) se pasa por html.EscapeString antes de interpolarse. Un nombre con
// "<script>" no debe poder romper el correo ni el cliente que lo renderiza.

// ── Layout base ──────────────────────────────────────────────────────────────

// layout envuelve el contenido en la plantilla corporativa (header oscuro con
// logo, cuerpo blanco, footer gris) usada en toda la plataforma.
func (c *Client) layout(title, bodyHTML, footerNote string) string {
	appName := html.EscapeString(c.cfg.AppName)
	logo := ""
	if base := safeURL(c.cfg.AppURL); base != "" {
		logo = fmt.Sprintf(
			`<img src="%s/logo-capacitaciones.png" width="60" height="60" alt="%s" style="display:block;margin:0 auto 14px;border-radius:12px" />`,
			base, appName,
		)
	}
	if footerNote == "" {
		footerNote = "Plataforma de Capacitación Empresarial"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>%s</title></head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:48px 0">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%">

        <tr><td style="background:#1c1d1f;border-radius:16px 16px 0 0;padding:32px 40px;text-align:center">
          %s
          <h1 style="margin:0;color:#ffffff;font-size:22px;font-weight:800;letter-spacing:-0.4px">%s</h1>
        </td></tr>

        <tr><td style="background:#ffffff;padding:40px 40px 32px">%s</td></tr>

        <tr><td style="background:#f9fafb;border-radius:0 0 16px 16px;padding:20px 40px;text-align:center;border-top:1px solid #f3f4f6">
          <p style="margin:0;font-size:12px;color:#9ca3af">© %s &nbsp;·&nbsp; %s</p>
        </td></tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`, html.EscapeString(title), logo, appName, bodyHTML, appName, html.EscapeString(footerNote))
}

// codeBox renderiza el recuadro destacado con un código de un solo uso.
func codeBox(label, code string) string {
	return fmt.Sprintf(`
<table width="100%%" cellpadding="0" cellspacing="0">
  <tr><td align="center" style="padding:4px 0 28px">
    <div style="display:inline-block;background:#fff7ed;border:2px solid #fdba74;border-radius:14px;padding:24px 40px;text-align:center">
      <p style="margin:0 0 6px;font-size:12px;font-weight:700;color:#f97316;letter-spacing:1.5px;text-transform:uppercase">%s</p>
      <span style="letter-spacing:12px;font-size:38px;font-weight:900;color:#111827;font-family:monospace">%s</span>
    </div>
  </td></tr>
</table>`, html.EscapeString(label), html.EscapeString(code))
}

// button renderiza el CTA principal.
func button(label, href, color string) string {
	safe := safeURL(href)
	if safe == "" {
		return ""
	}
	return fmt.Sprintf(`
<table width="100%%" cellpadding="0" cellspacing="0">
  <tr><td align="center" style="padding:12px 0 28px">
    <a href="%s" target="_blank" style="display:inline-block;background:%s;color:#ffffff;text-decoration:none;font-weight:700;font-size:16px;padding:14px 30px;border-radius:10px">%s</a>
  </td></tr>
</table>`, safe, color, html.EscapeString(label))
}

// ── Verificación de correo ───────────────────────────────────────────────────

// VerificationCode construye el correo con el código de 6 dígitos que confirma
// la propiedad del buzón durante el registro.
func (c *Client) VerificationCode(name, code string, minutes int) Message {
	body := fmt.Sprintf(`
<h2 style="margin:0 0 10px;font-size:20px;font-weight:800;color:#111827">Confirma tu correo electrónico</h2>
<p style="margin:0 0 8px;color:#374151;font-size:15px;line-height:1.65">Hola <strong>%s</strong>,</p>
<p style="margin:0 0 24px;color:#6b7280;font-size:15px;line-height:1.65">
  Tu cuenta está casi lista. Ingresa este código en la plataforma para activarla.
  <strong style="color:#374151">Expira en %d minutos.</strong>
</p>
%s
<p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
  Si no creaste esta cuenta, ignora este mensaje: sin el código nadie puede activarla.
</p>`, html.EscapeString(firstName(name)), minutes, codeBox("Tu código de verificación", code))

	return Message{
		Subject: fmt.Sprintf("%s es tu código de verificación — %s", code, c.cfg.AppName),
		HTML:    c.layout("Verifica tu correo", body, "Este código es de un solo uso."),
	}
}

// ── Recuperación de contraseña ───────────────────────────────────────────────

// PasswordResetCode: variante con código de 6 dígitos.
func (c *Client) PasswordResetCode(code string) Message {
	body := fmt.Sprintf(`
<h2 style="margin:0 0 10px;font-size:20px;font-weight:800;color:#111827">Recuperar contraseña</h2>
<p style="margin:0 0 24px;color:#6b7280;font-size:15px;line-height:1.65">
  Recibimos una solicitud para restablecer tu contraseña. Usa el siguiente código en la plataforma.
  <strong style="color:#374151">Expira en 15 minutos.</strong>
</p>
%s
<p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
  Si no solicitaste este cambio, ignora este mensaje. Tu contraseña permanecerá sin cambios.
</p>`, codeBox("Tu código de verificación", code))

	return Message{
		Subject: fmt.Sprintf("Código de recuperación — %s", c.cfg.AppName),
		HTML:    c.layout("Recuperar contraseña", body, "Este código es de un solo uso y válido por 15 minutos."),
	}
}

// PasswordResetLink: variante con enlace firmado (la que usa el auth service).
func (c *Client) PasswordResetLink(name, link string) Message {
	body := fmt.Sprintf(`
<h2 style="margin:0 0 10px;font-size:20px;font-weight:800;color:#111827">Recuperar contraseña</h2>
<p style="margin:0 0 8px;color:#374151;font-size:15px;line-height:1.65">Hola <strong>%s</strong>,</p>
<p style="margin:0 0 20px;color:#6b7280;font-size:15px;line-height:1.65">
  Pulsa el botón para crear una contraseña nueva. <strong style="color:#374151">El enlace caduca en 1 hora.</strong>
</p>
%s
<p style="margin:0 0 16px;font-size:13px;color:#9ca3af;line-height:1.6;word-break:break-all">
  Si el botón no funciona, copia esta dirección en tu navegador:<br />%s
</p>
<p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
  Si no solicitaste este cambio, ignora este mensaje.
</p>`, html.EscapeString(firstName(name)), button("Restablecer contraseña", link, "#f97316"), html.EscapeString(link))

	return Message{
		Subject: fmt.Sprintf("Restablece tu contraseña — %s", c.cfg.AppName),
		HTML:    c.layout("Recuperar contraseña", body, "Enlace válido por 1 hora."),
	}
}

// ── Compras ──────────────────────────────────────────────────────────────────

// PurchaseLine describe un renglón del resumen de compra.
type PurchaseLine struct {
	Titulo   string
	Tipo     string // "Inscripción individual" | "Licencias corporativas"
	Cantidad int
}

// PurchaseConfirmation es el acuse de recibo que se envía al comprador en
// cuanto Stripe confirma el pago, sin importar el tipo de producto.
func (c *Client) PurchaseConfirmation(name string, lines []PurchaseLine, total float64, ctaLabel, ctaURL string) Message {
	var rows strings.Builder
	for _, l := range lines {
		cantidad := ""
		if l.Cantidad > 1 {
			cantidad = fmt.Sprintf(` <span style="color:#f97316;font-weight:700">× %d</span>`, l.Cantidad)
		}
		rows.WriteString(fmt.Sprintf(`
<tr>
  <td style="padding:14px 0;border-bottom:1px solid #f3f4f6">
    <div style="font-size:15px;font-weight:700;color:#111827">%s%s</div>
    <div style="font-size:13px;color:#6b7280;margin-top:2px">%s</div>
  </td>
</tr>`, html.EscapeString(l.Titulo), cantidad, html.EscapeString(l.Tipo)))
	}

	body := fmt.Sprintf(`
<h2 style="margin:0 0 10px;font-size:20px;font-weight:800;color:#111827">✅ Pago confirmado</h2>
<p style="margin:0 0 8px;color:#374151;font-size:15px;line-height:1.65">Hola <strong>%s</strong>,</p>
<p style="margin:0 0 20px;color:#6b7280;font-size:15px;line-height:1.65">
  Tu pago se procesó correctamente. Este es el detalle de tu compra:
</p>

<table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:8px">%s</table>

<table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:24px">
  <tr>
    <td style="padding:16px 0 0;font-size:15px;color:#6b7280">Total pagado</td>
    <td style="padding:16px 0 0;font-size:20px;font-weight:800;color:#111827;text-align:right">$%s MXN</td>
  </tr>
</table>

%s
<p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
  Tu factura está disponible desde la plataforma. Si compraste licencias corporativas,
  recibirás un segundo correo con los accesos para tu equipo.
</p>`, html.EscapeString(firstName(name)), rows.String(), formatMoney(total), button(ctaLabel, ctaURL, "#10b981"))

	return Message{
		Subject: fmt.Sprintf("Confirmación de compra — %s", c.cfg.AppName),
		HTML:    c.layout("Pago confirmado", body, "Gracias por tu compra."),
	}
}

// ── Licencias corporativas ───────────────────────────────────────────────────

// CorporateLicense entrega al comprador el/los código(s) de acceso de la
// licencia recién adquirida. Sustituye la necesidad de entrar al módulo.
func (c *Client) CorporateLicense(name, cursoTitulo string, lugares int, codigoCompartido, gestionURL string) Message {
	accesos := `<p style="margin:0 0 24px;color:#6b7280;font-size:15px">Consulta tus accesos desde la plataforma.</p>`
	if codigoCompartido != "" {
		accesos = fmt.Sprintf(`
<p style="margin:0 0 4px;color:#374151;font-size:15px;line-height:1.65">
  Comparte este código con tu equipo para que se inscriban al curso:
</p>
%s`, codeBox("Código de acceso", codigoCompartido))
	}

	body := fmt.Sprintf(`
<h2 style="margin:0 0 10px;font-size:20px;font-weight:800;color:#111827">🎟️ Tus accesos corporativos</h2>
<p style="margin:0 0 8px;color:#374151;font-size:15px;line-height:1.65">Hola <strong>%s</strong>,</p>
<p style="margin:0 0 20px;color:#6b7280;font-size:15px;line-height:1.65">
  Adquiriste <strong>%d lugar(es)</strong> para <strong>%s</strong>. Aquí están los accesos.
</p>
%s
%s
<p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
  ¿Prefieres que los enviemos nosotros? Desde la plataforma puedes capturar los correos de tus
  participantes y cada uno recibirá su acceso automáticamente.
</p>`, html.EscapeString(firstName(name)), lugares, html.EscapeString(cursoTitulo), accesos,
		button("Repartir accesos por correo", gestionURL, "#f97316"))

	return Message{
		Subject: fmt.Sprintf("Accesos corporativos: %s — %s", cursoTitulo, c.cfg.AppName),
		HTML:    c.layout("Tus accesos corporativos", body, "Conserva estos códigos: son de un solo uso."),
	}
}

// ParticipantAccess es el correo que recibe cada trabajador/participante con su
// código individual, enviado por el comprador desde la plataforma.
func (c *Client) ParticipantAccess(nombreParticipante, empresaONombreComprador, cursoTitulo, codigo, accesoURL string) Message {
	saludo := "Hola"
	if n := firstName(nombreParticipante); n != "" {
		saludo = "Hola <strong>" + html.EscapeString(n) + "</strong>"
	}
	invitador := html.EscapeString(empresaONombreComprador)
	if invitador == "" {
		invitador = "tu organización"
	}

	body := fmt.Sprintf(`
<h2 style="margin:0 0 10px;font-size:20px;font-weight:800;color:#111827">Te inscribieron a una capacitación</h2>
<p style="margin:0 0 8px;color:#374151;font-size:15px;line-height:1.65">%s,</p>
<p style="margin:0 0 20px;color:#6b7280;font-size:15px;line-height:1.65">
  <strong>%s</strong> reservó un lugar para ti en <strong>%s</strong>.
  Usa el código de abajo para activar tu acceso.
</p>
%s
%s
<p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6">
  Este código es personal e intransferible: solo puede usarse una vez.
</p>`, saludo, invitador, html.EscapeString(cursoTitulo),
		codeBox("Tu código de acceso", codigo),
		button("Activar mi acceso", accesoURL, "#10b981"))

	return Message{
		Subject: fmt.Sprintf("Tu acceso a %s — %s", cursoTitulo, c.cfg.AppName),
		HTML:    c.layout("Tu acceso a la capacitación", body, "Acceso personal e intransferible."),
	}
}

// ── DC-3 ─────────────────────────────────────────────────────────────────────

// DC3Representative avisa al representante que ya puede tramitar constancias.
func (c *Client) DC3Representative(name, cursoTitulo string, duracionHoras int, formURL string) Message {
	body := fmt.Sprintf(`
<h2 style="margin:0 0 14px;font-size:20px;font-weight:800;color:#111827">📋 Trámite de Constancias DC-3</h2>
<p style="margin:0 0 18px;color:#374151;font-size:15px;line-height:1.65">Hola <strong>%s</strong>,</p>
<p style="margin:0 0 24px;color:#4b5563;font-size:15px;line-height:1.65">
  La capacitación <strong>%s</strong> ha concluido. Como representante del grupo, ya puedes generar
  y descargar las constancias DC-3 de tus participantes.
</p>
%s
<div style="background:#f9fafb;border-left:4px solid #3b82f6;padding:16px 20px;border-radius:0 8px 8px 0;margin-bottom:24px">
  <p style="margin:0;font-size:13px;color:#4b5563;line-height:1.5">
    <strong>Nota:</strong> el enlace ya lleva precargados el nombre de la capacitación y la duración
    (%d hrs). Verifica los datos del centro de trabajo y los nombres oficiales de los participantes.
  </p>
</div>
<p style="margin:0;font-size:14px;color:#6b7280;line-height:1.6">
  Agradecemos tu preferencia.<br /><strong>El equipo de %s</strong>
</p>`, html.EscapeString(firstName(name)), html.EscapeString(cursoTitulo),
		button("Formulario para Constancias DC-3", formURL, "#10b981"),
		duracionHoras, html.EscapeString(c.cfg.AppName))

	return Message{
		Subject: fmt.Sprintf("Constancias DC-3 disponibles — %s — %s", cursoTitulo, c.cfg.AppName),
		HTML:    c.layout("Constancias DC-3", body, ""),
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// firstName acorta el nombre completo al primer token para un saludo natural.
func firstName(full string) string {
	full = strings.TrimSpace(full)
	if full == "" {
		return ""
	}
	if i := strings.IndexByte(full, ' '); i > 0 {
		return full[:i]
	}
	return full
}

// formatMoney imprime el importe con separador de miles y dos decimales.
func formatMoney(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	intPart, decPart, _ := strings.Cut(s, ".")

	neg := strings.HasPrefix(intPart, "-")
	intPart = strings.TrimPrefix(intPart, "-")

	var b strings.Builder
	for i, d := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(d)
	}
	if neg {
		return "-" + b.String() + "." + decPart
	}
	return b.String() + "." + decPart
}
