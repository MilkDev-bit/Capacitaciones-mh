/**
 * Matemática de la escena de "Cómo funciona".
 *
 * Todo lo que se puede calcular sin tocar el DOM vive aquí: así la rotación del
 * cubo y el ciclo día-noche se prueban con vitest en lugar de comprobarse a ojo
 * haciendo scroll, que es como estas escenas se rompen sin que nadie se entere.
 *
 * La única entrada es `s`, el progreso de scroll normalizado a 0–1.
 */

// ── Utilidades ──────────────────────────────────────────────────────────────

export function limitar(v: number, min = 0, max = 1) {
  return Math.min(max, Math.max(min, v))
}

/** Suavizado simétrico: arranca y frena despacio, constante en el medio. */
export function suavizar(t: number) {
  return t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t
}

// ── Cubo ────────────────────────────────────────────────────────────────────

export interface Parada {
  rx: number
  ry: number
}

/**
 * Una parada por cara, en el orden en que las caras aparecen en el DOM:
 * arriba, frente, derecha, atrás, izquierda, abajo.
 *
 * El giro en Y es siempre negativo y creciente para que el cubo gire SIEMPRE
 * en el mismo sentido. Si alguna parada volviera a 0 el cubo desharía el giro
 * a mitad de la página y se leería como un error, no como una transición.
 */
export const PARADAS: readonly Parada[] = [
  { rx: 90, ry: 0 },
  { rx: 0, ry: 0 },
  { rx: 0, ry: -90 },
  { rx: 0, ry: -180 },
  { rx: 0, ry: -270 },
  { rx: -90, ry: -360 },
]

/** Rotación del cubo para un progreso dado, interpolando entre paradas. */
export function rotacionCubo(s: number, paradas: readonly Parada[] = PARADAS): Parada {
  const n = paradas.length
  if (n === 0) return { rx: 0, ry: 0 }
  if (n === 1) return { ...(paradas[0] as Parada) }

  const t = limitar(s) * (n - 1)
  const i = Math.min(Math.floor(t), n - 2)
  const f = suavizar(t - i)
  const a = paradas[i] as Parada
  const b = paradas[i + 1] as Parada
  return {
    rx: a.rx + (b.rx - a.rx) * f,
    ry: a.ry + (b.ry - a.ry) * f,
  }
}

/** Índice de la cara/sección que corresponde a un progreso dado. */
export function indiceParada(s: number, total: number) {
  if (total <= 0) return 0
  return Math.min(total - 1, Math.max(0, Math.round(limitar(s) * (total - 1))))
}

// ── Ciclo día → atardecer → noche ───────────────────────────────────────────

export type RGB = readonly [number, number, number]

export interface Paleta {
  cieloArriba: RGB
  cieloAbajo: RGB
  astro: RGB
  /** Cinco planos de colina, del más lejano al más cercano. */
  colinas: readonly RGB[]
}

/**
 * Tres fotogramas clave. El del medio es el atardecer: sin él, pasar de día a
 * noche interpolando en línea recta atraviesa un gris sucio en vez de un cielo.
 */
export const PALETAS: readonly Paleta[] = [
  {
    cieloArriba: [142, 197, 255],
    cieloAbajo: [255, 216, 168],
    astro: [245, 197, 78],
    colinas: [
      [217, 137, 129],
      [196, 105, 118],
      [145, 77, 100],
      [112, 55, 90],
      [67, 61, 108],
    ],
  },
  {
    cieloArriba: [122, 75, 124],
    cieloAbajo: [240, 117, 95],
    astro: [255, 145, 113],
    colinas: [
      [138, 84, 107],
      [110, 62, 96],
      [86, 58, 106],
      [61, 46, 84],
      [43, 40, 80],
    ],
  },
  {
    cieloArriba: [15, 16, 48],
    cieloAbajo: [44, 35, 80],
    astro: [232, 230, 245],
    colinas: [
      [56, 50, 99],
      [43, 40, 80],
      [32, 30, 62],
      [23, 20, 44],
      [14, 10, 26],
    ],
  },
]

function mezclarRGB(a: RGB, b: RGB, t: number): RGB {
  return [
    Math.round(a[0] + (b[0] - a[0]) * t),
    Math.round(a[1] + (b[1] - a[1]) * t),
    Math.round(a[2] + (b[2] - a[2]) * t),
  ]
}

export function css(c: RGB) {
  return `rgb(${c[0]}, ${c[1]}, ${c[2]})`
}

/** Paleta interpolada para un progreso dado. */
export function paletaEn(s: number, paletas: readonly Paleta[] = PALETAS): Paleta {
  const n = paletas.length
  const primera = paletas[0] as Paleta
  if (n === 1) return primera

  const t = limitar(s) * (n - 1)
  const i = Math.min(Math.floor(t), n - 2)
  const f = t - i
  const a = paletas[i] as Paleta
  const b = paletas[i + 1] as Paleta

  return {
    cieloArriba: mezclarRGB(a.cieloArriba, b.cieloArriba, f),
    cieloAbajo: mezclarRGB(a.cieloAbajo, b.cieloAbajo, f),
    astro: mezclarRGB(a.astro, b.astro, f),
    colinas: a.colinas.map((c, k) => mezclarRGB(c, (b.colinas[k] ?? c) as RGB, f)),
  }
}

/**
 * Posición del sol/luna en porcentaje del viewBox.
 *
 * Describe un arco: sube, cruza y se pone. La `y` usa una parábola porque un
 * descenso lineal haría que el astro tocara el horizonte demasiado pronto y el
 * cielo se quedara vacío durante media página.
 */
export function posicionAstro(s: number) {
  const p = limitar(s)
  const x = 16 + p * 68
  // Vértice en p≈0.42: el punto más alto cae poco después del inicio, como
  // en un atardecer real observado desde media tarde.
  const y = 14 + 74 * Math.pow(p - 0.12, 2) * 1.35
  return { x, y: limitar(y, 8, 96) }
}

/** Opacidad de las estrellas: aparecen solo cuando el cielo ya oscureció. */
export function opacidadEstrellas(s: number) {
  return limitar((limitar(s) - 0.55) / 0.3)
}
