<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../../api'
import { toast } from '../../utils/toast'

const licencias = ref<any[]>([])
const loading = ref(true)
const verifying = ref(false)
const route = useRoute()
const router = useRouter()

async function verifySession(sessionId: string) {
  verifying.value = true
  try {
    await api.post('/verify-checkout-session', { session_id: sessionId })
    toast.success('¡Licencias generadas correctamente!')
  } catch (e: any) {
    // Si ya fue procesado antes (duplicate), no es un error real
    const msg = e.response?.data?.error || ''
    if (!msg.includes('ya existe') && !msg.includes('conflict')) {
      console.warn('verify-checkout-session:', msg)
    }
  } finally {
    verifying.value = false
    // Limpiar el session_id de la URL para que no se reprocese al refrescar
    router.replace({ path: '/usuario/licencias' })
  }
}

async function fetchLicencias() {
  loading.value = true
  try {
    const res = await api.get('/licencias-compradas')
    licencias.value = res.data || []
  } catch (e: any) {
    toast.error('Error al cargar tus licencias')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const sessionId = route.query.session_id as string | undefined
  if (sessionId) {
    await verifySession(sessionId)
  }
  await fetchLicencias()
})

function copyCode(codigo: string) {
  navigator.clipboard.writeText(codigo)
  toast.success('Código copiado al portapapeles')
}
</script>

<template>
  <div class="mis-licencias">
    <div class="header-section">
      <div class="glass-icon-wrapper">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="glass-icon"><path d="M4 10h16"/><path d="M4 14h16"/><path d="M14 18V6a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2z"/><path d="M18 22H4"/><path d="M22 18V6a2 2 0 0 0-2-2h-4"/></svg>
      </div>
      <h2>Mis Licencias Corporativas</h2>
      <p class="subtitle">Gestiona los accesos que has adquirido para tu equipo.</p>
    </div>

    <div v-if="loading" class="loading">Cargando tus licencias...</div>
    
    <div v-else-if="licencias.length === 0" class="empty-state">
      <div class="empty-icon glass-icon-wrapper-large">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="glass-icon"><rect width="20" height="14" x="2" y="7" rx="2" ry="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/></svg>
      </div>
      <h3>Aún no has adquirido licencias</h3>
      <p>Cuando compres accesos grupales para tu empresa, aparecerán aquí con sus respectivos códigos de invitación.</p>
      <router-link to="/tienda" class="btn btn-primary mt-4">Explorar Catálogo</router-link>
    </div>

    <div v-else class="licencias-grid">
      <div v-for="lic in licencias" :key="lic.id" class="licencia-card">
        <div class="lic-header">
          <h3>{{ lic.nombre }}</h3>
          <span class="status-badge">Activa</span>
        </div>
        
        <div class="lic-body">
          <div class="stats-row">
            <div class="stat">
              <span class="label">Asientos Usados</span>
              <span class="value">{{ lic.usadas }} / {{ lic.capacidad_maxima > 0 ? lic.capacidad_maxima : '∞' }}</span>
            </div>
            <div class="stat">
              <span class="label">Fecha de Compra</span>
              <span class="value">{{ new Date(lic.created_at).toLocaleDateString() }}</span>
            </div>
          </div>

          <div class="code-section">
            <p class="code-instruction">Envía este código a tu equipo para que puedan acceder al curso:</p>
            <div class="code-box" @click="copyCode(lic.codigo_acceso)">
              {{ lic.codigo_acceso }}
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="copy-icon"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.mis-licencias {
  display: flex;
  flex-direction: column;
  gap: 24px;
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

.header-section {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}
.header-section h2 { margin: 0 0 8px 0; color: var(--dark); font-size: 1.8rem; width: 100%; }
.subtitle { margin: 0; color: var(--muted); width: 100%; }

.glass-icon-wrapper {
  background: rgba(249, 115, 22, 0.1);
  padding: 12px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(249, 115, 22, 0.2);
  backdrop-filter: blur(4px);
}
.glass-icon-wrapper-large {
  background: rgba(249, 115, 22, 0.1);
  padding: 24px;
  border-radius: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(249, 115, 22, 0.2);
  backdrop-filter: blur(4px);
}
.glass-icon {
  color: var(--primary);
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  background: var(--surface);
  border-radius: var(--r-lg);
  border: 1px dashed var(--border);
}
.empty-icon { font-size: 3rem; margin-bottom: 16px; opacity: 0.8; }
.empty-state h3 { margin: 0 0 8px 0; color: var(--dark); }
.empty-state p { color: var(--muted); max-width: 400px; margin: 0 auto; }
.mt-4 { margin-top: 16px; }

.licencias-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 24px;
}

.licencia-card {
  background: var(--surface);
  border-radius: var(--r-lg);
  border: 1px solid var(--border-light);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
  transition: transform 0.2s, box-shadow 0.2s;
}
.licencia-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.lic-header {
  padding: 20px;
  border-bottom: 1px solid var(--border-light);
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  background: rgba(52, 211, 153, 0.05);
}
.lic-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--dark);
  font-weight: 600;
}
.status-badge {
  background: #34d399;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 12px;
  text-transform: uppercase;
}

.lic-body {
  padding: 20px;
}

.stats-row {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
}
.stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.stat .label {
  font-size: 0.8rem;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.stat .value {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--dark);
}

.code-section {
  background: rgba(0,0,0,0.02);
  padding: 16px;
  border-radius: var(--r-md);
}
.code-instruction {
  margin: 0 0 12px 0;
  font-size: 0.85rem;
  color: var(--muted);
}
.code-box {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border: 2px dashed #34d399;
  padding: 12px 16px;
  border-radius: 8px;
  font-family: monospace;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--dark);
  cursor: pointer;
  transition: all 0.2s;
}
.code-box:hover {
  background: rgba(52, 211, 153, 0.1);
}
.copy-icon {
  font-size: 1.2rem;
  opacity: 0.5;
}
</style>
