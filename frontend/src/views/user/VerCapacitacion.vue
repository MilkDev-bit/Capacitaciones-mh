<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '../../api'

const route = useRoute()
const cursoId = route.params.id as string

const curso = ref<any>(null)
const lecciones = ref<any[]>([])
const selectedLeccion = ref<any | null>(null)
const loading = ref(true)

// Preguntas intermedias
const showIntermedias = ref(false)
const preguntas = ref<any[]>([])
const respuestas = ref<Record<string, any>>({})
const resultadoInt = ref<any | null>(null)

// Foro
const foroPosts = ref<any[]>([])
const nuevoPost = ref({ titulo: '', contenido: '' })
const showNuevoPost = ref(false)
const expandedPost = ref<string | null>(null)
const comentariosMap = ref<Record<string, any[]>>({})
const nuevoComentario = ref<Record<string, string>>({})

const progreso = computed(() => {
  if (!lecciones.value.length) return 0
  const completadas = lecciones.value.filter(l => l.completada).length
  return Math.round((completadas / lecciones.value.length) * 100)
})

const leccionesCompletadas = computed(() => lecciones.value.filter(l => l.completada).length)
const duracionTotal = computed(() =>
  lecciones.value.reduce((total, lec) => total + Number(lec.duracion_min || 0), 0)
)
const currentIndex = computed(() =>
  lecciones.value.findIndex(l => l.id === selectedLeccion.value?.id)
)
const previousLeccion = computed(() =>
  currentIndex.value > 0 ? lecciones.value[currentIndex.value - 1] : null
)
const nextLeccion = computed(() =>
  currentIndex.value >= 0 && currentIndex.value < lecciones.value.length - 1
    ? lecciones.value[currentIndex.value + 1]
    : null
)

async function load() {
  loading.value = true
  const [cRes, lRes] = await Promise.all([
    api.get(`/capacitaciones/${cursoId}`),
    api.get(`/capacitaciones/${cursoId}/lecciones`)
  ])
  curso.value = cRes.data
  lecciones.value = lRes.data || []
  // Instead of auto-selecting, we let selectedLeccion be null to show the Welcome Hero
  loading.value = false
}

function startCourse() {
  if (lecciones.value.length > 0) {
    // Find the first non-completed lesson, or default to the first one
    const firstPending = lecciones.value.find(l => !l.completada)
    selectLeccion(firstPending || lecciones.value[0])
  }
}

onMounted(load)

async function selectLeccion(lec: any) {
  selectedLeccion.value = lec
  showIntermedias.value = false
  resultadoInt.value = null
  respuestas.value = {}
  await loadForo(lec.id)
  await loadPreguntas(lec.id)
}

async function marcarCompleta() {
  if (!selectedLeccion.value || selectedLeccion.value.completada) return
  await api.post(`/lecciones/${selectedLeccion.value.id}/completar`)
  selectedLeccion.value.completada = true
  const idx = lecciones.value.findIndex(l => l.id === selectedLeccion.value.id)
  if (idx >= 0) lecciones.value[idx].completada = true
  // Mostrar preguntas intermedias si hay
  if (preguntas.value.length > 0) {
    showIntermedias.value = true
  }
}

async function loadPreguntas(leccionId: string) {
  const res = await api.get(`/capacitaciones/${cursoId}/intermedias?despues_de_leccion_id=${leccionId}`)
  preguntas.value = res.data || []
}

async function submitIntermedias() {
  const payload = preguntas.value.map(p => {
    const r: any = { pregunta_id: p.id }
    if (p.tipo === 'open_text') {
      r.respuesta_texto = respuestas.value[p.id] || ''
    } else {
      r.opcion_id = respuestas.value[p.id] || ''
    }
    return r
  })
  const res = await api.post(`/capacitaciones/${cursoId}/intermedias/submit`, payload)
  resultadoInt.value = res.data
}

// ── Foro ────────────────────────────────────────────────────────────────────
async function loadForo(leccionId: string) {
  const res = await api.get(`/lecciones/${leccionId}/foro`)
  foroPosts.value = res.data || []
}

async function crearPost() {
  if (!nuevoPost.value.titulo || !nuevoPost.value.contenido) return
  await api.post(`/lecciones/${selectedLeccion.value.id}/foro`, nuevoPost.value)
  nuevoPost.value = { titulo: '', contenido: '' }
  showNuevoPost.value = false
  await loadForo(selectedLeccion.value.id)
}

async function eliminarPost(postId: string) {
  if (!confirm('Eliminar este post?')) return
  await api.delete(`/foro/posts/${postId}`)
  await loadForo(selectedLeccion.value.id)
}

async function togglePost(postId: string) {
  if (expandedPost.value === postId) {
    expandedPost.value = null
    return
  }
  expandedPost.value = postId
  if (!comentariosMap.value[postId]) {
    const res = await api.get(`/foro/posts/${postId}/comentarios`)
    comentariosMap.value[postId] = res.data || []
  }
}

async function crearComentario(postId: string) {
  const texto = nuevoComentario.value[postId]
  if (!texto?.trim()) return
  await api.post(`/foro/posts/${postId}/comentarios`, { contenido: texto })
  nuevoComentario.value[postId] = ''
  const res = await api.get(`/foro/posts/${postId}/comentarios`)
  comentariosMap.value[postId] = res.data || []
}

function fileUrl(path: string) {
  // path ya viene con /uploads/... desde el backend
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
}

function getEmbedUrl(url: string): string {
  if (!url) return ''
  // YouTube
  const yt = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/)([^&?\s]+)/)
  if (yt) return `https://www.youtube.com/embed/${yt[1]}?rel=0`
  // Vimeo
  const vim = url.match(/vimeo\.com\/(\d+)/)
  if (vim) return `https://player.vimeo.com/video/${vim[1]}`
  // Otro (iframe generico)
  return url
}

function typeLabel(t: string) {
  const map: Record<string, string> = { video: 'Video', document: 'PDF / Documento', text: 'Lectura', link: 'Enlace / Video' }
  return map[t] || t
}

function typeIcon(t: string) {
  const map: Record<string, string> = { video: 'V', document: 'D', text: 'L', link: 'YT' }
  return map[t] || '?'
}

async function goToLesson(lec: any | null) {
  if (!lec) return
  await selectLeccion(lec)
}
</script>

<template>
  <div class="ver-curso-shell">
    <!-- Skeleton de carga -->
    <div v-if="loading" class="ver-skeleton">
      <div class="ver-sidebar-skel">
        <div class="skeleton" style="height:20px;width:80%;margin-bottom:12px"></div>
        <div class="skeleton" style="height:8px;width:100%;margin-bottom:20px;border-radius:4px"></div>
        <div v-for="n in 5" :key="n" style="display:flex;gap:10px;margin-bottom:10px;align-items:center">
          <div class="skeleton" style="width:22px;height:22px;border-radius:50%;flex-shrink:0"></div>
          <div style="flex:1">
            <div class="skeleton skel-line" style="width:75%"></div>
            <div class="skeleton skel-text-sm" style="margin-top:4px"></div>
          </div>
        </div>
      </div>
      <div class="ver-content-skel">
        <div class="skeleton" style="height:28px;width:50%;margin-bottom:8px"></div>
        <div class="skeleton skel-text" style="margin-bottom:20px"></div>
        <div class="skeleton" style="height:320px;width:100%;border-radius:12px"></div>
      </div>
    </div>

    <div v-else class="ver-layout">

      <!-- Sidebar lecciones -->
      <aside class="ver-sidebar">
        <div class="ver-sidebar-head">
          <h2 class="ver-curso-nombre">{{ curso?.title }}</h2>
          <p class="ver-course-meta">
            {{ leccionesCompletadas }} de {{ lecciones.length }} lecciones
            <span v-if="duracionTotal"> · {{ duracionTotal }} min</span>
          </p>
          <div class="ver-progress-wrap">
            <div class="ver-progress-top">
              <span>Progreso del curso</span>
              <span class="ver-progress-pct">{{ progreso }}%</span>
            </div>
            <div class="ver-progress-bg">
              <div class="ver-progress-fill" :style="`width:${progreso}%`" />
            </div>
          </div>
        </div>
        <nav class="ver-nav">
          <button v-for="(lec, idx) in lecciones" :key="lec.id"
            @click="selectLeccion(lec)"
            :class="['ver-nav-item', selectedLeccion?.id === lec.id ? 'active' : '', lec.completada ? 'done' : '']"
            :aria-current="selectedLeccion?.id === lec.id ? 'page' : undefined">
            <span class="ver-nav-num">
              <svg v-if="lec.completada" width="10" height="10" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
              <span v-else>{{ idx + 1 }}</span>
            </span>
            <div class="ver-nav-info">
              <p class="ver-nav-title">{{ lec.title }}</p>
              <p class="ver-nav-meta">{{ typeLabel(lec.type) }}<span v-if="lec.duracion_min"> · {{ lec.duracion_min }} min</span></p>
            </div>
            <span :class="['ver-type-pip', lec.type]"></span>
          </button>
          <div v-if="lecciones.length === 0" class="ver-nav-empty">Sin lecciones</div>
        </nav>
      </aside>

      <!-- Contenido principal -->
      <main class="ver-main">
        <div class="ver-main-inner">
          <div v-if="!selectedLeccion" class="ver-welcome-hero">
            <div class="ver-welcome-banner" :style="curso?.thumbnail_url ? `background-image: url('${fileUrl(curso.thumbnail_url)}')` : ''">
              <div class="ver-welcome-overlay"></div>
              <div class="ver-welcome-content">
                <span class="ver-welcome-badge">Módulo de Capacitación</span>
                <h1 class="ver-welcome-title">{{ curso?.title }}</h1>
                <p class="ver-welcome-desc">{{ curso?.description }}</p>
                <div class="ver-welcome-stats">
                  <div class="vw-stat">
                    <strong>{{ lecciones.length }}</strong>
                    <span>Lecciones</span>
                  </div>
                  <div class="vw-stat" v-if="duracionTotal">
                    <strong>{{ duracionTotal }}</strong>
                    <span>Minutos</span>
                  </div>
                  <div class="vw-stat">
                    <strong>{{ progreso }}%</strong>
                    <span>Completado</span>
                  </div>
                </div>
                <button class="btn btn-primary btn-large mt-6" @click="startCourse">
                  {{ progreso > 0 ? 'Continuar curso' : 'Comenzar curso' }}
                </button>
              </div>
            </div>
            <div class="ver-welcome-message" v-if="curso?.welcome_message">
              <h3>Acerca de este curso</h3>
              <p>{{ curso.welcome_message }}</p>
            </div>
          </div>

          <Transition name="fade" mode="out-in">
          <div v-if="selectedLeccion" :key="selectedLeccion.id">
            <!-- Header leccion -->
            <div class="ver-lec-header">
              <div class="ver-lec-header-left">
                <div class="ver-lec-breadcrumb">{{ curso?.title }}</div>
                <h1 class="ver-lec-title">{{ selectedLeccion.title }}</h1>
                <div class="ver-lec-meta-row">
                  <span>{{ currentIndex + 1 }} / {{ lecciones.length }}</span>
                  <span>{{ typeLabel(selectedLeccion.type) }}</span>
                  <span v-if="selectedLeccion.duracion_min">{{ selectedLeccion.duracion_min }} min</span>
                </div>
                <p v-if="selectedLeccion.description" class="ver-lec-desc">{{ selectedLeccion.description }}</p>
              </div>
              <button v-if="!selectedLeccion.completada"
                @click="marcarCompleta"
                class="btn btn-primary btn-sm" style="flex-shrink:0" aria-label="Marcar lección como completada">
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
                Marcar completada
              </button>
              <span v-else class="ver-done-chip">
                <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
                Completada
              </span>
            </div>

            <!-- Reproductor / contenido -->
            <section class="ver-content-card">
              <!-- Video subido -->
              <div v-if="selectedLeccion.type === 'video'" class="ver-media-frame ver-media-video">
                <video v-if="selectedLeccion.file_path" :src="fileUrl(selectedLeccion.file_path)" controls class="ver-media-fill" />
                <div v-else class="ver-media-empty">Sin video disponible</div>
              </div>

              <!-- PDF / Documento embebido -->
              <div v-else-if="selectedLeccion.type === 'document'">
                <div v-if="selectedLeccion.file_path">
                  <iframe :src="fileUrl(selectedLeccion.file_path)" class="ver-doc-frame" />
                  <div class="ver-resource-footer">
                    <a :href="fileUrl(selectedLeccion.file_path)" target="_blank"
                      class="ver-resource-link">Abrir en nueva pestaña</a>
                  </div>
                </div>
                <p v-else class="ver-media-empty ver-media-empty-light">Sin documento adjunto</p>
              </div>

              <!-- Texto / lectura -->
              <div v-else-if="selectedLeccion.type === 'text'" class="ver-reading">
                <div class="ver-reading-content">{{ selectedLeccion.content }}</div>
              </div>

              <!-- Enlace externo: YouTube, Vimeo, otro -->
              <div v-else-if="selectedLeccion.type === 'link'">
                <div v-if="selectedLeccion.content">
                  <div class="ver-media-frame">
                    <iframe :src="getEmbedUrl(selectedLeccion.content)"
                      class="ver-media-fill"
                      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                      allowfullscreen />
                  </div>
                  <div class="ver-resource-footer">
                    <a :href="selectedLeccion.content" target="_blank" rel="noopener"
                      class="ver-resource-link">Abrir enlace original</a>
                  </div>
                </div>
                <p v-else class="ver-media-empty ver-media-empty-light">Sin enlace configurado</p>
              </div>
            </section>

            <div class="ver-lesson-actions">
              <button class="btn btn-secondary" :disabled="!previousLeccion" @click="goToLesson(previousLeccion)">
                Anterior
              </button>
              <button v-if="!selectedLeccion.completada" class="btn btn-primary" @click="marcarCompleta">
                Marcar completada
              </button>
              <button class="btn btn-secondary" :disabled="!nextLeccion" @click="goToLesson(nextLeccion)">
                Siguiente
              </button>
            </div>

            <!-- Preguntas Intermedias -->
            <Transition name="slide-up">
              <div v-if="showIntermedias && preguntas.length > 0" class="ver-intermedias">
                <div class="ver-int-head">
                  <span style="font-size:1.4rem">🧠</span>
                  <div>
                    <h3 style="font-weight:700;color:var(--dark);font-size:1rem">Preguntas de la lección</h3>
                    <p style="font-size:0.82rem;color:var(--muted)">Responde para reforzar tu aprendizaje</p>
                  </div>
                </div>
                <div v-if="resultadoInt" class="ver-int-result">
                  <div style="font-size:2.5rem;font-weight:800;color:var(--brand)">{{ resultadoInt.puntaje.toFixed(1) }} / {{ resultadoInt.puntaje_max.toFixed(1) }}</div>
                  <p style="color:var(--muted);font-size:0.9rem">{{ resultadoInt.porcentaje?.toFixed(0) }}% correcto</p>
                  <button @click="showIntermedias = false" class="btn btn-secondary btn-sm" style="margin-top:12px">Continuar</button>
                </div>
                <div v-else style="display:flex;flex-direction:column;gap:16px">
                  <div v-for="p in preguntas" :key="p.id" class="ver-int-pregunta">
                    <p style="font-size:0.92rem;font-weight:600;color:var(--dark);margin-bottom:10px">{{ p.texto }}</p>
                    <div v-if="p.tipo === 'open_text'">
                      <textarea v-model="respuestas[p.id]" rows="3" placeholder="Tu respuesta..." class="field-input" style="resize:vertical" />
                    </div>
                    <div v-else style="display:flex;flex-direction:column;gap:8px">
                      <label v-for="op in p.opciones" :key="op.id" class="ver-option-label">
                        <input type="radio" :name="p.id" :value="op.id" v-model="respuestas[p.id]" style="accent-color:var(--brand)" />
                        <span>{{ op.texto }}</span>
                      </label>
                    </div>
                  </div>
                  <button @click="submitIntermedias" class="btn btn-primary">Enviar respuestas</button>
                </div>
              </div>
            </Transition>

            <!-- Foro -->
            <div class="ver-foro">
              <div class="ver-foro-head">
                <div>
                  <h3 style="font-weight:700;color:var(--dark);font-size:1rem">💬 Foro de la lección</h3>
                  <p style="font-size:0.8rem;color:var(--muted);margin-top:2px">Pregunta o comenta sobre este contenido</p>
                </div>
                <button @click="showNuevoPost = !showNuevoPost" class="btn btn-secondary btn-sm">
                  {{ showNuevoPost ? 'Cancelar' : '+ Nuevo post' }}
                </button>
              </div>

              <Transition name="slide-down">
                <div v-if="showNuevoPost" class="ver-new-post">
                  <input v-model="nuevoPost.titulo" placeholder="Título del post" class="field-input" style="margin-bottom:8px" />
                  <textarea v-model="nuevoPost.contenido" placeholder="Escribe tu pregunta o comentario..." rows="3" class="field-input" style="resize:vertical;margin-bottom:10px" />
                  <div style="display:flex;gap:8px">
                    <button @click="crearPost" class="btn btn-primary btn-sm">Publicar</button>
                    <button @click="showNuevoPost = false" class="btn btn-secondary btn-sm">Cancelar</button>
                  </div>
                </div>
              </Transition>

              <div v-if="foroPosts.length === 0" style="text-align:center;padding:28px;color:var(--muted);font-size:0.88rem">Sin posts aún. Sé el primero en preguntar.</div>

              <TransitionGroup name="list-item" tag="div" style="display:flex;flex-direction:column;gap:0;position:relative">
                <div v-for="post in foroPosts" :key="post.id" class="ver-post">
                  <div class="ver-post-head" @click="togglePost(post.id)">
                    <div style="flex:1;min-width:0">
                      <p class="ver-post-title">{{ post.titulo }}</p>
                      <p class="ver-post-meta">{{ post.user_name }} · {{ new Date(post.created_at).toLocaleDateString('es') }}</p>
                    </div>
                    <div style="display:flex;gap:6px;align-items:center">
                      <button @click.stop="eliminarPost(post.id)" class="lec-btn del" title="Eliminar post">✕</button>
                      <span class="ver-post-toggle">{{ expandedPost === post.id ? '▲' : '▼' }}</span>
                    </div>
                  </div>
                  <Transition name="slide-down">
                    <div v-if="expandedPost === post.id" class="ver-post-body">
                      <p style="font-size:0.88rem;color:var(--text);margin-bottom:16px;white-space:pre-wrap;line-height:1.6">{{ post.contenido }}</p>
                      <div style="display:flex;flex-direction:column;gap:6px;margin-bottom:12px">
                        <div v-for="com in (comentariosMap[post.id] || [])" :key="com.id" class="ver-comentario">
                          <p style="font-size:0.87rem;color:var(--dark)">{{ com.contenido }}</p>
                          <p style="font-size:0.75rem;color:var(--muted);margin-top:2px">{{ com.user_name }}</p>
                        </div>
                        <p v-if="!(comentariosMap[post.id]?.length)" style="font-size:0.82rem;color:var(--muted)">Sin comentarios aún.</p>
                      </div>
                      <div style="display:flex;gap:8px">
                        <input v-model="nuevoComentario[post.id]" placeholder="Agregar comentario..." class="field-input" style="flex:1" @keydown.enter="crearComentario(post.id)" />
                        <button @click="crearComentario(post.id)" class="btn btn-primary btn-sm">Enviar</button>
                      </div>
                    </div>
                  </Transition>
                </div>
              </TransitionGroup>
            </div>
          </div>
          </Transition>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* Layout shell */
.ver-curso-shell {
  min-height: calc(100vh - var(--topbar-h) - 56px);
}
.ver-layout {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  min-height: calc(100vh - var(--topbar-h) - 56px);
  background: var(--surface);
  border: 1px solid rgba(17, 24, 39, 0.08);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

/* Skeletons */
.ver-skeleton { display: flex; gap: 0; height: 100vh; overflow: hidden; }
.ver-sidebar-skel { width: 280px; background: var(--surface); border-right: 1px solid var(--border); padding: 24px; flex-shrink: 0; }
.ver-content-skel { flex: 1; padding: 36px; }

/* Sidebar */
.ver-sidebar {
  width: 100%; background: var(--surface); border-right: 1.5px solid var(--border);
  display: flex; flex-direction: column; min-width: 0;
}
.ver-sidebar-head { padding: 20px; border-bottom: 1px solid var(--border); }
.ver-curso-nombre { font-size: 0.92rem; font-weight: 800; color: var(--dark); line-height: 1.35; margin-bottom: 12px; }
.ver-course-meta { color: var(--muted); font-size: 0.78rem; margin: -6px 0 12px; }
.ver-progress-wrap { margin-top: 4px; }
.ver-progress-top { display: flex; justify-content: space-between; font-size: 0.77rem; color: var(--muted); margin-bottom: 5px; }
.ver-progress-pct { font-weight: 700; color: var(--brand); }
.ver-progress-bg { height: 7px; background: var(--border-light); border-radius: 4px; overflow: hidden; }
.ver-progress-fill { height: 100%; background: linear-gradient(90deg, var(--brand), var(--brand-dark)); border-radius: 4px; transition: width 0.5s cubic-bezier(0.25,0.46,0.45,0.94); }

.ver-nav { flex: 1; padding: 10px; overflow-y: auto; }
.ver-nav-item {
  width: 100%; text-align: left; padding: 10px 12px; border-radius: var(--r);
  display: flex; align-items: flex-start; gap: 10px; cursor: pointer; border: none;
  background: none; transition: all 0.15s; margin-bottom: 3px; position: relative;
}
.ver-nav-item:hover  { background: var(--bg); }
.ver-nav-item.active { background: var(--brand-light); }
.ver-nav-item.done .ver-nav-num { background: var(--success); color: #fff; }

.ver-nav-num {
  width: 22px; height: 22px; border-radius: 50%; background: var(--border-light);
  color: var(--muted); font-size: 0.72rem; font-weight: 700; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center; margin-top: 1px;
  transition: background 0.15s;
}
.ver-nav-item.active .ver-nav-num { background: var(--brand); color: #fff; }

.ver-nav-info { flex: 1; min-width: 0; }
.ver-nav-title { font-size: 0.87rem; font-weight: 600; color: var(--dark); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ver-nav-item.active .ver-nav-title { color: var(--brand-dark); }
.ver-nav-meta { font-size: 0.75rem; color: var(--muted); margin-top: 2px; }

.ver-type-pip { width: 5px; height: 5px; border-radius: 50%; flex-shrink: 0; margin-top: 8px; }
.ver-type-pip.video    { background: var(--brand); }
.ver-type-pip.document { background: var(--info); }
.ver-type-pip.text     { background: var(--success); }
.ver-type-pip.link     { background: #8b5cf6; }

.ver-nav-empty { text-align: center; padding: 24px; font-size: 0.85rem; color: var(--muted); }

/* Main content */
.ver-main { min-width: 0; background: #f8fafc; }
.ver-main-inner { max-width: 980px; margin: 0 auto; padding: 30px; }

/* Welcome Hero */
.ver-welcome-hero { display: flex; flex-direction: column; gap: 24px; }
.ver-welcome-banner {
  position: relative; padding: 60px 40px; border-radius: var(--r-xl);
  background: linear-gradient(135deg, var(--dark) 0%, #374151 100%);
  color: #fff; overflow: hidden; background-size: cover; background-position: center;
  box-shadow: var(--shadow-md);
}
.ver-welcome-overlay {
  position: absolute; top: 0; left: 0; width: 100%; height: 100%;
  background: linear-gradient(90deg, rgba(29,29,31,0.95) 0%, rgba(29,29,31,0.7) 100%);
  backdrop-filter: blur(4px);
}
.ver-welcome-content { position: relative; z-index: 10; max-width: 600px; }
.ver-welcome-badge {
  display: inline-block; padding: 4px 12px; border-radius: 999px;
  background: rgba(255,255,255,0.15); color: #fff; font-size: 0.75rem;
  font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 16px;
  border: 1px solid rgba(255,255,255,0.1);
}
.ver-welcome-title { font-size: 2.2rem; font-weight: 800; line-height: 1.15; letter-spacing: -0.02em; margin-bottom: 12px; }
.ver-welcome-desc { font-size: 1.05rem; color: rgba(255,255,255,0.8); line-height: 1.5; margin-bottom: 30px; }
.ver-welcome-stats { display: flex; gap: 24px; margin-bottom: 30px; }
.vw-stat { display: flex; flex-direction: column; gap: 4px; }
.vw-stat strong { font-size: 1.4rem; font-weight: 800; color: var(--brand); }
.vw-stat span { font-size: 0.75rem; color: rgba(255,255,255,0.6); text-transform: uppercase; letter-spacing: 0.04em; font-weight: 600; }
.ver-welcome-message {
  background: var(--surface); padding: 32px; border-radius: var(--r-lg);
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
}
.ver-welcome-message h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); margin-bottom: 12px; }
.ver-welcome-message p { font-size: 0.95rem; color: var(--text); line-height: 1.6; white-space: pre-wrap; }
.btn-large { padding: 12px 24px; font-size: 1rem; font-weight: 600; }

/* Lesson header */
.ver-lec-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 20px; flex-wrap: wrap; }
.ver-lec-header-left { flex: 1; min-width: 0; }
.ver-lec-breadcrumb { font-size: 0.78rem; color: var(--muted); margin-bottom: 4px; font-weight: 500; }
.ver-lec-title { font-size: 1.4rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; line-height: 1.25; }
.ver-lec-meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 8px;
}
.ver-lec-meta-row span {
  padding: 3px 9px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--surface);
  color: var(--muted);
  font-size: 0.75rem;
  font-weight: 700;
}
.ver-lec-desc { font-size: 0.88rem; color: var(--muted); margin-top: 6px; line-height: 1.55; }
.ver-done-chip {
  display: inline-flex; align-items: center; gap: 5px;
  background: var(--success-bg); color: var(--success); padding: 6px 14px;
  border-radius: 20px; font-size: 0.82rem; font-weight: 700; flex-shrink: 0;
}

/* Content player */
.ver-content-card {
  overflow: hidden;
  margin-bottom: 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  box-shadow: var(--shadow-xs);
}
.ver-media-frame {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: #0b0f19;
}
.ver-media-video {
  display: flex;
  align-items: center;
  justify-content: center;
}
.ver-media-fill {
  width: 100%;
  height: 100%;
  border: 0;
  display: block;
}
.ver-media-empty {
  min-height: 280px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 28px;
  color: var(--muted);
  font-size: 0.9rem;
  text-align: center;
}
.ver-media-empty-light {
  min-height: 180px;
  background: var(--surface);
}
.ver-doc-frame {
  width: 100%;
  height: min(72vh, 760px);
  min-height: 420px;
  border: 0;
  display: block;
  background: var(--border-light);
}
.ver-resource-footer {
  display: flex;
  justify-content: flex-end;
  padding: 10px 14px;
  border-top: 1px solid var(--border-light);
  background: var(--surface);
}
.ver-resource-link {
  color: var(--info);
  font-size: 0.82rem;
  font-weight: 700;
}
.ver-resource-link:hover {
  text-decoration: underline;
}
.ver-reading {
  padding: 28px;
}
.ver-reading-content {
  max-width: 72ch;
  color: var(--text);
  font-size: 1rem;
  line-height: 1.75;
  white-space: pre-wrap;
}
.ver-lesson-actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 22px;
}

/* Preguntas intermedias */
.ver-intermedias {
  background: linear-gradient(135deg, #fffbeb, #fef3c7); border: 1.5px solid #fcd34d;
  border-radius: var(--r-lg); padding: 22px; margin-bottom: 20px;
}
.ver-int-head { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; padding-bottom: 14px; border-bottom: 1px solid #fcd34d; }
.ver-int-pregunta { background: rgba(255,255,255,.7); border-radius: var(--r); padding: 14px 16px; border: 1px solid #fde68a; }
.ver-option-label { display: flex; align-items: center; gap: 9px; padding: 8px 12px; border-radius: var(--r-sm); cursor: pointer; transition: background 0.12s; font-size: 0.88rem; color: var(--text); }
.ver-option-label:hover { background: rgba(249,115,22,.07); }
.ver-int-result { text-align: center; padding: 20px 0; }

/* Foro */
.ver-foro { background: var(--surface); border-radius: var(--r-lg); border: 1.5px solid var(--border); overflow: hidden; }
.ver-foro-head { padding: 18px 20px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--border-light); }
.ver-new-post { padding: 16px 20px; background: var(--bg); border-bottom: 1px solid var(--border-light); }

.ver-post { border-bottom: 1px solid var(--border-light); }
.ver-post:last-child { border-bottom: none; }
.ver-post-head { padding: 13px 18px; display: flex; align-items: flex-start; cursor: pointer; transition: background 0.12s; }
.ver-post-head:hover { background: var(--bg); }
.ver-post-title { font-size: 0.9rem; font-weight: 600; color: var(--dark); }
.ver-post-meta { font-size: 0.77rem; color: var(--muted); margin-top: 3px; }
.ver-post-toggle { font-size: 0.72rem; color: var(--muted); margin-left: 4px; }
.ver-post-body { padding: 16px 18px; background: var(--bg); border-top: 1px solid var(--border-light); }

.ver-comentario { background: var(--surface); border-radius: var(--r); padding: 10px 12px; border: 1px solid var(--border-light); }

@media (max-width: 768px) {
  .ver-curso-shell {
    min-height: auto;
  }
  .ver-layout {
    grid-template-columns: 1fr;
    min-height: auto;
  }
  .ver-sidebar {
    width: 100%;
    max-height: 320px;
    overflow-y: auto;
    border-right: none;
    border-bottom: 1.5px solid var(--border);
  }
  .ver-main-inner {
    padding: 18px;
  }
  .ver-lec-title {
    font-size: 1.2rem;
  }
  .ver-doc-frame {
    min-height: 360px;
  }
  .ver-lesson-actions,
  .ver-foro-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
