<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const router = useRouter()

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
}
const typeIcon: Record<string, string> = { video: 'VID', document: 'PDF', text: 'TXT', link: 'URL' }
const typeLabel: Record<string, string> = {
  video: 'Video',
  document: 'Documento',
  text: 'Lectura',
  link: 'Enlace',
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

const publicosFiltrados = computed(() => {
  const term = normalize(search.value)
  return cursosPublicos.value.filter((curso) => !term || hasMatch(curso, term))
})

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
    const cursos = res.data || []
    const cursosConProgreso = await Promise.all(
      cursos.map(async (curso: any) => {
        try {
          const lRes = await api.get(`/capacitaciones/${curso.id}/lecciones`)
          const lecciones = lRes.data || []
          return {
            ...curso,
            total_lecciones: lecciones.length,
            lecciones_completadas: lecciones.filter((l: any) => l.completada).length,
          }
        } catch {
          return { ...curso, total_lecciones: 0, lecciones_completadas: 0 }
        }
      })
    )
    capacitaciones.value = cursosConProgreso
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

onMounted(() => {
  loadMis()
  loadPublicos()
})

function openCourse(id: string) {
  router.push('/usuario/capacitaciones/' + id)
}

async function inscribirse(id: string) {
  inscribiendose.value = id
  try {
    await api.post(`/inscribirse/${id}`)
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
    const res = await api.post('/unirse-con-codigo', { codigo: code })
    codigoSuccess.value = `Te uniste a "${res.data.title}"`
    codigoInput.value = ''
    await loadMis()
    setTimeout(() => {
      codigoSuccess.value = ''
      activeTab.value = 'mis'
    }, 1800)
  } catch (e: any) {
    codigoError.value = e.response?.data?.error || 'Codigo invalido'
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
          <span>{{ capacitaciones.length }}</span>
          <p>Cursos inscritos</p>
        </div>
        <div class="learning-stat">
          <span>{{ leccionesCompletadas }}/{{ totalLecciones }}</span>
          <p>Lecciones completadas</p>
        </div>
        <div class="learning-stat">
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
          v-for="c in cursosFiltrados"
          :key="c.id"
          class="course-card"
          tabindex="0"
          @click="openCourse(c.id)"
          @keyup.enter="openCourse(c.id)"
        >
          <div :class="['course-thumb', c.thumbnail_url ? 'has-image' : (thumbClass[c.type] || 'thumb-default')]">
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
              <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
              <span class="course-lessons">{{ c.total_lecciones || 0 }} lecciones</span>
            </div>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripcion disponible.' }}</p>
            <div class="progress-wrap">
              <div class="progress-top">
                <span class="progress-label">{{ c.lecciones_completadas || 0 }}/{{ c.total_lecciones || 0 }} completadas</span>
                <span class="progress-pct">{{ courseProgress(c) }}%</span>
              </div>
              <div class="progress-bar-bg" aria-hidden="true">
                <div class="progress-bar-fill" :style="`width:${courseProgress(c)}%`" />
              </div>
            </div>
            <div class="course-cta">
              {{ courseProgress(c) === 100 ? 'Repasar contenido' : 'Continuar aprendiendo' }}
              <span aria-hidden="true">&rarr;</span>
            </div>
          </div>
        </article>
      </div>

      <div v-else class="empty-state learning-empty">
        <div class="empty-icon">Cursos</div>
        <h3>{{ search ? 'No encontramos cursos con ese filtro' : 'Aun no tienes cursos activos' }}</h3>
        <p>
          {{ search ? 'Prueba con otro termino o cambia el estado del filtro.' : 'Explora el catalogo o usa un codigo de acceso de tu instructor.' }}
        </p>
        <button class="btn btn-primary" @click="activeTab = 'explorar'">Explorar cursos</button>
      </div>
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

      <div class="section-head">
        <div>
          <p class="section-label">Catalogo disponible</p>
          <span>{{ publicosFiltrados.length }} resultado{{ publicosFiltrados.length !== 1 ? 's' : '' }}</span>
        </div>
      </div>

      <div v-if="loadingPublicos" class="courses-grid">
        <article v-for="n in 3" :key="n" class="course-card course-card-skeleton">
          <div class="skeleton skel-thumb"></div>
          <div class="course-body">
            <div class="skeleton skel-title"></div>
            <div class="skeleton skel-text"></div>
          </div>
        </article>
      </div>

      <div v-else-if="publicosFiltrados.length" class="courses-grid">
        <article v-for="c in publicosFiltrados" :key="c.id" class="course-card public-course">
          <div :class="['course-thumb', c.thumbnail_url ? 'has-image' : (thumbClass[c.type] || 'thumb-default')]">
            <template v-if="c.thumbnail_url">
              <img :src="fileUrl(c.thumbnail_url)" alt="Portada del curso" class="course-thumb-img" />
              <span v-if="c.inscrito" class="enrolled-ribbon">Inscrito</span>
            </template>
            <template v-else>
              <span class="thumb-icon">{{ typeIcon[c.type] || 'CUR' }}</span>
              <span v-if="c.inscrito" class="enrolled-ribbon">Inscrito</span>
            </template>
          </div>
          <div class="course-body">
            <div class="course-card-top">
              <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
              <span class="course-lessons">Libre acceso</span>
            </div>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripcion disponible.' }}</p>
            <div class="course-footer-row">
              <span v-if="c.inscrito" class="badge badge-green">Ya inscrito</span>
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

      <div v-else class="empty-state learning-empty">
        <div class="empty-icon">Buscar</div>
        <h3>No hay cursos publicos disponibles</h3>
        <p>Pide a tu instructor que comparta un enlace o codigo de acceso.</p>
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
  padding: 32px 40px;
  border-radius: var(--r-xl);
  background: linear-gradient(135deg, var(--dark) 0%, #374151 100%);
  color: #fff;
  overflow: hidden;
  box-shadow: var(--shadow-md);
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
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.08);
}

.learning-stat span {
  display: block;
  color: #fff;
  font-size: 1.45rem;
  font-weight: 900;
}

.learning-stat p {
  margin-top: 2px;
  color: rgba(255, 255, 255, 0.66);
  font-size: 0.8rem;
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
  color: var(--dark);
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
}

.course-card:hover {
  border-color: rgba(249, 115, 22, 0.45);
}

.course-card-skeleton {
  pointer-events: none;
}

.course-thumb {
  height: 148px;
  align-items: flex-end;
  justify-content: flex-start;
  padding: 16px;
}

.course-thumb.has-image {
  background: none;
}
.course-thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.thumb-icon {
  color: rgba(255, 255, 255, 0.94);
  font-size: 1.85rem;
  font-weight: 900;
  letter-spacing: 0.02em;
  filter: none;
}

.thumb-link {
  background: linear-gradient(135deg, #0f766e 0%, #2563eb 100%);
}

.course-cover-pill {
  position: absolute;
  top: 12px;
  left: 12px;
  padding: 4px 9px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  color: rgba(255, 255, 255, 0.88);
  font-size: 0.72rem;
  font-weight: 800;
  backdrop-filter: blur(6px);
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
  padding-top: 8px;
  border-top: 1px solid var(--border-light);
}

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
  letter-spacing: 0.06em;
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
}
</style>
