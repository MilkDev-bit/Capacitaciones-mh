<script setup lang="ts">
import { computed } from 'vue'
import { resumenMarkdown } from '../../utils/markdown'

export interface StoreCourse {
  id: string
  title: string
  description?: string
  type?: string
  precio?: number
  thumbnail_url?: string
  color?: string
  duration?: number
  total_lecciones?: number
  created_at?: string
}

const props = withDefaults(
  defineProps<{
    course: StoreCourse
    /** 'grid' se estira al contenedor; 'carousel' usa ancho fijo para el scroll horizontal. */
    variant?: 'grid' | 'carousel'
    inCart?: boolean
    /** El usuario tiene una suscripción vigente: el curso ya está pagado. */
    incluidoEnPlan?: boolean
  }>(),
  { variant: 'grid', inCart: false, incluidoEnPlan: false }
)

const emit = defineEmits<{
  (e: 'open', id: string): void
  (e: 'add', course: StoreCourse): void
}>()

const TYPE_META: Record<string, { label: string; tint: string }> = {
  video: { label: 'Video', tint: '#6366f1' },
  document: { label: 'Documento', tint: '#0071e3' },
  text: { label: 'Lectura', tint: '#10b981' },
  link: { label: 'Enlace', tint: '#8b5cf6' },
}

const meta = computed(() => TYPE_META[props.course.type ?? ''] ?? { label: 'Curso', tint: '#f97316' })
const esGratis = computed(() => !props.course.precio || props.course.precio <= 0)

/**
 * Los gratuitos se inscriben desde el detalle, no se compran. Con suscripción
 * vigente tampoco: mostrar "Agregar al carrito" sobre algo que el usuario ya
 * paga cada mes es el camino más corto a un cobro duplicado y a un reembolso.
 */
const puedeAgregarDirecto = computed(() => !esGratis.value && !props.incluidoEnPlan)

const imagen = computed(() => {
  const path = props.course.thumbnail_url
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
})

/** Degradado determinista por id: dos cursos sin portada no se ven idénticos. */
const fondoPlaceholder = computed(() => {
  if (props.course.color) return `linear-gradient(140deg, ${props.course.color}, ${meta.value.tint})`
  let hash = 0
  for (const ch of props.course.id) hash = (hash * 31 + ch.charCodeAt(0)) % 360
  return `linear-gradient(140deg, hsl(${hash} 62% 32%), hsl(${(hash + 48) % 360} 58% 20%))`
})

const precioTexto = computed(() =>
  new Intl.NumberFormat('es-MX', {
    style: 'currency',
    currency: 'MXN',
    maximumFractionDigits: 0,
  }).format(props.course.precio ?? 0)
)

const duracionTexto = computed(() => {
  const min = props.course.duration ?? 0
  if (min <= 0) return ''
  if (min < 60) return `${min} min`
  const h = Math.floor(min / 60)
  const m = min % 60
  return m ? `${h} h ${m} min` : `${h} h`
})

const esNuevo = computed(() => {
  if (!props.course.created_at) return false
  const d = new Date(props.course.created_at).getTime()
  if (Number.isNaN(d)) return false
  return Date.now() - d < 1000 * 60 * 60 * 24 * 21
})
</script>

<template>
  <article
    :class="['tile', `tile--${variant}`]"
    tabindex="0"
    role="button"
    :aria-label="`Ver el curso ${course.title}`"
    @click="emit('open', course.id)"
    @keydown.enter="emit('open', course.id)"
    @keydown.space.prevent="emit('open', course.id)"
  >
    <div class="tile__media">
      <img v-if="imagen" :src="imagen" :alt="`Portada de ${course.title}`" class="tile__img" loading="lazy" />
      <div v-else class="tile__ph" :style="{ background: fondoPlaceholder }">
        <span class="tile__ph-letter">{{ course.title.charAt(0).toUpperCase() }}</span>
      </div>

      <div class="tile__badges">
        <span class="badge badge--type" :style="{ '--tint': meta.tint }">{{ meta.label }}</span>
        <span v-if="esNuevo" class="badge badge--new">Nuevo</span>
      </div>

      <span v-if="esGratis" class="badge badge--free">Gratis</span>
      <span v-else-if="incluidoEnPlan" class="badge badge--plan">Incluido en tu plan</span>

      <!-- Acción rápida: aparece en hover/focus en escritorio, siempre visible en táctil -->
      <div v-if="puedeAgregarDirecto" class="tile__quick">
        <button
          :class="['quick-btn', inCart && 'quick-btn--done']"
          :aria-label="inCart ? 'Ya está en el carrito' : `Agregar ${course.title} al carrito`"
          @click.stop="emit('add', course)"
        >
          <svg v-if="inCart" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1" /><circle cx="19" cy="21" r="1" /><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12" /></svg>
          {{ inCart ? 'En el carrito' : 'Agregar' }}
        </button>
      </div>
    </div>

    <div class="tile__body">
      <h3 class="tile__title">{{ course.title }}</h3>
      <p class="tile__desc">{{ resumenMarkdown(course.description) || 'Capacitación profesional certificada.' }}</p>

      <ul class="tile__meta" v-if="duracionTexto || course.total_lecciones">
        <li v-if="duracionTexto">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10" /><path d="M12 6v6l4 2" /></svg>
          {{ duracionTexto }}
        </li>
        <li v-if="course.total_lecciones">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20" /></svg>
          {{ course.total_lecciones }} lecciones
        </li>
      </ul>

      <div class="tile__foot">
        <span v-if="incluidoEnPlan && !esGratis" class="tile__price tile__price--plan">
          Incluido en tu plan
        </span>
        <span v-else :class="['tile__price', esGratis && 'tile__price--free']">
          {{ esGratis ? 'Gratis' : precioTexto }}
        </span>
        <span class="tile__link">
          {{ incluidoEnPlan && !esGratis ? 'Entrar' : 'Ver detalle' }}
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M12 5l7 7-7 7" /></svg>
        </span>
      </div>
    </div>
  </article>
</template>

<style scoped>
.tile {
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  overflow: hidden;
  cursor: pointer;
  text-align: left;
  transition:
    transform 0.35s var(--ease-apple),
    box-shadow 0.35s var(--ease-apple),
    border-color 0.35s var(--ease-apple);
}
.tile--grid { width: 100%; }
/* Ancho fijo: es lo que permite el scroll-snap horizontal del carrusel. */
.tile--carousel { width: 288px; flex: 0 0 288px; }

.tile:hover,
.tile:focus-visible {
  transform: translateY(-6px);
  box-shadow: var(--shadow-md);
  border-color: var(--brand-border);
  outline: none;
}
.tile:focus-visible { box-shadow: 0 0 0 3px var(--brand-light), var(--shadow-md); }

.badge--plan {
  position: absolute;
  bottom: 10px;
  left: 10px;
  background: rgba(16, 185, 129, 0.95);
  color: #fff;
  font-weight: 700;
}
.tile__price--plan { color: #10b981; font-size: 0.95rem; }

/* ── Media ─────────────────────────────────────────────── */
.tile__media {
  position: relative;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  background: var(--surface-soft);
}
.tile__img,
.tile__ph {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s var(--ease-apple);
}
.tile:hover .tile__img,
.tile:hover .tile__ph { transform: scale(1.05); }

.tile__ph { display: grid; place-items: center; }
.tile__ph-letter {
  font-size: 3.2rem;
  font-weight: 800;
  color: rgba(255, 255, 255, 0.9);
  letter-spacing: -0.03em;
}

.tile__badges {
  position: absolute;
  top: 10px;
  left: 10px;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  max-width: calc(100% - 20px);
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 9px;
  border-radius: 7px;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.01em;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  white-space: nowrap;
}
.badge--type { background: color-mix(in srgb, var(--tint) 88%, transparent); color: #fff; }
.badge--new { background: rgba(255, 255, 255, 0.92); color: var(--dark); }
.badge--free {
  position: absolute;
  top: 10px;
  right: 10px;
  background: var(--success);
  color: #fff;
}

/* ── Acción rápida ─────────────────────────────────────── */
.tile__quick {
  position: absolute;
  right: 10px;
  bottom: 10px;
  opacity: 0;
  transform: translateY(6px);
  transition: opacity 0.25s var(--ease-apple), transform 0.25s var(--ease-apple);
}
.tile:hover .tile__quick,
.tile:focus-within .tile__quick { opacity: 1; transform: translateY(0); }

.quick-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 13px;
  border: none;
  border-radius: 9px;
  background: var(--brand);
  color: #fff;
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(249, 115, 22, 0.35);
  transition: background 0.2s, transform 0.2s;
}
.quick-btn:hover { background: var(--brand-dark); transform: scale(1.04); }
.quick-btn--done { background: var(--success); box-shadow: 0 4px 14px rgba(52, 199, 89, 0.32); }
.quick-btn--done:hover { background: var(--success); }

/* En táctil no existe el hover: el botón vive siempre visible. */
@media (hover: none) {
  .tile__quick { opacity: 1; transform: none; }
}

/* ── Cuerpo ────────────────────────────────────────────── */
.tile__body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px 16px 14px;
  flex: 1;
}

.tile__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.35;
  color: var(--text);
  letter-spacing: -0.01em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tile__desc {
  margin: 0;
  font-size: 0.83rem;
  line-height: 1.5;
  color: var(--muted);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tile__meta {
  list-style: none;
  margin: 2px 0 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 4px 12px;
}
.tile__meta li {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.74rem;
  color: var(--subtle);
}

.tile__foot {
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid var(--border-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.tile__price {
  font-size: 1.05rem;
  font-weight: 800;
  color: var(--text);
  letter-spacing: -0.02em;
}
.tile__price--free { color: var(--success); }

.tile__link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--brand);
  transition: gap 0.2s var(--ease-apple);
}
.tile:hover .tile__link { gap: 7px; }

@media (max-width: 640px) {
  .tile--carousel { width: 232px; flex-basis: 232px; }
  .tile__body { padding: 13px 13px 12px; }
  .tile__title { font-size: 0.93rem; }
  .tile__desc { font-size: 0.79rem; }
}

@media (prefers-reduced-motion: reduce) {
  .tile,
  .tile__img,
  .tile__ph,
  .tile__quick { transition: none; animation: none; }
  .tile:hover { transform: none; }
  .tile:hover .tile__img,
  .tile:hover .tile__ph { transform: none; }
}
</style>
