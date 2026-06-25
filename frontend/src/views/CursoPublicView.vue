<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const id = route.params.id as string
const curso = ref<any>(null)
const licencias = ref<any[]>([])
const loading = ref(true)

const codigoForm = ref('')
const joiningCodigo = ref(false)

onMounted(async () => {
  try {
    // 1. Get course info. The preview endpoint takes code, but we need by ID.
    // Wait, do we have a public endpoint by ID? Let's check backend.
    // For now, let's assume we can fetch it, or we'll add a public endpoint if needed.
    const resC = await api.get(`/capacitaciones/${id}`) // This requires auth in current backend. We might need to handle this.
    curso.value = resC.data
    
    const resL = await api.get(`/capacitaciones/${id}/licencias`)
    licencias.value = resL.data || []
  } catch (e: any) {
    if (e.response?.status === 401) {
      toast.info('Debes iniciar sesión para ver este curso')
      router.push(`/login?redirect=/curso/${id}`)
    } else {
      toast.error('Error al cargar curso')
    }
  } finally {
    loading.value = false
  }
})

async function unirseConCodigo() {
  if (!codigoForm.value) return
  joiningCodigo.value = true
  try {
    // Try license code first
    await api.post('/inscripciones-licencia', {
      capacitacion_id: id,
      codigo_acceso: codigoForm.value
    })
    toast.success('Inscrito correctamente a la cohorte')
    router.push('/usuario/capacitaciones')
  } catch (e:any) {
    // If it fails, maybe it's a generic course code
    try {
      await api.post('/inscripciones', { codigo: codigoForm.value })
      toast.success('Inscrito correctamente al curso')
      router.push('/usuario/capacitaciones')
    } catch (err: any) {
      toast.error(err.response?.data?.error || e.response?.data?.error || 'Código inválido')
    }
  } finally {
    joiningCodigo.value = false
  }
}

async function buyLicencia(lic: any) {
  try {
    const res = await api.post('/checkout-session', {
      licencia_id: lic.id,
      success_url: window.location.origin + '/usuario/capacitaciones?success=true',
      cancel_url: window.location.href
    })
    if (res.data.url) {
      window.location.href = res.data.url
    }
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al iniciar pago')
  }
}
</script>

<template>
  <div class="landing-bg">
    <div class="header-bar">
      <div class="brand-bar" @click="router.push('/')">
        <svg width="28" height="28" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="8" fill="#f97316"/>
          <path d="M10 28L20 12L30 28H10Z" fill="white"/>
        </svg>
        <span class="brand-name">Capacitaciones MH</span>
      </div>
      <div v-if="auth.isLoggedIn" class="user-nav">
        <button class="btn btn-secondary small" @click="router.push('/usuario')">Ir al Panel</button>
      </div>
      <div v-else class="user-nav">
        <button class="btn btn-text" @click="router.push('/login')">Iniciar Sesión</button>
      </div>
    </div>

    <div v-if="loading" class="state-center" style="margin-top: 100px;">
      <div class="spin-ring"></div>
      <p>Cargando curso...</p>
    </div>

    <div v-else-if="curso" class="content-wrapper">
      <div class="hero-section" :style="{ background: curso.thumbnail_url ? `url(${curso.thumbnail_url}) center/cover` : (curso.color || '#f97316') }">
        <div class="hero-overlay">
          <h1>{{ curso.title }}</h1>
          <p>{{ curso.description || 'Sin descripción' }}</p>
        </div>
      </div>

      <div class="main-grid">
        <div class="left-col">
          <h3>Sobre este curso</h3>
          <p style="color: var(--muted); line-height: 1.6;">{{ curso.welcome_message || 'Aprende y desarrolla nuevas habilidades con este curso.' }}</p>
          
          <div class="code-section">
            <h4>¿Tienes un código de invitación?</h4>
            <div style="display:flex; gap: 8px;">
              <input v-model="codigoForm" class="field-input" placeholder="Ingresa tu código" />
              <button class="btn btn-primary" @click="unirseConCodigo" :disabled="joiningCodigo">
                {{ joiningCodigo ? 'Verificando...' : 'Canjear' }}
              </button>
            </div>
          </div>
        </div>

        <div class="right-col">
          <h3>Opciones de Acceso</h3>
          <div v-if="licencias.length === 0" class="empty-lic">
            Este curso no tiene licencias de pago disponibles.
          </div>
          
          <div v-else class="lic-list">
            <div v-for="lic in licencias" :key="lic.id" class="lic-card">
              <h4>{{ lic.nombre }}</h4>
              <div class="price">${{ lic.precio }} USD</div>
              <div class="cap" v-if="lic.capacidad_maxima > 0">
                Quedan {{ lic.capacidad_maxima - lic.usadas }} lugares
              </div>
              <button class="btn btn-primary w-100 mt-3" @click="buyLicencia(lic)">
                Comprar Licencia
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.landing-bg { min-height: 100vh; background: var(--bg); display: flex; flex-direction: column; }
.header-bar { display: flex; align-items: center; justify-content: space-between; padding: 16px 32px; background: #fff; border-bottom: 1px solid var(--border-light); }
.brand-bar { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.brand-name { font-size: 1.1rem; font-weight: 800; color: var(--dark); letter-spacing: -.01em; }
.content-wrapper { max-width: 1100px; margin: 0 auto; width: 100%; padding: 32px; }

.hero-section { border-radius: 24px; overflow: hidden; position: relative; min-height: 300px; display: flex; align-items: flex-end; margin-bottom: 40px; }
.hero-overlay { background: linear-gradient(to top, rgba(0,0,0,0.8), transparent); width: 100%; padding: 40px; color: #fff; }
.hero-overlay h1 { font-size: 2.5rem; font-weight: 900; margin: 0 0 12px 0; }
.hero-overlay p { font-size: 1.1rem; opacity: 0.9; margin: 0; max-width: 600px; }

.main-grid { display: grid; grid-template-columns: 2fr 1fr; gap: 40px; }
.left-col h3, .right-col h3 { font-size: 1.4rem; font-weight: 800; color: var(--dark); margin: 0 0 16px 0; }
.code-section { margin-top: 40px; padding: 24px; background: #fff; border-radius: 16px; border: 1px solid var(--border-light); }
.code-section h4 { margin: 0 0 16px 0; font-size: 1.1rem; color: var(--dark); }
.field-input { flex: 1; padding: 12px 16px; border: 1px solid var(--border); border-radius: var(--r-md); font-size: 1rem; outline: none; }
.field-input:focus { border-color: var(--brand); box-shadow: 0 0 0 3px var(--brand-light); }
.w-100 { width: 100%; }
.mt-3 { margin-top: 16px; }

.lic-list { display: flex; flex-direction: column; gap: 16px; }
.lic-card { background: #fff; padding: 24px; border-radius: 16px; border: 1px solid var(--border-light); box-shadow: 0 4px 12px rgba(0,0,0,0.05); text-align: center; }
.lic-card h4 { font-size: 1.15rem; font-weight: 800; margin: 0 0 8px 0; color: var(--dark); }
.price { font-size: 2rem; font-weight: 900; color: var(--brand); margin-bottom: 8px; }
.cap { font-size: 0.85rem; color: var(--muted); background: var(--surface-soft); padding: 4px 12px; border-radius: 20px; display: inline-block; }
.empty-lic { color: var(--muted); font-style: italic; }

.state-center { display: flex; flex-direction: column; align-items: center; gap: 16px; color: var(--muted); }
.spin-ring { width: 40px; height: 40px; border-radius: 50%; border: 3px solid rgba(249,115,22,.2); border-top-color: var(--brand); animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 768px) {
  .main-grid { grid-template-columns: 1fr; }
  .content-wrapper { padding: 16px; }
  .hero-section { min-height: 200px; }
  .hero-overlay h1 { font-size: 1.8rem; }
}
</style>
