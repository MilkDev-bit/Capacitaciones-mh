<script setup lang="ts">
import { onMounted, onBeforeUnmount, computed } from 'vue'

/**
 * Pantalla de timbre: la que sustituye al "comparte este código".
 *
 * Cubre los dos lados de una llamada que todavía no empieza — el que llama y
 * el que recibe — porque son el mismo momento visto desde dos sitios y
 * separarlos en dos componentes duplicaría el tono, la cuenta atrás y el
 * cierre.
 */

const props = defineProps<{
  /** 'entrante' muestra aceptar/rechazar; 'saliente' solo colgar. */
  modo: 'entrante' | 'saliente'
  nombre: string
  isGroup: boolean
  /** Segundos que quedan de timbre. */
  restantes: number
}>()

const emit = defineEmits<{
  (e: 'aceptar'): void
  (e: 'rechazar'): void
  (e: 'colgar'): void
}>()

const iniciales = computed(() =>
  (props.nombre || '?')
    .split(' ')
    .slice(0, 2)
    .map((p) => p.charAt(0).toUpperCase())
    .join(''),
)

const subtitulo = computed(() => {
  if (props.modo === 'saliente') return 'Llamando…'
  return props.isGroup ? 'Videollamada de grupo' : 'Videollamada entrante'
})

/**
 * Tono generado con la Web Audio API en lugar de un archivo de audio.
 *
 * Un MP3 son ~30 KB en el bundle y, sobre todo, los navegadores bloquean la
 * reproducción automática de <audio> sin interacción previa del usuario. Un
 * oscilador sí puede sonar tras un gesto (abrir el chat), y si el contexto
 * viene bloqueado simplemente no suena en vez de lanzar una excepción.
 */
let audioCtx: AudioContext | null = null
let intervaloTono: ReturnType<typeof setInterval> | undefined

function pulso(frecuencia: number, duracion: number, volumen: number) {
  if (!audioCtx) return
  const osc = audioCtx.createOscillator()
  const gain = audioCtx.createGain()
  osc.type = 'sine'
  osc.frequency.value = frecuencia
  // Rampa de entrada y salida: un oscilador que arranca y para en seco
  // produce un chasquido audible.
  const t = audioCtx.currentTime
  gain.gain.setValueAtTime(0, t)
  gain.gain.linearRampToValueAtTime(volumen, t + 0.05)
  gain.gain.linearRampToValueAtTime(0, t + duracion)
  osc.connect(gain).connect(audioCtx.destination)
  osc.start(t)
  osc.stop(t + duracion)
}

function tono() {
  // Entrante: dos pulsos agudos, como un timbre. Saliente: uno grave y
  // largo, como el tono de llamada de la red telefónica.
  if (props.modo === 'entrante') {
    pulso(880, 0.35, 0.16)
    setTimeout(() => pulso(660, 0.35, 0.16), 420)
  } else {
    pulso(420, 0.6, 0.08)
  }
}

onMounted(() => {
  // prefers-reduced-motion también cubre a quien pide menos estímulo: si lo
  // tiene activo, la pantalla se muestra sin sonido ni animación.
  if (window.matchMedia?.('(prefers-reduced-motion: reduce)').matches) return
  try {
    audioCtx = new AudioContext()
    tono()
    intervaloTono = setInterval(tono, props.modo === 'entrante' ? 1800 : 2600)
  } catch {
    audioCtx = null
  }
})

onBeforeUnmount(() => {
  if (intervaloTono) clearInterval(intervaloTono)
  audioCtx?.close().catch(() => {})
  audioCtx = null
})
</script>

<template>
  <div class="llamada-overlay" role="dialog" aria-modal="true" :aria-label="subtitulo">
    <div class="llamada-card">
      <div class="avatar-pulso">
        <div class="avatar">{{ iniciales }}</div>
      </div>

      <p class="subtitulo">{{ subtitulo }}</p>
      <h2 class="nombre">{{ nombre }}</h2>
      <p class="cuenta" v-if="restantes > 0">{{ restantes }}s</p>

      <div class="acciones">
        <template v-if="modo === 'entrante'">
          <button type="button" class="btn-accion rechazar" @click="emit('rechazar')" aria-label="Rechazar llamada">
            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.79 19.79 0 0 1 2.12 4.18 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.13.96.36 1.9.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.9.34 1.85.57 2.81.7A2 2 0 0 1 22 16.92z" />
              <line x1="2" y1="2" x2="22" y2="22" />
            </svg>
          </button>
          <button type="button" class="btn-accion aceptar" @click="emit('aceptar')" aria-label="Aceptar llamada">
            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M23 7l-7 5 7 5V7z" /><rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
            </svg>
          </button>
        </template>
        <button v-else type="button" class="btn-accion rechazar" @click="emit('colgar')" aria-label="Cancelar llamada">
          <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.79 19.79 0 0 1 2.12 4.18 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.13.96.36 1.9.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.9.34 1.85.57 2.81.7A2 2 0 0 1 22 16.92z" />
            <line x1="2" y1="2" x2="22" y2="22" />
          </svg>
        </button>
      </div>

      <p class="pista">{{ modo === 'entrante' ? 'Contesta o rechaza' : 'Esperando respuesta…' }}</p>
    </div>
  </div>
</template>

<style scoped>
.llamada-overlay {
  position: fixed;
  inset: 0;
  z-index: 10001;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(15, 15, 17, 0.82);
  backdrop-filter: blur(10px);
}

.llamada-card {
  width: 100%;
  max-width: 340px;
  text-align: center;
  color: #fff;
}

.avatar-pulso {
  position: relative;
  width: 116px;
  height: 116px;
  margin: 0 auto 26px;
  display: flex;
  align-items: center;
  justify-content: center;
}
/* Las ondas salen de un pseudo-elemento para no meter divs decorativos. */
.avatar-pulso::before,
.avatar-pulso::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 2px solid rgba(249, 115, 22, 0.55);
  animation: onda 2s ease-out infinite;
}
.avatar-pulso::after { animation-delay: 1s; }

@keyframes onda {
  0%   { transform: scale(1);    opacity: 0.7; }
  100% { transform: scale(1.55); opacity: 0;   }
}

.avatar {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  background: linear-gradient(140deg, #f97316, #ea580c);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 800;
  letter-spacing: 1px;
  position: relative;
  z-index: 1;
}

.subtitulo { margin: 0 0 6px; font-size: 0.85rem; color: rgba(255,255,255,0.65); }
.nombre { margin: 0 0 6px; font-size: 1.5rem; font-weight: 800; }
.cuenta { margin: 0; font-size: 0.85rem; color: rgba(255,255,255,0.45); font-variant-numeric: tabular-nums; }

.acciones {
  display: flex;
  justify-content: center;
  gap: 40px;
  margin: 36px 0 18px;
}

.btn-accion {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: none;
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.18s, filter 0.18s;
}
.btn-accion:hover { transform: scale(1.08); filter: brightness(1.1); }
.btn-accion:active { transform: scale(0.96); }
.aceptar  { background: #10b981; }
.rechazar { background: #ef4444; }

.pista { margin: 0; font-size: 0.78rem; color: rgba(255,255,255,0.4); }

@media (prefers-reduced-motion: reduce) {
  .avatar-pulso::before,
  .avatar-pulso::after { animation: none; opacity: 0.35; }
  .btn-accion:hover { transform: none; }
}
</style>
