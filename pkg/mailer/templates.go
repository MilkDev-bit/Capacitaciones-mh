package mailer

import (
	"fmt"
	"html"
	"strings"
)

// Este archivo concentra TODAS las plantillas HTML del sistema.
//
// Reglas de diseño de correo (no son preferencias estéticas, son restricciones
// del medio):
//
//   - Todo con <table>: Outlook renderiza con el motor de Word y descarta
//     flexbox, grid y la mayoría de position/float.
//   - Estilos en línea: Gmail elimina <style> en la vista de conversación.
//   - Colores explícitos en cada celda: sin ellos, el modo oscuro de Gmail y
//     Outlook invierte los fondos y deja texto gris sobre gris (que es
//     exactamente lo que se veía en el correo de accesos corporativos).
//   - Ancho máximo 600 px y fuentes de sistema: nada de webfonts.
//
// Seguridad: cualquier dato que provenga del usuario (nombre, título de curso,
// código) se pasa por html.EscapeString antes de interpolarse. Un nombre con
// "<script>" no debe poder romper el correo ni el cliente que lo renderiza.

// Paleta de marca — replica las variables CSS del frontend (main.css).
const (
	colBrand      = "#f97316"
	colBrandDark  = "#ea580c"
	colBrandLight = "#fff7ed"
	colInk        = "#111827"
	colBody       = "#374151"
	colMuted      = "#6b7280"
	colFaint      = "#9ca3af"
	colLine       = "#e5e7eb"
	colCanvas     = "#f2f2f4"
	colSuccess    = "#10b981"
	colHeader     = "#1c1d1f"
)

// ── Layout base ──────────────────────────────────────────────────────────────

// layout envuelve el contenido en la plantilla corporativa: barra de acento
// naranja, cabecera oscura con el logo, tarjeta blanca y pie discreto.
//
// preheader es el texto de vista previa que muestran Gmail/Outlook junto al
// asunto. Va oculto en el cuerpo; sin él los clientes rellenan ese espacio con
// las primeras palabras del HTML, que suelen ser basura.
func (c *Client) layout(title, preheader, bodyHTML, footerNote string) string {
	appName := html.EscapeString(c.cfg.AppName)

	// El logo se sirve desde la raíz pública del frontend. Si APP_URL no está
	// configurada o la imagen no carga, el alt text hereda el estilo del <img>
	// y se ve como un texto de marca en vez de un icono roto.
	logo := ""
	if base := safeURL(c.cfg.AppURL); base != "" {
		logo = fmt.Sprintf(`
          <tr><td align="center" style="padding:0 0 14px">
            <img src="%s/logo-capacitaciones.png" width="56" height="56" alt="%s"
                 style="display:block;border:0;outline:none;text-decoration:none;border-radius:14px;background:#ffffff;color:#ffffff;font-size:13px;font-weight:700;line-height:56px;text-align:center" />
          </td></tr>`, base, appName)
	}

	if footerNote == "" {
		footerNote = "Plataforma de Capacitación Empresarial"
	}

	// Enlace del pie hacia la plataforma; se omite si no hay APP_URL válida.
	footerLink := ""
	if base := safeURL(c.cfg.AppURL); base != "" {
		footerLink = fmt.Sprintf(
			`<a href="%s" target="_blank" style="color:%s;text-decoration:none;font-weight:600">Ir a la plataforma</a> &nbsp;·&nbsp; `,
			base, colBrand)
	}

	return fmt.Sprintf(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" lang="es">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1.0" />
  <meta http-equiv="X-UA-Compatible" content="IE=edge" />
  <meta name="color-scheme" content="light only" />
  <meta name="supported-color-schemes" content="light only" />
  <title>%s</title>
  <!--[if mso]><style>body,table,td,a{font-family:Arial,Helvetica,sans-serif !important}</style><![endif]-->
</head>
<body style="margin:0;padding:0;width:100%%;background-color:%s;color:%s;-webkit-text-size-adjust:100%%;-ms-text-size-adjust:100%%">

  <div style="display:none;font-size:1px;color:%s;line-height:1px;max-height:0;max-width:0;opacity:0;overflow:hidden">%s</div>
  <div style="display:none;max-height:0;overflow:hidden">&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;</div>

  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:%s">
    <tr><td align="center" style="padding:32px 12px 40px">

      <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="width:100%%;max-width:600px;background-color:#ffffff;border-radius:18px;overflow:hidden;box-shadow:0 6px 24px rgba(0,0,0,0.06)">

        <tr><td style="height:5px;line-height:5px;font-size:0;background-color:%s;background-image:linear-gradient(90deg,%s 0%%,%s 100%%)">&nbsp;</td></tr>

        <tr><td align="center" style="background-color:%s;padding:30px 40px 26px">
          <table role="presentation" cellpadding="0" cellspacing="0" border="0" align="center">
            %s
            <tr><td align="center" style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;font-size:20px;font-weight:800;color:#ffffff;letter-spacing:-0.3px;line-height:1.25">%s</td></tr>
          </table>
        </td></tr>

        <tr><td style="background-color:#ffffff;padding:36px 40px 30px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif">%s</td></tr>

        <tr><td style="background-color:#fafafa;padding:18px 40px 22px;text-align:center;border-top:1px solid #f1f1f3">
          <p style="margin:0 0 4px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;font-size:12px;line-height:1.6;color:%s">%s%s</p>
          <p style="margin:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;font-size:11px;line-height:1.6;color:%s">&copy; %s &nbsp;·&nbsp; Correo automático, no respondas a este mensaje.</p>
        </td></tr>

      </table>

    </td></tr>
  </table>
</body>
</html>`,
		html.EscapeString(title),
		colCanvas, colBody,
		colCanvas, html.EscapeString(preheader),
		colCanvas,
		colBrand, colBrand, colBrandDark,
		colHeader,
		logo,
		appName,
		bodyHTML,
		colFaint, footerLink, html.EscapeString(footerNote),
		colFaint, appName,
	)
}

// ── Bloques reutilizables ────────────────────────────────────────────────────

// heading renderiza el título del cuerpo. eyebrow es la etiqueta pequeña en
// naranja que va encima (opcional).
func heading(eyebrow, text string) string {
	out := ""
	if eyebrow != "" {
		out = fmt.Sprintf(
			`<p style="margin:0 0 8px;font-size:11px;font-weight:800;letter-spacing:1.4px;text-transform:uppercase;color:%s">%s</p>`,
			colBrand, html.EscapeString(eyebrow))
	}
	return out + fmt.Sprintf(
		`<h1 style="margin:0 0 14px;font-size:22px;line-height:1.3;font-weight:800;color:%s">%s</h1>`,
		colInk, html.EscapeString(text))
}

// paragraph escribe un párrafo de cuerpo. htmlContent ya viene escapado por el
// llamador porque suele mezclar texto fijo con <strong>.
func paragraph(htmlContent string) string {
	return fmt.Sprintf(
		`<p style="margin:0 0 16px;font-size:15px;line-height:1.7;color:%s">%s</p>`,
		colBody, htmlContent)
}

// note es el texto legal/aclaratorio del final, en gris pequeño.
func note(htmlContent string) string {
	return fmt.Sprintf(
		`<p style="margin:0;font-size:13px;line-height:1.65;color:%s">%s</p>`,
		colFaint, htmlContent)
}

// codeBox renderiza el recuadro destacado con un código de un solo uso.
//
// El tamaño se adapta al código: los OTP de 6 dígitos admiten tracking amplio
// y tipografía enorme, pero un token largo tipo "c90c2c4e-4b4" con
// letter-spacing:12px se parte en dos líneas y queda ilegible (era el defecto
// visible en el correo de accesos corporativos). Para esos casos se reduce el
// cuerpo, se recorta el tracking y se permite el corte controlado.
func codeBox(label, code string) string {
	size, tracking, wrap := "38px", "10px", "normal"
	if len(code) > 8 {
		size, tracking, wrap = "24px", "2px", "break-all"
	}
	if len(code) > 20 {
		size, tracking = "18px", "1px"
	}

	return fmt.Sprintf(`
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:4px 0 26px">
  <tr><td align="center" style="background-color:%s;border:1px solid #fed7aa;border-radius:14px;padding:22px 24px">
    <p style="margin:0 0 8px;font-size:11px;font-weight:800;color:%s;letter-spacing:1.4px;text-transform:uppercase">%s</p>
    <p style="margin:0;font-family:'SF Mono',SFMono-Regular,Menlo,Consolas,'Courier New',monospace;font-size:%s;line-height:1.25;font-weight:700;color:%s;letter-spacing:%s;word-break:%s">%s</p>
  </td></tr>
</table>`,
		colBrandLight, colBrandDark, html.EscapeString(label),
		size, colInk, tracking, wrap, html.EscapeString(code))
}

// button renderiza el CTA principal. Incluye el fallback VML para que Outlook
// (que ignora border-radius y padding en <a>) pinte un botón real y no un
// enlace suelto.
func button(label, href, color string) string {
	safe := safeURL(href)
	if safe == "" {
		return ""
	}
	esc := html.EscapeString(label)

	return fmt.Sprintf(`
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:6px 0 26px">
  <tr><td align="center">
    <!--[if mso]>
    <v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" xmlns:w="urn:schemas-microsoft-com:office:word"
                 href="%s" style="height:48px;v-text-anchor:middle;width:280px" arcsize="22%%" stroke="f" fillcolor="%s">
      <w:anchorlock/><center style="color:#ffffff;font-family:Arial,sans-serif;font-size:16px;font-weight:bold">%s</center>
    </v:roundrect>
    <![endif]-->
    <!--[if !mso]><!-- -->
    <a href="%s" target="_blank" rel="noopener"
       style="display:inline-block;background-color:%s;color:#ffffff;text-decoration:none;font-size:16px;font-weight:700;line-height:1.2;padding:15px 34px;border-radius:11px;mso-hide:all">%s</a>
    <!--<![endif]-->
  </td></tr>
</table>`, safe, color, esc, safe, color, esc)
}

// fallbackLink muestra la URL en texto plano bajo el botón. Algunos clientes
// corporativos desactivan los enlaces en HTML; sin esto el correo queda inútil.
func fallbackLink(href string) string {
	safe := safeURL(href)
	if safe == "" {
		return ""
	}
	return fmt.Sprintf(`
<p style="margin:0 0 18px;font-size:12px;line-height:1.6;color:%s;word-break:break-all">
  ¿El botón no funciona? Copia esta dirección en tu navegador:<br />
  <span style="color:%s">%s</span>
</p>`, colFaint, colMuted, html.EscapeString(safe))
}

// infoBox destaca una aclaración con una barra lateral de color.
func infoBox(accent, htmlContent string) string {
	return fmt.Sprintf(`
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 24px">
  <tr>
    <td style="width:4px;background-color:%s;border-radius:4px 0 0 4px">&nbsp;</td>
    <td style="background-color:#f9fafb;padding:14px 18px;border-radius:0 8px 8px 0;font-size:13px;line-height:1.65;color:%s">%s</td>
  </tr>
</table>`, accent, colBody, htmlContent)
}

// ── Verificación de correo ───────────────────────────────────────────────────

// VerificationCode construye el correo con el código de 6 dígitos que confirma
// la propiedad del buzón durante el registro.
func (c *Client) VerificationCode(name, code string, minutes int) Message {
	body := heading("Registro", "Confirma tu correo electrónico") +
		paragraph(fmt.Sprintf("Hola <strong style=\"color:%s\">%s</strong>, tu cuenta está casi lista.",
			colInk, html.EscapeString(firstName(name)))) +
		paragraph(fmt.Sprintf("Ingresa este código en la plataforma para activarla. <strong style=\"color:%s\">Expira en %d minutos.</strong>",
			colInk, minutes)) +
		codeBox("Tu código de verificación", code) +
		note("Si no creaste esta cuenta, ignora este mensaje: sin el código nadie puede activarla.")

	return Message{
		Subject: fmt.Sprintf("%s es tu código de verificación — %s", code, c.cfg.AppName),
		HTML: c.layout("Verifica tu correo",
			fmt.Sprintf("Tu código es %s. Expira en %d minutos.", code, minutes),
			body, "Este código es de un solo uso."),
	}
}

// ── Recuperación de contraseña ───────────────────────────────────────────────

// PasswordResetCode: variante con código de 6 dígitos.
func (c *Client) PasswordResetCode(code string) Message {
	body := heading("Seguridad", "Recuperar contraseña") +
		paragraph("Recibimos una solicitud para restablecer tu contraseña. Usa el siguiente código en la plataforma.") +
		codeBox("Tu código de recuperación", code) +
		infoBox(colBrand, "El código <strong>expira en 15 minutos</strong> y solo puede usarse una vez.") +
		note("Si no solicitaste este cambio, ignora este mensaje. Tu contraseña permanecerá sin cambios.")

	return Message{
		Subject: fmt.Sprintf("Código de recuperación — %s", c.cfg.AppName),
		HTML: c.layout("Recuperar contraseña",
			fmt.Sprintf("Tu código de recuperación es %s.", code),
			body, "Este código es de un solo uso y válido por 15 minutos."),
	}
}

// PasswordResetLink: variante con enlace firmado (la que usa el auth service).
func (c *Client) PasswordResetLink(name, link string) Message {
	body := heading("Seguridad", "Recuperar contraseña") +
		paragraph(fmt.Sprintf("Hola <strong style=\"color:%s\">%s</strong>,", colInk, html.EscapeString(firstName(name)))) +
		paragraph(fmt.Sprintf("Pulsa el botón para crear una contraseña nueva. <strong style=\"color:%s\">El enlace caduca en 1 hora.</strong>", colInk)) +
		button("Restablecer contraseña", link, colBrand) +
		fallbackLink(link) +
		note("Si no solicitaste este cambio, ignora este mensaje y tu contraseña seguirá igual.")

	return Message{
		Subject: fmt.Sprintf("Restablece tu contraseña — %s", c.cfg.AppName),
		HTML: c.layout("Recuperar contraseña",
			"Crea una contraseña nueva. El enlace caduca en 1 hora.",
			body, "Enlace válido por 1 hora."),
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
			cantidad = fmt.Sprintf(` <span style="color:%s;font-weight:800">× %d</span>`, colBrand, l.Cantidad)
		}
		rows.WriteString(fmt.Sprintf(`
<tr>
  <td style="padding:14px 0;border-bottom:1px solid #f1f1f3">
    <div style="font-size:15px;font-weight:700;line-height:1.4;color:%s">%s%s</div>
    <div style="font-size:13px;line-height:1.5;color:%s;margin-top:3px">%s</div>
  </td>
</tr>`, colInk, html.EscapeString(l.Titulo), cantidad, colMuted, html.EscapeString(l.Tipo)))
	}

	body := heading("Comprobante", "Pago confirmado") +
		paragraph(fmt.Sprintf("Hola <strong style=\"color:%s\">%s</strong>, tu pago se procesó correctamente. Este es el detalle de tu compra:",
			colInk, html.EscapeString(firstName(name)))) +
		fmt.Sprintf(`<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 4px">%s</table>`, rows.String()) +
		fmt.Sprintf(`
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 26px">
  <tr>
    <td style="padding:16px 0 0;font-size:14px;color:%s">Total pagado</td>
    <td style="padding:16px 0 0;font-size:21px;font-weight:800;color:%s;text-align:right;white-space:nowrap">$%s MXN</td>
  </tr>
</table>`, colMuted, colInk, formatMoney(total)) +
		button(ctaLabel, ctaURL, colSuccess) +
		note("Tu factura está disponible desde la plataforma. Si compraste licencias corporativas, recibirás un segundo correo con los accesos para tu equipo.")

	return Message{
		Subject: fmt.Sprintf("Confirmación de compra — %s", c.cfg.AppName),
		HTML: c.layout("Pago confirmado",
			fmt.Sprintf("Recibimos tu pago de $%s MXN. Gracias por tu compra.", formatMoney(total)),
			body, "Gracias por tu compra."),
	}
}

// ── Licencias corporativas ───────────────────────────────────────────────────

// CorporateLicense entrega al comprador el/los código(s) de acceso de la
// licencia recién adquirida. Sustituye la necesidad de entrar al módulo.
func (c *Client) CorporateLicense(name, cursoTitulo string, lugares int, codigoCompartido, gestionURL string) Message {
	accesos := paragraph("Consulta tus accesos desde la plataforma.")
	if codigoCompartido != "" {
		accesos = paragraph("Comparte este código con tu equipo para que se inscriban al curso:") +
			codeBox("Código de acceso", codigoCompartido)
	}

	body := heading("Licencias corporativas", "Tus accesos ya están listos") +
		paragraph(fmt.Sprintf("Hola <strong style=\"color:%s\">%s</strong>,", colInk, html.EscapeString(firstName(name)))) +
		paragraph(fmt.Sprintf("Adquiriste <strong style=\"color:%s\">%d lugar(es)</strong> para <strong style=\"color:%s\">%s</strong>.",
			colInk, lugares, colInk, html.EscapeString(cursoTitulo))) +
		accesos +
		button("Repartir accesos por correo", gestionURL, colBrand) +
		infoBox(colBrand, "¿Prefieres que los enviemos nosotros? Desde la plataforma puedes capturar los correos de tus participantes y cada uno recibirá su acceso automáticamente.") +
		note("Conserva estos códigos: cada uno es de un solo uso.")

	return Message{
		Subject: fmt.Sprintf("Accesos corporativos: %s — %s", cursoTitulo, c.cfg.AppName),
		HTML: c.layout("Tus accesos corporativos",
			fmt.Sprintf("%d lugar(es) para %s. Códigos de acceso incluidos.", lugares, cursoTitulo),
			body, "Conserva estos códigos: son de un solo uso."),
	}
}

// ParticipantAccess es el correo que recibe cada trabajador/participante con su
// código individual, enviado por el comprador desde la plataforma.
func (c *Client) ParticipantAccess(nombreParticipante, empresaONombreComprador, cursoTitulo, codigo, accesoURL string) Message {
	saludo := "Hola"
	if n := firstName(nombreParticipante); n != "" {
		saludo = fmt.Sprintf(`Hola <strong style="color:%s">%s</strong>`, colInk, html.EscapeString(n))
	}
	invitador := html.EscapeString(empresaONombreComprador)
	if invitador == "" {
		invitador = "tu organización"
	}

	body := heading("Invitación", "Te inscribieron a una capacitación") +
		paragraph(saludo+",") +
		paragraph(fmt.Sprintf(`<strong style="color:%s">%s</strong> reservó un lugar para ti en <strong style="color:%s">%s</strong>. Pulsa el botón para activar tu acceso — el código ya va incluido en el enlace.`,
			colInk, invitador, colInk, html.EscapeString(cursoTitulo))) +
		button("Activar mi acceso", accesoURL, colSuccess) +
		fallbackLink(accesoURL) +
		codeBox("Tu código de acceso", codigo) +
		note("Este código es personal e intransferible: solo puede usarse una vez.")

	return Message{
		Subject: fmt.Sprintf("Tu acceso a %s — %s", cursoTitulo, c.cfg.AppName),
		HTML: c.layout("Tu acceso a la capacitación",
			fmt.Sprintf("%s te reservó un lugar en %s. Activa tu acceso.", empresaONombreComprador, cursoTitulo),
			body, "Acceso personal e intransferible."),
	}
}

// ── DC-3 ─────────────────────────────────────────────────────────────────────

// DC3Representative avisa al representante que ya puede tramitar constancias.
func (c *Client) DC3Representative(name, cursoTitulo string, duracionHoras int, formURL string) Message {
	body := heading("Constancias", "Ya puedes tramitar los DC-3") +
		paragraph(fmt.Sprintf("Hola <strong style=\"color:%s\">%s</strong>,", colInk, html.EscapeString(firstName(name)))) +
		paragraph(fmt.Sprintf(`La capacitación <strong style="color:%s">%s</strong> ha concluido. Como representante del grupo, ya puedes generar y descargar las constancias DC-3 de tus participantes.`,
			colInk, html.EscapeString(cursoTitulo))) +
		button("Formulario para Constancias DC-3", formURL, colSuccess) +
		infoBox("#3b82f6", fmt.Sprintf("<strong>Nota:</strong> el enlace ya lleva precargados el nombre de la capacitación y la duración (%d hrs). Verifica los datos del centro de trabajo y los nombres oficiales de los participantes.", duracionHoras)) +
		note(fmt.Sprintf("Agradecemos tu preferencia.<br /><strong>El equipo de %s</strong>", html.EscapeString(c.cfg.AppName)))

	return Message{
		Subject: fmt.Sprintf("Constancias DC-3 disponibles — %s — %s", cursoTitulo, c.cfg.AppName),
		HTML: c.layout("Constancias DC-3",
			fmt.Sprintf("La capacitación %s concluyó. Genera las constancias de tu equipo.", cursoTitulo),
			body, ""),
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
