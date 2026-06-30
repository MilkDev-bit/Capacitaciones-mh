<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'
import DragDropUpload from '../../components/DragDropUpload.vue'
import EmptyState from '../../components/EmptyState.vue'
import { toast } from '../../utils/toast'
import { uploadToR2 } from '../../utils/upload'

const capacitaciones = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const showDrawer = ref(false)
const editingId = ref<string | null>(null)
const search = ref('')
const page = ref(1)
const totalPages = ref(1)
const limit = 20

const PRESET_COLORS = [
  { hex: '#f97316', name: 'Naranja' },
  { hex: '#ef4444', name: 'Rojo' },
  { hex: '#8b5cf6', name: 'Violeta' },
  { hex: '#3b82f6', name: 'Azul' },
  { hex: '#10b981', name: 'Verde' },
  { hex: '#f59e0b', name: 'Ambar' },
  { hex: '#ec4899', name: 'Rosa' },
  { hex: '#0ea5e9', name: 'Celeste' },
  { hex: '#64748b', name: 'Gris' },
  { hex: '#1e293b', name: 'Oscuro' },
]

const form = ref({
  title: '',
  description: '',
  type: 'video',
  content: '',
  welcome_message: '',
  is_public: false,
  color: '#f97316',
  thumbnail_preview: '',
})

const file = ref<File | null>(null)
const thumbnailFile = ref<File | null>(null)

const isEditing = computed(() => editingId.value !== null)

const filtered = computed(() => {
  const term = search.value.toLowerCase().trim()
  if (!term) return capacitaciones.value
  return capacitaciones.value.filter(c =>
    (c.title || '').toLowerCase().includes(term) ||
    (c.description || '').toLowerCase().includes(term)
  )
})

async function load() {
  loading.value = true
  try {
    const res = await api.get('/admin/capacitaciones', { params: { page: page.value, limit } })
    capacitaciones.value = res.data || []
    totalPages.value = 1
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  editingId.value = null
  form.value = {
    title: '', description: '', type: 'video', content: '',
    welcome_message: '', is_public: false, color: '#f97316', thumbnail_preview: '', duration: 0,
  }
  file.value = null
  thumbnailFile.value = null
  showDrawer.value = true
}

function openEdit(c: any) {
  editingId.value = c.id
  form.value = {
    title: c.title || '',
    description: c.description || '',
    type: c.type || 'video',
    content: c.content || '',
    welcome_message: c.welcome_message || '',
    is_public: !!c.is_public,
    color: c.color || '#f97316',
    thumbnail_preview: c.thumbnail_url || '',
    duration: c.duration || 0,
  }
  file.value = null
  thumbnailFile.value = null
  showDrawer.value = true
}

function closeDrawer() {
  showDrawer.value = false
  editingId.value = null
}

async function guardar() {
  if (!form.value.title.trim()) {
    toast.error('El titulo es requerido'); return
  }
  saving.value = true
  try {
    const payload: Record<string, any> = {
      title: form.value.title.trim(),
      description: form.value.description,
      type: form.value.type,
      content: form.value.content,
      welcome_message: form.value.welcome_message,
      is_public: form.value.is_public,
      color: form.value.color,
      thumbnail_url: form.value.thumbnail_preview || '',
      duration: Number(form.value.duration),
    }
    if (file.value) {
      const prefix = form.value.type === 'video' ? 'videos' : 'documents'
      payload.content = await uploadToR2(file.value, prefix)
    }
    if (thumbnailFile.value) {
      payload.thumbnail_url = await uploadToR2(thumbnailFile.value, 'thumbnails')
    }

    if (isEditing.value) {
      await api.put(`/admin/capacitaciones/${editingId.value}`, payload)
      toast.success('Capacitacion actualizada')
    } else {
      await api.post('/admin/capacitaciones', payload)
      toast.success('Capacitacion creada')
    }
    closeDrawer()
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al guardar')
  } finally {
    saving.value = false
  }
}

async function eliminar(id: string) {
  if (!await toast.confirm('Eliminar esta capacitacion? Esta accion no se puede deshacer.')) return
  await api.delete(`/admin/capacitaciones/${id}`)
  await load()
}

function typeLabel(t: string) {
  return { video: 'Video', document: 'Documento', text: 'Texto', videocall: 'Videollamada' }[t] || t
}

function fileUrl(path: string) {
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
}

function cardColor(c: any) {
  return c.color || '#f97316'
}
</script>
<template>
  <div class="ac-shell">
    <div class="ac-topbar">
      <div>
        <h1 class="ac-title">Capacitaciones</h1>
        <p class="ac-sub">{{ capacitaciones.length }} curso{{ capacitaciones.length !== 1 ? 's' : '' }} en la plataforma</p>
      </div>
      <button class="btn btn-primary ac-new-btn" @click="openCreate">
        <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
        Nueva capacitacion
      </button>
    </div>

    <div class="ac-search-bar">
      <svg width="17" height="17" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z"/>
      </svg>
      <input v-model="search" placeholder="Buscar capacitaciones..." />
    </div>

    <div v-if="loading && !capacitaciones.length" class="ac-grid">
      <div v-for="n in 6" :key="n" class="ac-card ac-card-skel">
        <div class="skeleton" style="height:130px;border-radius:var(--r-lg) var(--r-lg) 0 0"></div>
        <div style="padding:16px">
          <div class="skeleton skel-title"></div>
          <div class="skeleton skel-text" style="margin-top:8px"></div>
        </div>
      </div>
    </div>

    <div v-else-if="filtered.length" class="ac-grid">
      <div v-for="c in filtered" :key="c.id" class="ac-card">
        <div class="ac-card-cover" :style="c.thumbnail_url ? '' : ('background:' + cardColor(c))">
          <template v-if="c.thumbnail_url">
            <img :src="fileUrl(c.thumbnail_url)" alt="Portada" class="ac-card-cover-img" />
          </template>
          <template v-else>
            <span class="ac-card-icon">
              <svg v-if="c.type === 'video'" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
              <svg v-else-if="c.type === 'document'" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
              <svg v-else-if="c.type === 'videocall'" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
              <svg v-else width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 10h16M4 14h16M4 18h7"/></svg>
            </span>
          </template>
          <span class="ac-card-type-badge">{{ typeLabel(c.type) }}</span>
          <span v-if="c.is_public" class="ac-card-public-badge">Publico</span>
        </div>
        <div class="ac-card-body">
          <h3>{{ c.title }}</h3>
          <p class="ac-card-desc">{{ c.description || 'Sin descripcion' }}</p>
          <div class="ac-card-footer">
            <span class="ac-card-date">{{ new Date(c.created_at).toLocaleDateString('es') }}</span>
            <div class="ac-card-actions">
              <button class="btn-icon btn-icon-edit" title="Editar" @click="openEdit(c)">
                <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button class="btn-icon btn-icon-danger" title="Eliminar" @click="eliminar(c.id)">
                <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <EmptyState v-else
      :title="search ? 'Sin resultados' : 'No hay capacitaciones'"
      :description="search ? 'Prueba con otro termino de busqueda.' : 'Crea la primera capacitacion para empezar.'"
    >
      <template #action>
        <button v-if="!search" class="btn btn-primary" @click="openCreate">Crear primera capacitacion</button>
      </template>
    </EmptyState>

    <div v-if="totalPages > 1" class="pagination-controls" style="display:flex;gap:8px;justify-content:center;margin-top:16px;">
      <button class="btn btn-secondary" :disabled="page <= 1" @click="page--; load()">← Anterior</button>
      <span style="line-height:2">Página {{ page }} / {{ totalPages }}</span>
      <button class="btn btn-secondary" :disabled="page >= totalPages" @click="page++; load()">Siguiente →</button>
    </div>
  </div>

  <Teleport to="body">
    <Transition name="fade">
      <div v-if="showDrawer" class="drawer-backdrop" @click="closeDrawer" />
    </Transition>
    <Transition name="slide-right">
      <div v-if="showDrawer" class="ac-drawer" role="dialog">
        <div class="drawer-header">
          <div class="drawer-header-info">
            <div class="drawer-header-icon">
              <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
            </div>
            <div>
              <h2>{{ isEditing ? 'Editar capacitacion' : 'Nueva capacitacion' }}</h2>
              <p>{{ isEditing ? 'Modifica los datos del curso' : 'Completa los datos para crear el curso' }}</p>
            </div>
          </div>
          <button class="drawer-close" @click="closeDrawer" aria-label="Cerrar">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6 6 18M6 6l12 12"/></svg>
          </button>
        </div>

        <div class="drawer-body">
          <div class="drawer-section">
            <div class="drawer-section-label">
              <span class="section-num">1</span>
              Informacion basica
            </div>
            <div class="drawer-fields">
              <div class="dfield dfield-full">
                <label>Titulo <span class="req">*</span></label>
                <input class="field-input" v-model="form.title" placeholder="Ej: Seguridad en el trabajo" maxlength="200" />
                <span class="dfield-hint">{{ form.title.length }}/200 caracteres</span>
              </div>
              <div class="dfield dfield-full">
                <label>Descripcion</label>
                <textarea class="field-input" v-model="form.description" rows="3" placeholder="Breve descripcion del curso para los estudiantes..." style="resize:vertical"></textarea>
              </div>
              <div class="dfield dfield-full">
                <label>Mensaje de bienvenida</label>
                <textarea class="field-input" v-model="form.welcome_message" rows="2" placeholder="Mensaje que veran los estudiantes al abrir el curso..." style="resize:vertical"></textarea>
                <span class="dfield-hint">Se muestra en la pantalla principal del curso</span>
              </div>
            </div>
          </div>

          <div class="drawer-section">
            <div class="drawer-section-label">
              <span class="section-num">2</span>
              Tipo y contenido
            </div>
            <div class="drawer-fields">
              <div class="dfield dfield-full">
                <label>Tipo de capacitacion <span class="req">*</span></label>
                <div class="type-selector">
                  <label :class="['type-opt', form.type === 'video' && 'type-opt-active']">
                    <input type="radio" v-model="form.type" value="video" />
                    <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                    <span>Video</span>
                  </label>
                  <label :class="['type-opt', form.type === 'document' && 'type-opt-active']">
                    <input type="radio" v-model="form.type" value="document" />
                    <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
                    <span>Documento</span>
                  </label>
                  <label :class="['type-opt', form.type === 'text' && 'type-opt-active']">
                    <input type="radio" v-model="form.type" value="text" />
                    <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 10h16M4 14h16M4 18h7"/></svg>
                    <span>Texto</span>
                  </label>
                  <label :class="['type-opt', form.type === 'videocall' && 'type-opt-active']">
                    <input type="radio" v-model="form.type" value="videocall" />
                    <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
                    <span>Videollamada</span>
                  </label>
                </div>
              </div>
              <div v-if="form.type === 'videocall'" class="dfield dfield-full">
                <label>Duración (Minutos) <span class="req">*</span></label>
                <input class="field-input" type="number" v-model="form.duration" min="0" placeholder="Ej: 60" />
                <span class="dfield-hint">La duración en minutos de la videollamada.</span>
              </div>
              <div v-if="form.type === 'video' || form.type === 'document'" class="dfield dfield-full">
                <label>Archivo principal</label>
                <DragDropUpload
                  v-model="file"
                  :accept="form.type === 'video' ? 'video/mp4,video/webm' : '.pdf,.doc,.docx'"
                />
              </div>
              <div v-if="form.type === 'text'" class="dfield dfield-full">
                <label>Contenido de texto</label>
                <textarea class="field-input" v-model="form.content" rows="6" placeholder="Escribe el contenido del curso..." style="resize:vertical"></textarea>
              </div>
            </div>
          </div>

          <div class="drawer-section">
            <div class="drawer-section-label">
              <span class="section-num">3</span>
              Personalizacion
            </div>
            <div class="drawer-fields">
              <div class="dfield dfield-full">
                <label>Color de la tarjeta</label>
                <div class="color-palette">
                  <button
                    v-for="preset in PRESET_COLORS"
                    :key="preset.hex"
                    :title="preset.name"
                    :class="['color-dot', form.color === preset.hex && 'color-dot-active']"
                    :style="'background:' + preset.hex"
                    @click="form.color = preset.hex"
                    type="button"
                  />
                  <label class="color-custom-wrap" title="Color personalizado">
                    <input type="color" v-model="form.color" class="color-custom-input" />
                    <span class="color-custom-btn" :style="'background:' + form.color">
                      <svg width="13" height="13" fill="currentColor" viewBox="0 0 24 24"><path d="M12 3a9 9 0 100 18A9 9 0 0012 3zm0 16.5A7.5 7.5 0 1112 4.5a7.5 7.5 0 010 15zm-1-9.75a1.25 1.25 0 112.5 0 1.25 1.25 0 01-2.5 0zm-3 2.5a1.25 1.25 0 112.5 0 1.25 1.25 0 01-2.5 0zm6 0a1.25 1.25 0 112.5 0 1.25 1.25 0 01-2.5 0zm-3 3.5a1.25 1.25 0 112.5 0 1.25 1.25 0 01-2.5 0z"/></svg>
                    </span>
                  </label>
                </div>
                <div class="color-preview-chip">
                  <span class="color-preview-dot" :style="'background:' + form.color"></span>
                  <span class="color-preview-label">{{ form.title || 'Titulo del curso' }}</span>
                </div>
              </div>

              <div class="dfield dfield-full">
                <label>Miniatura (portada)</label>
                <div v-if="form.thumbnail_preview && !thumbnailFile" class="thumb-preview-wrap">
                  <img :src="fileUrl(form.thumbnail_preview)" class="thumb-preview-img" alt="Miniatura actual" />
                  <span class="thumb-preview-label">Miniatura actual &mdash; sube una nueva para reemplazarla</span>
                </div>
                <DragDropUpload v-model="thumbnailFile" accept=".jpg,.jpeg,.png,.webp" />
                <span class="dfield-hint">JPG, PNG o WEBP. Recomendado: 800x450 px.</span>
              </div>
            </div>
          </div>

          <div class="drawer-section">
            <div class="drawer-section-label">
              <span class="section-num">4</span>
              Acceso y visibilidad
            </div>
            <div class="drawer-fields">
              <div class="dfield dfield-full">
                <div class="toggle-row">
                  <div class="toggle-info">
                    <span class="toggle-title">Curso publico</span>
                    <span class="toggle-desc">Cualquier usuario puede unirse sin necesitar asignacion</span>
                  </div>
                  <button
                    type="button"
                    :class="['toggle-btn', form.is_public && 'toggle-btn-on']"
                    @click="form.is_public = !form.is_public"
                    role="switch"
                    :aria-checked="form.is_public"
                  >
                    <span class="toggle-thumb" />
                  </button>
                </div>
              </div>
            </div>
          </div>

        </div>

        <div class="drawer-footer">
          <button class="btn btn-primary drawer-save-btn" :disabled="saving" @click="guardar">
            <svg v-if="!saving" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
            <span v-if="saving" class="spinner-xs"></span>
            {{ saving ? 'Guardando...' : (isEditing ? 'Guardar cambios' : 'Crear capacitacion') }}
          </button>
          <button class="btn btn-secondary" @click="closeDrawer">Cancelar</button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.ac-shell { display: flex; flex-direction: column; gap: 20px; }
.ac-topbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.ac-new-btn { display: flex; align-items: center; gap: 7px; }
.ac-title { font-size: 1.65rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; }
.ac-sub { color: var(--muted); font-size: 0.88rem; margin-top: 3px; }

.ac-search-bar {
  display: flex; align-items: center; gap: 10px; padding: 11px 16px;
  border: 1.5px solid var(--border); border-radius: var(--r);
  background: var(--surface); color: var(--muted);
  box-shadow: var(--shadow-xs); max-width: 400px;
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
  height: 130px; display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden; background: var(--brand);
}
.ac-card-cover-img { width: 100%; height: 100%; object-fit: cover; }
.ac-card-icon { filter: drop-shadow(0 2px 6px rgba(0,0,0,.3)); color: rgba(255,255,255,.9); }
.ac-card-type-badge {
  position: absolute; top: 10px; left: 10px; padding: 3px 9px;
  border-radius: 999px; background: rgba(0,0,0,.35); color: #fff;
  font-size: 0.7rem; font-weight: 700; backdrop-filter: blur(6px);
}
.ac-card-public-badge {
  position: absolute; top: 10px; right: 10px;
  padding: 3px 9px; border-radius: 999px; background: rgba(16,185,129,.85); color: #fff;
  font-size: 0.7rem; font-weight: 700; backdrop-filter: blur(6px);
}
.ac-card-body { padding: 16px; display: flex; flex-direction: column; gap: 8px; }
.ac-card-body h3 { font-size: 0.97rem; font-weight: 700; color: var(--dark); line-height: 1.35; }
.ac-card-desc {
  font-size: 0.82rem; color: var(--muted); line-height: 1.45;
  display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.ac-card-footer { display: flex; align-items: center; justify-content: space-between; padding-top: 10px; border-top: 1px solid var(--border-light); }
.ac-card-date { font-size: 0.75rem; color: var(--subtle); }
.ac-card-actions { display: flex; gap: 6px; }

.btn-icon {
  display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: var(--r); border: 1px solid var(--border);
  background: var(--surface); cursor: pointer; transition: background .15s, border-color .15s; color: var(--muted);
}
.btn-icon-edit:hover { background: rgba(59,130,246,.1); border-color: #3b82f6; color: #3b82f6; }
.btn-icon-danger:hover { background: rgba(239,68,68,.1); border-color: var(--danger); color: var(--danger); }

.drawer-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,.45);
  backdrop-filter: blur(2px); z-index: 1400;
}

.ac-drawer {
  position: fixed; top: 0; right: 0; bottom: 0; width: 520px; max-width: 100vw;
  background: var(--bg); z-index: 1401;
  display: flex; flex-direction: column;
  box-shadow: -8px 0 40px rgba(0,0,0,.18);
  border-left: 1px solid var(--border);
}

.drawer-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px; border-bottom: 1px solid var(--border);
  background: var(--dark); flex-shrink: 0;
}
.drawer-header-info { display: flex; align-items: center; gap: 14px; }
.drawer-header-icon {
  display: flex; align-items: center; justify-content: center;
  width: 42px; height: 42px; border-radius: 12px;
  background: rgba(255,255,255,.15); color: #fff; flex-shrink: 0;
}
.drawer-header h2 { font-size: 1.05rem; font-weight: 800; color: #fff; margin: 0; }
.drawer-header p { font-size: 0.8rem; color: rgba(255,255,255,.6); margin: 2px 0 0; }
.drawer-close {
  display: flex; align-items: center; justify-content: center;
  width: 34px; height: 34px; border-radius: var(--r);
  background: rgba(255,255,255,.12); border: 0; cursor: pointer; color: rgba(255,255,255,.8);
  transition: background .15s;
}
.drawer-close:hover { background: rgba(255,255,255,.22); color: #fff; }

.drawer-body { flex: 1; overflow-y: auto; }
.drawer-section { padding: 22px 24px; border-bottom: 1px solid var(--border-light); }
.drawer-section:last-child { border-bottom: 0; }
.drawer-section-label {
  display: flex; align-items: center; gap: 10px;
  font-size: 0.78rem; font-weight: 800; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--muted); margin-bottom: 16px;
}
.section-num {
  display: flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; border-radius: 50%;
  background: var(--brand); color: #fff; font-size: 0.72rem; font-weight: 900; flex-shrink: 0;
}

.drawer-fields { display: flex; flex-direction: column; gap: 14px; }
.dfield { display: flex; flex-direction: column; gap: 5px; }
.dfield-full { width: 100%; }
.dfield label { font-size: 0.8rem; font-weight: 700; color: var(--dark); }
.dfield .req { color: var(--danger); }
.dfield-hint { font-size: 0.74rem; color: var(--muted); }

.type-selector { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.type-opt {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 14px 8px; border-radius: var(--r-lg);
  border: 2px solid var(--border); background: var(--surface);
  cursor: pointer; transition: border-color .18s, background .18s;
  font-size: 0.82rem; font-weight: 600; color: var(--muted); text-align: center;
}
.type-opt input[type=radio] { display: none; }
.type-opt:hover { border-color: var(--brand); color: var(--brand); }
.type-opt-active { border-color: var(--brand) !important; background: rgba(249,115,22,.06) !important; color: var(--brand) !important; }

.color-palette { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; margin-bottom: 10px; }
.color-dot {
  width: 28px; height: 28px; border-radius: 50%; border: 3px solid transparent;
  cursor: pointer; transition: transform .15s, box-shadow .15s; outline: none;
}
.color-dot:hover { transform: scale(1.15); }
.color-dot-active { border-color: white !important; box-shadow: 0 0 0 2px var(--dark); transform: scale(1.15); }

.color-custom-wrap { position: relative; cursor: pointer; }
.color-custom-input { position: absolute; opacity: 0; width: 0; height: 0; }
.color-custom-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 50%;
  border: 2px dashed rgba(255,255,255,.6); color: rgba(255,255,255,.9);
  cursor: pointer; transition: transform .15s;
}
.color-custom-btn:hover { transform: scale(1.15); }

.color-preview-chip {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; border-radius: var(--r);
  background: var(--surface); border: 1px solid var(--border);
}
.color-preview-dot { width: 14px; height: 14px; border-radius: 50%; flex-shrink: 0; }
.color-preview-label { font-size: 0.8rem; color: var(--dark); font-weight: 500; }

.thumb-preview-wrap {
  display: flex; align-items: center; gap: 12px; margin-bottom: 8px;
  padding: 10px; background: var(--surface); border-radius: var(--r);
  border: 1px solid var(--border);
}
.thumb-preview-img { width: 72px; height: 42px; object-fit: cover; border-radius: 6px; flex-shrink: 0; }
.thumb-preview-label { font-size: 0.76rem; color: var(--muted); }

.toggle-row {
  display: flex; align-items: center; justify-content: space-between; gap: 16px;
  padding: 14px 16px; background: var(--surface); border-radius: var(--r-lg);
  border: 1px solid var(--border);
}
.toggle-info { display: flex; flex-direction: column; gap: 2px; }
.toggle-title { font-size: 0.88rem; font-weight: 700; color: var(--dark); }
.toggle-desc { font-size: 0.78rem; color: var(--muted); }
.toggle-btn {
  position: relative; width: 44px; height: 24px; border-radius: 999px;
  background: var(--border); border: 0; cursor: pointer; flex-shrink: 0;
  transition: background .2s;
}
.toggle-btn-on { background: var(--brand); }
.toggle-thumb {
  position: absolute; top: 3px; left: 3px;
  width: 18px; height: 18px; border-radius: 50%; background: #fff;
  box-shadow: 0 1px 4px rgba(0,0,0,.2);
  transition: transform .2s cubic-bezier(0.34,1.56,0.64,1);
}
.toggle-btn-on .toggle-thumb { transform: translateX(20px); }

.drawer-footer {
  display: flex; align-items: center; gap: 10px; padding: 16px 24px;
  border-top: 1px solid var(--border); background: var(--surface); flex-shrink: 0;
}
.drawer-save-btn { display: flex; align-items: center; gap: 7px; flex: 1; justify-content: center; }

.spinner-xs {
  display: inline-block; width: 14px; height: 14px;
  border: 2px solid rgba(255,255,255,.4); border-top-color: #fff;
  border-radius: 50%; animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.fade-enter-active, .fade-leave-active { transition: opacity .25s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.slide-right-enter-active, .slide-right-leave-active { transition: transform .3s cubic-bezier(0.32,0.72,0,1); }
.slide-right-enter-from, .slide-right-leave-to { transform: translateX(100%); }

@media (max-width: 600px) {
  .ac-topbar { flex-direction: column; align-items: stretch; }
  .ac-grid { grid-template-columns: 1fr; }
  .ac-drawer { width: 100vw; }
}
</style>
