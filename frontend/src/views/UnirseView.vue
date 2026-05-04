<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const curso = ref<any>(null)
const loading = ref(true)
const joining = ref(false)
const error = ref('')
const joined = ref(false)

const codigo = route.params.codigo as string

onMounted(async () => {
  try {
    const res = await api.get(`/preview-curso/${codigo.toUpperCase()}`, { skipAuth: true } as any)
    curso.value = res.data
  } catch {
    error.value = 'El código no es válido o el curso no existe.'
  } finally {
    loading.value = false
  }
})

async function unirse() {
  if (!auth.isLoggedIn) {
    router.push(`/login?redirect=/unirse/${codigo}`)
    return
  }
  joining.value = true
  try {
    await api.post('/unirse-con-codigo', { codigo: codigo.toUpperCase() })
    joined.value = true
    setTimeout(() => router.push('/usuario/capacitaciones'), 1500)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al unirse al curso'
  } finally {
    joining.value = false
  }
}

function typeLabel(t: string) {
  return { video: '🎥 Video', document: '📄 Documento', text: '📝 Texto' }[t] || t
}
</script>

<template>
  <div class="join-bg">
    <div class="join-card">
      <!-- Brand -->
      <div class="brand-bar">
        <svg width="28" height="28" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="8" fill="#f97316"/>
          <path d="M10 28L20 12L30 28H10Z" fill="white"/>
        </svg>
        <span class="brand-name">Capacitaciones MH</span>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="state-center">
        <div class="spin-ring"></div>
        <p class="state-text">Buscando curso&hellip;</p>
      </div>

      <!-- &Eacute;xito al unirse -->
      <div v-else-if="joined" class="state-center">
        <div class="icon-circle success">&#10003;</div>
        <h2 class="state-title">&iexcl;Te uniste al curso!</h2>
        <p class="state-text">Redirigiendo a tus capacitaciones&hellip;</p>
      </div>

      <!-- Vista previa del curso -->
      <div v-else-if="curso" class="course-preview">
        <div class="invite-badge">Invitaci&oacute;n a curso</div>
        <div class="code-box">{{ codigo.toUpperCase() }}</div>
        <h2 class="course-title">{{ curso.title }}</h2>
        <p class="course-desc">{{ curso.description || 'Sin descripci\u00f3n' }}</p>
        <span class="type-tag">{{ typeLabel(curso.type) }}</span>

        <div v-if="error" class="alert alert-error" style="margin-top:8px">{{ error }}</div>

        <button v-if="auth.isLoggedIn" class="btn-join" :disabled="joining" @click="unirse">
          {{ joining ? 'Uni\u00e9ndose\u2026' : 'Unirme al curso' }}
        </button>
        <div v-else class="login-box">
          <p>Debes iniciar sesi&oacute;n para unirte.</p>
          <button class="btn-join" @click="router.push(`/login?redirect=/unirse/${codigo}`)">
            Iniciar sesi&oacute;n
          </button>
        </div>
      </div>

      <!-- Error / c&oacute;digo inv&aacute;lido -->
      <div v-else class="state-center">
        <div class="icon-circle danger">&#10007;</div>
        <h2 class="state-title">C&oacute;digo no v&aacute;lido</h2>
        <p class="state-text">{{ error }}</p>
        <button class="btn btn-secondary" style="margin-top:12px" @click="router.push('/')">Volver al inicio</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.join-bg {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #7c2d12 0%, #f97316 60%, #fbbf24 100%);
  padding: 20px;
}
.join-card {
  background: #fff; border-radius: 20px; padding: 36px 32px;
  width: 100%; max-width: 440px;
  box-shadow: 0 24px 60px rgba(0,0,0,.22);
  display: flex; flex-direction: column; gap: 20px;
}
.brand-bar { display: flex; align-items: center; gap: 10px; padding-bottom: 16px; border-bottom: 1px solid var(--border-light); }
.brand-name { font-size: 1rem; font-weight: 800; color: var(--dark); letter-spacing: -.01em; }
.state-center { text-align: center; padding: 8px 0; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.state-title { font-size: 1.25rem; font-weight: 800; color: var(--dark); }
.state-text { color: var(--muted); font-size: 0.9rem; }
.spin-ring {
  width: 40px; height: 40px; border-radius: 50%;
  border: 4px solid rgba(249,115,22,.2); border-top-color: var(--brand);
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.icon-circle {
  width: 60px; height: 60px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.6rem; font-weight: 900;
}
.icon-circle.success { background: var(--success-bg); color: var(--success); }
.icon-circle.danger { background: var(--danger-bg); color: var(--danger); }
.course-preview { display: flex; flex-direction: column; gap: 10px; }
.invite-badge { display: inline-block; background: var(--brand-light); color: var(--brand-dark); font-size: 0.72rem; font-weight: 700; padding: 3px 10px; border-radius: 20px; letter-spacing: .04em; width: fit-content; }
.code-box {
  font-size: 2rem; font-weight: 900; letter-spacing: .28em; color: var(--dark);
  font-family: 'Courier New', monospace; background: var(--bg);
  border: 2px dashed var(--brand-border); border-radius: 12px;
  padding: 14px; text-align: center;
}
.course-title { font-size: 1.3rem; font-weight: 800; color: var(--dark); line-height: 1.3; }
.course-desc { color: var(--muted); font-size: 0.9rem; line-height: 1.5; }
.type-tag { background: var(--brand-light); color: var(--brand-dark); font-size: 0.78rem; font-weight: 600; padding: 3px 10px; border-radius: 20px; display: inline-block; width: fit-content; }
.btn-join {
  background: var(--brand); color: #fff; border: none; border-radius: 10px;
  padding: 14px; font-size: 1rem; font-weight: 700; cursor: pointer;
  width: 100%; display: flex; align-items: center; justify-content: center;
  transition: background .15s;
}
.btn-join:hover:not(:disabled) { background: var(--brand-dark); }
.btn-join:disabled { opacity: .6; cursor: not-allowed; }
.login-box { text-align: center; display: flex; flex-direction: column; gap: 10px; }
.login-box p { color: var(--muted); font-size: 0.87rem; }
@media (max-width: 480px) { .join-card { padding: 24px 18px; } }
</style>
