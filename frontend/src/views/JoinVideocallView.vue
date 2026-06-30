<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const joining = ref(false)
const codigoInput = ref(route.query.code as string || '')

onMounted(() => {
  // Auto-join if code is present in URL and user is logged in
  if (codigoInput.value && auth.isLoggedIn) {
    unirse()
  }
})

async function unirse() {
  const code = codigoInput.value.trim().toUpperCase()
  if (!code) {
    toast.error('Por favor ingresa tu código de acceso')
    return
  }

  if (!auth.isLoggedIn) {
    router.push(`/login?redirect=/join?code=${code}`)
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

      <div class="course-preview">
        <div class="invite-badge">Acceso a Videollamada</div>
        <h2 class="course-title">Ingresa a tu sesión</h2>
        <p class="course-desc">Ingresa el código único que recibiste para entrar a la sala segura.</p>
        
        <div class="code-section">
          <label for="code-input">Código único de acceso:</label>
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
          {{ joining ? 'Conectando...' : (auth.isLoggedIn ? 'Unirme ahora' : 'Iniciar sesión para unirme') }}
        </button>
      </div>

    </div>
  </div>
</template>

<style scoped>
.join-bg {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
  background-image: 
    radial-gradient(circle at 15% 50%, rgba(249, 115, 22, 0.08), transparent 25%),
    radial-gradient(circle at 85% 30%, rgba(249, 115, 22, 0.08), transparent 25%);
  padding: 20px;
}

.join-card {
  width: 100%;
  max-width: 440px;
  background: rgba(30, 41, 59, 0.7);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 32px;
  box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5);
  animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.brand-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.brand-name {
  color: #fff;
  font-weight: 700;
  font-size: 1.1rem;
  letter-spacing: -0.01em;
}

.course-preview {
  text-align: center;
}

.invite-badge {
  display: inline-block;
  padding: 6px 12px;
  background: rgba(249, 115, 22, 0.15);
  color: #f97316;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 16px;
}

.course-title {
  color: #fff;
  font-size: 1.6rem;
  font-weight: 800;
  margin-bottom: 12px;
  line-height: 1.2;
}

.course-desc {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.95rem;
  line-height: 1.5;
  margin-bottom: 28px;
}

.code-section {
  text-align: left;
  margin-bottom: 24px;
}

.code-section label {
  display: block;
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.85rem;
  font-weight: 600;
  margin-bottom: 8px;
}

.code-field {
  width: 100%;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  font-family: monospace;
  font-size: 1.2rem;
  padding: 14px 16px;
  border-radius: 12px;
  text-align: center;
  letter-spacing: 0.05em;
  outline: none;
  transition: all 0.2s;
}

.code-field:focus {
  border-color: #f97316;
  background: rgba(15, 23, 42, 0.8);
  box-shadow: 0 0 0 4px rgba(249, 115, 22, 0.15);
}

.code-field::placeholder {
  color: rgba(255, 255, 255, 0.3);
}

.btn-join {
  width: 100%;
  background: linear-gradient(135deg, #f97316 0%, #ea580c 100%);
  color: #fff;
  border: none;
  padding: 14px 24px;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(249, 115, 22, 0.3);
}

.btn-join:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(249, 115, 22, 0.4);
}

.btn-join:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}
</style>
