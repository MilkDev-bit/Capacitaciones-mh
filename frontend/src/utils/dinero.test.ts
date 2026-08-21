import { describe, it, expect } from 'vitest'
import { pesos, entero, porcentaje, mesCorto, fechaHora } from './dinero'

describe('pesos', () => {
  it('convierte centavos a pesos con dos decimales', () => {
    expect(pesos(123456)).toMatch(/1,234\.56/)
    expect(pesos(1)).toMatch(/0\.01/)
  })

  it('omite los decimales cuando se pide', () => {
    const t = pesos(123456, false)
    expect(t).toMatch(/1,235/)
    expect(t).not.toMatch(/\./)
  })

  it('trata la ausencia de importe como cero', () => {
    expect(pesos(0)).toMatch(/0\.00/)
    expect(pesos(null)).toMatch(/0\.00/)
    expect(pesos(undefined)).toMatch(/0\.00/)
  })

  /**
   * Se suma en centavos, como enteros, y solo se divide al formatear. Mientras
   * el total quepa en un entero seguro de JS la suma es exacta, que es lo que
   * necesita el total anual de la pantalla de dirección.
   */
  it('formatea sumas grandes sin desviarse', () => {
    const centavos = [1, 2, 1013, 2027, 3041, 999]
    const total = centavos.reduce((a, b) => a + b, 0)
    expect(total).toBe(7083)
    expect(Number.isSafeInteger(total)).toBe(true)
    expect(pesos(total)).toMatch(/70\.83/)

    // Un año de ventas de siete cifras sigue siendo exacto.
    expect(pesos(123_456_789)).toMatch(/1,234,567\.89/)
  })
})

describe('porcentaje', () => {
  it('calcula la proporción', () => {
    expect(porcentaje(25, 100)).toBe(25)
    expect(porcentaje(1, 3)).toBeCloseTo(33.333, 2)
  })

  // El primer día no hay ventas: el anillo debe quedarse a cero, no en NaN.
  it('devuelve cero cuando el total es cero', () => {
    expect(porcentaje(10, 0)).toBe(0)
    expect(porcentaje(0, 0)).toBe(0)
  })

  // Un valor fuera de rango deja el anillo dando más de una vuelta.
  it('acota entre 0 y 100', () => {
    expect(porcentaje(150, 100)).toBe(100)
    expect(porcentaje(-50, 100)).toBe(0)
  })

  it('no propaga infinitos ni NaN', () => {
    expect(Number.isFinite(porcentaje(Infinity, 100))).toBe(true)
    expect(porcentaje(NaN, 100)).toBe(0)
  })
})

describe('mesCorto', () => {
  it('convierte AAAA-MM a mes y año', () => {
    expect(mesCorto('2026-08')).toMatch(/2026/)
    expect(mesCorto('2026-08').toLowerCase()).toContain('ago')
  })

  /**
   * Regresión de zona horaria: construyendo la fecha a medianoche UTC, en
   * México (UTC-6) el día 1 se convierte en el 30 del mes anterior y la
   * etiqueta del gráfico sale corrida un mes.
   */
  it('no se corre al mes anterior', () => {
    expect(mesCorto('2026-01').toLowerCase()).toContain('ene')
    expect(mesCorto('2026-01')).toContain('2026')
  })

  it('devuelve la entrada si no tiene el formato esperado', () => {
    expect(mesCorto('')).toBe('')
    expect(mesCorto('basura')).toBe('basura')
  })
})

describe('fechaHora', () => {
  it('formatea un ISO', () => {
    expect(fechaHora('2026-08-21T14:30:00Z')).toMatch(/2026/)
  })

  it('devuelve vacío si la fecha no sirve', () => {
    expect(fechaHora('')).toBe('')
    expect(fechaHora('no-es-fecha')).toBe('')
  })
})

describe('entero', () => {
  /**
   * Regresión de las tarjetas del panel: se pintaba el número en crudo y a
   * partir de cuatro cifras dejaba de leerse de un vistazo.
   */
  it('separa los miles', () => {
    expect(entero(15239)).toBe('15,239')
    expect(entero(1234567)).toBe('1,234,567')
  })

  it('no toca los números cortos', () => {
    expect(entero(5)).toBe('5')
    expect(entero(999)).toBe('999')
  })

  it('trata la ausencia de valor como cero', () => {
    expect(entero(0)).toBe('0')
    expect(entero(null)).toBe('0')
    expect(entero(undefined)).toBe('0')
  })
})

describe('formato de importes en las tarjetas', () => {
  /**
   * El panel hacía `$${v.toFixed(2)}` y salía "$15239.12": sin separador de
   * miles, sin locale y con el símbolo pegado a mano.
   */
  it('pone separador de miles donde antes no lo había', () => {
    const alaAntigua = `$${(15239.12).toFixed(2)}`
    expect(alaAntigua).toBe('$15239.12')

    const ahora = pesos(Math.round(15239.12 * 100))
    expect(ahora).toContain('15,239.12')
    expect(ahora).not.toBe(alaAntigua)
  })
})
