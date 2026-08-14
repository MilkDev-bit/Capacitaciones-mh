<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import logoSrc from '../../assets/logo-capacitaciones.png'
import {
  PARADAS,
  rotacionCubo,
  indiceParada,
  paletaEn,
  posicionAstro,
  opacidadEstrellas,
  css,
} from '../../composables/escena'

/**
 * Cómo funciona — escena guiada por scroll.
 *
 * Un cubo de seis caras, una por paso del flujo de compra, girando sobre un
 * paisaje que va de la tarde a la noche. Todo se deriva de `scrollY`: la página
 * NUNCA llama a scrollTo ni intercepta la rueda, así que trackpad, teclado,
 * barra de desplazamiento y lectores de pantalla siguen funcionando igual que
 * en cualquier otra página. La escena es una capa decorativa encima; el
 * contenido de las tarjetas se lee perfectamente sin ella.
 */

const router = useRouter()

interface Paso {
  clave: string
  num: string
  nombre: string
  titulo: string[]
  cuerpo: string
  icono: string
}

const PASOS: Paso[] = [
  {
    clave: 'ruta',
    num: '01',
    nombre: 'TU RUTA',
    titulo: ['ELIGE', 'CÓMO', 'PAGAR'],
    cuerpo:
      'Suscríbete y abre todo el catálogo mientras el plan siga activo, o compra solo el curso que necesitas y consérvalo para siempre. Las dos vías dan la misma constancia.',
    icono: 'M3 12h7l3-8 3 16 3-8h2',
  },
  {
    clave: 'carrito',
    num: '02',
    nombre: 'CARRITO',
    titulo: ['SIN', 'LETRA', 'CHICA'],
    cuerpo:
      'El total que ves es el que pagas.',
    icono: 'M2 3h3l2.7 12.4a2 2 0 0 0 2 1.6h9.8a2 2 0 0 0 1.9-1.6L23 7H5.1',
  },
  {
    clave: 'pago',
    num: '03',
    nombre: 'PAGO',
    titulo: ['PAGA', 'SEGURO'],
    cuerpo:
      'Ningún dato de tu tarjeta pasa por nuestros servidores. Al terminar vuelves a una confirmación con tus accesos ya listos.',
    icono: 'M3 11h18v10H3zM7 11V7a5 5 0 0 1 10 0v4',
  },
  {
    clave: 'accesos',
    num: '04',
    nombre: 'ACCESOS',
    titulo: ['REPARTE', 'AL', 'EQUIPO'],
    cuerpo:
      'Compra licencias corporativas, cada quien recibe su acceso. Si alguien deja la empresa, reasignas su lugar.',
    icono: 'M17 20h5v-2a3 3 0 0 0-5.4-1.9M17 20H7m10 0v-2M7 20H2v-2a3 3 0 0 1 5.4-1.9M15 7a3 3 0 1 1-6 0 3 3 0 0 1 6 0',
  },
  {
    clave: 'ritmo',
    num: '05',
    nombre: 'A TU RITMO',
    titulo: ['APRENDE', 'CUANDO', 'PUEDAS'],
    cuerpo:
      'Video, documentos y lecturas. El progreso se guarda por lección, así que puedes cerrar y retomar donde lo dejaste, desde el teléfono o desde la computadora.',
    icono: 'M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20',
  },
  {
    clave: 'dc3',
    num: '06',
    nombre: 'DC-3',
    titulo: ['TU', 'CONSTANCIA'],
    cuerpo:
      'Al completar el curso se emite tu constancia DC-3, con el formato que pide la STPS. En compras corporativas avisamos al representante para que la descargue.',
    icono: 'M12 15a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM8.2 13.8 7 22l5-3 5 3-1.2-8.2',
  },
]

const N = PASOS.length

// ── Preferencia de movimiento ───────────────────────────────────────────────

/**
 * Con reduced-motion la escena se sustituye por una rejilla estática de los
 * seis pasos. No es "lo mismo pero quieto": una escena 3D congelada muestra
 * seis caras superpuestas e ilegibles. Se cambia de formato, no de contenido.
 */
const reducido =
  typeof window !== 'undefined' &&
  !!window.matchMedia?.('(prefers-reduced-motion: reduce)').matches

// ── Estado derivado del scroll ──────────────────────────────────────────────

/** Progreso suavizado 0–1, el único valor que alimenta toda la escena. */
const s = ref(0)
const indice = computed(() => indiceParada(s.value, N))
const pct = computed(() => Math.round(s.value * 100))
const pasoActual = computed(() => PASOS[indice.value] ?? PASOS[0]!)

const rot = computed(() => rotacionCubo(s.value, PARADAS))
const estiloCubo = computed(() => ({
  transform: `rotateX(${rot.value.rx.toFixed(2)}deg) rotateY(${rot.value.ry.toFixed(2)}deg)`,
}))

const paleta = computed(() => paletaEn(s.value))
const astro = computed(() => posicionAstro(s.value))
const estrellas = computed(() => opacidadEstrellas(s.value))

/** Las colinas cercanas barren más que las lejanas: la cámara avanza. */
function estiloColina(i: number) {
  if (reducido) return undefined
  return { transform: `translate3d(0, ${(-s.value * (60 + i * 70)).toFixed(1)}px, 0)` }
}
const estiloNube1 = computed(() =>
  reducido ? undefined : { transform: `translate3d(${(s.value * 460).toFixed(0)}px, 0, 0)` }
)
const estiloNube2 = computed(() =>
  reducido ? undefined : { transform: `translate3d(${(-s.value * 560).toFixed(0)}px, 0, 0)` }
)

// ── Bucle ───────────────────────────────────────────────────────────────────

let raf = 0
let objetivo = 0
let ultimo = 0

function leerScroll() {
  const alto = Math.max(1, document.documentElement.scrollHeight - window.innerHeight)
  objetivo = Math.min(1, Math.max(0, window.scrollY / alto))
  // Sin bucle de animación el valor salta directo: alimenta el HUD y los dots,
  // que siguen siendo útiles aunque la escena 3D esté oculta.
  if (reducido) s.value = objetivo
}

/**
 * El suavizado es exponencial y ponderado por delta de tiempo: así el cubo gira
 * igual a 60 y a 144 Hz. Con un factor fijo por frame, un monitor rápido
 * adelantaría la animación y uno lento la arrastraría.
 */
function frame(ahora: number) {
  raf = requestAnimationFrame(frame)
  if (document.hidden) {
    ultimo = ahora
    return
  }
  const dt = Math.min((ahora - ultimo) / 1000, 0.05)
  ultimo = ahora
  s.value += (objetivo - s.value) * (1 - Math.exp(-dt * 9))
  // Se ancla al llegar: si no, el valor se queda oscilando en el sexto decimal
  // y Vue vuelve a pintar la escena entera en cada frame para siempre.
  if (Math.abs(objetivo - s.value) < 0.0004) s.value = objetivo
}

function irAPaso(i: number) {
  document.getElementById(`paso-${i}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

onMounted(() => {
  leerScroll()
  s.value = objetivo
  window.addEventListener('scroll', leerScroll, { passive: true })
  window.addEventListener('resize', leerScroll, { passive: true })
  if (!reducido) {
    ultimo = performance.now()
    raf = requestAnimationFrame(frame)
  }
})

onUnmounted(() => {
  cancelAnimationFrame(raf)
  window.removeEventListener('scroll', leerScroll)
  window.removeEventListener('resize', leerScroll)
})
</script>

<template>
  <div :class="['cf', reducido && 'cf--estatico']">
    <!-- ══════════ FONDO: PAISAJE ══════════ -->
    <div class="cf__paisaje" aria-hidden="true">
      <svg viewBox="0 0 1440 900" preserveAspectRatio="xMidYMax slice">
        <defs>
          <linearGradient id="cf-cielo" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="css(paleta.cieloArriba)" />
            <stop offset="72%" :stop-color="css(paleta.cieloAbajo)" />
          </linearGradient>
          <radialGradient id="cf-halo">
            <stop offset="0%" :stop-color="css(paleta.astro)" stop-opacity="0.5" />
            <stop offset="100%" :stop-color="css(paleta.astro)" stop-opacity="0" />
          </radialGradient>
        </defs>

        <rect width="1440" height="900" fill="url(#cf-cielo)" />

        <!-- Estrellas: aparecen solo cuando el cielo ya oscureció -->
        <g class="cf__estrellas" :style="{ opacity: estrellas }" fill="#fff">
          <circle cx="140" cy="120" r="1.6" />
          <circle cx="330" cy="72" r="1.1" />
          <circle cx="520" cy="160" r="1.8" />
          <circle cx="690" cy="60" r="1.2" />
          <circle cx="860" cy="140" r="1.5" />
          <circle cx="1030" cy="86" r="1.9" />
          <circle cx="1210" cy="150" r="1.2" />
          <circle cx="1360" cy="70" r="1.6" />
          <circle cx="240" cy="240" r="1.3" />
          <circle cx="600" cy="266" r="1.5" />
          <circle cx="960" cy="230" r="1.1" />
          <circle cx="1300" cy="270" r="1.7" />
          <circle cx="430" cy="330" r="1.2" />
          <circle cx="1120" cy="345" r="1.4" />
        </g>

        <!-- Astro: mismo elemento para el sol y la luna, solo cambia de color -->
        <circle :cx="astro.x * 14.4" :cy="astro.y * 9" r="150" fill="url(#cf-halo)" />
        <circle :cx="astro.x * 14.4" :cy="astro.y * 9" r="46" :fill="css(paleta.astro)" />

        <g class="cf__nubes" fill="#fff" opacity="0.34">
          <path :style="estiloNube1"
            d="M180 250c14-16 38-12 56-8 16-8 36-12 52 0 12-16 40-18 54-2 10-8 26-6 32 4H180z" />
          <path :style="estiloNube2"
            d="M980 178c12-14 32-10 48-7 14-7 31-10 45 0 10-14 34-15 46-2 9-7 22-5 27 3H980z" />
        </g>

        <!-- Cinco planos de colina, del más lejano al más cercano -->
        <path :style="estiloColina(0)" :fill="css(paleta.colinas[0]!)"
          d="M0,560 C240,500 420,600 720,540 C1020,480 1220,580 1440,520 L1440,960 L0,960 Z" />
        <path :style="estiloColina(1)" :fill="css(paleta.colinas[1]!)"
          d="M0,640 C260,580 460,690 760,630 C1060,570 1240,670 1440,610 L1440,960 L0,960 Z" />
        <path :style="estiloColina(2)" :fill="css(paleta.colinas[2]!)"
          d="M0,712 C300,660 520,760 820,706 C1120,652 1290,730 1440,690 L1440,960 L0,960 Z" />
        <path :style="estiloColina(3)" :fill="css(paleta.colinas[3]!)"
          d="M0,782 C280,742 540,822 840,780 C1140,738 1300,796 1440,764 L1440,960 L0,960 Z" />
        <path :style="estiloColina(4)" :fill="css(paleta.colinas[4]!)"
          d="M0,846 C320,818 560,872 860,842 C1160,812 1320,858 1440,836 L1440,960 L0,960 Z" />
      </svg>
    </div>

    <!-- ══════════ CUBO ══════════ -->
    <div v-if="!reducido" class="cf__escena" aria-hidden="true">
      <div class="cubo" :style="estiloCubo">
        <div v-for="(p, i) in PASOS" :key="p.clave" class="cara" :data-cara="i">
          <span class="cara__num">{{ p.num }}</span>
          <svg class="cara__icono" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4"
            stroke-linecap="round" stroke-linejoin="round">
            <path :d="p.icono" />
          </svg>
          <span class="cara__nombre">{{ p.nombre }}</span>
        </div>
      </div>
    </div>

    <!-- ══════════ HUD ══════════ -->
    <!-- aria-hidden: duplica la posición que el lector de pantalla ya conoce
         por el propio documento; anunciarlo en cada frame sería solo ruido. -->
    <div class="hud" aria-hidden="true">
      <div class="hud__pct">{{ String(pct).padStart(3, '0') }}%</div>
      <div class="hud__barra">
        <div class="hud__fill" :style="{ width: pct + '%' }" />
      </div>
      <div class="hud__etiqueta">{{ pasoActual.nombre }}</div>
    </div>

    <!-- Navegación lateral por pasos -->
    <nav class="strip" aria-label="Pasos">
      <button v-for="(p, i) in PASOS" :key="p.clave" :class="['strip__dot', indice === i && 'is-on']"
        :aria-label="`Paso ${p.num}: ${p.nombre}`" :aria-current="indice === i ? 'step' : undefined"
        @click="irAPaso(i)" />
    </nav>

    <div class="caption" aria-hidden="true">
      <div class="caption__num">{{ pasoActual.num }}</div>
      <div class="caption__nombre">{{ pasoActual.nombre }}</div>
    </div>

    <!-- ══════════ CABECERA ══════════ -->
    <header class="cf__hdr">
      <button class="cf__brand" aria-label="Ir al inicio" @click="router.push('/')">
        <img :src="logoSrc" alt="MH Capacitaciones" />
      </button>
      <div class="cf__hdr-acciones">
        <button class="cf__link" @click="router.push('/planes')">Planes</button>
        <button class="cf__link cf__link--solido" @click="router.push('/tienda')">Ver catálogo</button>
      </div>
    </header>

    <!-- ══════════ CONTENIDO ══════════ -->
    <div class="cf__scroll">
      <section v-for="(p, i) in PASOS" :id="`paso-${i}`" :key="p.clave" class="cf__sec">
        <article :class="['tarjeta', i % 2 === 1 && 'tarjeta--der']">
          <div class="tarjeta__linea" v-reveal />
          <p class="tarjeta__tag" v-reveal>{{ p.num }} — {{ p.nombre }}</p>
          <h2 class="tarjeta__titulo" v-reveal="1">
            <template v-for="(linea, k) in p.titulo" :key="k">{{ linea }}<br /></template>
          </h2>
          <p class="tarjeta__cuerpo" v-reveal="2">{{ p.cuerpo }}</p>

          <div v-if="i === N - 1" class="tarjeta__ctas" v-reveal="3">
            <button class="cta" @click="router.push('/tienda')">Ver catálogo</button>
            <button class="cta cta--fantasma" @click="router.push('/planes')">Comparar planes</button>
          </div>
        </article>
      </section>
    </div>

    <!-- Rejilla estática: sustituye a la escena con prefers-reduced-motion -->
    <div v-if="reducido" class="planos">
      <article v-for="p in PASOS" :key="`pl-${p.clave}`" class="plano">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
          stroke-linejoin="round">
          <path :d="p.icono" />
        </svg>
        <span class="plano__num">{{ p.num }}</span>
        <span class="plano__nombre">{{ p.nombre }}</span>
      </article>
    </div>
  </div>
</template>

<style scoped>
.cf {
  --tinta: #f4f1ea;
  /* 0.82 en vez de 0.62: sobre el panel da ~9:1, holgado por encima del 4.5:1
     de la AA. A 0.62 el cuerpo se caía justo cuando el cielo oscurecía. */
  --tinta-suave: rgba(244, 241, 234, 0.82);
  --acento: #f5c54e;
  /* Fondo del panel fijo y casi opaco. La escena cambia de la tarde a la noche,
     así que el contraste del texto NO puede depender de qué haya detrás: se
     resuelve contra el panel y sale igual al 0% que al 100% del recorrido. */
  --panel: rgba(13, 11, 30, 0.92);
  --panel-borde: rgba(245, 197, 78, 0.34);
  --hueco: clamp(1rem, 4vw, 2rem);
  position: relative;
  color: var(--tinta);
  background: #0f1030;
}

/* main.css declara `h1..h6 { color: var(--dark) }` a nivel global. Una regla
   directa gana siempre a la herencia, así que dentro de la escena hay que
   devolver el color de forma explícita en lugar de confiar en heredarlo. */
.cf :is(h1, h2, h3) {
  color: var(--tinta);
}

/* ── Paisaje ───────────────────────────────────────────── */
.cf__paisaje {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

.cf__paisaje svg {
  width: 100%;
  height: 100%;
  display: block;
}

.cf__paisaje path,
.cf__nubes path {
  will-change: transform;
}

.cf__estrellas {
  transition: opacity 0.2s linear;
}

/* ── Cubo ──────────────────────────────────────────────── */
.cf__escena {
  position: fixed;
  inset: 0;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  /* La perspectiva vive en el contenedor para que las seis caras compartan
     un único punto de fuga; si fuera por cara, cada una tendría el suyo. */
  perspective: 1200px;
  pointer-events: none;
}

.cubo {
  --lado: min(62vw, 58vh, 460px);
  width: var(--lado);
  height: var(--lado);
  position: relative;
  transform-style: preserve-3d;
  will-change: transform;
}

.cara {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.9rem;
  /* backface-visibility oculta la cara opuesta: sin esto se transparentaría
     el número invertido de la cara de atrás sobre la que estás leyendo. */
  backface-visibility: hidden;
  background:
    repeating-linear-gradient(0deg, rgba(255, 255, 255, 0.03) 0 1px, transparent 1px 44px),
    repeating-linear-gradient(90deg, rgba(255, 255, 255, 0.03) 0 1px, transparent 1px 44px),
    var(--panel);
  border: 1px solid var(--panel-borde);
  backdrop-filter: blur(3px);
  -webkit-backdrop-filter: blur(3px);
}

.cara__num {
  font-size: clamp(2.6rem, 9vw, 5rem);
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1;
  color: var(--acento);
}

.cara__icono {
  width: clamp(28px, 7vw, 44px);
  height: auto;
  color: var(--tinta);
  opacity: 0.85;
}

.cara__nombre {
  font-size: clamp(0.62rem, 1.6vw, 0.78rem);
  font-weight: 700;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: var(--tinta-suave);
}

.cara[data-cara='0'] {
  transform: rotateX(-90deg) translateZ(calc(var(--lado) / 2));
}

.cara[data-cara='1'] {
  transform: translateZ(calc(var(--lado) / 2));
}

.cara[data-cara='2'] {
  transform: rotateY(90deg) translateZ(calc(var(--lado) / 2));
}

.cara[data-cara='3'] {
  transform: rotateY(180deg) translateZ(calc(var(--lado) / 2));
}

.cara[data-cara='4'] {
  transform: rotateY(-90deg) translateZ(calc(var(--lado) / 2));
}

.cara[data-cara='5'] {
  transform: rotateX(90deg) translateZ(calc(var(--lado) / 2));
}

/* ── HUD ───────────────────────────────────────────────── */
.hud {
  position: fixed;
  top: var(--hueco);
  right: var(--hueco);
  z-index: 20;
  text-align: right;
  font-size: 0.66rem;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--tinta);
  font-variant-numeric: tabular-nums;
  /* El HUD flota sobre el cielo desnudo, que pasa de casi blanco a casi negro.
     La sombra es lo único que lo mantiene legible en los dos extremos. */
  text-shadow: 0 1px 6px rgba(0, 0, 0, 0.75);
}

.hud__barra {
  width: 7.5rem;
  height: 2px;
  margin: 0.5rem 0 0 auto;
  border-radius: 2px;
  background: rgba(16, 14, 38, 0.45);
  overflow: hidden;
}

.hud__fill {
  height: 100%;
  background: var(--acento);
}

.hud__etiqueta {
  margin-top: 0.4rem;
  color: var(--acento);
  font-size: 0.62rem;
  text-shadow: 0 1px 6px rgba(0, 0, 0, 0.8);
}

.strip {
  position: fixed;
  left: var(--hueco);
  top: 50%;
  transform: translateY(-50%);
  z-index: 20;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.strip__dot {
  position: relative;
  width: 7px;
  height: 7px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  /* Relleno oscuro con anillo claro: los dots flotan sobre el cielo desnudo,
     que va de casi blanco a casi negro. Un punto de un solo color desaparece
     en uno de los dos extremos — en claro se perdían al inicio del recorrido. */
  background: rgba(16, 14, 38, 0.6);
  box-shadow: 0 0 0 1.5px rgba(244, 241, 234, 0.7);
  cursor: pointer;
  transition: background 0.3s, transform 0.3s, box-shadow 0.3s;
}

/* Área táctil de 24px sin engordar el punto visible. */
.strip__dot::before {
  content: '';
  position: absolute;
  inset: -9px;
}

.strip__dot.is-on {
  background: var(--acento);
  box-shadow: 0 0 0 1.5px rgba(16, 14, 38, 0.75);
  transform: scale(1.7);
}

.strip__dot:focus-visible {
  outline: 2px solid var(--acento);
  outline-offset: 5px;
}

.caption {
  position: fixed;
  bottom: var(--hueco);
  left: 50%;
  transform: translateX(-50%);
  z-index: 20;
  text-align: center;
  pointer-events: none;
  user-select: none;
}

.caption__num {
  font-size: 0.6rem;
  letter-spacing: 0.3em;
  color: var(--acento);
  margin-bottom: 0.15rem;
  text-shadow: 0 1px 6px rgba(0, 0, 0, 0.75);
}

.caption__nombre {
  font-size: clamp(1.6rem, 5vw, 3rem);
  font-weight: 800;
  letter-spacing: 0.05em;
  line-height: 1;
  /* Marca de agua deliberada: no transporta información —el mismo dato está en
     el HUD y en la tarjeta—, así que se queda por debajo del umbral de lectura
     a propósito, con aria-hidden en el marcado. */
  color: rgba(244, 241, 234, 0.3);
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.4);
}

/* ── Cabecera ──────────────────────────────────────────── */
.cf__hdr {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem clamp(1rem, 4vw, 2rem);
  /* Sin fondo: el degradado del paisaje ya da contraste y un panel sólido
     partiría la escena justo donde debe sentirse continua. */
}

.cf__brand {
  background: none;
  border: 0;
  padding: 0;
  cursor: pointer;
}

.cf__brand img {
  height: 34px;
  width: auto;
  display: block;
}

.cf__hdr-acciones {
  display: flex;
  gap: 0.5rem;
}

.cf__link {
  border: 1px solid rgba(244, 241, 234, 0.3);
  background: rgba(15, 16, 48, 0.4);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  color: var(--tinta);
  border-radius: 999px;
  padding: 0.45rem 1rem;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s, color 0.2s;
}

.cf__link:hover {
  border-color: var(--acento);
  color: var(--acento);
}

.cf__link--solido {
  background: var(--acento);
  border-color: var(--acento);
  color: #14122c;
}

.cf__link--solido:hover {
  filter: brightness(1.08);
  color: #14122c;
}

/* ── Secciones ─────────────────────────────────────────── */
.cf__scroll {
  position: relative;
  z-index: 10;
}

.cf__sec {
  min-height: 100vh; min-height: 100dvh;
  display: flex;
  align-items: center;
  /* Aire a la derecha para el HUD y a la izquierda para el strip de pasos. */
  padding: 6rem calc(4rem + var(--hueco)) 6rem calc(3rem + var(--hueco));
}

.tarjeta {
  max-width: 24rem;
  padding: 2rem 1.75rem;
  background: var(--panel);
  border-left: 2px solid var(--panel-borde);
  /* La sombra despega el panel del cielo de noche, cuando fondo y panel llegan
     a tener casi la misma luminancia y el borde solo ya no basta. */
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(7px) saturate(120%);
  -webkit-backdrop-filter: blur(7px) saturate(120%);
}

.tarjeta--der {
  margin-left: auto;
  text-align: right;
  border-left: 0;
  border-right: 2px solid var(--panel-borde);
}

.tarjeta--der .tarjeta__linea {
  margin-left: auto;
}

.tarjeta__linea {
  width: 3rem;
  height: 1px;
  background: var(--acento);
  margin-bottom: 1.1rem;
}

.tarjeta__tag {
  margin: 0 0 1rem;
  font-size: 0.62rem;
  letter-spacing: 0.26em;
  text-transform: uppercase;
  color: var(--acento);
}

.tarjeta__titulo {
  margin: 0;
  color: var(--tinta);
  font-size: clamp(2rem, 5.5vw, 3.4rem);
  font-weight: 800;
  letter-spacing: -0.02em;
  line-height: 0.96;
}

.tarjeta__cuerpo {
  margin: 1.15rem 0 0;
  font-size: 0.95rem;
  line-height: 1.7;
  color: var(--tinta-suave);
}

.tarjeta__ctas {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin-top: 1.6rem;
}

.tarjeta--der .tarjeta__ctas {
  justify-content: flex-end;
}

.cta {
  border: 1px solid var(--acento);
  background: var(--acento);
  color: #14122c;
  border-radius: 999px;
  padding: 0.6rem 1.2rem;
  font-size: 0.84rem;
  font-weight: 700;
  cursor: pointer;
  transition: filter 0.2s, background 0.2s, color 0.2s;
}

.cta:hover {
  filter: brightness(1.08);
}

.cta--fantasma {
  background: transparent;
  color: var(--acento);
}

.cta--fantasma:hover {
  background: var(--acento);
  color: #14122c;
}

/* ── Alternativa estática ──────────────────────────────── */
.planos {
  position: relative;
  z-index: 10;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
  max-width: 1100px;
  margin: 0 auto;
  padding: 2rem var(--hueco) 4rem;
}

.plano {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1.4rem 1rem;
  background: var(--panel);
  border: 1px solid var(--panel-borde);
  border-radius: 14px;
  text-align: center;
}

.plano svg {
  width: 26px;
  height: 26px;
  color: var(--tinta);
  opacity: 0.85;
}

.plano__num {
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--acento);
}

.plano__nombre {
  font-size: 0.66rem;
  letter-spacing: 0.2em;
  color: var(--tinta-suave);
}

/* ── Responsive ────────────────────────────────────────── */
@media (max-width: 1023px) {
  .strip {
    display: none;
  }

  .cubo {
    --lado: min(70vw, 44vh, 320px);
  }

  .cf__sec {
    /* El cubo ocupa el centro: las tarjetas bajan al tercio inferior para no
       taparlo, que es donde el pulgar además las alcanza. */
    min-height: 118vh;
    align-items: flex-end;
    padding: 0 var(--hueco) 4.5rem;
  }

  .tarjeta,
  .tarjeta--der {
    max-width: 100%;
    margin: 0;
    text-align: left;
    border-left: 2px solid var(--panel-borde);
    border-right: 0;
    padding: 1.35rem 1.15rem;
  }

  .tarjeta--der .tarjeta__linea {
    margin-left: 0;
  }

  .tarjeta--der .tarjeta__ctas {
    justify-content: flex-start;
  }

  .caption__nombre {
    font-size: 1.5rem;
  }
}

/* La escena entera desaparece: quedan la rejilla estática y las tarjetas. */
@media (prefers-reduced-motion: reduce) {
  .cf__escena {
    display: none;
  }

  .cf__sec {
    min-height: auto;
    padding: 2.5rem var(--hueco);
  }

  .caption {
    display: none;
  }

  .strip__dot {
    transition: none;
  }

  .cf__estrellas {
    transition: none;
  }
}
</style>
