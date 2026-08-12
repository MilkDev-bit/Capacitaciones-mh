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
  <div class="verify-page">
    <div class="verify-card">
      <div class="icon-badge">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect width="20" height="16" x="2" y="4" rx="2" />
          <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
        </svg>
      </div>

      <h1>Verifica tu correo</h1>
      <p class="lead">
        Enviamos un código de 6 dígitos a<br />
        <strong>{{ email }}</strong>
      </p>

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

      <button class="btn-verify" :disabled="!complete || loading" @click="verify">
        {{ loading ? 'Verificando…' : 'Verificar y entrar' }}
      </button>

      <div class="resend-row">
        <span v-if="cooldown > 0" class="muted">Podrás reenviar el código en {{ cooldown }}s</span>
        <button v-else class="btn-link" :disabled="resending" @click="resend">
          {{ resending ? 'Enviando…' : 'Reenviar código' }}
        </button>
      </div>

      <p class="footnote">
        ¿No lo encuentras? Revisa la carpeta de spam o correo no deseado.
        El código expira a los 15 minutos.
      </p>

      <router-link to="/login" class="btn-back">← Volver al inicio de sesión</router-link>
    </div>
  </div>
</template>

<style scoped>
.verify-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: linear-gradient(160deg, #fff7ed 0%, #f4f4f5 60%);
}

.verify-card {
  width: 100%;
  max-width: 440px;
  background: #fff;
  border-radius: 24px;
  padding: 40px 36px 32px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.08);
  text-align: center;
}

.icon-badge {
  width: 56px;
  height: 56px;
  margin: 0 auto 20px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(249, 115, 22, 0.1);
  color: #f97316;
}

h1 { margin: 0 0 10px; font-size: 1.5rem; font-weight: 800; color: #111827; }
.lead { margin: 0 0 28px; color: #6b7280; font-size: 0.95rem; line-height: 1.6; }
.lead strong { color: #111827; word-break: break-all; }

.code-inputs { display: flex; gap: 10px; justify-content: center; margin-bottom: 24px; }

.code-box {
  width: 48px;
  height: 58px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  text-align: center;
  font-size: 1.6rem;
  font-weight: 800;
  color: #111827;
  background: #f9fafb;
  transition: border-color 0.15s, background 0.15s, transform 0.15s;
}
.code-box:focus {
  outline: none;
  border-color: #f97316;
  background: #fff;
  transform: translateY(-2px);
}
.code-box.filled { border-color: #fdba74; background: #fff; }
.code-box:disabled { opacity: 0.6; }

.btn-verify {
  width: 100%;
  padding: 15px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #f97316 0%, #ea580c 100%);
  color: #fff;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s, opacity 0.2s;
  box-shadow: 0 8px 20px rgba(249, 115, 22, 0.25);
}
.btn-verify:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 12px 26px rgba(249, 115, 22, 0.32); }
.btn-verify:disabled { opacity: 0.5; cursor: not-allowed; box-shadow: none; }

.resend-row { margin: 18px 0 8px; min-height: 22px; }
.muted { color: #9ca3af; font-size: 0.85rem; }
.btn-link {
  background: none; border: none; color: #f97316;
  font-size: 0.9rem; font-weight: 700; cursor: pointer; padding: 0;
}
.btn-link:hover { text-decoration: underline; }

.footnote { margin: 12px 0 22px; font-size: 0.8rem; color: #9ca3af; line-height: 1.6; }

.btn-back {
  display: inline-block; color: #6b7280; font-size: 0.85rem; text-decoration: none;
}
.btn-back:hover { color: #111827; }

@media (max-width: 420px) {
  .verify-card { padding: 32px 20px 26px; }
  .code-inputs { gap: 6px; }
  .code-box { width: 42px; height: 52px; font-size: 1.35rem; }
}
</style>
