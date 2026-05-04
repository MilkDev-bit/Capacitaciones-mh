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
  <div>
    <!-- Header -->
    <div class="ph">
      <div>
        <h1 class="ph-title">Mis cursos</h1>
        <p class="ph-sub">Crea y gestiona tus cursos. Comparte el c\u00f3digo con tus estudiantes.</p>
      </div>
      <button class="btn btn-primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancelar' : '+ Nuevo curso' }}
      </button>
    </div>

    <!-- Form -->
    <div v-if="showForm" class="form-card">
      <p class="form-card-title">Nuevo curso</p>
      <div class="form-grid">
        <div class="field">
          <label>T\u00edtulo *</label>
          <input class="field-input" v-model="form.title" placeholder="Ej: Introducci\u00f3n a la seguridad" />
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
          <label>Descripci\u00f3n</label>
          <textarea class="field-input" v-model="form.description" rows="2" placeholder="Breve descripci\u00f3n..."></textarea>
        </div>
        <div v-if="form.type === 'text'" class="field full">
          <label>Contenido HTML</label>
          <textarea class="field-input" v-model="form.content" rows="5" placeholder="<p>Contenido del curso...</p>"></textarea>
        </div>
        <div v-else class="field full">
          <label>Archivo ({{ form.type === 'video' ? 'MP4, MOV...' : 'PDF, DOCX...' }})</label>
          <input type="file" :accept="form.type === 'video' ? 'video/*' : '.pdf,.doc,.docx'" @change="onFile" />
        </div>
        <div class="field full">
          <label class="toggle-label">
            <input type="checkbox" v-model="form.is_public" />
            <span>Publicar curso (cualquier usuario puede inscribirse sin c\u00f3digo)</span>
          </label>
        </div>
      </div>
      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div v-if="success" class="alert alert-success">{{ success }}</div>
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="loading" @click="guardar">
          <span v-if="loading" class="spinner" style="width:16px;height:16px"></span>
          {{ loading ? 'Guardando\u2026' : 'Guardar curso' }}
        </button>
        <button class="btn btn-secondary" @click="showForm = false">Cancelar</button>
      </div>
    </div>

    <!-- Cards de cursos -->
    <div v-if="capacitaciones.length" class="courses-grid">
      <div v-for="c in capacitaciones" :key="c.id" class="course-card">
        <div :class="['course-thumb', { 'thumb-video': c.type==='video', 'thumb-document': c.type==='document', 'thumb-text': c.type==='text', 'thumb-default': !['video','document','text'].includes(c.type) }]">
          <span class="thumb-icon">{{ ({ video: '\ud83c\udfa5', document: '\ud83d\udcce', text: '\ud83d\udcdd' } as Record<string,string>)[c.type] || '\ud83d\udcda' }}</span>
          <button
            :class="['vis-badge', c.is_public ? 'vis-public' : 'vis-private']"
            @click="togglePublic(c.id)"
            :title="c.is_public ? 'Hacer privado' : 'Publicar'"
          >
            {{ c.is_public ? '\ud83c\udf10 P\u00fablico' : '\ud83d\udd12 Privado' }}
          </button>
        </div>
        <div class="course-body">
          <span class="course-type">{{ ({ video: 'Video', document: 'Documento', text: 'Texto' } as Record<string,string>)[c.type] || c.type }}</span>
          <h3>{{ c.title }}</h3>
          <p class="course-desc">{{ c.description || 'Sin descripci\u00f3n' }}</p>

          <!-- C\u00f3digo de acceso -->
          <div class="code-section">
            <div class="code-label">C\u00f3digo de acceso</div>
            <div class="code-row">
              <span class="code-val">{{ c.codigo_acceso || '\u2014' }}</span>
              <div class="code-btns" v-if="c.codigo_acceso">
                <button class="code-btn" @click="copiarCodigo(c.codigo_acceso, c.id)" title="Copiar c\u00f3digo">
                  {{ copiedId === c.id + '-code' ? '\u2713' : '\ud83d\udccb' }}
                </button>
                <button class="code-btn" @click="copiarEnlace(c.codigo_acceso, c.id)" title="Copiar enlace">
                  {{ copiedId === c.id + '-link' ? '\u2713 Copiado' : '\ud83d\udd17 Enlace' }}
                </button>
                <button class="code-btn danger" @click="resetCodigo(c.id)" title="Nuevo c\u00f3digo">\ud83d\udd04</button>
              </div>
            </div>
            <div v-if="copiedId === c.id + '-code'" class="copy-toast">C\u00f3digo copiado</div>
            <div v-if="copiedId === c.id + '-link'" class="copy-toast">Enlace copiado</div>
          </div>
        </div>
        <div class="course-footer">
          <span class="course-date">{{ new Date(c.created_at).toLocaleDateString() }}</span>
          <button class="btn btn-danger btn-sm" @click="eliminar(c.id)">Eliminar</button>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <div class="empty-icon">📚</div>
      <h3>Aún no has creado ningún curso</h3>
      <p>Crea tu primer curso y comparte el código con tus estudiantes.</p>
      <button class="btn btn-primary" @click="showForm = true">Crear mi primer curso</button>
    </div>
  </div>
</template>

<style scoped>
.ph { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 12px; margin-bottom: 24px; }
.ph-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.ph-sub { color: var(--muted); font-size: 0.87rem; margin-top: 4px; }

/* Form card */
.form-card { background: var(--surface); border-radius: var(--r-lg); padding: 24px; box-shadow: var(--shadow-sm); margin-bottom: 24px; border-top: 4px solid var(--brand); }
.form-card-title { font-size: 1rem; font-weight: 700; color: var(--dark); margin-bottom: 16px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field.full { grid-column: 1 / -1; }
fieldlabel, label { font-size: 0.78rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .04em; }
input[type="text"], input[type="file"], input[type="number"], select, textarea {
  padding: 9px 12px; border: 1.5px solid var(--border); border-radius: var(--r); font-size: 0.9rem; outline: none; font-family: inherit; background: var(--bg);
}
input:focus, select:focus, textarea:focus { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(249,115,22,.1); }
.toggle-label { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 0.88rem; color: var(--text); font-weight: 500; }
.form-actions { display: flex; gap: 10px; margin-top: 16px; align-items: center; }

/* Courses grid */
.courses-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; }
.course-card { background: var(--surface); border-radius: var(--r-lg); overflow: hidden; box-shadow: var(--shadow-sm); display: flex; flex-direction: column; }
.course-thumb {
  height: 120px; display: flex; align-items: center; justify-content: center; position: relative;
}
.thumb-icon { font-size: 2.5rem; filter: drop-shadow(0 2px 6px rgba(0,0,0,.2)); }

/* Visibility badge */
.vis-badge {
  position: absolute; top: 10px; right: 10px;
  font-size: 0.72rem; font-weight: 700; padding: 3px 10px; border-radius: 20px; border: none; cursor: pointer;
  transition: opacity .15s;
}
.vis-public  { background: rgba(0,200,80,.18); color: #065f46; }
.vis-private { background: rgba(0,0,0,.22); color: #fff; }
.vis-badge:hover { opacity: 0.75; }

.course-body { padding: 16px; display: flex; flex-direction: column; gap: 8px; flex: 1; }
.course-type { font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: .06em; color: var(--brand-dark); background: var(--brand-light); padding: 2px 8px; border-radius: 4px; display: inline-block; width: fit-content; }
.course-body h3 { font-size: 0.97rem; font-weight: 700; color: var(--dark); line-height: 1.35; }
.course-desc { font-size: 0.83rem; color: var(--muted); flex: 1; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

/* Code section */
.code-section { background: var(--bg); border: 1px solid var(--border); border-radius: var(--r); padding: 10px 13px; position: relative; }
.code-label { font-size: 0.68rem; font-weight: 700; color: var(--subtle); text-transform: uppercase; letter-spacing: .07em; margin-bottom: 6px; }
.code-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.code-val { font-size: 1.2rem; font-weight: 800; letter-spacing: .15em; color: var(--dark); font-family: 'Courier New', monospace; }
.code-btns { display: flex; gap: 4px; margin-left: auto; }
.code-btn {
  background: var(--surface); border: 1px solid var(--border); border-radius: 6px;
  padding: 4px 9px; cursor: pointer; font-size: 0.78rem; font-weight: 600; color: var(--text); transition: all .15s; white-space: nowrap;
}
.code-btn:hover { border-color: var(--brand); color: var(--brand); background: var(--brand-light); }
.code-btn.danger:hover { border-color: var(--danger); color: var(--danger); background: var(--danger-bg); }
.copy-toast {
  position: absolute; bottom: -26px; right: 4px;
  background: var(--dark); color: #fff; font-size: 0.7rem; padding: 3px 9px; border-radius: 4px; z-index: 10;
}

.course-footer { padding: 12px 16px; border-top: 1px solid var(--border-light); display: flex; justify-content: space-between; align-items: center; }
.course-date { font-size: 0.76rem; color: var(--subtle); }

.empty-state { text-align: center; padding: 60px 20px; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.empty-icon { font-size: 3rem; }
.empty-state h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
.empty-state p { color: var(--muted); font-size: 0.9rem; max-width: 360px; }

@media (max-width: 680px) {
  .form-grid { grid-template-columns: 1fr; }
  .courses-grid { grid-template-columns: 1fr; }
  .ph { flex-direction: column; }
}
</style>
