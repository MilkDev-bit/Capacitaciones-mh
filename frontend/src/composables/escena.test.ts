import { describe, it, expect } from 'vitest'
import {
  limitar,
  suavizar,
  PARADAS,
  rotacionCubo,
  indiceParada,
  PALETAS,
  paletaEn,
  posicionAstro,
  opacidadEstrellas,
  css,
} from './escena'

describe('utilidades', () => {
  it('limitar recorta fuera de rango', () => {
    expect(limitar(-3)).toBe(0)
    expect(limitar(4)).toBe(1)
    expect(limitar(0.4)).toBe(0.4)
  })

  it('suavizar conserva los extremos y el punto medio', () => {
    expect(suavizar(0)).toBe(0)
    expect(suavizar(1)).toBe(1)
    expect(suavizar(0.5)).toBeCloseTo(0.5, 5)
  })
})

describe('rotación del cubo', () => {
  it('cada parada deja una cara mirando al frente', () => {
    PARADAS.forEach((parada, i) => {
      const r = rotacionCubo(i / (PARADAS.length - 1))
      expect(r.rx).toBeCloseTo(parada.rx, 4)
      expect(r.ry).toBeCloseTo(parada.ry, 4)
    })
  })

  /**
   * Lo más fácil de romper al tocar PARADAS: si el giro en Y dejara de ser
   * monótono, el cubo desharía la rotación a mitad de página y se leería como
   * un fallo, no como una transición.
   */
  it('gira siempre en el mismo sentido', () => {
    let anterior = Infinity
    for (let i = 0; i <= 100; i++) {
      const { ry } = rotacionCubo(i / 100)
      expect(ry).toBeLessThanOrEqual(anterior + 1e-9)
      anterior = ry
    }
  })

  it('completa una vuelta entera de principio a fin', () => {
    expect(rotacionCubo(0).ry).toBe(0)
    expect(rotacionCubo(1).ry).toBe(-360)
  })

  it('recorta el progreso fuera de rango en vez de extrapolar', () => {
    expect(rotacionCubo(-5)).toEqual(rotacionCubo(0))
    expect(rotacionCubo(9)).toEqual(rotacionCubo(1))
  })

  it('no revienta con listas degeneradas', () => {
    expect(rotacionCubo(0.5, [])).toEqual({ rx: 0, ry: 0 })
    expect(rotacionCubo(0.5, [{ rx: 12, ry: 34 }])).toEqual({ rx: 12, ry: 34 })
  })
})

describe('índice de paso', () => {
  it('asigna un paso a cada tramo del recorrido', () => {
    expect(indiceParada(0, 6)).toBe(0)
    expect(indiceParada(0.5, 6)).toBe(3)
    expect(indiceParada(1, 6)).toBe(5)
  })

  it('nunca se sale del rango', () => {
    for (const s of [-2, -0.001, 1.001, 50]) {
      const i = indiceParada(s, 6)
      expect(i).toBeGreaterThanOrEqual(0)
      expect(i).toBeLessThanOrEqual(5)
    }
    expect(indiceParada(0.5, 0)).toBe(0)
  })
})

describe('ciclo día → noche', () => {
  it('los extremos coinciden con los fotogramas clave', () => {
    expect(paletaEn(0).cieloArriba).toEqual(PALETAS[0]!.cieloArriba)
    expect(paletaEn(1).cieloAbajo).toEqual(PALETAS[2]!.cieloAbajo)
  })

  it('el punto medio cae en el atardecer', () => {
    expect(paletaEn(0.5).cieloAbajo).toEqual(PALETAS[1]!.cieloAbajo)
  })

  // El cielo tiene que oscurecer de forma continua: un repunte de luminosidad a
  // mitad del recorrido delataría una interpolación mal encadenada.
  it('el cielo solo oscurece', () => {
    const luz = (c: readonly [number, number, number]) => c[0] + c[1] + c[2]
    let anterior = Infinity
    for (let i = 0; i <= 60; i++) {
      const actual = luz(paletaEn(i / 60).cieloArriba)
      expect(actual).toBeLessThanOrEqual(anterior + 1e-9)
      anterior = actual
    }
  })

  it('devuelve siempre cinco planos de colina con canales válidos', () => {
    for (const s of [0, 0.23, 0.5, 0.77, 1]) {
      const p = paletaEn(s)
      expect(p.colinas).toHaveLength(5)
      for (const c of p.colinas) {
        for (const canal of c) {
          expect(canal).toBeGreaterThanOrEqual(0)
          expect(canal).toBeLessThanOrEqual(255)
          expect(Number.isInteger(canal)).toBe(true)
        }
      }
    }
  })

  it('serializa a rgb() válido', () => {
    expect(css([15, 16, 48])).toBe('rgb(15, 16, 48)')
  })

  it('con un solo fotograma devuelve ese mismo', () => {
    expect(paletaEn(0.7, [PALETAS[1]!])).toBe(PALETAS[1]!)
  })
})

describe('astro y estrellas', () => {
  it('el astro cruza el cielo de izquierda a derecha', () => {
    expect(posicionAstro(0).x).toBeLessThan(posicionAstro(0.5).x)
    expect(posicionAstro(0.5).x).toBeLessThan(posicionAstro(1).x)
  })

  it('sube y luego se pone', () => {
    const inicio = posicionAstro(0).y
    const cima = posicionAstro(0.12).y
    const final = posicionAstro(1).y
    // Menor y = más alto en pantalla.
    expect(cima).toBeLessThan(inicio)
    expect(final).toBeGreaterThan(cima)
  })

  it('mantiene el astro dentro del encuadre', () => {
    for (let i = 0; i <= 40; i++) {
      const { y } = posicionAstro(i / 40)
      expect(y).toBeGreaterThanOrEqual(8)
      expect(y).toBeLessThanOrEqual(96)
    }
  })

  it('las estrellas solo salen con el cielo ya oscuro', () => {
    expect(opacidadEstrellas(0)).toBe(0)
    expect(opacidadEstrellas(0.55)).toBe(0)
    expect(opacidadEstrellas(0.7)).toBeGreaterThan(0)
    expect(opacidadEstrellas(0.7)).toBeLessThan(1)
    expect(opacidadEstrellas(1)).toBe(1)
  })
})
