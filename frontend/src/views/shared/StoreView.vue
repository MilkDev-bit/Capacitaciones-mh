<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import heroBg from '../../assets/store-hero-bg.png'

const router = useRouter()
const cursosPublicos = ref<any[]>([])
const loading = ref(true)
const search = ref('')
const exploreSort = ref('reciente')
const explorePage = ref(1)
const activeFilter = ref('todos')
const EXPLORE_PAGE_SIZE = 12

const typeLabel: Record<string, string> = {
  video: 'Video',
  document: 'Documento',
  text: 'Lectura',
  link: 'Enlace',
}

const typeEmoji: Record<string, string> = {
  video: '🎬',
  document: '📄',
  text: '📖',
  link: '🔗',
}

const typeGradient: Record<string, string> = {
  video:    'linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)',
  document: 'linear-gradient(135deg, #0d0d0d 0%, #1a1a1a 50%, #2d1b69 100%)',
  text:     'linear-gradient(135deg, #0f2027 0%, #203a43 50%, #2c5364 100%)',
  link:     'linear-gradient(135deg, #1a0533 0%, #2d1b69 50%, #11998e 100%)',
}

function normalize(text: string) {
  return (text || '').toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '').trim()
}

function fileUrl(path: string) {
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
}

function hasMatch(curso: any, term: string) {
  const haystack = `${curso.title || ''} ${curso.description || ''} ${typeLabel[curso.type] || curso.type || ''}`
  return normalize(haystack).includes(normalize(term))
}

const filterOptions = computed(() => {
  const counts: Record<string, number> = { todos: cursosPublicos.value.length }
  cursosPublicos.value.forEach((c: any) => {
    const key = c.type || 'otro'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const publicosFiltrados = computed(() => {
  const term = normalize(search.value)
  let list = cursosPublicos.value.filter((c: any) => {
    const matchSearch = !term || hasMatch(c, term)
    const matchFilter = activeFilter.value === 'todos' || c.type === activeFilter.value
    return matchSearch && matchFilter
  })
  if (exploreSort.value === 'az') list = [...list].sort((a: any, b: any) => a.title.localeCompare(b.title))
  else if (exploreSort.value === 'za') list = [...list].sort((a: any, b: any) => b.title.localeCompare(a.title))
  else if (exploreSort.value === 'precio_asc') list = [...list].sort((a: any, b: any) => (a.precio || 0) - (b.precio || 0))
  else if (exploreSort.value === 'precio_desc') list = [...list].sort((a: any, b: any) => (b.precio || 0) - (a.precio || 0))
  return list
})

const totalExplorePages = computed(() =>
  Math.max(1, Math.ceil(publicosFiltrados.value.length / EXPLORE_PAGE_SIZE))
)
const publicosPaginados = computed(() => {
  const start = (explorePage.value - 1) * EXPLORE_PAGE_SIZE
  return publicosFiltrados.value.slice(start, start + EXPLORE_PAGE_SIZE)
})

// Reset page when filter/search changes
watch([search, activeFilter, exploreSort], () => { explorePage.value = 1 })

async function loadPublicos() {
  loading.value = true
  try {
    const res = await api.get('/cursos-publicos')
    cursosPublicos.value = res.data || []
  } finally {
    loading.value = false
  }
}

onMounted(loadPublicos)

function goToCourse(id: string) {
  router.push(`/curso/${id}`)
}

function formatPrice(precio: number) {
  return new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN', maximumFractionDigits: 0 }).format(precio)
}
</script>

<template>
  <div class="store-root">

    <!-- ───── NAVBAR ───── -->
    <nav class="glass-nav">
      <div class="nav-inner">
        <button class="nav-brand" @click="router.push('/')">
          <span class="brand-dot"></span>
          MH Capacitaciones
        </button>
        <div class="nav-actions">
          <button class="nav-btn ghost" @click="router.push('/login')">Iniciar sesión</button>
          <button class="nav-btn primary" @click="router.push('/login')">Registrarse</button>
        </div>
      </div>
    </nav>

    <!-- ───── HERO ───── -->
    <section class="hero-section" :style="{ backgroundImage: `url(${heroBg})` }">
      <div class="hero-overlay"></div>
      <div class="hero-inner">
        <div class="hero-eyebrow">
          <span class="hero-pill">✦ Plataforma de Aprendizaje</span>
        </div>
        <h1 class="hero-title">
          Domina nuevas<br />
          <span class="hero-gradient-text">habilidades</span> hoy.
        </h1>
        <p class="hero-subtitle">
          Más de {{ cursosPublicos.length > 0 ? cursosPublicos.length : '∞' }} cursos profesionales creados por expertos en la industria.
          Aprende a tu ritmo, desde donde quieras.
        </p>

        <!-- Glass search bar -->
        <div class="hero-search-wrap">
          <div class="glass-search">
            <svg class="search-icon" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
            </svg>
            <input
              v-model="search"
              id="store-search"
              placeholder="Busca cursos, temas o formatos…"
              autocomplete="off"
            />
            <kbd v-if="!search" class="search-kbd">⌘K</kbd>
            <button v-else class="search-clear" @click="search = ''" aria-label="Limpiar búsqueda">✕</button>
          </div>
        </div>

        <!-- Stats pills -->
        <div class="hero-stats">
          <div class="stat-pill">
            <span class="stat-num">{{ cursosPublicos.filter((c:any) => !c.precio).length }}</span>
            <span class="stat-label">Cursos gratuitos</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-pill">
            <span class="stat-num">{{ cursosPublicos.filter((c:any) => c.precio > 0).length }}</span>
            <span class="stat-label">Cursos premium</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-pill">
            <span class="stat-num">100%</span>
            <span class="stat-label">Certificados</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ───── MAIN CATALOG ───── -->
    <main class="catalog-main">

      <!-- Filter strip -->
      <div class="filter-strip">
        <div class="filter-inner">
          <div class="filter-pills">
            <button
              :class="['filter-pill', activeFilter === 'todos' ? 'active' : '']"
              @click="activeFilter = 'todos'"
            >
              Todos
              <span class="pill-count">{{ filterOptions['todos'] || 0 }}</span>
            </button>
            <button
              v-if="filterOptions['video']"
              :class="['filter-pill', activeFilter === 'video' ? 'active' : '']"
              @click="activeFilter = 'video'"
            >
              🎬 Videos
              <span class="pill-count">{{ filterOptions['video'] }}</span>
            </button>
            <button
              v-if="filterOptions['document']"
              :class="['filter-pill', activeFilter === 'document' ? 'active' : '']"
              @click="activeFilter = 'document'"
            >
              📄 Documentos
              <span class="pill-count">{{ filterOptions['document'] }}</span>
            </button>
            <button
              v-if="filterOptions['text']"
              :class="['filter-pill', activeFilter === 'text' ? 'active' : '']"
              @click="activeFilter = 'text'"
            >
              📖 Lecturas
              <span class="pill-count">{{ filterOptions['text'] }}</span>
            </button>
          </div>

          <div class="sort-wrap">
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path d="M3 6h18M7 12h10M11 18h2"/>
            </svg>
            <select v-model="exploreSort" class="sort-select" aria-label="Ordenar resultados">
              <option value="reciente">Más recientes</option>
              <option value="az">A → Z</option>
              <option value="za">Z → A</option>
              <option value="precio_asc">Precio: menor a mayor</option>
              <option value="precio_desc">Precio: mayor a menor</option>
            </select>
          </div>
        </div>
      </div>

      <div class="catalog-inner">

        <!-- Results count -->
        <div class="results-bar">
          <p class="results-text">
            <strong>{{ publicosFiltrados.length }}</strong> resultado{{ publicosFiltrados.length !== 1 ? 's' : '' }}
            <span v-if="search"> para "<em>{{ search }}</em>"</span>
          </p>
        </div>

        <!-- Skeleton loading -->
        <div v-if="loading" class="courses-grid">
          <article v-for="n in 6" :key="n" class="course-card skeleton-card">
            <div class="skel-img"></div>
            <div class="skel-body">
              <div class="skel-badge"></div>
              <div class="skel-title"></div>
              <div class="skel-desc"></div>
              <div class="skel-desc short"></div>
              <div class="skel-footer"></div>
            </div>
          </article>
        </div>

        <!-- Course Grid -->
        <div v-else-if="publicosPaginados.length" class="courses-grid">
          <article
            v-for="c in publicosPaginados"
            :key="c.id"
            class="course-card"
            @click="goToCourse(c.id)"
            role="button"
            :aria-label="`Ver el curso ${c.title}`"
          >
            <!-- Thumbnail -->
            <div class="card-img-wrap">
              <img
                v-if="c.thumbnail_url"
                :src="fileUrl(c.thumbnail_url)"
                :alt="`Portada de ${c.title}`"
                class="card-img"
              />
              <div
                v-else
                class="card-img-placeholder"
                :style="{ background: typeGradient[c.type] || typeGradient['video'] }"
              >
                <span class="placeholder-emoji">{{ typeEmoji[c.type] || '📚' }}</span>
              </div>

              <!-- Overlay gradient -->
              <div class="card-img-overlay"></div>

              <!-- Type badge on image -->
              <span class="card-type-chip">{{ typeLabel[c.type] || 'Curso' }}</span>

              <!-- Free badge -->
              <span v-if="!c.precio || c.precio === 0" class="card-free-chip">Gratis</span>
            </div>

            <!-- Body -->
            <div class="card-body">
              <h3 class="card-title">{{ c.title }}</h3>
              <p class="card-desc">{{ c.description || 'Curso profesional certificado.' }}</p>

              <!-- Footer -->
              <div class="card-footer">
                <div class="card-price-block">
                  <span v-if="c.precio > 0" class="card-price">{{ formatPrice(c.precio) }}</span>
                  <span v-else class="card-price free">Gratis</span>
                </div>
                <button class="card-cta" @click.stop="goToCourse(c.id)">
                  Ver curso
                  <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                    <path d="M5 12h14M12 5l7 7-7 7"/>
                  </svg>
                </button>
              </div>
            </div>
          </article>
        </div>

        <!-- Empty state -->
        <div v-else class="empty-wrap">
          <div class="empty-glass">
            <div class="empty-icon">🔍</div>
            <h3>Sin resultados</h3>
            <p>Intenta con un término diferente o limpia los filtros.</p>
            <button class="glass-btn" @click="search = ''; activeFilter = 'todos'">Limpiar filtros</button>
          </div>
        </div>

        <!-- Pagination -->
        <nav v-if="!loading && totalExplorePages > 1" class="pagination" aria-label="Paginación del catálogo">
          <button
            class="page-btn"
            :disabled="explorePage === 1"
            @click="explorePage--"
            aria-label="Página anterior"
          >
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path d="M15 18l-6-6 6-6"/>
            </svg>
          </button>
          <template v-for="p in totalExplorePages" :key="p">
            <button
              v-if="p === 1 || p === totalExplorePages || Math.abs(p - explorePage) <= 1"
              :class="['page-num', explorePage === p ? 'active' : '']"
              @click="explorePage = p"
            >{{ p }}</button>
            <span v-else-if="p === explorePage - 2 || p === explorePage + 2" class="page-ellipsis">…</span>
          </template>
          <button
            class="page-btn"
            :disabled="explorePage === totalExplorePages"
            @click="explorePage++"
            aria-label="Página siguiente"
          >
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path d="M9 18l6-6-6-6"/>
            </svg>
          </button>
        </nav>
      </div>
    </main>

    <!-- ───── FOOTER ───── -->
    <footer class="store-footer">
      <p>© {{ new Date().getFullYear() }} MH Soluciones Empresariales · Todos los derechos reservados</p>
    </footer>

  </div>
</template>

<style scoped>
/* ── Base ──────────────────────────────────────────────── */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800;900&display=swap');

.store-root {
  min-height: 100vh;
  background: #08090a;
  color: #f1f5f9;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  -webkit-font-smoothing: antialiased;
}

/* ── NAV ───────────────────────────────────────────────── */
.glass-nav {
  position: fixed;
  top: 0; left: 0; right: 0;
  z-index: 100;
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  background: rgba(8, 9, 10, 0.72);
  border-bottom: 1px solid rgba(255,255,255,0.08);
}

.nav-inner {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 32px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  background: none;
  border: none;
  color: #f1f5f9;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  letter-spacing: -0.01em;
}

.brand-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: #f97316;
  box-shadow: 0 0 8px rgba(249,115,22,0.8);
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.nav-btn {
  padding: 8px 18px;
  border-radius: 9999px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.nav-btn.ghost {
  background: transparent;
  border: 1px solid rgba(255,255,255,0.15);
  color: rgba(255,255,255,0.8);
}

.nav-btn.ghost:hover {
  border-color: rgba(255,255,255,0.35);
  color: #fff;
}

.nav-btn.primary {
  background: rgba(249,115,22,1);
  border: 1px solid rgba(249,115,22,0.6);
  color: #fff;
  box-shadow: 0 0 16px rgba(249,115,22,0.3);
}

.nav-btn.primary:hover {
  background: rgba(234,88,12,1);
  box-shadow: 0 0 24px rgba(249,115,22,0.5);
}

/* ── HERO ──────────────────────────────────────────────── */
.hero-section {
  position: relative;
  padding-top: 160px;
  padding-bottom: 100px;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  overflow: hidden;
}

.hero-overlay {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(to bottom, rgba(8,9,10,0.35) 0%, rgba(8,9,10,0.55) 60%, rgba(8,9,10,1) 100%),
    radial-gradient(ellipse 60% 50% at 50% 40%, rgba(99,102,241,0.18), transparent);
  pointer-events: none;
}

.hero-inner {
  position: relative;
  z-index: 2;
  max-width: 780px;
  margin: 0 auto;
  text-align: center;
  padding: 0 24px;
}

.hero-eyebrow {
  margin-bottom: 24px;
}

.hero-pill {
  display: inline-block;
  padding: 6px 18px;
  border-radius: 9999px;
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.14);
  backdrop-filter: blur(12px);
  font-size: 0.8rem;
  font-weight: 600;
  color: rgba(255,255,255,0.8);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.hero-title {
  font-size: clamp(2.8rem, 6vw, 5rem);
  font-weight: 900;
  line-height: 1.05;
  letter-spacing: -0.03em;
  color: #fff;
  margin: 0 0 20px 0;
}

.hero-gradient-text {
  background: linear-gradient(135deg, #fb923c 0%, #f97316 40%, #a78bfa 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-subtitle {
  font-size: 1.1rem;
  color: rgba(255,255,255,0.65);
  line-height: 1.7;
  margin: 0 auto 40px;
  max-width: 560px;
  font-weight: 400;
}

/* Glass search */
.hero-search-wrap {
  margin-bottom: 40px;
}

.glass-search {
  display: flex;
  align-items: center;
  gap: 12px;
  backdrop-filter: blur(32px) saturate(200%);
  -webkit-backdrop-filter: blur(32px) saturate(200%);
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.18);
  border-radius: 16px;
  padding: 14px 20px;
  max-width: 580px;
  margin: 0 auto;
  transition: border-color 0.25s, box-shadow 0.25s;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.1);
}

.glass-search:focus-within {
  border-color: rgba(249,115,22,0.5);
  box-shadow: 0 8px 32px rgba(0,0,0,0.4), 0 0 0 3px rgba(249,115,22,0.15), inset 0 1px 0 rgba(255,255,255,0.1);
}

.search-icon {
  color: rgba(255,255,255,0.4);
  flex-shrink: 0;
}

.glass-search input {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  font-size: 1rem;
  color: #fff;
  font-family: inherit;
  font-weight: 400;
  caret-color: #f97316;
}

.glass-search input::placeholder {
  color: rgba(255,255,255,0.35);
}

.search-kbd {
  display: inline-block;
  padding: 3px 8px;
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 6px;
  font-size: 0.72rem;
  color: rgba(255,255,255,0.45);
  font-family: monospace;
  white-space: nowrap;
}

.search-clear {
  background: rgba(255,255,255,0.1);
  border: none;
  border-radius: 50%;
  width: 24px; height: 24px;
  color: rgba(255,255,255,0.6);
  cursor: pointer;
  font-size: 0.75rem;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.2s;
}

.search-clear:hover {
  background: rgba(255,255,255,0.2);
}

/* Stats */
.hero-stats {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
  flex-wrap: wrap;
}

.stat-pill {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 28px;
}

.stat-num {
  font-size: 1.75rem;
  font-weight: 800;
  color: #fff;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 0.75rem;
  color: rgba(255,255,255,0.45);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: 500;
}

.stat-divider {
  width: 1px;
  height: 32px;
  background: rgba(255,255,255,0.12);
}

/* ── FILTER STRIP ──────────────────────────────────────── */
.filter-strip {
  position: sticky;
  top: 60px;
  z-index: 50;
  backdrop-filter: blur(20px) saturate(160%);
  -webkit-backdrop-filter: blur(20px) saturate(160%);
  background: rgba(8,9,10,0.8);
  border-bottom: 1px solid rgba(255,255,255,0.07);
}

.filter-inner {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 32px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.filter-pills {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.filter-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 9999px;
  border: 1px solid rgba(255,255,255,0.1);
  background: rgba(255,255,255,0.04);
  color: rgba(255,255,255,0.55);
  font-size: 0.82rem;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-pill:hover {
  border-color: rgba(255,255,255,0.25);
  color: rgba(255,255,255,0.9);
}

.filter-pill.active {
  background: rgba(249,115,22,0.15);
  border-color: rgba(249,115,22,0.5);
  color: #fb923c;
}

.pill-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9999px;
  background: rgba(255,255,255,0.1);
  font-size: 0.7rem;
  font-weight: 700;
}

.filter-pill.active .pill-count {
  background: rgba(249,115,22,0.3);
}

.sort-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255,255,255,0.35);
  flex-shrink: 0;
}

.sort-select {
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: rgba(255,255,255,0.7);
  font-size: 0.82rem;
  font-family: inherit;
  padding: 6px 10px;
  outline: none;
  cursor: pointer;
  transition: border-color 0.2s;
}

.sort-select:hover {
  border-color: rgba(255,255,255,0.22);
}

.sort-select option {
  background: #1a1a2e;
}

/* ── CATALOG ───────────────────────────────────────────── */
.catalog-main {
  background: #08090a;
}

.catalog-inner {
  max-width: 1280px;
  margin: 0 auto;
  padding: 40px 32px 80px;
}

.results-bar {
  margin-bottom: 28px;
}

.results-text {
  font-size: 0.88rem;
  color: rgba(255,255,255,0.4);
  font-weight: 400;
}

.results-text strong {
  color: rgba(255,255,255,0.8);
  font-weight: 700;
}

.results-text em {
  color: rgba(249,115,22,0.8);
  font-style: normal;
}

/* ── GRID ──────────────────────────────────────────────── */
.courses-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

/* ── COURSE CARD ───────────────────────────────────────── */
.course-card {
  background: rgba(255,255,255,0.035);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 20px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.3s cubic-bezier(.34,1.56,.64,1), box-shadow 0.3s ease, border-color 0.3s ease;
  display: flex;
  flex-direction: column;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

.course-card:hover {
  transform: translateY(-6px) scale(1.01);
  box-shadow:
    0 24px 48px rgba(0,0,0,0.5),
    0 0 0 1px rgba(249,115,22,0.2),
    inset 0 1px 0 rgba(255,255,255,0.08);
  border-color: rgba(249,115,22,0.25);
}

/* Image area */
.card-img-wrap {
  position: relative;
  height: 200px;
  overflow: hidden;
}

.card-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s ease;
}

.course-card:hover .card-img {
  transform: scale(1.07);
}

.card-img-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.5s ease;
}

.course-card:hover .card-img-placeholder {
  transform: scale(1.05);
}

.placeholder-emoji {
  font-size: 3.5rem;
  filter: drop-shadow(0 4px 12px rgba(0,0,0,0.5));
}

.card-img-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, transparent 40%, rgba(8,9,10,0.8) 100%);
  pointer-events: none;
}

.card-type-chip {
  position: absolute;
  top: 12px;
  left: 12px;
  padding: 4px 11px;
  border-radius: 9999px;
  background: rgba(8,9,10,0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.12);
  font-size: 0.7rem;
  font-weight: 700;
  color: rgba(255,255,255,0.85);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.card-free-chip {
  position: absolute;
  top: 12px;
  right: 12px;
  padding: 4px 11px;
  border-radius: 9999px;
  background: rgba(16,185,129,0.2);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(16,185,129,0.4);
  font-size: 0.7rem;
  font-weight: 800;
  color: #34d399;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Card body */
.card-body {
  padding: 22px 22px 20px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.card-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: rgba(255,255,255,0.95);
  margin: 0 0 10px 0;
  line-height: 1.35;
  letter-spacing: -0.01em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-desc {
  font-size: 0.85rem;
  color: rgba(255,255,255,0.42);
  line-height: 1.6;
  margin: 0 0 20px 0;
  flex: 1;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 16px;
  border-top: 1px solid rgba(255,255,255,0.07);
}

.card-price {
  font-size: 1.2rem;
  font-weight: 800;
  color: rgba(255,255,255,0.95);
  letter-spacing: -0.02em;
}

.card-price.free {
  color: #34d399;
  font-size: 1rem;
  font-weight: 700;
}

.card-cta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(249,115,22,0.15);
  border: 1px solid rgba(249,115,22,0.35);
  border-radius: 9999px;
  color: #fb923c;
  font-size: 0.82rem;
  font-weight: 700;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s;
}

.card-cta:hover {
  background: rgba(249,115,22,0.28);
  border-color: rgba(249,115,22,0.6);
  box-shadow: 0 0 16px rgba(249,115,22,0.2);
}

/* ── SKELETON ──────────────────────────────────────────── */
.skeleton-card {
  pointer-events: none;
  background: rgba(255,255,255,0.03);
  border-color: rgba(255,255,255,0.05);
}

.skel-img {
  height: 200px;
  background: linear-gradient(90deg, rgba(255,255,255,0.04) 25%, rgba(255,255,255,0.08) 50%, rgba(255,255,255,0.04) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.6s infinite;
}

.skel-body {
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.skel-badge {
  height: 16px;
  width: 60px;
  border-radius: 999px;
  background: rgba(255,255,255,0.06);
  animation: shimmer 1.6s infinite;
}

.skel-title {
  height: 22px;
  width: 85%;
  border-radius: 6px;
  background: rgba(255,255,255,0.07);
  animation: shimmer 1.6s infinite;
}

.skel-desc {
  height: 14px;
  width: 100%;
  border-radius: 4px;
  background: rgba(255,255,255,0.04);
  animation: shimmer 1.6s infinite;
}

.skel-desc.short {
  width: 65%;
}

.skel-footer {
  height: 36px;
  border-radius: 8px;
  background: rgba(255,255,255,0.05);
  animation: shimmer 1.6s infinite;
  margin-top: 8px;
}

@keyframes shimmer {
  0%   { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

/* ── EMPTY STATE ───────────────────────────────────────── */
.empty-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
}

.empty-glass {
  text-align: center;
  padding: 48px 40px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 24px;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  max-width: 380px;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 16px;
}

.empty-glass h3 {
  font-size: 1.25rem;
  font-weight: 700;
  color: rgba(255,255,255,0.9);
  margin: 0 0 8px 0;
}

.empty-glass p {
  font-size: 0.9rem;
  color: rgba(255,255,255,0.45);
  margin: 0 0 24px 0;
  line-height: 1.6;
}

.glass-btn {
  padding: 10px 22px;
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 9999px;
  color: rgba(255,255,255,0.85);
  font-size: 0.88rem;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s;
}

.glass-btn:hover {
  background: rgba(255,255,255,0.14);
  border-color: rgba(255,255,255,0.28);
}

/* ── PAGINATION ────────────────────────────────────────── */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 56px;
}

.page-btn {
  width: 40px; height: 40px;
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.1);
  background: rgba(255,255,255,0.04);
  color: rgba(255,255,255,0.6);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-btn:hover:not(:disabled) {
  border-color: rgba(249,115,22,0.4);
  color: #fb923c;
}

.page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.page-num {
  width: 40px; height: 40px;
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.1);
  background: rgba(255,255,255,0.04);
  color: rgba(255,255,255,0.6);
  font-family: inherit;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.page-num:hover {
  border-color: rgba(249,115,22,0.4);
  color: #fb923c;
}

.page-num.active {
  background: rgba(249,115,22,0.2);
  border-color: rgba(249,115,22,0.55);
  color: #fb923c;
}

.page-ellipsis {
  color: rgba(255,255,255,0.25);
  width: 30px;
  text-align: center;
  font-size: 1rem;
}

/* ── FOOTER ────────────────────────────────────────────── */
.store-footer {
  border-top: 1px solid rgba(255,255,255,0.06);
  padding: 28px 32px;
  text-align: center;
}

.store-footer p {
  font-size: 0.8rem;
  color: rgba(255,255,255,0.22);
}

/* ── RESPONSIVE ────────────────────────────────────────── */
@media (max-width: 768px) {
  .nav-inner { padding: 0 20px; }
  .hero-section { padding-top: 110px; padding-bottom: 64px; }
  .hero-title { font-size: 2.4rem; }
  .hero-stats { gap: 0; }
  .stat-pill { padding: 0 16px; }
  .filter-inner { padding: 0 20px; flex-wrap: wrap; height: auto; padding-top: 10px; padding-bottom: 10px; gap: 8px; }
  .filter-pills { gap: 4px; }
  .catalog-inner { padding: 28px 20px 60px; }
  .courses-grid { grid-template-columns: 1fr 1fr; gap: 14px; }
  .card-img-wrap { height: 160px; }
}

@media (max-width: 480px) {
  .hero-subtitle { display: none; }
  .nav-btn.ghost { display: none; }
  .courses-grid { grid-template-columns: 1fr; }
}
</style>
