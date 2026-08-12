<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
const invoiceLoading = ref(false)

function openDetails(lic: any) {
  selectedLic.value = lic
}

function closeModal() {
  selectedLic.value = null
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

async function copyCode(codigo: string) {
  navigator.clipboard.writeText(codigo)
  toast.success('Código copiado al portapapeles')
}

// ── Reparto de accesos por correo ──────────────────────────────────────────
// El comprador captura los correos de su equipo y cada persona recibe su
// acceso individual, en lugar de tener que copiar códigos a mano desde aquí.

interface FilaParticipante { nombre: string; email: string }

const envioLic = ref<any>(null)
const participantes = ref<FilaParticipante[]>([{ nombre: '', email: '' }])
const enviando = ref(false)
const invitaciones = ref<any[]>([])
const loadingInvitaciones = ref(false)
const pegadoMasivo = ref('')

const lugaresLicencia = computed(() =>
  envioLic.value?.capacidad_maxima > 0 ? envioLic.value.capacidad_maxima : Infinity
)
const yaEnviados = computed(() => invitaciones.value.length)
const disponibles = computed(() =>
  lugaresLicencia.value === Infinity ? Infinity : Math.max(0, lugaresLicencia.value - yaEnviados.value)
)
const correosValidos = computed(() =>
  participantes.value.filter((p) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(p.email.trim()))
)
const excedeCupo = computed(() => correosValidos.value.length > disponibles.value)

async function abrirEnvio(lic: any) {
  envioLic.value = lic
  participantes.value = [{ nombre: '', email: '' }]
  pegadoMasivo.value = ''
  await cargarInvitaciones(lic.id)
}

async function cargarInvitaciones(licenciaId: string) {
  loadingInvitaciones.value = true
  try {
    const res = await api.get(`/licencias/${licenciaId}/invitaciones`)
    invitaciones.value = res.data || []
  } catch {
    invitaciones.value = []
  } finally {
    loadingInvitaciones.value = false
  }
}

function cerrarEnvio() {
  envioLic.value = null
  participantes.value = [{ nombre: '', email: '' }]
  invitaciones.value = []
}

function agregarFila() {
  participantes.value.push({ nombre: '', email: '' })
}

function quitarFila(i: number) {
  participantes.value.splice(i, 1)
  if (participantes.value.length === 0) agregarFila()
}

/** Convierte una lista pegada (uno por línea, o "Nombre <correo>") en filas. */
function procesarPegado() {
  const lineas = pegadoMasivo.value.split(/[\n,;]+/).map((l) => l.trim()).filter(Boolean)
  if (!lineas.length) return

  const nuevas: FilaParticipante[] = []
  for (const linea of lineas) {
    const match = linea.match(/^(.*?)[<\s]*([^\s<>]+@[^\s<>]+)>?$/)
    if (!match?.[2]) continue
    nuevas.push({ nombre: (match[1] ?? '').replace(/["']/g, '').trim(), email: match[2].trim() })
  }
  if (!nuevas.length) {
    toast.error('No encontramos correos válidos en el texto pegado')
    return
  }
  // Se conservan las filas que el usuario ya había llenado a mano.
  const existentes = participantes.value.filter((p) => p.email.trim())
  participantes.value = [...existentes, ...nuevas]
  pegadoMasivo.value = ''
  toast.success(`${nuevas.length} participante(s) agregado(s)`)
}

async function enviarAccesos() {
  if (!correosValidos.value.length) {
    toast.error('Captura al menos un correo válido')
    return
  }
  if (excedeCupo.value) {
    toast.error(`Solo quedan ${disponibles.value} accesos disponibles en esta licencia`)
    return
  }

  enviando.value = true
  try {
    const res = await api.post(`/licencias/${envioLic.value.id}/enviar-accesos`, {
      participantes: correosValidos.value.map((p) => ({ nombre: p.nombre.trim(), email: p.email.trim() })),
    })
    toast.success(`Enviamos ${res.data.enviados} acceso(s) por correo`)
    participantes.value = [{ nombre: '', email: '' }]
    await cargarInvitaciones(envioLic.value.id)
    await fetchLicencias()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos enviar los accesos')
  } finally {
    enviando.value = false
  }
}

function tramitarDC3Licencia(lic: any) {
  const nombreCurso = lic.nombre && lic.nombre !== 'Licencia Corporativa' ? lic.nombre : 'Capacitación'
  const duracion = Math.ceil((lic.curso_duracion || 60) / 60)
  const url = `https://dc3.mhsolucionesempresariales.com/formulario-dc3-8f9d3a2b?nombre_curso=${encodeURIComponent(nombreCurso)}&duracion_horas=${duracion}&area_tematica=6000`
  window.open(url, '_blank')
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

          <div class="code-section">
            <p class="code-instruction">Envía este código a tu equipo para que puedan acceder al curso:</p>
            <div class="code-box" @click="copyCode(lic.codigo_acceso)">
              {{ lic.codigo_acceso }}
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="copy-icon"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
            </div>
          </div>

          <!-- Acción principal: repartir los accesos por correo -->
          <button class="btn-enviar" @click="abrirEnvio(lic)">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="16" x="2" y="4" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
            Enviar accesos por correo
          </button>
          <p v-if="lic.accesos_enviados > 0" class="enviados-hint">
            {{ lic.accesos_enviados }} acceso(s) ya enviados
          </p>

          <div style="display: flex; gap: 8px; margin-top: 12px;">
            <button class="btn-details" style="margin-top: 0; flex: 1;" @click="openDetails(lic)">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
              Ver Detalles
            </button>
            <button class="btn-details" style="margin-top: 0; flex: 1; background: rgba(16, 185, 129, 0.08); color: #10b981; border-color: rgba(16, 185, 129, 0.3);" @click="tramitarDC3Licencia(lic)">
              📋 Tramitar DC-3
            </button>
          </div>
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

            </div>
            <div class="modal-footer">
              <button style="display: flex; align-items: center; justify-content: center; gap: 8px; background-color: #10b981; color: #ffffff; border: none; border-radius: 10px; padding: 10px 14px; font-size: 0.85rem; font-weight: 600; cursor: pointer; flex: 1;" @click="tramitarDC3Licencia(selectedLic)">
                📋 Tramitar DC-3
              </button>
              <button style="display: flex; align-items: center; justify-content: center; gap: 8px; background-color: #f97316; color: #ffffff; border: none; border-radius: 10px; padding: 10px 14px; font-size: 0.85rem; font-weight: 600; cursor: pointer; flex: 1;" class="action-doc-btn" @click="downloadInvoice(selectedLic)" :disabled="invoiceLoading">
                <svg v-if="!invoiceLoading" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                <span v-if="invoiceLoading">Buscando...</span>
                <span v-else>Recibo</span>
              </button>
              <button class="btn-secondary" style="padding: 10px 14px; font-size: 0.85rem;" @click="copyCode(selectedLic.codigo_acceso)">Copiar Código</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Modal: repartir accesos por correo -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="envioLic" class="modal-overlay" @click.self="cerrarEnvio">
          <div class="modal-card modal-wide">
            <div class="modal-header">
              <div>
                <h3>Enviar accesos por correo</h3>
                <p class="modal-sub">{{ envioLic.capacitacion_titulo || envioLic.nombre }}</p>
              </div>
              <button class="modal-close" @click="cerrarEnvio">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>

            <div class="modal-body">
              <div class="cupo-bar">
                <span><strong>{{ disponibles === Infinity ? '∞' : disponibles }}</strong> accesos disponibles</span>
                <span class="cupo-sep">·</span>
                <span>{{ yaEnviados }} enviados</span>
              </div>

              <!-- Pegado masivo: lo normal es que RR. HH. tenga la lista en un correo -->
              <details class="bulk">
                <summary>Pegar una lista de correos</summary>
                <textarea
                  v-model="pegadoMasivo"
                  class="bulk-input"
                  rows="4"
                  placeholder="ana@empresa.com&#10;Luis Pérez <luis@empresa.com>&#10;maria@empresa.com"
                ></textarea>
                <button class="btn-secondary btn-sm" :disabled="!pegadoMasivo.trim()" @click="procesarPegado">
                  Agregar a la lista
                </button>
              </details>

              <div class="participantes">
                <div v-for="(p, i) in participantes" :key="i" class="participante-row">
                  <input v-model="p.nombre" class="input-nombre" type="text" placeholder="Nombre (opcional)" />
                  <input v-model="p.email" class="input-email" type="email" placeholder="correo@empresa.com" />
                  <button class="btn-quitar" :aria-label="`Quitar participante ${i + 1}`" @click="quitarFila(i)">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </div>
              </div>

              <button class="btn-agregar" @click="agregarFila">+ Agregar participante</button>

              <p v-if="excedeCupo" class="alerta">
                Capturaste {{ correosValidos.length }} correos pero solo quedan {{ disponibles }} accesos.
              </p>

              <div v-if="invitaciones.length" class="historial">
                <h4>Accesos ya enviados</h4>
                <div v-if="loadingInvitaciones" class="loading" style="font-size: 0.85rem;">Cargando…</div>
                <div v-else class="historial-list">
                  <div v-for="inv in invitaciones" :key="inv.id" class="historial-item">
                    <div class="historial-info">
                      <span class="historial-email">{{ inv.email }}</span>
                      <span v-if="inv.nombre" class="historial-nombre">{{ inv.nombre }}</span>
                    </div>
                    <span class="mono historial-codigo">{{ inv.codigo }}</span>
                    <span :class="['ticket-status', inv.estado === 'usado' ? 'status-used' : 'status-free']">
                      {{ inv.estado === 'usado' ? 'Activado' : 'Enviado' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div class="modal-footer">
              <button class="btn-secondary" style="padding: 10px 14px; font-size: 0.85rem;" @click="cerrarEnvio">Cerrar</button>
              <button
                class="btn-enviar"
                style="margin: 0; flex: 1;"
                :disabled="enviando || !correosValidos.length || excedeCupo"
                @click="enviarAccesos"
              >
                {{ enviando ? 'Enviando…' : `Enviar ${correosValidos.length || ''} acceso(s)` }}
              </button>
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



/* ── Reparto de accesos por correo ─────────────────────────────────────── */

.btn-enviar {
  width: 100%;
  margin-top: 16px;
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 12px 14px; border: none; border-radius: 10px;
  background: linear-gradient(135deg, #f97316 0%, #ea580c 100%);
  color: #fff; font-size: 0.9rem; font-weight: 700; cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s, opacity 0.2s;
  box-shadow: 0 6px 16px rgba(249, 115, 22, 0.22);
}
.btn-enviar:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 10px 22px rgba(249, 115, 22, 0.3); }
.btn-enviar:disabled { opacity: 0.5; cursor: not-allowed; box-shadow: none; transform: none; }

.enviados-hint { margin: 8px 0 0; font-size: 0.78rem; color: var(--muted); text-align: center; }

.modal-wide { max-width: 640px; }
.modal-sub { margin: 4px 0 0; font-size: 0.85rem; color: var(--muted); }

.cupo-bar {
  display: flex; align-items: center; gap: 8px; flex-wrap: wrap;
  background: rgba(249, 115, 22, 0.07); border-radius: 10px;
  padding: 10px 14px; font-size: 0.85rem; color: var(--text-muted, #6b7280);
  margin-bottom: 16px;
}
.cupo-bar strong { color: #ea580c; font-size: 1rem; }
.cupo-sep { opacity: 0.4; }

.bulk { margin-bottom: 16px; }
.bulk summary {
  cursor: pointer; font-size: 0.85rem; font-weight: 600; color: #f97316;
  padding: 6px 0; user-select: none;
}
.bulk-input {
  width: 100%; margin: 10px 0 8px; padding: 10px 12px; font-size: 0.85rem;
  border: 1px solid rgba(0, 0, 0, 0.1); border-radius: 10px; resize: vertical;
  font-family: inherit; box-sizing: border-box;
}
.bulk-input:focus { outline: none; border-color: #f97316; }

.participantes { display: flex; flex-direction: column; gap: 8px; }
.participante-row { display: flex; gap: 8px; align-items: center; }
.input-nombre, .input-email {
  padding: 10px 12px; font-size: 0.87rem; border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 9px; font-family: inherit; min-width: 0;
}
.input-nombre { flex: 0 1 38%; }
.input-email { flex: 1 1 auto; }
.input-nombre:focus, .input-email:focus { outline: none; border-color: #f97316; }

.btn-quitar {
  flex-shrink: 0; width: 32px; height: 32px; border: none; border-radius: 8px;
  background: rgba(239, 68, 68, 0.08); color: #ef4444; cursor: pointer;
  display: flex; align-items: center; justify-content: center; transition: background 0.15s;
}
.btn-quitar:hover { background: rgba(239, 68, 68, 0.18); }

.btn-agregar {
  margin-top: 10px; background: none; border: 1px dashed rgba(0, 0, 0, 0.15);
  border-radius: 9px; padding: 9px 14px; width: 100%;
  font-size: 0.85rem; font-weight: 600; color: var(--muted); cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}
.btn-agregar:hover { border-color: #f97316; color: #f97316; }

.alerta {
  margin: 14px 0 0; padding: 10px 14px; border-radius: 9px;
  background: rgba(239, 68, 68, 0.08); color: #b91c1c; font-size: 0.83rem;
}

.historial { margin-top: 24px; }
.historial h4 { margin: 0 0 10px; font-size: 0.9rem; color: var(--text-color); }
.historial-list { display: flex; flex-direction: column; gap: 6px; max-height: 220px; overflow-y: auto; }
.historial-item {
  display: flex; align-items: center; gap: 10px;
  background: rgba(0, 0, 0, 0.02); border-radius: 9px; padding: 9px 12px;
}
.historial-info { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.historial-email { font-size: 0.85rem; color: var(--text-color); overflow: hidden; text-overflow: ellipsis; }
.historial-nombre { font-size: 0.75rem; color: var(--muted); }
.historial-codigo { font-size: 0.8rem; color: var(--muted); }

@media (max-width: 560px) {
  .participante-row { flex-wrap: wrap; }
  .input-nombre { flex: 1 1 100%; }
}
</style>
