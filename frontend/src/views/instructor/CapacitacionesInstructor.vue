<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import api from '../../api'
import DragDropUpload from '../../components/DragDropUpload.vue'

const capacitaciones = ref<any[]>([])
const loading = ref(false)
const showForm = ref(false)
const error = ref('')
const success = ref('')
const form = ref({ title: '', description: '', type: 'video', content: '', is_public: false, welcome_message: '', thumbnail_url: '' })
const file = ref<File | null>(null)
const thumbnailFile = ref<File | null>(null)

const selectedCurso = ref<any | null>(null)
const activeTab = ref<'lecciones' | 'intermedias' | 'examen'>('lecciones')

const lecciones = ref<any[]>([])
const loadingLec = ref(false)
const showLecForm = ref(false)
const lecForm = ref({ title: '', description: '', type: 'video', content: '', orden: 1, duracion_min: 0 })
const lecFile = ref<File | null>(null)

const intermedias = ref<any[]>([])
const loadingInt = ref(false)
const showIntForm = ref(false)
const intForm = ref({
  texto: '',
  tipo: 'multiple_choice',
  orden: 1,
  despues_de_leccion_id: '',
  opciones: [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }]
})

const misExamenes = ref<any[]>([])

async function load() {
  loading.value = true
  try {
    const res = await api.get('/instructor/capacitaciones')
    capacitaciones.value = res.data || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al cargar los cursos'
  } finally {
    loading.value = false
  }
}

onMounted(load)

function onFile(e: Event) {
  file.value = (e.target as HTMLInputElement).files?.[0] ?? null
}
function onLecFile(e: Event) {
  lecFile.value = (e.target as HTMLInputElement).files?.[0] ?? null
}

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.title) { error.value = 'El titulo es requerido'; return }
  loading.value = true
  try {
    const fd = new FormData()
    fd.append('title', form.value.title)
    fd.append('description', form.value.description)
    fd.append('type', form.value.type)
    fd.append('content', form.value.content)
    fd.append('is_public', String(form.value.is_public))
    fd.append('welcome_message', form.value.welcome_message)
    if (file.value) fd.append('file', file.value)
    if (thumbnailFile.value) fd.append('thumbnail', thumbnailFile.value)
    
    await api.post('/instructor/capacitaciones', fd)
    success.value = 'Curso creado'
    showForm.value = false
    form.value = { title: '', description: '', type: 'video', content: '', is_public: false, welcome_message: '', thumbnail_url: '' }
    file.value = null
    thumbnailFile.value = null
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('Eliminar este curso?')) return
  await api.delete(`/instructor/capacitaciones/${id}`)
  await load()
}

async function togglePublic(id: string) {
  await api.patch(`/instructor/capacitaciones/${id}/toggle-public`)
  await load()
}

async function resetCodigo(id: string) {
  await api.post(`/instructor/capacitaciones/${id}/reset-codigo`)
  await load()
}

async function selectCurso(c: any) {
  if (selectedCurso.value?.id === c.id) { selectedCurso.value = null; return }
  selectedCurso.value = c
  activeTab.value = 'lecciones'
  await Promise.all([loadLecciones(), loadIntermedias(), loadMisExamenes()])
}

async function loadLecciones() {
  if (!selectedCurso.value) return
  loadingLec.value = true
  const res = await api.get(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones`)
  lecciones.value = res.data || []
  loadingLec.value = false
}

async function guardarLeccion() {
  if (!lecForm.value.title) return
  const fd = new FormData()
  fd.append('title', lecForm.value.title)
  fd.append('description', lecForm.value.description)
  fd.append('type', lecForm.value.type)
  fd.append('content', lecForm.value.content)
  fd.append('orden', String(lecForm.value.orden))
  fd.append('duracion_min', String(lecForm.value.duracion_min || 0))
  if (lecFile.value) fd.append('file', lecFile.value)
  await api.post(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones`, fd)
  showLecForm.value = false
  lecForm.value = { title: '', description: '', type: 'video', content: '', orden: lecciones.value.length + 2, duracion_min: 0 }
  lecFile.value = null
  await loadLecciones()
}

async function eliminarLeccion(leccionId: string) {
  if (!confirm('Eliminar esta leccion?')) return
  await api.delete(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones/${leccionId}`)
  await loadLecciones()
}

async function moverLeccion(idx: number, dir: -1 | 1) {
  const arr = [...lecciones.value]
  const target = idx + dir
  if (target < 0 || target >= arr.length) return
  ;[arr[idx], arr[target]] = [arr[target], arr[idx]]
  const reorder = arr.map((l, i) => ({ id: l.id, orden: i + 1 }))
  await api.put(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones/reorder`, reorder)
  await loadLecciones()
}

async function loadIntermedias() {
  if (!selectedCurso.value) return
  loadingInt.value = true
  const res = await api.get(`/instructor/capacitaciones/${selectedCurso.value.id}/intermedias`)
  intermedias.value = res.data || []
  loadingInt.value = false
}

function addOpcion() { intForm.value.opciones.push({ texto: '', es_correcta: false }) }
function removeOpcion(i: number) { intForm.value.opciones.splice(i, 1) }

async function guardarIntermedia() {
  if (!intForm.value.texto) return
  const payload: any = { texto: intForm.value.texto, tipo: intForm.value.tipo, orden: intForm.value.orden }
  if (intForm.value.despues_de_leccion_id) payload.despues_de_leccion_id = intForm.value.despues_de_leccion_id
  if (intForm.value.tipo !== 'open_text') payload.opciones = intForm.value.tipo === 'true_false'
    ? [{ texto: 'Verdadero', es_correcta: false }, { texto: 'Falso', es_correcta: false }]
    : intForm.value.opciones
  await api.post(`/instructor/capacitaciones/${selectedCurso.value.id}/intermedias`, payload)
  showIntForm.value = false
  intForm.value = { texto: '', tipo: 'multiple_choice', orden: 1, despues_de_leccion_id: '', opciones: [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }] }
  await loadIntermedias()
}

async function eliminarIntermedia(preguntaId: string) {
  if (!confirm('Eliminar esta pregunta?')) return
  await api.delete(`/instructor/capacitaciones/${selectedCurso.value.id}/intermedias/${preguntaId}`)
  await loadIntermedias()
}

async function loadMisExamenes() {
  const res = await api.get('/instructor/examenes')
  misExamenes.value = res.data || []
}

function fileUrl(path: string) {
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
}
</script>

<template>
  <div class="ci-shell">
    <!-- Page Header estilo Udemy -->
    <div class="ci-topbar">
      <div class="ci-topbar-left">
        <h1 class="ci-topbar-title">Mis Cursos</h1>
        <p class="ci-topbar-sub">{{ capacitaciones.length }} curso{{ capacitaciones.length !== 1 ? 's' : '' }} publicados</p>
      </div>
      <button class="ci-new-btn" @click="showForm = !showForm" :aria-expanded="showForm">
        <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
          <path v-if="!showForm" d="M12 5v14M5 12h14" stroke-linecap="round"/>
          <path v-else d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/>
        </svg>
        {{ showForm ? 'Cancelar' : 'Crear nuevo curso' }}
      </button>
    </div>

    <!-- Alertas globales -->
    <Transition name="slide-down">
      <div v-if="error" class="ci-alert ci-alert-error" role="alert">
        <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/></svg>
        {{ error }}
      </div>
    </Transition>
    <Transition name="slide-down">
      <div v-if="success" class="ci-alert ci-alert-success" role="status">
        <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7" stroke-linecap="round"/></svg>
        {{ success }}
      </div>
    </Transition>

    <!-- Formulario nuevo curso estilo Udemy -->
    <Transition name="slide-down">
      <div v-if="showForm" class="ci-form-card">
        <div class="ci-form-header">
          <div class="ci-form-header-icon">📚</div>
          <div>
            <h2 class="ci-form-title">Crear nuevo curso</h2>
            <p class="ci-form-sub">Completa la información básica para empezar</p>
          </div>
        </div>

        <div class="ci-form-section">
          <h3 class="ci-form-section-title">Información del curso</h3>
          <div class="ci-form-grid">
            <div class="ci-field ci-field-full">
              <label class="ci-label" for="f-title">Título del curso <span class="ci-req">*</span></label>
              <input id="f-title" class="ci-input" v-model="form.title" placeholder="Ej: Seguridad Industrial – Nivel Básico" />
              <span class="ci-hint">Un buen título es claro y describe lo que el estudiante aprenderá.</span>
            </div>
            <div class="ci-field ci-field-full">
              <label class="ci-label" for="f-desc">Descripción del curso</label>
              <textarea id="f-desc" class="ci-input ci-textarea" v-model="form.description" rows="3" placeholder="¿Qué aprenderán los estudiantes? ¿A quién va dirigido?" style="resize:vertical" />
            </div>
          </div>
        </div>

        <div class="ci-form-section">
          <h3 class="ci-form-section-title">Diseño y Bienvenida</h3>
          <div class="ci-form-grid">
            <div class="ci-field ci-field-full">
              <label class="ci-label">Imagen de portada <span class="ci-optional">(opcional)</span></label>
              <DragDropUpload 
                v-model="thumbnailFile" 
                accept=".jpg,.jpeg,.png,.webp" 
              />
              <span class="ci-hint">Sube una imagen que representará el curso (Recomendado: 1280x720px. Formatos: JPG, PNG, WEBP)</span>
            </div>
            <div class="ci-field ci-field-full">
              <label class="ci-label" for="f-welcome">Mensaje de Bienvenida <span class="ci-optional">(opcional)</span></label>
              <textarea id="f-welcome" class="ci-input ci-textarea" v-model="form.welcome_message" rows="3" placeholder="Mensaje visible para los estudiantes al entrar al curso por primera vez" style="resize:vertical" />
            </div>
          </div>
        </div>

        <div class="ci-form-section">
          <h3 class="ci-form-section-title">Configuración</h3>
          <div class="ci-form-grid">
            <div class="ci-field">
              <label class="ci-label" for="f-type">Tipo de contenido principal</label>
              <div class="ci-select-wrap">
                <select id="f-type" class="ci-input ci-select" v-model="form.type">
                  <option value="video">🎥 Video</option>
                  <option value="document">📄 Documento PDF</option>
                  <option value="text">📝 Texto / Lectura</option>
                </select>
              </div>
            </div>
            <div class="ci-field">
              <label class="ci-label" for="f-file">Archivo principal <span class="ci-optional">(opcional)</span></label>
              <div class="ci-file-drop">
                <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 16l4-4m0 0l4 4m-4-4v9"/><path d="M20 16.58A5 5 0 0018 7h-1.26A8 8 0 104 15.25"/></svg>
                <span>Arrastra o elige un archivo</span>
                <input id="f-file" type="file" @change="onFile" :accept="form.type === 'video' ? 'video/*' : '.pdf,.doc,.docx'" />
              </div>
            </div>
            <div v-if="form.type === 'text'" class="ci-field ci-field-full">
              <label class="ci-label">Contenido del texto</label>
              <textarea class="ci-input ci-textarea" v-model="form.content" rows="5" placeholder="Escribe el contenido del curso..." style="resize:vertical" />
            </div>
            <div class="ci-field ci-field-full">
              <label class="ci-toggle-label">
                <div class="ci-toggle" :class="{ active: form.is_public }" @click="form.is_public = !form.is_public" role="switch" :aria-checked="form.is_public" tabindex="0" @keydown.enter="form.is_public = !form.is_public">
                  <div class="ci-toggle-thumb"></div>
                </div>
                <div>
                  <span class="ci-toggle-text">Curso público</span>
                  <span class="ci-toggle-hint">{{ form.is_public ? 'Visible sin necesidad de código de acceso' : 'Solo accesible con código de acceso' }}</span>
                </div>
              </label>
            </div>
          </div>
        </div>

        <div class="ci-form-footer">
          <button class="ci-btn-primary" @click="guardar" :disabled="loading" :aria-busy="loading">
            <svg v-if="!loading" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7" stroke-linecap="round"/></svg>
            <span v-if="loading" class="ci-spinner"></span>
            {{ loading ? 'Guardando...' : 'Publicar curso' }}
          </button>
          <button class="ci-btn-ghost" @click="showForm = false" type="button">Cancelar</button>
        </div>
      </div>
    </Transition>

    <!-- Grid de cursos estilo Udemy -->
    <div class="ci-grid">
      <!-- Skeleton loading -->
      <template v-if="!capacitaciones.length && loading">
        <div v-for="n in 3" :key="n" class="ci-card ci-card-skeleton">
          <div class="ci-card-cover skeleton"></div>
          <div class="ci-card-body">
            <div class="skeleton skel-title" style="width:75%"></div>
            <div class="skeleton skel-text" style="margin-top:8px"></div>
            <div class="skeleton skel-text-sm" style="margin-top:4px;width:50%"></div>
          </div>
        </div>
      </template>

      <TransitionGroup name="list-item" tag="div" style="display:contents">
        <div v-for="c in capacitaciones" :key="c.id"
          :class="['ci-card', selectedCurso?.id === c.id && 'ci-card-selected']">

          <!-- Cover / Thumbnail -->
          <div :class="['ci-card-cover', c.thumbnail_url ? 'has-image' : `cover-${c.type}`]">
            <template v-if="c.thumbnail_url">
              <img :src="fileUrl(c.thumbnail_url)" alt="Portada del curso" class="ci-card-cover-img" />
            </template>
            <template v-else>
              <div class="ci-card-cover-icon">{{ c.type === 'video' ? '🎥' : c.type === 'document' ? '📄' : '📝' }}</div>
            </template>
            <div class="ci-card-cover-badges">
              <span :class="['ci-badge-pub', c.is_public ? 'pub' : 'priv']">
                {{ c.is_public ? '🌐 Público' : '🔒 Privado' }}
              </span>
            </div>
            <div class="ci-card-cover-overlay">
              <button class="ci-overlay-btn" @click="selectCurso(c)">
                {{ selectedCurso?.id === c.id ? 'Cerrar' : 'Gestionar contenido' }}
              </button>
            </div>
          </div>

          <!-- Card Body -->
          <div class="ci-card-body">
            <h3 class="ci-card-title">{{ c.title }}</h3>
            <p class="ci-card-desc">{{ c.description || 'Sin descripción' }}</p>
            <div class="ci-card-stats">
              <span class="ci-stat">
                <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6h16M4 10h16M4 14h8"/></svg>
                {{ c.type === 'video' ? 'Video' : c.type === 'document' ? 'Documento' : 'Texto' }}
              </span>
              <span v-if="c.codigo" class="ci-stat ci-stat-code">
                <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="5" y="11" width="14" height="10" rx="2"/><path d="M12 17v.01M8 11V7a4 4 0 018 0v4"/></svg>
                {{ c.codigo }}
              </span>
            </div>
          </div>

          <!-- Card Footer Actions -->
          <div class="ci-card-footer">
            <button class="ci-action-btn" :title="c.is_public ? 'Ocultar curso' : 'Publicar curso'" @click="togglePublic(c.id)">
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <template v-if="c.is_public">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                </template>
                <template v-else>
                  <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94"/><path d="M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/>
                </template>
              </svg>
              {{ c.is_public ? 'Ocultar' : 'Publicar' }}
            </button>
            <button class="ci-action-btn" title="Regenerar código" @click="resetCodigo(c.id)">
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 11-2.12-9.36L23 10"/></svg>
              Nuevo código
            </button>
            <button class="ci-action-btn ci-action-danger" title="Eliminar curso" @click="eliminar(c.id)">
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/></svg>
              Eliminar
            </button>
          </div>
        </div>
      </TransitionGroup>
    </div>

    <!-- Empty state -->
    <Transition name="fade">
      <div v-if="capacitaciones.length === 0 && !loading" class="ci-empty">
        <div class="ci-empty-illustration">
          <svg width="64" height="64" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color:var(--border)"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
        </div>
        <h3 class="ci-empty-title">Aún no tienes cursos</h3>
        <p class="ci-empty-sub">Crea tu primer curso y empieza a capacitar a tus estudiantes.</p>
        <button class="ci-btn-primary" @click="showForm = true">Crear mi primer curso</button>
      </div>
    </Transition>

    <!-- Panel de gestión del curso seleccionado (full width, debajo del grid) -->
    <Transition name="slide-down">
      <div v-if="selectedCurso" class="ci-mgmt">
        <!-- Header del panel -->
        <div class="ci-mgmt-head">
          <div class="ci-mgmt-head-left">
            <div :class="['ci-mgmt-thumb', `cover-${selectedCurso.type}`]">
              {{ selectedCurso.type === 'video' ? '🎥' : selectedCurso.type === 'document' ? '📄' : '📝' }}
            </div>
            <div>
              <h2 class="ci-mgmt-title">{{ selectedCurso.title }}</h2>
              <p class="ci-mgmt-sub">Gestionar contenido del curso</p>
            </div>
          </div>
          <button class="ci-mgmt-close" @click="selectedCurso = null" aria-label="Cerrar panel">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/></svg>
          </button>
        </div>

        <!-- Tabs internos -->
        <div class="ci-tabs" role="tablist">
          <button role="tab" :aria-selected="activeTab === 'lecciones'"
            :class="['ci-tab', activeTab === 'lecciones' && 'ci-tab-active']"
            @click="activeTab = 'lecciones'">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6h16M4 10h16M4 14h8"/></svg>
            Lecciones
            <span class="ci-tab-pill">{{ lecciones.length }}</span>
          </button>
          <button role="tab" :aria-selected="activeTab === 'intermedias'"
            :class="['ci-tab', activeTab === 'intermedias' && 'ci-tab-active']"
            @click="activeTab = 'intermedias'">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01"/><circle cx="12" cy="12" r="10"/></svg>
            Preguntas intermedias
            <span class="ci-tab-pill">{{ intermedias.length }}</span>
          </button>
          <button role="tab" :aria-selected="activeTab === 'examen'"
            :class="['ci-tab', activeTab === 'examen' && 'ci-tab-active']"
            @click="activeTab = 'examen'">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
            Examen final
          </button>
        </div>

        <div class="ci-tab-body">

          <!-- ── TAB LECCIONES ── -->
          <div v-if="activeTab === 'lecciones'" role="tabpanel">
            <div class="ci-tab-toolbar">
              <span class="ci-tab-count">{{ lecciones.length }} lección{{ lecciones.length !== 1 ? 'es' : '' }}</span>
              <button class="ci-btn-primary ci-btn-sm" @click="showLecForm = !showLecForm; lecForm.orden = lecciones.length + 1">
                <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14" stroke-linecap="round"/></svg>
                {{ showLecForm ? 'Cancelar' : 'Agregar lección' }}
              </button>
            </div>

            <!-- Form nueva lección -->
            <Transition name="slide-down">
              <div v-if="showLecForm" class="ci-sub-form">
                <div class="ci-sub-form-header">
                  <span>📖 Nueva lección</span>
                </div>
                <div class="ci-form-grid">
                  <div class="ci-field ci-field-full">
                    <label class="ci-label">Título <span class="ci-req">*</span></label>
                    <input class="ci-input" v-model="lecForm.title" placeholder="Ej: Introducción al módulo" />
                  </div>
                  <div class="ci-field ci-field-full">
                    <label class="ci-label">Descripción <span class="ci-optional">(opcional)</span></label>
                    <input class="ci-input" v-model="lecForm.description" placeholder="Breve descripción de la lección" />
                  </div>
                  <div class="ci-field">
                    <label class="ci-label">Tipo de contenido</label>
                    <div class="ci-select-wrap">
                      <select class="ci-input ci-select" v-model="lecForm.type">
                        <option value="video">🎥 Video</option>
                        <option value="document">📄 PDF / Documento</option>
                        <option value="text">📝 Texto</option>
                        <option value="link">🔗 Enlace externo</option>
                      </select>
                    </div>
                  </div>
                  <div class="ci-field">
                    <label class="ci-label">Duración (min)</label>
                    <input type="number" class="ci-input" v-model="lecForm.duracion_min" min="0" placeholder="0" />
                  </div>
                  <div v-if="lecForm.type === 'video' || lecForm.type === 'document'" class="ci-field ci-field-full">
                    <label>Archivo *</label>
                    <DragDropUpload 
                      v-model="lecFile" 
                      :accept="lecForm.type === 'video' ? 'video/mp4,video/webm' : '.pdf,.doc,.docx'" 
                    />
                  </div>
                  <div v-if="lecForm.type === 'text'" class="ci-field ci-field-full">
                    <label>Archivo (no aplica)</label>
                    <p style="font-size:0.82rem;color:var(--muted);padding:10px 0;background:var(--surface-soft);border-radius:var(--r);text-align:center">El contenido se escribe abajo</p>
                  </div>
                  <div v-if="lecForm.type === 'text'" class="ci-field ci-field-full">
                    <label class="ci-label">Contenido</label>
                    <textarea class="ci-input ci-textarea" v-model="lecForm.content" placeholder="Escribe el contenido..." rows="5" style="resize:vertical" />
                  </div>
                  <div v-if="lecForm.type === 'link'" class="ci-field ci-field-full">
                    <label class="ci-label">URL del recurso</label>
                    <input class="ci-input" v-model="lecForm.content" placeholder="https://www.youtube.com/watch?v=..." />
                    <span class="ci-hint">Soporta YouTube, Vimeo y cualquier URL embebible.</span>
                  </div>
                </div>
                <div class="ci-sub-form-footer">
                  <button class="ci-btn-primary ci-btn-sm" @click="guardarLeccion">Guardar lección</button>
                  <button class="ci-btn-ghost ci-btn-sm" @click="showLecForm = false">Cancelar</button>
                </div>
              </div>
            </Transition>

            <!-- Skeleton lecciones -->
            <div v-if="loadingLec" class="ci-lec-list">
              <div v-for="n in 3" :key="n" class="ci-lec-item">
                <div class="skeleton" style="width:28px;height:28px;border-radius:50%;flex-shrink:0"></div>
                <div style="flex:1">
                  <div class="skeleton skel-line" style="width:55%"></div>
                  <div class="skeleton skel-text-sm" style="margin-top:5px;width:35%"></div>
                </div>
              </div>
            </div>

            <!-- Empty lecciones -->
            <div v-else-if="lecciones.length === 0" class="ci-tab-empty">
              <span style="font-size:2.2rem">📋</span>
              <p>Sin lecciones aún. Agrega la primera con el botón de arriba.</p>
            </div>

            <!-- Lista de lecciones -->
            <TransitionGroup v-else name="list-item" tag="div" class="ci-lec-list">
              <div v-for="(lec, idx) in lecciones" :key="lec.id" class="ci-lec-item">
                <div class="ci-lec-num">{{ idx + 1 }}</div>
                <div :class="['ci-lec-pip', lec.type]"></div>
                <div class="ci-lec-info">
                  <span class="ci-lec-title">{{ lec.title }}</span>
                  <span class="ci-lec-meta">
                    {{ lec.type === 'video' ? '🎥 Video' : lec.type === 'document' ? '📄 PDF' : lec.type === 'link' ? '🔗 Enlace' : '📝 Texto' }}
                    <template v-if="lec.duracion_min"> · {{ lec.duracion_min }} min</template>
                    <template v-if="lec.description"> · {{ lec.description }}</template>
                  </span>
                </div>
                <div class="ci-lec-actions">
                  <button class="ci-lec-btn" @click="moverLeccion(idx, -1)" :disabled="idx === 0" aria-label="Subir">↑</button>
                  <button class="ci-lec-btn" @click="moverLeccion(idx, 1)" :disabled="idx === lecciones.length - 1" aria-label="Bajar">↓</button>
                  <button class="ci-lec-btn ci-lec-btn-del" @click="eliminarLeccion(lec.id)" aria-label="Eliminar">✕</button>
                </div>
              </div>
            </TransitionGroup>
          </div>

          <!-- ── TAB PREGUNTAS INTERMEDIAS ── -->
          <div v-if="activeTab === 'intermedias'" role="tabpanel">
            <div class="ci-tab-toolbar">
              <span class="ci-tab-count">{{ intermedias.length }} pregunta{{ intermedias.length !== 1 ? 's' : '' }}</span>
              <button class="ci-btn-primary ci-btn-sm" @click="showIntForm = !showIntForm">
                <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14" stroke-linecap="round"/></svg>
                {{ showIntForm ? 'Cancelar' : 'Agregar pregunta' }}
              </button>
            </div>

            <Transition name="slide-down">
              <div v-if="showIntForm" class="ci-sub-form">
                <div class="ci-sub-form-header"><span>💬 Nueva pregunta intermedia</span></div>
                <div style="display:flex;flex-direction:column;gap:12px">
                  <div class="ci-field ci-field-full">
                    <label class="ci-label">Texto de la pregunta <span class="ci-req">*</span></label>
                    <textarea class="ci-input ci-textarea" v-model="intForm.texto" placeholder="¿Cuál es el procedimiento correcto para...?" rows="2" style="resize:vertical" />
                  </div>
                  <div class="ci-form-grid">
                    <div class="ci-field">
                      <label class="ci-label">Tipo de pregunta</label>
                      <div class="ci-select-wrap">
                        <select class="ci-input ci-select" v-model="intForm.tipo">
                          <option value="multiple_choice">Opción múltiple</option>
                          <option value="true_false">Verdadero / Falso</option>
                          <option value="open_text">Respuesta abierta</option>
                        </select>
                      </div>
                    </div>
                    <div class="ci-field">
                      <label class="ci-label">Mostrar después de</label>
                      <div class="ci-select-wrap">
                        <select class="ci-input ci-select" v-model="intForm.despues_de_leccion_id">
                          <option value="">Al inicio del curso</option>
                          <option v-for="lec in lecciones" :key="lec.id" :value="lec.id">{{ lec.title }}</option>
                        </select>
                      </div>
                    </div>
                  </div>

                  <div v-if="intForm.tipo === 'multiple_choice'">
                    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
                      <span class="ci-label" style="margin:0">Opciones — marca la correcta</span>
                      <button class="ci-btn-ghost ci-btn-sm" @click="addOpcion" type="button">+ Opción</button>
                    </div>
                    <div v-for="(op, i) in intForm.opciones" :key="i" class="ci-opcion-row" :class="{ correct: op.es_correcta }">
                      <input class="ci-input" v-model="op.texto" :placeholder="`Opción ${i + 1}`" style="flex:1" />
                      <label class="ci-radio-opt">
                        <input type="radio" :name="`int-c-${selectedCurso.id}`" :checked="op.es_correcta"
                          @change="intForm.opciones.forEach((o, j) => o.es_correcta = j === i)"
                          style="accent-color:var(--success)" />
                        <span>Correcta</span>
                      </label>
                      <button v-if="intForm.opciones.length > 2" class="ci-lec-btn ci-lec-btn-del" @click="removeOpcion(i)" type="button">✕</button>
                    </div>
                  </div>

                  <div v-if="intForm.tipo === 'true_false'" class="ci-info-note">
                    Se generarán las opciones <strong>Verdadero</strong> y <strong>Falso</strong> automáticamente.
                  </div>
                </div>
                <div class="ci-sub-form-footer">
                  <button class="ci-btn-primary ci-btn-sm" @click="guardarIntermedia">Guardar pregunta</button>
                  <button class="ci-btn-ghost ci-btn-sm" @click="showIntForm = false">Cancelar</button>
                </div>
              </div>
            </Transition>

            <div v-if="loadingInt" class="ci-lec-list">
              <div v-for="n in 2" :key="n" class="ci-lec-item">
                <div class="skeleton" style="width:70px;height:24px;border-radius:4px;flex-shrink:0"></div>
                <div style="flex:1"><div class="skeleton skel-line" style="width:75%"></div></div>
              </div>
            </div>

            <div v-else-if="intermedias.length === 0" class="ci-tab-empty">
              <span style="font-size:2.2rem">💡</span>
              <p>Sin preguntas intermedias. Aparecen entre lecciones para reforzar el aprendizaje.</p>
            </div>

            <TransitionGroup v-else name="list-item" tag="div" class="ci-lec-list">
              <div v-for="preg in intermedias" :key="preg.id" class="ci-lec-item">
                <span :class="['ci-int-badge', preg.tipo === 'multiple_choice' ? 'mc' : preg.tipo === 'true_false' ? 'tf' : 'ot']">
                  {{ preg.tipo === 'multiple_choice' ? 'Múltiple' : preg.tipo === 'true_false' ? 'V/F' : 'Abierta' }}
                </span>
                <div class="ci-lec-info">
                  <span class="ci-lec-title">{{ preg.texto }}</span>
                </div>
                <button class="ci-lec-btn ci-lec-btn-del" @click="eliminarIntermedia(preg.id)" aria-label="Eliminar pregunta">✕</button>
              </div>
            </TransitionGroup>
          </div>

          <!-- ── TAB EXAMEN ── -->
          <div v-if="activeTab === 'examen'" role="tabpanel">
            <div class="ci-info-note" style="margin-bottom:16px">
              Para enlazar un examen, selecciona <strong>{{ selectedCurso.title }}</strong> al crear el examen en la sección "Exámenes".
            </div>

            <div v-if="misExamenes.filter(e => e.capacitacion_id === selectedCurso.id).length === 0" class="ci-tab-empty">
              <span style="font-size:2.2rem">📋</span>
              <p>Sin examen enlazado. Ve a la sección Exámenes y selecciona este curso.</p>
            </div>

            <div v-for="ex in misExamenes.filter(e => e.capacitacion_id === selectedCurso.id)" :key="ex.id" class="ci-lec-item" style="margin-bottom:8px">
              <div class="ci-lec-pip" style="background:var(--success)"></div>
              <div class="ci-lec-info">
                <span class="ci-lec-title">{{ ex.title }}</span>
                <span class="ci-lec-meta"><span class="ci-int-badge mc" style="font-size:0.75rem">✓ Enlazado</span></span>
              </div>
            </div>
          </div>

        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
/* ── Shell ── */
.ci-shell { min-height: 100vh; background: var(--bg); padding-bottom: 60px; }

/* ── Topbar ── */
.ci-topbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 28px 32px 20px; gap: 16px; flex-wrap: wrap;
}
.ci-topbar-title { font-size: 1.75rem; font-weight: 800; color: var(--dark); letter-spacing: -0.03em; margin: 0; }
.ci-topbar-sub { font-size: 0.88rem; color: var(--muted); margin: 3px 0 0; }

/* ── New Course Button ── */
.ci-new-btn {
  display: inline-flex; align-items: center; gap: 8px;
  background: var(--dark); color: #fff;
  padding: 11px 20px; border-radius: 6px; border: none; cursor: pointer;
  font-size: 0.9rem; font-weight: 700; letter-spacing: 0.01em;
  transition: background 0.15s, transform 0.1s; white-space: nowrap;
}
.ci-new-btn:hover { background: #000; transform: translateY(-1px); }
.ci-new-btn:active { transform: translateY(0); }

/* ── Alerts ── */
.ci-alert {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 20px; margin: 0 32px 16px; border-radius: 6px; font-size: 0.88rem; font-weight: 600;
}
.ci-alert-error   { background: var(--danger-bg); color: var(--danger); border-left: 3px solid var(--danger); }
.ci-alert-success { background: var(--success-bg); color: var(--success); border-left: 3px solid var(--success); }

/* ── Create Course Form ── */
.ci-form-card {
  margin: 0 32px 28px; background: var(--surface); border-radius: 12px;
  border: 1.5px solid var(--border); box-shadow: 0 4px 20px rgba(0,0,0,.06); overflow: hidden;
}
.ci-form-header {
  display: flex; align-items: center; gap: 16px;
  padding: 24px 28px; background: var(--dark); color: #fff;
}
.ci-form-header-icon { font-size: 2rem; }
.ci-form-title { font-size: 1.1rem; font-weight: 800; color: #fff; margin: 0 0 2px; }
.ci-form-sub   { font-size: 0.83rem; color: rgba(255,255,255,.65); margin: 0; }

.ci-form-section { padding: 22px 28px; border-bottom: 1px solid var(--border-light); }
.ci-form-section:last-of-type { border-bottom: none; }
.ci-form-section-title { font-size: 0.8rem; font-weight: 700; text-transform: uppercase; letter-spacing: .07em; color: var(--muted); margin: 0 0 16px; }

.ci-form-footer {
  padding: 18px 28px; display: flex; align-items: center; gap: 10px;
  background: var(--bg); border-top: 1px solid var(--border);
}

/* ── Fields ── */
.ci-form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.ci-field { display: flex; flex-direction: column; gap: 6px; }
.ci-field-full { grid-column: 1 / -1; }
.ci-label { font-size: 0.83rem; font-weight: 700; color: var(--dark); }
.ci-req { color: var(--danger); margin-left: 2px; }
.ci-optional { font-weight: 400; color: var(--muted); }
.ci-hint { font-size: 0.77rem; color: var(--muted); }

.ci-input {
  width: 100%; padding: 10px 13px; border: 1.5px solid var(--border);
  border-radius: 6px; font-size: 0.9rem; color: var(--dark); background: var(--surface);
  transition: border-color 0.15s, box-shadow 0.15s; outline: none; box-sizing: border-box;
}
.ci-input:focus { border-color: var(--brand); box-shadow: 0 0 0 3px var(--brand-light); }
.ci-textarea { resize: vertical; min-height: 80px; font-family: inherit; }
.ci-select { appearance: none; padding-right: 32px; }
.ci-select-wrap { position: relative; }
.ci-select-wrap::after {
  content: ''; position: absolute; right: 12px; top: 50%; transform: translateY(-50%);
  width: 0; height: 0; border-left: 4px solid transparent; border-right: 4px solid transparent;
  border-top: 5px solid var(--muted); pointer-events: none;
}

.ci-file-drop {
  display: flex; align-items: center; gap: 10px; padding: 12px 14px;
  border: 1.5px dashed var(--border); border-radius: 6px; cursor: pointer;
  background: var(--bg); color: var(--muted); font-size: 0.85rem; position: relative;
  transition: border-color 0.15s, background 0.15s;
}
.ci-file-drop:hover { border-color: var(--brand); background: var(--brand-light); }
.ci-file-drop input[type="file"] { position: absolute; inset: 0; opacity: 0; cursor: pointer; width: 100%; }

/* Toggle */
.ci-toggle-label { display: flex; align-items: flex-start; gap: 14px; cursor: pointer; }
.ci-toggle {
  width: 44px; height: 24px; background: var(--border); border-radius: 12px;
  position: relative; flex-shrink: 0; transition: background 0.2s; margin-top: 2px; cursor: pointer;
}
.ci-toggle.active { background: var(--brand); }
.ci-toggle-thumb {
  width: 18px; height: 18px; background: #fff; border-radius: 50%;
  position: absolute; top: 3px; left: 3px; transition: left 0.2s; box-shadow: 0 1px 3px rgba(0,0,0,.2);
}
.ci-toggle.active .ci-toggle-thumb { left: 23px; }
.ci-toggle-text { font-size: 0.88rem; font-weight: 700; color: var(--dark); display: block; }
.ci-toggle-hint { font-size: 0.78rem; color: var(--muted); margin-top: 1px; display: block; }

/* ── Buttons ── */
.ci-btn-primary {
  display: inline-flex; align-items: center; gap: 7px;
  background: var(--brand); color: #fff; padding: 11px 22px;
  border: none; border-radius: 6px; cursor: pointer; font-size: 0.9rem; font-weight: 700;
  transition: background 0.15s, transform 0.1s;
}
.ci-btn-primary:hover:not(:disabled) { background: var(--brand-dark); transform: translateY(-1px); }
.ci-btn-primary:disabled { opacity: .55; cursor: not-allowed; }
.ci-btn-ghost {
  display: inline-flex; align-items: center; gap: 7px;
  background: none; color: var(--dark); padding: 10px 18px;
  border: 1.5px solid var(--border); border-radius: 6px; cursor: pointer; font-size: 0.9rem; font-weight: 600;
  transition: background 0.12s, border-color 0.12s;
}
.ci-btn-ghost:hover { background: var(--bg); border-color: var(--dark); }
.ci-btn-sm { padding: 7px 14px !important; font-size: 0.82rem !important; }

/* Spinner */
.ci-spinner {
  width: 15px; height: 15px; border: 2px solid rgba(255,255,255,.3);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; display: inline-block;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Course Grid ── */
.ci-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px; padding: 4px 32px 0;
}

.ci-card {
  background: var(--surface); border-radius: 10px;
  border: 1.5px solid var(--border); overflow: hidden;
  display: flex; flex-direction: column;
  transition: box-shadow 0.2s, transform 0.2s, border-color 0.2s;
}
.ci-card:hover { box-shadow: 0 8px 28px rgba(0,0,0,.1); transform: translateY(-3px); }
.ci-card-selected { border-color: var(--brand); box-shadow: 0 0 0 3px var(--brand-light); transform: none; }
.ci-card-skeleton .ci-card-cover { height: 160px; }

/* Card cover */
.ci-card-cover {
  height: 155px; position: relative; display: flex; align-items: center; justify-content: center;
  overflow: hidden;
}
.ci-card-cover.has-image { background: none; }
.ci-card-cover-img { width: 100%; height: 100%; object-fit: cover; }
.cover-video    { background: linear-gradient(135deg, #1e1b4b 0%, #312e81 40%, #4338ca 100%); }
.cover-document { background: linear-gradient(135deg, #14532d 0%, #15803d 40%, #16a34a 100%); }
.cover-text     { background: linear-gradient(135deg, #1e3a5f 0%, #1d4ed8 40%, #3b82f6 100%); }

.ci-card-cover-icon {
  font-size: 3rem; z-index: 1; filter: drop-shadow(0 4px 8px rgba(0,0,0,.3));
  transition: transform 0.2s;
}
.ci-card:hover .ci-card-cover-icon { transform: scale(1.15); }

.ci-card-cover-badges {
  position: absolute; top: 10px; right: 10px; z-index: 2;
}
.ci-badge-pub {
  font-size: 0.72rem; font-weight: 700; padding: 3px 10px; border-radius: 20px;
  backdrop-filter: blur(6px);
}
.ci-badge-pub.pub  { background: rgba(22,163,74,.85); color: #fff; }
.ci-badge-pub.priv { background: rgba(0,0,0,.55); color: rgba(255,255,255,.9); }

.ci-card-cover-overlay {
  position: absolute; inset: 0; background: rgba(0,0,0,.55);
  display: flex; align-items: center; justify-content: center;
  opacity: 0; transition: opacity 0.2s; z-index: 3;
}
.ci-card:hover .ci-card-cover-overlay { opacity: 1; }
.ci-overlay-btn {
  background: #fff; color: var(--dark); border: none; cursor: pointer;
  padding: 9px 18px; border-radius: 6px; font-size: 0.85rem; font-weight: 800;
  transition: transform 0.1s;
}
.ci-overlay-btn:hover { transform: scale(1.03); }

/* Card body */
.ci-card-body { padding: 14px 16px; flex: 1; }
.ci-card-title { font-size: 0.95rem; font-weight: 800; color: var(--dark); margin: 0 0 5px; line-height: 1.3; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.ci-card-desc { font-size: 0.8rem; color: var(--muted); margin: 0 0 10px; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.ci-card-stats { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.ci-stat { display: flex; align-items: center; gap: 4px; font-size: 0.77rem; color: var(--muted); }
.ci-stat-code { background: var(--bg); border: 1px solid var(--border); padding: 2px 8px; border-radius: 4px; font-family: monospace; font-size: 0.8rem; font-weight: 700; color: var(--dark); }

/* Card footer */
.ci-card-footer {
  display: flex; align-items: center; gap: 1px;
  border-top: 1px solid var(--border-light); background: var(--bg);
}
.ci-action-btn {
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 5px;
  padding: 9px 6px; border: none; background: none; cursor: pointer;
  font-size: 0.76rem; font-weight: 600; color: var(--muted);
  transition: background 0.12s, color 0.12s;
}
.ci-action-btn:hover { background: var(--border-light); color: var(--dark); }
.ci-action-btn:not(:last-child) { border-right: 1px solid var(--border-light); }
.ci-action-danger:hover { background: var(--danger-bg); color: var(--danger); }

/* ── Empty state ── */
.ci-empty {
  text-align: center; padding: 80px 20px; display: flex; flex-direction: column; align-items: center; gap: 12px;
}
.ci-empty-illustration { margin-bottom: 4px; }
.ci-empty-title { font-size: 1.25rem; font-weight: 800; color: var(--dark); margin: 0; }
.ci-empty-sub   { font-size: 0.88rem; color: var(--muted); margin: 0; max-width: 320px; }

/* ── Management Panel ── */
.ci-mgmt {
  margin: 24px 32px 0; background: var(--surface); border-radius: 12px;
  border: 1.5px solid var(--brand); box-shadow: 0 6px 28px rgba(249,115,22,.1); overflow: hidden;
}
.ci-mgmt-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px; background: var(--dark); gap: 14px;
}
.ci-mgmt-head-left { display: flex; align-items: center; gap: 14px; }
.ci-mgmt-thumb {
  width: 42px; height: 42px; border-radius: 8px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center; font-size: 1.4rem;
}
.ci-mgmt-title { font-size: 1.05rem; font-weight: 800; color: #fff; margin: 0; }
.ci-mgmt-sub   { font-size: 0.8rem; color: rgba(255,255,255,.55); margin: 2px 0 0; }
.ci-mgmt-close {
  background: rgba(255,255,255,.1); border: none; cursor: pointer; color: #fff;
  width: 34px; height: 34px; border-radius: 50%; display: flex; align-items: center; justify-content: center;
  transition: background 0.15s; flex-shrink: 0;
}
.ci-mgmt-close:hover { background: rgba(255,255,255,.2); }

/* Tabs */
.ci-tabs { display: flex; padding: 0 24px; border-bottom: 1.5px solid var(--border-light); gap: 0; }
.ci-tab {
  display: flex; align-items: center; gap: 7px; padding: 14px 16px;
  border: none; background: none; cursor: pointer; font-size: 0.88rem; font-weight: 600;
  color: var(--muted); border-bottom: 2.5px solid transparent; margin-bottom: -1.5px;
  transition: color 0.15s, border-color 0.15s; white-space: nowrap;
}
.ci-tab:hover { color: var(--dark); }
.ci-tab-active { color: var(--brand) !important; border-bottom-color: var(--brand); }
.ci-tab-pill {
  background: var(--border-light); color: var(--muted); font-size: 0.72rem; font-weight: 700;
  padding: 1px 7px; border-radius: 10px; min-width: 20px; text-align: center;
}
.ci-tab-active .ci-tab-pill { background: var(--brand-light); color: var(--brand-dark); }

.ci-tab-body { padding: 22px 24px; }

.ci-tab-toolbar {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;
}
.ci-tab-count { font-size: 0.85rem; font-weight: 700; color: var(--muted); }
.ci-tab-empty {
  text-align: center; padding: 36px 20px; display: flex; flex-direction: column;
  align-items: center; gap: 8px; color: var(--muted); font-size: 0.88rem;
}

/* Sub form inside panel */
.ci-sub-form {
  background: var(--bg); border-radius: 8px; border: 1.5px dashed var(--border);
  overflow: hidden; margin-bottom: 16px;
}
.ci-sub-form-header {
  padding: 12px 16px; background: var(--border-light); font-size: 0.88rem; font-weight: 700; color: var(--dark);
}
.ci-sub-form > .ci-form-grid, .ci-sub-form > div { padding: 16px; }
.ci-sub-form-footer {
  display: flex; gap: 8px; padding: 12px 16px;
  border-top: 1px solid var(--border-light); background: var(--surface);
}

/* Lesson list */
.ci-lec-list { display: flex; flex-direction: column; gap: 6px; position: relative; }
.ci-lec-item {
  display: flex; align-items: center; gap: 10px;
  background: var(--surface); border: 1.5px solid var(--border); border-radius: 8px; padding: 10px 12px;
  transition: border-color 0.12s, box-shadow 0.12s;
}
.ci-lec-item:hover { border-color: var(--brand-border); box-shadow: 0 2px 8px rgba(249,115,22,.08); }
.ci-lec-num {
  width: 24px; height: 24px; border-radius: 50%; background: var(--bg); border: 1.5px solid var(--border);
  color: var(--muted); font-size: 0.75rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.ci-lec-pip { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.ci-lec-pip.video    { background: var(--brand); }
.ci-lec-pip.document { background: var(--success); }
.ci-lec-pip.text     { background: var(--info); }
.ci-lec-pip.link     { background: #8b5cf6; }
.ci-lec-info { flex: 1; min-width: 0; }
.ci-lec-title { font-size: 0.88rem; font-weight: 700; color: var(--dark); display: block; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ci-lec-meta  { font-size: 0.77rem; color: var(--muted); display: flex; align-items: center; gap: 5px; margin-top: 2px; flex-wrap: wrap; }
.ci-lec-actions { display: flex; gap: 4px; flex-shrink: 0; }
.ci-lec-btn {
  width: 28px; height: 28px; border-radius: 6px; border: 1.5px solid var(--border);
  background: none; cursor: pointer; color: var(--muted); font-size: 0.82rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center; transition: all 0.12s;
}
.ci-lec-btn:hover:not(:disabled) { background: var(--bg); color: var(--dark); border-color: var(--dark); }
.ci-lec-btn:disabled { opacity: .3; cursor: not-allowed; }
.ci-lec-btn-del:hover { background: var(--danger-bg) !important; color: var(--danger) !important; border-color: var(--danger) !important; }

/* Intermediate question badges */
.ci-int-badge { font-size: 0.73rem; font-weight: 700; padding: 3px 10px; border-radius: 4px; flex-shrink: 0; white-space: nowrap; }
.ci-int-badge.mc  { background: #dbeafe; color: #1d4ed8; }
.ci-int-badge.tf  { background: #fef9c3; color: #a16207; }
.ci-int-badge.ot  { background: #d1fae5; color: #065f46; }

/* Option rows */
.ci-opcion-row {
  display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px;
  background: var(--surface); border: 1.5px solid var(--border); margin-bottom: 5px; transition: border-color 0.12s;
}
.ci-opcion-row.correct { border-color: var(--success); background: var(--success-bg); }
.ci-radio-opt { display: flex; align-items: center; gap: 5px; font-size: 0.8rem; color: var(--muted); cursor: pointer; white-space: nowrap; }

/* Info note */
.ci-info-note {
  background: var(--info-bg); color: var(--info); border-radius: 6px;
  padding: 10px 14px; font-size: 0.83rem; border-left: 3px solid var(--info);
}

@media (max-width: 900px) {
  .ci-topbar, .ci-alert, .ci-form-card, .ci-grid, .ci-mgmt { margin-left: 16px; margin-right: 16px; padding-left: 16px; padding-right: 16px; }
  .ci-grid { grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 14px; padding: 4px 0; margin: 0 16px; }
  .ci-form-grid { grid-template-columns: 1fr; }
  .ci-mgmt { margin-left: 0; margin-right: 0; border-radius: 0; }
}
@media (max-width: 600px) {
  .ci-grid { grid-template-columns: 1fr; }
  .ci-topbar { padding: 18px 16px 14px; }
  .ci-tabs { overflow-x: auto; }
  .ci-tab { padding: 12px 12px; font-size: 0.82rem; }
}
</style>
