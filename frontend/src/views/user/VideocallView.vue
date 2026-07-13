<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
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
const cursoInfo = ref<any>(null)
const showConclusionModal = ref(false)

const dc3FormUrl = computed(() => {
  const title = cursoInfo.value?.title || 'Capacitación en Vivo'
  const horas = Math.max(1, Math.ceil((cursoInfo.value?.duration || 60) / 60))
  return `https://dc3.mhsolucionesempresariales.com/formulario-dc3-8f9d3a2b?nombre_curso=${encodeURIComponent(title)}&duracion_horas=${horas}`
})

onMounted(async () => {
  if (!roomName || !codigo) {
    iziToast.warning({ title: 'Aviso', message: 'Falta la sala o el código de acceso. Ingresa tu código.' })
    router.push('/join')
    return
  }

  try {
    // Validar el código con el backend ANTES de unirse a Jitsi
    await api.post('/videocalls/join', { codigo })
    try {
      const cursoRes = await api.get(`/capacitaciones/${roomName}`)
      cursoInfo.value = cursoRes.data
    } catch {
      try {
        const pubRes = await api.get(`/cursos-publicos/${roomName}`)
        cursoInfo.value = pubRes.data
      } catch {
        cursoInfo.value = { title: 'Capacitación en Vivo', duration: 60 }
      }
    }
  } catch (e: any) {
    iziToast.error({ title: 'Error', message: e.response?.data?.error || 'No tienes acceso a esta videollamada.' })
    router.push('/join')
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
  showConclusionModal.value = true
}

const handleVideoConferenceLeft = () => {
  if (jitsiApi) { jitsiApi.dispose(); jitsiApi = null; }
  showConclusionModal.value = true
}

</script>

<template>
  <div class="videocall-container">
    <div class="header-bar">
      <h2>Videollamada en Curso</h2>
      <button class="btn btn-outline" @click="router.push('/usuario/dashboard')">Salir</button>
    </div>
    <div ref="jitsiContainer" class="jitsi-wrapper"></div>

    <!-- Modal de Conclusión para Estudiante/Participante -->
    <div v-if="showConclusionModal" class="modal-overlay">
      <div class="modal-card">
        <div class="modal-icon">🎓</div>
        <h3>Sesión de Capacitación Concluida</h3>
        <p class="modal-desc">
          ¡Gracias por tu participación! Se ha enviado una notificación al <strong>Comprador / Representante Legal</strong> de tu empresa o grupo con las instrucciones y el acceso prellenado para generar las constancias oficiales DC-3 de esta sesión.
        </p>
        <div v-if="cursoInfo?.dc3_enabled === true" class="dc3-box">
          <p>Si tú eres el responsable o representante legal encargado de tramitar las constancias, puedes ingresar al formulario aquí:</p>
          <a :href="dc3FormUrl" target="_blank" class="btn btn-dc3">
            📋 Trámite y Emisión de Constancias DC-3
          </a>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="router.push('/usuario/dashboard')">Ir a mi Dashboard</button>
        </div>
      </div>
    </div>
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

/* ── Modal Conclusión DC-3 ────────────────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  padding: 1rem;
  backdrop-filter: blur(5px);
}
.modal-card {
  background: #1f2937;
  color: #f9fafb;
  border-radius: 1rem;
  padding: 2.2rem;
  max-width: 500px;
  width: 100%;
  text-align: center;
  border: 1px solid #374151;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
}
.modal-icon {
  font-size: 3rem;
  margin-bottom: 0.8rem;
}
.modal-card h3 {
  font-size: 1.5rem;
  font-weight: 800;
  margin-bottom: 0.8rem;
  color: #fff;
}
.modal-desc {
  font-size: 0.95rem;
  color: #d1d5db;
  line-height: 1.5;
  margin-bottom: 1.5rem;
}
.dc3-box {
  background: rgba(59, 130, 246, 0.1);
  border: 1px dashed #3b82f6;
  border-radius: 0.75rem;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
}
.dc3-box p {
  font-size: 0.88rem;
  color: #93c5fd;
  margin-bottom: 1rem;
}
.btn-dc3 {
  display: inline-block;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff !important;
  font-weight: 700;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  text-decoration: none;
  transition: all 0.2s;
  box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.4);
}
.btn-dc3:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px -2px rgba(37, 99, 235, 0.6);
}
.modal-footer {
  margin-top: 1rem;
}
.btn-primary {
  background: #4f46e5;
  color: #fff;
  border: none;
  padding: 0.7rem 1.4rem;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
}
.btn-primary:hover {
  background: #4338ca;
}
</style>
