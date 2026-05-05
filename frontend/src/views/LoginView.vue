<script setup lang="ts">
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
    error.value = e.response?.data?.error || 'Correo o contraseña incorrectos'
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
    setTimeout(() => { tab.value = 'login'; regSuccess.value = '' }, 1500)
  } catch (e: any) {
    regError.value = e.response?.data?.error || 'Error al registrarse'
  } finally {
    regLoading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-hero">
      <div class="hero-content">
        <div class="hero-logo">
          <svg width="44" height="44" viewBox="0 0 44 44" fill="none">
            <rect width="44" height="44" rx="12" fill="rgba(255,255,255,0.15)"/>
            <path d="M12 32L22 14L32 32H12Z" fill="white"/>
          </svg>
        </div>
        <h1 class="hero-title">Capacitaciones<br><span>MH</span></h1>
        <p class="hero-subtitle">Cursos, lecciones, examenes y seguimiento en una sola experiencia de aprendizaje.</p>
        <div class="hero-metrics" aria-label="Resumen de la plataforma">
          <div>
            <strong>24/7</strong>
            <span>Acceso</span>
          </div>
          <div>
            <strong>3</strong>
            <span>Roles</span>
          </div>
          <div>
            <strong>100%</strong>
            <span>Progreso</span>
          </div>
        </div>
        <div class="hero-preview" aria-hidden="true">
          <div class="preview-top">
            <span>Aula activa</span>
            <strong>68%</strong>
          </div>
          <div class="preview-progress"><span></span></div>
          <div class="preview-course">
            <i>VID</i>
            <div>
              <strong>Seguridad operativa</strong>
              <span>12 lecciones · examen final</span>
            </div>
          </div>
          <div class="preview-course">
            <i>PDF</i>
            <div>
              <strong>Protocolos internos</strong>
              <span>Material de consulta</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="auth-form-panel">
      <div class="auth-form-wrap">
        <div class="mobile-logo">
          <svg width="32" height="32" viewBox="0 0 44 44" fill="none">
            <rect width="44" height="44" rx="12" fill="var(--brand)"/>
            <path d="M12 32L22 14L32 32H12Z" fill="white"/>
          </svg>
          <span>Capacitaciones MH</span>
        </div>

        <div class="form-tabs">
          <button :class="['form-tab', tab === 'login' ? 'active' : '']" @click="tab = 'login'">
            Iniciar sesión
          </button>
          <button :class="['form-tab', tab === 'register' ? 'active' : '']" @click="tab = 'register'">
            Registrarse
          </button>
        </div>

        <form v-if="tab === 'login'" @submit.prevent="submit" class="auth-form">
          <div class="form-group">
            <label>Correo electrónico</label>
            <input class="field-input" v-model="email" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>Contraseña</label>
            <div class="pass-wrap">
              <input class="field-input" v-model="password" :type="showPass ? 'text' : 'password'" placeholder="••••••••" autocomplete="current-password" required />
              <button type="button" class="pass-toggle" @click="showPass = !showPass">
                {{ showPass ? 'Ocultar' : 'Ver' }}
              </button>
            </div>
          </div>
          <div v-if="error" class="alert alert-error">{{ error }}</div>
          <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="loading">
            <span v-if="loading" class="btn-spinner"></span>
            {{ loading ? 'Entrando...' : 'Entrar a la plataforma' }}
          </button>
          <p class="form-footer">
            ¿No tienes cuenta? <button type="button" class="link-btn" @click="tab = 'register'">Regístrate gratis</button>
          </p>
        </form>

        <form v-if="tab === 'register'" @submit.prevent="register" class="auth-form">
          <div class="form-group">
            <label>Nombre completo</label>
            <input class="field-input" v-model="regName" type="text" placeholder="Tu nombre completo" autocomplete="name" required />
          </div>
          <div class="form-group">
            <label>Correo electrónico</label>
            <input class="field-input" v-model="regEmail" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>Contraseña</label>
            <input class="field-input" v-model="regPassword" type="password" placeholder="Mínimo 6 caracteres" autocomplete="new-password" required minlength="6" />
          </div>
          <div class="form-group">
            <label>Tipo de cuenta</label>
            <div class="role-grid">
              <button type="button" :class="['role-card', regRole === 'user' ? 'selected' : '']" @click="regRole = 'user'">
                <span class="role-icon">EST</span>
                <strong>Estudiante</strong>
                <small>Accede y aprende a tu ritmo</small>
              </button>
              <button type="button" :class="['role-card', regRole === 'instructor' ? 'selected' : '']" @click="regRole = 'instructor'">
                <span class="role-icon">INS</span>
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
            ¿Ya tienes cuenta? <button type="button" class="link-btn" @click="tab = 'login'">Inicia sesión</button>
          </p>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page { display: flex; min-height: 100vh; }
.auth-hero {
  flex: 0 0 47%;
  background:
    linear-gradient(145deg, rgba(28,29,31,.98) 0%, rgba(38,35,31,.96) 100%),
    linear-gradient(135deg, rgba(249,115,22,.22), rgba(37,99,235,.18));
  display: flex; flex-direction: column; justify-content: center; padding: 60px 56px;
  position: relative; overflow: hidden;
}
.hero-content { position: relative; z-index: 1; max-width: 520px; }
.hero-logo { margin-bottom: 20px; }
.hero-title { font-size: 2.4rem; font-weight: 900; color: #fff; line-height: 1.15; margin-bottom: 16px; }
.hero-title span { color: var(--brand); }
.hero-subtitle { color: rgba(255,255,255,.72); font-size: 0.96rem; line-height: 1.6; max-width: 430px; }
.hero-metrics {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin-top: 26px;
}
.hero-metrics div {
  padding: 12px; border: 1px solid rgba(255,255,255,.14); border-radius: 8px;
  background: rgba(255,255,255,.07);
}
.hero-metrics strong { display: block; color: #fff; font-size: 1.05rem; font-weight: 900; }
.hero-metrics span { color: rgba(255,255,255,.58); font-size: 0.75rem; font-weight: 700; }
.hero-preview {
  margin-top: 24px; padding: 16px; border: 1px solid rgba(255,255,255,.16); border-radius: 8px;
  background: rgba(255,255,255,.1); box-shadow: 0 20px 60px rgba(0,0,0,.22);
}
.preview-top { display: flex; align-items: center; justify-content: space-between; color: rgba(255,255,255,.72); font-size: 0.78rem; font-weight: 800; }
.preview-top strong { color: #fff; }
.preview-progress { height: 7px; margin: 10px 0 14px; border-radius: 999px; background: rgba(255,255,255,.12); overflow: hidden; }
.preview-progress span { display: block; width: 68%; height: 100%; background: linear-gradient(90deg, var(--brand), #22c55e); }
.preview-course { display: flex; gap: 10px; align-items: center; padding: 10px 0; border-top: 1px solid rgba(255,255,255,.1); }
.preview-course i {
  width: 38px; height: 38px; border-radius: 8px; display: flex; align-items: center; justify-content: center;
  background: rgba(249,115,22,.22); color: #fff; font-style: normal; font-size: 0.72rem; font-weight: 900;
}
.preview-course strong { display: block; color: #fff; font-size: 0.86rem; }
.preview-course span { color: rgba(255,255,255,.58); font-size: 0.76rem; }
.auth-form-panel { flex: 1; display: flex; align-items: center; justify-content: center; padding: 40px 24px; background: var(--bg); }
.auth-form-wrap { width: 100%; max-width: 420px; }
.mobile-logo { display: none; align-items: center; gap: 10px; font-size: 1rem; font-weight: 800; color: var(--dark); margin-bottom: 28px; }
.form-tabs { display: flex; background: var(--border-light); border-radius: var(--r); padding: 4px; gap: 4px; margin-bottom: 28px; }
.form-tab { flex: 1; padding: 9px; border: none; border-radius: var(--r-sm); background: transparent; font-size: 0.9rem; font-weight: 600; color: var(--muted); transition: all 0.18s; cursor: pointer; }
.form-tab.active { background: var(--surface); color: var(--dark); box-shadow: var(--shadow-sm); }
.auth-form { display: flex; flex-direction: column; gap: 18px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.pass-wrap { position: relative; }
.pass-wrap .field-input { padding-right: 78px; }
.pass-toggle { position: absolute; right: 8px; top: 50%; transform: translateY(-50%); background: var(--border-light); border: none; border-radius: 6px; color: var(--dark); font-size: 0.76rem; font-weight: 800; cursor: pointer; line-height: 1; padding: 7px 9px; }
.role-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.role-card { display: flex; flex-direction: column; align-items: center; gap: 6px; text-align: center; padding: 16px 12px; border: 2px solid var(--border); border-radius: var(--r); background: var(--surface); cursor: pointer; transition: all 0.18s; }
.role-card:hover { border-color: var(--brand); }
.role-card.selected { border-color: var(--brand); background: var(--brand-light); }
.role-icon {
  width: 38px; height: 30px; display: inline-flex; align-items: center; justify-content: center;
  border-radius: 7px; background: var(--border-light); color: var(--brand-dark); font-size: 0.72rem; font-weight: 900;
}
.role-card.selected .role-icon { background: var(--brand); color: #fff; }
.role-card strong { font-size: 0.88rem; color: var(--dark); }
.role-card small { font-size: 0.75rem; color: var(--muted); line-height: 1.3; }
.submit-btn { width: 100%; margin-top: 4px; }
.btn-spinner { width: 14px; height: 14px; border: 2px solid rgba(255,255,255,.4); border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; flex-shrink: 0; }
.form-footer { text-align: center; font-size: 0.85rem; color: var(--muted); }
.link-btn { background: none; border: none; color: var(--brand); font-weight: 700; cursor: pointer; padding: 0; font-size: inherit; }
.link-btn:hover { text-decoration: underline; }
@media (max-width: 860px) {
  .auth-hero { display: none; }
  .auth-form-panel { padding: 32px 20px; }
  .mobile-logo { display: flex; }
}
@media (max-width: 420px) { .role-grid { grid-template-columns: 1fr; } }
</style>
