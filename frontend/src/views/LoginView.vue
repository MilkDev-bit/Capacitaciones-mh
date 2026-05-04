Ã¯Â»Â¿ÃÂ¯ÃÂ»ÃÂ¿<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../api'
import router from '../router'
import { useRoute } from 'vue-router'

const auth = useAuthStore()
const route = useRoute()
const tab = ref<'login' | 'register'>('login')

const email = ref('')
const password = ref('')
const showPass = ref(false)
const error = ref('')
const loading = ref(false)

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
    const res = await api.post('/login', { email: email.value, password: password.value })
    auth.token = res.data.token
    auth.user = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
    const redirect = route.query.redirect as string | undefined
    if (redirect) router.push(redirect)
    else if (res.data.user?.role === 'admin') router.push('/admin')
    else if (res.data.user?.role === 'instructor') router.push('/instructor')
    else router.push('/usuario')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Correo o contraseÃÂÃÂÃÂÃÂ±a incorrectos'
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
    regSuccess.value = 'ÃÂÃ¢ÂÂÃÂÃÂ¡Cuenta creada! Ya puedes iniciar sesiÃÂÃÂÃÂÃÂ³n.'
    regName.value = ''; regEmail.value = ''; regPassword.value = ''; regRole.value = 'user'
    setTimeout(() => { tab.value = 'login'; regSuccess.value = '' }, 1500)
  } catch (e: any) {
    regError.value = e.response?.data?.error || 'Error al registrarse'
  } finally {
    regLoading.value = false
  }
}

function initials(name: string) {
  return name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2)
}
</script>

<template>
  <div class="auth-page">
    <!-- Panel izquierdo ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂ¢Ã¢ÂÂ¬ÃÂ hero -->
    <div class="auth-hero">
      <div class="hero-content">
        <div class="hero-logo">
          <svg width="44" height="44" viewBox="0 0 44 44" fill="none">
            <rect width="44" height="44" rx="12" fill="rgba(255,255,255,0.15)"/>
            <path d="M12 32L22 14L32 32H12Z" fill="white"/>
          </svg>
        </div>
        <h1 class="hero-title">Capacitaciones<br><span>MH</span></h1>
        <p class="hero-subtitle">Aprende, crece y certifica tus habilidades con la plataforma de capacitaciÃÂÃÂÃÂÃÂ³n corporativa.</p>
        <div class="hero-stats">
          <div class="stat">
            <span class="stat-num">500+</span>
            <span class="stat-lbl">Cursos</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat">
            <span class="stat-num">12K+</span>
            <span class="stat-lbl">Estudiantes</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat">
            <span class="stat-num">98%</span>
            <span class="stat-lbl">SatisfacciÃÂÃÂÃÂÃÂ³n</span>
          </div>
        </div>
        <!-- Feature list -->
        <ul class="hero-features">
          <li><span class="feat-check">ÃÂÃÂ¢ÃÂÃ¢ÂÂÃÂ¢Ã¢ÂÂ¬ÃÂ</span> Cursos de video, documentos y texto</li>
          <li><span class="feat-check">ÃÂÃÂ¢ÃÂÃ¢ÂÂÃÂ¢Ã¢ÂÂ¬ÃÂ</span> ExÃÂÃÂÃÂÃÂ¡menes con retroalimentaciÃÂÃÂÃÂÃÂ³n</li>
          <li><span class="feat-check">ÃÂÃÂ¢ÃÂÃ¢ÂÂÃÂ¢Ã¢ÂÂ¬ÃÂ</span> Acceso por cÃÂÃÂÃÂÃÂ³digo o enlace de invitaciÃÂÃÂÃÂÃÂ³n</li>
        </ul>
      </div>
      <div class="hero-decoration"></div>
    </div>

    <!-- Panel derecho ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂ¢Ã¢ÂÂ¬ÃÂ formulario -->
    <div class="auth-form-panel">
      <div class="auth-form-wrap">
        <!-- Logo mobile -->
        <div class="mobile-logo">
          <svg width="32" height="32" viewBox="0 0 44 44" fill="none">
            <rect width="44" height="44" rx="12" fill="var(--brand)"/>
            <path d="M12 32L22 14L32 32H12Z" fill="white"/>
          </svg>
          <span>Capacitaciones MH</span>
        </div>

        <!-- Tabs -->
        <div class="form-tabs">
          <button :class="['form-tab', tab === 'login' ? 'active' : '']" @click="tab = 'login'">
            Iniciar sesiÃÂÃÂÃÂÃÂ³n
          </button>
          <button :class="['form-tab', tab === 'register' ? 'active' : '']" @click="tab = 'register'">
            Registrarse
          </button>
        </div>

        <!-- LOGIN -->
        <form v-if="tab === 'login'" @submit.prevent="submit" class="auth-form">
          <div class="form-group">
            <label>Correo electrÃÂÃÂÃÂÃÂ³nico</label>
            <input class="field-input" v-model="email" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>ContraseÃÂÃÂÃÂÃÂ±a</label>
            <div class="pass-wrap">
              <input class="field-input" v-model="password" :type="showPass ? 'text' : 'password'" placeholder="ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢" autocomplete="current-password" required />
              <button type="button" class="pass-toggle" @click="showPass = !showPass">
                {{ showPass ? 'ÃÂÃÂ°ÃÂÃÂ¸ÃÂ¢Ã¢ÂÂÃÂ¢ÃÂÃ¢ÂÂ ' : 'ÃÂÃÂ°ÃÂÃÂ¸ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂÃÂ' }}
              </button>
            </div>
          </div>
          <div v-if="error" class="alert alert-error">{{ error }}</div>
          <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="loading">
            <span v-if="loading" class="btn-spinner"></span>
            {{ loading ? 'Entrando...' : 'Entrar a la plataforma' }}
          </button>
          <p class="form-footer">
            ÃÂÃ¢ÂÂÃÂÃÂ¿No tienes cuenta? <button type="button" class="link-btn" @click="tab = 'register'">RegÃÂÃÂÃÂÃÂ­strate gratis</button>
          </p>
        </form>

        <!-- REGISTER -->
        <form v-if="tab === 'register'" @submit.prevent="register" class="auth-form">
          <div class="form-group">
            <label>Nombre completo</label>
            <input class="field-input" v-model="regName" type="text" placeholder="Tu nombre completo" autocomplete="name" required />
          </div>
          <div class="form-group">
            <label>Correo electrÃÂÃÂÃÂÃÂ³nico</label>
            <input class="field-input" v-model="regEmail" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>ContraseÃÂÃÂÃÂÃÂ±a</label>
            <input class="field-input" v-model="regPassword" type="password" placeholder="MÃÂÃÂÃÂÃÂ­nimo 6 caracteres" autocomplete="new-password" required minlength="6" />
          </div>
          <div class="form-group">
            <label>Tipo de cuenta</label>
            <div class="role-grid">
              <button
                type="button"
                :class="['role-card', regRole === 'user' ? 'selected' : '']"
                @click="regRole = 'user'"
              >
                <span class="role-icon">ÃÂÃÂ°ÃÂÃÂ¸ÃÂÃÂ½ÃÂ¢Ã¢ÂÂ¬ÃÂ</span>
                <strong>Estudiante</strong>
                <small>Accede y aprende a tu ritmo</small>
              </button>
              <button
                type="button"
                :class="['role-card', regRole === 'instructor' ? 'selected' : '']"
                @click="regRole = 'instructor'"
              >
                <span class="role-icon">ÃÂÃÂ°ÃÂÃÂ¸ÃÂÃÂÃÂÃÂ«</span>
                <strong>Instructor</strong>
                <small>Crea y gestiona cursos</small>
              </button>
            </div>
          </div>
          <div v-if="regError" class="alert alert-error">{{ regError }}</div>
          <div v-if="regSuccess" class="alert alert-success">{{ regSuccess }}</div>
          <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="regLoading">
            <span v-if="regLoading" class="btn-spinner"></span>
            {{ regLoading ? 'Creando cuenta...' : 'Crear cuenta gratis' }}
          </button>
          <p class="form-footer">
            ÃÂÃ¢ÂÂÃÂÃÂ¿Ya tienes cuenta? <button type="button" class="link-btn" @click="tab = 'login'">Inicia sesiÃÂÃÂÃÂÃÂ³n</button>
          </p>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ Layout ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ */
.auth-page { display: flex; min-height: 100vh; }

/* ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ Hero (left) ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ */
.auth-hero {
  flex: 0 0 45%; background: linear-gradient(145deg, #1c1d1f 0%, #2d2f31 100%);
  display: flex; flex-direction: column; justify-content: center; padding: 60px 56px;
  position: relative; overflow: hidden;
}
.hero-decoration {
  position: absolute; right: -80px; bottom: -80px;
  width: 320px; height: 320px; border-radius: 50%;
  background: radial-gradient(circle, rgba(249,115,22,.25) 0%, transparent 70%);
  pointer-events: none;
}
.hero-content { position: relative; z-index: 1; }
.hero-logo { margin-bottom: 20px; }
.hero-title {
  font-size: 2.4rem; font-weight: 900; color: #fff; line-height: 1.15; margin-bottom: 16px;
}
.hero-title span { color: var(--brand); }
.hero-subtitle { color: rgba(255,255,255,.65); font-size: 1rem; line-height: 1.6; margin-bottom: 36px; max-width: 340px; }
.hero-stats { display: flex; align-items: center; gap: 20px; margin-bottom: 36px; }
.stat { text-align: center; }
.stat-num { display: block; font-size: 1.4rem; font-weight: 800; color: var(--brand); }
.stat-lbl { font-size: 0.78rem; color: rgba(255,255,255,.5); text-transform: uppercase; letter-spacing: .05em; }
.stat-divider { width: 1px; height: 36px; background: rgba(255,255,255,.15); }
.hero-features { list-style: none; display: flex; flex-direction: column; gap: 10px; }
.hero-features li { color: rgba(255,255,255,.75); font-size: 0.92rem; display: flex; align-items: center; gap: 10px; }
.feat-check { background: rgba(249,115,22,.2); color: var(--brand); border-radius: 50%; width: 22px; height: 22px; display: flex; align-items: center; justify-content: center; font-size: 0.78rem; font-weight: 900; flex-shrink: 0; }

/* ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ Form panel (right) ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ */
.auth-form-panel {
  flex: 1; display: flex; align-items: center; justify-content: center;
  padding: 40px 24px; background: var(--bg);
}
.auth-form-wrap { width: 100%; max-width: 420px; }

.mobile-logo { display: none; align-items: center; gap: 10px; font-size: 1rem; font-weight: 800; color: var(--dark); margin-bottom: 28px; }

/* Tabs */
.form-tabs { display: flex; background: var(--border-light); border-radius: var(--r); padding: 4px; gap: 4px; margin-bottom: 28px; }
.form-tab {
  flex: 1; padding: 9px; border: none; border-radius: var(--r-sm); background: transparent;
  font-size: 0.9rem; font-weight: 600; color: var(--muted); transition: all 0.18s; cursor: pointer;
}
.form-tab.active { background: var(--surface); color: var(--dark); box-shadow: var(--shadow-sm); }

/* Form */
.auth-form { display: flex; flex-direction: column; gap: 18px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.pass-wrap { position: relative; }
.pass-wrap .field-input { padding-right: 44px; }
.pass-toggle {
  position: absolute; right: 12px; top: 50%; transform: translateY(-50%);
  background: none; border: none; font-size: 1rem; cursor: pointer; line-height: 1;
}

/* Role selector */
.role-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.role-card {
  display: flex; flex-direction: column; align-items: center; gap: 6px; text-align: center;
  padding: 16px 12px; border: 2px solid var(--border); border-radius: var(--r);
  background: var(--surface); cursor: pointer; transition: all 0.18s;
}
.role-card:hover { border-color: var(--brand); }
.role-card.selected { border-color: var(--brand); background: var(--brand-light); }
.role-icon { font-size: 1.6rem; }
.role-card strong { font-size: 0.88rem; color: var(--dark); }
.role-card small { font-size: 0.75rem; color: var(--muted); line-height: 1.3; }

/* Submit */
.submit-btn { width: 100%; margin-top: 4px; }
.btn-spinner {
  width: 14px; height: 14px; border: 2px solid rgba(255,255,255,.4);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; flex-shrink: 0;
}

/* Footer */
.form-footer { text-align: center; font-size: 0.85rem; color: var(--muted); }
.link-btn { background: none; border: none; color: var(--brand); font-weight: 700; cursor: pointer; padding: 0; font-size: inherit; }
.link-btn:hover { text-decoration: underline; }

/* ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ Responsive ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ÃÂÃÂ¢ÃÂ¢Ã¢ÂÂ¬ÃÂÃÂ¢Ã¢ÂÂÃÂ¬ */
@media (max-width: 860px) {
  .auth-hero { display: none; }
  .auth-form-panel { padding: 32px 20px; }
  .mobile-logo { display: flex; }
}
@media (max-width: 420px) {
  .role-grid { grid-template-columns: 1fr; }
}
</style>
