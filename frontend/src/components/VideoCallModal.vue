<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

declare global {
  interface Window {
    JitsiMeetExternalAPI: any
  }
}

const props = defineProps<{
  /** Sala negociada por el servidor. El cliente nunca la inventa. */
  roomName: string
  userName: string
  /** Dominio del servidor Jitsi autohospedado. */
  domain: string
  /**
   * Token JWT emitido por el gateway para ESTA sala. Sin él, Prosody rechaza
   * la conexión: es lo que sustituye al "código de acceso" — la autorización
   * la lleva el token, no el usuario.
   */
  jwt: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const jitsiContainer = ref<HTMLElement | null>(null)
let api: any = null
const isLoading = ref(true)
const errorMsg = ref('')

onMounted(() => {
  if (window.JitsiMeetExternalAPI) {
    initJitsi()
    return
  }
  const script = document.createElement('script')
  // El script se carga del propio servidor Jitsi, no de meet.jit.si: es el
  // que corresponde a la versión desplegada y evita depender de un tercero.
  script.src = `https://${props.domain}/external_api.js`
  script.async = true
  script.onload = initJitsi
  script.onerror = () => {
    isLoading.value = false
    errorMsg.value = 'No se pudo contactar con el servidor de videollamadas.'
  }
  document.head.appendChild(script)
})

function initJitsi() {
  if (!jitsiContainer.value) return

  const options = {
    roomName: props.roomName,
    jwt: props.jwt,
    width: '100%',
    height: '100%',
    parentNode: jitsiContainer.value,
    userInfo: { displayName: props.userName },
    configOverwrite: {
      // `prejoinConfig.enabled` es la clave vigente. La antigua
      // `prejoinPageEnabled` está obsoleta desde Jitsi 2022 y es la razón de
      // que siguiera apareciendo la pantalla de "Entrar a la reunión" pese a
      // estar puesta en false.
      prejoinConfig: { enabled: false },
      disableDeepLinking: true,
      // Entrar con cámara y micrófono activos, como en cualquier app de
      // mensajería: el usuario ya aceptó la llamada, pedirle otra confirmación
      // reintroduce la fricción que estamos quitando.
      startWithAudioMuted: false,
      startWithVideoMuted: false,
      disableInviteFunctions: true,
      // Sin lobby ni sala de espera: quién entra ya lo decidió el token.
      enableLobbyChat: false,
      hideConferenceSubject: true,
      hideConferenceTimer: false,
      // El chat de Jitsi sobra: la conversación ya está en la plataforma.
      disableChat: true,
    },
    interfaceConfigOverwrite: {
      SHOW_JITSI_WATERMARK: false,
      SHOW_WATERMARK_FOR_GUESTS: false,
      SHOW_BRAND_WATERMARK: false,
      DISABLE_JOIN_LEAVE_NOTIFICATIONS: false,
      MOBILE_APP_PROMO: false,
      TOOLBAR_BUTTONS: [
        'microphone', 'camera', 'desktop', 'fullscreen',
        'fodeviceselection', 'hangup', 'profile', 'settings',
        'videoquality', 'filmstrip', 'tileview', 'select-background',
      ],
    },
  }

  api = new (window as any).JitsiMeetExternalAPI(props.domain, options)
  isLoading.value = false

  api.addListener('videoConferenceLeft', () => emit('close'))
  api.addListener('readyToClose', () => emit('close'))
  // Un token inválido o vencido se manifiesta aquí; sin este listener el
  // usuario se queda mirando una pantalla en blanco sin saber por qué.
  api.addListener('errorOccurred', (e: any) => {
    if (e?.error?.isFatal) {
      errorMsg.value = 'La llamada se interrumpió. Vuelve a intentarlo.'
    }
  })
}

onUnmounted(() => {
  if (api) {
    api.dispose()
    api = null
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
      <div v-if="errorMsg" class="loading-state">
        <p>{{ errorMsg }}</p>
        <button type="button" class="retry-btn" @click="emit('close')">Cerrar</button>
      </div>
      <div v-else-if="isLoading" class="loading-state">
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

.retry-btn {
  background: #f97316;
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 10px 22px;
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
}
.retry-btn:hover { background: #ea580c; }

@keyframes spin { 100% { transform: rotate(360deg); } }
</style>
