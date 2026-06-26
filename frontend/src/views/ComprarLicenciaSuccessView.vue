<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { toast } from '../utils/toast'
import logoSrc from '../assets/logo-capacitaciones.png'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const licencia = ref<any>(null)
const justCopied = ref(false)

async function loadLicencia() {
  try {
    // Para ver el código de acceso, llamamos a la ruta de "Mis licencias compradas"
    // ya que requiere autenticación y ser el propietario
    const res = await api.get('/licencias-compradas')
    const match = res.data?.find((l: any) => l.id === route.params.id)
    if (match) {
      licencia.value = match
    } else {
      throw new Error("No tienes acceso a esta licencia")
    }
  } catch (e: any) {
    toast.error('Error al verificar tu compra. Revisa la sección Mis Licencias.')
    router.push('/usuario/licencias')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  // Simular un pequeño delay para asegurar que el webhook procesó
  setTimeout(() => loadLicencia(), 1500)
})

function copyCode() {
  if (licencia.value?.codigo_acceso) {
    navigator.clipboard.writeText(licencia.value.codigo_acceso)
    justCopied.value = true
    toast.success('Código copiado')
    setTimeout(() => justCopied.value = false, 2000)
  }
}
</script>

<template>
  <div class="success-view">
    <nav class="top-nav">
      <div class="nav-content">
        <router-link to="/tienda" class="logo-link">
          <img :src="logoSrc" alt="Logo" class="logo-img" v-if="logoSrc"/>
          <span class="logo-text">
            <span class="white">MH</span> <span class="orange">Capacitaciones</span> <span class="badge-b2b">Empresas</span>
          </span>
        </router-link>
      </div>
    </nav>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Confirmando tu pago...</p>
    </div>

    <div v-else-if="licencia" class="success-content">
      <div class="glass-card success-card">
        <div class="success-icon">✨</div>
        <h1>¡Compra Exitosa!</h1>
        <p class="subtitle">Ya tienes acceso corporativo para <strong>{{ licencia.nombre }}</strong>.</p>
        
        <div class="code-box">
          <p class="code-label">CÓDIGO DE ACCESO PARA TUS EMPLEADOS</p>
          <div class="code-value" @click="copyCode">
            {{ licencia.codigo_acceso }}
          </div>
          <button class="btn-copy" @click="copyCode">
            {{ justCopied ? '¡Copiado!' : 'Copiar Código' }}
          </button>
        </div>

        <div class="instructions">
          <h3>¿Cómo usarlo?</h3>
          <ol>
            <li>Comparte este código con tus colaboradores.</li>
            <li>Ellos deben entrar a nuestra plataforma y crear una cuenta gratuita.</li>
            <li>En su panel, ingresan el código en "Unirse con Código" y tendrán acceso inmediato al curso.</li>
          </ol>
        </div>

        <router-link to="/usuario/licencias" class="btn-primary">
          Ir a Mi Panel de Empresa
        </router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.success-view {
  min-height: 100vh;
  background: radial-gradient(circle at top, #13271f 0%, #0f0f1a 100%);
  color: #fff;
  font-family: 'Inter', sans-serif;
  display: flex;
  flex-direction: column;
}

.top-nav {
  padding: 20px 0;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}
.nav-content { max-width: 1000px; margin: 0 auto; padding: 0 24px; }
.logo-link { display: flex; align-items: center; gap: 12px; text-decoration: none; }
.logo-img { height: 32px; }
.logo-text { font-size: 1.25rem; font-weight: 800; display: flex; align-items: center; gap: 6px; }
.white { color: #fff; }
.orange { color: #f97316; }
.badge-b2b { font-size: 0.7rem; background: rgba(52,211,153,0.2); color: #34d399; padding: 2px 8px; border-radius: 12px; margin-left: 8px; text-transform: uppercase; }

.loading-state {
  flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px; color: #34d399;
}
.spinner {
  width: 40px; height: 40px; border: 4px solid rgba(52,211,153,0.2); border-top-color: #34d399; border-radius: 50%; animation: spin 1s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.success-content {
  flex: 1; display: flex; align-items: center; justify-content: center; padding: 40px 24px;
}
.success-card {
  width: 100%; max-width: 600px;
  background: rgba(255, 255, 255, 0.03); backdrop-filter: blur(20px);
  border: 1px solid rgba(52, 211, 153, 0.2); border-radius: 24px;
  padding: 40px; text-align: center;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  animation: popIn 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
@keyframes popIn { 0% { opacity: 0; transform: scale(0.95); } 100% { opacity: 1; transform: scale(1); } }

.success-icon { font-size: 4rem; margin-bottom: 16px; animation: float 3s ease-in-out infinite; }
@keyframes float { 0%,100% { transform: translateY(0); } 50% { transform: translateY(-10px); } }

h1 { font-size: 2.2rem; margin: 0 0 12px 0; color: #34d399; }
.subtitle { color: #94a3b8; font-size: 1.05rem; margin-bottom: 32px; }
.subtitle strong { color: #fff; }

.code-box {
  background: rgba(0,0,0,0.3);
  border: 1px dashed rgba(52, 211, 153, 0.4);
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 32px;
}
.code-label { font-size: 0.8rem; color: #94a3b8; letter-spacing: 1px; margin: 0 0 12px 0; }
.code-value {
  font-family: monospace; font-size: 2.5rem; font-weight: 700; color: #fff;
  letter-spacing: 4px; cursor: pointer; margin-bottom: 16px;
  text-shadow: 0 0 20px rgba(52, 211, 153, 0.4);
}
.btn-copy {
  background: rgba(52, 211, 153, 0.15); color: #34d399; border: 1px solid rgba(52, 211, 153, 0.3);
  padding: 8px 24px; border-radius: 20px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.btn-copy:hover { background: rgba(52, 211, 153, 0.25); transform: translateY(-2px); }

.instructions {
  text-align: left; background: rgba(255,255,255,0.02); padding: 24px; border-radius: 16px; margin-bottom: 32px;
}
.instructions h3 { margin: 0 0 16px 0; font-size: 1.1rem; color: #fff; }
.instructions ol { margin: 0; padding-left: 20px; color: #cbd5e1; line-height: 1.6; font-size: 0.95rem; }
.instructions li { margin-bottom: 8px; }

.btn-primary {
  display: inline-block; width: 100%; padding: 18px;
  background: linear-gradient(135deg, #34d399, #10b981); color: #000;
  border-radius: 14px; font-size: 1.1rem; font-weight: 700; text-decoration: none;
  transition: all 0.3s ease; box-shadow: 0 10px 25px -5px rgba(52, 211, 153, 0.3);
}
.btn-primary:hover { transform: translateY(-2px); box-shadow: 0 15px 30px -5px rgba(52, 211, 153, 0.4); }
</style>
