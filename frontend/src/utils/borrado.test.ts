import { describe, it, expect } from 'vitest'
import { puedeBorrarParaTodos, VENTANA_BORRADO_MS, type MensajeBorrable } from './borrado'

const YO = 'u-yo'
const AHORA = new Date('2026-09-11T12:00:00Z').getTime()

function mensaje(extra: Partial<MensajeBorrable> = {}): MensajeBorrable {
  return {
    emisor_id: YO,
    created_at: new Date(AHORA - 60_000).toISOString(), // hace un minuto
    ...extra,
  }
}

describe('puedeBorrarParaTodos', () => {
  it('permite retirar lo que acabo de escribir', () => {
    expect(puedeBorrarParaTodos(mensaje(), YO, AHORA)).toBe(true)
  })

  /**
   * Retirar el mensaje de otra persona sería reescribir su conversación. El
   * servidor lo rechaza; esto evita además enseñar el botón.
   */
  it('no permite retirar el mensaje de otra persona', () => {
    expect(puedeBorrarParaTodos(mensaje({ emisor_id: 'u-otro' }), YO, AHORA)).toBe(false)
  })

  it('no ofrece nada a una sesión sin usuario', () => {
    expect(puedeBorrarParaTodos(mensaje(), undefined, AHORA)).toBe(false)
  })

  /**
   * El límite exacto. Se prueba a los dos lados porque es justo el punto donde
   * el frontend y `ventanaBorrado` del servicio Go tienen que coincidir: si se
   * separan, el menú ofrece una opción que devuelve error.
   */
  it('deja de ofrecerlo al cumplirse la ventana', () => {
    const justoDentro = mensaje({
      created_at: new Date(AHORA - VENTANA_BORRADO_MS + 1000).toISOString(),
    })
    const justoFuera = mensaje({
      created_at: new Date(AHORA - VENTANA_BORRADO_MS - 1000).toISOString(),
    })
    expect(puedeBorrarParaTodos(justoDentro, YO, AHORA)).toBe(true)
    expect(puedeBorrarParaTodos(justoFuera, YO, AHORA)).toBe(false)
  })

  it('la ventana son sesenta minutos', () => {
    expect(VENTANA_BORRADO_MS).toBe(3_600_000)
  })

  it('no reofrece borrar lo ya borrado', () => {
    expect(puedeBorrarParaTodos(mensaje({ eliminado: true }), YO, AHORA)).toBe(false)
  })

  /**
   * Un mensaje en vuelo todavía no tiene id en el servidor: el DELETE iría
   * contra un identificador temporal que allí no existe.
   */
  it('no ofrece borrar un mensaje que aún se está enviando', () => {
    expect(puedeBorrarParaTodos(mensaje({ _status: 'sending', _tempId: 'tmp-1' }), YO, AHORA)).toBe(false)
    expect(puedeBorrarParaTodos(mensaje({ _status: 'error', _tempId: 'tmp-1' }), YO, AHORA)).toBe(false)
    expect(puedeBorrarParaTodos(mensaje({ _status: 'sent' }), YO, AHORA)).toBe(true)
  })

  /**
   * Una fecha ilegible da NaN y toda comparación con NaN es falsa, así que el
   * caso ya salía bien por accidente. Se fija con una prueba para que siga
   * saliendo bien si alguien cambia la comparación por otra cosa.
   */
  it('descarta un mensaje con fecha ilegible', () => {
    expect(puedeBorrarParaTodos(mensaje({ created_at: 'no-es-fecha' }), YO, AHORA)).toBe(false)
  })
})
