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
  <div class="page">
    <div class="page-header">
      <h2>Capacitaciones</h2>
      <button class="btn-primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancelar' : '+ Nueva capacitación' }}
      </button>
    </div>

    <div v-if="showForm" class="card form-card">
      <h3>Nueva capacitación</h3>
      <div class="form-grid">
        <div class="field">
          <label>Título *</label>
          <input v-model="form.title" placeholder="Ej: Seguridad en el trabajo" />
        </div>
        <div class="field">
          <label>Tipo *</label>
          <select v-model="form.type">
            <option value="video">Video</option>
            <option value="document">Documento PDF</option>
            <option value="text">Texto enriquecido</option>
          </select>
        </div>
        <div class="field full">
          <label>Descripción</label>
          <textarea v-model="form.description" rows="2" placeholder="Breve descripción..."></textarea>
        </div>
        <div v-if="form.type === 'text'" class="field full">
          <label>Contenido HTML</label>
          <textarea v-model="form.content" rows="5" placeholder="<p>Contenido de la capacitación...</p>"></textarea>
        </div>
        <div v-else class="field full">
          <label>Archivo ({{ form.type === 'video' ? 'MP4, MOV...' : 'PDF, DOCX...' }})</label>
          <input type="file" :accept="form.type === 'video' ? 'video/*' : '.pdf,.doc,.docx'" @change="onFile" />
        </div>
      </div>
      <p v-if="error" class="msg error">{{ error }}</p>
      <p v-if="success" class="msg success">{{ success }}</p>
      <button class="btn-primary" :disabled="loading" @click="guardar">
        {{ loading ? 'Guardando...' : 'Guardar capacitación' }}
      </button>
    </div>

    <div class="table-wrap">
      <table v-if="capacitaciones.length">
        <thead>
          <tr><th>Título</th><th>Tipo</th><th>Descripción</th><th>Fecha</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="c in capacitaciones" :key="c.id">
            <td><strong>{{ c.title }}</strong></td>
            <td><span class="badge type">{{ typeLabel(c.type) }}</span></td>
            <td class="desc">{{ c.description || '—' }}</td>
            <td>{{ new Date(c.created_at).toLocaleDateString() }}</td>
            <td>
              <button class="btn-danger-sm" @click="eliminar(c.id)">Eliminar</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">No hay capacitaciones registradas aún.</div>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; }
.btn-primary {
  background: #3b82f6; color: white; border: none; padding: 9px 18px;
  border-radius: 8px; cursor: pointer; font-weight: 600; font-size: 0.9rem;
}
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.card { background: white; border-radius: 12px; padding: 1.5rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); margin-bottom: 1.5rem; }
.form-card h3 { font-size: 1rem; font-weight: 700; margin-bottom: 1rem; color: #334155; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.field { display: flex; flex-direction: column; gap: 4px; }
.field.full { grid-column: 1 / -1; }
label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
input, select, textarea {
  padding: 9px 12px; border: 1.5px solid #e2e8f0; border-radius: 8px;
  font-size: 0.9rem; outline: none; font-family: inherit;
}
input:focus, select:focus, textarea:focus { border-color: #3b82f6; }
.msg { padding: 8px 12px; border-radius: 6px; font-size: 0.85rem; margin: 8px 0; }
.msg.error { background: #fee2e2; color: #b91c1c; }
.msg.success { background: #dcfce7; color: #15803d; }
.table-wrap { background: white; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.07); overflow: hidden; }
table { width: 100%; border-collapse: collapse; }
th { background: #f8fafc; padding: 12px 16px; text-align: left; font-size: 0.78rem; font-weight: 700; color: #64748b; border-bottom: 1px solid #e2e8f0; }
td { padding: 12px 16px; border-bottom: 1px solid #f1f5f9; font-size: 0.88rem; }
td.desc { color: #64748b; max-width: 280px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.badge.type { font-size: 0.78rem; padding: 3px 8px; background: #eff6ff; color: #1d4ed8; border-radius: 20px; }
.btn-danger-sm { background: #fee2e2; color: #b91c1c; border: none; padding: 5px 10px; border-radius: 6px; cursor: pointer; font-size: 0.8rem; font-weight: 600; }
.btn-danger-sm:hover { background: #fecaca; }
.empty { padding: 3rem; text-align: center; color: #94a3b8; font-size: 0.95rem; }
</style>
