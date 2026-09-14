import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import SearchUserModal from './SearchUserModal.vue'
import CreateGroupModal from './CreateGroupModal.vue'

vi.mock('../api', () => ({ default: { get: vi.fn().mockResolvedValue({ data: [] }), post: vi.fn() } }))
vi.mock('../utils/toast', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

/**
 * Regresión: "ReferenceError: props is not defined".
 *
 * Los dos modales llamaban a `toRef(props, 'show')` pero declaraban las props
 * con `defineProps<...>()` sin asignar el resultado. En una compilación de
 * desarrollo eso pasa desapercibido hasta que alguien abre el modal; en
 * producción reventaba el setup entero y la pantalla se quedaba en blanco.
 *
 * El type-check lo señalaba desde hacía tiempo y se ignoró por venir "de
 * antes". Esta prueba lo convierte en un fallo de CI, que sí se mira.
 *
 * Montar basta: el error ocurría en `setup`, antes de pintar nada.
 */
describe('modales de la pantalla de mensajes', () => {
  beforeEach(() => vi.clearAllMocks())

  it('SearchUserModal se monta abierto sin reventar', () => {
    const w = mount(SearchUserModal, { props: { show: true } })
    expect(w.text()).toContain('Nueva Conversación')
  })

  it('CreateGroupModal se monta abierto sin reventar', () => {
    const w = mount(CreateGroupModal, { props: { show: true } })
    expect(w.html()).not.toBe('')
  })

  // Cerrados tampoco deben fallar: `useScrollLock` recibe la prop igualmente.
  it('se montan cerrados sin reventar', () => {
    expect(() => mount(SearchUserModal, { props: { show: false } })).not.toThrow()
    expect(() => mount(CreateGroupModal, { props: { show: false } })).not.toThrow()
  })
})
