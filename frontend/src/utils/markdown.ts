/**
 * Markdown a HTML para los textos largos que escribe el instructor.
 *
 * Se usa en la descripción del curso y en el mensaje de bienvenida: campos que
 * antes se interpolaban como texto plano, así que ni los saltos de línea
 * sobrevivían y todo salía como un único bloque.
 *
 * El saneado NO es opcional ni un extra defensivo: la descripción se muestra en
 * la ficha PÚBLICA del curso, sin sesión, y quien la escribe es un instructor,
 * no un administrador. Sin sanear, un `<script>` o un `onerror=` en ese campo se
 * ejecutaría en el navegador de cualquier visitante. Por eso la conversión y el
 * saneado viven juntos en esta única función: no hay forma de llamar a una sin
 * la otra.
 */
import { marked } from 'marked'
import DOMPurify from 'dompurify'

/**
 * Etiquetas permitidas en la salida.
 *
 * Es una lista blanca corta a propósito. Una descripción de curso necesita
 * párrafos, énfasis, listas y poco más; admitir tablas o imágenes remotas
 * añadiría superficie de ataque y roturas de maquetación a cambio de casi nada.
 *
 * Sin `img`: una imagen remota permite rastrear a quien visita la ficha y sacar
 * la maquetación de quicio con un archivo enorme.
 */
const ETIQUETAS = [
  'p', 'br', 'strong', 'em', 'u', 's',
  'ul', 'ol', 'li',
  'h3', 'h4',
  'blockquote', 'code', 'pre',
  'a', 'hr',
]

/**
 * Subconjunto para burbujas de chat.
 *
 * Sin títulos, sin separadores y sin bloques de cita: en un mensaje corto
 * quedan fuera de lugar, y un `h3` permite además escribir a tamaño gigante en
 * una conversación ajena. Lo que sí se conserva es lo que la gente usa de
 * verdad al escribir: énfasis, tachado, listas y enlaces.
 */
const ETIQUETAS_CHAT = [
  'p', 'br', 'strong', 'em', 'u', 's',
  'ul', 'ol', 'li',
  'code',
  'a',
]

/** Solo lo imprescindible para los enlaces. */
const ATRIBUTOS = ['href', 'title', 'target', 'rel']

marked.setOptions({
  // Un salto simple es un <br>. Quien escribe la descripción de un curso espera
  // que Enter haga lo que hace en cualquier editor, no el comportamiento
  // estándar de Markdown, que exige una línea en blanco para separar.
  breaks: true,
  gfm: true,
})

/**
 * Los enlaces salen siempre con `rel="noopener noreferrer"`.
 *
 * `target="_blank"` sin `noopener` deja que la página destino manipule la
 * nuestra a través de `window.opener`. Se aplica DESPUÉS de sanear, sobre el
 * DOM ya limpio, para que no dependa de lo que escribiera el instructor.
 */
DOMPurify.addHook('afterSanitizeAttributes', (nodo) => {
  if (nodo instanceof HTMLAnchorElement && nodo.hasAttribute('href')) {
    nodo.setAttribute('target', '_blank')
    nodo.setAttribute('rel', 'noopener noreferrer')
  }
})

/**
 * Convierte Markdown a HTML seguro para insertar con v-html.
 *
 * Devuelve cadena vacía para una entrada vacía, de modo que quien llama pueda
 * decidir si muestra un texto por defecto.
 */
export function renderMarkdown(
  texto: string | null | undefined,
  variante: 'completo' | 'chat' = 'completo',
): string {
  const fuente = (texto || '').trim()
  if (!fuente) return ''

  // `marked.parse` es síncrono con estas opciones, pero su firma admite promesa;
  // se fuerza a string para que el tipo no se propague a las vistas.
  const html = marked.parse(fuente, { async: false }) as string

  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: variante === 'chat' ? ETIQUETAS_CHAT : ETIQUETAS,
    ALLOWED_ATTR: ATRIBUTOS,
    // Los esquemas peligrosos (`javascript:`, `data:`) quedan fuera.
    ALLOWED_URI_REGEXP: /^(?:https?:|mailto:|#|\/)/i,
  })
}

/**
 * Texto plano equivalente, para recortes y metadatos.
 *
 * Las tarjetas del catálogo y la etiqueta <meta name="description"> necesitan
 * una versión sin marcas: pintar ahí el Markdown en crudo dejaría asteriscos y
 * guiones a la vista.
 */
export function markdownATexto(texto: string | null | undefined): string {
  const html = renderMarkdown(texto)
  if (!html) return ''
  const tmp = document.createElement('div')
  tmp.innerHTML = html
  return (tmp.textContent || '').replace(/\s+/g, ' ').trim()
}
