<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import api from '../../api'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const cursoId = route.params.id as string

const jitsiContainer = ref<HTMLElement | null>(null)
let jitsiApi: any = null

onMounted(() => {
  if (!cursoId) {
    alert('Falta el curso.')
    router.push('/instructor/dashboard')
    return
  }
  
  const domain = 'meet.jit.si'
  const options = {
    roomName: cursoId,
    width: '100%',
    height: '100%',
    parentNode: jitsiContainer.value,
    userInfo: {
      displayName: auth.user?.name || 'Instructor'
    },
    configOverwrite: {
      prejoinPageEnabled: false,
    },
    interfaceConfigOverwrite: {
      TOOLBAR_BUTTONS: [
        'microphone', 'camera', 'closedcaptions', 'desktop', 'fullscreen',
        'fodeviceselection', 'hangup', 'profile', 'chat', 'settings', 'videoquality',
        'tileview', 'download', 'help', 'mute-everyone', 'security'
      ],
    }
  }

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

async function terminarLlamada() {
  try {
    await api.post(`/videocalls/${cursoId}/end`)
    alert('Llamada finalizada. Los códigos ya no son válidos.')
    if (jitsiApi) {
      jitsiApi.dispose()
      jitsiApi = null
    }
    router.push('/instructor/capacitaciones')
  } catch (e) {
    console.error('Error al finalizar la llamada', e)
  }
}

onBeforeUnmount(() => {
  if (jitsiApi) jitsiApi.dispose()
})

const handleReadyToClose = () => {
  if (jitsiApi) { jitsiApi.dispose(); jitsiApi = null; }
  router.push('/instructor/capacitaciones')
}

const handleVideoConferenceLeft = () => {
  if (jitsiApi) { jitsiApi.dispose(); jitsiApi = null; }
  router.push('/instructor/capacitaciones')
}

</script>

<template>
  <div class="videocall-container">
    <div class="header-bar">
      <h2>Panel del Instructor - Videollamada</h2>
      <div class="actions">
        <button class="btn btn-outline" @click="router.push('/instructor/capacitaciones')">Salir (sin finalizar)</button>
        <button class="btn btn-danger" @click="terminarLlamada">Finalizar Videollamada para todos</button>
      </div>
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
.actions {
  display: flex;
  gap: 10px;
}
.btn-danger {
  background: #ef4444;
  color: white;
  border: none;
}
.btn-danger:hover {
  background: #dc2626;
}
.jitsi-wrapper {
  flex: 1;
}
</style>
