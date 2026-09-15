import { describe, it, expect } from 'vitest'
import {
  preguntasQuitadas,
  respuestasQueSePierden,
  avisoDeBorrado,
  type PreguntaOriginal,
} from './examen'

const ORIGINALES: PreguntaOriginal[] = [
  { id: 'p1', texto: '¿Qué es un arnés?', respuestas: 12 },
  { id: 'p2', texto: 'Altura mínima', respuestas: 0 },
  { id: 'p3', texto: 'Uso del casco', respuestas: 5 },
]

describe('preguntasQuitadas', () => {
  it('no reporta nada si no se quitó ninguna', () => {
    const actuales = [{ id: 'p1' }, { id: 'p2' }, { id: 'p3' }]
    expect(preguntasQuitadas(ORIGINALES, actuales)).toEqual([])
  })

  it('detecta la pregunta con respuestas que desapareció', () => {
    const actuales = [{ id: 'p2' }, { id: 'p3' }]
    expect(preguntasQuitadas(ORIGINALES, actuales).map(p => p.id)).toEqual(['p1'])
  })

  /**
   * Quitar una pregunta que nadie contestó no destruye nada. Avisar de eso
   * acostumbra al instructor a aceptar el diálogo sin leerlo, y entonces el
   * aviso deja de servir justo cuando sí hay algo en juego.
   */
  it('ignora las preguntas sin respuestas', () => {
    const actuales = [{ id: 'p1' }, { id: 'p3' }]
    expect(preguntasQuitadas(ORIGINALES, actuales)).toEqual([])
  })

  /** Una pregunta nueva no tiene id: no debe confundirse con una conservada. */
  it('las preguntas nuevas no salvan a las borradas', () => {
    const actuales = [{ id: 'p2' }, {}, {}]
    expect(preguntasQuitadas(ORIGINALES, actuales).map(p => p.id)).toEqual(['p1', 'p3'])
  })

  it('vaciar el examen reporta todas las que tenían respuestas', () => {
    expect(preguntasQuitadas(ORIGINALES, []).map(p => p.id)).toEqual(['p1', 'p3'])
  })

  it('un examen que nadie ha presentado nunca avisa', () => {
    const sinRespuestas = ORIGINALES.map(p => ({ ...p, respuestas: 0 }))
    expect(preguntasQuitadas(sinRespuestas, [])).toEqual([])
  })
})

describe('respuestasQueSePierden', () => {
  it('suma las respuestas de todas las quitadas', () => {
    expect(respuestasQueSePierden(preguntasQuitadas(ORIGINALES, []))).toBe(17)
  })

  it('es cero cuando no se quita nada', () => {
    expect(respuestasQueSePierden([])).toBe(0)
  })
})

describe('avisoDeBorrado', () => {
  /**
   * El número de respuestas es el dato que hace dudar: quitar UNA pregunta
   * puede borrar doce respuestas, y sin esa cifra el aviso parece menor de lo
   * que es.
   */
  it('nombra las preguntas y las respuestas por separado', () => {
    const texto = avisoDeBorrado(preguntasQuitadas(ORIGINALES, [{ id: 'p2' }, { id: 'p3' }]))
    expect(texto).toContain('1 pregunta que ya fue respondida')
    expect(texto).toContain('12 respuestas de estudiantes')
  })

  it('concuerda en plural con varias preguntas', () => {
    const texto = avisoDeBorrado(preguntasQuitadas(ORIGINALES, []))
    expect(texto).toContain('2 preguntas que ya fueron respondidas')
    expect(texto).toContain('17 respuestas de estudiantes')
  })

  it('concuerda en singular con una sola respuesta', () => {
    const texto = avisoDeBorrado([{ id: 'x', texto: 'Única', respuestas: 1 }])
    expect(texto).toContain('1 pregunta que ya fue respondida')
    expect(texto).toContain('1 respuesta de estudiantes')
  })

  it('advierte de que no se puede deshacer', () => {
    expect(avisoDeBorrado(preguntasQuitadas(ORIGINALES, []))).toContain('no se pueden recuperar')
  })
})
