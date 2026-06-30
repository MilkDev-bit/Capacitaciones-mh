<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../../api'
import { toast } from '../../utils/toast'
import { useCartStore } from '../../stores/cart'

const cartStore = useCartStore()
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
    cartStore.clearCart()
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
const selectedLic = ref<any>(null)
const selectedTickets = ref<any[]>([])
const loadingTickets = ref(false)
const invoiceLoading = ref(false)

async function openDetails(lic: any) {
  selectedLic.value = lic
  selectedTickets.value = []
  loadingTickets.value = true
  try {
    const res = await api.get(`/licencias/${lic.id}/tickets`)
    selectedTickets.value = res.data || []
  } catch {
    selectedTickets.value = []
  } finally {
    loadingTickets.value = false
  }
}

const baseOrigin = ref(window.location.origin)

function closeModal() {
  selectedLic.value = null
  selectedTickets.value = []
}

async function downloadInvoice(lic: any) {
  invoiceLoading.value = true
  try {
    const res = await api.get(`/licencias/${lic.id}/invoice`)
    const pdfUrl = res.data.invoice_pdf || res.data.invoice_url
    if (pdfUrl) {
      window.open(pdfUrl, '_blank')
    } else {
      toast.error('No se encontró la factura para esta licencia')
    }
  } catch {
    toast.error('Factura no disponible. Solo aparece si la compra fue realizada con la facturación activada en Stripe.')
  } finally {
    invoiceLoading.value = false
  }
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    toast.success('Copiado al portapapeles')
  } catch (err) {
    toast.error('Error al copiar')
  }
}

async function copyCode(codigo: string) {
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

    <div v-if="loading || verifying" class="loading">
      {{ verifying ? 'Procesando tu compra...' : 'Cargando tus licencias...' }}
    </div>
    
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
              <span class="label">Lugares usados</span>
              <span class="value">{{ lic.usadas }} / {{ lic.capacidad_maxima > 0 ? lic.capacidad_maxima : '∞' }}</span>
            </div>
            <div class="stat">
              <span class="label">Fecha de Compra</span>
              <span class="value">{{ new Date(lic.created_at).toLocaleDateString() }}</span>
            </div>
            <div class="stat">
              <span class="label">Precio total</span>
              <span class="value">${{ lic.precio?.toLocaleString('es-MX', { minimumFractionDigits: 2 }) }} MXN</span>
            </div>
          </div>

          <div class="code-section" v-if="lic.curso_type === 'videocall'">
            <p class="code-instruction">Esta es una capacitación por <strong>Videollamada</strong> ({{ lic.curso_duracion }} min). En lugar de un solo código, se generaron códigos únicos para cada participante.</p>
            <p class="code-instruction" style="margin-top: 8px;">Abre los detalles para ver y copiar los códigos individuales.</p>
          </div>
          <div class="code-section" v-else>
            <p class="code-instruction">Envía este código a tu equipo para que puedan acceder al curso:</p>
            <div class="code-box" @click="copyCode(lic.codigo_acceso)">
              {{ lic.codigo_acceso }}
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="copy-icon"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
            </div>
          </div>

          <button class="btn-details" @click="openDetails(lic)">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
            Ver Detalles y Factura
          </button>
        </div>
      </div>
    </div>

    <!-- Modal de detalles -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="selectedLic" class="modal-overlay" @click.self="closeModal">
          <div class="modal-card">
            <div class="modal-header">
              <h3>Detalles de Licencia</h3>
              <button class="modal-close" @click="closeModal">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="modal-body">
              <div class="detail-row"><span class="detail-label">Nombre</span><span class="detail-value">{{ selectedLic.nombre }}</span></div>
              <div class="detail-row"><span class="detail-label">Código de acceso</span><span class="detail-value mono">{{ selectedLic.codigo_acceso }}</span></div>
              <div class="detail-row"><span class="detail-label">Lugares totales</span><span class="detail-value">{{ selectedLic.capacidad_maxima > 0 ? selectedLic.capacidad_maxima : 'Ilimitados' }}</span></div>
              <div class="detail-row"><span class="detail-label">Lugares usados</span><span class="detail-value">{{ selectedLic.usadas || 0 }}</span></div>
              <div class="detail-row"><span class="detail-label">Fecha de Compra</span><span class="detail-value">{{ new Date(selectedLic.created_at).toLocaleDateString() }}</span></div>
              <div class="detail-row"><span class="detail-label">Precio Pagado</span><span class="detail-value">${{ selectedLic.precio?.toLocaleString('es-MX', { minimumFractionDigits: 2 }) }} MXN</span></div>
              <div class="detail-row"><span class="detail-label">Factura</span><span class="detail-value">
                <button class="btn btn-secondary btn-sm" @click="downloadInvoice(selectedLic)" :disabled="invoiceLoading">
                  {{ invoiceLoading ? 'Descargando...' : 'Descargar Factura PDF' }}
                </button>
              </span></div>

              <div v-if="selectedTickets && selectedTickets.length > 0" class="tickets-section">
                <h4 style="margin-top: 1.5rem; margin-bottom: 0.5rem; color: var(--text-color);">Enlace de Invitación</h4>
                <p style="font-size: 0.85rem; color: var(--text-muted); margin-bottom: 0.5rem;">Comparte este enlace junto con los códigos a tus participantes.</p>
                <div class="invite-link-box">
                  <span class="mono link-text">{{ baseOrigin }}/invitacion/{{ selectedLic.capacitacion_id }}</span>
                  <button class="btn-copy-small" @click="copyText(`${baseOrigin}/invitacion/${selectedLic.capacitacion_id}`)" title="Copiar enlace">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
                  </button>
                </div>

                <h4 style="margin-top: 1.5rem; margin-bottom: 0.5rem; color: var(--text-color);">Códigos Únicos para Videollamada</h4>
                <p style="font-size: 0.85rem; color: var(--text-muted); margin-bottom: 1rem;">Envía un código distinto a cada participante de la videollamada.</p>
                <div v-if="loadingTickets" class="loading" style="font-size: 0.9rem;">Cargando códigos...</div>
                <div v-else class="tickets-list">
                  <div v-for="(t, i) in selectedTickets" :key="t.id" class="ticket-item">
                    <span class="ticket-num">#{{ i + 1 }}</span>
                    <span class="ticket-code mono">{{ t.codigo }}</span>
                    <span :class="['ticket-status', t.in_use_by_user ? 'status-used' : 'status-free']">
                      {{ t.in_use_by_user ? 'En uso' : 'Disponible' }}
                    </span>
                    <button class="btn-copy-small" @click="copyCode(t.codigo)" title="Copiar código">
                      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button style="display: flex; align-items: center; justify-content: center; gap: 8px; background-color: #f97316; color: #ffffff; border: none; border-radius: 10px; padding: 10px 18px; font-size: 0.9rem; font-weight: 600; cursor: pointer; flex: 1; opacity: 1 !important; visibility: visible !important;" class="action-doc-btn" @click="downloadInvoice(selectedLic)" :disabled="invoiceLoading">
                <svg v-if="!invoiceLoading" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                <span v-if="invoiceLoading">Buscando documento...</span>
                <span v-else>Descargar Recibo</span>
              </button>
              <button class="btn-secondary" @click="copyCode(selectedLic.codigo_acceso)">Copiar Código</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
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

/* ── Modal & Buttons ───────────────────────────────────────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
}

.modal-card {
  background: var(--surface);
  border-radius: 20px;
  width: 100%;
  max-width: 520px;
  box-shadow: 0 24px 60px rgba(0,0,0,0.2);
  overflow: hidden;
  border: 1px solid var(--border);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--dark);
}

.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--muted);
  padding: 4px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.modal-close:hover { background: var(--border); color: var(--dark); }

.modal-body {
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 50vh;
  overflow-y: auto;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border);
}
.detail-row:last-child { border-bottom: none; }

.detail-label {
  font-size: 0.82rem;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
  flex-shrink: 0;
}

.detail-value {
  font-size: 0.9rem;
  color: var(--dark);
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.detail-value.mono { font-family: monospace; }
.detail-value.small { font-size: 0.75rem; }
.detail-value.highlight { color: var(--primary); font-weight: 700; }

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border);
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.btn-comprobante {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  flex: 1;
  justify-content: center;
}
.btn-comprobante:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-comprobante:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-secondary {
  background: var(--border);
  color: var(--dark);
  border: none;
  border-radius: 10px;
  padding: 10px 16px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-secondary:hover { background: var(--muted); color: white; }

.btn-details {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  width: 100%;
  justify-content: center;
  background: rgba(249, 115, 22, 0.08);
  color: var(--primary);
  border: 1px solid rgba(249, 115, 22, 0.25);
  border-radius: 10px;
  padding: 10px;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-details:hover {
  background: rgba(249, 115, 22, 0.15);
  transform: translateY(-1px);
}

.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.2s ease; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }

.tickets-section {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  padding: 1.25rem;
  border-radius: 12px;
  margin-top: 1.5rem;
}

.invite-link-box {
  display: flex; align-items: center; justify-content: space-between;
  background: var(--bg); border: 1px solid var(--border); border-radius: 8px;
  padding: 8px 12px; margin-bottom: 1rem;
}
.link-text {
  font-size: 0.85rem; color: var(--brand); word-break: break-all; padding-right: 10px;
}

.tickets-list {
  display: flex; flex-direction: column; gap: 8px;
}

.ticket-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  gap: 1rem;
  font-size: 0.9rem;
}

.ticket-num {
  font-weight: 700;
  color: var(--text-muted);
  width: 30px;
}

.ticket-code {
  font-family: monospace;
  font-weight: 600;
  flex: 1;
}

.ticket-status {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
}

.status-free {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.status-used {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.btn-copy-small {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-copy-small:hover {
  background: rgba(0, 0, 0, 0.05);
  color: var(--primary-color);
}
</style>
