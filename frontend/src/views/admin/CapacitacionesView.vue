<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const loading = ref(false)
const showForm = ref(false)
const error = ref('')
const success = ref('')

const form = ref({ title: '', description: '', type: 'video', content: '' })
const file = ref<File | null>(null)

async function load() {
  const res = await api.get('/admin/capacitaciones')
  capacitaciones.value = res.data || []
}

onMounted(load)

function onFile(e: Event) {
  const input = e.target as HTMLInputElement
  file.value = input.files?.[0] ?? null
}

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.title || !form.value.type) {
    error.value = 'El título y tipo son requeridos'; return
  }
  loading.value = true
  try {
    const fd = new FormData()
    fd.append('title', form.value.title)
    fd.append('description', form.value.description)
    fd.append('type', form.value.type)
    fd.append('content', form.value.content)
    if (file.value) fd.append('file', file.value)
    await api.post('/admin/capacitaciones', fd)
    success.value = 'Capacitación creada exitosamente'
    showForm.value = false
    form.value = { title: '', description: '', type: 'video', content: '' }
    file.value = null
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('¿Eliminar esta capacitación?')) return
  await api.delete(`/admin/capacitaciones/${id}`)
  await load()
}

function typeLabel(t: string) {
  return { video: '🎥 Video', document: '📄 Documento', text: '📝 Texto' }[t] || t
}
</script>

<template>
  <div>
    <div class="ph">
      <div>
        <h1 class="ph-title">Capacitaciones</h1>
        <p class="ph-sub">Gestiona todas las capacitaciones de la plataforma.</p>
      </div>
      <button class="btn btn-primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancelar' : '+ Nueva capacitación' }}
      </button>
    </div>

    <div v-if="showForm" class="form-card">
      <p class="form-card-title">Nueva capacitación</p>
      <div class="form-grid">
        <div class="field">
          <label>Título *</label>
          <input class="field-input" v-model="form.title" placeholder="Ej: Seguridad en el trabajo" />
        </div>
        <div class="field">
          <label>Tipo *</label>
          <select class="field-input" v-model="form.type">
            <option value="video">Video</option>
            <option value="document">Documento PDF</option>
            <option value="text">Texto enriquecido</option>
          </select>
        </div>
        <div class="field full">
          <label>Descripción</label>
          <textarea class="field-input" v-model="form.description" rows="2" placeholder="Breve descripción..."></textarea>
        </div>
        <div v-if="form.type === 'text'" class="field full">
          <label>Contenido HTML</label>
          <textarea class="field-input" v-model="form.content" rows="5" placeholder="<p>Contenido...</p>"></textarea>
        </div>
        <div v-else class="field full">
          <label>Archivo</label>
          <input type="file" :accept="form.type === 'video' ? 'video/*' : '.pdf,.doc,.docx'" @change="onFile" />
        </div>
      </div>
      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div v-if="success" class="alert alert-success">{{ success }}</div>
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="loading" @click="guardar">
          <span v-if="loading" class="spinner" style="width:16px;height:16px"></span>
          {{ loading ? 'Guardando…' : 'Guardar capacitación' }}
        </button>
        <button class="btn btn-secondary" @click="showForm = false">Cancelar</button>
      </div>
    </div>

    <div class="table-card">
      <table v-if="capacitaciones.length">
        <thead>
          <tr><th>Título</th><th>Tipo</th><th>Descripción</th><th>Fecha</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="c in capacitaciones" :key="c.id">
            <td><strong>{{ c.title }}</strong></td>
            <td><span class="type-pill">{{ ({ video: '🎥 Video', document: '📎 Doc', text: '📝 Texto' } as Record<string,string>)[c.type] || c.type }}</span></td>
            <td class="cell-muted">{{ c.description || '—' }}</td>
            <td class="cell-muted">{{ new Date(c.created_at).toLocaleDateString() }}</td>
            <td><button class="btn btn-danger btn-sm" @click="eliminar(c.id)">Eliminar</button></td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <div class="empty-icon">📚</div>
        <h3>No hay capacitaciones registradas</h3>
        <button class="btn btn-primary" @click="showForm = true">Crear primera capacitación</button>
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
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field.full { grid-column: 1 / -1; }
label { font-size: 0.78rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .04em; }
.form-actions { display: flex; gap: 10px; margin-top: 16px; }
.table-card { background: var(--surface); border-radius: var(--r-lg); box-shadow: var(--shadow-sm); overflow: hidden; }
table { width: 100%; border-collapse: collapse; }
th { background: var(--bg); padding: 11px 16px; text-align: left; font-size: 0.75rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .05em; border-bottom: 1px solid var(--border); }
td { padding: 12px 16px; border-top: 1px solid var(--border-light); font-size: 0.9rem; }
.cell-muted { color: var(--muted); max-width: 260px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.type-pill { font-size: 0.75rem; padding: 3px 9px; background: var(--brand-light); color: var(--brand-dark); border-radius: 20px; font-weight: 600; white-space: nowrap; }
.empty-state { padding: 60px 20px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.empty-icon { font-size: 3rem; }
.empty-state h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
@media (max-width: 600px) { .form-grid { grid-template-columns: 1fr; } .ph { flex-direction: column; } }
</style>
