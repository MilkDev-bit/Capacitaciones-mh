import os

path = r"C:\Proyectos\capacitaciones'mh\frontend\src\views\LoginView.vue"
login_content = """\
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
    error.value = e.response?.data?.error || 'Correo o contrase\\u00f1a incorrectos'
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
    regSuccess.value = '\\u00a1Cuenta creada! Ya puedes iniciar sesi\\u00f3n.'
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
        <div class="hero-stats">
          <div class="stat"><span class="stat-num">500+</span><span class="stat-lbl">Cursos</span></div>
          <div class="stat-divider"></div>
          <div class="stat"><span class="stat-num">12K+</span><span class="stat-lbl">Estudiantes</span></div>
          <div class="stat-divider"></div>
          <div class="stat"><span class="stat-num">98%</span><span class="stat-lbl">Satisfacci\\u00f3n</span></div>
        </div>
        <ul class="hero-features">
          <li><span class="feat-check">\\u2713</span> Cursos de video, documentos y texto</li>
          <li><span class="feat-check">\\u2713</span> Ex\\u00e1menes con retroalimentaci\\u00f3n</li>
          <li><span class="feat-check">\\u2713</span> Acceso por c\\u00f3digo o enlace de invitaci\\u00f3n</li>
        </ul>
      </div>
      <div class="hero-decoration"></div>
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
          <button :class="[\\'form-tab\\', tab === \\'login\\' ? \\'active\\' : \\'\\']" @click="tab = \\'login\\'">
            Iniciar sesi\\u00f3n
          </button>
          <button :class="[\\'form-tab\\', tab === \\'register\\' ? \\'active\\' : \\'\\']" @click="tab = \\'register\\'">
            Registrarse
          </button>
        </div>

        <form v-if="tab === \\'login\\'" @submit.prevent="submit" class="auth-form">
          <div class="form-group">
            <label>Correo electr\\u00f3nico</label>
            <input class="field-input" v-model="email" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>Contrase\\u00f1a</label>
            <div class="pass-wrap">
              <input class="field-input" v-model="password" :type="showPass ? \\'text\\' : \\'password\\'" placeholder="\\u2022\\u2022\\u2022\\u2022\\u2022\\u2022\\u2022\\u2022" autocomplete="current-password" required />
              <button type="button" class="pass-toggle" @click="showPass = !showPass">
                {{ showPass ? \\'\\U0001f648\\' : \\'\\U0001f441\\' }}
              </button>
            </div>
          </div>
          <div v-if="error" class="alert alert-error">{{ error }}</div>
          <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="loading">
            <span v-if="loading" class="btn-spinner"></span>
            {{ loading ? \\'Entrando...\\' : \\'Entrar a la plataforma\\' }}
          </button>
          <p class="form-footer">
            \\u00bfNo tienes cuenta? <button type="button" class="link-btn" @click="tab = \\'register\\'">Reg\\u00edstrate gratis</button>
          </p>
        </form>

        <form v-if="tab === \\'register\\'" @submit.prevent="register" class="auth-form">
          <div class="form-group">
            <label>Nombre completo</label>
            <input class="field-input" v-model="regName" type="text" placeholder="Tu nombre completo" autocomplete="name" required />
          </div>
          <div class="form-group">
            <label>Correo electr\\u00f3nico</label>
            <input class="field-input" v-model="regEmail" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>Contrase\\u00f1a</label>
            <input class="field-input" v-model="regPassword" type="password" placeholder="M\\u00ednimo 6 caracteres" autocomplete="new-password" required minlength="6" />
          </div>
          <div class="form-group">
            <label>Tipo de cuenta</label>
            <div class="role-grid">
              <button type="button" :class="[\\'role-card\\', regRole === \\'user\\' ? \\'selected\\' : \\'\\']" @click="regRole = \\'user\\'">
                <span class="role-icon">\\U0001f393</span>
                <strong>Estudiante</strong>
                <small>Accede y aprende a tu ritmo</small>
              </button>
              <button type="button" :class="[\\'role-card\\', regRole === \\'instructor\\' ? \\'selected\\' : \\'\\']" @click="regRole = \\'instructor\\'">
                <span class="role-icon">\\U0001f3eb</span>
                <strong>Instructor</strong>
                <small>Crea y gestiona cursos</small>
              </button>
            </div>
          </div>
          <div v-if="regError" class="alert alert-error">{{ regError }}</div>
          <div v-if="regSuccess" class="alert alert-success">{{ regSuccess }}</div>
          <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="regLoading">
            <span v-if="regLoading" class="btn-spinner"></span>
            {{ regLoading ? \\'Creando cuenta...\\' : \\'Crear cuenta gratis\\' }}
          </button>
          <p class="form-footer">
            \\u00bfYa tienes cuenta? <button type="button" class="link-btn" @click="tab = \\'login\\'">Inicia sesi\\u00f3n</button>
          </p>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page { display: flex; min-height: 100vh; }
.auth-hero {
  flex: 0 0 45%; background: linear-gradient(145deg, #1c1d1f 0%, #2d2f31 100%);
  display: flex; flex-direction: column; justify-content: center; padding: 60px 56px;
  position: relative; overflow: hidden;
}
.hero-decoration {
  position: absolute; right: -80px; bottom: -80px; width: 320px; height: 320px; border-radius: 50%;
  background: radial-gradient(circle, rgba(249,115,22,.25) 0%, transparent 70%); pointer-events: none;
}
.hero-content { position: relative; z-index: 1; }
.hero-logo { margin-bottom: 20px; }
.hero-title { font-size: 2.4rem; font-weight: 900; color: #fff; line-height: 1.15; margin-bottom: 16px; }
.hero-title span { color: var(--brand); }
.hero-stats { display: flex; align-items: center; gap: 20px; margin-bottom: 36px; }
.stat { text-align: center; }
.stat-num { display: block; font-size: 1.4rem; font-weight: 800; color: var(--brand); }
.stat-lbl { font-size: 0.78rem; color: rgba(255,255,255,.5); text-transform: uppercase; letter-spacing: .05em; }
.stat-divider { width: 1px; height: 36px; background: rgba(255,255,255,.15); }
.hero-features { list-style: none; display: flex; flex-direction: column; gap: 10px; }
.hero-features li { color: rgba(255,255,255,.75); font-size: 0.92rem; display: flex; align-items: center; gap: 10px; }
.feat-check { background: rgba(249,115,22,.2); color: var(--brand); border-radius: 50%; width: 22px; height: 22px; display: flex; align-items: center; justify-content: center; font-size: 0.78rem; font-weight: 900; flex-shrink: 0; }
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
.pass-wrap .field-input { padding-right: 44px; }
.pass-toggle { position: absolute; right: 12px; top: 50%; transform: translateY(-50%); background: none; border: none; font-size: 1rem; cursor: pointer; line-height: 1; }
.role-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.role-card { display: flex; flex-direction: column; align-items: center; gap: 6px; text-align: center; padding: 16px 12px; border: 2px solid var(--border); border-radius: var(--r); background: var(--surface); cursor: pointer; transition: all 0.18s; }
.role-card:hover { border-color: var(--brand); }
.role-card.selected { border-color: var(--brand); background: var(--brand-light); }
.role-icon { font-size: 1.6rem; }
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
"""

with open(login_path, 'w', encoding='utf-8') as f:
    f.write(login_content.encode('raw_unicode_escape').decode('unicode_escape'))
print("OK:", login_path)
content = """\
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
    error.value = e.response?.data?.error || 'Correo o contrase\u00f1a incorrectos'
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
    regSuccess.value = '\u00a1Cuenta creada! Ya puedes iniciar sesi\u00f3n.'
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
    <!-- Panel izquierdo -->
    <div class="auth-hero">
      <div class="hero-content">
        <div class="hero-logo">
          <svg width="44" height="44" viewBox="0 0 44 44" fill="none">
            <rect width="44" height="44" rx="12" fill="rgba(255,255,255,0.15)"/>
            <path d="M12 32L22 14L32 32H12Z" fill="white"/>
          </svg>
        </div>
        <h1 class="hero-title">Capacitaciones<br><span>MH</span></h1>
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
            <span class="stat-lbl">Satisfacci\u00f3n</span>
          </div>
        </div>
        <ul class="hero-features">
          <li><span class="feat-check">\u2713</span> Cursos de video, documentos y texto</li>
          <li><span class="feat-check">\u2713</span> Ex\u00e1menes con retroalimentaci\u00f3n</li>
          <li><span class="feat-check">\u2713</span> Acceso por c\u00f3digo o enlace de invitaci\u00f3n</li>
        </ul>
      </div>
      <div class="hero-decoration"></div>
    </div>

    <!-- Panel derecho -->
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
            Iniciar sesi\u00f3n
          </button>
          <button :class="['form-tab', tab === 'register' ? 'active' : '']" @click="tab = 'register'">
            Registrarse
          </button>
        </div>

        <!-- LOGIN -->
        <form v-if="tab === 'login'" @submit.prevent="submit" class="auth-form">
          <div class="form-group">
            <label>Correo electr\u00f3nico</label>
            <input class="field-input" v-model="email" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>Contrase\u00f1a</label>
            <div class="pass-wrap">
              <input class="field-input" v-model="password" :type="showPass ? 'text' : 'password'" placeholder="\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022" autocomplete="current-password" required />
              <button type="button" class="pass-toggle" @click="showPass = !showPass">
                {{ showPass ? '\U0001f648' : '\U0001f441' }}
              </button>
            </div>
          </div>
          <div v-if="error" class="alert alert-error">{{ error }}</div>
          <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="loading">
            <span v-if="loading" class="btn-spinner"></span>
            {{ loading ? 'Entrando...' : 'Entrar a la plataforma' }}
          </button>
          <p class="form-footer">
            \u00bfNo tienes cuenta? <button type="button" class="link-btn" @click="tab = 'register'">Reg\u00edstrate gratis</button>
          </p>
        </form>

        <!-- REGISTER -->
        <form v-if="tab === 'register'" @submit.prevent="register" class="auth-form">
          <div class="form-group">
            <label>Nombre completo</label>
            <input class="field-input" v-model="regName" type="text" placeholder="Tu nombre completo" autocomplete="name" required />
          </div>
          <div class="form-group">
            <label>Correo electr\u00f3nico</label>
            <input class="field-input" v-model="regEmail" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
          </div>
          <div class="form-group">
            <label>Contrase\u00f1a</label>
            <input class="field-input" v-model="regPassword" type="password" placeholder="M\u00ednimo 6 caracteres" autocomplete="new-password" required minlength="6" />
          </div>
          <div class="form-group">
            <label>Tipo de cuenta</label>
            <div class="role-grid">
              <button
                type="button"
                :class="['role-card', regRole === 'user' ? 'selected' : '']"
                @click="regRole = 'user'"
              >
                <span class="role-icon">\U0001f393</span>
                <strong>Estudiante</strong>
                <small>Accede y aprende a tu ritmo</small>
              </button>
              <button
                type="button"
                :class="['role-card', regRole === 'instructor' ? 'selected' : '']"
                @click="regRole = 'instructor'"
              >
                <span class="role-icon">\U0001f3eb</span>
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
            \u00bfYa tienes cuenta? <button type="button" class="link-btn" @click="tab = 'login'">Inicia sesi\u00f3n</button>
          </p>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page { display: flex; min-height: 100vh; }

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

.auth-form-panel {
  flex: 1; display: flex; align-items: center; justify-content: center;
  padding: 40px 24px; background: var(--bg);
}
.auth-form-wrap { width: 100%; max-width: 420px; }
.mobile-logo { display: none; align-items: center; gap: 10px; font-size: 1rem; font-weight: 800; color: var(--dark); margin-bottom: 28px; }

.form-tabs { display: flex; background: var(--border-light); border-radius: var(--r); padding: 4px; gap: 4px; margin-bottom: 28px; }
.form-tab {
  flex: 1; padding: 9px; border: none; border-radius: var(--r-sm); background: transparent;
  font-size: 0.9rem; font-weight: 600; color: var(--muted); transition: all 0.18s; cursor: pointer;
}
.form-tab.active { background: var(--surface); color: var(--dark); box-shadow: var(--shadow-sm); }

.auth-form { display: flex; flex-direction: column; gap: 18px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.pass-wrap { position: relative; }
.pass-wrap .field-input { padding-right: 44px; }
.pass-toggle {
  position: absolute; right: 12px; top: 50%; transform: translateY(-50%);
  background: none; border: none; font-size: 1rem; cursor: pointer; line-height: 1;
}

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

.submit-btn { width: 100%; margin-top: 4px; }
.btn-spinner {
  width: 14px; height: 14px; border: 2px solid rgba(255,255,255,.4);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; flex-shrink: 0;
}

.form-footer { text-align: center; font-size: 0.85rem; color: var(--muted); }
.link-btn { background: none; border: none; color: var(--brand); font-weight: 700; cursor: pointer; padding: 0; font-size: inherit; }
.link-btn:hover { text-decoration: underline; }

@media (max-width: 860px) {
  .auth-hero { display: none; }
  .auth-form-panel { padding: 32px 20px; }
  .mobile-logo { display: flex; }
}
@media (max-width: 420px) {
  .role-grid { grid-template-columns: 1fr; }
}
</style>
"""

with open(path, 'w', encoding='utf-8') as f:
    f.write(content)
print("OK:", path)

content = """\
<script setup lang="ts">

import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const cursosPublicos = ref<any[]>([])
const router = useRouter()
const activeTab = ref<'mis' | 'explorar'>('mis')
const inscribiendose = ref<string | null>(null)

const codigoInput = ref('')
const codigoLoading = ref(false)
const codigoError = ref('')
const codigoSuccess = ref('')

async function loadMis() {
  const res = await api.get('/mis-capacitaciones')
  capacitaciones.value = res.data || []
}
async function loadPublicos() {
  const res = await api.get('/cursos-publicos')
  cursosPublicos.value = res.data || []
}
onMounted(() => { loadMis(); loadPublicos() })

async function inscribirse(id: string) {
  inscribiendose.value = id
  try {
    await api.post(`/inscribirse/${id}`)
    await Promise.all([loadMis(), loadPublicos()])
    activeTab.value = 'mis'
  } finally { inscribiendose.value = null }
}

async function unirseConCodigo() {
  const code = codigoInput.value.trim().toUpperCase()
  if (!code) { codigoError.value = 'Ingresa un c\u00f3digo'; return }
  codigoError.value = ''; codigoSuccess.value = ''
  codigoLoading.value = true
  try {
    const res = await api.post('/unirse-con-codigo', { codigo: code })
    codigoSuccess.value = `\u00a1Te uniste a "${res.data.title}"!`
    codigoInput.value = ''
    await loadMis()
    setTimeout(() => { codigoSuccess.value = ''; activeTab.value = 'mis' }, 2000)
  } catch (e: any) {
    codigoError.value = e.response?.data?.error || 'C\u00f3digo inv\u00e1lido'
  } finally { codigoLoading.value = false }
}

const thumbClass: Record<string, string> = { video: 'thumb-video', document: 'thumb-document', text: 'thumb-text' }
const typeIcon: Record<string, string> = { video: '\U0001f3a5', document: '\U0001f4c4', text: '\U0001f4dd' }
const typeLabel: Record<string, string> = { video: 'Video', document: 'Documento', text: 'Texto' }
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="ph">
      <div>
        <h1 class="ph-title">Mis aprendizajes</h1>
        <p class="ph-sub">Todos tus cursos y capacitaciones asignadas</p>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs-bar">
      <button :class="['tab-pill', activeTab === 'mis' ? 'active' : '']" @click="activeTab = 'mis'">
        Mis cursos
        <span class="pill-count">{{ capacitaciones.length }}</span>
      </button>
      <button :class="['tab-pill', activeTab === 'explorar' ? 'active' : '']" @click="activeTab = 'explorar'">
        Explorar
        <span class="pill-count">{{ cursosPublicos.length }}</span>
      </button>
    </div>

    <!-- Mis cursos -->
    <div v-if="activeTab === 'mis'">
      <div v-if="capacitaciones.length" class="courses-grid">
        <div
          v-for="c in capacitaciones" :key="c.id"
          class="course-card"
          @click="router.push('/usuario/capacitaciones/' + c.id)"
          tabindex="0" @keyup.enter="router.push('/usuario/capacitaciones/' + c.id)"
        >
          <div :class="['course-thumb', thumbClass[c.type] || 'thumb-default']">
            <span class="thumb-icon">{{ typeIcon[c.type] || '\U0001f4da' }}</span>
          </div>
          <div class="course-body">
            <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripci\u00f3n' }}</p>
            <div class="course-cta">Continuar aprendiendo \u2192</div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <div class="empty-icon">\U0001f4da</div>
        <h3>No tienes cursos asignados a\u00fan</h3>
        <p>Explora los cursos disponibles o pide a tu instructor un c\u00f3digo de acceso.</p>
        <button class="btn btn-primary" @click="activeTab = 'explorar'">Explorar cursos</button>
      </div>
    </div>

    <!-- Explorar -->
    <div v-if="activeTab === 'explorar'">
      <!-- Join by code -->
      <div class="code-banner">
        <div class="code-banner-left">
          <div class="code-key-icon">\U0001f511</div>
          <div>
            <strong>\u00bfTienes un c\u00f3digo de acceso?</strong>
            <p>Ingresa el c\u00f3digo de tu instructor para unirte a un curso privado.</p>
          </div>
        </div>
        <div class="code-banner-right">
          <input
            v-model="codigoInput"
            class="code-field"
            placeholder="ABC123"
            maxlength="12"
            @keyup.enter="unirseConCodigo"
          />
          <button class="btn btn-primary" :disabled="codigoLoading" @click="unirseConCodigo">
            {{ codigoLoading ? 'Cargando...' : 'Unirme' }}
          </button>
        </div>
      </div>
      <div v-if="codigoError" class="alert alert-error" style="margin-bottom:16px">{{ codigoError }}</div>
      <div v-if="codigoSuccess" class="alert alert-success" style="margin-bottom:16px">{{ codigoSuccess }}</div>

      <p class="section-label">Cursos disponibles para todos</p>
      <div v-if="cursosPublicos.length" class="courses-grid">
        <div v-for="c in cursosPublicos" :key="c.id" class="course-card public-course">
          <div :class="['course-thumb', thumbClass[c.type] || 'thumb-default']">
            <span class="thumb-icon">{{ typeIcon[c.type] || '\U0001f4da' }}</span>
            <span v-if="c.inscrito" class="enrolled-ribbon">\u2713 Inscrito</span>
          </div>
          <div class="course-body">
            <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripci\u00f3n' }}</p>
            <div class="course-footer-row">
              <span v-if="c.inscrito" class="badge badge-green">\u2713 Ya inscrito</span>
              <button
                v-else
                class="btn btn-primary btn-sm"
                :disabled="inscribiendose === c.id"
                @click.stop="inscribirse(c.id)"
              >
                {{ inscribiendose === c.id ? 'Inscribiendo...' : '+ Inscribirse gratis' }}
              </button>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <div class="empty-icon">\U0001f50d</div>
        <h3>No hay cursos p\u00fablicos disponibles</h3>
        <p>Pide a tu instructor que comparta el enlace o c\u00f3digo de su curso.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ph { margin-bottom: 24px; }
.ph-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.ph-sub { color: var(--muted); font-size: 0.9rem; margin-top: 4px; }

/* Tabs */
.tabs-bar { display: flex; gap: 8px; margin-bottom: 24px; }
.tab-pill {
  padding: 9px 20px; border: 2px solid var(--border); border-radius: 24px; background: var(--surface);
  font-size: 0.88rem; font-weight: 600; color: var(--muted); cursor: pointer;
  display: flex; align-items: center; gap: 8px; transition: all 0.18s;
}
.tab-pill:hover { border-color: var(--brand); color: var(--brand); }
.tab-pill.active { border-color: var(--brand); background: var(--brand); color: #fff; }
.tab-pill.active .pill-count { background: rgba(255,255,255,.25); color: #fff; }
.pill-count { background: var(--border-light); color: var(--muted); font-size: 0.75rem; padding: 1px 7px; border-radius: 12px; font-weight: 700; }

/* Course grid */
.courses-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(268px, 1fr)); gap: 20px; }
.course-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  box-shadow: var(--shadow-sm); cursor: pointer; transition: transform 0.2s, box-shadow 0.2s;
  display: flex; flex-direction: column;
}
.course-card:hover { transform: translateY(-4px); box-shadow: var(--shadow-md); }

/* Thumbnail banner */
.course-thumb {
  height: 140px; display: flex; align-items: center; justify-content: center;
  position: relative; flex-shrink: 0;
}
.thumb-icon { font-size: 2.8rem; filter: drop-shadow(0 2px 6px rgba(0,0,0,.25)); }
.enrolled-ribbon {
  position: absolute; top: 10px; right: 10px;
  background: rgba(0,0,0,.55); color: #fff; font-size: 0.72rem; font-weight: 700;
  padding: 3px 10px; border-radius: 20px; backdrop-filter: blur(4px);
}

/* Course body */
.course-body { padding: 16px; display: flex; flex-direction: column; gap: 6px; flex: 1; }
.course-type-badge {
  font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: .06em;
  color: var(--brand-dark); background: var(--brand-light); padding: 2px 8px; border-radius: 4px;
  display: inline-block; width: fit-content;
}
.course-title { font-size: 0.97rem; font-weight: 700; color: var(--dark); line-height: 1.35; }
.course-desc { font-size: 0.83rem; color: var(--muted); line-height: 1.45; flex: 1; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.course-cta { font-size: 0.83rem; font-weight: 700; color: var(--brand); margin-top: 4px; }
.course-footer-row { display: flex; align-items: center; gap: 10px; margin-top: 4px; }

/* Code banner */
.code-banner {
  display: flex; align-items: center; gap: 20px; flex-wrap: wrap;
  background: var(--surface); border-radius: var(--r-lg); padding: 20px 24px;
  box-shadow: var(--shadow-sm); border-left: 4px solid var(--brand); margin-bottom: 24px;
}
.code-banner-left { display: flex; align-items: center; gap: 14px; flex: 1; min-width: 200px; }
.code-key-icon { font-size: 1.8rem; }
.code-banner-left strong { font-size: 0.95rem; color: var(--dark); display: block; }
.code-banner-left p { font-size: 0.82rem; color: var(--muted); margin-top: 2px; }
.code-banner-right { display: flex; gap: 10px; align-items: center; }
.code-field {
  padding: 9px 14px; border: 2px solid var(--border); border-radius: var(--r);
  font-size: 1.1rem; font-weight: 800; letter-spacing: .15em; text-transform: uppercase;
  width: 130px; outline: none; font-family: 'Courier New', monospace; background: var(--bg);
}
.code-field:focus { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(249,115,22,.12); }

.section-label { font-size: 1rem; font-weight: 700; color: var(--dark); margin-bottom: 16px; }

/* Empty */
.empty-state { text-align: center; padding: 60px 20px; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.empty-icon { font-size: 3rem; }
.empty-state h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
.empty-state p { color: var(--muted); font-size: 0.9rem; max-width: 360px; }

@media (max-width: 560px) {
  .courses-grid { grid-template-columns: 1fr; }
  .code-banner { flex-direction: column; align-items: flex-start; }
  .code-banner-right { width: 100%; }
  .code-field { flex: 1; }
}
</style>
"""

with open(path, 'w', encoding='utf-8') as f:
    f.write(content)
print("OK:", path)
