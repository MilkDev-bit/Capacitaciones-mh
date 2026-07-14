<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const curso = ref<any>(null)
const loading = ref(true)
const joining = ref(false)
const error = ref('')

const roomId = route.params.id as string
const codigoInput = ref(route.query.codigo as string || '')

onMounted(async () => {
  try {
    const res = await api.get(`/cursos-publicos/${roomId}`)
    curso.value = res.data
    // Auto-join if code is present in URL and user is logged in
    if (codigoInput.value && auth.isLoggedIn) {
      unirse()
    }
  } catch {
    error.value = 'El enlace de la videollamada no es válido o ya no existe.'
  } finally {
    loading.value = false
  }
})

async function unirse() {
  const code = codigoInput.value.trim().toUpperCase()
  if (!code) {
    toast.error('Por favor ingresa tu código de acceso')
    return
  }

  if (!auth.isLoggedIn) {
    // Redirect to login, preserving the room id and code
    router.push(`/login?redirect=/invitacion/${roomId}?codigo=${code}`)
    return
  }
  
  joining.value = true
  try {
    const res = await api.post('/videocalls/join', { codigo: code })
    toast.success('Uniéndose a la videollamada...')
    setTimeout(() => {
      router.push(`/usuario/videocall/${res.data.room_name}?codigo=${code}`)
    }, 1000)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al unirse a la videollamada')
  } finally {
    joining.value = false
  }
}
</script>

<template>
  <div class="join-bg">
    <div class="join-card">
    
      <div class="brand-bar">
        <svg width="28" height="28" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="8" fill="#f97316"/>
          <path d="M10 28L20 12L30 28H10Z" fill="white"/>
        </svg>
        <span class="brand-name">Capacitaciones MH</span>
      </div>

      <div v-if="loading" class="state-center">
        <div class="spin-ring"></div>
        <p class="state-text">Cargando sala&hellip;</p>
      </div>

      <div v-else-if="curso" class="course-preview">
        <div class="invite-badge">Invitaci&oacute;n a Videollamada</div>
        <h2 class="course-title">{{ curso.title }}</h2>
        <p class="course-desc">{{ curso.description || 'Sin descripci\u00f3n disponible.' }}</p>
        
        <div class="code-section">
          <label for="code-input">Ingresa tu c&oacute;digo &uacute;nico de acceso:</label>
          <input 
            id="code-input"
            v-model="codigoInput" 
            class="code-field" 
            placeholder="VC-ABC12345" 
            maxlength="20"
            autocomplete="off"
            @keyup.enter="unirse"
          />
        </div>

        <button class="btn-join" :disabled="joining" @click="unirse">
          {{ joining ? 'Conectando\u2026' : (auth.isLoggedIn ? 'Unirme ahora' : 'Iniciar sesi\u00f3n para unirme') }}
        </button>
      </div>

      <div v-else class="state-center">
        <div class="icon-circle danger">&#10007;</div>
        <h2 class="state-title">Enlace inv&aacute;lido</h2>
        <p class="state-text">{{ error }}</p>
        <button class="btn btn-secondary" style="margin-top:12px" @click="router.push('/')">Volver al inicio</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.join-bg {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #1e3a8a 0%, #3b82f6 60%, #93c5fd 100%);
  padding: 20px;
}
.join-card {
  background: var(--surface); border-radius: 20px; padding: 36px 32px;
  width: 100%; max-width: 460px;
  box-shadow: 0 24px 60px rgba(0,0,0,.22);
  display: flex; flex-direction: column; gap: 20px;
  color: var(--text);
}
.brand-bar { display: flex; align-items: center; gap: 10px; padding-bottom: 16px; border-bottom: 1px solid var(--border-light); }
.brand-name { font-size: 1rem; font-weight: 800; color: var(--dark); letter-spacing: -.01em; }
.state-center { text-align: center; padding: 8px 0; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.state-title { font-size: 1.25rem; font-weight: 800; color: var(--dark); }
.state-text { color: var(--muted); font-size: 0.9rem; }
.spin-ring {
  width: 40px; height: 40px; border-radius: 50%;
  border: 4px solid rgba(59,130,246,.2); border-top-color: #3b82f6;
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.icon-circle {
  width: 60px; height: 60px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.6rem; font-weight: 900;
}
.icon-circle.danger { background: var(--danger-bg); color: var(--danger); }
.course-preview { display: flex; flex-direction: column; gap: 14px; }
.invite-badge { display: inline-block; background: rgba(59,130,246,0.1); color: #2563eb; font-size: 0.72rem; font-weight: 700; padding: 4px 12px; border-radius: 20px; letter-spacing: .04em; width: fit-content; text-transform: uppercase; }
.course-title { font-size: 1.4rem; font-weight: 800; color: var(--dark); line-height: 1.3; }
.course-desc { color: var(--muted); font-size: 0.95rem; line-height: 1.5; }

.code-section {
  display: flex; flex-direction: column; gap: 8px; margin-top: 10px; margin-bottom: 10px;
}
.code-section label {
  font-size: 0.85rem; font-weight: 600; color: var(--dark);
}
.code-field {
  font-size: 1.3rem; font-weight: 800; letter-spacing: .15em; color: var(--text);
  font-family: 'Courier New', monospace; background: var(--bg);
  border: 2px solid var(--border); border-radius: 12px;
  padding: 12px 14px; text-align: center; outline: none; transition: all 0.2s;
}
.code-field:focus {
  border-color: #3b82f6; background: var(--surface); color: var(--text); box-shadow: 0 0 0 4px rgba(59,130,246,0.15);
}

.btn-join {
  background: #2563eb; color: #fff; border: none; border-radius: 12px;
  padding: 16px; font-size: 1.05rem; font-weight: 700; cursor: pointer;
  width: 100%; display: flex; align-items: center; justify-content: center;
  transition: background .15s, transform 0.1s;
}
.btn-join:hover:not(:disabled) { background: #1d4ed8; transform: translateY(-1px); }
.btn-join:active:not(:disabled) { transform: translateY(1px); }
.btn-join:disabled { opacity: .6; cursor: not-allowed; }
@media (max-width: 480px) { .join-card { padding: 24px 18px; } }
</style>
