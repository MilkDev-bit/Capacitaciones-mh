<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'
import { toast } from '../../utils/toast'

const examenes = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const showForm = ref(false)
const loading = ref(false)

const resModal = ref<any>(null)
const resLista = ref<any[]>([])
const resDetalle = ref<any[] | null>(null)
const resEstudiante = ref<any>(null)
const resLoading = ref(false)
const detalleFiltro = ref<'all' | 'correct' | 'wrong' | 'open'>('all')

const detalleVisible = computed(() => resEstudiante.value !== null)

const detalleFiltrada = computed(() => {
  if (!resDetalle.value) return []
  if (detalleFiltro.value === 'all') return resDetalle.value
  if (detalleFiltro.value === 'correct') return resDetalle.value.filter((p: any) => p.tipo !== 'open_text' && p.es_correcta)
  if (detalleFiltro.value === 'wrong') return resDetalle.value.filter((p: any) => p.tipo !== 'open_text' && !p.es_correcta)
  if (detalleFiltro.value === 'open') return resDetalle.value.filter((p: any) => p.tipo === 'open_text')
  return resDetalle.value
})

const detalleStats = computed(() => {
  if (!resDetalle.value) return { correctas: 0, incorrectas: 0, open: 0, sinRespuesta: 0 }
  return {
    correctas:    resDetalle.value.filter((p: any) => p.tipo !== 'open_text' && p.es_correcta && p.respuesta_dada).length,
    incorrectas:  resDetalle.value.filter((p: any) => p.tipo !== 'open_text' && !p.es_correcta && p.respuesta_dada).length,
    open:         resDetalle.value.filter((p: any) => p.tipo === 'open_text').length,
    sinRespuesta: resDetalle.value.filter((p: any) => !p.respuesta_dada).length,
  }
})

function cerrarDetalle() {
  resDetalle.value = null
  resEstudiante.value = null
  detalleFiltro.value = 'all'
}

async function verResultados(ex: any) {
  resModal.value = ex
  resDetalle.value = null
  resEstudiante.value = null
  detalleFiltro.value = 'all'
  resLoading.value = true
  try {
    const r = await api.get(`/instructor/examenes/${ex.id}/resultados`)
    resLista.value = r.data || []
  } finally {
    resLoading.value = false
  }
}

async function verDetalle(est: any) {
  if (!resModal.value) return
  resEstudiante.value = est
  detalleFiltro.value = 'all'
  resLoading.value = true
  try {
    const r = await api.get(`/instructor/examenes/${resModal.value.id}/resultados/${est.user_id}`)
    resDetalle.value = r.data || []
  } finally {
    resLoading.value = false
  }
}

function scoreColor(pct: number) {
  if (pct >= 80) return '#10b981'
  if (pct >= 60) return '#f59e0b'
  return '#ef4444'
}

const form = ref({
  title: '',
  description: '',
  capacitacion_id: null as string | null,
  preguntas: [] as Array<{
    texto: string
    tipo: string
    valor: number
    orden: number
    opciones: Array<{ texto: string; es_correcta: boolean }>
  }>
})

function addPregunta() {
  form.value.preguntas.push({
    texto: '', tipo: 'multiple_choice', valor: 1, orden: form.value.preguntas.length + 1,
    opciones: [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }]
  })
}

function removePregunta(i: number) { form.value.preguntas.splice(i, 1) }

function onTipoChange(pi: number) {
  const p = form.value.preguntas[pi]
  if (!p) return
  if (p.tipo === 'true_false') {
    p.opciones = [{ texto: 'Verdadero', es_correcta: false }, { texto: 'Falso', es_correcta: false }]
  } else if (p.tipo === 'open_text') {
    p.opciones = []
  } else {
    p.opciones = [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }]
  }
}

function addOpcion(pi: number) {
  const p = form.value.preguntas[pi]
  if (p) p.opciones.push({ texto: '', es_correcta: false })
}
function removeOpcion(pi: number, oi: number) {
  const p = form.value.preguntas[pi]
  if (p && p.opciones.length > 2) p.opciones.splice(oi, 1)
}
function setCorrecta(pi: number, oi: number) {
  const p = form.value.preguntas[pi]
  if (p) p.opciones.forEach((o, idx) => o.es_correcta = idx === oi)
}

async function load() {
  const [eRes, cRes] = await Promise.all([
    api.get('/instructor/examenes'),
    api.get('/instructor/capacitaciones')
  ])
  examenes.value = eRes.data || []
  capacitaciones.value = cRes.data || []
}

onMounted(load)

async function guardar() {
  if (!form.value.title) { toast.error('El titulo es requerido'); return }
  if (!form.value.preguntas.length) { toast.error('Agrega al menos una pregunta'); return }
  for (const p of form.value.preguntas) {
    if (!p.texto) { toast.error('Todas las preguntas necesitan texto'); return }
    if (p.valor <= 0) { toast.error('El valor debe ser mayor a 0'); return }
    if (p.tipo !== 'open_text' && !p.opciones.some(o => o.es_correcta)) {
      toast.error('Cada pregunta debe tener una respuesta correcta'); return
    }
    if (p.tipo === 'multiple_choice' && p.opciones.some(o => !o.texto)) {
      toast.error('Todas las opciones deben tener texto'); return
    }
  }
  loading.value = true
  try {
    const payload: any = {
      title: form.value.title,
      description: form.value.description,
      preguntas: form.value.preguntas.map(p => ({
        texto: p.texto,
        tipo: p.tipo,
        valor: p.valor,
        orden: p.orden,
        opciones: p.tipo !== 'open_text' ? p.opciones : []
      }))
    }
    if (form.value.capacitacion_id) payload.capacitacion_id = form.value.capacitacion_id
    await api.post('/instructor/examenes', payload)
    toast.success('Examen creado exitosamente')
    showForm.value = false
    form.value = { title: '', description: '', capacitacion_id: null, preguntas: [] }
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al guardar')
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!await toast.confirm('Eliminar este examen?')) return
  await api.delete(`/instructor/examenes/${id}`)
  await load()
}
</script>

<template>
  <div class="ex-shell">
    <!-- Page Header -->
    <div class="ph">
      <div>
        <h1 class="ph-title">Mis Exámenes</h1>
        <p class="ph-sub">Crea y administra tus evaluaciones</p>
      </div>
      <div class="ph-actions">
        <button @click="showForm = !showForm" class="btn btn-primary" :aria-expanded="showForm">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24" aria-hidden="true">
            <path v-if="!showForm" d="M12 5v14M5 12h14" stroke-linecap="round"/>
            <path v-else d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/>
          </svg>
          {{ showForm ? 'Cancelar' : 'Nuevo examen' }}
        </button>
      </div>
    </div>

    <div class="ex-body">
      <Transition name="slide-down">
        <div v-if="showForm" class="form-card ex-form-card">
          <h2 class="form-card-title">Crear nuevo examen</h2>
          <div class="form-grid">
            <div class="field full">
              <label class="field-label" for="ex-title">Título del examen <span class="field-req">*</span></label>
              <input id="ex-title" v-model="form.title" placeholder="Ej. Evaluación de seguridad laboral" class="field-input" />
            </div>
            <div class="field full">
              <label class="field-label" for="ex-desc">Descripción <span style="color:var(--muted);font-weight:400">(opcional)</span></label>
              <textarea id="ex-desc" v-model="form.description" placeholder="Instrucciones o contexto del examen..." rows="2" class="field-input" style="resize:vertical" />
            </div>
            <div class="field full">
              <label class="field-label" for="ex-cap">Enlazar a curso <span style="color:var(--muted);font-weight:400">(opcional)</span></label>
              <select id="ex-cap" v-model="form.capacitacion_id" class="field-input">
                <option :value="null">Sin curso enlazado</option>
                <option v-for="cap in capacitaciones" :key="cap.id" :value="cap.id">{{ cap.title }}</option>
              </select>
            </div>
          </div>

          <div class="ex-preguntas-head">
            <span class="ex-preguntas-label">Preguntas <span class="pill-count">{{ form.preguntas.length }}</span></span>
            <button @click="addPregunta" class="btn btn-secondary btn-sm" type="button">
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14" stroke-linecap="round"/></svg>
              Agregar pregunta
            </button>
          </div>

          <TransitionGroup name="list-item" tag="div" class="ex-preguntas-list">
            <div v-for="(p, pi) in form.preguntas" :key="pi" class="ex-pregunta-card">
              <div class="ex-pregunta-header">
                <span class="ex-pregunta-num">{{ pi + 1 }}</span>
                <span class="ex-pregunta-tipo-badge" :class="p.tipo">
                  {{ p.tipo === 'multiple_choice' ? 'Opción múltiple' : p.tipo === 'true_false' ? 'V / F' : 'Abierta' }}
                </span>
                <button @click="removePregunta(pi)" class="icon-btn danger" type="button" :aria-label="`Eliminar pregunta ${pi + 1}`">
                  <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/></svg>
                </button>
              </div>

              <textarea v-model="p.texto" :placeholder="`Texto de la pregunta ${pi + 1}...`" rows="2" class="field-input ex-pregunta-texto" :aria-label="`Pregunta ${pi + 1}`" />

              <div class="ex-pregunta-meta">
                <div class="field">
                  <label class="field-label" :for="`tipo-${pi}`">Tipo</label>
                  <select :id="`tipo-${pi}`" v-model="p.tipo" @change="onTipoChange(pi)" class="field-input">
                    <option value="multiple_choice">Opción múltiple</option>
                    <option value="true_false">Verdadero / Falso</option>
                    <option value="open_text">Respuesta abierta</option>
                  </select>
                </div>
                <div class="field">
                  <label class="field-label" :for="`valor-${pi}`">Puntos</label>
                  <input :id="`valor-${pi}`" type="number" v-model="p.valor" min="0.1" step="0.5" class="field-input" />
                </div>
                <div class="field">
                  <label class="field-label" :for="`orden-${pi}`">Orden</label>
                  <input :id="`orden-${pi}`" type="number" v-model="p.orden" min="1" class="field-input" />
                </div>
              </div>

              <div v-if="p.tipo === 'open_text'" class="ex-open-notice">
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                Pregunta de respuesta abierta. No requiere opciones.
              </div>

              <div v-if="p.tipo === 'true_false'" class="ex-tf-opts">
                <p class="ex-opts-label">¿Cuál es correcta?</p>
                <div style="display:flex;gap:12px">
                  <label class="ex-tf-label">
                    <input type="radio" :name="`tf-${pi}`" @change="setCorrecta(pi, 0)" :checked="p.opciones[0]?.es_correcta" style="accent-color:var(--brand)" />
                    <span :class="{ 'ex-tf-active': p.opciones[0]?.es_correcta }">✓ Verdadero</span>
                  </label>
                  <label class="ex-tf-label">
                    <input type="radio" :name="`tf-${pi}`" @change="setCorrecta(pi, 1)" :checked="p.opciones[1]?.es_correcta" style="accent-color:var(--brand)" />
                    <span :class="{ 'ex-tf-active': p.opciones[1]?.es_correcta }">✗ Falso</span>
                  </label>
                </div>
              </div>

              <div v-if="p.tipo === 'multiple_choice'" class="ex-mc-opts">
                <div class="ex-opts-bar">
                  <p class="ex-opts-label">Opciones — marca la correcta</p>
                  <button @click="addOpcion(pi)" class="btn btn-secondary btn-sm" type="button" style="padding:4px 10px">+ Opción</button>
                </div>
                <div v-for="(op, oi) in p.opciones" :key="oi" class="ex-opcion-row" :class="{ correct: op.es_correcta }">
                  <input v-model="op.texto" :placeholder="`Opción ${oi + 1}`" class="field-input" style="flex:1" />
                  <label class="ex-radio-label" :title="op.es_correcta ? 'Correcta' : 'Marcar como correcta'">
                    <input type="radio" :name="`correcta-${pi}`" @change="setCorrecta(pi, oi)" :checked="op.es_correcta" style="accent-color:var(--success)" />
                    <span>Correcta</span>
                  </label>
                  <button v-if="p.opciones.length > 2" @click="removeOpcion(pi, oi)" class="icon-btn danger" type="button" :aria-label="`Eliminar opción ${oi + 1}`">
                    <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/></svg>
                  </button>
                </div>
              </div>
            </div>
          </TransitionGroup>

          <div v-if="form.preguntas.length === 0" class="empty-state" style="padding:24px;background:var(--bg);border-radius:var(--r);border:1px dashed var(--border)">
            <span class="empty-icon">❓</span>
            <p>Sin preguntas aún. Haz clic en "Agregar pregunta" para comenzar.</p>
          </div>

          <div class="form-actions" style="margin-top:20px">
            <button @click="guardar" :disabled="loading" class="btn btn-primary" :aria-busy="loading">
              <svg v-if="!loading" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7" stroke-linecap="round"/></svg>
              {{ loading ? 'Guardando...' : 'Guardar examen' }}
            </button>
            <button @click="showForm = false" class="btn btn-secondary" type="button">Cancelar</button>
          </div>
        </div>
      </Transition>

      <TransitionGroup name="list-item" tag="div" class="ex-list">
        <div v-for="ex in examenes" :key="ex.id" class="ex-card">
          <div class="ex-card-left">
            <div class="ex-card-icon" aria-hidden="true"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg></div>
            <div class="ex-card-info">
              <h2 class="ex-card-title">{{ ex.title }}</h2>
              <p v-if="ex.description" class="ex-card-desc">{{ ex.description }}</p>
              <div class="ex-card-meta">
                <span v-if="ex.capacitacion_nombre" class="ex-linked-badge">
                  <svg width="11" height="11" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg>
                  {{ ex.capacitacion_nombre }}
                </span>
                <span v-else class="ex-nolink-badge">Sin curso enlazado</span>
              </div>
            </div>
          </div>
          <div class="ex-card-actions">
            <button @click.stop="verResultados(ex)" class="btn btn-secondary btn-sm" title="Ver respuestas de estudiantes">
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/></svg>
              Resultados
            </button>
            <button @click="eliminar(ex.id)" class="icon-btn danger" :aria-label="`Eliminar examen ${ex.title}`">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/></svg>
            </button>
          </div>
        </div>
      </TransitionGroup>

      <Transition name="fade">
        <div v-if="examenes.length === 0" class="empty-state">
          <span class="empty-icon"><svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg></span>
          <p style="font-weight:600;color:var(--dark)">Sin exámenes aún</p>
          <p style="font-size:0.85rem">Crea tu primer examen con el botón de arriba.</p>
        </div>
      </Transition>
    </div>
  </div>

  <Teleport to="body">
    <Transition name="fade">
      <div v-if="resModal && !detalleVisible" class="res-overlay" @click.self="resModal = null">
        <div class="res-panel">
          <div class="res-header">
            <div>
              <p class="res-header-label">Resultados del examen</p>
              <h2 class="res-header-title">{{ resModal.title }}</h2>
            </div>
            <button class="icon-btn" @click="resModal = null" title="Cerrar">
              <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/></svg>
            </button>
          </div>

          <div v-if="resLoading" class="res-loading">
            <div class="spinner"></div>
          </div>

          <template v-else-if="!resDetalle">
            <div v-if="resLista.length === 0" class="res-empty">
              <svg width="36" height="36" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
              <p>Ningún estudiante ha respondido aún.</p>
            </div>
            <div v-else class="res-table-wrap">
              <p class="res-count">{{ resLista.length }} estudiante{{ resLista.length !== 1 ? 's' : '' }} respondieron</p>
              <table class="res-table">
                <thead>
                  <tr>
                    <th>Estudiante</th>
                    <th>Puntaje</th>
                    <th>%</th>
                    <th>Fecha</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="est in resLista" :key="est.user_id">
                    <td>
                      <div class="res-student-name">{{ est.nombre }}</div>
                      <div class="res-student-email">{{ est.email }}</div>
                    </td>
                    <td class="res-score">{{ est.puntaje.toFixed(1) }} / {{ est.puntaje_max.toFixed(1) }}</td>
                    <td>
                      <span class="res-pct-badge" :style="{ background: scoreColor(est.porcentaje) + '20', color: scoreColor(est.porcentaje) }">
                        {{ est.porcentaje.toFixed(0) }}%
                      </span>
                    </td>
                    <td class="res-date">{{ new Date(est.respondido_at).toLocaleDateString('es-MX', { day:'2-digit', month:'short', year:'numeric', hour:'2-digit', minute:'2-digit' }) }}</td>
                    <td>
                      <button class="btn btn-secondary btn-sm" @click="verDetalle(est)">Ver detalle</button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </template>
        </div>
      </div>
    </Transition>
  </Teleport>

  <Teleport to="body">
    <div v-if="detalleVisible" class="det-overlay">
      <div class="det-header">
        <div class="det-header-left">
          <button class="det-back" @click="cerrarDetalle">
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M19 12H5M12 5l-7 7 7 7" stroke-linecap="round" stroke-linejoin="round"/></svg>
            Volver a resultados
          </button>
          <div class="det-title-group">
            <span class="det-exam-name">{{ resModal?.title }}</span>
            <span class="det-sep">›</span>
            <span class="det-student-name">{{ resEstudiante?.nombre }}</span>
          </div>
        </div>
        <div class="det-header-right">
          <span class="det-score-pill" :style="{ background: scoreColor(resEstudiante?.porcentaje) + '18', color: scoreColor(resEstudiante?.porcentaje) }">
            {{ resEstudiante?.porcentaje?.toFixed(0) }}%
          </span>
          <span class="det-pts">{{ resEstudiante?.puntaje?.toFixed(1) }} / {{ resEstudiante?.puntaje_max?.toFixed(1) }} pts</span>
        </div>
      </div>

      <div class="det-stats-bar">
        <button
          v-for="f in [
            { key: 'all',     label: 'Todas',       count: resDetalle?.length ?? 0,   color: '' },
            { key: 'correct', label: 'Correctas',   count: detalleStats.correctas,    color: '#10b981' },
            { key: 'wrong',   label: 'Incorrectas', count: detalleStats.incorrectas,  color: '#ef4444' },
            { key: 'open',    label: 'Texto libre', count: detalleStats.open,         color: '#6366f1' },
          ]"
          :key="f.key"
          class="det-stat-chip"
          :class="{ active: detalleFiltro === f.key }"
          :style="detalleFiltro === f.key && f.color ? { background: f.color + '18', color: f.color, borderColor: f.color + '50' } : {}"
          @click="detalleFiltro = f.key as any"
        >
          <span class="det-chip-count" :style="f.color && detalleFiltro !== f.key ? { color: f.color } : {}">{{ f.count }}</span>
          {{ f.label }}
        </button>
      </div>

      <div class="det-body">
        <div v-if="resLoading" class="det-loading"><div class="spinner"></div></div>
        <template v-else>
          <div
            v-for="p in detalleFiltrada"
            :key="p.pregunta_id"
            class="det-q-card"
            :class="!p.respuesta_dada ? 'q-unanswered' : p.tipo === 'open_text' ? 'q-open' : p.es_correcta ? 'q-correct' : 'q-wrong'"
          >
            <div class="det-q-head">
              <span class="det-q-num">{{ (resDetalle?.indexOf(p) ?? 0) + 1 }}</span>
              <p class="det-q-texto">{{ p.texto }}</p>
              <div class="det-q-right">
                <span v-if="!p.respuesta_dada" class="det-q-tag tag-unanswered">Sin respuesta</span>
                <span v-else-if="p.tipo === 'open_text'" class="det-q-tag tag-open">Texto libre</span>
                <span v-else-if="p.es_correcta" class="det-q-tag tag-correct">
                  <svg width="11" height="11" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7" stroke-linecap="round"/></svg>
                  Correcto
                </span>
                <span v-else class="det-q-tag tag-wrong">
                  <svg width="11" height="11" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/></svg>
                  Incorrecto
                </span>
                <span class="det-q-val">{{ p.valor }} pt</span>
              </div>
            </div>
            <div v-if="p.respuesta_dada" class="det-q-answers">
              <div class="det-ans-row">
                <span class="det-ans-label">Respondió</span>
                <span class="det-ans-val" :class="p.tipo === 'open_text' ? 'val-open' : p.es_correcta ? 'val-correct' : 'val-wrong'">{{ p.respuesta_dada }}</span>
              </div>
              <div v-if="p.tipo !== 'open_text' && !p.es_correcta && p.respuesta_correcta" class="det-ans-row">
                <span class="det-ans-label">Correcta</span>
                <span class="det-ans-val val-correct">{{ p.respuesta_correcta }}</span>
              </div>
            </div>
          </div>
          <p v-if="detalleFiltrada.length === 0" class="det-empty">No hay preguntas en esta categoría.</p>
        </template>
      </div>

    </div>
  </Teleport>
</template>

<style scoped>
.ex-shell { min-height: 100vh; background: var(--bg); }
.ex-body { max-width: 860px; margin: 0 auto; padding: 0 28px 40px; }

.ex-form-card { margin-bottom: 24px; }

.ex-preguntas-head {
  display: flex; align-items: center; justify-content: space-between;
  margin: 20px 0 12px; padding-top: 16px; border-top: 1.5px solid var(--border-light);
}
.ex-preguntas-label { font-size: 0.88rem; font-weight: 700; color: var(--dark); display: flex; align-items: center; gap: 6px; }
.ex-preguntas-list { display: flex; flex-direction: column; gap: 14px; position: relative; }

.ex-pregunta-card {
  background: var(--surface); border-radius: var(--r-lg); border: 1px solid var(--border-light);
  padding: 20px; display: flex; flex-direction: column; gap: 12px;
  box-shadow: var(--shadow-sm); transition: box-shadow 0.15s, border-color 0.15s;
}
.ex-pregunta-card:focus-within { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(249,115,22,.12); }

.ex-pregunta-header {
  display: flex; align-items: center; gap: 10px;
}
.ex-pregunta-num {
  width: 24px; height: 24px; border-radius: 50%; background: var(--brand); color: #fff;
  font-size: 0.75rem; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.ex-pregunta-tipo-badge {
  font-size: 0.72rem; font-weight: 700; padding: 2px 9px; border-radius: 10px;
  background: var(--border-light); color: var(--muted);
}
.ex-pregunta-tipo-badge.multiple_choice { background: #dbeafe; color: #1d4ed8; }
.ex-pregunta-tipo-badge.true_false      { background: #fef9c3; color: #a16207; }
.ex-pregunta-tipo-badge.open_text       { background: #d1fae5; color: #065f46; }

.ex-pregunta-texto { resize: vertical; }

.ex-pregunta-meta {
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px;
}

.ex-open-notice {
  display: flex; align-items: center; gap: 6px; font-size: 0.8rem;
  color: var(--muted); background: var(--surface); padding: 8px 10px; border-radius: var(--r-sm);
  border: 1px solid var(--border-light);
}

.ex-tf-opts { padding: 6px 0; }
.ex-opts-label { font-size: 0.8rem; font-weight: 600; color: var(--muted); margin-bottom: 7px; }
.ex-tf-label { display: flex; align-items: center; gap: 6px; cursor: pointer; font-size: 0.88rem; color: var(--text); }
.ex-tf-active { font-weight: 700; color: var(--brand); }

.ex-mc-opts { display: flex; flex-direction: column; gap: 6px; }
.ex-opts-bar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
.ex-opcion-row {
  display: flex; align-items: center; gap: 8px; padding: 5px 8px; border-radius: var(--r-sm);
  background: var(--surface); border: 1.5px solid var(--border-light); transition: border-color 0.12s;
}
.ex-opcion-row.correct { border-color: var(--success); background: var(--success-bg); }
.ex-radio-label { display: flex; align-items: center; gap: 5px; font-size: 0.8rem; color: var(--muted); cursor: pointer; white-space: nowrap; }

.ex-list { display: flex; flex-direction: column; gap: 10px; position: relative; }
.ex-card {
  background: var(--surface); border-radius: var(--r-lg); border: 1px solid var(--border-light);
  padding: 20px; display: flex; align-items: flex-start; gap: 16px;
  box-shadow: var(--shadow-sm); transition: box-shadow 0.2s, transform 0.2s;
}
.ex-card:hover { box-shadow: var(--shadow-lg); transform: translateY(-3px); }
.ex-card-left { display: flex; align-items: flex-start; gap: 12px; flex: 1; min-width: 0; }
.ex-card-icon { font-size: 1.6rem; flex-shrink: 0; line-height: 1; margin-top: 2px; }
.ex-card-info { flex: 1; min-width: 0; }
.ex-card-title { font-size: 1rem; font-weight: 700; color: var(--dark); margin-bottom: 3px; }
.ex-card-desc { font-size: 0.85rem; color: var(--muted); margin-bottom: 7px; }
.ex-card-meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.ex-linked-badge {
  display: inline-flex; align-items: center; gap: 4px; font-size: 0.76rem; font-weight: 600;
  background: var(--success-bg); color: var(--success); padding: 3px 10px; border-radius: 10px;
}
.ex-nolink-badge {
  font-size: 0.76rem; font-weight: 500; color: var(--muted); background: var(--bg);
  padding: 3px 10px; border-radius: 10px; border: 1px solid var(--border-light);
}

@media (max-width: 768px) {
  .ex-body { padding: 0 14px 32px; }
  .ex-pregunta-meta { grid-template-columns: 1fr 1fr; }
  .det-overlay { inset: 0; }
  .res-overlay  { inset: 0; }
}

.ex-card-actions { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }

.res-overlay {
  position: fixed; inset: 0 0 0 var(--sidebar-w); background: rgba(15,23,42,.45);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}
.res-panel {
  background: var(--surface); border-radius: var(--r-xl);
  box-shadow: 0 20px 60px rgba(0,0,0,.25);
  width: 100%; max-width: 780px; max-height: 88vh;
  display: flex; flex-direction: column; overflow: hidden;
}

.res-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  padding: 22px 24px 18px; border-bottom: 1px solid var(--border-light); flex-shrink: 0;
}
.res-header-label { font-size: 0.74rem; font-weight: 700; text-transform: uppercase; letter-spacing: .06em; color: var(--brand); margin-bottom: 2px; }
.res-header-title { font-size: 1.15rem; font-weight: 800; color: var(--dark); margin: 0; }

.res-loading { display: flex; justify-content: center; padding: 48px; }
.res-empty { display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 48px; color: var(--muted); font-size: 0.9rem; }

.res-table-wrap { overflow-y: auto; padding: 20px 24px; flex: 1; }
.res-count { font-size: 0.82rem; font-weight: 600; color: var(--muted); margin-bottom: 14px; }
.res-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; }
.res-table th { text-align: left; padding: 8px 12px; font-size: 0.75rem; font-weight: 700; text-transform: uppercase; letter-spacing: .05em; color: var(--muted); border-bottom: 2px solid var(--border-light); }
.res-table td { padding: 12px 12px; border-bottom: 1px solid var(--border-light); vertical-align: middle; }
.res-table tbody tr:hover { background: var(--bg); }
.res-student-name { font-weight: 600; color: var(--dark); }
.res-student-email { font-size: 0.78rem; color: var(--muted); }
.res-score { font-weight: 600; color: var(--dark); white-space: nowrap; }
.res-pct-badge { font-size: 0.78rem; font-weight: 700; padding: 3px 9px; border-radius: 10px; }
.res-date { font-size: 0.78rem; color: var(--muted); white-space: nowrap; }
.res-back-btn {
  display: inline-flex; align-items: center; gap: 6px; margin: 16px 24px 0;
  font-size: 0.83rem; font-weight: 600; color: var(--muted);
  background: none; border: none; cursor: pointer; transition: color .15s;
}
.res-back-btn:hover { color: var(--brand); }
.res-detalle-header { display: flex; align-items: baseline; justify-content: space-between; padding: 10px 24px 14px; border-bottom: 1px solid var(--border-light); }
.res-detalle-user { font-size: 1rem; font-weight: 700; color: var(--dark); }
.res-detalle-meta { font-size: 0.88rem; color: var(--muted); }
.res-preguntas { overflow-y: auto; flex: 1; padding: 16px 24px 24px; display: flex; flex-direction: column; gap: 12px; }
.res-pregunta {
  border-radius: var(--r); padding: 14px 16px;
  border-left: 4px solid var(--border); background: var(--bg);
}
.res-pregunta.correct  { border-left-color: #10b981; background: #f0fdf4; }
.res-pregunta.wrong    { border-left-color: #ef4444; background: #fef2f2; }
.res-pregunta.open     { border-left-color: #6366f1; background: #eef2ff; }
.res-pregunta.unanswered { border-left-color: var(--border); opacity: .7; }
.res-pregunta-head { display: flex; align-items: flex-start; gap: 10px; margin-bottom: 8px; }
.res-qnum { font-size: 0.72rem; font-weight: 800; background: var(--brand); color: #fff; padding: 2px 7px; border-radius: 10px; flex-shrink: 0; margin-top: 2px; }
.res-qtexto { flex: 1; font-weight: 600; color: var(--dark); font-size: 0.9rem; line-height: 1.5; }
.res-qval { font-size: 0.75rem; font-weight: 700; color: var(--success); white-space: nowrap; flex-shrink: 0; }
.res-pregunta-body { font-size: 0.87rem; display: flex; flex-direction: column; gap: 4px; padding-left: 4px; }
.res-resp-row { display: flex; align-items: center; gap: 8px; }
.res-resp-label { font-size: 0.76rem; font-weight: 700; color: var(--muted); min-width: 72px; }
.res-resp-val { color: var(--dark); flex: 1; }
.res-resp-icon { flex-shrink: 0; display: flex; }
.res-correcta-row { display: flex; align-items: center; gap: 8px; }
.correct-text { color: #10b981; font-weight: 600; }
.res-sin-respuesta { font-size: 0.82rem; color: var(--muted); font-style: italic; }

.det-overlay {
  position: fixed; inset: 0 0 0 var(--sidebar-w); z-index: 1100;
  background: var(--bg);
  display: flex; flex-direction: column;
  overflow: hidden;
}

.slide-up-enter-active { transition: transform .3s cubic-bezier(.22,1,.36,1), opacity .25s; }
.slide-up-leave-active { transition: transform .2s ease-in, opacity .15s; }
.slide-up-enter-from  { transform: translateY(30px); opacity: 0; }
.slide-up-leave-to    { transform: translateY(30px); opacity: 0; }

.det-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 28px; background: var(--surface); border-bottom: 1px solid var(--border-light);
  flex-shrink: 0; gap: 16px; flex-wrap: wrap;
}
.det-header-left { display: flex; flex-direction: column; gap: 4px; }
.det-back {
  display: inline-flex; align-items: center; gap: 6px;
  font-size: 0.8rem; font-weight: 600; color: var(--muted);
  background: none; border: none; cursor: pointer; padding: 0; transition: color .15s;
}
.det-back:hover { color: var(--brand); }
.det-title-group { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.det-exam-name { font-size: 0.88rem; color: var(--muted); }
.det-sep { color: var(--border); font-size: 1rem; }
.det-student-name { font-size: 0.95rem; font-weight: 800; color: var(--dark); }
.det-header-right { display: flex; align-items: center; gap: 10px; flex-shrink: 0; }
.det-score-pill {
  font-size: 0.9rem; font-weight: 800;
  padding: 5px 14px; border-radius: 20px;
}
.det-pts { font-size: 0.85rem; font-weight: 600; color: var(--muted); }

.det-stats-bar {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 28px; background: var(--surface); border-bottom: 1px solid var(--border-light);
  flex-shrink: 0; flex-wrap: wrap;
}
.det-stat-chip {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px; border-radius: 20px;
  font-size: 0.82rem; font-weight: 600; cursor: pointer;
  background: var(--bg); border: 1.5px solid var(--border-light); color: var(--muted);
  transition: all .15s;
}
.det-stat-chip:hover { border-color: var(--border); color: var(--dark); }
.det-stat-chip.active { border-color: var(--brand); color: var(--brand); background: rgba(249,115,22,.08); }
.det-chip-count { font-size: 0.78rem; font-weight: 800; }

.det-body {
  flex: 1; overflow-y: auto; padding: 28px;
  display: flex; flex-direction: column; gap: 16px;
  max-width: 860px; width: 100%; margin: 0 auto;
}
.det-loading { display: flex; justify-content: center; padding: 60px; }
.det-empty { text-align: center; color: var(--muted); font-size: 0.9rem; padding: 40px; }

.det-q-card {
  background: var(--surface); border-radius: var(--r-lg);
  border: 1.5px solid var(--border-light); overflow: hidden;
  box-shadow: var(--shadow-sm);
  transition: box-shadow .15s;
}
.det-q-card:hover { box-shadow: var(--shadow-md); }
.det-q-card.q-correct { border-color: #bbf7d0; }
.det-q-card.q-wrong   { border-color: #fecaca; }
.det-q-card.q-open    { border-color: #c7d2fe; }
.det-q-card.q-unanswered { opacity: .7; }

.det-q-head {
  display: flex; align-items: flex-start; gap: 14px;
  padding: 18px 20px 14px;
  border-bottom: 1px solid var(--border-light);
}
.det-q-num {
  flex-shrink: 0; min-width: 28px; height: 28px; border-radius: 50%;
  background: var(--brand); color: #fff;
  font-size: 0.78rem; font-weight: 800;
  display: flex; align-items: center; justify-content: center;
  margin-top: 1px;
}
.det-q-texto { flex: 1; font-size: 0.95rem; font-weight: 600; color: var(--dark); line-height: 1.6; margin: 0; }
.det-q-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.det-q-tag {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 0.72rem; font-weight: 700; padding: 3px 9px; border-radius: 10px;
}
.tag-correct    { background: #dcfce7; color: #15803d; }
.tag-wrong      { background: #fee2e2; color: #b91c1c; }
.tag-open       { background: #e0e7ff; color: #4338ca; }
.tag-unanswered { background: var(--bg); color: var(--muted); border: 1px solid var(--border-light); }
.det-q-val { font-size: 0.78rem; font-weight: 700; color: var(--success); white-space: nowrap; }

.det-q-answers {
  padding: 14px 20px 16px 62px;
  display: flex; flex-direction: column; gap: 8px;
}
.det-ans-row { display: flex; align-items: flex-start; gap: 12px; }
.det-ans-label {
  font-size: 0.75rem; font-weight: 700; color: var(--muted); text-transform: uppercase;
  letter-spacing: .04em; min-width: 64px; padding-top: 2px;
}
.det-ans-val {
  flex: 1; font-size: 0.9rem; color: var(--dark);
  line-height: 1.5;
}
.val-correct { color: #15803d; font-weight: 600; }
.val-wrong   { color: #b91c1c; font-weight: 600; }
.val-open    { color: var(--dark); font-style: italic; }
</style>
