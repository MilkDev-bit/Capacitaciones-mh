<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const email = ref((route.query.email as string) || localStorage.getItem('pending_verification_email') || '')
const digits = ref<string[]>(['', '', '', '', '', ''])
const inputs = ref<HTMLInputElement[]>([])
const loading = ref(false)
const resending = ref(false)
const cooldown = ref(0)
let cooldownTimer: ReturnType<typeof setInterval> | undefined

const code = computed(() => digits.value.join(''))
const complete = computed(() => code.value.length === 6)

onMounted(() => {
  if (!email.value) {
    toast.error('No sabemos qué cuenta verificar. Inicia sesión de nuevo.')
    router.replace('/login')
    return
  }
  startCooldown(60)
  nextTick(() => inputs.value[0]?.focus())
})

onUnmounted(() => clearInterval(cooldownTimer))

function startCooldown(seconds: number) {
  cooldown.value = seconds
  clearInterval(cooldownTimer)
  cooldownTimer = setInterval(() => {
    cooldown.value -= 1
    if (cooldown.value <= 0) clearInterval(cooldownTimer)
  }, 1000)
}

function setInputRef(el: any, i: number) {
  if (el) inputs.value[i] = el as HTMLInputElement
}

function onInput(i: number, event: Event) {
  const target = event.target as HTMLInputElement
  const value = target.value.replace(/\D/g, '')

  if (value.length > 1) {
    // El usuario pegó el código completo dentro de una casilla.
    fill(value)
    return
  }

  digits.value[i] = value
  if (value && i < 5) inputs.value[i + 1]?.focus()
  if (complete.value) verify()
}

function onKeydown(i: number, event: KeyboardEvent) {
  // Retroceso sobre casilla vacía: se salta a la anterior, que es lo que
  // espera cualquiera que se equivocó al teclear.
  if (event.key === 'Backspace' && !digits.value[i] && i > 0) {
    inputs.value[i - 1]?.focus()
    digits.value[i - 1] = ''
    event.preventDefault()
  }
  if (event.key === 'ArrowLeft' && i > 0) inputs.value[i - 1]?.focus()
  if (event.key === 'ArrowRight' && i < 5) inputs.value[i + 1]?.focus()
}

function onPaste(event: ClipboardEvent) {
  event.preventDefault()
  fill(event.clipboardData?.getData('text') ?? '')
}

function fill(raw: string) {
  const clean = raw.replace(/\D/g, '').slice(0, 6)
  digits.value = ['', '', '', '', '', ''].map((_, i) => clean[i] ?? '')
  nextTick(() => {
    inputs.value[Math.min(clean.length, 5)]?.focus()
    if (complete.value) verify()
  })
}

async function verify() {
  if (!complete.value || loading.value) return
  loading.value = true
  try {
    const res = await api.post('/verify-email', { email: email.value, code: code.value })
    localStorage.removeItem('pending_verification_email')
    auth.setUser(res.data.user)
    toast.success('¡Correo verificado! Bienvenido.')
    auth.redirectToHome()
  } catch (e: any) {
    const status = e.response?.status
    const msg = e.response?.data?.error || 'No pudimos verificar el código'
    toast.error(msg)
    digits.value = ['', '', '', '', '', '']
    nextTick(() => inputs.value[0]?.focus())
    // Código caducado o intentos agotados: solo sirve pedir uno nuevo.
    if (status === 410 || status === 429) startCooldown(0)
  } finally {
    loading.value = false
  }
}

async function resend() {
  if (cooldown.value > 0 || resending.value) return
  resending.value = true
  try {
    await api.post('/resend-verification', { email: email.value })
    toast.success('Te enviamos un código nuevo. Revisa tu bandeja.')
    startCooldown(60)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos reenviar el código')
    startCooldown(30)
  } finally {
    resending.value = false
  }
}
</script>

<template>
  <!--
    Mismo esqueleto que LoginView: hero oscuro con burbujas a la izquierda y
    panel de formulario a la derecha. Reutilizar el layout evita que el usuario
    sienta que salió de la plataforma justo en el paso más frágil del alta.
  -->
  <div class="auth-page">
    <div class="auth-hero">
      <div class="bubbles" aria-hidden="true">
        <span></span><span></span><span></span><span></span><span></span>
        <span></span><span></span><span></span><span></span><span></span>
      </div>
      <div class="hero-content">
        <div class="hero-logo">
          <img src="../assets/logo-capacitaciones.png" alt="Capacitaciones MH" class="hero-logo-img" />
        </div>
        <h1 class="hero-title">Un paso<br><span>y listo</span></h1>
        <p class="hero-subtitle">
          Confirmamos tu correo para proteger tu cuenta y poder emitir tus constancias a nombre correcto.
        </p>
        <div class="hero-features">
          <div class="hero-feature">
            <span class="hero-feature-icon">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
            </span>
            <span>El código expira en 15 minutos</span>
          </div>
          <div class="hero-feature">
            <span class="hero-feature-icon">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
            </span>
            <span>¿No llega? Revisa spam o correo no deseado</span>
          </div>
        </div>
      </div>
    </div>

    <div class="auth-form-panel">
      <div class="auth-form-wrap">
        <div class="mobile-logo">
          <img src="../assets/logo-capacitaciones.png" alt="Capacitaciones MH" class="mobile-logo-img" />
          <span>Capacitaciones MH</span>
        </div>

        <div class="verify-head">
          <div class="verify-icon">
            <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <rect width="20" height="16" x="2" y="4" rx="2" />
              <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
            </svg>
          </div>
          <h2>Verifica tu correo</h2>
          <p>
            Enviamos un código de 6 dígitos a<br />
            <strong>{{ email }}</strong>
          </p>
        </div>

        <form class="auth-form" @submit.prevent="verify">
          <div class="code-inputs" @paste="onPaste">
            <input
              v-for="(d, i) in digits"
              :key="i"
              :ref="(el) => setInputRef(el, i)"
              :value="d"
              type="text"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="6"
              class="code-box"
              :class="{ filled: !!d }"
              :disabled="loading"
              :aria-label="`Dígito ${i + 1} de 6`"
              @input="onInput(i, $event)"
              @keydown="onKeydown(i, $event)"
            />
          </div>

          <button type="submit" class="btn btn-primary btn-lg submit-btn" :disabled="!complete || loading">
            <span v-if="loading" class="btn-spinner"></span>
            {{ loading ? 'Verificando…' : 'Verificar y entrar' }}
          </button>

          <div class="resend-row">
            <span v-if="cooldown > 0" class="resend-muted">
              Podrás reenviar el código en <strong>{{ cooldown }}s</strong>
            </span>
            <span v-else>
              ¿No lo recibiste?
              <button type="button" class="link-btn" :disabled="resending" @click="resend">
                {{ resending ? 'Enviando…' : 'Reenviar código' }}
              </button>
            </span>
          </div>

          <p class="form-footnote">
            Revisa la carpeta de spam o correo no deseado. El código expira a los 15 minutos.
          </p>

          <p class="form-footer">
            <router-link to="/login" class="link-btn back-link">← Volver al inicio de sesión</router-link>
          </p>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Layout compartido con LoginView ──────────────────────────────────────── */
.auth-page { display: flex; min-height: 100vh; min-height: 100dvh; }
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
.hero-logo-img { width: 72px; height: 72px; object-fit: contain; filter: drop-shadow(0 4px 20px rgba(249,115,22,.4)); }
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

.bubbles { position: absolute; inset: 0; overflow: hidden; pointer-events: none; z-index: 0; }
.bubbles span {
  position: absolute; bottom: -120px;
  border-radius: 50%;
  background: rgba(249,115,22,.12);
  border: 1px solid rgba(249,115,22,.18);
  animation: bubble-rise linear infinite;
  backdrop-filter: blur(2px);
}
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

/* Respeta la preferencia del sistema: la animación es decorativa. */
@media (prefers-reduced-motion: reduce) {
  .bubbles span { animation: none; opacity: .35; }
}

.auth-form-panel { flex: 1; display: flex; align-items: center; justify-content: center; padding: 40px 24px; background: var(--bg); overflow-y: auto; }
.auth-form-wrap { width: 100%; max-width: 440px; }
.mobile-logo { display: none; align-items: center; gap: 10px; font-size: 1rem; font-weight: 800; color: var(--dark); margin-bottom: 28px; }
.mobile-logo-img { width: 36px; height: 36px; object-fit: contain; }
.auth-form { display: flex; flex-direction: column; gap: 18px; }

/* ── Cabecera de la pantalla ──────────────────────────────────────────────── */
.verify-head { text-align: center; margin-bottom: 26px; }
.verify-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 64px; height: 64px; border-radius: 50%; margin-bottom: 14px;
  background: rgba(249,115,22,.12); color: var(--brand);
}
.verify-head h2 { font-size: 1.5rem; font-weight: 800; color: var(--dark); margin-bottom: 8px; }
.verify-head p { font-size: 0.9rem; color: var(--muted); line-height: 1.6; }
.verify-head strong { color: var(--dark); word-break: break-all; }

/* ── Casillas del código ──────────────────────────────────────────────────── */
.code-inputs { display: flex; gap: 10px; justify-content: center; }
.code-box {
  width: 100%; max-width: 56px; height: 60px;
  border: 2px solid var(--border); border-radius: var(--r-sm);
  background: var(--surface-soft); color: var(--dark);
  text-align: center; font-size: 1.6rem; font-weight: 800;
  transition: border-color .18s var(--ease-apple), background .18s var(--ease-apple), transform .18s var(--ease-apple), box-shadow .18s var(--ease-apple);
}
.code-box:focus {
  outline: none;
  border-color: var(--brand);
  background: var(--surface);
  transform: translateY(-2px);
  box-shadow: 0 6px 18px rgba(249,115,22,.18);
}
.code-box.filled { border-color: var(--brand-border); background: var(--surface); }
.code-box:disabled { opacity: .55; cursor: not-allowed; }

.submit-btn { width: 100%; margin-top: 4px; display: flex; justify-content: center; align-items: center; gap: 8px; }
.submit-btn:disabled { opacity: .6; cursor: not-allowed; }
.btn-spinner { width: 16px; height: 16px; border: 2.5px solid rgba(255,255,255,.4); border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite; flex-shrink: 0; }

.resend-row { text-align: center; font-size: 0.85rem; color: var(--muted); min-height: 22px; }
.resend-muted strong { color: var(--dark); font-variant-numeric: tabular-nums; }
.link-btn { background: none; border: none; color: var(--brand); font-weight: 700; cursor: pointer; padding: 0; font-size: inherit; text-decoration: none; }
.link-btn:hover:not(:disabled) { text-decoration: underline; color: var(--brand-dark); }
.link-btn:disabled { opacity: .6; cursor: not-allowed; }
.back-link { color: var(--muted); font-weight: 600; }
.back-link:hover { color: var(--dark); text-decoration: none; }

.form-footnote { text-align: center; font-size: 0.78rem; color: var(--subtle); line-height: 1.6; margin: -4px 0 0; }
.form-footer { text-align: center; font-size: 0.85rem; margin: 0; }

@media (max-width: 1023px) {
  .auth-hero { display: none; }
  .auth-form-panel { padding: 32px 20px; }
  .mobile-logo { display: flex; }
}
@media (max-width: 639px) {
  .code-inputs { gap: 6px; }
  .code-box { height: 54px; font-size: 1.35rem; }
}
</style>
