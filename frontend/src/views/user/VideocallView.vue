<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import api from '../../api'
import iziToast from 'izitoast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const roomName = route.params.id as string
const codigo = route.query.codigo as string

const jitsiContainer = ref<HTMLElement | null>(null)
let jitsiApi: any = null

onMounted(async () => {
  if (!roomName || !codigo) {
    iziToast.warning({ title: 'Aviso', message: 'Falta la sala o el código de acceso.' })
    router.push('/usuario/dashboard')
    return
  }

  try {
    // Validar el código con el backend ANTES de unirse a Jitsi
    await api.post('/videocalls/join', { codigo })
  } catch (e: any) {
    iziToast.error({ title: 'Error', message: e.response?.data?.error || 'No tienes acceso a esta videollamada.' })
    router.push('/usuario/dashboard')
    return
  }
  
  const domain = 'meet.jit.si'
  const options = {
    roomName: roomName,
    width: '100%',
    height: '100%',
    parentNode: jitsiContainer.value,
    userInfo: {
      displayName: auth.user?.name || 'Estudiante'
    },
    configOverwrite: {
      prejoinPageEnabled: false,
      startWithAudioMuted: true,
      startWithVideoMuted: true,
    },
    interfaceConfigOverwrite: {
      TOOLBAR_BUTTONS: [
        'microphone', 'camera', 'closedcaptions', 'desktop', 'fullscreen',
        'fodeviceselection', 'hangup', 'profile', 'chat', 'settings', 'videoquality',
        'tileview', 'download', 'help', 'mute-everyone', 'security'
      ],
    }
  }

  // Cargar el script de Jitsi dinámicamente si no está en index.html
  if (!(window as any).JitsiMeetExternalAPI) {
    const script = document.createElement('script')
    script.src = `https://${domain}/external_api.js`
    script.async = true
    script.onload = () => {
      jitsiApi = new (window as any).JitsiMeetExternalAPI(domain, options)
      
      jitsiApi.addEventListeners({
        readyToClose: handleReadyToClose,
        videoConferenceLeft: handleVideoConferenceLeft
      })
    }
    document.head.appendChild(script)
  } else {
    jitsiApi = new (window as any).JitsiMeetExternalAPI(domain, options)
    jitsiApi.addEventListeners({
      readyToClose: handleReadyToClose,
      videoConferenceLeft: handleVideoConferenceLeft
    })
  }
})

async function liberarCodigo() {
  if (codigo) {
    try {
      await api.post('/videocalls/leave', { codigo })
    } catch (e) {
      console.error('Error al liberar código:', e)
    }
  }
}

onBeforeUnmount(() => {
  if (jitsiApi) jitsiApi.dispose()
  liberarCodigo()
})

const handleReadyToClose = () => {
  if (jitsiApi) { jitsiApi.dispose(); jitsiApi = null; }
  router.push('/usuario/dashboard')
}

const handleVideoConferenceLeft = () => {
  if (jitsiApi) { jitsiApi.dispose(); jitsiApi = null; }
  router.push('/usuario/dashboard')
}

</script>

<template>
  <div class="videocall-container">
    <div class="header-bar">
      <h2>Videollamada en Curso</h2>
      <button class="btn btn-outline" @click="router.push('/usuario/dashboard')">Salir</button>
    </div>
    <div ref="jitsiContainer" class="jitsi-wrapper"></div>
  </div>
</template>

<style scoped>
.videocall-container {
  position: fixed;
  top: 0;
  left: 0;
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  z-index: 9999;
  background: #111;
  color: #fff;
}
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: #1f2937;
  border-bottom: 1px solid #374151;
}
.header-bar h2 {
  margin: 0;
  font-size: 1.1rem;
}
.jitsi-wrapper {
  flex: 1;
}
</style>
