import axios from 'axios'
import { useAuthStore } from './stores/auth'

const api = axios.create({
  baseURL: '/api',
  withCredentials: true, // envía la cookie auth_token HttpOnly automáticamente
})

api.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.handleSessionExpired()
    }
    return Promise.reject(err)
  }
)

export default api
