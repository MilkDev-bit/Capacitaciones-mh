import { describe, it, expect } from 'vitest'
import { listaDe } from './lista'

describe('listaDe', () => {
  it('saca el arreglo de la clave indicada', () => {
    expect(listaDe({ entregas: [1, 2] }, 'entregas')).toEqual([1, 2])
  })

  it('acepta una respuesta que ya es un arreglo', () => {
    expect(listaDe([1, 2], 'entregas')).toEqual([1, 2])
    expect(listaDe([1, 2])).toEqual([1, 2])
  })

  /**
   * Regresión de la pantalla de Entregas en blanco.
   *
   * protobuf-go serializa un `repeated` vacío borrando el campo, así que la
   * respuesta llega como `{}`. Con el idioma anterior —`data?.entregas ||
   * data || []`— ese objeto vacío es truthy, ganaba a `[]`, y la vista
   * reventaba al llamar a `.filter()` sobre un objeto.
   */
  it('devuelve un arreglo cuando la respuesta es un objeto vacío', () => {
    const r = listaDe({}, 'entregas')
    expect(Array.isArray(r)).toBe(true)
    expect(r).toEqual([])
    // Lo que rompía la vista: esto debe poder llamarse sin lanzar.
    expect(() => r.filter(Boolean)).not.toThrow()
    expect(r.length).toBe(0)
  })

  it('devuelve un arreglo cuando la clave viene nula o no es lista', () => {
    expect(listaDe({ entregas: null }, 'entregas')).toEqual([])
    expect(listaDe({ entregas: {} }, 'entregas')).toEqual([])
    expect(listaDe({ entregas: 'texto' }, 'entregas')).toEqual([])
  })

  it('devuelve un arreglo ante ausencia de datos', () => {
    expect(listaDe(null, 'entregas')).toEqual([])
    expect(listaDe(undefined, 'entregas')).toEqual([])
    expect(listaDe(0, 'entregas')).toEqual([])
    expect(listaDe('', 'entregas')).toEqual([])
  })

  // Deja constancia de por qué el idioma anterior fallaba, para que no vuelva.
  it('un objeto vacío es truthy: por eso no basta con ||', () => {
    expect(Boolean({})).toBe(true)
    const alaAntigua = (undefined as any) || {} || []
    expect(Array.isArray(alaAntigua)).toBe(false)
  })
})
