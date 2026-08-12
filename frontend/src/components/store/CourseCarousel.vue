<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import CourseTile, { type StoreCourse } from './CourseTile.vue'

const props = withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    courses: StoreCourse[]
    /** IDs ya presentes en el carrito, para marcar las tarjetas. */
    cartIds?: string[]
    loading?: boolean
  }>(),
  { cartIds: () => [], loading: false }
)

const emit = defineEmits<{
  (e: 'open', id: string): void
  (e: 'add', course: StoreCourse): void
}>()

const track = ref<HTMLElement | null>(null)
const puedeIzq = ref(false)
const puedeDer = ref(false)

/**
 * El scroll nativo es la fuente de verdad (así los gestos táctiles y el
 * trackpad funcionan gratis); las flechas solo lo empujan.
 */
function actualizarFlechas() {
  const el = track.value
  if (!el) return
  // El margen de 4px absorbe los redondeos subpíxel del scroll en algunos navegadores.
  puedeIzq.value = el.scrollLeft > 4
  puedeDer.value = el.scrollLeft + el.clientWidth < el.scrollWidth - 4
}

function desplazar(direccion: -1 | 1) {
  const el = track.value
  if (!el) return
  // Se avanza casi una pantalla, dejando una tarjeta visible como ancla.
  el.scrollBy({ left: direccion * (el.clientWidth * 0.85), behavior: 'smooth' })
}

let observer: ResizeObserver | undefined

onMounted(async () => {
  await nextTick()
  actualizarFlechas()
  if (typeof ResizeObserver !== 'undefined' && track.value) {
    observer = new ResizeObserver(actualizarFlechas)
    observer.observe(track.value)
  }
})

onUnmounted(() => observer?.disconnect())

watch(() => props.courses, () => nextTick(actualizarFlechas), { deep: false })
</script>

<template>
  <section v-if="loading || courses.length" class="carousel">
    <header class="carousel__head">
      <div class="carousel__titles">
        <h2 class="carousel__title">{{ title }}</h2>
        <p v-if="subtitle" class="carousel__sub">{{ subtitle }}</p>
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
      <div ref="track" class="carousel__track" @scroll.passive="actualizarFlechas">
        <template v-if="loading">
          <div v-for="n in 4" :key="`s-${n}`" class="skeleton" aria-hidden="true">
            <div class="skeleton__media" />
            <div class="skeleton__line" />
            <div class="skeleton__line skeleton__line--short" />
          </div>
        </template>

        <CourseTile
          v-else
          v-for="c in courses"
          :key="c.id"
          :course="c"
          variant="carousel"
          :in-cart="cartIds.includes(c.id)"
          @open="emit('open', $event)"
          @add="emit('add', $event)"
        />
      </div>

      <!-- Difuminados que insinúan que hay más contenido hacia los lados -->
      <div :class="['fade', 'fade--left', puedeIzq && 'fade--on']" aria-hidden="true" />
      <div :class="['fade', 'fade--right', puedeDer && 'fade--on']" aria-hidden="true" />
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
  /* Las flechas sobran donde el gesto de swipe es lo natural. */
  .carousel__nav { display: none; }
  .carousel__track { gap: 13px; }
  .skeleton { width: 232px; flex-basis: 232px; }
}

@media (prefers-reduced-motion: reduce) {
  .carousel__track { scroll-behavior: auto; }
  .skeleton__media,
  .skeleton__line { animation: none; }
}
</style>
