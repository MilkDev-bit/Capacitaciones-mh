import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// ── Mocks ──────────────────────────────────────────────────────────────────
// vi.mock factories are hoisted — use vi.hoisted() for shared variables.
const { mockToast } = vi.hoisted(() => ({
  mockToast: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    confirm: vi.fn().mockResolvedValue(true),
  },
}))

vi.mock('../api', () => ({
  default: {
    post: vi.fn(),
  },
}))

vi.mock('../router', () => ({
  default: {
    push: vi.fn(),
  },
}))

vi.mock('../utils/toast', () => ({
  toast: mockToast,
}))

// ── Importar store DESPUÉS de los mocks ────────────────────────────────────
import { useAuthStore } from './auth'
import api from '../api'
import router from '../router'

// ── Helpers ────────────────────────────────────────────────────────────────
const fakeUser = { id: 'u1', name: 'Test User', email: 'test@test.com', role: 'user' }

beforeEach(() => {
  setActivePinia(createPinia())
  localStorage.clear()
  vi.clearAllMocks()
})

// ── Tests ──────────────────────────────────────────────────────────────────
describe('useAuthStore – login', () => {
  it('guarda usuario en el store tras login exitoso', async () => {
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { user: fakeUser },
    })

    const store = useAuthStore()
    await store.login(fakeUser.email, 'password123')

    expect(store.user?.id).toBe(fakeUser.id)
    expect(store.isLoggedIn).toBe(true)
  })

  it('persiste perfil de usuario en localStorage', async () => {
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { user: fakeUser },
    })

    const store = useAuthStore()
    await store.login(fakeUser.email, 'password123')

    expect(JSON.parse(localStorage.getItem('user')!).id).toBe(fakeUser.id)
  })

  it('redirige al admin si rol es admin', async () => {
    const adminUser = { ...fakeUser, role: 'admin' }
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { user: adminUser },
    })

    const store = useAuthStore()
    await store.login(adminUser.email, 'password123')

    expect(vi.mocked(router.push)).toHaveBeenCalledWith('/admin')
  })

  it('redirige al instructor si rol es instructor', async () => {
    const instUser = { ...fakeUser, role: 'instructor' }
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { user: instUser },
    })

    const store = useAuthStore()
    await store.login(instUser.email, 'password123')

    expect(vi.mocked(router.push)).toHaveBeenCalledWith('/instructor')
  })
})

describe('useAuthStore – verificación de correo', () => {
  const errorNoVerificado = {
    response: { status: 403, data: { code: 'email_not_verified', error: 'verifica tu correo' } },
  }

  it('redirige a /verificar-correo cuando el correo no está confirmado', async () => {
    vi.mocked(api.post).mockRejectedValueOnce(errorNoVerificado)

    const store = useAuthStore()
    await store.login('sin-verificar@test.com', 'password123')

    expect(vi.mocked(router.push)).toHaveBeenCalledWith({
      path: '/verificar-correo',
      query: { email: 'sin-verificar@test.com' },
    })
    expect(store.isLoggedIn).toBe(false)
  })

  it('recuerda el correo pendiente para precargar la pantalla de verificación', async () => {
    vi.mocked(api.post).mockRejectedValueOnce(errorNoVerificado)

    const store = useAuthStore()
    await store.login('sin-verificar@test.com', 'password123')

    expect(localStorage.getItem('pending_verification_email')).toBe('sin-verificar@test.com')
  })

  it('propaga cualquier otro error de login sin redirigir', async () => {
    vi.mocked(api.post).mockRejectedValueOnce({
      response: { status: 401, data: { error: 'credenciales inválidas' } },
    })

    const store = useAuthStore()
    await expect(store.login('a@test.com', 'mala')).rejects.toBeDefined()
    expect(vi.mocked(router.push)).not.toHaveBeenCalled()
  })
})

describe('useAuthStore – logout', () => {
  it('limpia usuario al hacer logout', async () => {
    vi.mocked(api.post)
      .mockResolvedValueOnce({ data: { user: fakeUser } }) // login
      .mockResolvedValueOnce({})                           // logout
    const store = useAuthStore()
    await store.login(fakeUser.email, 'password123')

    await store.logout()

    expect(store.user).toBeNull()
    expect(store.isLoggedIn).toBe(false)
    expect(localStorage.getItem('user')).toBeNull()
  })
})

describe('useAuthStore – handleSessionExpired', () => {
  it('limpia el estado cuando la sesión expira', async () => {
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { user: fakeUser },
    })
    const store = useAuthStore()
    await store.login(fakeUser.email, 'password123')

    store.handleSessionExpired()

    expect(store.user).toBeNull()
    expect(store.isLoggedIn).toBe(false)
    expect(localStorage.getItem('user')).toBeNull()
  })

  it('no muestra toast si no había sesión activa', () => {
    const store = useAuthStore()
    store.handleSessionExpired() // sin token previo
    // toast.warning no debe llamarse
    expect(mockToast.warning).not.toHaveBeenCalled()
  })
})

describe('useAuthStore – computed', () => {
  it('isAdmin es true solo cuando el rol es admin', async () => {
    const adminUser = { ...fakeUser, role: 'admin' }
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { user: adminUser },
    })
    const store = useAuthStore()
    await store.login(adminUser.email, 'pass')

    expect(store.isAdmin).toBe(true)
    expect(store.isInstructor).toBe(false)
  })

  it('isInstructor es true solo cuando el rol es instructor', async () => {
    const instUser = { ...fakeUser, role: 'instructor' }
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { user: instUser },
    })
    const store = useAuthStore()
    await store.login(instUser.email, 'pass')

    expect(store.isInstructor).toBe(true)
    expect(store.isAdmin).toBe(false)
  })
})
