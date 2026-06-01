<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { toast } from '../utils/toast'

const route = useRoute()
const router = useRouter()

const token = computed(() => (route.query.token as string) || '')
const password = ref('')
const confirm = ref('')
const showPass = ref(false)
const loading = ref(false)
const success = ref(false)

const passwordsMatch = computed(() => !confirm.value || password.value === confirm.value)

async function submit() {
  if (!token.value) {
    toast.error('Enlace inválido o expirado')
    return
  }
  if (password.value.length < 8) {
    toast.error('La contraseña debe tener al menos 8 caracteres')
    return
  }
  if (password.value !== confirm.value) {
    toast.error('Las contraseñas no coinciden')
    return
  }
  loading.value = true
  try {
    await api.post('/reset-password', { token: token.value, password: password.value })
    success.value = true
    toast.success('¡Contraseña actualizada! Redirigiendo...')
    setTimeout(() => router.push('/login'), 2500)
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Enlace inválido o expirado')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="reset-page">
    <!-- Hero lateral -->
    <div class="reset-hero" aria-hidden="true">
      <div class="bubbles">
        <span></span><span></span><span></span><span></span>
        <span></span><span></span><span></span><span></span>
      </div>
      <div class="hero-inner">
        <img src="/logo-capacitaciones.png" alt="Capacitaciones MH" class="hero-logo" />
        <h1 class="hero-title">Capacitaciones<br><span>MH</span></h1>
        <p class="hero-sub">Plataforma de aprendizaje empresarial</p>
      </div>
    </div>

    <!-- Panel del formulario -->
    <div class="reset-panel">
      <div class="reset-wrap">

        <!-- Logo móvil -->
        <div class="mobile-brand">
          <img src="/logo-capacitaciones.png" alt="Capacitaciones MH" class="mobile-logo-img" />
          <span>Capacitaciones MH</span>
        </div>

        <!-- Sin token -->
        <div v-if="!token" class="invalid-state">
          <div class="invalid-icon">
            <svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10"/>
              <path d="M15 9l-6 6M9 9l6 6"/>
            </svg>
          </div>
          <h2>Enlace inválido</h2>
          <p>Este enlace de recuperación no es válido o ya expiró. Solicita uno nuevo desde la pantalla de inicio de sesión.</p>
          <a href="/login" class="btn btn-primary btn-lg block-btn">Ir al inicio de sesión</a>
        </div>

        <!-- Éxito -->
        <div v-else-if="success" class="success-state">
          <div class="success-icon">
            <svg width="44" height="44" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
          </div>
          <h2>¡Contraseña actualizada!</h2>
          <p>Tu contraseña fue restablecida exitosamente. Serás redirigido al inicio de sesión en unos segundos.</p>
          <a href="/login" class="btn btn-primary btn-lg block-btn">Ir al inicio de sesión</a>
        </div>

        <!-- Formulario -->
        <template v-else>
          <div class="form-header">
            <div class="lock-icon">
              <svg width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                <path d="M7 11V7a5 5 0 0110 0v4"/>
              </svg>
            </div>
            <h2>Nueva contraseña</h2>
            <p>Elige una contraseña segura de al menos 8 caracteres.</p>
          </div>

          <form @submit.prevent="submit" class="reset-form">
            <div class="form-group">
              <label>Nueva contraseña</label>
              <div class="pass-wrap">
                <input
                  class="field-input"
                  v-model="password"
                  :type="showPass ? 'text' : 'password'"
                  placeholder="Mínimo 8 caracteres"
                  autocomplete="new-password"
                  required
                  minlength="8"
                />
                <button type="button" class="pass-toggle" @click="showPass = !showPass">
                  {{ showPass ? 'Ocultar' : 'Ver' }}
                </button>
              </div>
            </div>

            <div class="form-group">
              <label>Confirmar contraseña</label>
              <div class="pass-wrap">
                <input
                  class="field-input"
                  v-model="confirm"
                  :type="showPass ? 'text' : 'password'"
                  placeholder="Repite tu contraseña"
                  autocomplete="new-password"
                  required
                  minlength="8"
                />
              </div>
              <p v-if="confirm && !passwordsMatch" class="hint error-hint">
                Las contraseñas no coinciden
              </p>
            </div>

            <button
              type="submit"
              class="btn btn-primary btn-lg block-btn"
              :disabled="loading || !passwordsMatch || password.length < 8"
            >
              <span v-if="loading" class="btn-spinner"></span>
              {{ loading ? 'Actualizando...' : 'Restablecer contraseña' }}
            </button>

            <p class="back-link">
              <a href="/login">← Volver al inicio de sesión</a>
            </p>
          </form>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reset-page { display: flex; min-height: 100vh; }

/* ── Hero ── */
.reset-hero {
  flex: 0 0 42%;
  background:
    linear-gradient(145deg, rgba(28,29,31,.98) 0%, rgba(38,35,31,.96) 100%),
    linear-gradient(135deg, rgba(249,115,22,.22), rgba(37,99,235,.18));
  display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden;
}
.hero-inner { position: relative; z-index: 1; text-align: center; padding: 40px; }
.hero-logo { width: 80px; height: 80px; object-fit: contain; margin: 0 auto 20px; filter: drop-shadow(0 4px 16px rgba(249,115,22,.35)); }
.hero-title { font-size: 2.2rem; font-weight: 900; color: #fff; line-height: 1.15; margin: 0 0 12px; }
.hero-title span { color: var(--brand); }
.hero-sub { color: rgba(255,255,255,.6); font-size: 0.95rem; }

/* Bubbles */
.bubbles { position: absolute; inset: 0; overflow: hidden; pointer-events: none; z-index: 0; }
.bubbles span {
  position: absolute; bottom: -80px; border-radius: 50%;
  background: rgba(249,115,22,.1); border: 1px solid rgba(249,115,22,.16);
  animation: bubble-rise linear infinite;
}
.bubbles span:nth-child(1) { width:30px; height:30px; left:10%; animation-duration:8s;   animation-delay:0s   }
.bubbles span:nth-child(2) { width:18px; height:18px; left:28%; animation-duration:6s;   animation-delay:1.4s }
.bubbles span:nth-child(3) { width:46px; height:46px; left:45%; animation-duration:11s;  animation-delay:0.6s; background:rgba(255,255,255,.04) }
.bubbles span:nth-child(4) { width:14px; height:14px; left:62%; animation-duration:7s;   animation-delay:3s   }
.bubbles span:nth-child(5) { width:36px; height:36px; left:78%; animation-duration:9s;   animation-delay:1.8s }
.bubbles span:nth-child(6) { width:22px; height:22px; left:18%; animation-duration:10s;  animation-delay:4.5s }
.bubbles span:nth-child(7) { width:40px; height:40px; left:55%; animation-duration:7.5s; animation-delay:2.2s }
.bubbles span:nth-child(8) { width:16px; height:16px; left:88%; animation-duration:6.5s; animation-delay:0.9s }
@keyframes bubble-rise {
  0%  { transform:translateY(0) scale(1);      opacity:0  }
  12% { opacity:.85 }
  85% { opacity:.5  }
  100%{ transform:translateY(-110vh) scale(1.1); opacity:0 }
}

/* ── Panel ── */
.reset-panel { flex: 1; display: flex; align-items: center; justify-content: center; padding: 40px 24px; background: var(--bg); overflow-y: auto; }
.reset-wrap { width: 100%; max-width: 420px; }

.mobile-brand { display: none; align-items: center; gap: 10px; margin-bottom: 28px; }
.mobile-logo-img { width: 36px; height: 36px; object-fit: contain; }
.mobile-brand span { font-size: 1rem; font-weight: 800; color: var(--dark); }

/* ── Form header ── */
.form-header { text-align: center; margin-bottom: 28px; }
.lock-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 60px; height: 60px; border-radius: 16px;
  background: var(--brand-light, rgba(249,115,22,.12));
  color: var(--brand); margin-bottom: 14px;
}
.form-header h2 { font-size: 1.6rem; font-weight: 800; color: var(--dark); margin: 0 0 6px; }
.form-header p { color: var(--muted); font-size: 0.9rem; line-height: 1.5; margin: 0; }

/* ── Reset form ── */
.reset-form { display: flex; flex-direction: column; gap: 18px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.pass-wrap { position: relative; }
.pass-wrap .field-input { padding-right: 78px; }
.pass-toggle {
  position: absolute; right: 8px; top: 50%; transform: translateY(-50%);
  background: var(--border-light); border: none; border-radius: 6px;
  color: var(--dark); font-size: 0.76rem; font-weight: 800; cursor: pointer;
  padding: 7px 9px; transition: all 0.2s;
}
.pass-toggle:hover { background: var(--border); }
.hint { font-size: 0.75rem; margin-top: 4px; }
.error-hint { color: var(--danger); font-weight: 600; }
.block-btn { width: 100%; display: flex; justify-content: center; align-items: center; gap: 8px; }
.block-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-spinner { width: 16px; height: 16px; border: 2.5px solid rgba(255,255,255,.4); border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; flex-shrink: 0; }
@keyframes spin { to { transform: rotate(360deg); } }
.back-link { text-align: center; margin-top: 4px; }
.back-link a { color: var(--muted); font-size: 0.88rem; text-decoration: none; font-weight: 500; }
.back-link a:hover { color: var(--brand); text-decoration: underline; }

/* ── States ── */
.invalid-state, .success-state { text-align: center; }
.invalid-icon, .success-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 72px; height: 72px; border-radius: 50%; margin: 0 auto 20px;
}
.invalid-icon { background: rgba(239,68,68,.1); color: #ef4444; }
.success-icon { background: rgba(16,185,129,.1); color: #10b981; }
.invalid-state h2, .success-state h2 { font-size: 1.5rem; font-weight: 800; color: var(--dark); margin: 0 0 10px; }
.invalid-state p, .success-state p { color: var(--muted); font-size: 0.9rem; line-height: 1.6; margin: 0 0 24px; }

@media (max-width: 860px) {
  .reset-hero { display: none; }
  .mobile-brand { display: flex; }
}
</style>
