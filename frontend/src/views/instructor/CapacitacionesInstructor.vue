<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const loading = ref(false)
const showForm = ref(false)
const error = ref('')
const success = ref('')
const copiedId = ref<string | null>(null)
const expandedCode = ref<string | null>(null)

const form = ref({ title: '', description: '', type: 'video', content: '', is_public: false })
const file = ref<File | null>(null)

async function load() {
  const res = await api.get('/instructor/capacitaciones')
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
    fd.append('is_public', String(form.value.is_public))
    if (file.value) fd.append('file', file.value)
    await api.post('/instructor/capacitaciones', fd)
    success.value = 'Curso creado exitosamente'
    showForm.value = false
    form.value = { title: '', description: '', type: 'video', content: '', is_public: false }
    file.value = null
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('¿Eliminar este curso?')) return
  await api.delete(`/instructor/capacitaciones/${id}`)
  await load()
}

async function togglePublic(id: string) {
  await api.patch(`/instructor/capacitaciones/${id}/toggle-public`)
  await load()
}

async function resetCodigo(id: string) {
  if (!confirm('¿Generar un nuevo código? El anterior dejará de funcionar.')) return
  const res = await api.post(`/instructor/capacitaciones/${id}/reset-codigo`)
  const cap = capacitaciones.value.find(c => c.id === id)
  if (cap) cap.codigo_acceso = res.data.codigo_acceso
}

function shareLink(codigo: string) {
  return `${window.location.origin}/unirse/${codigo}`
}

async function copiarCodigo(codigo: string, id: string) {
  await navigator.clipboard.writeText(codigo)
  copiedId.value = id + '-code'
  setTimeout(() => { copiedId.value = null }, 2000)
}

async function copiarEnlace(codigo: string, id: string) {
  await navigator.clipboard.writeText(shareLink(codigo))
  copiedId.value = id + '-link'
  setTimeout(() => { copiedId.value = null }, 2000)
}

function typeLabel(t: string) {
  return { video: '🎥 Video', document: '📄 Documento', text: '📝 Texto' }[t] || t
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>Mis cursos</h2>
        <p class="subtitle">Crea y gestiona tus cursos. Comparte el código o enlace para que los estudiantes se unan.</p>
      </div>
      <button class="btn-primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancelar' : '+ Nuevo curso' }}
      </button>
    </div>

    <div v-if="showForm" class="card form-card">
      <h3>Nuevo curso</h3>
      <div class="form-grid">
        <div class="field">
          <label>Título *</label>
          <input v-model="form.title" placeholder="Ej: Introducción a la seguridad" />
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
          <textarea v-model="form.content" rows="5" placeholder="<p>Contenido del curso...</p>"></textarea>
        </div>
        <div v-else class="field full">
          <label>Archivo ({{ form.type === 'video' ? 'MP4, MOV...' : 'PDF, DOCX...' }})</label>
          <input type="file" :accept="form.type === 'video' ? 'video/*' : '.pdf,.doc,.docx'" @change="onFile" />
        </div>
        <div class="field full">
          <label class="toggle-label">
            <input type="checkbox" v-model="form.is_public" />
            <span>Publicar curso (cualquier usuario puede inscribirse sin código)</span>
          </label>
        </div>
      </div>
      <p v-if="error" class="msg error">{{ error }}</p>
      <p v-if="success" class="msg success">{{ success }}</p>
      <button class="btn-primary" :disabled="loading" @click="guardar">
        {{ loading ? 'Guardando...' : 'Guardar curso' }}
      </button>
    </div>

    <!-- Cards de cursos -->
    <div v-if="capacitaciones.length" class="courses-grid">
      <div v-for="c in capacitaciones" :key="c.id" class="course-card">
        <div class="course-top">
          <span class="badge type">{{ typeLabel(c.type) }}</span>
          <button
            class="badge-btn"
            :class="c.is_public ? 'public' : 'private'"
            @click="togglePublic(c.id)"
            :title="c.is_public ? 'Clic para hacer privado' : 'Clic para publicar'"
          >
            {{ c.is_public ? '🌐 Público' : '🔒 Privado' }}
          </button>
        </div>
        <h3>{{ c.title }}</h3>
        <p class="desc">{{ c.description || 'Sin descripción' }}</p>

        <!-- Sección de código de acceso -->
        <div class="code-section">
          <div class="code-label">Código de acceso</div>
          <div class="code-box">
            <span class="code-value">{{ c.codigo_acceso || '—' }}</span>
            <div class="code-actions" v-if="c.codigo_acceso">
              <button
                class="icon-btn"
                @click="copiarCodigo(c.codigo_acceso, c.id)"
                title="Copiar código"
              >
                {{ copiedId === c.id + '-code' ? '✓' : '📋' }}
              </button>
              <button
                class="icon-btn"
                @click="copiarEnlace(c.codigo_acceso, c.id)"
                title="Copiar enlace de invitación"
              >
                {{ copiedId === c.id + '-link' ? '✓ Copiado' : '🔗 Copiar enlace' }}
              </button>
              <button
                class="icon-btn danger"
                @click="resetCodigo(c.id)"
                title="Generar nuevo código"
              >
                🔄
              </button>
            </div>
          </div>
          <div v-if="copiedId === c.id + '-code'" class="copy-toast">Código copiado</div>
          <div v-if="copiedId === c.id + '-link'" class="copy-toast">Enlace copiado</div>
        </div>

        <div class="course-footer">
          <span class="date">{{ new Date(c.created_at).toLocaleDateString() }}</span>
          <button class="btn-danger-sm" @click="eliminar(c.id)">Eliminar</button>
        </div>
      </div>
    </div>

    <div v-else class="empty">
      <p>Aún no has creado ningún curso.</p>
      <button class="btn-primary" @click="showForm = true">Crear mi primer curso</button>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; gap: 1rem; }
.page-header h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; margin: 0; }
.subtitle { color: #64748b; font-size: 0.85rem; margin-top: 4px; }
.btn-primary {
  background: #7c3aed; color: white; border: none;
  border-radius: 8px; padding: 9px 18px; cursor: pointer; font-size: 0.9rem; font-weight: 600; white-space: nowrap;
}
.btn-primary:hover:not(:disabled) { background: #6d28d9; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.card { background: white; border-radius: 12px; padding: 1.5rem; box-shadow: 0 1px 6px rgba(0,0,0,0.08); margin-bottom: 1.5rem; }
.form-card h3 { font-size: 1rem; font-weight: 700; color: #334155; margin-bottom: 1rem; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.field { display: flex; flex-direction: column; gap: 4px; }
.field.full { grid-column: 1 / -1; }
label { font-size: 0.8rem; font-weight: 600; color: #64748b; }
input[type="text"], input[type="file"], select, textarea {
  border: 1.5px solid #e2e8f0; border-radius: 8px; padding: 8px 10px; font-size: 0.9rem; outline: none;
}
input:focus, select:focus, textarea:focus { border-color: #7c3aed; }
.toggle-label { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 0.9rem; color: #334155; font-weight: 500; }
.toggle-label input[type="checkbox"] { width: 16px; height: 16px; cursor: pointer; }
.msg { font-size: 0.85rem; margin-top: 8px; padding: 8px 12px; border-radius: 6px; }
.msg.error { background: #fee2e2; color: #dc2626; }
.msg.success { background: #d1fae5; color: #065f46; }
/* Courses grid */
.courses-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 1.2rem; }
.course-card {
  background: white; border-radius: 12px; padding: 1.4rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.07); display: flex; flex-direction: column; gap: 0.6rem;
}
.course-top { display: flex; align-items: center; gap: 8px; }
.course-card h3 { font-size: 1rem; font-weight: 700; color: #1e293b; margin: 0; }
.desc { font-size: 0.85rem; color: #64748b; line-height: 1.4; flex: 1; }
.badge { font-size: 0.75rem; padding: 3px 8px; border-radius: 20px; font-weight: 600; }
.badge.type { background: #ede9fe; color: #6d28d9; }
.badge-btn {
  font-size: 0.75rem; padding: 4px 10px; border-radius: 20px; font-weight: 600;
  border: none; cursor: pointer; transition: all 0.15s;
}
.badge-btn.public { background: #d1fae5; color: #065f46; }
.badge-btn.private { background: #f1f5f9; color: #475569; }
.badge-btn:hover { opacity: 0.8; }
/* Code section */
.code-section { background: #f8fafc; border-radius: 8px; padding: 10px 12px; border: 1px solid #e2e8f0; position: relative; }
.code-label { font-size: 0.72rem; font-weight: 700; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 6px; }
.code-box { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.code-value { font-size: 1.3rem; font-weight: 800; letter-spacing: 0.15em; color: #1e293b; font-family: 'Courier New', monospace; }
.code-actions { display: flex; align-items: center; gap: 4px; margin-left: auto; }
.icon-btn {
  background: white; border: 1px solid #e2e8f0; border-radius: 6px;
  padding: 4px 8px; cursor: pointer; font-size: 0.8rem; transition: all 0.15s; white-space: nowrap;
}
.icon-btn:hover { background: #ede9fe; border-color: #c4b5fd; }
.icon-btn.danger:hover { background: #fee2e2; border-color: #fca5a5; }
.copy-toast {
  position: absolute; bottom: -24px; right: 0;
  background: #1e293b; color: white; font-size: 0.72rem; padding: 3px 8px; border-radius: 4px;
}
.course-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 4px; }
.date { font-size: 0.78rem; color: #94a3b8; }
.btn-danger-sm {
  background: #fee2e2; color: #dc2626; border: none;
  border-radius: 6px; padding: 5px 12px; cursor: pointer; font-size: 0.8rem; font-weight: 600;
}
.btn-danger-sm:hover { background: #fecaca; }
.empty { padding: 3rem; text-align: center; color: #64748b; display: flex; flex-direction: column; align-items: center; gap: 1rem; }
</style>
