<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'
import { useCartStore } from '../../stores/cart'
import { useTheme } from '../../composables/useTheme'
import { toast } from '../../utils/toast'
import CourseTile, { type StoreCourse } from '../../components/store/CourseTile.vue'
import CourseCarousel from '../../components/store/CourseCarousel.vue'
import logoSrc from '../../assets/logo-capacitaciones.png'

const router = useRouter()
const auth = useAuthStore()
const cart = useCartStore()
const { isDark, toggleTheme } = useTheme()

const cursos = ref<StoreCourse[]>([])
const loading = ref(true)

// ── Filtros del catálogo ────────────────────────────────────────────────────
const search = ref('')
const tipoFiltro = ref('todos')
const precioFiltro = ref<'todos' | 'gratis' | 'pago'>('todos')
const orden = ref('reciente')
const pagina = ref(1)
const POR_PAGINA = 12

const TIPOS = [
  { key: 'todos', label: 'Todos' },
  { key: 'video', label: 'Video' },
  { key: 'document', label: 'Documentos' },
  { key: 'text', label: 'Lecturas' },
]

// ── Header ──────────────────────────────────────────────────────────────────
const scrolled = ref(false)
const menuAbierto = ref(false)
function onScroll() { scrolled.value = window.scrollY > 20 }

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
  cargar()
})
onUnmounted(() => window.removeEventListener('scroll', onScroll))

async function cargar() {
  loading.value = true
  try {
    const res = await api.get('/cursos-publicos')
    cursos.value = res.data || []
  } catch {
    toast.error('No pudimos cargar el catálogo. Intenta recargar la página.')
  } finally {
    loading.value = false
  }
}

// ── Normalización de búsqueda ───────────────────────────────────────────────
/** Quita acentos para que "capacitacion" encuentre "Capacitación". */
function normalizar(txt: string) {
  return (txt || '').toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').trim()
}

const gratis = computed(() => cursos.value.filter((c) => !c.precio || c.precio <= 0))
const premium = computed(() => cursos.value.filter((c) => (c.precio ?? 0) > 0))

/** Destacados: los de pago más recientes; si no hay, lo que exista. */
const destacados = computed(() => {
  const base = premium.value.length ? premium.value : cursos.value
  return [...base]
    .sort((a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime())
    .slice(0, 10)
})

const autoformativos = computed(() =>
  cursos.value.filter((c) => c.type === 'document' || c.type === 'text' || c.type === 'link').slice(0, 10)
)

const conteoPorTipo = computed(() => {
  const acc: Record<string, number> = { todos: cursos.value.length }
  for (const c of cursos.value) {
    const k = c.type || 'otro'
    acc[k] = (acc[k] || 0) + 1
  }
  return acc
})

const filtrados = computed(() => {
  const term = normalizar(search.value)
  let list = cursos.value.filter((c) => {
    const heno = normalizar(`${c.title || ''} ${c.description || ''}`)
    const okBusqueda = !term || heno.includes(term)
    const okTipo = tipoFiltro.value === 'todos' || c.type === tipoFiltro.value
    const okPrecio =
      precioFiltro.value === 'todos' ||
      (precioFiltro.value === 'gratis' && (!c.precio || c.precio <= 0)) ||
      (precioFiltro.value === 'pago' && (c.precio ?? 0) > 0)
    return okBusqueda && okTipo && okPrecio
  })

  const porFecha = (a: StoreCourse, b: StoreCourse) =>
    new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime()

  switch (orden.value) {
    case 'az': list = [...list].sort((a, b) => a.title.localeCompare(b.title)); break
    case 'za': list = [...list].sort((a, b) => b.title.localeCompare(a.title)); break
    case 'precio_asc': list = [...list].sort((a, b) => (a.precio || 0) - (b.precio || 0)); break
    case 'precio_desc': list = [...list].sort((a, b) => (b.precio || 0) - (a.precio || 0)); break
    default: list = [...list].sort(porFecha)
  }
  return list
})

const totalPaginas = computed(() => Math.max(1, Math.ceil(filtrados.value.length / POR_PAGINA)))
const paginados = computed(() => filtrados.value.slice((pagina.value - 1) * POR_PAGINA, pagina.value * POR_PAGINA))

const hayFiltrosActivos = computed(
  () => !!search.value || tipoFiltro.value !== 'todos' || precioFiltro.value !== 'todos'
)

watch([search, tipoFiltro, precioFiltro, orden], () => { pagina.value = 1 })

function limpiarFiltros() {
  search.value = ''
  tipoFiltro.value = 'todos'
  precioFiltro.value = 'todos'
}

// ── Acciones ────────────────────────────────────────────────────────────────
const idsEnCarrito = computed(() => cart.items.map((i) => i.curso_id))

function abrirCurso(id: string) {
  router.push(`/curso/${id}`)
}

function agregarAlCarrito(curso: StoreCourse) {
  if (idsEnCarrito.value.includes(curso.id)) {
    cart.openDrawer()
    return
  }
  cart.addItem({
    curso_id: curso.id,
    title: curso.title,
    thumbnail: curso.thumbnail_url,
    precio: curso.precio ?? 0,
    cantidad: 1,
    type: 'b2c',
  })
  toast.success(`"${curso.title}" se agregó al carrito`)
}

function irACatalogo() {
  document.getElementById('catalogo')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function irAPanel() {
  const rol = auth.user?.role
  router.push(rol === 'admin' ? '/admin' : rol === 'instructor' ? '/instructor' : '/usuario')
}

// Atajo ⌘K / Ctrl+K para enfocar el buscador — patrón esperado en un catálogo.
function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    document.getElementById('store-search')?.focus()
  }
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="store">
    <!-- ══════════ HEADER ══════════ -->
    <header :class="['hdr', scrolled && 'hdr--scrolled']">
      <div class="hdr__inner">
        <button class="brand" @click="router.push('/')" aria-label="MH Capacitaciones — ir al inicio">
          <img :src="logoSrc" alt="MH Capacitaciones" class="brand__logo" />
        </button>

        <nav class="hdr__links" aria-label="Secciones">
          <button class="hdr__link" @click="irACatalogo">Catálogo</button>
          <button class="hdr__link" @click="precioFiltro = 'gratis'; irACatalogo()">Gratis</button>
        </nav>

        <div class="hdr__actions">
          <button class="icon-btn" :aria-label="isDark ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'" @click="toggleTheme">
            <svg v-if="isDark" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M6.3 17.7l-1.4 1.4M19.1 4.9l-1.4 1.4" /></svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" /></svg>
          </button>

          <button class="icon-btn icon-btn--cart" aria-label="Abrir carrito" @click="cart.openDrawer">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1" /><circle cx="19" cy="21" r="1" /><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12" /></svg>
            <span v-if="cart.totalItems" class="cart-dot">{{ cart.totalItems }}</span>
          </button>

          <!-- Autenticación: primero en la jerarquía visual del header -->
          <template v-if="auth.isLoggedIn">
            <button class="btn btn--solid" @click="irAPanel">Mi panel</button>
          </template>
          <template v-else>
            <button class="btn btn--ghost hide-sm" @click="router.push('/login')">Iniciar sesión</button>
            <button class="btn btn--solid" @click="router.push('/login?tab=register')">Registrarse</button>
          </template>

          <button
            class="icon-btn burger"
            :aria-expanded="menuAbierto"
            aria-label="Abrir menú"
            @click="menuAbierto = !menuAbierto"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path v-if="!menuAbierto" d="M3 6h18M3 12h18M3 18h18" /><path v-else d="M18 6 6 18M6 6l12 12" /></svg>
          </button>
        </div>
      </div>

      <Transition name="drop">
        <nav v-if="menuAbierto" class="hdr__mobile" aria-label="Menú móvil">
          <button @click="menuAbierto = false; irACatalogo()">Catálogo</button>
          <button @click="menuAbierto = false; precioFiltro = 'gratis'; irACatalogo()">Cursos gratis</button>
          <button v-if="!auth.isLoggedIn" @click="menuAbierto = false; router.push('/login')">Iniciar sesión</button>
        </nav>
      </Transition>
    </header>

    <!-- ══════════ HERO ══════════ -->
    <section class="hero">
      <div class="hero__glow" aria-hidden="true" />
      <div class="hero__inner">
        <p class="hero__sub">
          Cursos en línea impartidos por instructores certificados.
          Compra licencias para toda tu empresa y reparte los accesos por correo en un clic.
        </p>

        <div class="hero__search">
          <svg class="hero__search-icon" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="11" cy="11" r="7" /><path d="m21 21-4.3-4.3" /></svg>
          <input
            id="store-search"
            v-model="search"
            type="search"
            placeholder="Busca por tema, curso o formato…"
            autocomplete="off"
            aria-label="Buscar cursos"
            @keydown.enter="irACatalogo"
          />
          <kbd v-if="!search" class="hero__kbd">⌘K</kbd>
          <button v-else class="hero__clear" aria-label="Limpiar búsqueda" @click="search = ''">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12" /></svg>
          </button>
        </div>

        <div class="hero__ctas">
          <button class="btn btn--solid btn--lg" @click="irACatalogo">Explorar catálogo</button>
          <button v-if="!auth.isLoggedIn" class="btn btn--ghost btn--lg" @click="router.push('/login?tab=register')">
            Crear cuenta gratis
          </button>
        </div>

        <dl class="hero__stats">
          <div><dt>{{ cursos.length || '—' }}</dt><dd>Capacitaciones</dd></div>
          <div><dt>{{ gratis.length || '—' }}</dt><dd>Gratuitas</dd></div>
          <div><dt>{{ premium.length || '—' }}</dt><dd>Empresariales</dd></div>
          <div><dt>DC-3</dt><dd>Constancia STPS</dd></div>
        </dl>
      </div>
    </section>

    <!-- ══════════ CARRUSELES ══════════ -->
    <div class="rows">
      <CourseCarousel
        title="Destacados"
        subtitle="Lo más reciente de nuestro catálogo"
        :courses="destacados"
        :cart-ids="idsEnCarrito"
        :loading="loading"
        @open="abrirCurso"
        @add="agregarAlCarrito"
      />

      <CourseCarousel
        v-if="!loading && gratis.length"
        title="Empieza gratis"
        subtitle="Sin costo, sin tarjeta"
        :courses="gratis"
        :cart-ids="idsEnCarrito"
        @open="abrirCurso"
        @add="agregarAlCarrito"
      />

      <CourseCarousel
        v-if="!loading && autoformativos.length"
        title="A tu ritmo"
        subtitle="Material autoformativo que avanzas cuando quieras"
        :courses="autoformativos"
        :cart-ids="idsEnCarrito"
        @open="abrirCurso"
        @add="agregarAlCarrito"
      />
    </div>

    <!-- ══════════ BANDA B2B ══════════ -->
    <section class="b2b">
      <div class="b2b__inner">
        <div class="b2b__copy">
          <h2>¿Capacitas a un equipo completo?</h2>
          <p>
            Compra licencias corporativas, captura los correos de tus colaboradores y cada
            uno recibe su acceso por correo. Con constancia DC-3 lista para la STPS.
          </p>
          <button class="btn btn--solid btn--lg" @click="precioFiltro = 'pago'; irACatalogo()">
            Ver cursos empresariales
          </button>
        </div>
        <ul class="b2b__list">
          <li><span class="b2b__num">1</span> Elige el curso y el número de lugares</li>
          <li><span class="b2b__num">2</span> Paga en línea con factura automática</li>
          <li><span class="b2b__num">3</span> Reparte los accesos por correo</li>
        </ul>
      </div>
    </section>

    <!-- ══════════ CATÁLOGO ══════════ -->
    <main id="catalogo" class="catalog">
      <header class="catalog__head">
        <h2 class="catalog__title">Catálogo completo</h2>
        <p class="catalog__sub">Filtra por formato o precio para encontrar lo que necesitas.</p>
      </header>

      <div class="filters">
        <div class="chips" role="group" aria-label="Filtrar por formato">
          <button
            v-for="t in TIPOS"
            :key="t.key"
            v-show="t.key === 'todos' || conteoPorTipo[t.key]"
            :class="['chip', tipoFiltro === t.key && 'chip--on']"
            :aria-pressed="tipoFiltro === t.key"
            @click="tipoFiltro = t.key"
          >
            {{ t.label }}
            <span class="chip__n">{{ conteoPorTipo[t.key] || 0 }}</span>
          </button>

          <span class="chips__sep" aria-hidden="true" />

          <button
            :class="['chip', precioFiltro === 'gratis' && 'chip--on chip--green']"
            :aria-pressed="precioFiltro === 'gratis'"
            @click="precioFiltro = precioFiltro === 'gratis' ? 'todos' : 'gratis'"
          >
            Gratis <span class="chip__n">{{ gratis.length }}</span>
          </button>
          <button
            :class="['chip', precioFiltro === 'pago' && 'chip--on']"
            :aria-pressed="precioFiltro === 'pago'"
            @click="precioFiltro = precioFiltro === 'pago' ? 'todos' : 'pago'"
          >
            De pago <span class="chip__n">{{ premium.length }}</span>
          </button>
        </div>

        <label class="sort">
          <span class="sr-only">Ordenar resultados</span>
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M3 6h18M7 12h10M11 18h2" /></svg>
          <select v-model="orden">
            <option value="reciente">Más recientes</option>
            <option value="az">A → Z</option>
            <option value="za">Z → A</option>
            <option value="precio_asc">Precio: menor a mayor</option>
            <option value="precio_desc">Precio: mayor a menor</option>
          </select>
        </label>
      </div>

      <p class="results">
        <strong>{{ filtrados.length }}</strong>
        {{ filtrados.length === 1 ? 'resultado' : 'resultados' }}
        <template v-if="search"> para “{{ search }}”</template>
        <button v-if="hayFiltrosActivos" class="results__clear" @click="limpiarFiltros">Limpiar filtros</button>
      </p>

      <div v-if="loading" class="grid">
        <div v-for="n in 8" :key="n" class="gskel" aria-hidden="true">
          <div class="gskel__media" />
          <div class="gskel__line" />
          <div class="gskel__line gskel__line--short" />
        </div>
      </div>

      <div v-else-if="paginados.length" class="grid">
        <CourseTile
          v-for="c in paginados"
          :key="c.id"
          :course="c"
          :in-cart="idsEnCarrito.includes(c.id)"
          @open="abrirCurso"
          @add="agregarAlCarrito"
        />
      </div>

      <div v-else class="empty">
        <div class="empty__icon">
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="11" cy="11" r="7" /><path d="m21 21-4.3-4.3" /></svg>
        </div>
        <h3>Sin resultados</h3>
        <p>No encontramos capacitaciones con esos criterios.</p>
        <button class="btn btn--solid" @click="limpiarFiltros">Limpiar filtros</button>
      </div>

      <nav v-if="!loading && totalPaginas > 1" class="pager" aria-label="Paginación">
        <button class="pager__arrow" :disabled="pagina === 1" aria-label="Página anterior" @click="pagina--">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M15 18l-6-6 6-6" /></svg>
        </button>
        <template v-for="p in totalPaginas" :key="p">
          <button
            v-if="p === 1 || p === totalPaginas || Math.abs(p - pagina) <= 1"
            :class="['pager__n', pagina === p && 'pager__n--on']"
            :aria-current="pagina === p ? 'page' : undefined"
            @click="pagina = p"
          >{{ p }}</button>
          <span v-else-if="p === pagina - 2 || p === pagina + 2" class="pager__gap">…</span>
        </template>
        <button class="pager__arrow" :disabled="pagina === totalPaginas" aria-label="Página siguiente" @click="pagina++">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M9 18l6-6-6-6" /></svg>
        </button>
      </nav>
    </main>

    <!-- ══════════ FOOTER ══════════ -->
    <footer class="foot">
      <div class="foot__inner">
        <div class="foot__brand">
          <img :src="logoSrc" alt="" class="brand__logo" />
          <div>
            <strong>MH Capacitaciones</strong>
            <p>Capacitación empresarial con constancia DC-3.</p>
          </div>
        </div>
        <p class="foot__legal">© {{ new Date().getFullYear() }} MH Soluciones Empresariales · Todos los derechos reservados</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.store {
  min-height: 100vh;
  background: var(--bg);
  color: var(--text);
  -webkit-font-smoothing: antialiased;
}

.sr-only {
  position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px;
  overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; border: 0;
}

/* ── Botones compartidos ───────────────────────────────── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 9px 17px;
  border-radius: 10px;
  border: 1px solid transparent;
  font-size: 0.87rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.2s, border-color 0.2s, transform 0.2s var(--ease-apple), box-shadow 0.2s;
}
.btn--solid {
  background: var(--brand);
  color: #fff;
  box-shadow: 0 4px 14px rgba(249, 115, 22, 0.28);
}
.btn--solid:hover { background: var(--brand-dark); transform: translateY(-1px); box-shadow: 0 8px 20px rgba(249, 115, 22, 0.34); }
.btn--ghost { background: transparent; color: var(--text); border-color: var(--border); }
.btn--ghost:hover { background: var(--surface-soft); border-color: var(--border-hover, var(--border)); }
.btn--lg { padding: 13px 26px; font-size: 0.97rem; border-radius: 12px; }

/* ── Header ────────────────────────────────────────────── */
.hdr {
  position: sticky;
  top: 0;
  z-index: 60;
  background: color-mix(in srgb, var(--surface) 78%, transparent);
  backdrop-filter: saturate(180%) blur(20px);
  -webkit-backdrop-filter: saturate(180%) blur(20px);
  border-bottom: 1px solid transparent;
  transition: border-color 0.3s, box-shadow 0.3s;
}
.hdr--scrolled { border-bottom-color: var(--border); box-shadow: var(--shadow-xs); }

.hdr__inner {
  max-width: 1240px;
  margin: 0 auto;
  padding: 0 24px;
  height: 62px;
  display: flex;
  align-items: center;
  gap: 20px;
}

.brand {
  display: flex; align-items: center;
  background: none; border: none; padding: 0; cursor: pointer;
  flex-shrink: 0;
  transition: opacity 0.2s;
}
.brand:hover { opacity: 0.75; }
/* Sin texto al lado, el logo carga solo la identidad de la marca: sube de 32 a
   38 px para no perderse frente a los botones del header. */
.brand__logo { width: 38px; height: 38px; object-fit: contain; border-radius: 9px; display: block; }

.hdr__links { display: flex; gap: 4px; margin-right: auto; }
.hdr__link {
  background: none; border: none; cursor: pointer;
  padding: 7px 12px; border-radius: 8px;
  font-size: 0.87rem; font-weight: 600; color: var(--muted);
  transition: color 0.2s, background 0.2s;
}
.hdr__link:hover { color: var(--text); background: var(--surface-soft); }

.hdr__actions { display: flex; align-items: center; gap: 9px; margin-left: auto; }

.icon-btn {
  position: relative;
  width: 36px; height: 36px;
  display: grid; place-items: center;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
}
.icon-btn:hover { background: var(--surface-soft); border-color: var(--brand-border); }

.cart-dot {
  position: absolute;
  top: -5px; right: -5px;
  min-width: 18px; height: 18px;
  padding: 0 5px;
  display: grid; place-items: center;
  background: var(--brand); color: #fff;
  border-radius: 9px;
  font-size: 0.66rem; font-weight: 800;
  border: 2px solid var(--surface);
}

.burger { display: none; }

.hdr__mobile {
  display: none;
  flex-direction: column;
  padding: 8px 24px 14px;
  border-top: 1px solid var(--border);
  background: var(--surface);
}
.hdr__mobile button {
  background: none; border: none; text-align: left;
  padding: 11px 4px; font-size: 0.92rem; font-weight: 600;
  color: var(--text); cursor: pointer;
  border-bottom: 1px solid var(--border-light);
}
.hdr__mobile button:last-child { border-bottom: none; }

.drop-enter-active, .drop-leave-active { transition: opacity 0.2s, transform 0.2s var(--ease-apple); }
.drop-enter-from, .drop-leave-to { opacity: 0; transform: translateY(-8px); }

/* ── Hero ──────────────────────────────────────────────── */
.hero { position: relative; overflow: hidden; padding: 76px 24px 64px; }

/* Halo cálido que da profundidad sin recurrir a una imagen pesada. */
.hero__glow {
  position: absolute;
  top: -180px; left: 50%;
  width: 900px; height: 600px;
  transform: translateX(-50%);
  background: radial-gradient(ellipse at center, rgba(249, 115, 22, 0.18), transparent 68%);
  pointer-events: none;
}

.hero__inner { position: relative; max-width: 860px; margin: 0 auto; text-align: center; }

/* Sin titular, el subtítulo pasa a ser el texto principal del hero: sube de
   tamaño y de contraste para que la sección no se vea decapitada. */
.hero__sub {
  max-width: 640px;
  margin: 0 auto 30px;
  font-size: clamp(1.25rem, 3vw, 1.8rem);
  font-weight: 600;
  line-height: 1.35;
  letter-spacing: -0.02em;
  color: var(--dark);
}

.hero__search {
  position: relative;
  display: flex; align-items: center;
  max-width: 560px; margin: 0 auto 24px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  box-shadow: var(--shadow-sm);
  transition: border-color 0.2s, box-shadow 0.2s;
}
.hero__search:focus-within { border-color: var(--brand); box-shadow: 0 0 0 4px var(--brand-light); }
.hero__search-icon { position: absolute; left: 17px; color: var(--subtle); pointer-events: none; }
.hero__search input {
  flex: 1;
  padding: 15px 52px 15px 46px;
  border: none; background: none; outline: none;
  font-size: 0.97rem; color: var(--text); font-family: inherit;
  border-radius: 14px;
}
.hero__search input::placeholder { color: var(--subtle); }
.hero__search input::-webkit-search-cancel-button { display: none; }

.hero__kbd {
  position: absolute; right: 14px;
  padding: 3px 7px;
  background: var(--surface-soft);
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.7rem; font-weight: 600; color: var(--subtle);
  font-family: inherit;
}
.hero__clear {
  position: absolute; right: 12px;
  width: 26px; height: 26px;
  display: grid; place-items: center;
  border: none; border-radius: 50%;
  background: var(--surface-soft); color: var(--muted);
  cursor: pointer;
}
.hero__clear:hover { background: var(--border); }

.hero__ctas { display: flex; gap: 11px; justify-content: center; flex-wrap: wrap; margin-bottom: 42px; }

.hero__stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-width: 620px;
  margin: 0 auto;
  padding: 0;
}
.hero__stats div { text-align: center; }
.hero__stats dt {
  font-size: 1.55rem; font-weight: 800; color: var(--text);
  letter-spacing: -0.03em; line-height: 1.1;
}
.hero__stats dd { margin: 3px 0 0; font-size: 0.76rem; color: var(--subtle); }

/* ── Carruseles ────────────────────────────────────────── */
.rows { max-width: 1240px; margin: 0 auto; padding: 34px 24px 0; }

/* ── Banda B2B ─────────────────────────────────────────── */
.b2b { padding: 20px 24px 56px; }
.b2b__inner {
  max-width: 1240px; margin: 0 auto;
  display: grid; grid-template-columns: 1.1fr 0.9fr; gap: 40px; align-items: center;
  padding: 44px;
  background: linear-gradient(135deg, var(--brand-light), color-mix(in srgb, var(--brand-light) 40%, var(--surface)));
  border: 1px solid var(--brand-border);
  border-radius: var(--r-xl);
}
.b2b__copy h2 {
  margin: 0 0 12px;
  font-size: clamp(1.4rem, 3vw, 1.9rem); font-weight: 800;
  letter-spacing: -0.03em; color: var(--dark);
}
.b2b__copy p { margin: 0 0 22px; font-size: 0.97rem; line-height: 1.6; color: var(--muted); max-width: 46ch; }

.b2b__list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 14px; }
.b2b__list li {
  display: flex; align-items: center; gap: 13px;
  padding: 14px 17px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r);
  font-size: 0.89rem; font-weight: 600; color: var(--text);
}
.b2b__num {
  flex-shrink: 0;
  width: 26px; height: 26px;
  display: grid; place-items: center;
  background: var(--brand); color: #fff;
  border-radius: 50%;
  font-size: 0.77rem; font-weight: 800;
}

/* ── Catálogo ──────────────────────────────────────────── */
.catalog { max-width: 1240px; margin: 0 auto; padding: 12px 24px 72px; scroll-margin-top: 74px; }

.catalog__head { margin-bottom: 22px; }
.catalog__title {
  margin: 0; font-size: clamp(1.5rem, 3vw, 2rem); font-weight: 800;
  letter-spacing: -0.03em; color: var(--dark);
}
.catalog__sub { margin: 5px 0 0; font-size: 0.92rem; color: var(--muted); }

.filters {
  display: flex; align-items: center; justify-content: space-between;
  gap: 16px; flex-wrap: wrap; margin-bottom: 16px;
  padding-bottom: 16px; border-bottom: 1px solid var(--border);
}

.chips { display: flex; align-items: center; gap: 7px; flex-wrap: wrap; }
.chip {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 13px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--surface);
  color: var(--text);
  font-size: 0.83rem; font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s, color 0.2s;
}
.chip:hover { border-color: var(--brand-border); }
.chip--on { background: var(--brand); border-color: var(--brand); color: #fff; }
.chip--green { background: var(--success); border-color: var(--success); }
.chip__n {
  padding: 1px 6px; border-radius: 999px;
  background: var(--surface-soft); color: var(--muted);
  font-size: 0.71rem; font-weight: 700;
}
.chip--on .chip__n { background: rgba(255, 255, 255, 0.24); color: #fff; }
.chips__sep { width: 1px; height: 20px; background: var(--border); margin: 0 3px; }

.sort {
  display: inline-flex; align-items: center; gap: 7px;
  padding: 7px 12px;
  border: 1px solid var(--border); border-radius: 10px;
  background: var(--surface); color: var(--muted);
}
.sort select {
  border: none; background: none; outline: none;
  font-size: 0.84rem; font-weight: 600; color: var(--text);
  font-family: inherit; cursor: pointer;
}

.results { margin: 0 0 20px; font-size: 0.87rem; color: var(--muted); }
.results strong { color: var(--text); }
.results__clear {
  margin-left: 10px; background: none; border: none;
  color: var(--brand); font-size: 0.84rem; font-weight: 700;
  cursor: pointer; padding: 0; font-family: inherit;
}
.results__clear:hover { text-decoration: underline; }

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(255px, 1fr));
  gap: 22px;
}

.gskel {
  border: 1px solid var(--border); border-radius: var(--r-lg);
  overflow: hidden; padding-bottom: 16px; background: var(--surface);
}
.gskel__media { aspect-ratio: 16 / 10; }
.gskel__line { height: 12px; margin: 12px 16px 0; border-radius: 6px; }
.gskel__line--short { width: 55%; }
.gskel__media, .gskel__line {
  background: linear-gradient(90deg, var(--surface-soft) 25%, var(--border-light) 50%, var(--surface-soft) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s ease-in-out infinite;
}
@keyframes shimmer { from { background-position: 200% 0; } to { background-position: -200% 0; } }

.empty { text-align: center; padding: 68px 24px; }
.empty__icon {
  width: 62px; height: 62px; margin: 0 auto 18px;
  display: grid; place-items: center; border-radius: 50%;
  background: var(--brand-light); color: var(--brand);
}
.empty h3 { margin: 0 0 7px; font-size: 1.15rem; font-weight: 700; color: var(--text); }
.empty p { margin: 0 0 20px; font-size: 0.92rem; color: var(--muted); }

.pager { display: flex; align-items: center; justify-content: center; gap: 7px; margin-top: 42px; }
.pager__arrow, .pager__n {
  min-width: 36px; height: 36px; padding: 0 10px;
  display: grid; place-items: center;
  border: 1px solid var(--border); border-radius: 9px;
  background: var(--surface); color: var(--text);
  font-size: 0.85rem; font-weight: 700; cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
}
.pager__arrow:hover:not(:disabled), .pager__n:hover { border-color: var(--brand-border); background: var(--surface-soft); }
.pager__arrow:disabled { opacity: 0.35; cursor: default; }
.pager__n--on { background: var(--brand); border-color: var(--brand); color: #fff; }
.pager__gap { color: var(--subtle); padding: 0 3px; }

/* ── Footer ────────────────────────────────────────────── */
.foot { border-top: 1px solid var(--border); background: var(--surface); }
.foot__inner {
  max-width: 1240px; margin: 0 auto; padding: 34px 24px;
  display: flex; align-items: center; justify-content: space-between;
  gap: 22px; flex-wrap: wrap;
}
.foot__brand { display: flex; align-items: center; gap: 13px; }
.foot__brand strong { font-size: 0.94rem; font-weight: 800; color: var(--text); }
.foot__brand p { margin: 2px 0 0; font-size: 0.82rem; color: var(--muted); }
.foot__legal { margin: 0; font-size: 0.8rem; color: var(--subtle); }

/* ══════════ RESPONSIVO ══════════ */
@media (max-width: 940px) {
  .b2b__inner { grid-template-columns: 1fr; gap: 28px; padding: 34px 26px; }
}

@media (max-width: 780px) {
  .hdr__links { display: none; }
  .burger { display: grid; }
  .hdr__mobile { display: flex; }
  .hide-sm { display: none; }
  .hero { padding: 52px 20px 44px; }
  .hero__stats { grid-template-columns: repeat(2, 1fr); gap: 20px 12px; }
  .rows, .catalog, .foot__inner, .hdr__inner { padding-left: 20px; padding-right: 20px; }
  .b2b { padding: 12px 20px 44px; }
}

@media (max-width: 560px) {
  .hdr__inner { height: 56px; gap: 10px; }
  .brand__logo { width: 34px; height: 34px; }
  .hero__ctas { flex-direction: column; }
  .hero__ctas .btn { width: 100%; }
  .filters { flex-direction: column; align-items: stretch; }
  /* Los chips se desbordan lateralmente en vez de apilarse y comerse la pantalla. */
  .chips {
    flex-wrap: nowrap;
    overflow-x: auto;
    padding-bottom: 4px;
    scrollbar-width: none;
  }
  .chips::-webkit-scrollbar { display: none; }
  .chips__sep { display: none; }
  .sort { justify-content: space-between; }
  .grid { grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 14px; }
  .foot__inner { flex-direction: column; align-items: flex-start; }
}

@media (prefers-reduced-motion: reduce) {
  .btn, .icon-btn, .chip { transition: none; animation: none; }
  .btn--solid:hover { transform: none; }
  .gskel__media, .gskel__line { animation: none; }
}
</style>
