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
  if (!form.value.title) { error.value = 'El título es requerido'; return }
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
  if (!confirm('¿Eliminar este examen?')) return
  await api.delete(`/admin/examenes/${id}`)
  await load()
}

function puntajeTotal(exam: any) {
  return exam.puntaje_total ?? '—'
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>Exámenes</h2>
      <button class="btn-primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancelar' : '+ Nuevo examen' }}
      </button>
    </div>

    <div v-if="showForm" class="card form-card">
      <h3>Nuevo examen</h3>
      <div class="form-row">
        <div class="field">
          <label>Título *</label>
          <input v-model="form.title" placeholder="Ej: Examen de inducción" />
        </div>
        <div class="field">
          <label>Descripción</label>
          <input v-model="form.description" placeholder="Opcional..." />
        </div>
      </div>

      <div class="preguntas">
        <div v-for="(p, pi) in form.preguntas" :key="pi" class="pregunta-card">
          <div class="pregunta-header">
            <span class="num">Pregunta {{ pi + 1 }}</span>
            <button class="btn-icon" @click="removePregunta(pi)">✕</button>
          </div>
          <div class="form-row">
            <div class="field flex-3">
              <label>Enunciado *</label>
              <textarea v-model="p.texto" rows="2" placeholder="Escribe la pregunta..."></textarea>
            </div>
            <div class="field flex-1">
              <label>Valor (pts) *</label>
              <input v-model.number="p.valor" type="number" min="0.5" step="0.5" />
            </div>
          </div>
          <div class="opciones">
            <label>Opciones (selecciona la correcta)</label>
            <div v-for="(o, oi) in p.opciones" :key="oi" class="opcion-row">
              <input
                type="radio"
                :name="'correcta-' + pi"
                :checked="o.es_correcta"
                @change="setCorrecta(pi, oi)"
                title="Marcar como correcta"
              />
              <input v-model="o.texto" placeholder="Texto de la opción..." class="opcion-input" />
              <button class="btn-icon red" @click="removeOpcion(pi, oi)" :disabled="p.opciones.length <= 2">✕</button>
            </div>
            <button class="btn-ghost" @click="addOpcion(pi)">+ Agregar opción</button>
          </div>
        </div>
        <button class="btn-add-pregunta" @click="addPregunta">+ Agregar pregunta</button>
      </div>

      <p v-if="error" class="msg error">{{ error }}</p>
      <p v-if="success" class="msg success">{{ success }}</p>
      <button class="btn-primary" :disabled="loading" @click="guardar">
        {{ loading ? 'Guardando...' : 'Guardar examen' }}
      </button>
    </div>

    <div class="table-wrap">
      <table v-if="examenes.length">
        <thead>
          <tr><th>Título</th><th>Descripción</th><th>Fecha</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="e in examenes" :key="e.id">
            <td><strong>{{ e.title }}</strong></td>
            <td class="desc">{{ e.description || '—' }}</td>
            <td>{{ new Date(e.created_at).toLocaleDateString() }}</td>
            <td><button class="btn-danger-sm" @click="eliminar(e.id)">Eliminar</button></td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">No hay exámenes registrados aún.</div>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; }
.btn-primary { background: #3b82f6; color: white; border: none; padding: 9px 18px; border-radius: 8px; cursor: pointer; font-weight: 600; font-size: 0.9rem; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.card { background: white; border-radius: 12px; padding: 1.5rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); margin-bottom: 1.5rem; }
.form-card h3 { font-size: 1rem; font-weight: 700; margin-bottom: 1rem; color: #334155; }
.form-row { display: flex; gap: 1rem; margin-bottom: 1rem; }
.field { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.field.flex-3 { flex: 3; }
.field.flex-1 { flex: 1; }
label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
input, select, textarea { padding: 9px 12px; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: 0.9rem; outline: none; font-family: inherit; }
input:focus, select:focus, textarea:focus { border-color: #3b82f6; }
.preguntas { display: flex; flex-direction: column; gap: 1rem; margin: 1rem 0; }
.pregunta-card { border: 1.5px solid #e2e8f0; border-radius: 10px; padding: 1rem; background: #f8fafc; }
.pregunta-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; }
.num { font-weight: 700; color: #3b82f6; font-size: 0.9rem; }
.btn-icon { background: none; border: none; cursor: pointer; font-size: 1rem; color: #94a3b8; padding: 2px 6px; border-radius: 4px; }
.btn-icon:hover { background: #e2e8f0; }
.btn-icon.red { color: #ef4444; }
.btn-icon:disabled { opacity: 0.3; cursor: not-allowed; }
.opciones { margin-top: 0.75rem; }
.opciones > label { display: block; margin-bottom: 6px; }
.opcion-row { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.opcion-input { flex: 1; }
.btn-ghost { background: none; border: 1.5px dashed #94a3b8; color: #64748b; padding: 6px 12px; border-radius: 6px; cursor: pointer; font-size: 0.82rem; margin-top: 4px; }
.btn-ghost:hover { border-color: #3b82f6; color: #3b82f6; }
.btn-add-pregunta { background: #eff6ff; color: #3b82f6; border: none; padding: 10px; border-radius: 8px; cursor: pointer; font-weight: 600; font-size: 0.9rem; width: 100%; }
.btn-add-pregunta:hover { background: #dbeafe; }
.msg { padding: 8px 12px; border-radius: 6px; font-size: 0.85rem; margin: 8px 0; }
.msg.error { background: #fee2e2; color: #b91c1c; }
.msg.success { background: #dcfce7; color: #15803d; }
.table-wrap { background: white; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.07); overflow: hidden; }
table { width: 100%; border-collapse: collapse; }
th { background: #f8fafc; padding: 12px 16px; text-align: left; font-size: 0.78rem; font-weight: 700; color: #64748b; border-bottom: 1px solid #e2e8f0; }
td { padding: 12px 16px; border-bottom: 1px solid #f1f5f9; font-size: 0.88rem; }
td.desc { color: #64748b; max-width: 280px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.btn-danger-sm { background: #fee2e2; color: #b91c1c; border: none; padding: 5px 10px; border-radius: 6px; cursor: pointer; font-size: 0.8rem; font-weight: 600; }
.btn-danger-sm:hover { background: #fecaca; }
.empty { padding: 3rem; text-align: center; color: #94a3b8; font-size: 0.95rem; }
</style>
