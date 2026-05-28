<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../api'
import router from '../router'
import { useRoute } from 'vue-router'
import { toast } from '../utils/toast'
import { getRecaptchaToken } from '../utils/recaptcha'

const auth = useAuthStore()
const route = useRoute()
const tab = ref<'login' | 'register' | 'forgot'>('login')

const email = ref('')
const password = ref('')
const showPass = ref(false)
const loading = ref(false)

const regName = ref('')
const regEmail = ref('')
const regPassword = ref('')
const regConfirmPassword = ref('')
const showRegPass = ref(false)
const regLoading = ref(false)


const forgotEmail = ref('')
const forgotLoading = ref(false)
const forgotStep = ref<1 | 2>(1)
const resetCode = ref('')
const newPassword = ref('')
const newPasswordConfirm = ref('')
const showNewPass = ref(false)

const isValidEmail = (e: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e)

const passStrength = computed(() => {
  const p = regPassword.value
  if (!p) return 0
  let score = 0
  if (p.length >= 8) score += 25
  if (p.length >= 12) score += 25
  if (/[A-Z]/.test(p)) score += 25
  if (/[0-9]/.test(p) && /[^A-Za-z0-9]/.test(p)) score += 25
  return Math.min(100, score)
})

const strengthColor = computed(() => {
  if (passStrength.value <= 25) return 'var(--danger)'
  if (passStrength.value <= 50) return 'var(--warning)'
  if (passStrength.value <= 75) return 'var(--brand)'
  return 'var(--success)'
})

const strengthText = computed(() => {
  if (!regPassword.value) return 'Ingresa una contraseña'
  if (passStrength.value <= 25) return 'Débil'
  if (passStrength.value <= 50) return 'Regular'
  if (passStrength.value <= 75) return 'Buena'
  return 'Fuerte'
})

const passwordsMatch = computed(() => {
  return regPassword.value && regConfirmPassword.value && regPassword.value === regConfirmPassword.value
})

async function submit() {
  if (!isValidEmail(email.value)) {
    toast.error('Formato de correo inválido')
    return
  }
  loading.value = true
  try {
    const recaptcha_token = await getRecaptchaToken('login')
    const res = await api.post('/login', { email: email.value, password: password.value, recaptcha_token })
    // El token JWT viene en cookie HttpOnly — solo guardamos el perfil
    auth.user = res.data.user
    localStorage.setItem('user', JSON.stringify(res.data.user))
    toast.success(`¡Bienvenido/a, ${res.data.user.name}!`, 'Sesión iniciada')
    const redirect = route.query.redirect as string | undefined
    if (redirect) router.push(redirect)
    else if (res.data.user?.role === 'admin') router.push('/admin')
    else if (res.data.user?.role === 'instructor') router.push('/instructor')
    else router.push('/usuario')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Correo o contraseña incorrectos')
  } finally {
    loading.value = false
  }
}

async function register() {
  if (!regName.value || !regEmail.value || !regPassword.value || !regConfirmPassword.value) {
    toast.error('Todos los campos son requeridos'); return
  }
  if (!isValidEmail(regEmail.value)) {
    toast.error('Formato de correo inválido'); return
  }
  if (!passwordsMatch.value) {
    toast.error('Las contraseñas no coinciden'); return
  }
  if (passStrength.value < 50) {
    toast.error('La contraseña es muy débil. Usa mayúsculas y números.'); return
  }
  
  regLoading.value = true
  try {
    const recaptcha_token = await getRecaptchaToken('register')
    await api.post('/register', { name: regName.value, email: regEmail.value, password: regPassword.value, recaptcha_token })
    toast.success('¡Cuenta creada! Redirigiendo...')
    setTimeout(() => { 
      tab.value = 'login'
      email.value = regEmail.value
      password.value = regPassword.value
      regName.value = ''; regEmail.value = ''; regPassword.value = ''; regConfirmPassword.value = '';
    }, 1500)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al registrarse')
  } finally {
    regLoading.value = false
  }
}

async function forgotPassword() {
  if (!isValidEmail(forgotEmail.value)) {
    toast.error('Ingresa un correo válido')
    return
  }
  forgotLoading.value = true
  try {
    await api.post('/forgot-password', { email: forgotEmail.value })
    toast.success('Si el correo existe, recibirás el código de recuperación.')
    forgotStep.value = 2
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No se pudo enviar el correo de recuperación')
  } finally {
    forgotLoading.value = false
  }
}

async function resetPassword() {
  if (!resetCode.value.trim()) {
    toast.error('Ingresa el código recibido en tu correo')
    return
  }
  if (newPassword.value.length < 8) {
    toast.error('La contraseña debe tener al menos 8 caracteres')
    return
  }
  if (newPassword.value !== newPasswordConfirm.value) {
    toast.error('Las contraseñas no coinciden')
    return
  }
  forgotLoading.value = true
  try {
    await api.post('/reset-password', {
      email: forgotEmail.value,
      code: resetCode.value.trim(),
      new_password: newPassword.value,
    })
    toast.success('¡Contraseña actualizada! Ya puedes iniciar sesión.')
    setTimeout(() => {
      tab.value = 'login'
      forgotEmail.value = ''
      resetCode.value = ''
      newPassword.value = ''
      newPasswordConfirm.value = ''
      forgotStep.value = 1
    }, 1500)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Código inválido o expirado')
  } finally {
    forgotLoading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-hero">
      <div class="bubbles" aria-hidden="true">
        <span></span><span></span><span></span><span></span><span></span>
        <span></span><span></span><span></span><span></span><span></span>
      </div>
      <div class="hero-content">
        <div class="hero-logo">
          <svg width="44" height="44" viewBox="0 0 44 44" fill="none">
            <rect width="44" height="44" rx="12" fill="rgba(255,255,255,0.15)"/>
            <path d="M12 32L22 14L32 32H12Z" fill="white"/>
          </svg>
        </div>
        <h1 class="hero-title">Capacitaciones<br><span>MH</span></h1>
        <p class="hero-subtitle">Cursos, lecciones, examenes y seguimiento en una sola experiencia de aprendizaje.</p>
        <div class="hero-features">
          <div class="hero-feature">
            <span class="hero-feature-icon">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
            </span>
            <span>Cursos con video, documentos y lecturas</span>
          </div>
          <div class="hero-feature">
            <span class="hero-feature-icon">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg>
            </span>
            <span>Exámenes y seguimiento de progreso</span>
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

        <div class="form-tabs" v-if="tab !== 'forgot'">
          <button :class="['form-tab', tab === 'login' ? 'active' : '']" @click="tab = 'login'">
            Iniciar sesión
          </button>
          <button :class="['form-tab', tab === 'register' ? 'active' : '']" @click="tab = 'register'">
            Registrarse
          </button>
        </div>

        <Transition name="fade-slide" mode="out-in">
          <form v-if="tab === 'login'" @submit.prevent="submit" class="auth-form">
            <div class="form-group">
              <label>Correo electrónico</label>
              <input class="field-input" v-model="email" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
            </div>
            <div class="form-group">
              <div style="display:flex; justify-content: space-between; align-items: baseline;">
                <label>Contraseña</label>
                <button type="button" class="link-btn-sm" @click="tab = 'forgot'; forgotStep = 1">¿Olvidaste tu contraseña?</button>
              </div>
              <div class="pass-wrap">
                <input class="field-input" v-model="password" :type="showPass ? 'text' : 'password'" placeholder="••••••••" autocomplete="current-password" required />
                <button type="button" class="pass-toggle" @click="showPass = !showPass">
                  {{ showPass ? 'Ocultar' : 'Ver' }}
                </button>
              </div>
            </div>
            <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="loading">
              <span v-if="loading" class="btn-spinner"></span>
              {{ loading ? 'Entrando...' : 'Entrar a la plataforma' }}
            </button>
            <p class="recaptcha-notice">
              Protegido por reCAPTCHA &middot;
              <a href="https://policies.google.com/privacy" target="_blank" rel="noopener">Privacidad</a> &middot;
              <a href="https://policies.google.com/terms" target="_blank" rel="noopener">Términos</a>
            </p>
            <p class="form-footer">
              ¿No tienes cuenta? <button type="button" class="link-btn" @click="tab = 'register'">Regístrate gratis</button>
            </p>
          </form>

          <form v-else-if="tab === 'register'" @submit.prevent="register" class="auth-form">
            <div class="form-group">
              <label>Nombre completo</label>
              <input class="field-input" v-model="regName" type="text" placeholder="Tu nombre completo" autocomplete="name" required />
            </div>
            <div class="form-group">
              <label>Correo electrónico</label>
              <input class="field-input" v-model="regEmail" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
            </div>
            
            <div class="pass-row">
              <div class="form-group" style="flex: 1;">
                <label>Contraseña</label>
                <div class="pass-wrap">
                  <input class="field-input" v-model="regPassword" :type="showRegPass ? 'text' : 'password'" placeholder="Mínimo 8 caracteres" autocomplete="new-password" required minlength="8" />
                  <button type="button" class="pass-toggle" @click="showRegPass = !showRegPass">
                    {{ showRegPass ? 'Ocultar' : 'Ver' }}
                  </button>
                </div>
          
                <div class="pass-strength" v-if="regPassword">
                  <div class="strength-bar">
                    <div class="strength-fill" :style="{ width: `${passStrength}%`, backgroundColor: strengthColor }"></div>
                  </div>
                  <span class="strength-text" :style="{ color: strengthColor }">{{ strengthText }}</span>
                </div>
              </div>

              <div class="form-group" style="flex: 1;">
                <label>Confirmar contraseña</label>
                <div class="pass-wrap">
                  <input class="field-input" v-model="regConfirmPassword" :type="showRegPass ? 'text' : 'password'" placeholder="Repite tu contraseña" autocomplete="new-password" required minlength="8" />
                </div>
                <div v-if="regConfirmPassword && !passwordsMatch" class="form-hint error-hint">
                  Las contraseñas no coinciden
                </div>
              </div>
            </div>

            <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="regLoading || !passwordsMatch || passStrength < 50">
              <span v-if="regLoading" class="btn-spinner"></span>
              {{ regLoading ? 'Creando cuenta...' : 'Crear cuenta gratis' }}
            </button>
            <p class="recaptcha-notice">
              Protegido por reCAPTCHA &middot;
              <a href="https://policies.google.com/privacy" target="_blank" rel="noopener">Privacidad</a> &middot;
              <a href="https://policies.google.com/terms" target="_blank" rel="noopener">Términos</a>
            </p>
            <p class="form-footer">
              ¿Ya tienes cuenta? <button type="button" class="link-btn" @click="tab = 'login'">Inicia sesión</button>
            </p>
          </form>

          <form v-else-if="tab === 'forgot' && forgotStep === 1" @submit.prevent="forgotPassword" class="auth-form forgot-form">
            <div class="forgot-header">
              <div class="forgot-icon"><svg width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg></div>
              <h2>Recuperar contraseña</h2>
              <p>Ingresa tu correo y te enviaremos un código para restablecer tu contraseña.</p>
            </div>
            
            <div class="form-group">
              <label>Correo electrónico</label>
              <input class="field-input" v-model="forgotEmail" type="email" placeholder="correo@empresa.com" autocomplete="email" required />
            </div>

            <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="forgotLoading">
              <span v-if="forgotLoading" class="btn-spinner"></span>
              {{ forgotLoading ? 'Enviando...' : 'Enviar código' }}
            </button>
            <button type="button" class="btn btn-secondary btn-lg submit-btn" style="margin-top:0" @click="tab = 'login'" :disabled="forgotLoading">
              Volver al inicio de sesión
            </button>
          </form>

          <form v-else-if="tab === 'forgot' && forgotStep === 2" @submit.prevent="resetPassword" class="auth-form forgot-form">
            <div class="forgot-header">
              <div class="forgot-icon"><svg width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07A19.5 19.5 0 013.07 9.8a19.79 19.79 0 01-3-8.57A2 2 0 012.02 3h3a2 2 0 012 1.72c.127.96.361 1.903.7 2.81a2 2 0 01-.45 2.11L6.09 10.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0122 16.92z"/></svg></div>
              <h2>Ingresa el código</h2>
              <p>Revisa tu correo <strong>{{ forgotEmail }}</strong> e ingresa el código recibido.</p>
            </div>

            <div class="form-group">
              <label>Código de verificación</label>
              <input class="field-input" v-model="resetCode" type="text" placeholder="XXX-XXX" autocomplete="one-time-code" required />
            </div>

            <div class="form-group">
              <label>Nueva contraseña</label>
              <div class="pass-wrap">
                <input class="field-input" v-model="newPassword" :type="showNewPass ? 'text' : 'password'" placeholder="Mínimo 8 caracteres" autocomplete="new-password" required minlength="8" />
                <button type="button" class="pass-toggle" @click="showNewPass = !showNewPass">
                  {{ showNewPass ? 'Ocultar' : 'Ver' }}
                </button>
              </div>
            </div>

            <div class="form-group">
              <label>Confirmar contraseña</label>
              <input class="field-input" v-model="newPasswordConfirm" :type="showNewPass ? 'text' : 'password'" placeholder="Repite tu nueva contraseña" autocomplete="new-password" required minlength="8" />
              <div v-if="newPasswordConfirm && newPassword !== newPasswordConfirm" class="form-hint error-hint">
                Las contraseñas no coinciden
              </div>
            </div>

            <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="forgotLoading || (!!newPasswordConfirm && newPassword !== newPasswordConfirm)">
              <span v-if="forgotLoading" class="btn-spinner"></span>
              {{ forgotLoading ? 'Actualizando...' : 'Restablecer contraseña' }}
            </button>
            <button type="button" class="btn btn-secondary btn-lg submit-btn" style="margin-top:0" @click="forgotStep = 1" :disabled="forgotLoading">
              Volver atrás
            </button>
          </form>
        </Transition>
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
.hero-features { display: flex; flex-direction: column; gap: 12px; margin-top: 28px; }
.hero-feature {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 14px; border-radius: 10px;
  background: rgba(255,255,255,.07); border: 1px solid rgba(255,255,255,.1);
  color: rgba(255,255,255,.82); font-size: 0.88rem; font-weight: 500;
}
.hero-feature-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: 8px; flex-shrink: 0;
  background: rgba(249,115,22,.2); border: 1px solid rgba(249,115,22,.3);
  color: var(--brand); backdrop-filter: blur(4px);
}

/* Bubbles animation */
.bubbles {
  position: absolute; inset: 0; overflow: hidden; pointer-events: none; z-index: 0;
}
.bubbles span {
  position: absolute; bottom: -120px;
  border-radius: 50%;
  background: rgba(249,115,22,.12);
  border: 1px solid rgba(249,115,22,.18);
  animation: bubble-rise linear infinite;
  backdrop-filter: blur(2px);
}

/* Each bubble: different size, position, duration, delay */
.bubbles span:nth-child(1)  { width: 48px;  height: 48px;  left: 8%;   animation-duration: 9s;  animation-delay: 0s;   }
.bubbles span:nth-child(2)  { width: 28px;  height: 28px;  left: 18%;  animation-duration: 7s;  animation-delay: 1.5s; }
.bubbles span:nth-child(3)  { width: 70px;  height: 70px;  left: 32%;  animation-duration: 11s; animation-delay: 0.8s; background: rgba(255,255,255,.05); }
.bubbles span:nth-child(4)  { width: 22px;  height: 22px;  left: 45%;  animation-duration: 6s;  animation-delay: 3s;   }
.bubbles span:nth-child(5)  { width: 54px;  height: 54px;  left: 58%;  animation-duration: 10s; animation-delay: 1s;   background: rgba(249,115,22,.08); }
.bubbles span:nth-child(6)  { width: 18px;  height: 18px;  left: 68%;  animation-duration: 7.5s;animation-delay: 2.2s; }
.bubbles span:nth-child(7)  { width: 36px;  height: 36px;  left: 76%;  animation-duration: 8s;  animation-delay: 0.3s; }
.bubbles span:nth-child(8)  { width: 60px;  height: 60px;  left: 85%;  animation-duration: 12s; animation-delay: 4s;   background: rgba(255,255,255,.04); }
.bubbles span:nth-child(9)  { width: 26px;  height: 26px;  left: 24%;  animation-duration: 9.5s;animation-delay: 5s;   }
.bubbles span:nth-child(10) { width: 42px;  height: 42px;  left: 52%;  animation-duration: 8.5s;animation-delay: 2.8s; background: rgba(249,115,22,.1); }

@keyframes bubble-rise {
  0%   { transform: translateY(0)   scale(1)   rotate(0deg);   opacity: 0; }
  10%  { opacity: 1; }
  80%  { opacity: .7; }
  100% { transform: translateY(-110vh) scale(1.15) rotate(30deg); opacity: 0; }
}

.auth-form-panel { flex: 1; display: flex; align-items: center; justify-content: center; padding: 40px 24px; background: var(--bg); overflow-y: auto; }
.auth-form-wrap { width: 100%; max-width: 440px; }
.mobile-logo { display: none; align-items: center; gap: 10px; font-size: 1rem; font-weight: 800; color: var(--dark); margin-bottom: 28px; }
.form-tabs { display: flex; background: var(--border-light); border-radius: var(--r); padding: 4px; gap: 4px; margin-bottom: 28px; }
.form-tab { flex: 1; padding: 9px; border: none; border-radius: var(--r-sm); background: transparent; font-size: 0.9rem; font-weight: 600; color: var(--muted); transition: all 0.18s; cursor: pointer; }
.form-tab.active { background: var(--surface); color: var(--dark); box-shadow: var(--shadow-sm); }
.auth-form { display: flex; flex-direction: column; gap: 18px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.pass-wrap { position: relative; }
.pass-wrap .field-input { padding-right: 78px; }
.pass-toggle { position: absolute; right: 8px; top: 50%; transform: translateY(-50%); background: var(--border-light); border: none; border-radius: 6px; color: var(--dark); font-size: 0.76rem; font-weight: 800; cursor: pointer; line-height: 1; padding: 7px 9px; transition: all 0.2s; }
.pass-toggle:hover { background: var(--border); }
.pass-row { display: flex; gap: 14px; flex-wrap: wrap; }
.pass-strength { display: flex; align-items: center; justify-content: space-between; margin-top: 6px; gap: 10px; }
.strength-bar { flex: 1; height: 5px; background: var(--border-light); border-radius: 4px; overflow: hidden; }
.strength-fill { height: 100%; transition: all 0.3s ease; }
.strength-text { font-size: 0.72rem; font-weight: 700; }
.form-hint { font-size: 0.75rem; margin-top: 4px; }
.error-hint { color: var(--danger); font-weight: 600; }
.role-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.role-card { display: flex; flex-direction: column; align-items: center; gap: 6px; text-align: center; padding: 14px 10px; border: 2px solid var(--border); border-radius: var(--r); background: var(--surface); cursor: pointer; transition: all 0.18s; }
.role-card:hover { border-color: var(--brand); }
.role-card.selected { border-color: var(--brand); background: var(--brand-light); }
.role-icon {
  width: 38px; height: 30px; display: inline-flex; align-items: center; justify-content: center;
  border-radius: 7px; background: var(--border-light); color: var(--brand-dark); font-size: 0.72rem; font-weight: 900;
}
.role-card.selected .role-icon { background: var(--brand); color: #fff; }
.role-card strong { font-size: 0.88rem; color: var(--dark); }
.role-card small { font-size: 0.72rem; color: var(--muted); line-height: 1.2; }
.submit-btn { width: 100%; margin-top: 4px; display: flex; justify-content: center; align-items: center; gap: 8px; }
.submit-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.recaptcha-notice {
  text-align: center; font-size: 0.72rem; color: var(--muted);
  display: flex; align-items: center; justify-content: center; gap: 4px; flex-wrap: wrap;
}
.recaptcha-notice a { color: var(--muted); text-decoration: underline; }
.recaptcha-notice a:hover { color: var(--brand); }
.btn-spinner { width: 16px; height: 16px; border: 2.5px solid rgba(255,255,255,.4); border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; flex-shrink: 0; }
.form-footer { text-align: center; font-size: 0.85rem; color: var(--muted); }
.link-btn { background: none; border: none; color: var(--brand); font-weight: 700; cursor: pointer; padding: 0; font-size: inherit; }
.link-btn-sm { background: none; border: none; color: var(--brand); font-weight: 600; cursor: pointer; padding: 0; font-size: 0.78rem; transition: color 0.2s; }
.link-btn:hover, .link-btn-sm:hover { text-decoration: underline; color: var(--brand-dark); }

/* Forgot Password Styles */
.forgot-header { text-align: center; margin-bottom: 24px; }
.forgot-icon { font-size: 3rem; margin-bottom: 12px; }
.forgot-header h2 { font-size: 1.5rem; font-weight: 800; color: var(--dark); margin-bottom: 8px; }
.forgot-header p { font-size: 0.9rem; color: var(--muted); line-height: 1.5; }

/* Transitions */
.fade-slide-enter-active, .fade-slide-leave-active { transition: all 0.25s ease; }
.fade-slide-enter-from { opacity: 0; transform: translateY(10px); }
.fade-slide-leave-to { opacity: 0; transform: translateY(-10px); }

@media (max-width: 860px) {
  .auth-hero { display: none; }
  .auth-form-panel { padding: 32px 20px; }
  .mobile-logo { display: flex; }
}
@media (max-width: 420px) { 
  .pass-row { flex-direction: column; gap: 14px; }
}
</style>
