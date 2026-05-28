import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as Sentry from '@sentry/vue'
import api from '../api'
import router from '../router'
import { toast } from '../utils/toast'

interface UserProfile {
  id: string
  name: string
  email: string
  role: string
  avatar_url?: string
}

interface LoginResponse {
  user: UserProfile
}

export const useAuthStore = defineStore('auth', () => {
  // El token JWT vive en una cookie HttpOnly — no accesible desde JS.
  // Solo almacenamos el perfil del usuario en localStorage (no es sensible).
  const user = ref<UserProfile | null>(
    (() => {
      try {
        const raw = localStorage.getItem('user')
        if (!raw) return null
        return JSON.parse(raw)
      } catch {
        localStorage.removeItem('user')
        return null
      }
    })()
  )

  const isLoggedIn = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isInstructor = computed(() => user.value?.role === 'instructor')

  async function login(email: string, password: string) {
    const res = await api.post<LoginResponse>('/login', { email, password })
    // El servidor ya seteó la cookie HttpOnly; solo guardamos el perfil.
    user.value = res.data.user
    localStorage.setItem('user', JSON.stringify(res.data.user))
    if (user.value?.role === 'admin') {
      router.push('/admin')
    } else if (user.value?.role === 'instructor') {
      router.push('/instructor')
    } else {
      router.push('/usuario')
    }
  }

  async function logout() {
    if (!await toast.confirm('¿Deseas cerrar sesión?', 'Cerrar sesión')) return
    await api.post('/logout').catch((err) => {
      Sentry.captureException(err, { tags: { action: 'logout' } })
    }) // limpia la cookie en el servidor
    user.value = null
    localStorage.removeItem('user')
    toast.success('¡Hasta pronto! Sesión cerrada correctamente.', 'Hasta pronto')
    router.push('/login')
  }

  function handleSessionExpired() {
    const hadUser = !!user.value
    user.value = null
    localStorage.removeItem('user')
    if (hadUser) {
      toast.warning('Tu sesión ha expirado. Por favor inicia sesión nuevamente.', 'Sesión expirada')
      setTimeout(() => router.push('/login'), 2000)
    }
  }

  return { user, isLoggedIn, isAdmin, isInstructor, login, logout, handleSessionExpired }
})
