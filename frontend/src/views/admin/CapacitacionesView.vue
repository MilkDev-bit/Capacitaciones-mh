<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const loading = ref(true)
const showForm = ref(false)
const error = ref('')
const success = ref('')
const search = ref('')

const form = ref({ title: '', description: '', type: 'video', content: '' })
const file = ref<File | null>(null)

const filtered = computed(() => {
  const term = search.value.toLowerCase().trim()
  if (!term) return capacitaciones.value
  return capacitaciones.value.filter(c =>
    (c.title || '').toLowerCase().includes(term) || (c.description || '').toLowerCase().includes(term)
  )
})

async function load() {
  loading.value = true
  try {
    const res = await api.get('/admin/capacitaciones')
    capacitaciones.value = res.data || []
  } finally {
    loading.value = false
  }
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
    setTimeout(() => { success.value = '' }, 3000)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('¿Eliminar esta capacitación? Esta acción no se puede deshacer.')) return
  await api.delete(`/admin/capacitaciones/${id}`)
  await load()
}

function typeLabel(t: string) {
  return { video: 'Video', document: 'Documento', text: 'Texto' }[t] || t
}
function typeIcon(t: string) {
  return { video: '🎥', document: '📄', text: '📝' }[t] || '📁'
}
function thumbClass(t: string) {
  return { video: 'thumb-video', document: 'thumb-document', text: 'thumb-text' }[t] || 'thumb-default'
}
</script>

<template>
  <div class="ac-shell">
    <!-- Header -->
    <div class="ac-topbar">
      <div>
        <h1 class="ac-title">Capacitaciones</h1>
        <p class="ac-sub">{{ capacitaciones.length }} curso{{ capacitaciones.length !== 1 ? 's' : '' }} en la plataforma</p>
      </div>
      <button class="btn btn-primary" @click="showForm = !showForm">
        {{ showForm ? '✕ Cancelar' : '+ Nueva capacitación' }}
      </button>
    </div>

    <!-- Alerts -->
    <Transition name="slide-down">
      <div v-if="error" class="alert alert-error">{{ error }}</div>
    </Transition>
    <Transition name="slide-down">
      <div v-if="success" class="alert alert-success">{{ success }}</div>
    </Transition>

    <!-- Form -->
    <Transition name="slide-down">
      <div v-if="showForm" class="ac-form-card">
        <div class="ac-form-header">
          <span>📚</span>
          <div>
            <h2>Nueva capacitación</h2>
            <p>Completa los datos para crear un nuevo curso</p>
          </div>
        </div>
        <div class="ac-form-body">
          <div class="ac-form-grid">
            <div class="ac-field ac-field-full">
              <label>Título *</label>
              <input class="field-input" v-model="form.title" placeholder="Ej: Seguridad en el trabajo" />
            </div>
            <div class="ac-field">
              <label>Tipo *</label>
              <select class="field-input" v-model="form.type">
                <option value="video">🎥 Video</option>
                <option value="document">📄 Documento PDF</option>
                <option value="text">📝 Texto enriquecido</option>
              </select>
            </div>
            <div class="ac-field">
              <label>Archivo {{ form.type === 'text' ? '(no aplica)' : '' }}</label>
              <input v-if="form.type !== 'text'" type="file" :accept="form.type === 'video' ? 'video/*' : '.pdf,.doc,.docx'" @change="onFile" />
              <p v-else style="font-size:0.82rem;color:var(--muted);padding:10px 0">El contenido se escribe abajo</p>
            </div>
            <div class="ac-field ac-field-full">
              <label>Descripción</label>
              <textarea class="field-input" v-model="form.description" rows="2" placeholder="Breve descripción del curso..." style="resize:vertical"></textarea>
            </div>
            <div v-if="form.type === 'text'" class="ac-field ac-field-full">
              <label>Contenido</label>
              <textarea class="field-input" v-model="form.content" rows="5" placeholder="Escribe el contenido..." style="resize:vertical"></textarea>
            </div>
          </div>
          <div class="ac-form-actions">
            <button class="btn btn-primary" :disabled="loading" @click="guardar">
              {{ loading ? 'Guardando…' : 'Crear capacitación' }}
            </button>
            <button class="btn btn-secondary" @click="showForm = false">Cancelar</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Search -->
    <div class="ac-search-bar">
      <svg width="17" height="17" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
      </svg>
      <input v-model="search" placeholder="Buscar capacitaciones..." />
    </div>

    <!-- Grid -->
    <div v-if="loading && !capacitaciones.length" class="ac-grid">
      <div v-for="n in 6" :key="n" class="ac-card ac-card-skel">
        <div class="skeleton" style="height:120px;border-radius:var(--r-lg) var(--r-lg) 0 0"></div>
        <div style="padding:16px"><div class="skeleton skel-title"></div><div class="skeleton skel-text" style="margin-top:8px"></div></div>
      </div>
    </div>

    <div v-else-if="filtered.length" class="ac-grid">
      <div v-for="c in filtered" :key="c.id" class="ac-card">
        <div :class="['ac-card-cover', thumbClass(c.type)]">
          <span class="ac-card-icon">{{ typeIcon(c.type) }}</span>
          <span class="ac-card-type-badge">{{ typeLabel(c.type) }}</span>
        </div>
        <div class="ac-card-body">
          <h3>{{ c.title }}</h3>
          <p class="ac-card-desc">{{ c.description || 'Sin descripción' }}</p>
          <div class="ac-card-footer">
            <span class="ac-card-date">{{ new Date(c.created_at).toLocaleDateString('es') }}</span>
            <button class="btn btn-danger btn-sm" @click="eliminar(c.id)">Eliminar</button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <div class="empty-icon">📚</div>
      <h3>{{ search ? 'Sin resultados' : 'No hay capacitaciones' }}</h3>
      <p>{{ search ? 'Prueba con otro término de búsqueda.' : 'Crea la primera capacitación para empezar.' }}</p>
      <button v-if="!search" class="btn btn-primary" @click="showForm = true">Crear primera capacitación</button>
    </div>
  </div>
</template>

<style scoped>
.ac-shell { display: flex; flex-direction: column; gap: 20px; }
.ac-topbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.ac-title { font-size: 1.65rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; }
.ac-sub { color: var(--muted); font-size: 0.88rem; margin-top: 3px; }

.ac-form-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  border: 1.5px solid var(--border); box-shadow: 0 4px 20px rgba(0,0,0,.06);
}
.ac-form-header {
  display: flex; align-items: center; gap: 16px;
  padding: 22px 28px; background: var(--dark); color: #fff;
}
.ac-form-header span { font-size: 2rem; }
.ac-form-header h2 { font-size: 1.1rem; font-weight: 800; color: #fff; margin: 0; }
.ac-form-header p { font-size: 0.83rem; color: rgba(255,255,255,.65); margin: 2px 0 0; }
.ac-form-body { padding: 24px 28px; }
.ac-form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.ac-field { display: flex; flex-direction: column; gap: 6px; }
.ac-field-full { grid-column: 1 / -1; }
.ac-field label { font-size: 0.82rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: 0.04em; }
.ac-form-actions { display: flex; gap: 10px; margin-top: 20px; }

.ac-search-bar {
  display: flex; align-items: center; gap: 10px; padding: 11px 16px;
  border: 1.5px solid var(--border); border-radius: var(--r); background: var(--surface);
  color: var(--muted); box-shadow: var(--shadow-xs); max-width: 400px;
}
.ac-search-bar input { width: 100%; border: 0; outline: 0; background: transparent; color: var(--dark); font-size: 0.9rem; }
.ac-search-bar:focus-within { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(249,115,22,.12); }

.ac-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 20px; }
.ac-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
  transition: transform 0.22s cubic-bezier(0.34,1.56,0.64,1), box-shadow 0.22s;
}
.ac-card:hover { transform: translateY(-4px); box-shadow: var(--shadow-lg); }
.ac-card-skel { pointer-events: none; }
.ac-card-cover {
  height: 120px; display: flex; align-items: center; justify-content: center; position: relative;
}
.ac-card-icon { font-size: 2.4rem; filter: drop-shadow(0 2px 6px rgba(0,0,0,.25)); }
.ac-card-type-badge {
  position: absolute; top: 10px; left: 10px; padding: 3px 9px; border-radius: 999px;
  background: rgba(255,255,255,.18); color: rgba(255,255,255,.9); font-size: 0.72rem;
  font-weight: 800; backdrop-filter: blur(6px);
}
.ac-card-body { padding: 16px; display: flex; flex-direction: column; gap: 8px; }
.ac-card-body h3 { font-size: 0.97rem; font-weight: 700; color: var(--dark); line-height: 1.35; }
.ac-card-desc {
  font-size: 0.82rem; color: var(--muted); line-height: 1.45;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.ac-card-footer { display: flex; align-items: center; justify-content: space-between; padding-top: 10px; border-top: 1px solid var(--border-light); }
.ac-card-date { font-size: 0.75rem; color: var(--subtle); }

@media (max-width: 600px) {
  .ac-topbar { flex-direction: column; align-items: stretch; }
  .ac-form-grid { grid-template-columns: 1fr; }
  .ac-grid { grid-template-columns: 1fr; }
}
</style>
