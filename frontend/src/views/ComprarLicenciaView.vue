<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { toast } from '../utils/toast'
import logoSrc from '../assets/logo-capacitaciones.png'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const licencia = ref<any>(null)
const purchasing = ref(false)

async function loadLicencia() {
  try {
    const res = await api.get(`/licencias-publicas/${route.params.id}`)
    licencia.value = res.data
  } catch (e: any) {
    toast.error('Licencia corporativa no encontrada')
    router.push('/tienda')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadLicencia()
})

async function buyLicencia() {
  const token = localStorage.getItem('token')
  if (!token) {
    // Si no está autenticado, lo mandamos al login y que regrese acá
    sessionStorage.setItem('redirect_after_login', route.fullPath)
    toast.info('Crea una cuenta o inicia sesión primero')
    router.push('/login')
    return
  }

  purchasing.value = true
  try {
    const res = await api.post('/checkout-session', {
      licencia_id: licencia.value.id,
      success_url: window.location.origin + `/comprar-licencia/${licencia.value.id}/success`,
      cancel_url: window.location.href,
    })
    if (res.data.url) {
      window.location.href = res.data.url
    } else {
      throw new Error("No URL returned from Stripe")
    }
  } catch (e: any) {
    toast.error('Error al iniciar el proceso de pago B2B. Intenta más tarde.')
    purchasing.value = false
  }
}
</script>

<template>
  <div class="comprar-licencia-view">
    <!-- Navbar simple -->
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

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Cargando información corporativa...</p>
    </div>

    <!-- Content -->
    <div v-else-if="licencia" class="licencia-content">
      <div class="glass-card main-card">
        <div class="card-header">
          <h1>{{ licencia.nombre }}</h1>
          <p class="subtitle">Adquisición de Licencia Corporativa</p>
        </div>

        <div class="course-preview">
          <div class="thumbnail" :style="{ backgroundImage: `url(${licencia.capacitacion_thumbnail || 'https://images.unsplash.com/photo-1552664730-d307ca884978?q=80&w=2940&auto=format&fit=crop'})` }"></div>
          <div class="course-info">
            <small>Aplica para el curso:</small>
            <h3>{{ licencia.capacitacion_titulo }}</h3>
          </div>
        </div>

        <div class="details-grid">
          <div class="detail-box">
            <div class="icon glass-icon">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
            </div>
            <div class="detail-text">
              <span class="label">Capacidad Máxima</span>
              <span class="value">{{ licencia.capacidad_maxima > 0 ? `${licencia.capacidad_maxima} empleados` : 'Ilimitado' }}</span>
            </div>
          </div>
          <div class="detail-box">
            <div class="icon glass-icon">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>
            </div>
            <div class="detail-text">
              <span class="label">Precio Total</span>
              <span class="value price">${{ licencia.precio }} MXN</span>
            </div>
          </div>
        </div>

        <div class="purchase-action">
          <p class="disclaimer">Al pagar, obtendrás un <strong>Código de Acceso Corporativo</strong> que podrás repartir entre tus colaboradores.</p>
          <button class="btn-buy" :disabled="purchasing" @click="buyLicencia">
            <span v-if="purchasing">Procesando...</span>
            <span v-else>Pagar con Tarjeta</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.comprar-licencia-view {
  min-height: 100vh;
  background: radial-gradient(circle at top left, #1a1a2e 0%, #0f0f1a 100%);
  color: #fff;
  font-family: 'Inter', sans-serif;
  display: flex;
  flex-direction: column;
}

/* Navbar */
.top-nav {
  padding: 20px 0;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}
.nav-content {
  max-width: 1000px;
  margin: 0 auto;
  padding: 0 24px;
}
.logo-link {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
}
.logo-img { height: 32px; filter: drop-shadow(0 0 8px rgba(249,115,22,0.5)); }
.logo-text { font-size: 1.25rem; font-weight: 800; letter-spacing: -0.5px; display: flex; align-items: center; gap: 6px; }
.white { color: #fff; }
.orange { color: #f97316; }
.badge-b2b {
  font-size: 0.7rem;
  background: rgba(249,115,22,0.2);
  color: #f97316;
  padding: 2px 8px;
  border-radius: 12px;
  margin-left: 8px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

/* Loading */
.loading-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: #94a3b8;
}
.spinner {
  width: 40px; height: 40px;
  border: 4px solid rgba(255,255,255,0.1);
  border-top-color: #f97316;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Content */
.licencia-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
}

/* Glass Card */
.main-card {
  width: 100%;
  max-width: 500px;
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  padding: 32px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
@keyframes slideUp { from { opacity: 0; transform: translateY(30px); } to { opacity: 1; transform: translateY(0); } }

.card-header {
  text-align: center;
  margin-bottom: 32px;
}
.card-header h1 {
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, #fff, #a5b4fc);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.subtitle {
  color: #94a3b8;
  font-size: 0.95rem;
  margin: 0;
}

.course-preview {
  display: flex;
  align-items: center;
  gap: 16px;
  background: rgba(0,0,0,0.2);
  padding: 12px;
  border-radius: 16px;
  margin-bottom: 24px;
}
.thumbnail {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  background-size: cover;
  background-position: center;
}
.course-info small { color: #94a3b8; font-size: 0.8rem; }
.course-info h3 { margin: 4px 0 0 0; font-size: 1.1rem; font-weight: 600; color: #fff; }

.details-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 32px;
}
.detail-box {
  background: rgba(255,255,255,0.05);
  padding: 16px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.icon.glass-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(255,255,255,0.15) 0%, rgba(255,255,255,0.05) 100%);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  color: #fff;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
  flex-shrink: 0;
}
.icon.glass-icon svg {
  width: 20px;
  height: 20px;
  opacity: 0.9;
}
.detail-text { display: flex; flex-direction: column; gap: 4px; }
.label { font-size: 0.75rem; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px; }
.value { font-size: 1.1rem; font-weight: 600; color: #fff; }
.price { color: #f97316; font-size: 1.3rem; }

.purchase-action {
  text-align: center;
}
.disclaimer {
  color: #94a3b8;
  font-size: 0.85rem;
  line-height: 1.5;
  margin-bottom: 24px;
}
.disclaimer strong { color: #fff; }

.btn-buy {
  width: 100%;
  padding: 18px;
  background: linear-gradient(135deg, #f97316, #fb923c);
  color: #fff;
  border: none;
  border-radius: 14px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 10px 25px -5px rgba(249, 115, 22, 0.4);
}
.btn-buy:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 15px 30px -5px rgba(249, 115, 22, 0.5);
}
.btn-buy:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}
</style>
