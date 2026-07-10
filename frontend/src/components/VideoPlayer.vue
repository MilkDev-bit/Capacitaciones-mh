<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import Plyr from 'plyr'
import 'plyr/dist/plyr.css'

const props = defineProps<{
  src: string
  // Posición guardada en segundos (opcional)
  savedTime?: number
}>()

const emit = defineEmits<{
  // Emite el tiempo actual periódicamente para guardarlo
  (e: 'timeupdate', seconds: number): void
  (e: 'ended'): void
}>()

const videoEl = ref<HTMLVideoElement | null>(null)
let player: Plyr | null = null

function initPlayer() {
  if (!videoEl.value) return
  if (player) { player.destroy(); player = null }

  player = new Plyr(videoEl.value, {
    controls: [
      'play-large', 'rewind', 'play', 'fast-forward', 'progress',
      'current-time', 'duration', 'mute', 'volume', 'captions',
      'settings', 'pip', 'fullscreen',
    ],
    settings: ['captions', 'quality', 'speed', 'loop'],
    speed: { selected: 1, options: [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2] },
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
  if (props.savedTime && props.savedTime > 0) {
    player.once('ready', () => {
      if (player && props.savedTime) player.currentTime = props.savedTime
    })
  }

  // Emitir progreso periódicamente y ante eventos clave para que el padre lo persista en BD
  let lastEmit = -1
  let endedEmitted = false
  player.on('timeupdate', () => {
    if (!player) return
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

  player.on('seeked', () => {
    if (!player) return
    const t = Math.floor(player.currentTime)
    if (t > 0) emit('timeupdate', t)
  })

  player.on('ended', () => {
    if (!endedEmitted) {
      endedEmitted = true
      emit('ended')
    }
  })
}

onMounted(() => initPlayer())

// Cuando cambia el src (el usuario cambia de lección), reinicializar
watch(() => props.src, () => {
  if (videoEl.value) initPlayer()
})

onBeforeUnmount(() => {
  player?.destroy()
  player = null
})
</script>

<template>
  <div class="plyr-wrapper">
    <video ref="videoEl" :src="src" playsinline crossorigin="anonymous" />
  </div>
</template>

<style scoped>
.plyr-wrapper {
  width: 100%;
  border-radius: 12px;
  overflow: hidden;
  background: #000;
}
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
