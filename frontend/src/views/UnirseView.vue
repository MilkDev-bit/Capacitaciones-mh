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
      <div class="brand">
        <svg width="32" height="32" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="8" fill="#3b82f6"/>
          <path d="M10 28L20 12L30 28H10Z" fill="white"/>
        </svg>
        <span>Capacitaciones MH</span>
      </div>

      <div v-if="loading" class="state">
        <div class="spinner"></div>
        <p>Buscando curso…</p>
      </div>

      <div v-else-if="joined" class="state success">
        <div class="check">✓</div>
        <h2>¡Te uniste al curso!</h2>
        <p>Redirigiendo a tus capacitaciones…</p>
      </div>

      <div v-else-if="curso" class="course-preview">
        <div class="invite-badge">Invitación a curso</div>
        <div class="code-display">{{ codigo.toUpperCase() }}</div>
        <h2>{{ curso.title }}</h2>
        <p class="course-desc">{{ curso.description || 'Sin descripción' }}</p>
        <span class="type-tag">{{ typeLabel(curso.type) }}</span>

        <div v-if="error" class="msg error">{{ error }}</div>

        <button v-if="auth.isLoggedIn" class="btn-join" :disabled="joining" @click="unirse">
          {{ joining ? 'Uniéndose…' : 'Unirme al curso' }}
        </button>
        <div v-else class="login-prompt">
          <p>Debes iniciar sesión para unirte al curso.</p>
          <button class="btn-join" @click="router.push(`/login?redirect=/unirse/${codigo}`)">
            Iniciar sesión
          </button>
        </div>
      </div>

      <div v-else class="state">
        <div class="error-icon">✗</div>
        <h2>Código no válido</h2>
        <p>{{ error }}</p>
        <button class="btn-back" @click="router.push('/')">Volver al inicio</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.join-bg {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #1e3a5f 0%, #3b82f6 100%);
}
.join-card {
  background: white; border-radius: 16px; padding: 2.5rem 2rem;
  width: 100%; max-width: 420px; box-shadow: 0 20px 40px rgba(0,0,0,0.2);
  display: flex; flex-direction: column; gap: 1rem;
}
.brand { display: flex; align-items: center; gap: 10px; font-size: 1rem; font-weight: 700; color: #1e3a5f; }
.state { text-align: center; padding: 1.5rem 0; }
.state h2 { font-size: 1.2rem; font-weight: 700; color: #1e293b; margin-bottom: 6px; }
.state p { color: #64748b; }
.spinner {
  width: 36px; height: 36px; border: 3px solid #e2e8f0; border-top-color: #3b82f6;
  border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 12px;
}
@keyframes spin { to { transform: rotate(360deg); } }
.state.success .check {
  width: 56px; height: 56px; background: #d1fae5; color: #059669;
  border-radius: 50%; font-size: 1.5rem; display: flex; align-items: center; justify-content: center;
  margin: 0 auto 12px; font-weight: 900;
}
.error-icon {
  width: 56px; height: 56px; background: #fee2e2; color: #dc2626;
  border-radius: 50%; font-size: 1.5rem; display: flex; align-items: center; justify-content: center;
  margin: 0 auto 12px; font-weight: 900;
}
.course-preview { display: flex; flex-direction: column; gap: 0.75rem; }
.invite-badge {
  display: inline-block; background: #eff6ff; color: #3b82f6;
  font-size: 0.75rem; font-weight: 700; padding: 3px 10px; border-radius: 20px;
}
.code-display {
  font-size: 2rem; font-weight: 900; letter-spacing: 0.3em; color: #1e293b;
  font-family: 'Courier New', monospace; background: #f8fafc;
  border: 2px dashed #cbd5e1; border-radius: 10px; padding: 12px; text-align: center;
}
.course-preview h2 { font-size: 1.3rem; font-weight: 800; color: #1e293b; margin: 0; }
.course-desc { color: #64748b; font-size: 0.9rem; line-height: 1.5; }
.type-tag { background: #ede9fe; color: #6d28d9; font-size: 0.78rem; font-weight: 600; padding: 3px 10px; border-radius: 20px; display: inline-block; }
.btn-join {
  background: #3b82f6; color: white; border: none; border-radius: 10px;
  padding: 12px; font-size: 1rem; font-weight: 700; cursor: pointer; width: 100%; transition: background 0.2s;
}
.btn-join:hover:not(:disabled) { background: #2563eb; }
.btn-join:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-back {
  background: #f1f5f9; color: #475569; border: none; border-radius: 8px;
  padding: 10px 20px; cursor: pointer; font-size: 0.9rem; font-weight: 600; margin-top: 8px;
}
.login-prompt { text-align: center; }
.login-prompt p { color: #64748b; font-size: 0.87rem; margin-bottom: 10px; }
.msg { font-size: 0.85rem; padding: 8px 12px; border-radius: 6px; }
.msg.error { background: #fee2e2; color: #dc2626; }
</style>
