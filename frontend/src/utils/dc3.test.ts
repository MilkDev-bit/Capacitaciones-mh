import { describe, it, expect } from 'vitest'
import { estadoDC3, cursoDelCatalogo, AREAS_TEMATICAS_DC3, type EntradaEstadoDC3 } from './dc3'

/** Alumno recién llegado: consultado, sin nada guardado. */
const base: EntradaEstadoDC3 = {
  consultado: true,
  constanciaUrl: '',
  trabajadorCompleto: false,
  empresaCompleta: false,
  editando: false,
}

describe('estadoDC3', () => {
  it('espera a la respuesta antes de decidir nada', () => {
    // Sin esto, el alumno ve un parpadeo de "captura tus datos" en cada carga,
    // incluso cuando su constancia ya existe.
    expect(estadoDC3({ ...base, consultado: false })).toBe('cargando')
    expect(estadoDC3({ ...base, consultado: false, constanciaUrl: 'https://r2/x.docx' }))
      .toBe('cargando')
  })

  it('pide los datos del alumno cuando faltan', () => {
    expect(estadoDC3(base)).toBe('faltan-mios')
  })

  it('ofrece la descarga en cuanto hay documento', () => {
    expect(estadoDC3({
      ...base,
      constanciaUrl: 'https://r2/constancia.docx',
      trabajadorCompleto: true,
      empresaCompleta: true,
    })).toBe('lista')
  })

  it('señala al instructor solo cuando la empresa está incompleta', () => {
    expect(estadoDC3({ ...base, trabajadorCompleto: true })).toBe('falta-empresa')
  })

  /**
   * Regresión del fallo reportado.
   *
   * Con los datos del alumno y de la empresa completos pero sin documento, el
   * panel decía "falta que el instructor complete los datos de la empresa".
   * Culpaba a un tercero que ya había cumplido y dejaba al alumno sin ninguna
   * acción posible: no había reintento ni forma de volver al formulario.
   */
  it('no culpa al instructor cuando la empresa ya está completa', () => {
    const estado = estadoDC3({
      ...base,
      trabajadorCompleto: true,
      empresaCompleta: true,
    })
    expect(estado).toBe('sin-emitir')
    expect(estado).not.toBe('falta-empresa')
  })

  /**
   * Regresión: una CURP mal tecleada quedaba congelada.
   *
   * El formulario solo aparecía mientras `trabajadorCompleto` fuera falso, así
   * que en cuanto se guardaba una vez no había forma de volver a él desde
   * ninguno de los tres estados finales.
   */
  it('devuelve el formulario desde cualquier estado cuando se pide corregir', () => {
    const finales = [
      { ...base, trabajadorCompleto: true, empresaCompleta: true, constanciaUrl: 'https://r2/c.docx' },
      { ...base, trabajadorCompleto: true, empresaCompleta: true },
      { ...base, trabajadorCompleto: true },
    ]
    for (const estado of finales) {
      expect(estadoDC3({ ...estado, editando: true })).toBe('faltan-mios')
    }
  })

  it('no reabre el formulario mientras aún se está consultando', () => {
    expect(estadoDC3({ ...base, consultado: false, editando: true })).toBe('cargando')
  })
})

describe('cursoDelCatalogo', () => {
  it('encuentra el curso ignorando mayúsculas y espacios sobrantes', () => {
    const c = cursoDelCatalogo('  manejo seguro de montacargas  ')
    expect(c?.horas).toBe(6)
    expect(c?.area).toBe('6000')
  })

  it('devuelve undefined para un curso ajeno al catálogo', () => {
    expect(cursoDelCatalogo('Curso inventado')).toBeUndefined()
  })

  it('todas las áreas del catálogo existen en el listado oficial', () => {
    // Un área inventada haría que la STPS rechazara la constancia, y el error
    // solo se vería en el documento final.
    const claves = new Set(AREAS_TEMATICAS_DC3.map(a => a.clave))
    for (const curso of [
      cursoDelCatalogo('ISO 14001-2015 SISTEMA DE GESTION AMBIENTAL'),
      cursoDelCatalogo('SEGURIDAD EN EXCAVACIONES Y ZANJAS'),
    ]) {
      expect(claves.has(curso!.area)).toBe(true)
    }
  })
})
