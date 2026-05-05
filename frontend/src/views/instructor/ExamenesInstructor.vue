<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const examenes = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const showForm = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')

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
  error.value = ''; success.value = ''
  if (!form.value.title) { error.value = 'El titulo es requerido'; return }
  if (!form.value.preguntas.length) { error.value = 'Agrega al menos una pregunta'; return }
  for (const p of form.value.preguntas) {
    if (!p.texto) { error.value = 'Todas las preguntas necesitan texto'; return }
    if (p.valor <= 0) { error.value = 'El valor debe ser mayor a 0'; return }
    if (p.tipo !== 'open_text' && !p.opciones.some(o => o.es_correcta)) {
      error.value = 'Cada pregunta debe tener una respuesta correcta'; return
    }
    if (p.tipo === 'multiple_choice' && p.opciones.some(o => !o.texto)) {
      error.value = 'Todas las opciones deben tener texto'; return
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
    success.value = 'Examen creado exitosamente'
    showForm.value = false
    form.value = { title: '', description: '', capacitacion_id: null, preguntas: [] }
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('Eliminar este examen?')) return
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
      <!-- Alerts -->
      <Transition name="slide-down">
        <div v-if="error" class="alert alert-error" role="alert">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/></svg>
          {{ error }}
        </div>
      </Transition>
      <Transition name="slide-down">
        <div v-if="success" class="alert alert-success" role="status">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 13l4 4L19 7"/></svg>
          {{ success }}
        </div>
      </Transition>

      <!-- New Exam Form -->
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

          <!-- Questions -->
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

              <!-- Open text notice -->
              <div v-if="p.tipo === 'open_text'" class="ex-open-notice">
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                Pregunta de respuesta abierta. No requiere opciones.
              </div>

              <!-- True/False options -->
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

              <!-- Multiple choice options -->
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

      <!-- Examenes List -->
      <TransitionGroup name="list-item" tag="div" class="ex-list">
        <div v-for="ex in examenes" :key="ex.id" class="ex-card">
          <div class="ex-card-left">
            <div class="ex-card-icon" aria-hidden="true">📋</div>
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
          <button @click="eliminar(ex.id)" class="icon-btn danger" :aria-label="`Eliminar examen ${ex.title}`">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/></svg>
          </button>
        </div>
      </TransitionGroup>

      <Transition name="fade">
        <div v-if="examenes.length === 0" class="empty-state">
          <span class="empty-icon">📋</span>
          <p style="font-weight:600;color:var(--dark)">Sin exámenes aún</p>
          <p style="font-size:0.85rem">Crea tu primer examen con el botón de arriba.</p>
        </div>
      </Transition>
    </div>
  </div>
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

/* Exam list cards */
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
}
</style>
