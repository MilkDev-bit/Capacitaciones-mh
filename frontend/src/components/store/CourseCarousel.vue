<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import CourseTile, { type StoreCourse } from './CourseTile.vue'

const props = withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    courses: StoreCourse[]
    /** IDs ya presentes en el carrito, para marcar las tarjetas. */
    cartIds?: string[]
    loading?: boolean
    /** Suscripción vigente: todas las tarjetas se marcan como incluidas. */
    incluidoEnPlan?: boolean
  }>(),
  { cartIds: () => [], loading: false, incluidoEnPlan: false }
)

const emit = defineEmits<{
  (e: 'open', id: string): void
  (e: 'add', course: StoreCourse): void
}>()

const track = ref<HTMLElement | null>(null)
const puedeIzq = ref(false)
const puedeDer = ref(false)
const indiceActivo = ref(0)

/** Con menos de 3 tarjetas los dots y la profundidad sobran. */
const hayControles = computed(() => !props.loading && props.courses.length > 2)

/**
 * La profundidad 3D es decoración pura: se apaga con reduced-motion y en
 * pantallas táctiles pequeñas, donde el coste por frame no compensa y el
 * gesto de swipe ya comunica que hay más contenido.
 */
const efecto3D =
  typeof window !== 'undefined' &&
  !window.matchMedia?.('(prefers-reduced-motion: reduce)').matches &&
  !window.matchMedia?.('(max-width: 768px)').matches

// ── Medición ────────────────────────────────────────────────────────────────

/**
 * El scroll nativo es la fuente de verdad (así los gestos táctiles, el
 * trackpad y el teclado funcionan gratis); flechas y dots solo lo empujan.
 *
 * Todo se mide en un único rAF por frame de scroll: leer getBoundingClientRect
 * de cada tarjeta en el propio handler provocaría un reflow sincrónico por
 * evento y el carrusel iría a tirones justo al arrastrar.
 */
let pendiente = false

function alHacerScroll() {
  if (pendiente) return
  pendiente = true
  requestAnimationFrame(medir)
}

function medir() {
  pendiente = false
  const el = track.value
  if (!el) return

  // El margen de 4px absorbe los redondeos subpíxel del scroll.
  puedeIzq.value = el.scrollLeft > 4
  puedeDer.value = el.scrollLeft + el.clientWidth < el.scrollWidth - 4

  const rect = el.getBoundingClientRect()
  const centro = rect.left + rect.width / 2
  const mitad = rect.width / 2 || 1

  let mejorIdx = 0
  let mejorDist = Infinity

  const hijos = Array.from(el.children) as HTMLElement[]
  hijos.forEach((slide, i) => {
    const r = slide.getBoundingClientRect()
    const cSlide = r.left + r.width / 2

    // Distancia al borde izquierdo: marca cuál es la tarjeta "actual" para los dots.
    const distBorde = Math.abs(r.left - rect.left)
    if (distBorde < mejorDist) {
      mejorDist = distBorde
      mejorIdx = i
    }

    if (!efecto3D) return

    // −1 (izquierda) … 0 (centro) … 1 (derecha)
    const d = Math.max(-1, Math.min(1, (cSlide - centro) / mitad))
    const abs = Math.abs(d)
    // 10° y 4% de escala: suficiente para leerse como profundidad, lo bastante
    // poco para que la portada y el precio sigan siendo legibles de un vistazo.
    slide.style.setProperty('--rot', `${-d * 10}deg`)
    slide.style.setProperty('--esc', `${1 - abs * 0.04}`)
    slide.style.setProperty('--op', `${1 - abs * 0.2}`)
  })

  indiceActivo.value = mejorIdx
}

// ── Navegación ──────────────────────────────────────────────────────────────

function desplazar(direccion: -1 | 1) {
  const el = track.value
  if (!el) return
  // Se avanza casi una pantalla, dejando una tarjeta visible como ancla.
  el.scrollBy({ left: direccion * (el.clientWidth * 0.85), behavior: 'smooth' })
}

function irA(i: number) {
  const el = track.value
  const slide = el?.children[i] as HTMLElement | undefined
  if (!el || !slide) return
  el.scrollTo({ left: slide.offsetLeft - el.offsetLeft, behavior: 'smooth' })
}

/** Flechas del teclado sobre la pista: navegación esperada en un carrusel. */
function alPulsarTecla(e: KeyboardEvent) {
  if (e.key === 'ArrowRight') {
    e.preventDefault()
    desplazar(1)
  } else if (e.key === 'ArrowLeft') {
    e.preventDefault()
    desplazar(-1)
  }
}

let observer: ResizeObserver | undefined

onMounted(async () => {
  await nextTick()
  medir()
  if (typeof ResizeObserver !== 'undefined' && track.value) {
    observer = new ResizeObserver(alHacerScroll)
    observer.observe(track.value)
  }
})

onUnmounted(() => observer?.disconnect())

watch(() => props.courses, () => nextTick(medir), { deep: false })
</script>

<template>
  <section v-if="loading || courses.length" class="carousel">
    <header class="carousel__head">
      <div class="carousel__titles">
        <h2 class="carousel__title" v-reveal>{{ title }}</h2>
        <p v-if="subtitle" class="carousel__sub" v-reveal="1">{{ subtitle }}</p>
      </div>

      <div class="carousel__nav" v-if="!loading && courses.length > 1">
        <button
          class="nav-arrow"
          :disabled="!puedeIzq"
          aria-label="Desplazar a la izquierda"
          @click="desplazar(-1)"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6" /></svg>
        </button>
        <button
          class="nav-arrow"
          :disabled="!puedeDer"
          aria-label="Desplazar a la derecha"
          @click="desplazar(1)"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18l6-6-6-6" /></svg>
        </button>
      </div>
    </header>

    <div class="carousel__viewport">
      <div
        ref="track"
        :class="['carousel__track', efecto3D && 'carousel__track--3d']"
        role="group"
        :aria-label="`Carrusel: ${title}`"
        tabindex="0"
        @scroll.passive="alHacerScroll"
        @keydown="alPulsarTecla"
      >
        <template v-if="loading">
          <div v-for="n in 4" :key="`s-${n}`" class="skeleton" aria-hidden="true">
            <div class="skeleton__media" />
            <div class="skeleton__line" />
            <div class="skeleton__line skeleton__line--short" />
          </div>
        </template>

        <!--
          Cada tarjeta va envuelta: el wrapper carga la transformación 3D y la
          tarjeta conserva íntegro su propio hover, que también usa transform.
          Sin la envoltura, uno pisaría al otro.
        -->
        <div v-else v-for="c in courses" :key="c.id" class="slide">
          <CourseTile
            :course="c"
            variant="carousel"
            :in-cart="cartIds.includes(c.id)"
            :incluido-en-plan="incluidoEnPlan"
            @open="emit('open', $event)"
            @add="emit('add', $event)"
          />
        </div>
      </div>

      <!-- Difuminados que insinúan que hay más contenido hacia los lados -->
      <div :class="['fade', 'fade--left', puedeIzq && 'fade--on']" aria-hidden="true" />
      <div :class="['fade', 'fade--right', puedeDer && 'fade--on']" aria-hidden="true" />
    </div>

    <!-- ── Barra de control: flechas + dots ─────────────────────── -->
    <div v-if="hayControles" class="dock">
      <button
        class="dock__arrow"
        :disabled="!puedeIzq"
        aria-label="Anterior"
        @click="desplazar(-1)"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6" /></svg>
      </button>

      <div class="dock__dots">
        <button
          v-for="(c, i) in courses"
          :key="`d-${c.id}`"
          :class="['dot', indiceActivo === i && 'dot--on']"
          :aria-label="`Ir a ${c.title}`"
          :aria-current="indiceActivo === i ? 'true' : undefined"
          @click="irA(i)"
        />
      </div>

      <button
        class="dock__arrow"
        :disabled="!puedeDer"
        aria-label="Siguiente"
        @click="desplazar(1)"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18l6-6-6-6" /></svg>
      </button>
    </div>
  </section>
</template>

<style scoped>
.carousel { margin-bottom: 52px; }

.carousel__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.carousel__title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.025em;
  color: var(--text);
}
.carousel__sub {
  margin: 4px 0 0;
  font-size: 0.9rem;
  color: var(--muted);
}

.carousel__nav { display: flex; gap: 8px; flex-shrink: 0; }

.nav-arrow {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: 50%;
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s, opacity 0.2s, transform 0.2s;
}
.nav-arrow:hover:not(:disabled) {
  background: var(--brand);
  border-color: var(--brand);
  color: #fff;
  transform: scale(1.06);
}
.nav-arrow:disabled { opacity: 0.3; cursor: default; }

.carousel__viewport { position: relative; }

.carousel__track {
  display: flex;
  gap: 18px;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
  /* El padding vertical deja aire para el translateY del hover de las tarjetas. */
  padding: 6px 2px 14px;
  scrollbar-width: none;
  -ms-overflow-style: none;
}
.carousel__track::-webkit-scrollbar { display: none; }
.carousel__track > * { scroll-snap-align: start; }

.carousel__track:focus-visible {
  outline: 2px solid var(--brand);
  outline-offset: 4px;
  border-radius: var(--r-sm);
}

/* ── Profundidad 3D ────────────────────────────────────── */
/* La perspectiva vive en la pista, no en cada slide: así todas las tarjetas
   comparten un mismo punto de fuga y la fila se lee como un plano único. */
.carousel__track--3d {
  perspective: 1400px;
  perspective-origin: 50% 50%;
}
.carousel__track--3d .slide {
  transform: rotateY(var(--rot, 0deg)) scale(var(--esc, 1));
  opacity: var(--op, 1);
  transform-origin: 50% 50%;
  transition: opacity 0.3s linear;
  will-change: transform;
}
.slide { flex: 0 0 auto; }

.fade {
  position: absolute;
  top: 0;
  bottom: 14px;
  width: 56px;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.25s var(--ease-apple);
}
.fade--on { opacity: 1; }
.fade--left { left: 0; background: linear-gradient(90deg, var(--bg), transparent); }
.fade--right { right: 0; background: linear-gradient(270deg, var(--bg), transparent); }

/* ── Barra de control ──────────────────────────────────── */
.dock {
  width: fit-content;
  max-width: 100%;
  margin: 4px auto 0;
  padding: 4px 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--surface-soft) 70%, transparent);
  border: 1px solid var(--border-light);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
}

.dock__arrow {
  flex: 0 0 auto;
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  transition: color 0.2s, background 0.2s, opacity 0.2s;
}
.dock__arrow:hover:not(:disabled) { color: var(--text); background: var(--surface); }
.dock__arrow:disabled { opacity: 0.25; cursor: default; }

.dock__dots {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  scrollbar-width: none;
  padding: 4px 2px;
}
.dock__dots::-webkit-scrollbar { display: none; }

.dot {
  flex: 0 0 auto;
  width: 7px;
  height: 7px;
  padding: 0;
  border: 0;
  border-radius: 999px;
  background: color-mix(in srgb, var(--muted) 40%, transparent);
  cursor: pointer;
  /* Solo se anima el ancho: el dot activo se estira en vez de crecer, lo que
     evita que la fila entera se desplace al cambiar de tarjeta. */
  transition: width 0.3s var(--ease-apple), background 0.3s var(--ease-apple);
}
.dot--on { width: 24px; background: var(--brand); }
.dot:hover { background: var(--brand); }
/* Área táctil de 24px sin agrandar el punto visible. */
.dot::before { content: ''; position: absolute; inset: -9px; }
.dot { position: relative; }
.dot:focus-visible { outline: 2px solid var(--brand); outline-offset: 3px; }

/* ── Esqueleto de carga ────────────────────────────────── */
.skeleton {
  width: 288px;
  flex: 0 0 288px;
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  overflow: hidden;
  padding-bottom: 16px;
  background: var(--surface);
}
.skeleton__media { aspect-ratio: 16 / 10; }
.skeleton__line { height: 12px; margin: 12px 16px 0; border-radius: 6px; }
.skeleton__line--short { width: 55%; }
.skeleton__media,
.skeleton__line {
  background: linear-gradient(90deg, var(--surface-soft) 25%, var(--border-light) 50%, var(--surface-soft) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s ease-in-out infinite;
}
@keyframes shimmer {
  from { background-position: 200% 0; }
  to { background-position: -200% 0; }
}

@media (max-width: 768px) {
  .carousel { margin-bottom: 40px; }
  .carousel__title { font-size: 1.25rem; }
  .carousel__sub { font-size: 0.84rem; }
  /* Las flechas de la cabecera sobran donde el gesto de swipe es lo natural;
     la barra de dots se queda porque además indica la posición. */
  .carousel__nav { display: none; }
  .carousel__track { gap: 13px; }
  .skeleton { width: 232px; flex-basis: 232px; }
  .dock__arrow { display: none; }
}

@media (prefers-reduced-motion: reduce) {
  .carousel__track { scroll-behavior: auto; }
  .carousel__track--3d .slide { transform: none; opacity: 1; }
  .dot { transition: none; }
  .skeleton__media,
  .skeleton__line { animation: none; }
}
</style>
