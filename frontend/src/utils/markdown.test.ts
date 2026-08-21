import { describe, it, expect } from 'vitest'
import { renderMarkdown, markdownATexto, resumenMarkdown } from './markdown'

describe('renderMarkdown · formato', () => {
  it('convierte negritas, cursivas y listas', () => {
    const html = renderMarkdown('Texto **fuerte** y *suave*\n\n- uno\n- dos')
    expect(html).toContain('<strong>fuerte</strong>')
    expect(html).toContain('<em>suave</em>')
    expect(html).toContain('<li>uno</li>')
  })

  it('separa párrafos con una línea en blanco', () => {
    expect((renderMarkdown('Uno\n\nDos').match(/<p>/g) || []).length).toBe(2)
  })

  it('un salto simple es un <br>, no una unión', () => {
    // `breaks: true`: quien escribe la descripción de un curso espera que Enter
    // haga lo que hace en cualquier editor. Sin esto, dos líneas seguidas se
    // pegaban en una sola y el texto salía como un ladrillo.
    expect(renderMarkdown('Uno\nDos')).toContain('<br>')
  })

  it('devuelve cadena vacía para entradas vacías', () => {
    for (const v of ['', '   ', null, undefined]) {
      expect(renderMarkdown(v)).toBe('')
    }
  })
})

describe('renderMarkdown · saneado', () => {
  /**
   * La descripción se muestra en la ficha PÚBLICA, sin sesión, y quien la
   * escribe es un instructor. Estos casos son la razón de que la conversión y
   * el saneado vivan en la misma función: no debe existir forma de llamar a
   * una sin la otra.
   */
  it('elimina scripts', () => {
    const html = renderMarkdown('Hola <script>alert(1)</script> mundo')
    expect(html).not.toContain('<script')
    expect(html).not.toContain('alert(1)')
  })

  it('elimina manejadores de eventos', () => {
    const html = renderMarkdown('<img src=x onerror="alert(1)">')
    expect(html.toLowerCase()).not.toContain('onerror')
  })

  it('descarta las imágenes remotas', () => {
    // Fuera de la lista blanca: permiten rastrear a quien visita la ficha y
    // reventar la maquetación con un archivo enorme.
    expect(renderMarkdown('![x](https://ejemplo.mx/a.png)')).not.toContain('<img')
  })

  it('rechaza enlaces con esquemas peligrosos', () => {
    const html = renderMarkdown('[pulsa](javascript:alert(1))')
    expect(html).not.toContain('javascript:')
  })

  it('conserva los enlaces http y les pone rel de seguridad', () => {
    const html = renderMarkdown('[MH](https://mhsolucionesempresariales.com)')
    expect(html).toContain('href="https://mhsolucionesempresariales.com"')
    // Sin `noopener`, la página destino puede manipular la nuestra por
    // window.opener.
    expect(html).toContain('rel="noopener noreferrer"')
    expect(html).toContain('target="_blank"')
  })

  it('descarta etiquetas fuera de la lista blanca', () => {
    const html = renderMarkdown('<iframe src="https://x.mx"></iframe><table><tr><td>a</td></tr></table>')
    expect(html).not.toContain('<iframe')
    expect(html).not.toContain('<table')
  })
})

describe('markdownATexto', () => {
  it('deja el texto sin marcas ni etiquetas', () => {
    expect(markdownATexto('**Hola** *mundo*\n\n- uno')).toBe('Hola mundo uno')
  })

  it('vacío para entrada vacía', () => {
    expect(markdownATexto('')).toBe('')
  })
})

describe('renderMarkdown · variante chat', () => {
  it('conserva lo que se usa al escribir', () => {
    const html = renderMarkdown('**urgente** y ~~listo~~\n\n- uno\n- dos', 'chat')
    expect(html).toContain('<strong>urgente</strong>')
    expect(html).toContain('<li>uno</li>')
  })

  it('descarta títulos y separadores', () => {
    // En una burbuja quedan fuera de lugar, y un h3 permite escribir a tamaño
    // gigante en la conversación de otra persona.
    const html = renderMarkdown('### GRITO\n\n---\n\n> cita', 'chat')
    expect(html).not.toContain('<h3')
    expect(html).not.toContain('<hr')
    expect(html).not.toContain('<blockquote')
  })

  it('sanea igual que la variante completa', () => {
    const html = renderMarkdown('<script>alert(1)</script>', 'chat')
    expect(html).not.toContain('<script')
  })

  it('un esquema peligroso nunca llega a ser un href', () => {
    // Se afirma sobre el ATRIBUTO, no sobre la cadena suelta: si marked no
    // llega a formar el enlace, "javascript:" se queda como texto escapado y
    // visible, que es inofensivo. Lo que no debe existir nunca es un <a> que
    // apunte ahí.
    const html = renderMarkdown('[pulsa](javascript:alert(1))', 'chat')
    expect(html).not.toMatch(/href=["']?javascript:/i)
  })

  it('la variante completa sí permite títulos', () => {
    expect(renderMarkdown('### Temario', 'completo')).toContain('<h3')
  })
})

describe('resumenMarkdown', () => {
  /**
   * Regresión de las tarjetas del catálogo y del subtítulo de la ficha: la
   * descripción salía en crudo, con los asteriscos y los guiones a la vista.
   */
  it('quita las marcas de Markdown', () => {
    const r = resumenMarkdown('**Subtítulo** Aprende a identificar riesgos.')
    expect(r).toBe('Subtítulo Aprende a identificar riesgos.')
    expect(r).not.toContain('*')
  })

  it('aplana listas y saltos en una sola línea', () => {
    const r = resumenMarkdown('# Temario\n\n- Arneses\n- Líneas de vida')
    expect(r).toBe('Temario Arneses Líneas de vida')
    expect(r).not.toContain('\n')
  })

  it('devuelve cadena vacía si no hay texto', () => {
    expect(resumenMarkdown('')).toBe('')
    expect(resumenMarkdown(null)).toBe('')
    expect(resumenMarkdown(undefined)).toBe('')
    expect(resumenMarkdown('   ')).toBe('')
  })

  it('no toca lo que ya cabe', () => {
    expect(resumenMarkdown('Curso corto', 220)).toBe('Curso corto')
  })

  it('recorta por palabra entera y añade puntos suspensivos', () => {
    const r = resumenMarkdown('palabra '.repeat(60), 40)
    expect(r.length).toBeLessThanOrEqual(41)
    expect(r.endsWith('…')).toBe(true)
    // Sin sílabas partidas: lo anterior a los puntos es una palabra completa.
    expect(r.slice(0, -1).trim().split(' ').pop()).toBe('palabra')
  })

  // El HTML que colara el instructor tampoco debe llegar como etiqueta.
  it('no deja etiquetas HTML', () => {
    expect(resumenMarkdown('<img src=x onerror=alert(1)>hola')).not.toContain('<')
  })
})
