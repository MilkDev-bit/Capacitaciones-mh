<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'

const router = useRouter()
const cursosPublicos = ref<any[]>([])
const loading = ref(true)
const search = ref('')
const exploreSort = ref('reciente')
const explorePage = ref(1)
const EXPLORE_PAGE_SIZE = 12

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

function hasMatch(curso: any, term: string) {
  const haystack = `${curso.title || ''} ${curso.description || ''} ${typeLabel[curso.type] || curso.type || ''}`
  return normalize(haystack).includes(term)
}

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

async function loadPublicos() {
  loading.value = true
  try {
    const res = await api.get('/cursos-publicos')
    cursosPublicos.value = res.data || []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadPublicos()
})

function goToCourse(id: string) {
  router.push(`/curso/${id}`)
}
</script>

<template>
  <div class="store-layout">
    <header class="store-header">
      <div class="content-wrapper">
        <div class="brand" @click="router.push('/')">
          MH Soluciones Empresariales
        </div>
      </div>
    </header>

    <div class="store-hero">
      <div class="content-wrapper hero-content">
        <h1>Catálogo de Capacitaciones</h1>
        <p>Aprende nuevas habilidades con nuestros cursos creados por expertos en la materia.</p>
        
        <div class="search-bar">
          <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
          </svg>
          <input v-model="search" placeholder="Busca por título, tema o formato..." />
        </div>
      </div>
    </div>

    <main class="content-wrapper store-main">
      <div class="explore-toolbar">
        <div class="explore-sort">
          <select v-model="exploreSort" class="sort-select" aria-label="Ordenar por">
            <option value="reciente">Más reciente</option>
            <option value="az">A → Z</option>
            <option value="za">Z → A</option>
          </select>
        </div>
        <div class="results-count">
          <span>{{ publicosFiltrados.length }} resultado{{ publicosFiltrados.length !== 1 ? 's' : '' }}</span>
        </div>
      </div>

      <div v-if="loading" class="courses-grid">
        <article v-for="n in 6" :key="n" class="course-card skeleton">
          <div class="skel-thumb"></div>
          <div class="course-body">
            <div class="skel-title"></div>
            <div class="skel-text"></div>
          </div>
        </article>
      </div>

      <div v-else-if="publicosPaginados.length" class="courses-grid">
        <article
          v-for="c in publicosPaginados"
          :key="c.id"
          class="course-card"
          @click="goToCourse(c.id)"
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
              <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
            </div>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripción disponible.' }}</p>
            <div class="course-footer-row">
              <span v-if="c.precio > 0" class="price-tag">${{ c.precio }} MXN</span>
              <span v-else class="price-tag free">Gratis</span>
              <button class="btn btn-primary btn-sm">Ver detalles</button>
            </div>
          </div>
        </article>
      </div>

      <div v-else class="empty-state">
        <svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z"></path>
        </svg>
        <h3>No se encontraron cursos</h3>
        <p>Prueba con otro término de búsqueda.</p>
      </div>

      <div v-if="!loading && totalExplorePages > 1" class="explore-pagination" aria-label="Paginación">
        <button class="page-btn" :disabled="explorePage === 1" aria-label="Página anterior" @click="explorePage--">
          Anterior
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
          Siguiente
        </button>
      </div>
    </main>
  </div>
</template>

<style scoped>
.store-layout {
  min-height: 100vh;
  background: var(--bg);
}

.store-header {
  background: #fff;
  border-bottom: 1px solid var(--border-light);
  padding: 16px 0;
}

.content-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 32px;
}

.brand {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--dark);
  cursor: pointer;
}

.store-hero {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 55%, #1c1917 100%);
  color: #fff;
  padding: 64px 0;
  text-align: center;
}

.hero-content h1 {
  font-size: 3rem;
  font-weight: 900;
  margin: 0 0 16px 0;
}

.hero-content p {
  font-size: 1.15rem;
  color: rgba(255,255,255,0.8);
  max-width: 600px;
  margin: 0 auto 32px auto;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  border-radius: 999px;
  padding: 12px 24px;
  max-width: 600px;
  margin: 0 auto;
  color: var(--dark);
  box-shadow: 0 12px 24px rgba(0,0,0,0.1);
}

.search-bar svg {
  color: var(--muted);
}

.search-bar input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 1.1rem;
}

.store-main {
  padding-top: 48px;
  padding-bottom: 64px;
}

.explore-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.sort-select {
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: #fff;
  font-size: 0.95rem;
  outline: none;
}

.results-count {
  color: var(--muted);
  font-weight: 500;
}

.courses-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
}

.course-card {
  background: #fff;
  border: 1px solid var(--border-light);
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  display: flex;
  flex-direction: column;
}

.course-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0,0,0,0.08);
}

.course-thumb {
  height: 180px;
  position: relative;
  display: flex;
  align-items: flex-end;
  padding: 16px;
}

.course-thumb-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.4s ease;
}

.course-card:hover .course-thumb-img {
  transform: scale(1.05);
}

.has-image::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.5), transparent);
}

.thumb-icon {
  position: relative;
  z-index: 2;
  color: #fff;
  font-size: 2rem;
  font-weight: 900;
}

.course-cover-pill {
  position: absolute;
  top: 12px;
  left: 12px;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(4px);
  color: #fff;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
  z-index: 2;
}

.course-body {
  padding: 24px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.course-type-badge {
  font-size: 0.75rem;
  font-weight: 800;
  color: var(--brand);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 8px;
  display: block;
}

.course-title {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--dark);
  margin: 0 0 12px 0;
  line-height: 1.3;
}

.course-desc {
  font-size: 0.95rem;
  color: var(--muted);
  line-height: 1.5;
  margin: 0 0 24px 0;
  flex: 1;
}

.course-footer-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid var(--border-light);
  padding-top: 16px;
}

.price-tag {
  font-size: 1.15rem;
  font-weight: 900;
  color: var(--dark);
}

.price-tag.free {
  color: #10b981;
}

.empty-state {
  text-align: center;
  padding: 64px 0;
  color: var(--muted);
}

.empty-state svg {
  color: var(--border);
  margin-bottom: 16px;
}

.explore-pagination {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 48px;
}

.page-btn {
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: #fff;
  color: var(--dark);
  font-weight: 600;
  cursor: pointer;
  transition: 0.2s;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--brand);
  color: var(--brand);
}

.page-btn.active {
  background: var(--brand);
  color: #fff;
  border-color: var(--brand);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Skeletons */
.skeleton { pointer-events: none; }
.skel-thumb { height: 180px; background: var(--surface-soft); }
.skel-title { height: 24px; background: var(--surface-soft); border-radius: 4px; margin-bottom: 12px; width: 80%; }
.skel-text { height: 16px; background: var(--surface-soft); border-radius: 4px; width: 100%; margin-bottom: 8px; }
</style>
