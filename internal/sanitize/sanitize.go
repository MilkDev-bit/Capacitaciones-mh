// Package sanitize provee utilidades para limpiar HTML de entrada del usuario
// antes de persistirlo en base de datos (defensa contra XSS en el backend).
package sanitize

import "github.com/microcosm-cc/bluemonday"

// ugc permite etiquetas de formato seguro: b, i, u, a, ul, ol, li,
// p, br, strong, em, h1-h6, blockquote, code, pre.
// Elimina silenciosamente <script>, <iframe>, atributos on*, javascript:, etc.
var ugc = bluemonday.UGCPolicy()

// strict elimina absolutamente todo HTML, dejando solo texto plano.
var strict = bluemonday.StrictPolicy()

// HTML limpia HTML de contenido enriquecido (posts de foro, descripciones).
// Permite un subconjunto seguro de etiquetas de formato.
func HTML(s string) string {
	return ugc.Sanitize(s)
}

// Text elimina todo HTML y retorna texto plano.
// Úsalo en campos como títulos o nombres que no deberían contener HTML.
func Text(s string) string {
	return strict.Sanitize(s)
}
