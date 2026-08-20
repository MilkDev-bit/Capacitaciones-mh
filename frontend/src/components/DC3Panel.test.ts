import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import DC3Panel from './DC3Panel.vue'

vi.mock('../api', () => ({
  default: { get: vi.fn(), post: vi.fn() },
}))
vi.mock('../utils/toast', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

import api from '../api'

const get = api.get as unknown as ReturnType<typeof vi.fn>

/** Alumno con el curso al 100% que aún no ha capturado nada. */
const sinDatos = {
  constancia_url: '',
  trabajador_completo: false,
  empresa_completa: true,
  trabajador: null,
  empresa: { razon_social: 'MH SOLUCIONES EMPRESARIALES' },
  empresa_origen: 'instructor',
}

describe('DC3Panel', () => {
  beforeEach(() => vi.clearAllMocks())

  /**
   * Regresión del modal en blanco.
   *
   * Dentro del modal se monta con `completado` en true y sin `habilitado`. Si
   * el panel no pinta el formulario, el alumno no tiene forma de capturar su
   * CURP, y sin CURP el backend nunca emite la constancia: el modal vacío
   * bloqueaba toda la generación de DC-3.
   */
  it('pinta el formulario cuando faltan los datos del alumno', async () => {
    get.mockResolvedValue({ data: sinDatos })

    const w = mount(DC3Panel, {
      props: { cursoId: 'abc', completado: true, variante: 'plano' },
    })
    await flushPromises()

    expect(w.text()).not.toBe('')
    expect(w.find('input').exists()).toBe(true)
  })

  it('no se queda en blanco cuando la consulta falla', async () => {
    // Un curso sin DC-3 responde con error. Aun así el alumno debe ver algo:
    // un panel vacío es indistinguible de una pantalla rota.
    get.mockRejectedValue(new Error('500'))

    const w = mount(DC3Panel, {
      props: { cursoId: 'abc', completado: true, variante: 'plano' },
    })
    await flushPromises()

    expect(w.text()).not.toBe('')
  })

  /**
   * Regresión: el 202 no puede inventarse la causa.
   *
   * El panel ponía `empresaCompleta = false` al recibir un 202, sin haber
   * comprobado nada de la empresa. Bastaba con que faltara la duración del
   * curso para que acusara al instructor de no haber capturado unos datos que
   * sí tenía. Ahora el 202 obliga a repreguntar al servidor y muestra la lista
   * que este devuelve.
   */
  it('no acusa al instructor cuando el 202 se debe a otra cosa', async () => {
    get.mockResolvedValue({ data: { ...sinDatos, trabajador_completo: false } })
    const post = api.post as unknown as ReturnType<typeof vi.fn>
    post.mockResolvedValue({
      status: 202,
      data: { guardado: true, faltan: ['duración en horas'], mensaje: 'Falta: duración en horas.' },
    })

    const w = mount(DC3Panel, {
      props: { cursoId: 'abc', completado: true, variante: 'plano' },
    })
    await flushPromises()

    // Tras guardar, el servidor ya reporta al trabajador como completo.
    get.mockResolvedValue({ data: { ...sinDatos, trabajador_completo: true } })

    const [curp, puesto] = w.findAll('input')
    await curp!.setValue('ABCD901031HDFXYZ01')
    await puesto!.setValue('Supervisor')
    // La ocupación es un selector del Catálogo Nacional de Ocupaciones, no
    // texto libre: la casilla del formato oficial pide la clave.
    await w.find('select').setValue('04.6')
    await w.find('button.btn-primary').trigger('click')
    await flushPromises()

    // El GET posterior dice empresa_completa: true, así que el panel NO puede
    // caer en el mensaje que culpa al instructor.
    expect(w.text()).toContain('duración en horas')
    expect(w.text()).not.toContain('Falta que el instructor')
    // Y se repreguntó en vez de deducirlo: dos GET, el inicial y el de después.
    expect(get).toHaveBeenCalledTimes(2)
  })

  it('ofrece la descarga cuando ya hay documento', async () => {
    get.mockResolvedValue({
      data: { ...sinDatos, trabajador_completo: true, constancia_url: 'https://r2/c.docx' },
    })

    const w = mount(DC3Panel, {
      props: { cursoId: 'abc', completado: true, variante: 'plano' },
    })
    await flushPromises()

    const a = w.find('a[download]')
    expect(a.exists()).toBe(true)
    expect(a.attributes('href')).toBe('https://r2/c.docx')
  })
})
