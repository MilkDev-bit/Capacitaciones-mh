<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import Plyr from 'plyr'
import 'plyr/dist/plyr.css'

const props = defineProps<{
  src: string
  /** Miniatura de la lección: evita el rectángulo negro mientras carga. */
  poster?: string
  // Posición guardada en segundos (opcional)
  savedTime?: number
  /**
   * Impide saltar hacia delante: hay que ver el video para avanzar.
   *
   * Retroceder SÍ se permite. Bloquear también el retroceso convertiría
   * cualquier despiste en volver a ver el video entero, y no aporta nada:
   * lo que se quiere evitar es certificar a quien no vio el contenido.
   */
  bloquearAdelanto?: boolean
  /** Ya completada: se libera la barra para poder repasar sin fricción. */
  yaCompletada?: boolean
}>()

const emit = defineEmits<{
  // Emite el tiempo actual periódicamente para guardarlo
  (e: 'timeupdate', seconds: number): void
  (e: 'ended'): void
  /** El usuario intentó adelantar mientras estaba bloqueado. */
  (e: 'seek-bloqueado'): void
}>()

/**
 * Marca de agua: el segundo más lejano realmente reproducido.
 *
 * Es lo único que hace de tope. Se sube solo desde `timeupdate`, es decir
 * cuando el video de verdad avanzó; si se actualizara en `seeking` bastaría
 * arrastrar la barra para subirla y el bloqueo no serviría de nada.
 */
let maxVisto = 0
/** Margen para el goteo normal de timeupdate, que llega a saltos de ~250 ms. */
const TOLERANCIA_S = 1.5

const bloqueoActivo = () => !!props.bloquearAdelanto && !props.yaCompletada

const videoEl = ref<HTMLVideoElement | null>(null)
let player: Plyr | null = null

/**
 * Estado de carga visible.
 *
 * Antes el reproductor mostraba un rectángulo negro inerte mientras el
 * navegador descargaba el archivo: los mismos segundos se perciben mucho más
 * largos sin ninguna señal de que algo está pasando. `cargando` alimenta el
 * spinner y `tardando` aparece a los 4 s para explicar la espera.
 */
const cargando = ref(true)
const tardando = ref(false)
const errorCarga = ref(false)
let temporizadorTardanza: ReturnType<typeof setTimeout> | undefined

function marcarCargando() {
  cargando.value = true
  errorCarga.value = false
  tardando.value = false
  clearTimeout(temporizadorTardanza)
  temporizadorTardanza = setTimeout(() => { tardando.value = true }, 4000)
}

function marcarListo() {
  cargando.value = false
  tardando.value = false
  clearTimeout(temporizadorTardanza)
}

/**
 * Construye (o reconstruye) el reproductor.
 *
 * `mismoMedio` indica que se reinicializa sobre el MISMO vídeo, ya cargado —el
 * caso de levantar el bloqueo al completar la lección—. Importa porque en esa
 * situación el elemento no vuelve a emitir `loadedmetadata` y hay que retirar
 * el overlay a mano. No vale con mirar `readyState` siempre: al cambiar de
 * lección el watch corre ANTES de que el DOM actualice el src, así que
 * `readyState` aún describe el vídeo anterior y el spinner del nuevo
 * desaparecería antes de tiempo.
 */
function initPlayer(mismoMedio = false) {
  if (!videoEl.value) return
  if (player) { player.destroy(); player = null }
  marcarCargando()

  const conBloqueo = bloqueoActivo()

  player = new Plyr(videoEl.value, {
    controls: [
      'play-large', 'rewind', 'play',
      // Adelantar 10s se quita del todo mientras hay bloqueo: dejarlo visible
      // y que no haga nada se lee como que el reproductor está roto.
      ...(conBloqueo ? [] : ['fast-forward']),
      'progress',
      'current-time', 'duration', 'mute', 'volume', 'captions',
      'settings', 'pip', 'fullscreen',
    ],
    settings: conBloqueo ? ['captions', 'quality'] : ['captions', 'quality', 'speed', 'loop'],
    // Sin bloqueo se puede acelerar; con bloqueo no, porque a 2x el video
    // termina en la mitad de tiempo y el requisito deja de significar nada.
    speed: conBloqueo
      ? { selected: 1, options: [1] }
      : { selected: 1, options: [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2] },
    keyboard: { focused: true, global: false },
    tooltips: { controls: true, seek: true },
    seekTime: 10,
    i18n: {
      restart: 'Reiniciar',
      rewind: 'Retroceder {seektime}s',
      play: 'Reproducir',
      pause: 'Pausar',
      fastForward: 'Adelantar {seektime}s',
      seek: 'Buscar',
      seekLabel: '{currentTime} de {duration}',
      played: 'Reproducido',
      buffered: 'Almacenado',
      currentTime: 'Tiempo actual',
      duration: 'Duración',
      volume: 'Volumen',
      mute: 'Silenciar',
      unmute: 'Activar sonido',
      enableCaptions: 'Activar subtítulos',
      disableCaptions: 'Desactivar subtítulos',
      download: 'Descargar',
      enterFullscreen: 'Pantalla completa',
      exitFullscreen: 'Salir de pantalla completa',
      frameTitle: 'Reproductor de {title}',
      captions: 'Subtítulos',
      settings: 'Configuración',
      pip: 'Imagen en imagen',
      menuBack: 'Volver',
      speed: 'Velocidad',
      normal: 'Normal',
      quality: 'Calidad',
      loop: 'Repetir',
      start: 'Inicio',
      end: 'Fin',
      all: 'Todo',
      reset: 'Reiniciar',
      disabled: 'Desactivado',
      enabled: 'Activado',
      advertisement: 'Anuncio',
      qualityBadge: { 2160: '4K', 1440: 'HD', 1080: 'HD', 720: 'HD', 576: 'SD', 480: 'SD' },
    },
  })

  // Reanudar desde posición guardada
  const seekToSaved = () => {
    if (player && props.savedTime && props.savedTime > 0) {
      // El tope sube ANTES de mover la cabeza: ese tramo ya se vio en una
      // sesión anterior. Sin esto, reanudar se leería como un salto prohibido
      // y el guardia devolvería al alumno al segundo cero.
      if (props.savedTime > maxVisto) maxVisto = props.savedTime
      if (Math.abs(player.currentTime - props.savedTime) > 2) {
        player.currentTime = props.savedTime
      }
    }
  }
  if (props.savedTime && props.savedTime > 0 && props.savedTime > maxVisto) {
    maxVisto = props.savedTime
  }
  if (props.savedTime && props.savedTime > 0) {
    player.once('ready', seekToSaved)
    player.once('canplay', seekToSaved)
  }

  // Emitir progreso periódicamente y ante eventos clave para que el padre lo persista en BD
  let lastEmit = -1
  let endedEmitted = false
  player.on('timeupdate', () => {
    if (!player) return
    // Solo aquí sube el tope, y solo hacia arriba: es la prueba de que el
    // video avanzó de verdad en lugar de que alguien movió la barra.
    if (player.currentTime > maxVisto) maxVisto = player.currentTime
    const t = Math.floor(player.currentTime)
    if (t > 0 && Math.abs(t - lastEmit) >= 3) {
      lastEmit = t
      emit('timeupdate', t)
    }
    if (!endedEmitted && player.duration > 0 && player.currentTime >= player.duration - 1) {
      endedEmitted = true
      emit('ended')
    }
  })

  player.on('pause', () => {
    if (!player) return
    const t = Math.floor(player.currentTime)
    if (t > 0) emit('timeupdate', t)
  })

  /**
   * Guardia del salto hacia delante.
   *
   * Se escuchan `seeking` y `seeked`: el primero atrapa el arrastre en cuanto
   * empieza, el segundo cubre los saltos que el navegador aplica de golpe sin
   * pasar por `seeking` (teclado y algunos móviles).
   */
  const frenarAdelanto = () => {
    if (!player || !bloqueoActivo()) return false
    if (player.currentTime <= maxVisto + TOLERANCIA_S) return false
    player.currentTime = maxVisto
    emit('seek-bloqueado')
    return true
  }
  player.on('seeking', frenarAdelanto)

  player.on('seeked', () => {
    if (!player) return
    if (frenarAdelanto()) return
    const t = Math.floor(player.currentTime)
    if (t > 0) emit('timeupdate', t)
  })

  player.on('ended', () => {
    // Un vídeo terminado NO está cargando.
    //
    // Al llegar al final, el navegador dispara `waiting` —el elemento se queda
    // esperando datos que ya no van a llegar— y eso encendía el overlay. Como
    // `playing` es lo único que lo apagaba y ya no vuelve a dispararse, el
    // spinner se quedaba para siempre sobre el último fotograma.
    marcarListo()
    if (!endedEmitted) {
      endedEmitted = true
      emit('ended')
    }
  })

  // `loadedmetadata` es la señal correcta para retirar el spinner: significa
  // que el navegador ya leyó el atom moov y sabe duración y dimensiones, que
  // es justo el punto donde el reproductor pasa a ser usable. Esperar a
  // `canplay` mantendría el overlay durante todo el primer buffer.
  player.on('loadedmetadata', marcarListo)
  player.on('playing', marcarListo)

  // Este es el caso que dejaba el overlay colgado al terminar un vídeo: al
  // completarse la lección se reconstruye el reproductor (watch de
  // `yaCompletada`, para que reaparezcan velocidad y avance), pero el elemento
  // multimedia NO se recarga. `loadedmetadata` no vuelve a dispararse porque ya
  // ocurrió, `playing` tampoco porque la reproducción terminó, y `error` menos.
  // Sin ningún evento que lo apague, el spinner se quedaba indefinidamente.
  //
  // readyState >= HAVE_METADATA (1) significa que ya se conocen duración y
  // dimensiones: exactamente lo que espera marcarListo.
  if (mismoMedio && videoEl.value.readyState >= 1) marcarListo()
  // Pausar tampoco es cargar. Sin esto, un `waiting` justo antes de una pausa
  // dejaba el overlay puesto hasta que el usuario le diera a reproducir.
  player.on('pause', marcarListo)
  player.on('waiting', () => {
    // `waiting` al final del vídeo es un falso positivo: no hay más datos que
    // esperar. Encender el spinner ahí lo dejaría encendido para siempre,
    // porque en estado terminado ya no llega ningún evento que lo apague.
    if (player?.ended) return
    cargando.value = true
  })
  player.on('error', () => {
    marcarListo()
    errorCarga.value = true
  })
}

onMounted(() => initPlayer())

watch(() => props.savedTime, (newTime) => {
  if (player && newTime && newTime > 0 && Math.abs((player.currentTime || 0) - newTime) > 3) {
    player.currentTime = newTime
  }
})

// Cuando cambia el src (el usuario cambia de lección), reinicializar
// Otro video es otro tope: si no se reinicia, el minuto 12 del anterior
// dejaría saltar libremente los primeros 12 minutos del nuevo.
watch(() => props.src, () => {
  maxVisto = 0
  if (videoEl.value) initPlayer()
})

// Al completarse la lección el bloqueo se levanta, y eso cambia los controles:
// hay que reconstruir el reproductor para que reaparezcan velocidad y avance.
watch(() => props.yaCompletada, (ahora, antes) => {
  if (ahora && !antes && videoEl.value) {
    const t = player?.currentTime ?? 0
    // true: es el mismo vídeo, solo cambian los controles.
    initPlayer(true)
    if (t > 0) player?.once('ready', () => { if (player) player.currentTime = t })
  }
})

onBeforeUnmount(() => {
  clearTimeout(temporizadorTardanza)
  player?.destroy()
  player = null
})
</script>

<template>
  <div class="plyr-wrapper">
    <!--
      Sin `crossorigin`: el atributo obliga al navegador a tratar el MP4 como
      recurso CORS y a esperar la cabecera Access-Control-Allow-Origin antes de
      empezar a decodificar. Solo hace falta para pintar el video en un canvas
      o cargar subtítulos de otro origen, y aquí no se hace ninguna de las dos.

      `preload="metadata"` pide solo la cabecera del contenedor en lugar de
      dejar la decisión al navegador (Chrome tiende a `auto` y se trae varios
      megabytes antes del primer fotograma).
    -->
    <video
      ref="videoEl"
      :src="src"
      :poster="poster || undefined"
      preload="metadata"
      playsinline
    />

    <div v-if="cargando || errorCarga" class="video-overlay">
      <template v-if="errorCarga">
        <p class="overlay-text">No se pudo cargar el video.</p>
        <button type="button" class="overlay-btn" @click="initPlayer()">Reintentar</button>
      </template>
      <template v-else>
        <span class="overlay-spinner" aria-hidden="true"></span>
        <p class="overlay-text">{{ tardando ? 'Preparando el video…' : 'Cargando…' }}</p>
      </template>
    </div>
  </div>
</template>

<style scoped>
.plyr-wrapper {
  position: relative;
  width: 100%;
  border-radius: 12px;
  overflow: hidden;
  background: #000;
}

/* Overlay de carga/error. `pointer-events: none` en el estado normal para no
   robarle el clic al botón grande de reproducir de Plyr. */
.video-overlay {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(2px);
}
.overlay-spinner {
  width: 34px;
  height: 34px;
  border: 3px solid rgba(255, 255, 255, 0.25);
  border-top-color: #f97316;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) {
  .overlay-spinner { animation-duration: 2s; }
}
.overlay-text { margin: 0; color: rgba(255, 255, 255, 0.9); font-size: 0.85rem; font-weight: 600; }
.overlay-btn {
  background: #f97316; color: #fff; border: none; border-radius: 8px;
  padding: 8px 18px; font-size: 0.85rem; font-weight: 700; cursor: pointer;
}
.overlay-btn:hover { background: #ea580c; }
.plyr-wrapper :deep(.plyr) {
  width: 100%;
  border-radius: 12px;
}
/* Personalizar color de acento al naranja de marca */
.plyr-wrapper :deep(.plyr--video .plyr__control.plyr__tab-focus),
.plyr-wrapper :deep(.plyr--video .plyr__control:hover),
.plyr-wrapper :deep(.plyr--video .plyr__control[aria-expanded='true']) {
  background: #f97316;
}
.plyr-wrapper :deep(.plyr__control--overlaid) {
  background: rgba(249, 115, 22, 0.85);
}
.plyr-wrapper :deep(.plyr--full-ui input[type='range']) {
  color: #f97316;
}
</style>
