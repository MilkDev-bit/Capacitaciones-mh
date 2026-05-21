import axios from 'axios'
import router from './router'
import { toast } from './utils/toast'

const api = axios.create({
  baseURL: '/api',
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let sessionExpiredShown = false

api.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err.response?.status === 401) {
      const hadToken = !!localStorage.getItem('token')
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      if (hadToken && !sessionExpiredShown) {
        sessionExpiredShown = true
        toast.warning('Tu sesión ha expirado. Por favor inicia sesión nuevamente.', 'Sesión expirada')
        setTimeout(() => {
          sessionExpiredShown = false
          router.push('/login')
        }, 2000)
      }
    }
    return Promise.reject(err)
  }
)

export default api
