<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
import { toast } from '../../utils/toast'

const licencias = ref<any[]>([])
const loading = ref(true)

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

onMounted(() => {
  fetchLicencias()
})

function copyCode(codigo: string) {
  navigator.clipboard.writeText(codigo)
  toast.success('Código copiado al portapapeles')
}
</script>

<template>
  <div class="mis-licencias">
    <div class="header-section">
      <h2>🏢 Mis Licencias Corporativas</h2>
      <p class="subtitle">Gestiona los accesos que has adquirido para tu equipo.</p>
    </div>

    <div v-if="loading" class="loading">Cargando tus licencias...</div>
    
    <div v-else-if="licencias.length === 0" class="empty-state">
      <div class="empty-icon">💼</div>
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
              <span class="copy-icon">📋</span>
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

.header-section h2 { margin: 0 0 8px 0; color: var(--dark); font-size: 1.8rem; }
.subtitle { margin: 0; color: var(--muted); }

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
