<script setup lang="ts">
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'
import { useCartStore } from '../../stores/cart'
import heroBg from '../../assets/store-hero-bg.png'
import logoSrc from '../../assets/logo-capacitaciones.png'

const router = useRouter()
const auth = useAuthStore()
const cart = useCartStore()

// ── Scroll-based nav opacity ──────────────────────────────
const scrolled = ref(false)
function onScroll() { scrolled.value = window.scrollY > 40 }
onMounted(() => window.addEventListener('scroll', onScroll, { passive: true }))
onUnmounted(() => window.removeEventListener('scroll', onScroll))

// ── Card entry stagger ───────────────────────────────────
const cardsVisible = ref(false)
const cursosPublicos = ref<any[]>([])
const loading = ref(true)
const search = ref('')
const exploreSort = ref('reciente')
const explorePage = ref(1)
const activeFilter = ref('todos')
const priceFilter = ref('todos') // 'todos' | 'gratis' | 'pago'
const EXPLORE_PAGE_SIZE = 12

const typeLabel: Record<string, string> = {
  video: 'Video',
  document: 'Documento',
  text: 'Lectura',
  link: 'Enlace',
}

const typeIcon: Record<string, string> = {
  video: '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>',
  document: '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>',
  text: '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>',
  link: '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>',
}
const defaultIcon = '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20"/></svg>'

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
    const matchPrice =
      priceFilter.value === 'todos' ||
      (priceFilter.value === 'gratis' && (!c.precio || c.precio === 0)) ||
      (priceFilter.value === 'pago'   && c.precio > 0)
    return matchSearch && matchFilter && matchPrice
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

// Reset page when filter/search/priceFilter changes
watch([search, activeFilter, exploreSort, priceFilter], () => { explorePage.value = 1 })

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
  // trigger stagger after a tick
  setTimeout(() => { cardsVisible.value = true }, 80)
})

watch(publicosPaginados, () => {
  cardsVisible.value = false
  setTimeout(() => { cardsVisible.value = true }, 50)
})

function goToCourse(id: string) {
  router.push(`/curso/${id}`)
}

function formatPrice(precio: number) {
  return new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN', maximumFractionDigits: 0 }).format(precio)
}

// ── 3-D card tilt ────────────────────────────────────────
function onCardEnter(e: MouseEvent) {
  const card = e.currentTarget as HTMLElement
  card.style.transition = 'transform 0.1s ease, box-shadow 0.2s ease, border-color 0.2s ease'
}
function onCardLeave(e: MouseEvent) {
  const card = e.currentTarget as HTMLElement
  card.style.transform = ''
  card.style.transition = 'transform 0.4s ease, box-shadow 0.3s ease, border-color 0.3s ease'
}
function onCardMove(e: MouseEvent) {
  const card = e.currentTarget as HTMLElement
  const rect = card.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  const cx = rect.width / 2
  const cy = rect.height / 2
  const rotateX = ((y - cy) / cy) * -6
  const rotateY = ((x - cx) / cx) * 6
  card.style.transform = `perspective(800px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateY(-6px) scale(1.01)`
}
</script>

<template>
  <div class="store-root">

    <!-- ───── PARTICLES ───── -->
    <div class="particles" aria-hidden="true">
      <span v-for="n in 18" :key="n" :class="`p p-${n}`"></span>
    </div>

    <!-- ───── NAVBAR ───── -->
    <nav :class="['glass-nav', scrolled ? 'nav-scrolled' : '']">
      <div class="nav-inner">
        <button class="nav-brand" @click="router.push('/')">
          <img :src="logoSrc" alt="Logo MH" class="nav-logo" />
          <span class="brand-name">MH <span class="brand-accent">Capacitaciones</span></span>
        </button>
        <div class="nav-actions">
          <template v-if="auth.isLoggedIn">
            <button class="nav-btn ghost" @click="cart.openDrawer">
              🛒 Carrito ({{ cart.totalItems }})
            </button>
            <button class="nav-btn primary" @click="router.push(auth.user?.role === 'instructor' ? '/instructor' : '/usuario')">
              Mi Panel
            </button>
          </template>
          <template v-else>
            <button class="nav-btn ghost" @click="cart.openDrawer">
              🛒 Carrito ({{ cart.totalItems }})
            </button>
            <button class="nav-btn ghost" @click="router.push('/login')">Iniciar sesión</button>
            <button class="nav-btn primary" @click="router.push('/login')">Registrarse</button>
          </template>
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
              <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              Videos
              <span class="pill-count">{{ filterOptions['video'] }}</span>
            </button>
            <button
              v-if="filterOptions['document']"
              :class="['filter-pill', activeFilter === 'document' ? 'active' : '']"
              @click="activeFilter = 'document'"
            >
              <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
              Documentos
              <span class="pill-count">{{ filterOptions['document'] }}</span>
            </button>
            <button
              v-if="filterOptions['text']"
              :class="['filter-pill', activeFilter === 'text' ? 'active' : '']"
              @click="activeFilter = 'text'"
            >
              <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
              Lecturas
              <span class="pill-count">{{ filterOptions['text'] }}</span>
            </button>

            <!-- Price filters -->
            <div class="filter-divider"></div>
            <button
              :class="['filter-pill price-pill', priceFilter === 'gratis' ? 'active green' : '']"
              @click="priceFilter = priceFilter === 'gratis' ? 'todos' : 'gratis'"
            >
              🆓 Gratis
              <span class="pill-count">{{ cursosPublicos.filter((c:any) => !c.precio || c.precio === 0).length }}</span>
            </button>
            <button
              :class="['filter-pill price-pill', priceFilter === 'pago' ? 'active orange' : '']"
              @click="priceFilter = priceFilter === 'pago' ? 'todos' : 'pago'"
            >
              <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
              Premium
              <span class="pill-count">{{ cursosPublicos.filter((c:any) => c.precio > 0).length }}</span>
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
            <span v-if="priceFilter !== 'todos'"> · <button class="clear-chip" @click="priceFilter='todos'">{{ priceFilter === 'gratis' ? 'Gratis' : 'Premium' }} ✕</button></span>
            <span v-if="activeFilter !== 'todos'"> · <button class="clear-chip" @click="activeFilter='todos'">{{ activeFilter }} ✕</button></span>
          </p>
        </div>

        <!-- Skeleton loading -->
        <div v-if="loading" class="courses-grid">
          <article v-for="n in 6" :key="n" class="course-card skeleton-card" :style="{ '--i': n }">
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
        <TransitionGroup
          v-else-if="publicosPaginados.length"
          tag="div"
          class="courses-grid"
          name="card-stagger"
        >
          <article
            v-for="(c, idx) in publicosPaginados"
            :key="c.id"
            class="course-card"
            :class="{ 'card-visible': cardsVisible }"
            :style="{ '--i': idx }"
            @click="goToCourse(c.id)"
            @mouseenter="onCardEnter($event)"
            @mouseleave="onCardLeave($event)"
            @mousemove="onCardMove($event)"
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
                <div class="glass-icon" v-html="typeIcon[c.type] || defaultIcon"></div>
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
        </TransitionGroup>

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
  gap: 10px;
  background: none;
  border: none;
  color: #f1f5f9;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  letter-spacing: -0.01em;
  transition: opacity 0.2s;
}

.nav-brand:hover {
  opacity: 0.85;
}

.nav-logo {
  width: 34px;
  height: 34px;
  object-fit: contain;
  filter: drop-shadow(0 0 8px rgba(249,115,22,0.55));
  animation: logo-pulse 3s ease-in-out infinite;
}

@keyframes logo-pulse {
  0%, 100% { filter: drop-shadow(0 0 8px rgba(249,115,22,0.5)); }
  50%       { filter: drop-shadow(0 0 14px rgba(249,115,22,0.9)); }
}

.brand-name {
  font-size: 1.02rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.brand-accent {
  color: #fb923c;
}

/* scrolled nav gets a stronger bg */
.nav-scrolled {
  background: rgba(8,9,10,0.92);
  box-shadow: 0 1px 24px rgba(0,0,0,0.5);
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

.btn-icon {
  width: 16px;
  height: 16px;
  margin-right: 6px;
  display: inline-block;
  vertical-align: middle;
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

.filter-pill.active.green {
  background: rgba(52,211,153,0.15);
  border-color: rgba(52,211,153,0.45);
  color: #34d399;
  box-shadow: 0 0 10px rgba(52,211,153,0.15);
}

.filter-pill.active.orange {
  background: rgba(249,115,22,0.15);
  border-color: rgba(249,115,22,0.5);
  color: #fb923c;
  box-shadow: 0 0 10px rgba(249,115,22,0.15);
}

.filter-divider {
  width: 1px;
  height: 20px;
  background: rgba(255,255,255,0.1);
  margin: 0 4px;
  flex-shrink: 0;
}

.clear-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(255,255,255,0.07);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 9999px;
  color: rgba(255,255,255,0.6);
  font-size: 0.78rem;
  font-family: inherit;
  padding: 2px 8px;
  cursor: pointer;
  transition: all 0.2s;
}
.clear-chip:hover {
  background: rgba(255,255,255,0.14);
  color: #fff;
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

.glass-icon {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  padding: 12px;
  border-radius: 50%;
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
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

/* ── PARTICLES ─────────────────────────────────────────── */
.particles {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.p {
  position: absolute;
  border-radius: 50%;
  opacity: 0;
  animation: float-particle linear infinite;
}

/* Individual particle sizes, positions, delays */
.p-1  { width: 4px;  height: 4px;  background: rgba(249,115,22,.7);  left: 8%;  animation-duration: 14s; animation-delay: 0s;   }
.p-2  { width: 2px;  height: 2px;  background: rgba(255,255,255,.5); left: 16%; animation-duration: 18s; animation-delay: 2s;   }
.p-3  { width: 3px;  height: 3px;  background: rgba(167,139,250,.6); left: 23%; animation-duration: 12s; animation-delay: 5s;   }
.p-4  { width: 5px;  height: 5px;  background: rgba(249,115,22,.5);  left: 30%; animation-duration: 20s; animation-delay: 1s;   }
.p-5  { width: 2px;  height: 2px;  background: rgba(255,255,255,.4); left: 38%; animation-duration: 15s; animation-delay: 7s;   }
.p-6  { width: 4px;  height: 4px;  background: rgba(52,211,153,.6);  left: 45%; animation-duration: 11s; animation-delay: 3s;   }
.p-7  { width: 3px;  height: 3px;  background: rgba(249,115,22,.8);  left: 53%; animation-duration: 17s; animation-delay: 9s;   }
.p-8  { width: 2px;  height: 2px;  background: rgba(255,255,255,.5); left: 60%; animation-duration: 13s; animation-delay: 4s;   }
.p-9  { width: 6px;  height: 6px;  background: rgba(167,139,250,.4); left: 67%; animation-duration: 22s; animation-delay: 6s;   }
.p-10 { width: 3px;  height: 3px;  background: rgba(249,115,22,.6);  left: 74%; animation-duration: 16s; animation-delay: 0.5s; }
.p-11 { width: 2px;  height: 2px;  background: rgba(255,255,255,.3); left: 81%; animation-duration: 19s; animation-delay: 8s;   }
.p-12 { width: 4px;  height: 4px;  background: rgba(52,211,153,.5);  left: 88%; animation-duration: 10s; animation-delay: 2.5s; }
.p-13 { width: 3px;  height: 3px;  background: rgba(249,115,22,.7);  left: 12%; animation-duration: 21s; animation-delay: 11s;  }
.p-14 { width: 5px;  height: 5px;  background: rgba(255,255,255,.2); left: 27%; animation-duration: 14s; animation-delay: 13s;  }
.p-15 { width: 2px;  height: 2px;  background: rgba(167,139,250,.7); left: 50%; animation-duration: 9s;  animation-delay: 0.8s; }
.p-16 { width: 4px;  height: 4px;  background: rgba(249,115,22,.4);  left: 71%; animation-duration: 23s; animation-delay: 15s;  }
.p-17 { width: 3px;  height: 3px;  background: rgba(52,211,153,.6);  left: 91%; animation-duration: 12s; animation-delay: 6.5s; }
.p-18 { width: 2px;  height: 2px;  background: rgba(255,255,255,.4); left: 4%;  animation-duration: 17s; animation-delay: 10s;  }

@keyframes float-particle {
  0%   { transform: translateY(100vh) scale(0); opacity: 0; }
  10%  { opacity: 1; }
  90%  { opacity: 0.6; }
  100% { transform: translateY(-120px) scale(1.4) rotate(360deg); opacity: 0; }
}

/* ── HERO AURORA ────────────────────────────────────────── */
.hero-section::before {
  content: '';
  position: absolute;
  top: -60px; left: 50%;
  transform: translateX(-50%);
  width: 800px; height: 400px;
  background: radial-gradient(ellipse, rgba(249,115,22,0.12) 0%, rgba(167,139,250,0.08) 40%, transparent 70%);
  filter: blur(60px);
  animation: aurora-pulse 6s ease-in-out infinite alternate;
  pointer-events: none;
  z-index: 1;
}

@keyframes aurora-pulse {
  0%   { opacity: 0.6; transform: translateX(-50%) scale(1); }
  100% { opacity: 1;   transform: translateX(-50%) scale(1.15); }
}

/* ── HERO CONTENT ANIMATION ─────────────────────────────── */
.hero-eyebrow   { animation: hero-fade-up 0.7s ease both; animation-delay: 0.1s; }
.hero-title     { animation: hero-fade-up 0.7s ease both; animation-delay: 0.25s; }
.hero-subtitle  { animation: hero-fade-up 0.7s ease both; animation-delay: 0.4s; }
.hero-search-wrap { animation: hero-fade-up 0.7s ease both; animation-delay: 0.55s; }
.hero-stats     { animation: hero-fade-up 0.7s ease both; animation-delay: 0.7s; }

@keyframes hero-fade-up {
  from { opacity: 0; transform: translateY(28px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* ── CARD STAGGER ───────────────────────────────────────── */
.course-card {
  opacity: 0;
  transform: translateY(24px);
  transition:
    opacity 0.5s ease,
    transform 0.5s cubic-bezier(.34,1.2,.64,1),
    box-shadow 0.3s ease,
    border-color 0.3s ease;
  transition-delay: calc(var(--i, 0) * 60ms);
}

.course-card.card-visible {
  opacity: 1;
  transform: translateY(0);
}

/* ── CARD GLOW ON HOVER ─────────────────────────────────── */
.course-card:hover {
  box-shadow:
    0 24px 48px rgba(0,0,0,0.55),
    0 0 0 1px rgba(249,115,22,0.25),
    0 0 40px rgba(249,115,22,0.08),
    inset 0 1px 0 rgba(255,255,255,0.09);
  border-color: rgba(249,115,22,0.3);
}

/* ── CARD CTA PULSE ─────────────────────────────────────── */
.card-cta {
  position: relative;
  overflow: hidden;
}

.card-cta::after {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 50%, rgba(255,255,255,0.2), transparent 70%);
  opacity: 0;
  transition: opacity 0.3s;
}

.card-cta:hover::after {
  opacity: 1;
}

/* ── SEARCH BAR ANIMATED BORDER ─────────────────────────── */
.glass-search {
  position: relative;
}

.glass-search::before {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: 17px;
  background: linear-gradient(135deg, rgba(249,115,22,0), rgba(167,139,250,0), rgba(249,115,22,0));
  opacity: 0;
  transition: opacity 0.4s;
  z-index: -1;
}

.glass-search:focus-within::before {
  background: linear-gradient(135deg, rgba(249,115,22,0.5), rgba(167,139,250,0.4), rgba(249,115,22,0.5));
  opacity: 1;
  animation: border-spin 3s linear infinite;
}

@keyframes border-spin {
  0%   { background-position: 0% 50%; }
  50%  { background-position: 100% 50%; }
   100% { background-position: 0% 50%; }
}

/* ── STAT NUMBER COUNT-UP FEEL ──────────────────────────── */
.stat-num {
  display: inline-block;
  animation: stat-pop 0.6s cubic-bezier(.34,1.56,.64,1) both;
  animation-delay: 0.8s;
}

@keyframes stat-pop {
  from { transform: scale(0.5); opacity: 0; }
  to   { transform: scale(1);   opacity: 1; }
}

/* ── CARD SHIMMER ON HOVER ──────────────────────────────── */
.card-img-wrap::after {
  content: '';
  position: absolute;
  top: -100%;
  left: -60%;
  width: 60%;
  height: 300%;
  background: linear-gradient(105deg, transparent 40%, rgba(255,255,255,0.12) 50%, transparent 60%);
  transform: skewX(-15deg);
  transition: left 0s;
  pointer-events: none;
  z-index: 3;
}

.course-card:hover .card-img-wrap::after {
  left: 160%;
  transition: left 0.6s ease;
}

/* ── FILTER PILL ACTIVE GLOW ────────────────────────────── */
.filter-pill.active {
  background: rgba(249,115,22,0.15);
  border-color: rgba(249,115,22,0.5);
  color: #fb923c;
  box-shadow: 0 0 12px rgba(249,115,22,0.2);
}

/* ── CARD STAGGER TRANSITION (Vue TransitionGroup) ──────── */
.card-stagger-enter-active { transition: all 0.4s ease; }
.card-stagger-leave-active { transition: all 0.2s ease; }
.card-stagger-enter-from  { opacity: 0; transform: translateY(20px) scale(0.97); }
.card-stagger-leave-to    { opacity: 0; transform: scale(0.96); }

/* ── LOGO GLOW ON HERO SECTION ──────────────────────────── */
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
