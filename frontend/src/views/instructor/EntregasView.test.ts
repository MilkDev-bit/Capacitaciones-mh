import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import EntregasView from './EntregasView.vue'

vi.mock('../../api', () => ({
  default: { get: vi.fn(), post: vi.fn() },
}))
vi.mock('../../utils/toast', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

import api from '../../api'
const get = api.get as unknown as ReturnType<typeof vi.fn>

/** Responde según la ruta pedida, como hace el gateway. */
function responder(entregas: any, cursos: any) {
  get.mockImplementation((url: string) =>
    Promise.resolve({ data: url.includes('entregas') ? entregas : cursos })
  )
}

describe('EntregasView', () => {
  beforeEach(() => vi.clearAllMocks())

  /**
   * Regresión de la pantalla en blanco.
   *
   * Sin entregas, el gateway responde `{}`: protobuf-go borra el `repeated`
   * vacío al serializar. El código hacía `res.data?.entregas || res.data || []`
   * y, como un objeto vacío es truthy, la lista acababa siendo `{}`. El
   * computed `stats` llamaba entonces a `.filter()` sobre un objeto, lanzaba
   * TypeError durante el render y la vista entera desaparecía.
   */
  it('no revienta cuando el backend responde un objeto vacío', async () => {
    responder({}, [])
    const w = mount(EntregasView)
    await flushPromises()

    expect(w.text()).toContain('Entregas de Actividades')
    expect(w.text()).toContain('No hay entregas registradas')
  })

  it('tampoco revienta si la clave viene nula', async () => {
    responder({ entregas: null }, null)
    const w = mount(EntregasView)
    await flushPromises()
    expect(w.text()).toContain('No hay entregas registradas')
  })

  it('pinta las entregas cuando las hay', async () => {
    responder(
      {
        entregas: [
          {
            id: '1',
            user_name: 'Ana López',
            user_email: 'ana@ejemplo.mx',
            curso_title: 'Trabajos en alturas',
            leccion_title: 'Actividad 1',
            file_name: 'reporte.pdf',
            file_size: 2048,
            created_at: '2026-08-20T10:00:00Z',
            capacitacion_id: 'c1',
          },
        ],
      },
      [{ id: 'c1', title: 'Trabajos en alturas' }]
    )
    const w = mount(EntregasView)
    await flushPromises()

    expect(w.text()).toContain('Ana López')
    expect(w.text()).toContain('reporte.pdf')
    expect(w.text()).not.toContain('No hay entregas registradas')
  })

  // Un fallo de red tampoco debe dejar la pantalla vacía sin explicación.
  it('sobrevive a un error de la API', async () => {
    get.mockRejectedValue({ response: { data: { error: 'boom' } } })
    const w = mount(EntregasView)
    await flushPromises()
    expect(w.text()).toContain('Entregas de Actividades')
  })
})
