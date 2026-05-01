<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../api'
import router from '../router'
import { useRoute } from 'vue-router'

const auth = useAuthStore()
const route = useRoute()
const tab = ref<'login' | 'register'>('login')

// Login
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

// Register
const regName = ref('')
const regEmail = ref('')
const regPassword = ref('')
const regRole = ref('user')
const regError = ref('')
const regSuccess = ref('')
const regLoading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    // login directo sin redirigir automático si hay redirect param
    const res = await api.post('/login', { email: email.value, password: password.value })
    auth.token.value = res.data.token
    auth.user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))

    const redirect = route.query.redirect as string | undefined
    if (redirect) {
      router.push(redirect)
    } else if (res.data.user?.role === 'admin') {
      router.push('/admin')
    } else if (res.data.user?.role === 'instructor') {
      router.push('/instructor')
    } else {
      router.push('/usuario')
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al iniciar sesión'
  } finally {
    loading.value = false
  }
}

async function register() {
  regError.value = ''; regSuccess.value = ''
  if (!regName.value || !regEmail.value || !regPassword.value) {
    regError.value = 'Todos los campos son requeridos'; return
  }
  regLoading.value = true
  try {
    await api.post('/register', { name: regName.value, email: regEmail.value, password: regPassword.value, role: regRole.value })
    regSuccess.value = '¡Cuenta creada! Ya puedes iniciar sesión.'
    regName.value = ''; regEmail.value = ''; regPassword.value = ''; regRole.value = 'user'
    tab.value = 'login'
  } catch (e: any) {
    regError.value = e.response?.data?.error || 'Error al registrarse'
  } finally {
    regLoading.value = false
  }
}
</script>

<template>
  <div class="login-bg">
    <div class="login-card">
      <div class="login-logo">
        <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="10" fill="#3b82f6"/>
          <path d="M10 28L20 12L30 28H10Z" fill="white"/>
        </svg>
        <h1>Capacitaciones MH</h1>
      </div>

      <!-- Tabs -->
      <div class="tabs">
        <button :class="['tab', tab === 'login' ? 'active' : '']" @click="tab = 'login'">Iniciar sesión</button>
        <button :class="['tab', tab === 'register' ? 'active' : '']" @click="tab = 'register'">Registrarse</button>
      </div>

      <!-- Login -->
      <form v-if="tab === 'login'" @submit.prevent="submit">
        <label>Correo electrónico</label>
        <input v-model="email" type="email" placeholder="correo@empresa.com" required />
        <label>Contraseña</label>
        <input v-model="password" type="password" placeholder="••••••••" required />
        <p v-if="error" class="error">{{ error }}</p>
        <button type="submit" :disabled="loading">
          {{ loading ? 'Entrando...' : 'Entrar' }}
        </button>
      </form>

      <!-- Register -->
      <form v-if="tab === 'register'" @submit.prevent="register">
        <label>Nombre completo</label>
        <input v-model="regName" type="text" placeholder="Tu nombre" required />
        <label>Correo electrónico</label>
        <input v-model="regEmail" type="email" placeholder="correo@empresa.com" required />
        <label>Contraseña</label>
        <input v-model="regPassword" type="password" placeholder="Mínimo 6 caracteres" required minlength="6" />
        <label>Tipo de cuenta</label>
        <div class="role-selector">
          <label :class="['role-card', regRole === 'user' ? 'selected' : '']" @click="regRole = 'user'">
            <input type="radio" name="role" value="user" v-model="regRole" hidden />
            <span class="role-icon">🎓</span>
            <strong>Estudiante</strong>
            <small>Accede a cursos asignados e inscríbete en cursos públicos</small>
          </label>
          <label :class="['role-card', regRole === 'instructor' ? 'selected' : '']" @click="regRole = 'instructor'">
            <input type="radio" name="role" value="instructor" v-model="regRole" hidden />
            <span class="role-icon">🧑‍🏫</span>
            <strong>Instructor</strong>
            <small>Crea y comparte tus propios cursos y exámenes</small>
          </label>
        </div>
        <p v-if="regError" class="error">{{ regError }}</p>
        <p v-if="regSuccess" class="success-msg">{{ regSuccess }}</p>
        <button type="submit" :disabled="regLoading">
          {{ regLoading ? 'Creando cuenta...' : 'Crear cuenta' }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-bg {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1e3a5f 0%, #3b82f6 100%);
}
.login-card {
  background: white;
  border-radius: 16px;
  padding: 2.5rem 2rem;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 20px 40px rgba(0,0,0,0.18);
}
.login-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 1.5rem;
}
.login-logo h1 { font-size: 1.1rem; font-weight: 700; color: #1e3a5f; }
.tabs { display: flex; background: #f1f5f9; border-radius: 10px; padding: 4px; margin-bottom: 1.5rem; gap: 4px; }
.tab { flex: 1; padding: 8px; border: none; border-radius: 7px; cursor: pointer; font-size: 0.88rem; font-weight: 600; color: #64748b; background: transparent; transition: all 0.15s; }
.tab.active { background: white; color: #1e293b; box-shadow: 0 1px 4px rgba(0,0,0,0.1); }
label { display: block; font-size: 0.8rem; font-weight: 600; color: #64748b; margin-bottom: 4px; margin-top: 1rem; }
input[type="text"],
input[type="email"],
input[type="password"] {
  width: 100%; padding: 10px 12px; border: 1.5px solid #e2e8f0;
  border-radius: 8px; font-size: 0.95rem; outline: none; transition: border 0.2s;
  box-sizing: border-box;
}
input:focus { border-color: #3b82f6; }
.role-selector { display: flex; gap: 10px; margin-top: 8px; }
.role-card {
  flex: 1; border: 2px solid #e2e8f0; border-radius: 10px; padding: 12px;
  cursor: pointer; text-align: center; transition: all 0.15s;
  display: flex; flex-direction: column; align-items: center; gap: 4px;
}
.role-card:hover { border-color: #3b82f6; }
.role-card.selected { border-color: #3b82f6; background: #eff6ff; }
.role-icon { font-size: 1.5rem; }
.role-card strong { font-size: 0.85rem; color: #1e293b; }
.role-card small { font-size: 0.72rem; color: #64748b; line-height: 1.3; }
button[type="submit"] {
  margin-top: 1.5rem; width: 100%; padding: 11px;
  background: #3b82f6; color: white; border: none;
  border-radius: 8px; font-size: 1rem; font-weight: 600; cursor: pointer; transition: background 0.2s;
}
button[type="submit"]:hover:not(:disabled) { background: #2563eb; }
button[type="submit"]:disabled { opacity: 0.6; cursor: not-allowed; }
.error { color: #ef4444; font-size: 0.85rem; margin-top: 8px; }
.success-msg { color: #059669; font-size: 0.85rem; margin-top: 8px; background: #d1fae5; padding: 8px 10px; border-radius: 6px; }
</style>

