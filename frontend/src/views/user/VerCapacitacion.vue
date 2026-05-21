<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'
import { toast } from '../../utils/toast'

const route = useRoute()
const router = useRouter()
const cursoId = route.params.id as string
const authStore = useAuthStore()
const currentUser = computed(() => authStore.user)

const curso = ref<any>(null)
const lecciones = ref<any[]>([])
const selectedLeccion = ref<any | null>(null)
const loading = ref(true)
const loadError = ref('')

// Mobile sidebar
const sidebarOpen = ref(false)

// Preguntas intermedias
const showIntermedias = ref(false)
const preguntas = ref<any[]>([])
const respuestas = ref<Record<string, any>>({})
const resultadoInt = ref<any | null>(null)

// Foro
const foroPosts = ref<any[]>([])
const nuevoPost = ref({ titulo: '', contenido: '' })
const showNuevoPost = ref(false)
const postFileInput = ref<HTMLInputElement | null>(null)
const postFile = ref<File | null>(null)
const postFilePreview = ref<string | null>(null)
const postFileIsVideo = ref(false)
const expandedPost = ref<string | null>(null)
const comentariosMap = ref<Record<string, any[]>>({})
const nuevoComentario = ref<Record<string, string>>({})
const foroLoading = ref(false)
const foroError = ref('')
const postLoading = ref(false)

// Learning Aids
const focusMode = ref(false)
const notasPersonales = ref<Record<string, string>>({})

const progreso = computed(() => {
  if (!lecciones.value.length) return 0
  const completadas = lecciones.value.filter(l => l.completada).length
  return Math.round((completadas / lecciones.value.length) * 100)
})

const leccionesCompletadas = computed(() => lecciones.value.filter(l => l.completada).length)
const duracionTotal = computed(() =>
  lecciones.value.reduce((total, lec) => total + Number(lec.duracion_min || 0), 0)
)
const tiempoRestante = computed(() => 
  lecciones.value.filter(l => !l.completada).reduce((total, lec) => total + Number(lec.duracion_min || 0), 0)
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
const nextPendingLeccion = computed(() =>
  lecciones.value.find(l => !l.completada && l.id !== selectedLeccion.value?.id)
)

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const [cRes, lRes] = await Promise.all([
      api.get(`/capacitaciones/${cursoId}`),
      api.get(`/capacitaciones/${cursoId}/lecciones`)
    ])
    curso.value = cRes.data
    lecciones.value = lRes.data || []
  } catch (e: any) {
    loadError.value = e.response?.data?.error || 'No pudimos cargar el curso. Verifica tu conexión.'
  } finally {
    loading.value = false
  }
}

function startCourse() {
  if (lecciones.value.length > 0) {
    const firstPending = lecciones.value.find(l => !l.completada)
    selectLeccion(firstPending || lecciones.value[0])
  }
}

onMounted(load)

async function selectLeccion(lec: any) {
  selectedLeccion.value = lec
  sidebarOpen.value = false
  showIntermedias.value = false
  resultadoInt.value = null
  respuestas.value = {}
  
  // Cargar nota local — siempre inicializar la clave para reactividad
  notasPersonales.value[lec.id] = localStorage.getItem(`cap_nota_${cursoId}_${lec.id}`) || ''

  await loadForo(lec.id) // maneja sus errores internamente
  try {
    await loadPreguntas(lec.id)
  } catch { /* silently fail */ }
}

let notaTimer: ReturnType<typeof setTimeout> | null = null
function guardarNota() {
  if (selectedLeccion.value) {
    localStorage.setItem(`cap_nota_${cursoId}_${selectedLeccion.value.id}`, notasPersonales.value[selectedLeccion.value.id] || '')
    if (notaTimer) clearTimeout(notaTimer)
    notaTimer = setTimeout(() => toast.success('Nota guardada'), 1200)
  }
}

const showConfetti = ref(false)

async function marcarCompleta() {
  if (!selectedLeccion.value || selectedLeccion.value.completada) return
  try {
    await api.post(`/lecciones/${selectedLeccion.value.id}/completar`)
    selectedLeccion.value.completada = true
    const idx = lecciones.value.findIndex(l => l.id === selectedLeccion.value.id)
    if (idx >= 0) lecciones.value[idx].completada = true
    toast.success('Lección completada')
    
    if (progreso.value === 100) {
      showConfetti.value = true
      setTimeout(() => { showConfetti.value = false }, 5000)
    }

    // Mostrar preguntas intermedias si hay
    if (preguntas.value.length > 0) {
      showIntermedias.value = true
    }
  } catch {
    toast.error('Error al marcar la lección')
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
  foroLoading.value = true
  foroError.value = ''
  try {
    const res = await api.get(`/lecciones/${leccionId}/foro`)
    foroPosts.value = res.data || []
    // Pre-cargar comentarios para todos los posts
    for (const post of foroPosts.value) {
      const cRes = await api.get(`/foro/posts/${post.id}/comentarios`)
      comentariosMap.value[post.id] = cRes.data || []
    }
  } catch {
    foroError.value = 'No se pudieron cargar los posts del foro'
  } finally {
    foroLoading.value = false
  }
}

async function crearPost() {
  if (!nuevoPost.value.titulo || !nuevoPost.value.contenido) return
  postLoading.value = true
  foroError.value = ''
  const loadingToast = postFile.value && postFileIsVideo.value
    ? toast.loading('Subiendo video al foro...')
    : null
  try {
    const fd = new FormData()
    fd.append('titulo', nuevoPost.value.titulo)
    fd.append('contenido', nuevoPost.value.contenido)
    if (postFile.value) fd.append('media', postFile.value)
    await api.post(`/lecciones/${selectedLeccion.value.id}/foro`, fd)
    nuevoPost.value = { titulo: '', contenido: '' }
    removePostFile()
    showNuevoPost.value = false
    await loadForo(selectedLeccion.value.id)
    toast.success('Post publicado')
  } catch {
    foroError.value = 'Error al publicar el post. Inténtalo de nuevo.'
    toast.error('Error al publicar el post')
  } finally {
    postLoading.value = false
    loadingToast?.close()
  }
}

function onPostFile(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  if (!f) return
  if (f.size > 50 * 1024 * 1024) {
    toast.error('El archivo no puede superar 50 MB')
    return
  }
  postFile.value = f
  postFileIsVideo.value = f.type.startsWith('video/')
  if (postFilePreview.value) URL.revokeObjectURL(postFilePreview.value)
  postFilePreview.value = URL.createObjectURL(f)
}

function removePostFile() {
  postFile.value = null
  if (postFilePreview.value) URL.revokeObjectURL(postFilePreview.value)
  postFilePreview.value = null
  postFileIsVideo.value = false
  if (postFileInput.value) postFileInput.value.value = ''
}

function openMedia(url: string) { window.open(url, '_blank') }

async function eliminarPost(postId: string) {
  if (!await toast.confirm('Eliminar este post?')) return
  await api.delete(`/foro/posts/${postId}`)
  await loadForo(selectedLeccion.value.id)
}

async function togglePost(postId: string) {
  if (expandedPost.value === postId) {
    expandedPost.value = null
    return
  }
  expandedPost.value = postId
  const res = await api.get(`/foro/posts/${postId}/comentarios`)
  comentariosMap.value[postId] = res.data || []
}

async function toggleLike(postId: string) {
  const post = foroPosts.value.find(p => p.id === postId)
  if (!post) return
  const wasLiked = post.user_liked
  post.user_liked = !wasLiked
  post.like_count = (post.like_count || 0) + (post.user_liked ? 1 : -1)
  try {
    const res = await api.post(`/foro/posts/${postId}/like`)
    post.user_liked = res.data.liked
    post.like_count = res.data.count
  } catch {
    post.user_liked = wasLiked
    post.like_count = (post.like_count || 0) + (wasLiked ? 1 : -1)
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

function foroInitials(name: string) {
  return name.split(' ').slice(0, 2).map((w: string) => w[0]).join('').toUpperCase() || '?'
}

function meInitials() {
  const n = currentUser.value?.name || ''
  return n.split(' ').slice(0, 2).map((w: string) => w[0]).join('').toUpperCase() || 'YO'
}

function timeAgo(dateStr: string) {
  const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (diff < 60) return 'ahora mismo'
  if (diff < 3600) return `${Math.floor(diff / 60)}m`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h`
  if (diff < 604800) return `${Math.floor(diff / 86400)}d`
  return new Date(dateStr).toLocaleDateString('es', { month: 'short', day: 'numeric' })
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

function isNextPending(lec: any) {
  if (lec.completada) return false
  const firstPending = lecciones.value.find(l => !l.completada)
  return firstPending?.id === lec.id
}

async function goToLesson(lec: any | null) {
  if (!lec) return
  await selectLeccion(lec)
}

// Forum profile card popup
const foroProfileCard = ref<null | { id: string; name: string }>(null)

function openForoProfile(id: string, name: string) {
  foroProfileCard.value = { id, name }
}

function iniciarConversacion() {
  foroProfileCard.value = null
  toast.info('La mensajería directa estará disponible próximamente.')
}

function goBack() {
  router.push('/usuario/capacitaciones')
}
</script>

<template>
  <div class="ver-curso-shell">
    <!-- Main content -->

    <!-- Error state -->
    <div v-if="loadError && !loading" class="ver-error-state">
      <div class="ver-error-icon">
        <svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/></svg>
      </div>
      <h2>No se pudo cargar el curso</h2>
      <p>{{ loadError }}</p>
      <div style="display:flex;gap:10px">
        <button class="btn btn-primary" @click="load">Reintentar</button>
        <button class="btn btn-secondary" @click="goBack">Volver a mis cursos</button>
      </div>
    </div>

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

    <div v-else-if="!loadError" :class="['ver-layout', focusMode ? 'focus-mode' : '']">

      <!-- Mobile sidebar overlay -->
      <div :class="['ver-sidebar-overlay', sidebarOpen ? 'open' : '']" @click="sidebarOpen = false"></div>

      <!-- Mobile toggle button -->
      <button class="ver-mobile-toggle" @click="sidebarOpen = !sidebarOpen" aria-label="Ver índice de lecciones">
        <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6h16M4 12h16M4 18h16"/></svg>
        <span>Índice</span>
        <span class="ver-mobile-toggle-count">{{ leccionesCompletadas }}/{{ lecciones.length }}</span>
      </button>

      <!-- Sidebar lecciones -->
      <aside :class="['ver-sidebar', sidebarOpen ? 'open' : '']">
        <div class="ver-sidebar-head">
          <button class="ver-back-btn" @click="goBack" title="Volver a mis cursos">
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M15 19l-7-7 7-7" stroke-linecap="round" stroke-linejoin="round"/></svg>
            Mis cursos
          </button>
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
            :class="['ver-nav-item', selectedLeccion?.id === lec.id ? 'active' : '', lec.completada ? 'done' : '', isNextPending(lec) ? 'next-pending' : '']"
            :aria-current="selectedLeccion?.id === lec.id ? 'page' : undefined">
            <span class="ver-nav-num">
              <svg v-if="lec.completada" width="10" height="10" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
              <span v-else>{{ idx + 1 }}</span>
            </span>
            <div class="ver-nav-info">
              <p class="ver-nav-title">{{ lec.title }}</p>
              <p class="ver-nav-meta">{{ typeLabel(lec.type) }}<span v-if="lec.duracion_min"> · {{ lec.duracion_min }} min</span></p>
            </div>
            <span v-if="isNextPending(lec)" class="ver-next-badge">Siguiente</span>
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
                  <div class="vw-stat" v-if="tiempoRestante > 0">
                    <strong>{{ tiempoRestante }}</strong>
                    <span>Minutos restantes</span>
                  </div>
                  <div class="vw-stat" v-else-if="duracionTotal">
                    <strong>{{ duracionTotal }}</strong>
                    <span>Minutos totales</span>
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

            <!-- Cómo funciona -->
            <div class="ver-how-it-works">
              <h3>¿Cómo funciona?</h3>
              <div class="ver-steps">
                <div class="ver-step">
                  <div class="ver-step-num">1</div>
                  <div>
                    <strong>Ve el contenido</strong>
                    <p>Mira el video, lee el documento o revisa el material de la lección.</p>
                  </div>
                </div>
                <div class="ver-step">
                  <div class="ver-step-num">2</div>
                  <div>
                    <strong>Marca completada</strong>
                    <p>Cuando termines, presiona el botón "Marcar completada" para registrar tu avance.</p>
                  </div>
                </div>
                <div class="ver-step">
                  <div class="ver-step-num">3</div>
                  <div>
                    <strong>Contesta las preguntas</strong>
                    <p>Algunas lecciones tienen preguntas para reforzar lo aprendido.</p>
                  </div>
                </div>
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
              <div class="ver-lec-header-right" style="display:flex;gap:12px;align-items:center;">
                <button class="btn btn-secondary btn-sm" @click="focusMode = !focusMode" aria-label="Alternar modo enfoque">
                  <svg v-if="!focusMode" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 8V4h4m12 4V4h-4M4 16v4h4m12-4v4h-4"/></svg>
                  <svg v-else width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 14h6v6m10-10h-6V4m0 10l7 7M10 10L3 3"/></svg>
                  {{ focusMode ? 'Salir del Enfoque' : 'Modo Enfoque' }}
                </button>
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
                ← Anterior
              </button>
              <button v-if="!selectedLeccion.completada" class="btn btn-primary" @click="marcarCompleta">
                ✓ Marcar completada
              </button>
              <button class="btn btn-secondary" :disabled="!nextLeccion" @click="goToLesson(nextLeccion)">
                Siguiente →
              </button>
            </div>

            <!-- Mis Notas -->
            <div class="ver-notes-section">
              <div class="ver-notes-head">
                <h3 class="ver-section-title">
                  <span class="gm-icon gm-icon-notes">
                    <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5"/><path d="M17.5 2.5a2.121 2.121 0 013 3L12 14l-4 1 1-4 7.5-7.5z"/></svg>
                  </span>
                  Mis Notas
                </h3>
                <span class="ver-notes-status">{{ notasPersonales[selectedLeccion.id] ? 'Guardado localmente' : 'Escribe para guardar' }}</span>
              </div>
              <textarea 
                v-model="notasPersonales[selectedLeccion.id]" 
                @input="guardarNota"
                placeholder="Escribe tus apuntes personales para esta lección aquí..." 
                class="field-input ver-notes-input" 
                rows="4"
              ></textarea>
            </div>

            <!-- Sugerencia siguiente lección -->
            <Transition name="slide-up">
              <div v-if="selectedLeccion.completada && nextPendingLeccion" class="ver-next-suggestion">
                <div class="ver-next-suggestion-left">
                  <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M13 7l5 5m0 0l-5 5m5-5H6"/></svg>
                  <div>
                    <strong>Siguiente lección</strong>
                    <p>{{ nextPendingLeccion.title }}</p>
                  </div>
                </div>
                <button class="btn btn-primary btn-sm" @click="goToLesson(nextPendingLeccion)">Continuar →</button>
              </div>
            </Transition>

            <!-- Preguntas Intermedias -->
            <Transition name="slide-up">
              <div v-if="showIntermedias && preguntas.length > 0" class="ver-intermedias">
                <div class="ver-int-head">
                  <span class="gm-icon gm-icon-brain" style="width:38px;height:38px;flex-shrink:0">
                    <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path d="M12 2a7 7 0 014.9 11.9c-.7.7-1.2 1.6-1.4 2.6H8.5c-.2-1-.7-1.9-1.4-2.6A7 7 0 0112 2z"/><path d="M9 18h6M10 22h4"/></svg>
                  </span>
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
              <!-- Foro header -->
              <div class="ver-foro-head">
                <h3 class="ver-section-title">
                  <span class="gm-icon gm-icon-forum">
                    <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>
                  </span>
                  Foro de la lección
                </h3>
                <p class="ver-foro-sub">Participa, pregunta y comparte con la comunidad</p>
              </div>

              <div v-if="foroError" class="foro-msg foro-msg-error">{{ foroError }}</div>
              <div v-if="foroLoading" class="foro-loading">
                <span class="btn-spinner" style="border-color:var(--brand-light);border-top-color:var(--brand)"></span>
                <span>Cargando...</span>
              </div>

              <!-- Caja "¿Qué estás pensando?" al estilo Facebook -->
              <div v-if="!foroLoading" class="fb-create-box">
                <div class="fb-create-avatar">{{ meInitials() }}</div>
                <button class="fb-create-trigger" @click="showNuevoPost = true">
                  ¿Qué quieres preguntar sobre esta lección?
                </button>
              </div>

              <Transition name="slide-down">
                <div v-if="showNuevoPost" class="fb-post-form-card">
                  <div class="fb-post-form-header">
                    <h4>Crear publicación</h4>
                    <button class="fb-close-btn" @click="showNuevoPost = false">
                      <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>
                    </button>
                  </div>
                  <div class="fb-post-form-author">
                    <div class="fb-create-avatar">{{ meInitials() }}</div>
                    <span class="fb-post-form-name">{{ currentUser?.name }}</span>
                  </div>
                  <input v-model="nuevoPost.titulo" placeholder="Título de tu publicación..." class="field-input" style="margin-bottom:10px" />
                  <textarea v-model="nuevoPost.contenido" placeholder="¿Qué quieres compartir con el grupo?" rows="4" class="field-input" style="resize:vertical;margin-bottom:12px" />
                  <!-- Previsualización de adjunto -->
                  <div v-if="postFilePreview" class="fb-file-preview">
                    <video v-if="postFileIsVideo" :src="postFilePreview" class="fb-preview-media" controls muted />
                    <img v-else :src="postFilePreview" class="fb-preview-media" />
                    <button class="fb-remove-file" @click="removePostFile" type="button" title="Quitar adjunto">
                      <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>
                    </button>
                  </div>
                  <div class="fb-post-form-footer">
                    <div class="fb-attach-area">
                      <button type="button" class="fb-attach-btn" @click="postFileInput?.click()" title="Adjuntar imagen o video">
                        <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/></svg>
                        Adjuntar
                      </button>
                      <input ref="postFileInput" type="file" accept="image/jpeg,image/png,image/webp,image/gif,video/mp4,video/webm,video/quicktime" style="display:none" @change="onPostFile" />
                    </div>
                    <button @click="showNuevoPost = false" class="btn btn-secondary btn-sm">Cancelar</button>
                    <button @click="crearPost" class="btn btn-primary btn-sm" :disabled="postLoading || !nuevoPost.titulo || !nuevoPost.contenido">
                      <span v-if="postLoading" class="btn-spinner"></span>
                      {{ postLoading ? 'Publicando...' : 'Publicar' }}
                    </button>
                  </div>
                </div>
              </Transition>

              <div v-if="!foroLoading && foroPosts.length === 0" class="fb-empty-foro">
                <svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>
                <p>Aún no hay publicaciones en este foro.</p>
                <span>¡Sé el primero en preguntar o compartir algo!</span>
              </div>

              <!-- Posts estilo Facebook -->
              <TransitionGroup name="list-item" tag="div" class="fb-posts-list">
                <div v-for="post in foroPosts" :key="post.id" class="fb-post-card">

                  <!-- Card header: avatar, nombre, tiempo, opciones -->
                  <div class="fb-post-header">
                    <div class="fb-post-avatar-wrap" @click.stop="openForoProfile(post.user_id, post.user_name)" style="cursor:pointer" title="Ver perfil">
                      <div class="fb-post-avatar">{{ foroInitials(post.user_name) }}</div>
                    </div>
                    <div class="fb-post-meta">
                      <router-link :to="`/usuario/perfil/${post.user_id}`" class="fb-post-author">{{ post.user_name }}</router-link>
                      <span class="fb-post-time">{{ timeAgo(post.created_at) }}</span>
                    </div>
                    <button @click="eliminarPost(post.id)" class="fb-delete-btn" title="Eliminar publicación">
                      <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>
                    </button>
                  </div>

                  <!-- Contenido del post -->
                  <div class="fb-post-body">
                    <h4 class="fb-post-title">{{ post.titulo }}</h4>
                    <p class="fb-post-content">{{ post.contenido }}</p>
                    <!-- Media adjunta al post -->
                    <div v-if="post.media_url" class="fb-post-media">
                      <video v-if="post.media_type === 'video'" :src="post.media_url" class="fb-post-media-video" controls />
                      <img v-else :src="post.media_url" class="fb-post-media-img" @click="openMedia(post.media_url)" title="Ver imagen completa" />
                    </div>
                  </div>

                  <!-- Contador de likes y comentarios -->
                  <div v-if="post.like_count > 0 || (comentariosMap[post.id] || []).length > 0" class="fb-post-stats">
                    <span v-if="post.like_count > 0" class="fb-stat-likes">
                      <span class="fb-like-bubble">
                        <svg width="10" height="10" fill="currentColor" viewBox="0 0 20 20"><path d="M2 10.5a1.5 1.5 0 113 0v6a1.5 1.5 0 01-3 0v-6zm6-10A1.5 1.5 0 006.5 2v1.5a.5.5 0 01-.5.5H4a2 2 0 00-2 2v1a2 2 0 002 2h.5v1a.5.5 0 01-.5.5H3a1 1 0 000 2h.5a.5.5 0 01.5.5V17a1 1 0 001 1h8a1 1 0 001-1v-2.5a.5.5 0 01.5-.5H16a2 2 0 002-2V8a2 2 0 00-2-2h-2a.5.5 0 01-.5-.5V4a2 2 0 00-2-2h-1.5A1.5 1.5 0 008 3v-.5z"/></svg>
                      </span>
                      {{ post.like_count }}
                    </span>
                    <span v-if="(comentariosMap[post.id] || []).length > 0" class="fb-stat-comments">
                      {{ (comentariosMap[post.id] || []).length }} {{ (comentariosMap[post.id] || []).length === 1 ? 'comentario' : 'comentarios' }}
                    </span>
                  </div>

                  <!-- Botones de acción: Me gusta y Comentar -->
                  <div class="fb-post-actions">
                    <button @click="toggleLike(post.id)" :class="['fb-action-btn', post.user_liked ? 'fb-action-liked' : '']">
                      <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path d="M14 9V5a3 3 0 00-3-3l-4 9v11h11.28a2 2 0 002-1.7l1.38-9a2 2 0 00-2-2.3H14z" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M7 22H4a2 2 0 01-2-2v-7a2 2 0 012-2h3" stroke-linecap="round" stroke-linejoin="round"/>
                      </svg>
                      {{ post.user_liked ? 'Me gusta' : 'Me gusta' }}
                    </button>
                    <button @click="togglePost(post.id)" :class="['fb-action-btn', expandedPost === post.id ? 'fb-action-active' : '']">
                      <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                      Comentar
                    </button>
                  </div>

                  <!-- Sección de comentarios -->
                  <Transition name="slide-down">
                    <div v-if="expandedPost === post.id" class="fb-comments-section">
                      <div v-for="com in (comentariosMap[post.id] || [])" :key="com.id" class="fb-comment">
                        <div class="fb-comment-avatar" @click.stop="openForoProfile(com.user_id, com.user_name)" style="cursor:pointer" title="Ver perfil">{{ foroInitials(com.user_name) }}</div>
                        <div class="fb-comment-bubble">
                          <router-link :to="`/usuario/perfil/${com.user_id}`" class="fb-comment-author">{{ com.user_name }}</router-link>
                          <p class="fb-comment-text">{{ com.contenido }}</p>
                        </div>
                      </div>
                      <p v-if="!(comentariosMap[post.id] || []).length" class="fb-no-comments">Sin comentarios aún.</p>
                      <!-- Input nuevo comentario -->
                      <div class="fb-new-comment-row">
                        <div class="fb-comment-avatar me">{{ meInitials() }}</div>
                        <div class="fb-comment-input-wrap">
                          <input v-model="nuevoComentario[post.id]"
                            @keydown.enter="crearComentario(post.id)"
                            placeholder="Escribe un comentario..."
                            class="fb-comment-input" />
                          <button @click="crearComentario(post.id)" class="fb-comment-send" :disabled="!nuevoComentario[post.id]?.trim()" title="Enviar">
                            <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                          </button>
                        </div>
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

    <!-- Forum Profile Card Popup -->
    <Transition name="fade">
      <div v-if="foroProfileCard" class="foro-card-backdrop" @click="foroProfileCard = null">
        <div class="foro-profile-card" @click.stop>
          <button class="fpc-close" @click="foroProfileCard = null" aria-label="Cerrar">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
          <div class="fpc-avatar">{{ foroInitials(foroProfileCard.name) }}</div>
          <div class="fpc-name">{{ foroProfileCard.name }}</div>
          <div class="fpc-role">Participante del foro</div>
          <div class="fpc-actions">
            <RouterLink :to="`/usuario/perfil/${foroProfileCard.id}`" class="fpc-btn fpc-btn-primary" @click="foroProfileCard = null">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/></svg>
              Ver perfil
            </RouterLink>
            <button class="fpc-btn fpc-btn-secondary" @click="iniciarConversacion">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
              Iniciar conversación
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Celebration Confetti Overlay -->
    <Transition name="fade">
      <div v-if="showConfetti" class="ver-confetti-overlay">
        <div class="ver-confetti-card">
          <div class="ver-confetti-icon">
            <span class="gm-icon gm-icon-trophy">
              <svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M8.21 13.89L7 23l5-3 5 3-1.21-9.12M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
            </span>
          </div>
          <h2>¡Felicidades!</h2>
          <p>Has completado el 100% de <strong>{{ curso?.title }}</strong>.</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
/* Toast notification */
.ver-toast {
  position: fixed; top: 20px; left: 50%; transform: translateX(-50%); z-index: 9999;
  background: var(--dark); color: #fff; padding: 12px 28px; border-radius: 999px;
  font-size: 0.9rem; font-weight: 700; box-shadow: 0 8px 30px rgba(0,0,0,.2);
  pointer-events: none;
}

/* Error state */
.ver-error-state {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 14px; padding: 80px 24px; text-align: center; min-height: 50vh;
}
.ver-error-icon { color: var(--muted); }
.ver-error-state h2 { font-size: 1.3rem; font-weight: 800; color: var(--dark); }
.ver-error-state p { color: var(--muted); font-size: 0.92rem; max-width: 420px; }

/* Back button */
.ver-back-btn {
  display: inline-flex; align-items: center; gap: 6px; background: none; border: none;
  color: var(--muted); font-size: 0.8rem; font-weight: 600; cursor: pointer;
  padding: 4px 0; margin-bottom: 10px; transition: color 0.15s;
}
.ver-back-btn:hover { color: var(--brand); }

/* Mobile toggle */
.ver-mobile-toggle {
  display: none; align-items: center; gap: 8px; padding: 10px 16px;
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r);
  font-size: 0.85rem; font-weight: 700; color: var(--dark); cursor: pointer;
  grid-column: 1 / -1; transition: all 0.15s;
}
.ver-mobile-toggle:hover { border-color: var(--brand); }
.ver-mobile-toggle-count {
  margin-left: auto; background: var(--brand-light); color: var(--brand-dark);
  padding: 2px 8px; border-radius: 999px; font-size: 0.75rem;
}

/* Mobile sidebar overlay */
.ver-sidebar-overlay {
  display: none; position: fixed; inset: 0; background: rgba(0,0,0,.5); z-index: 200;
}
.ver-sidebar-overlay.open { display: block; }

/* Next pending badge */
.ver-nav-item.next-pending { border-left: 3px solid var(--brand); }
.ver-next-badge {
  font-size: 0.65rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em;
  color: var(--brand-dark); background: var(--brand-light); padding: 2px 7px;
  border-radius: 999px; flex-shrink: 0; white-space: nowrap;
}

/* How it works steps */
.ver-how-it-works {
  background: var(--surface); padding: 28px; border-radius: var(--r-lg);
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
}
.ver-how-it-works h3 { font-size: 1.05rem; font-weight: 700; color: var(--dark); margin-bottom: 18px; }
.ver-steps { display: flex; flex-direction: column; gap: 16px; }
.ver-step { display: flex; align-items: flex-start; gap: 14px; }
.ver-step-num {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, var(--brand), var(--brand-dark)); color: #fff;
  font-size: 0.85rem; font-weight: 800; display: flex; align-items: center; justify-content: center;
}
.ver-step strong { font-size: 0.9rem; font-weight: 700; color: var(--dark); display: block; }
.ver-step p { font-size: 0.82rem; color: var(--muted); margin-top: 2px; line-height: 1.45; }

/* Next lesson suggestion */
.ver-next-suggestion {
  display: flex; align-items: center; justify-content: space-between; gap: 14px;
  background: linear-gradient(135deg, #eff6ff, #dbeafe); border: 1.5px solid #93c5fd;
  border-radius: var(--r-lg); padding: 16px 20px; margin-bottom: 20px;
}
.ver-next-suggestion-left { display: flex; align-items: center; gap: 12px; color: var(--info); flex: 1; min-width: 0; }
.ver-next-suggestion-left strong { font-size: 0.82rem; font-weight: 700; color: var(--dark); display: block; }
.ver-next-suggestion-left p { font-size: 0.85rem; color: var(--muted); margin-top: 1px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

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
.ver-foro { background: transparent; display: flex; flex-direction: column; gap: 14px; margin-top: 24px; }
.ver-foro-head { margin-bottom: 2px; }
.ver-foro-sub { font-size: 0.8rem; color: var(--muted); margin-top: 2px; }

/* Facebook: caja crear publicación */
.fb-create-box {
  display: flex; align-items: center; gap: 12px;
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: var(--r-lg); padding: 12px 16px;
  box-shadow: var(--shadow-sm);
}
.fb-create-avatar {
  width: 40px; height: 40px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, var(--brand), #ef4444);
  color: #fff; font-size: 0.85rem; font-weight: 800;
}
.fb-create-trigger {
  flex: 1; background: var(--bg); border: 1px solid var(--border-light);
  border-radius: 999px; padding: 9px 16px; text-align: left;
  color: var(--muted); font-size: 0.88rem; cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}
.fb-create-trigger:hover { background: rgba(249,115,22,.06); border-color: var(--brand); }

/* Facebook: formulario publicación */
.fb-post-form-card {
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: var(--r-lg); padding: 20px; box-shadow: var(--shadow-md);
  display: flex; flex-direction: column; gap: 12px;
}
.fb-post-form-header { display: flex; align-items: center; justify-content: space-between; }
.fb-post-form-header h4 { font-size: 1rem; font-weight: 800; color: var(--dark); }
.fb-close-btn {
  width: 32px; height: 32px; border-radius: 50%; border: none;
  background: var(--bg); color: var(--muted); display: flex; align-items: center; justify-content: center;
  cursor: pointer; transition: all 0.15s;
}
.fb-close-btn:hover { background: #fef2f2; color: #dc2626; }
.fb-post-form-author { display: flex; align-items: center; gap: 10px; }
.fb-post-form-name { font-weight: 600; font-size: 0.9rem; color: var(--dark); }
.fb-post-form-footer { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.fb-attach-area { display: flex; align-items: center; gap: 6px; }
.fb-attach-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 12px; border-radius: var(--r); border: 1px solid var(--border);
  background: transparent; color: var(--muted); cursor: pointer;
  font-size: 0.8rem; font-weight: 500; transition: background .15s, color .15s;
}
.fb-attach-btn:hover { background: var(--brand); color: #fff; border-color: var(--brand); }
.fb-file-preview {
  position: relative; margin-bottom: 12px; border-radius: var(--r);
  overflow: hidden; max-height: 220px; background: #000;
}
.fb-preview-media {
  width: 100%; max-height: 220px; object-fit: contain; display: block;
}
.fb-remove-file {
  position: absolute; top: 6px; right: 6px;
  width: 24px; height: 24px; border-radius: 50%;
  background: rgba(0,0,0,.55); border: none; color: #fff;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
}
.fb-remove-file:hover { background: rgba(0,0,0,.8); }
.fb-post-media { margin-top: 10px; border-radius: var(--r); overflow: hidden; }
.fb-post-media-img {
  width: 100%; max-height: 400px; object-fit: cover; display: block; cursor: pointer;
  transition: opacity .15s;
}
.fb-post-media-img:hover { opacity: .9; }
.fb-post-media-video { width: 100%; max-height: 400px; display: block; background: #000; }

/* Facebook: estado vacío */
.fb-empty-foro {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 40px 24px; text-align: center; color: var(--muted);
}
.fb-empty-foro p { font-size: 0.95rem; font-weight: 600; color: var(--dark); margin: 0; }
.fb-empty-foro span { font-size: 0.83rem; }

/* Facebook: lista de posts */
.fb-posts-list { display: flex; flex-direction: column; gap: 14px; }

/* Facebook: card de post */
.fb-post-card {
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: var(--r-lg); overflow: hidden; box-shadow: var(--shadow-sm);
  transition: box-shadow 0.2s;
}
.fb-post-card:hover { box-shadow: var(--shadow-md); }

.fb-post-header { display: flex; align-items: flex-start; gap: 10px; padding: 14px 16px 10px; }
.fb-post-avatar-wrap { flex-shrink: 0; }
.fb-post-avatar {
  width: 40px; height: 40px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff; font-size: 0.8rem; font-weight: 800;
}
.fb-post-meta { flex: 1; display: flex; flex-direction: column; gap: 2px; }
.fb-post-author { font-size: 0.88rem; font-weight: 700; color: var(--dark); text-decoration: none; transition: color 0.15s; }
.fb-post-author:hover { color: var(--brand); }
.fb-post-time { font-size: 0.75rem; color: var(--muted); }
.fb-delete-btn {
  width: 30px; height: 30px; border-radius: 50%; border: none;
  background: transparent; color: var(--muted); cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s; flex-shrink: 0;
}
.fb-delete-btn:hover { background: #fef2f2; color: #dc2626; }

.fb-post-body { padding: 0 16px 14px; }
.fb-post-title { font-size: 0.95rem; font-weight: 700; color: var(--dark); margin-bottom: 6px; }
.fb-post-content { font-size: 0.88rem; color: var(--text); line-height: 1.65; white-space: pre-wrap; }

/* Contadores */
.fb-post-stats {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 16px 10px; font-size: 0.78rem; color: var(--muted);
}
.fb-stat-likes { display: flex; align-items: center; gap: 5px; }
.fb-like-bubble {
  width: 18px; height: 18px; border-radius: 50%;
  background: var(--brand); color: #fff;
  display: inline-flex; align-items: center; justify-content: center;
}
.fb-stat-comments { cursor: pointer; }
.fb-stat-comments:hover { text-decoration: underline; }

/* Botones acción */
.fb-post-actions {
  display: flex; border-top: 1px solid var(--border-light);
  padding: 4px 8px;
}
.fb-action-btn {
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 7px;
  padding: 8px 12px; border-radius: var(--r); border: none; background: transparent;
  color: var(--muted); font-size: 0.85rem; font-weight: 600; cursor: pointer;
  transition: all 0.15s;
}
.fb-action-btn:hover { background: var(--bg); color: var(--dark); }
.fb-action-liked { color: var(--brand) !important; }
.fb-action-liked:hover { background: rgba(249,115,22,.08) !important; }
.fb-action-active { color: #3b82f6 !important; }

/* Sección comentarios */
.fb-comments-section {
  background: var(--bg); border-top: 1px solid var(--border-light);
  padding: 12px 16px; display: flex; flex-direction: column; gap: 10px;
}
.fb-comment { display: flex; gap: 9px; align-items: flex-start; }
.fb-comment-avatar {
  width: 30px; height: 30px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff; font-size: 0.7rem; font-weight: 800;
}
.fb-comment-avatar.me { background: linear-gradient(135deg, var(--brand), #ef4444); }
.fb-comment-bubble {
  background: var(--surface); border-radius: 0 var(--r-lg) var(--r-lg) var(--r-lg);
  padding: 8px 12px; max-width: calc(100% - 42px);
  border: 1px solid var(--border-light);
}
.fb-comment-author { font-size: 0.8rem; font-weight: 700; color: var(--dark); text-decoration: none; display: block; margin-bottom: 3px; }
.fb-comment-author:hover { color: var(--brand); }
.fb-comment-text { font-size: 0.84rem; color: var(--text); line-height: 1.5; white-space: pre-wrap; }
.fb-no-comments { font-size: 0.8rem; color: var(--muted); text-align: center; padding: 4px 0; }

/* Input nuevo comentario */
.fb-new-comment-row { display: flex; gap: 9px; align-items: center; padding-top: 4px; }
.fb-comment-input-wrap {
  flex: 1; display: flex; align-items: center; gap: 6px;
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: 999px; padding: 6px 10px 6px 14px;
  transition: border-color 0.15s;
}
.fb-comment-input-wrap:focus-within { border-color: var(--brand); }
.fb-comment-input {
  flex: 1; background: none; border: none; outline: none;
  font-size: 0.84rem; color: var(--dark);
}
.fb-comment-send {
  width: 28px; height: 28px; border-radius: 50%; border: none;
  background: var(--brand); color: #fff;
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; transition: background 0.15s; flex-shrink: 0;
}
.fb-comment-send:disabled { background: var(--border); cursor: default; }
.fb-comment-send:not(:disabled):hover { background: var(--brand-dark); }

@media (max-width: 768px) {
  .ver-curso-shell {
    min-height: auto;
  }
  .ver-layout {
    grid-template-columns: 1fr;
    min-height: auto;
  }
  .ver-mobile-toggle {
    display: flex;
  }
  .ver-sidebar {
    position: fixed; top: 0; left: 0; bottom: 0; width: 300px; z-index: 201;
    transform: translateX(-100%); transition: transform 0.3s cubic-bezier(0.25,0.1,0.25,1);
    border-right: none; max-height: none; overflow-y: auto;
    box-shadow: 4px 0 20px rgba(0,0,0,.15);
  }
  .ver-sidebar.open {
    transform: translateX(0);
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
  .fb-post-form-card {
    flex-direction: column;
    align-items: stretch;
  }
  .ver-next-suggestion {
    flex-direction: column;
    align-items: stretch;
  }
  .ver-welcome-title {
    font-size: 1.5rem !important;
  }
}

/* Focus Mode */
.ver-layout.focus-mode { grid-template-columns: 1fr; }
.ver-layout.focus-mode .ver-sidebar { display: none; }
.ver-layout.focus-mode .ver-main { width: 100%; }
.ver-layout.focus-mode .ver-main-inner { max-width: 860px; }
.ver-layout.focus-mode .ver-lec-header { max-width: 860px; }
.ver-layout.focus-mode .ver-content-card { box-shadow: var(--shadow-md); }
.ver-layout.focus-mode .ver-lesson-actions,
.ver-layout.focus-mode .ver-next-suggestion,
.ver-layout.focus-mode .ver-intermedias,
.ver-layout.focus-mode .ver-notes-section,
.ver-layout.focus-mode .ver-foro { max-width: 860px; }

/* Notes Section */
.ver-notes-section { margin-top: 24px; background: var(--surface); border: 1px solid var(--border); border-radius: var(--r); padding: 20px; }
.ver-notes-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.ver-section-title { display: flex; align-items: center; gap: 8px; font-size: 1rem; font-weight: 700; color: var(--dark); margin: 0; }
.ver-notes-status { font-size: 0.75rem; color: var(--muted); background: var(--surface-soft); padding: 4px 8px; border-radius: 4px; }
.ver-notes-input { font-size: 0.9rem; line-height: 1.5; resize: vertical; background: var(--bg); border: 1px solid var(--border-light); }
.ver-notes-input:focus { background: var(--surface); border-color: var(--brand); }

/* Glassmorphism icons */
.gm-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 8px; flex-shrink: 0;
  backdrop-filter: blur(8px);
}
.gm-icon-notes {
  background: rgba(249,115,22,.12);
  border: 1px solid rgba(249,115,22,.25);
  color: var(--brand);
}
.gm-icon-forum {
  background: rgba(59,130,246,.12);
  border: 1px solid rgba(59,130,246,.28);
  color: #3b82f6;
}
.gm-icon-brain {
  background: rgba(234,179,8,.12);
  border: 1px solid rgba(234,179,8,.28);
  color: #ca8a04;
}
.gm-icon-trophy {
  background: rgba(249,115,22,.15);
  border: 1px solid rgba(249,115,22,.3);
  color: var(--brand);
  width: 80px !important; height: 80px !important;
  border-radius: 20px;
}

/* Foro loading/error */
.foro-loading { display: flex; align-items: center; gap: 10px; padding: 20px; color: var(--muted); font-size: 0.88rem; justify-content: center; }
.foro-msg { padding: 10px 18px; font-size: 0.85rem; font-weight: 600; border-radius: var(--r-sm); margin: 8px 16px; }
.foro-msg-error { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }

/* Forum author link */
.ver-user-link { color: var(--brand-dark); font-weight: 600; text-decoration: none; transition: color 0.15s; }
.ver-user-link:hover { color: var(--brand); text-decoration: underline; }

/* Lec delete button */
.lec-btn.del {
  display: inline-flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border-radius: 6px; border: none;
  background: transparent; color: var(--muted); cursor: pointer; transition: all 0.15s;
  flex-shrink: 0;
}
.lec-btn.del:hover { background: #fef2f2; color: #dc2626; }

/* Confetti Overlay */
.ver-confetti-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.6); z-index: 10000;
  display: flex; align-items: center; justify-content: center;
  backdrop-filter: blur(4px);
}
.ver-confetti-card {
  background: var(--surface); padding: 40px; border-radius: 20px;
  text-align: center; max-width: 400px; width: 90%;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
  animation: popIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
.ver-confetti-icon { font-size: 4rem; margin-bottom: 16px; animation: bounce 2s infinite; }
.ver-confetti-card h2 { font-size: 1.8rem; font-weight: 900; color: var(--dark); margin-bottom: 8px; }
.ver-confetti-card p { font-size: 1rem; color: var(--muted); }

@keyframes popIn {
  0% { opacity: 0; transform: scale(0.8); }
  100% { opacity: 1; transform: scale(1); }
}
@keyframes bounce {
  0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
  40% { transform: translateY(-20px); }
  60% { transform: translateY(-10px); }
}

/* ── Forum Profile Card Popup ─────────────────────────── */
.foro-card-backdrop {
  position: fixed; inset: 0; z-index: 999;
  background: rgba(0,0,0,.45);
  display: flex; align-items: center; justify-content: center; padding: 20px;
}
.foro-profile-card {
  position: relative;
  background: var(--surface); border-radius: var(--r-xl);
  padding: 32px 24px 24px; text-align: center;
  box-shadow: 0 24px 60px rgba(0,0,0,.18);
  min-width: 240px; max-width: 300px; width: 100%;
}
.fpc-close {
  position: absolute; top: 12px; right: 12px;
  background: var(--surface-soft); border: none; cursor: pointer;
  color: var(--muted); border-radius: 50%; width: 28px; height: 28px;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.12s;
}
.fpc-close:hover { background: var(--border); }
.fpc-avatar {
  width: 72px; height: 72px; border-radius: 50%; margin: 0 auto 14px;
  background: linear-gradient(135deg, var(--brand), var(--brand-dark));
  color: #fff; font-size: 1.6rem; font-weight: 900;
  display: flex; align-items: center; justify-content: center;
}
.fpc-name { font-size: 1.05rem; font-weight: 800; color: var(--dark); margin-bottom: 4px; }
.fpc-role { font-size: 0.78rem; color: var(--muted); margin-bottom: 20px; }
.fpc-actions { display: flex; flex-direction: column; gap: 8px; }
.fpc-btn {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 10px 16px; border-radius: 9999px; font-size: 0.88rem; font-weight: 600;
  border: none; cursor: pointer; text-decoration: none; transition: all 0.15s;
}
.fpc-btn-primary { background: var(--brand); color: #fff; box-shadow: 0 4px 12px rgba(249,115,22,.25); }
.fpc-btn-primary:hover { background: var(--brand-dark); }
.fpc-btn-secondary { background: var(--surface-soft); color: var(--dark); }
.fpc-btn-secondary:hover { background: rgba(0,0,0,.06); }
</style>
