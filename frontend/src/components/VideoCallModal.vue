<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

declare global {
  interface Window {
    JitsiMeetExternalAPI: any
  }
}

const props = defineProps<{
  roomName: string
  userName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const jitsiContainer = ref<HTMLElement | null>(null)
let api: any = null
const isLoading = ref(true)

onMounted(() => {
  // Load Jitsi external API script dynamically
  if (!window.JitsiMeetExternalAPI) {
    const script = document.createElement('script')
    script.src = 'https://meet.jit.si/external_api.js'
    script.async = true
    script.onload = initJitsi
    document.head.appendChild(script)
  } else {
    initJitsi()
  }
})

function initJitsi() {
  isLoading.value = false
  if (!jitsiContainer.value) return

  const domain = 'meet.jit.si'
  const options = {
    roomName: props.roomName,
    width: '100%',
    height: '100%',
    parentNode: jitsiContainer.value,
    userInfo: {
      displayName: props.userName
    },
    configOverwrite: {
      prejoinPageEnabled: false, // Skip prejoin page to enter directly
      disableDeepLinking: true // Prevent mobile app redirect prompts
    },
    interfaceConfigOverwrite: {
      // Customize interface slightly if needed
      SHOW_JITSI_WATERMARK: false
    }
  }

  // eslint-disable-next-line no-undef
  api = new (window as any).JitsiMeetExternalAPI(domain, options)

  api.addListener('videoConferenceLeft', () => {
    emit('close')
  })
}

onUnmounted(() => {
  if (api) {
    api.dispose()
  }
})
</script>

<template>
  <div class="video-call-modal">
    <div class="modal-header">
      <span>Videollamada segura</span>
      <button class="close-btn" @click="emit('close')" aria-label="Cerrar">✕</button>
    </div>
    <div class="jitsi-wrapper" ref="jitsiContainer">
      <div v-if="isLoading" class="loading-state">
        <span class="spinner"></span>
        <p>Conectando llamada...</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.video-call-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #000;
  z-index: 10000;
  display: flex;
  flex-direction: column;
}

.modal-header {
  height: 50px;
  background-color: #1a1a1a;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1rem;
  font-weight: 600;
}

.close-btn {
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: #fff;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.25);
}

.jitsi-wrapper {
  flex: 1;
  position: relative;
  background-color: #111;
}

.loading-state {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #fff;
  gap: 1rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255,255,255,0.3);
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin { 100% { transform: rotate(360deg); } }
</style>
