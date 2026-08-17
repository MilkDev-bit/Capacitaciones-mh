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
