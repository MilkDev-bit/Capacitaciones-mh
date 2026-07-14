<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'
import EmptyState from '../../components/EmptyState.vue'
import { toast } from '../../utils/toast'
import { useCartStore } from '../../stores/cart'

const auth = useAuthStore()
const cartStore = useCartStore()
const router = useRouter()
const route = useRoute()

const capacitaciones = ref<any[]>([])
const cursosPublicos = ref<any[]>([])
const activeTab = ref<'mis' | 'explorar'>('mis')
const estadoFiltro = ref<'todos' | 'en-curso' | 'completados'>('todos')
const search = ref('')
const inscribiendose = ref<string | null>(null)
const loadingMis = ref(true)
const loadingPublicos = ref(true)
const loadError = ref('')

const codigoInput = ref('')
const codigoLoading = ref(false)
const codigoError = ref('')
const codigoSuccess = ref('')

const thumbClass: Record<string, string> = {
  video: 'thumb-video',
  document: 'thumb-document',
  text: 'thumb-text',
  link: 'thumb-link',
  course: 'thumb-default',
  mixto: 'thumb-default',
}
const typeIcon: Record<string, string> = { video: 'VID', document: 'PDF', text: 'TXT', link: 'URL', course: 'CUR', mixto: 'CUR' }
const typeLabel: Record<string, string> = {
  video: 'Video',
  document: 'Documento',
  text: 'Lectura',
  link: 'Enlace',
  course: 'Curso Completo',
  mixto: 'Curso Completo',
}

function normalize(text: string) {
  return (text || '').toLowerCase().trim()
}

function fileUrl(path: string) {
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
}

function courseProgress(curso: any) {
  if (!curso.total_lecciones) return 0
  return Math.round((curso.lecciones_completadas / curso.total_lecciones) * 100)
}

function hasMatch(curso: any, term: string) {
  const haystack = `${curso.title || ''} ${curso.description || ''} ${typeLabel[curso.type] || curso.type || ''}`
  return normalize(haystack).includes(term)
}

const cursosFiltrados = computed(() => {
  const term = normalize(search.value)
  return capacitaciones.value.filter((curso) => {
    const progreso = courseProgress(curso)
    const statusMatch =
      estadoFiltro.value === 'todos' ||
      (estadoFiltro.value === 'en-curso' && progreso < 100) ||
      (estadoFiltro.value === 'completados' && progreso === 100)

    return statusMatch && (!term || hasMatch(curso, term))
  })
})

const exploreSort = ref('reciente')
const explorePage = ref(1)
const EXPLORE_PAGE_SIZE = 12

const publicosFiltrados = computed(() => {
  const term = normalize(search.value)
  let list = cursosPublicos.value.filter((c: any) => !term || hasMatch(c, term))
  if (exploreSort.value === 'az') list = [...list].sort((a: any, b: any) => a.title.localeCompare(b.title))
  else if (exploreSort.value === 'za') list = [...list].sort((a: any, b: any) => b.title.localeCompare(a.title))
  return list
})

const totalExplorePages = computed(() =>
  Math.max(1, Math.ceil(publicosFiltrados.value.length / EXPLORE_PAGE_SIZE))
)
const publicosPaginados = computed(() => {
  const start = (explorePage.value - 1) * EXPLORE_PAGE_SIZE
  return publicosFiltrados.value.slice(start, start + EXPLORE_PAGE_SIZE)
})
watch([search, exploreSort], () => { explorePage.value = 1 })

const totalLecciones = computed(() =>
  capacitaciones.value.reduce((total, curso) => total + (curso.total_lecciones || 0), 0)
)
const leccionesCompletadas = computed(() =>
  capacitaciones.value.reduce((total, curso) => total + (curso.lecciones_completadas || 0), 0)
)
const cursosCompletados = computed(() =>
  capacitaciones.value.filter((curso) => courseProgress(curso) === 100).length
)
const promedioProgreso = computed(() => {
  if (!capacitaciones.value.length) return 0
  const total = capacitaciones.value.reduce((sum, curso) => sum + courseProgress(curso), 0)
  return Math.round(total / capacitaciones.value.length)
})
const cursoSugerido = computed(() => {
  return (
    capacitaciones.value.find((curso) => courseProgress(curso) > 0 && courseProgress(curso) < 100) ||
    capacitaciones.value.find((curso) => courseProgress(curso) === 0) ||
    capacitaciones.value[0]
  )
})

async function loadMis() {
  loadingMis.value = true
  loadError.value = ''
  try {
    const res = await api.get('/mis-capacitaciones')
    capacitaciones.value = res.data || []
  } catch (e: any) {
    loadError.value = e.response?.data?.error || 'No pudimos cargar tus cursos.'
  } finally {
    loadingMis.value = false
  }
}

async function loadPublicos() {
  loadingPublicos.value = true
  try {
    const res = await api.get('/cursos-publicos')
    cursosPublicos.value = res.data || []
  } finally {
    loadingPublicos.value = false
  }
}

async function verifySession(sessionId: string) {
  try {
    await api.post('/verify-checkout-session', { session_id: sessionId })
    toast.success('¡Compra procesada correctamente!')
    cartStore.clearCart()
  } catch (e: any) {
    const msg = e.response?.data?.error || ''
    if (!msg.includes('ya existe') && !msg.includes('conflict')) {
      console.warn('verify-checkout-session:', msg)
    }
  } finally {
    router.replace({ path: '/usuario/capacitaciones' })
  }
}

onMounted(async () => {
  const sessionId = route.query.session_id as string | undefined
  if (sessionId) {
    await verifySession(sessionId)
  }
  loadMis()
  loadPublicos()
})

async function openCourse(id: string) {
  const course = capacitaciones.value.find(c => c.id === id)
  if (course && course.type === 'videocall') {
    router.push(`/join?code=${course.codigo_acceso || ''}`)
  } else {
    router.push('/usuario/capacitaciones/' + id)
  }
}

function tramitarDC3(c: any) {
  const nombreCurso = c.title || ''
  const duracion = Math.ceil((c.duration || 60) / 60)
  const area = c.area_tematica || '6000'
  const url = `https://dc3.mhsolucionesempresariales.com/formulario-dc3-8f9d3a2b?nombre_curso=${encodeURIComponent(nombreCurso)}&duracion_horas=${duracion}&area_tematica=${encodeURIComponent(area)}`
  window.open(url, '_blank')
}

async function inscribirse(id: string) {
  inscribiendose.value = id
  try {
    await api.post(`/cursos/${id}/inscripciones`)
    await Promise.all([loadMis(), loadPublicos()])
    activeTab.value = 'mis'
  } finally {
    inscribiendose.value = null
  }
}

async function unirseConCodigo() {
  const code = codigoInput.value.trim().toUpperCase()
  if (!code) {
    codigoError.value = 'Ingresa un codigo'
    return
  }

  codigoError.value = ''
  codigoSuccess.value = ''
  codigoLoading.value = true

  try {
    if (code.startsWith('VC-')) {
      const res = await api.post('/videocalls/join', { codigo: code })
      if (res.data && res.data.token) {
        toast.success('Uniéndose a la videollamada...')
        router.push(`/usuario/videocall/${res.data.room_name}?codigo=${code}`)
      }
    } else {
      const res = await api.post('/inscripciones', { codigo: code })
      codigoSuccess.value = `Te uniste a "${res.data.title}"`
      codigoInput.value = ''
      await loadMis()
      setTimeout(() => {
        codigoSuccess.value = ''
        activeTab.value = 'mis'
      }, 1800)
    }
  } catch (e: any) {
    codigoError.value = e.response?.data?.error || 'Codigo invalido o en uso'
  } finally {
    codigoLoading.value = false
  }
}
</script>

<template>
  <div class="learning-page">
    <section class="learning-hero">
      <div class="learning-hero-copy">
        <span class="hero-kicker">Tu aula</span>
        <h1>Hola{{ auth.user?.name ? `, ${auth.user.name.split(' ')[0]}` : '' }}. Continua donde lo dejaste.</h1>
        <p>
          Revisa tu progreso, encuentra nuevas capacitaciones y entra al contenido con menos pasos.
        </p>
        <div class="hero-actions">
          <button v-if="cursoSugerido" class="btn btn-primary btn-lg" @click="openCourse(cursoSugerido.id)">
            Continuar curso
          </button>
          <button class="btn btn-secondary btn-lg" @click="activeTab = 'explorar'">
            Explorar catalogo
          </button>
        </div>
      </div>

      <div class="learning-stats" aria-label="Resumen de aprendizaje">
        <div class="learning-stat">
          <div class="stat-icon">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
          </div>
          <span>{{ capacitaciones.length }}</span>
          <p>Cursos inscritos</p>
        </div>
        <div class="learning-stat">
          <div class="stat-icon">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </div>
          <span>{{ leccionesCompletadas }}/{{ totalLecciones }}</span>
          <p>Lecciones completadas</p>
        </div>
        <div class="learning-stat">
          <div class="stat-icon">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </div>
          <span>{{ promedioProgreso }}%</span>
          <p>Avance promedio</p>
        </div>
      </div>
    </section>

    <div v-if="loadError" class="alert alert-error learning-alert">{{ loadError }}</div>

    <div class="learning-toolbar">
      <div class="tabs-bar learning-tabs" role="tablist" aria-label="Secciones de cursos">
        <button
          :class="['tab-pill', activeTab === 'mis' ? 'active' : '']"
          role="tab"
          :aria-selected="activeTab === 'mis'"
          @click="activeTab = 'mis'"
        >
          Mis cursos
          <span class="pill-count">{{ capacitaciones.length }}</span>
        </button>
        <button
          :class="['tab-pill', activeTab === 'explorar' ? 'active' : '']"
          role="tab"
          :aria-selected="activeTab === 'explorar'"
          @click="activeTab = 'explorar'"
        >
          Explorar
          <span class="pill-count">{{ cursosPublicos.length }}</span>
        </button>
      </div>

      <label class="course-search">
        <svg width="17" height="17" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
        </svg>
        <input v-model="search" placeholder="Buscar cursos, temas o formato" />
      </label>
    </div>

    <div v-if="activeTab === 'mis'" class="learning-section">
      <div class="learning-filters" aria-label="Filtrar cursos por estado">
        <button :class="{ active: estadoFiltro === 'todos' }" @click="estadoFiltro = 'todos'">
          Todos
        </button>
        <button :class="{ active: estadoFiltro === 'en-curso' }" @click="estadoFiltro = 'en-curso'">
          En curso
        </button>
        <button :class="{ active: estadoFiltro === 'completados' }" @click="estadoFiltro = 'completados'">
          Completados
          <span>{{ cursosCompletados }}</span>
        </button>
      </div>

      <div v-if="loadingMis" class="courses-grid">
        <article v-for="n in 6" :key="n" class="course-card course-card-skeleton">
          <div class="skeleton skel-thumb"></div>
          <div class="course-body">
            <div class="skeleton skel-title"></div>
            <div class="skeleton skel-text"></div>
            <div class="skeleton skel-text-sm"></div>
          </div>
        </article>
      </div>

      <div v-else-if="cursosFiltrados.length" class="courses-grid">
        <article
          v-for="(c, i) in cursosFiltrados"
          :key="c.id"
          class="course-card"
          :style="{ '--anim-delay': `${i * 60}ms` }"
          tabindex="0"
          @click="openCourse(c.id)"
          @keyup.enter="openCourse(c.id)"
        >
          <div :class="['course-thumb', c.thumbnail_url ? 'has-image' : (thumbClass[c.type] || 'thumb-default')]" :style="c.thumbnail_url ? '' : { background: c.color || '#f97316' }">
            <template v-if="c.thumbnail_url">
              <img :src="fileUrl(c.thumbnail_url)" alt="Portada del curso" class="course-thumb-img" />
            </template>
            <template v-else>
              <span class="thumb-icon">{{ typeIcon[c.type] || 'CUR' }}</span>
              <span class="course-cover-pill">{{ typeLabel[c.type] || c.type }}</span>
            </template>
          </div>
          <div class="course-body">
            <div class="course-card-top">
            </div>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripcion disponible.' }}</p>
            <div class="progress-wrap" v-if="c.type !== 'videocall'">
              <div class="progress-top">
                <span class="progress-label">{{ c.lecciones_completadas || 0 }}/{{ c.total_lecciones || 0 }} completadas</span>
                <span class="progress-pct">{{ courseProgress(c) }}%</span>
              </div>
              <div class="progress-bar-bg" aria-hidden="true">
                <div class="progress-bar-fill" :style="`width:${courseProgress(c)}%`" />
              </div>
            </div>
            
            <div class="course-code-display" v-if="c.type === 'videocall'">
              <span class="code-label">Tu código de acceso:</span>
              <div class="code-value">{{ c.codigo_acceso || 'Generando...' }}</div>
              <span class="code-hint">Deberás ingresarlo en la sala.</span>
            </div>

            <div v-if="courseProgress(c) === 100 && c.type !== 'videocall' && c.dc3_enabled === true" style="margin-top: 12px; margin-bottom: -4px;">
              <button
                class="btn btn-secondary btn-sm"
                style="width: 100%; display: flex; align-items: center; justify-content: center; gap: 6px; border-color: #f97316; color: #f97316; font-weight: 600;"
                @click.stop="tramitarDC3(c)"
              >
                📋 Tramitar Constancia DC-3
              </button>
            </div>

            <div class="course-cta">
              <template v-if="c.type === 'videocall'">
                Ir a la sala
              </template>
              <template v-else>
                {{ courseProgress(c) === 100 ? 'Repasar contenido' : 'Continuar aprendiendo' }}
              </template>
              <span aria-hidden="true">&rarr;</span>
            </div>
          </div>
        </article>
      </div>

      <EmptyState v-else class="learning-empty"
        :title="search ? 'No encontramos cursos con ese filtro' : 'Aun no tienes cursos activos'"
        :description="search ? 'Prueba con otro termino o cambia el estado del filtro.' : 'Explora el catalogo o usa un codigo de acceso de tu instructor.'"
      >
        <template #action>
          <button class="btn btn-primary" @click="activeTab = 'explorar'">Explorar cursos</button>
        </template>
      </EmptyState>
    </div>

    <div v-if="activeTab === 'explorar'" class="learning-section">
      <div class="code-banner code-banner-polished">
        <div class="code-banner-left">
          <div class="code-key-icon" aria-hidden="true">
            <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path d="M15.5 7.5a5 5 0 1 0 1 5.8L22 18.8V22h-3.2l-1.6-1.6H15v-2.2l-1.7-1.7" />
              <circle cx="8.5" cy="10.5" r="1.2" />
            </svg>
          </div>
          <div>
            <strong>Tienes un codigo de acceso?</strong>
            <p>Unete a cursos privados compartidos por tu instructor.</p>
          </div>
        </div>
        <div class="code-banner-right">
          <input
            v-model="codigoInput"
            class="code-field"
            placeholder="ABC123"
            maxlength="12"
            autocomplete="off"
            @keyup.enter="unirseConCodigo"
          />
          <button class="btn btn-primary" :disabled="codigoLoading" @click="unirseConCodigo">
            {{ codigoLoading ? 'Validando...' : 'Unirme' }}
          </button>
        </div>
      </div>
      <div v-if="codigoError" class="alert alert-error learning-alert">{{ codigoError }}</div>
      <div v-if="codigoSuccess" class="alert alert-success learning-alert">{{ codigoSuccess }}</div>

      <!-- Ordenamiento -->
      <div class="explore-toolbar">
        <div class="explore-sort">
          <select v-model="exploreSort" class="sort-select" aria-label="Ordenar por">
            <option value="reciente">Más reciente</option>
            <option value="az">A → Z</option>
            <option value="za">Z → A</option>
          </select>
        </div>
      </div>

      <div class="section-head">
        <div>
          <p class="section-label">Catálogo disponible</p>
          <span>{{ publicosFiltrados.length }} resultado{{ publicosFiltrados.length !== 1 ? 's' : '' }}{{ totalExplorePages > 1 ? ' · pág. ' + explorePage + '/' + totalExplorePages : '' }}</span>
        </div>
      </div>

      <div v-if="loadingPublicos" class="courses-grid">
        <article v-for="n in 6" :key="n" class="course-card course-card-skeleton">
          <div class="skeleton skel-thumb"></div>
          <div class="course-body">
            <div class="skeleton skel-title"></div>
            <div class="skeleton skel-text"></div>
          </div>
        </article>
      </div>

      <div v-else-if="publicosPaginados.length" class="courses-grid">
        <article
          v-for="(c, i) in publicosPaginados"
          :key="c.id"
          class="course-card public-course"
          :style="{ '--anim-delay': `${i * 50}ms` }"
        >
          <div :class="['course-thumb', c.thumbnail_url ? 'has-image' : (thumbClass[c.type] || 'thumb-default')]" :style="c.thumbnail_url ? '' : { background: c.color || '#f97316' }">
            <template v-if="c.thumbnail_url">
              <img :src="fileUrl(c.thumbnail_url)" alt="Portada del curso" class="course-thumb-img" />
            </template>
            <template v-else>
              <span class="thumb-icon">{{ typeIcon[c.type] || 'CUR' }}</span>
              <span class="course-cover-pill">{{ typeLabel[c.type] || c.type }}</span>
            </template>
            <span v-if="c.inscrito" class="enrolled-ribbon">Inscrito</span>
          </div>
          <div class="course-body">
            <div class="course-card-top">
              <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
              <span class="course-lessons">Libre acceso</span>
            </div>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripción disponible.' }}</p>
            <div class="course-footer-row">
              <span v-if="c.inscrito" class="badge badge-green">Ya inscrito</span>
              <button
                v-else-if="c.precio > 0"
                class="btn btn-primary btn-sm"
                @click.stop="openCourse(c.id)"
              >
                Comprar curso - ${{ c.precio }} MXN
              </button>
              <button
                v-else
                class="btn btn-primary btn-sm"
                :disabled="inscribiendose === c.id"
                @click.stop="inscribirse(c.id)"
              >
                {{ inscribiendose === c.id ? 'Inscribiendo...' : 'Inscribirse gratis' }}
              </button>
            </div>
          </div>
        </article>
      </div>

      <EmptyState v-else class="learning-empty"
        :title="search ? 'No encontramos cursos con ese término' : 'No hay cursos públicos disponibles'"
        :description="search ? 'Prueba con otro término de búsqueda.' : 'Pide a tu instructor que comparta un código de acceso.'"
      />

      <!-- Paginación -->
      <div v-if="!loadingPublicos && totalExplorePages > 1" class="explore-pagination" aria-label="Paginación">
        <button class="page-btn" :disabled="explorePage === 1" aria-label="Página anterior" @click="explorePage--">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <template v-for="p in totalExplorePages" :key="p">
          <button
            v-if="p === 1 || p === totalExplorePages || Math.abs(p - explorePage) <= 1"
            :class="['page-btn', explorePage === p ? 'active' : '']"
            @click="explorePage = p"
          >{{ p }}</button>
          <span v-else-if="p === explorePage - 2 || p === explorePage + 2" class="page-ellipsis">…</span>
        </template>
        <button class="page-btn" :disabled="explorePage === totalExplorePages" aria-label="Página siguiente" @click="explorePage++">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.learning-page {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.learning-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.8fr);
  gap: 22px;
  padding: 36px 44px;
  border-radius: var(--r-xl);
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 55%, #1c1917 100%);
  color: #fff;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  position: relative;
}
.learning-hero::before {
  content: '';
  position: absolute;
  top: -60px; right: -60px;
  width: 260px; height: 260px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(249,115,22,0.18) 0%, transparent 70%);
  pointer-events: none;
}

.learning-hero-copy {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 720px;
}

.hero-kicker {
  width: fit-content;
  padding: 4px 10px;
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.82);
  font-size: 0.76rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.learning-hero h1 {
  max-width: 640px;
  color: #fff;
  font-size: 2rem;
  font-weight: 900;
  line-height: 1.12;
}

.learning-hero p {
  max-width: 620px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 0.95rem;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 4px;
}

.hero-actions .btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.22);
  color: #fff;
}

.learning-stats {
  display: grid;
  gap: 10px;
  align-self: stretch;
}

.learning-stat {
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.07);
  backdrop-filter: blur(8px);
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: background 0.2s;
}
.learning-stat:hover {
  background: rgba(255, 255, 255, 0.12);
}
.stat-icon {
  width: 34px; height: 34px;
  border-radius: 8px;
  background: rgba(249,115,22,0.22);
  color: #fb923c;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 2px;
}
.learning-stat span {
  display: block;
  color: #fff;
  font-size: 1.4rem;
  font-weight: 900;
  line-height: 1;
}
.learning-stat p {
  margin: 0;
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.76rem;
  font-weight: 500;
}

.learning-alert {
  margin: 0;
}

.learning-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
}

.learning-tabs {
  margin: 0;
}

.course-search {
  min-width: min(100%, 360px);
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 10px 13px;
  border: 1.5px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--muted);
  box-shadow: var(--shadow-xs);
}

.course-search input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--text);
  font-size: 0.9rem;
}

.course-search:focus-within {
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(249, 115, 22, 0.12);
}

.learning-section {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.learning-filters {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.learning-filters button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 13px;
  border: 1.5px solid var(--border);
  border-radius: 999px;
  background: var(--surface);
  color: var(--muted);
  font-size: 0.84rem;
  font-weight: 800;
  white-space: nowrap;
  transition: border-color 0.16s, color 0.16s, background 0.16s;
}

.learning-filters button:hover,
.learning-filters button.active {
  border-color: var(--brand);
  background: var(--brand-light);
  color: var(--brand-dark);
}

.learning-filters span {
  min-width: 20px;
  padding: 1px 7px;
  border-radius: 999px;
  background: rgba(249, 115, 22, 0.14);
  color: var(--brand-dark);
  font-size: 0.74rem;
}

.course-card {
  border: 1px solid var(--border-light);
  border-radius: var(--r-lg);
  background: var(--surface);
  overflow: hidden;
  transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease;
  cursor: pointer;
  animation: fadeInUp 0.35s ease both;
  animation-delay: var(--anim-delay, 0ms);
}
.course-card:hover {
  transform: translateY(-4px);
  border-color: rgba(249, 115, 22, 0.45);
  box-shadow: 0 12px 28px rgba(0,0,0,0.12);
}

.course-card-skeleton {
  pointer-events: none;
  animation: none;
}

.course-thumb {
  display: flex;
  position: relative;
  overflow: hidden;
  height: 168px;
  align-items: flex-end;
  justify-content: flex-start;
  padding: 14px;
  border-radius: var(--r-lg) var(--r-lg) 0 0;
}
.course-thumb.has-image::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.45) 0%, rgba(0,0,0,0.08) 50%, transparent 100%);
  pointer-events: none;
  z-index: 1;
}
.course-thumb.has-image {
  background: #111;
}
.course-thumb-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  z-index: 0;
  transition: transform 0.4s ease;
}
.course-card:hover .course-thumb-img { transform: scale(1.06); }
.thumb-icon {
  position: relative;
  z-index: 2;
  color: rgba(255, 255, 255, 0.95);
  font-size: 1.85rem;
  font-weight: 900;
  letter-spacing: 0.02em;
}

.thumb-link {
  background: linear-gradient(135deg, #0f766e 0%, #2563eb 100%);
}

.course-cover-pill {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 3;
  padding: 4px 9px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.35);
  color: rgba(255, 255, 255, 0.92);
  font-size: 0.72rem;
  font-weight: 800;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255,255,255,0.15);
}

.enrolled-ribbon {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 3;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(249, 115, 22, 0.9);
  backdrop-filter: blur(6px);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  border: 1px solid rgba(255,255,255,0.2);
}

.course-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.course-lessons {
  color: var(--muted);
  font-size: 0.75rem;
  font-weight: 700;
  white-space: nowrap;
}

.course-title {
  min-height: 2.6em;
}

.course-cta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px solid var(--border-light);
  color: var(--brand-dark);
  font-weight: 700;
  font-size: 0.85rem;
}
.course-card:hover .course-cta { color: var(--brand); }

.code-banner-polished {
  border: 1px solid rgba(249, 115, 22, 0.22);
  border-left: 4px solid var(--brand);
  border-radius: 8px;
}

.code-key-icon {
  width: 42px;
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--brand-light);
  color: var(--brand-dark);
  flex-shrink: 0;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.section-head span {
  color: var(--muted);
  font-size: 0.82rem;
}

.learning-empty .empty-icon {
  min-width: 80px;
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--brand-light);
  color: var(--brand-dark);
  font-size: 0.82rem;
  font-weight: 900;
  text-transform: uppercase;
  font-weight: 700;
  color: #fff;
}

.course-code-display {
  background: rgba(15, 23, 42, 0.4);
  border: 1px dashed rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  padding: 12px;
  margin: 12px 0;
  text-align: center;
}

.course-code-display .code-label {
  display: block;
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.6);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 4px;
}

.course-code-display .code-value {
  font-family: monospace;
  font-size: 1.1rem;
  font-weight: 700;
  color: #e2e8f0;
  letter-spacing: 0.1em;
}

.course-code-display .code-hint {
  display: block;
  font-size: 0.7rem;
  color: rgba(255, 255, 255, 0.4);
  margin-top: 4px;
}

@media (max-width: 900px) {
  .learning-hero {
    grid-template-columns: 1fr;
  }

  .learning-stats {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 640px) {
  .learning-hero {
    padding: 22px 18px;
  }

  .learning-hero h1 {
    font-size: 1.55rem;
  }

  .learning-stats {
    grid-template-columns: 1fr;
  }

  .learning-toolbar,
  .course-search,
  .code-banner-right,
  .hero-actions .btn {
    width: 100%;
  }

  .code-banner-right {
    flex-direction: column;
    align-items: stretch;
  }

  .code-field {
    width: 100%;
  }
  .explore-toolbar { gap: 8px; }
  .explore-sort { margin-left: 0; width: 100%; }
  .sort-select { width: 100%; }
  .explore-pagination { flex-wrap: wrap; }
}

/* ── Explorar toolbar ─────────────────────── */
.explore-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.explore-type-filters {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.explore-filter-chip {
  display: inline-flex;
  align-items: center;
  padding: 7px 14px;
  border-radius: 999px;
  border: 1.5px solid var(--border);
  background: var(--surface);
  color: var(--muted);
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: border-color 0.16s, color 0.16s, background 0.16s;
}
.explore-filter-chip:hover,
.explore-filter-chip.active {
  border-color: var(--brand);
  background: var(--brand-light);
  color: var(--brand-dark);
}
.explore-sort { margin-left: auto; }
.sort-select {
  padding: 7px 12px;
  border: 1.5px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  font-size: 0.84rem;
  font-weight: 600;
  outline: none;
  cursor: pointer;
  transition: border-color 0.16s;
}
.sort-select:focus { border-color: var(--brand); }

/* ── Paginación ───────────────────────────── */
.explore-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding-top: 6px;
}
.page-btn {
  min-width: 38px;
  height: 38px;
  padding: 0 10px;
  border-radius: 8px;
  border: 1.5px solid var(--border);
  background: var(--surface);
  color: var(--text);
  font-size: 0.85rem;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.16s, color 0.16s, background 0.16s;
}
.page-btn:hover:not(:disabled):not(.active) {
  border-color: var(--brand);
  color: var(--brand-dark);
}
.page-btn.active {
  background: var(--brand);
  border-color: var(--brand);
  color: #fff;
  box-shadow: 0 3px 10px rgba(249,115,22,0.35);
}
.page-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.page-ellipsis {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  color: var(--muted);
  font-size: 0.9rem;
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>
