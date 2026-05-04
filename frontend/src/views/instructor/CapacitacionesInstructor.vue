<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const loading = ref(false)
const showForm = ref(false)
const error = ref('')
const success = ref('')
const form = ref({ title: '', description: '', type: 'video', content: '', is_public: false })
const file = ref<File | null>(null)

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
  const res = await api.get('/instructor/capacitaciones')
  capacitaciones.value = res.data || []
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
    if (file.value) fd.append('file', file.value)
    await api.post('/instructor/capacitaciones', fd)
    success.value = 'Curso creado'
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
</script>

<template>
  <div>
    <!-- Page Header -->
    <div class="ph">
      <div>
        <h1 class="ph-title">Mis Cursos</h1>
        <p class="ph-sub">Gestiona el contenido y estructura de tus capacitaciones</p>
      </div>
      <button class="btn btn-primary" @click="showForm = !showForm" :aria-expanded="showForm">
        <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
        Nuevo Curso
      </button>
    </div>

    <!-- Alertas globales -->
    <Transition name="slide-down">
      <div v-if="error" class="alert alert-error" style="margin-bottom:16px" role="alert">{{ error }}</div>
    </Transition>
    <Transition name="slide-down">
      <div v-if="success" class="alert alert-success" style="margin-bottom:16px" role="status">{{ success }}</div>
    </Transition>

    <!-- Formulario nuevo curso -->
    <Transition name="slide-down">
      <div v-if="showForm" class="form-card" aria-label="Formulario nuevo curso">
        <p class="form-card-title">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="display:inline;vertical-align:middle;margin-right:6px"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
          Nuevo Curso
        </p>
        <div class="form-grid">
          <div class="field full">
            <label>Título del curso *</label>
            <input class="field-input" v-model="form.title" placeholder="Ej: Introducción a la Seguridad Industrial" />
          </div>
          <div class="field full">
            <label>Descripción</label>
            <textarea class="field-input" v-model="form.description" rows="2" placeholder="Breve descripción del curso y sus objetivos..." style="resize:vertical" />
          </div>
          <div class="field">
            <label>Tipo de curso</label>
            <select class="field-input" v-model="form.type">
              <option value="video">🎥 Video</option>
              <option value="document">📄 Documento PDF</option>
              <option value="text">📝 Texto / Lectura</option>
            </select>
          </div>
          <div class="field">
            <label>Archivo principal (opcional)</label>
            <input type="file" @change="onFile" class="field-input" style="padding:7px" :accept="form.type === 'video' ? 'video/*' : '.pdf,.doc,.docx'" />
          </div>
          <div v-if="form.type === 'text'" class="field full">
            <label>Contenido</label>
            <textarea class="field-input" v-model="form.content" rows="4" placeholder="Texto del contenido..." style="resize:vertical" />
          </div>
          <div class="field">
            <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
              <input type="checkbox" v-model="form.is_public" style="width:16px;height:16px;accent-color:var(--brand)" />
              Curso público (visible sin código)
            </label>
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" @click="guardar" :disabled="loading" aria-label="Guardar curso">
            <span v-if="loading" class="spinner" style="width:15px;height:15px;border-width:2px"></span>
            {{ loading ? 'Guardando…' : 'Guardar curso' }}
          </button>
          <button class="btn btn-secondary" @click="showForm = false">Cancelar</button>
        </div>
      </div>
    </Transition>

    <!-- Lista de cursos -->
    <div class="cursos-list">
      <!-- Skeleton loading -->
      <template v-if="!capacitaciones.length && loading">
        <div v-for="n in 3" :key="n" class="curso-card" style="margin-bottom:12px">
          <div class="curso-header" style="cursor:default">
            <div class="skeleton skel-thumb" style="width:50px;height:50px;border-radius:10px;flex-shrink:0"></div>
            <div style="flex:1">
              <div class="skeleton skel-title"></div>
              <div class="skeleton skel-text"></div>
            </div>
          </div>
        </div>
      </template>

      <TransitionGroup name="list-item" tag="div" class="cursos-list-inner">
        <div v-for="c in capacitaciones" :key="c.id" class="curso-card" style="margin-bottom:12px">

          <!-- Cabecera del curso -->
          <div class="curso-header" @click="selectCurso(c)" :aria-expanded="selectedCurso?.id === c.id">
            <!-- Miniatura tipo -->
            <div :class="['curso-thumb-mini', c.type === 'video' ? 'thumb-video' : c.type === 'document' ? 'thumb-document' : 'thumb-text']">
              {{ c.type === 'video' ? '🎥' : c.type === 'document' ? '📄' : '📝' }}
            </div>

            <!-- Info -->
            <div class="curso-info">
              <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap">
                <span class="curso-title">{{ c.title }}</span>
                <span :class="['badge', c.is_public ? 'badge-green' : 'badge-gray']">
                  {{ c.is_public ? '🌐 Público' : '🔒 Privado' }}
                </span>
              </div>
              <div class="curso-meta">
                <span>{{ c.description || 'Sin descripción' }}</span>
                <span v-if="c.codigo">Código: <span class="curso-code">{{ c.codigo }}</span></span>
              </div>
            </div>

            <!-- Acciones -->
            <div class="curso-actions" @click.stop>
              <button class="icon-btn" :title="c.is_public ? 'Ocultar curso' : 'Publicar curso'" @click="togglePublic(c.id)">
                {{ c.is_public ? '👁' : '🔓' }}
              </button>
              <button class="icon-btn" title="Regenerar código de acceso" @click="resetCodigo(c.id)">
                🔄
              </button>
              <button class="icon-btn danger" title="Eliminar curso" @click="eliminar(c.id)">
                🗑
              </button>
              <button :class="['icon-btn', selectedCurso?.id === c.id ? 'active' : '']" @click="selectCurso(c)" :title="selectedCurso?.id === c.id ? 'Cerrar' : 'Editar contenido'">
                {{ selectedCurso?.id === c.id ? '✕' : '✏️' }}
              </button>
            </div>
          </div>

          <!-- Panel expandible de contenido -->
          <Transition name="slide-down">
            <div v-if="selectedCurso?.id === c.id" class="curso-panel">

              <!-- Tabs internos -->
              <div class="curso-inner-tabs" role="tablist">
                <button role="tab" :aria-selected="activeTab === 'lecciones'"
                  :class="['curso-tab', activeTab === 'lecciones' ? 'active' : '']"
                  @click="activeTab = 'lecciones'">
                  <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6h16M4 10h16M4 14h8"/></svg>
                  Lecciones
                  <span class="pill-count">{{ lecciones.length }}</span>
                </button>
                <button role="tab" :aria-selected="activeTab === 'intermedias'"
                  :class="['curso-tab', activeTab === 'intermedias' ? 'active' : '']"
                  @click="activeTab = 'intermedias'">
                  <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01"/><circle cx="12" cy="12" r="10"/></svg>
                  Preguntas Intermedias
                  <span class="pill-count">{{ intermedias.length }}</span>
                </button>
                <button role="tab" :aria-selected="activeTab === 'examen'"
                  :class="['curso-tab', activeTab === 'examen' ? 'active' : '']"
                  @click="activeTab = 'examen'">
                  <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
                  Examen Final
                </button>
              </div>

              <!-- TAB: LECCIONES -->
              <div v-if="activeTab === 'lecciones'" class="tab-pane" role="tabpanel">
                <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
                  <span style="font-size:0.88rem;font-weight:600;color:var(--muted)">{{ lecciones.length }} lección(es) en este curso</span>
                  <button class="btn btn-primary btn-sm" @click="showLecForm = !showLecForm; lecForm.orden = lecciones.length + 1">
                    <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
                    Agregar lección
                  </button>
                </div>

                <!-- Formulario nueva lección -->
                <Transition name="slide-down">
                  <div v-if="showLecForm" class="lec-form-card">
                    <p style="font-size:0.9rem;font-weight:700;color:var(--dark);margin-bottom:14px">Nueva lección</p>
                    <div class="form-grid" style="gap:12px">
                      <div class="field full">
                        <label>Título *</label>
                        <input class="field-input" v-model="lecForm.title" placeholder="Ej: Introducción al módulo" />
                      </div>
                      <div class="field full">
                        <label>Descripción</label>
                        <input class="field-input" v-model="lecForm.description" placeholder="Descripción breve" />
                      </div>
                      <div class="field">
                        <label>Tipo de contenido</label>
                        <select class="field-input" v-model="lecForm.type">
                          <option value="video">🎥 Video (subir archivo)</option>
                          <option value="document">📄 PDF / Documento</option>
                          <option value="text">📝 Texto / Lectura</option>
                          <option value="link">🔗 Enlace externo (YouTube, Vimeo…)</option>
                        </select>
                      </div>
                      <div class="field">
                        <label>Duración (minutos)</label>
                        <input type="number" class="field-input" v-model="lecForm.duracion_min" min="0" placeholder="Ej: 15" />
                      </div>
                      <div v-if="lecForm.type === 'video' || lecForm.type === 'document'" class="field full">
                        <label>{{ lecForm.type === 'video' ? 'Archivo de video (mp4, webm…)' : 'Archivo PDF o documento' }}</label>
                        <input type="file" class="field-input" @change="onLecFile" style="padding:7px"
                          :accept="lecForm.type === 'video' ? 'video/*' : '.pdf,.doc,.docx,.ppt,.pptx'" />
                      </div>
                      <div v-if="lecForm.type === 'text'" class="field full">
                        <label>Contenido de la lectura</label>
                        <textarea class="field-input" v-model="lecForm.content" placeholder="Escribe el contenido aquí…" rows="5" style="resize:vertical" />
                      </div>
                      <div v-if="lecForm.type === 'link'" class="field full">
                        <label>URL del recurso</label>
                        <input class="field-input" v-model="lecForm.content" placeholder="https://www.youtube.com/watch?v=…" />
                        <span style="font-size:0.77rem;color:var(--muted);margin-top:4px">Soporta YouTube, Vimeo y cualquier URL embebible.</span>
                      </div>
                    </div>
                    <div class="form-actions">
                      <button class="btn btn-primary btn-sm" @click="guardarLeccion">Guardar lección</button>
                      <button class="btn btn-secondary btn-sm" @click="showLecForm = false">Cancelar</button>
                    </div>
                  </div>
                </Transition>

                <!-- Lista de lecciones -->
                <div v-if="loadingLec" style="padding:32px 0">
                  <div v-for="n in 3" :key="n" class="lec-item" style="margin-bottom:8px">
                    <div class="skeleton" style="width:26px;height:26px;border-radius:50%;flex-shrink:0"></div>
                    <div style="flex:1">
                      <div class="skeleton skel-line" style="width:60%"></div>
                      <div class="skeleton skel-text-sm" style="margin-top:4px"></div>
                    </div>
                  </div>
                </div>

                <div v-else-if="lecciones.length === 0" class="empty-state" style="padding:40px 20px">
                  <div class="empty-icon">📋</div>
                  <p style="font-weight:600;color:var(--dark)">Sin lecciones aún</p>
                  <p style="font-size:0.85rem;color:var(--muted)">Agrega la primera lección con el botón de arriba.</p>
                </div>

                <TransitionGroup v-else name="list-item" tag="div" style="display:flex;flex-direction:column;gap:8px;position:relative">
                  <div v-for="(lec, idx) in lecciones" :key="lec.id" class="lec-item">
                    <div class="lec-num">{{ idx + 1 }}</div>
                    <div :class="['lec-type-dot', lec.type]"></div>
                    <div class="lec-info">
                      <div class="lec-title">{{ lec.title }}</div>
                      <div class="lec-meta">
                        <span>{{ lec.type === 'video' ? '🎥 Video' : lec.type === 'document' ? '📄 PDF' : lec.type === 'link' ? '🔗 Enlace' : '📝 Texto' }}</span>
                        <span v-if="lec.duracion_min">· {{ lec.duracion_min }} min</span>
                        <span v-if="lec.description" style="color:var(--subtle)">· {{ lec.description }}</span>
                      </div>
                    </div>
                    <div class="lec-actions">
                      <button class="lec-btn" @click="moverLeccion(idx, -1)" :disabled="idx === 0" title="Mover arriba" aria-label="Subir lección">↑</button>
                      <button class="lec-btn" @click="moverLeccion(idx, 1)" :disabled="idx === lecciones.length - 1" title="Mover abajo" aria-label="Bajar lección">↓</button>
                      <button class="lec-btn del" @click="eliminarLeccion(lec.id)" title="Eliminar lección" aria-label="Eliminar">✕</button>
                    </div>
                  </div>
                </TransitionGroup>
              </div>

              <!-- TAB: PREGUNTAS INTERMEDIAS -->
              <div v-if="activeTab === 'intermedias'" class="tab-pane" role="tabpanel">
                <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
                  <span style="font-size:0.88rem;font-weight:600;color:var(--muted)">{{ intermedias.length }} pregunta(s) intermedia(s)</span>
                  <button class="btn btn-primary btn-sm" @click="showIntForm = !showIntForm">
                    <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
                    Agregar pregunta
                  </button>
                </div>

                <Transition name="slide-down">
                  <div v-if="showIntForm" class="lec-form-card">
                    <p style="font-size:0.9rem;font-weight:700;color:var(--dark);margin-bottom:14px">Nueva pregunta intermedia</p>
                    <div style="display:flex;flex-direction:column;gap:12px">
                      <div class="field full">
                        <label>Pregunta *</label>
                        <textarea class="field-input" v-model="intForm.texto" placeholder="¿Cuál es la respuesta correcta sobre…?" rows="2" style="resize:vertical" />
                      </div>
                      <div class="form-grid" style="gap:12px">
                        <div class="field">
                          <label>Tipo de pregunta</label>
                          <select class="field-input" v-model="intForm.tipo">
                            <option value="multiple_choice">Opción múltiple</option>
                            <option value="true_false">Verdadero / Falso</option>
                            <option value="open_text">Respuesta abierta</option>
                          </select>
                        </div>
                        <div class="field">
                          <label>Mostrar después de</label>
                          <select class="field-input" v-model="intForm.despues_de_leccion_id">
                            <option value="">Al inicio del curso</option>
                            <option v-for="lec in lecciones" :key="lec.id" :value="lec.id">{{ lec.title }}</option>
                          </select>
                        </div>
                      </div>

                      <div v-if="intForm.tipo === 'multiple_choice'">
                        <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
                          <span style="font-size:0.82rem;font-weight:700;color:var(--muted);text-transform:uppercase;letter-spacing:.04em">Opciones</span>
                          <button class="btn btn-secondary btn-sm" @click="addOpcion">+ Opción</button>
                        </div>
                        <div v-for="(op, i) in intForm.opciones" :key="i" style="display:flex;align-items:center;gap:8px;margin-bottom:6px">
                          <input class="field-input" v-model="op.texto" :placeholder="`Opción ${i+1}`" style="flex:1" />
                          <label style="display:flex;align-items:center;gap:5px;font-size:0.82rem;color:var(--muted);white-space:nowrap;cursor:pointer">
                            <input type="radio" :name="`int-c-${c.id}`" :checked="op.es_correcta"
                              @change="intForm.opciones.forEach((o,j) => o.es_correcta = j===i)"
                              style="accent-color:var(--success)" />
                            Correcta
                          </label>
                          <button v-if="intForm.opciones.length > 2" class="lec-btn del" @click="removeOpcion(i)" title="Eliminar opción">✕</button>
                        </div>
                      </div>

                      <div v-if="intForm.tipo === 'true_false'" class="alert alert-warning" style="font-size:0.82rem">
                        Se generarán las opciones <strong>Verdadero</strong> y <strong>Falso</strong> automáticamente.
                      </div>
                    </div>
                    <div class="form-actions">
                      <button class="btn btn-primary btn-sm" @click="guardarIntermedia">Guardar pregunta</button>
                      <button class="btn btn-secondary btn-sm" @click="showIntForm = false">Cancelar</button>
                    </div>
                  </div>
                </Transition>

                <div v-if="loadingInt" style="padding:32px 0">
                  <div v-for="n in 2" :key="n" class="int-item" style="margin-bottom:8px">
                    <div class="skeleton" style="width:70px;height:22px;border-radius:4px;flex-shrink:0"></div>
                    <div style="flex:1">
                      <div class="skeleton skel-line" style="width:80%"></div>
                      <div class="skeleton skel-text-sm" style="margin-top:4px"></div>
                    </div>
                  </div>
                </div>

                <div v-else-if="intermedias.length === 0" class="empty-state" style="padding:40px 20px">
                  <div class="empty-icon">💬</div>
                  <p style="font-weight:600;color:var(--dark)">Sin preguntas intermedias</p>
                  <p style="font-size:0.85rem;color:var(--muted)">Las preguntas intermedias aparecen entre lecciones para reforzar el aprendizaje.</p>
                </div>

                <TransitionGroup v-else name="list-item" tag="div" style="display:flex;flex-direction:column;gap:8px;position:relative">
                  <div v-for="preg in intermedias" :key="preg.id" class="int-item">
                    <span :class="['int-badge', preg.tipo === 'multiple_choice' ? 'mc' : preg.tipo === 'true_false' ? 'tf' : 'ot']">
                      {{ preg.tipo === 'multiple_choice' ? 'Opción múltiple' : preg.tipo === 'true_false' ? 'V/F' : 'Abierta' }}
                    </span>
                    <div style="flex:1;min-width:0">
                      <p style="font-size:0.88rem;color:var(--dark);line-height:1.4">{{ preg.texto }}</p>
                    </div>
                    <button class="lec-btn del" @click="eliminarIntermedia(preg.id)" title="Eliminar pregunta">✕</button>
                  </div>
                </TransitionGroup>
              </div>

              <!-- TAB: EXAMEN -->
              <div v-if="activeTab === 'examen'" class="tab-pane" role="tabpanel">
                <div class="alert alert-warning" style="margin-bottom:16px;font-size:0.85rem">
                  Para enlazar un examen a este curso, selecciona <strong>{{ c.title }}</strong> al crear o editar el examen en la sección "Exámenes".
                </div>

                <div v-if="misExamenes.filter(e => e.capacitacion_id === c.id).length === 0" class="empty-state" style="padding:32px 20px">
                  <div class="empty-icon">📋</div>
                  <p style="font-weight:600;color:var(--dark)">Sin examen enlazado</p>
                  <p style="font-size:0.85rem;color:var(--muted)">Ve a la sección Exámenes y selecciona este curso para enlazarlo.</p>
                </div>

                <div v-for="ex in misExamenes.filter(e => e.capacitacion_id === c.id)" :key="ex.id" class="lec-item" style="margin-bottom:8px">
                  <div class="lec-type-dot" style="background:var(--success)"></div>
                  <div class="lec-info">
                    <div class="lec-title">{{ ex.title }}</div>
                    <div class="lec-meta">
                      <span class="badge badge-green">✓ Enlazado</span>
                    </div>
                  </div>
                </div>
              </div>

            </div>
          </Transition>
        </div>
      </TransitionGroup>

      <!-- Empty state -->
      <Transition name="fade">
        <div v-if="capacitaciones.length === 0 && !loading" class="empty-state">
          <div class="empty-icon">📚</div>
          <h3>Aún no tienes cursos</h3>
          <p>Crea tu primer curso con el botón <strong>Nuevo Curso</strong> en la parte superior.</p>
          <button class="btn btn-primary" @click="showForm = true">Crear mi primer curso</button>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.cursos-list { /* container */ }
.cursos-list-inner { position: relative; }

.lec-form-card {
  background: var(--bg); border: 1.5px dashed var(--border); border-radius: var(--r);
  padding: 18px; margin-bottom: 16px;
}

.curso-panel { border-top: 1.5px solid var(--border); }
</style>
