<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const examenes = ref<any[]>([])
const showForm = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')

const form = ref({
  title: '',
  description: '',
  preguntas: [] as Array<{
    texto: string; valor: number; orden: number;
    opciones: Array<{ texto: string; es_correcta: boolean }>
  }>
})

function addPregunta() {
  form.value.preguntas.push({
    texto: '', valor: 1, orden: form.value.preguntas.length,
    opciones: [
      { texto: '', es_correcta: false },
      { texto: '', es_correcta: false },
    ]
  })
}

function removePregunta(i: number) { form.value.preguntas.splice(i, 1) }
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
  const res = await api.get('/admin/examenes')
  examenes.value = res.data || []
}

onMounted(load)

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.title) { error.value = 'El titulo es requerido'; return }
  if (!form.value.preguntas.length) { error.value = 'Agrega al menos una pregunta'; return }
  for (const p of form.value.preguntas) {
    if (!p.texto) { error.value = 'Todas las preguntas necesitan texto'; return }
    if (p.valor <= 0) { error.value = 'El valor de cada pregunta debe ser mayor a 0'; return }
    if (!p.opciones.some(o => o.es_correcta)) { error.value = 'Cada pregunta debe tener una respuesta correcta'; return }
    if (p.opciones.some(o => !o.texto)) { error.value = 'Todas las opciones deben tener texto'; return }
  }
  loading.value = true
  try {
    await api.post('/admin/examenes', form.value)
    success.value = 'Examen creado exitosamente'
    showForm.value = false
    form.value = { title: '', description: '', preguntas: [] }
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('Eliminar este examen?')) return
  await api.delete(`/admin/examenes/${id}`)
  await load()
}
</script>

<template>
  <div>
    <div class="ph">
      <div>
        <h1 class="ph-title">Examenes</h1>
        <p class="ph-sub">Crea y gestiona los examenes de la plataforma.</p>
      </div>
      <button class="btn btn-primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancelar' : '+ Nuevo examen' }}
      </button>
    </div>

    <div v-if="showForm" class="form-card">
      <p class="form-card-title">Nuevo examen</p>
      <div class="form-row">
        <div class="field">
          <label>Titulo *</label>
          <input class="field-input" v-model="form.title" placeholder="Ej: Examen de induccion" />
        </div>
        <div class="field">
          <label>Descripcion</label>
          <input class="field-input" v-model="form.description" placeholder="Opcional..." />
        </div>
      </div>

      <div class="preguntas-list">
        <div v-for="(p, pi) in form.preguntas" :key="pi" class="pregunta-item">
          <div class="pi-header">
            <span class="pi-num">Pregunta {{ pi + 1 }}</span>
            <button class="icon-close" @click="removePregunta(pi)">x</button>
          </div>
          <div class="pi-body">
            <div class="field" style="flex:3">
              <label>Enunciado *</label>
              <textarea class="field-input" v-model="p.texto" rows="2" placeholder="Escribe la pregunta..."></textarea>
            </div>
            <div class="field" style="flex:1">
              <label>Valor (pts)</label>
              <input class="field-input" v-model.number="p.valor" type="number" min="0.5" step="0.5" />
            </div>
          </div>
          <div class="opciones-group">
            <label>Opciones</label>
            <div v-for="(o, oi) in p.opciones" :key="oi" class="opcion-row">
              <input type="radio" :name="'ok-' + pi" :checked="o.es_correcta" @change="setCorrecta(pi, oi)" class="radio-ok" />
              <input class="field-input" v-model="o.texto" placeholder="Texto de la opcion..." style="flex:1" />
              <button class="icon-close sm" @click="removeOpcion(pi, oi)" :disabled="p.opciones.length <= 2">x</button>
            </div>
            <button class="btn-add-op" @click="addOpcion(pi)">+ Opcion</button>
          </div>
        </div>
        <button class="btn-add-preg" @click="addPregunta">+ Agregar pregunta</button>
      </div>

      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div v-if="success" class="alert alert-success">{{ success }}</div>
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="loading" @click="guardar">
          <span v-if="loading" class="spinner" style="width:16px;height:16px"></span>
          {{ loading ? 'Guardando...' : 'Guardar examen' }}
        </button>
        <button class="btn btn-secondary" @click="showForm = false">Cancelar</button>
      </div>
    </div>

    <div class="table-card">
      <table v-if="examenes.length">
        <thead>
          <tr><th>Titulo</th><th>Descripcion</th><th>Fecha</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="e in examenes" :key="e.id">
            <td><strong>{{ e.title }}</strong></td>
            <td class="cell-muted">{{ e.description || '-' }}</td>
            <td class="cell-muted">{{ new Date(e.created_at).toLocaleDateString() }}</td>
            <td><button class="btn btn-danger btn-sm" @click="eliminar(e.id)">Eliminar</button></td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <div class="empty-icon">📝</div>
        <h3>No hay examenes registrados</h3>
        <button class="btn btn-primary" @click="showForm = true">Crear primer examen</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ph { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 12px; margin-bottom: 24px; }
.ph-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.ph-sub { color: var(--muted); font-size: 0.87rem; margin-top: 4px; }
.form-card { background: var(--surface); border-radius: var(--r-lg); padding: 24px; box-shadow: var(--shadow-sm); margin-bottom: 24px; border-top: 4px solid var(--brand); }
.form-card-title { font-size: 1rem; font-weight: 700; color: var(--dark); margin-bottom: 16px; }
.form-row { display: flex; gap: 14px; flex-wrap: wrap; margin-bottom: 16px; }
.field { display: flex; flex-direction: column; gap: 5px; flex: 1; min-width: 200px; }
label { font-size: 0.78rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .04em; }
.form-actions { display: flex; gap: 10px; margin-top: 16px; }
.preguntas-list { display: flex; flex-direction: column; gap: 14px; margin: 16px 0; }
.pregunta-item { border: 1.5px solid var(--border); border-radius: var(--r); padding: 16px; background: var(--bg); }
.pi-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.pi-num { font-weight: 700; color: var(--brand); font-size: 0.88rem; }
.icon-close { background: none; border: none; cursor: pointer; color: var(--subtle); font-size: 0.9rem; padding: 2px 6px; border-radius: 4px; }
.icon-close:hover { color: var(--danger); background: var(--danger-bg); }
.icon-close.sm { font-size: 0.78rem; }
.icon-close:disabled { opacity: 0.3; cursor: not-allowed; }
.pi-body { display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 12px; }
.opciones-group > label { display: block; margin-bottom: 8px; }
.opcion-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.radio-ok { width: 16px; height: 16px; accent-color: var(--brand); cursor: pointer; flex-shrink: 0; }
.btn-add-op { background: none; border: 1.5px dashed var(--border); color: var(--muted); padding: 5px 12px; border-radius: 6px; cursor: pointer; font-size: 0.8rem; margin-top: 4px; }
.btn-add-op:hover { border-color: var(--brand); color: var(--brand); }
.btn-add-preg { background: var(--brand-light); color: var(--brand-dark); border: 2px dashed var(--brand-border); border-radius: var(--r); padding: 10px; cursor: pointer; font-weight: 600; font-size: 0.9rem; width: 100%; }
.btn-add-preg:hover { background: var(--brand-border); }
.table-card { background: var(--surface); border-radius: var(--r-lg); box-shadow: var(--shadow-sm); overflow: hidden; }
table { width: 100%; border-collapse: collapse; }
th { background: var(--bg); padding: 11px 16px; text-align: left; font-size: 0.75rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .05em; border-bottom: 1px solid var(--border); }
td { padding: 12px 16px; border-top: 1px solid var(--border-light); font-size: 0.9rem; }
.cell-muted { color: var(--muted); }
.empty-state { padding: 60px 20px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.empty-icon { font-size: 3rem; }
.empty-state h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
@media (max-width: 600px) { .ph { flex-direction: column; } }
</style>
