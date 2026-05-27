import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'
import router from '../router'
import { toast } from '../utils/toast'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<{ id: string; name: string; email: string; role: string; avatar_url?: string } | null>(
    JSON.parse(localStorage.getItem('user') || 'null')
  )

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isInstructor = computed(() => user.value?.role === 'instructor')

  async function login(email: string, password: string) {
    const res = await api.post('/login', { email, password })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
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
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    toast.success('¡Hasta pronto! Sesión cerrada correctamente.', 'Hasta pronto')
    setTimeout(() => router.push('/login'), 800)
  }

  function handleSessionExpired() {
    const hadToken = !!token.value
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    if (hadToken) {
      toast.warning('Tu sesión ha expirado. Por favor inicia sesión nuevamente.', 'Sesión expirada')
      setTimeout(() => router.push('/login'), 2000)
    }
  }

  return { token, user, isLoggedIn, isAdmin, isInstructor, login, logout, handleSessionExpired }
})
