<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const emit = defineEmits<{
  (e: 'capture', file: File): void
  (e: 'close'): void
  (e: 'gallery'): void
}>()

const videoRef = ref<HTMLVideoElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

const stream = ref<MediaStream | null>(null)
const facingMode = ref<'user' | 'environment'>('environment')
const errorMsg = ref('')
const isLoading = ref(true)

async function initCamera() {
  stopCamera()
  isLoading.value = true
  errorMsg.value = ''
  try {
    const s = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: facingMode.value, width: { ideal: 1920 }, height: { ideal: 1080 } },
      audio: false
    })
    stream.value = s
    if (videoRef.value) {
      videoRef.value.srcObject = s
    }
  } catch (err: any) {
    errorMsg.value = 'No se pudo acceder a la cámara. Revisa los permisos.'
    console.error('Error al iniciar cámara:', err)
  } finally {
    isLoading.value = false
  }
}

function stopCamera() {
  if (stream.value) {
    stream.value.getTracks().forEach(t => t.stop())
    stream.value = null
  }
}

function toggleCamera() {
  facingMode.value = facingMode.value === 'user' ? 'environment' : 'user'
  initCamera()
}

function takePhoto() {
  if (!videoRef.value || !canvasRef.value) return
  const video = videoRef.value
  const canvas = canvasRef.value
  canvas.width = video.videoWidth
  canvas.height = video.videoHeight
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.drawImage(video, 0, 0, canvas.width, canvas.height)
  
  canvas.toBlob((blob) => {
    if (!blob) return
    const file = new File([blob], `captura_${Date.now()}.jpg`, { type: 'image/jpeg' })
    emit('capture', file)
    stopCamera()
  }, 'image/jpeg', 0.85)
}

function closeCamera() {
  stopCamera()
  emit('close')
}

function openGallery() {
  stopCamera()
  emit('gallery')
}

onMounted(() => {
  initCamera()
})

onUnmounted(() => {
  stopCamera()
})
</script>

<template>
  <div class="camera-overlay">
    <div class="camera-header">
      <button class="icon-btn" @click="closeCamera" aria-label="Cerrar cámara">
        <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>
      </button>
      <span class="camera-title">Cámara</span>
      <!-- Div fantasma para centrar el título -->
      <div style="width: 24px;"></div>
    </div>

    <div class="camera-view">
      <div v-if="isLoading" class="camera-loading">
        <span class="spinner"></span>
      </div>
      <div v-if="errorMsg" class="camera-error">
        <p>{{ errorMsg }}</p>
      </div>
      <video
        ref="videoRef"
        autoplay
        playsinline
        muted
        :class="['live-video', facingMode === 'user' ? 'mirrored' : '']"
      ></video>
      <canvas ref="canvasRef" style="display: none;"></canvas>
    </div>

    <div class="camera-controls">
      <!-- Botón de Galería a la izquierda -->
      <button class="gallery-btn" @click="openGallery" title="Abrir Galería">
        <svg width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
          <circle cx="8.5" cy="8.5" r="1.5"/>
          <polyline points="21 15 16 10 5 21"/>
        </svg>
      </button>

      <!-- Botón de Captura al centro -->
      <button class="capture-btn" @click="takePhoto" aria-label="Tomar foto">
        <div class="capture-inner"></div>
      </button>

      <!-- Botón para Cambiar Cámara a la derecha -->
      <button class="flip-btn" @click="toggleCamera" aria-label="Cambiar cámara">
        <svg width="28" height="28" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path d="M23 4v6h-6M1 20v-6h6"/>
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.camera-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #000;
  z-index: 9999;
  display: flex;
  flex-direction: column;
}

.camera-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1rem;
  background: linear-gradient(to bottom, rgba(0,0,0,0.6), transparent);
  z-index: 10;
}

.camera-title {
  color: #fff;
  font-weight: 600;
  font-size: 1.1rem;
}

.icon-btn {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 50%;
}

.camera-view {
  flex: 1;
  position: relative;
  background-color: #111;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.live-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.live-video.mirrored {
  transform: scaleX(-1);
}

.camera-loading, .camera-error {
  position: absolute;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  z-index: 5;
}

.camera-error {
  background: rgba(0,0,0,0.5);
  padding: 1rem;
  border-radius: 8px;
  text-align: center;
}

.spinner {
  width: 30px;
  height: 30px;
  border: 3px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin { 100% { transform: rotate(360deg); } }

.camera-controls {
  height: 120px;
  background-color: #000;
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding-bottom: env(safe-area-inset-bottom);
}

.capture-btn {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: none;
  border: 4px solid #fff;
  padding: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.capture-inner {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background-color: #fff;
  transition: transform 0.1s;
}

.capture-btn:active .capture-inner {
  transform: scale(0.9);
}

.flip-btn, .gallery-btn {
  background: rgba(255,255,255,0.15);
  border: none;
  color: #fff;
  width: 50px;
  height: 50px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s;
}

.flip-btn:hover, .gallery-btn:hover {
  background: rgba(255,255,255,0.25);
}
</style>
