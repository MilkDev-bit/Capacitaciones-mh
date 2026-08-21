import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ModalAviso from './ModalAviso.vue'
import { AVISO_VERSION } from '../utils/aviso'

vi.mock('../api', () => ({
  default: { get: vi.fn(), post: vi.fn() },
}))
vi.mock('../utils/toast', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))
vi.mock('../stores/auth', () => ({
  useAuthStore: () => ({ isLoggedIn: true, logout: vi.fn() }),
}))

import api from '../api'

const get = api.get as unknown as ReturnType<typeof vi.fn>
const post = api.post as unknown as ReturnType<typeof vi.fn>

/** Monta el modal y espera a que resuelva la consulta del perfil. */
async function montar() {
  const w = mount(ModalAviso, {
    global: { stubs: { Teleport: true, RouterLink: true } },
  })
  await flushPromises()
  return w
}

const visible = (w: Awaited<ReturnType<typeof montar>>) => w.find('.ma-modal').exists()

describe('ModalAviso', () => {
  beforeEach(() => vi.clearAllMocks())

  /**
   * Regresión de la lectura del perfil.
   *
   * GET /perfil responde { user, stats }: la versión aceptada viene anidada
   * bajo `user`. Leerla del nivel de arriba daba undefined, y como "sin
   * versión" significa "no ha aceptado", el modal reaparecía en cada carga
   * incluso para quien acababa de aceptar. Bloqueante e imposible de esquivar.
   */
  it('no se muestra a quien ya aceptó la versión vigente', async () => {
    get.mockResolvedValue({ data: { user: { aviso_version: AVISO_VERSION }, stats: {} } })
    expect(visible(await montar())).toBe(false)
  })

  it('se muestra a quien nunca aceptó', async () => {
    get.mockResolvedValue({ data: { user: { aviso_version: '' }, stats: {} } })
    expect(visible(await montar())).toBe(true)
  })

  it('se muestra a quien aceptó una versión anterior', async () => {
    get.mockResolvedValue({ data: { user: { aviso_version: '2020-01-01' }, stats: {} } })
    expect(visible(await montar())).toBe(true)
  })

  /**
   * Un fallo de red no debe dejar a nadie encerrado fuera de su cuenta: se
   * prefiere volver a preguntar en la siguiente carga.
   */
  it('no bloquea si la consulta del perfil falla', async () => {
    get.mockRejectedValue(new Error('sin red'))
    expect(visible(await montar())).toBe(false)
  })

  it('manda la versión vigente al aceptar y se cierra', async () => {
    get.mockResolvedValue({ data: { user: { aviso_version: '' }, stats: {} } })
    post.mockResolvedValue({ data: { aceptado: true } })
    const w = await montar()

    await w.find('.ma-check input').setValue(true)
    await w.find('.btn-primary').trigger('click')
    await flushPromises()

    expect(post).toHaveBeenCalledWith('/perfil/aviso', { version: AVISO_VERSION })
    expect(visible(w)).toBe(false)
  })
})
