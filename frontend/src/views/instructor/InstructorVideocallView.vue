<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import api from '../../api'
import iziToast from 'izitoast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const cursoId = route.params.id as string

const jitsiContainer = ref<HTMLElement | null>(null)
let jitsiApi: any = null
const loading = ref(true)
let scheduleId = ''
const cursoInfo = ref<any>(null)
const showConclusionModal = ref(false)

const dc3FormUrl = computed(() => {
  const title = cursoInfo.value?.title || 'Capacitación en Vivo'
  const horas = Math.max(1, Math.ceil((cursoInfo.value?.duration || 60) / 60))
  return `https://dc3.mhsolucionesempresariales.com/formulario-dc3-8f9d3a2b?nombre_curso=${encodeURIComponent(title)}&duracion_horas=${horas}`
})

onMounted(async () => {
  if (!cursoId) {
    iziToast.warning({ title: 'Aviso', message: 'Falta el curso.' })
    router.push('/instructor/capacitaciones')
    return
  }

  let actualRoomName = cursoId
  try {
    // Validar acceso al curso antes de unirse a Jitsi
    const cursoRes = await api.get(`/cursos/${cursoId}`)
    cursoInfo.value = cursoRes.data
    
    // Obtener la sala activa actual (schedule_id) si existe
    const roomRes = await api.get(`/instructor/capacitaciones/${cursoId}/current-room`)
    if (roomRes.data && roomRes.data.room_name) {
      actualRoomName = roomRes.data.room_name
      scheduleId = roomRes.data.schedule_id || ''
    }
  } catch (e: any) {
    iziToast.error({ title: 'Error', message: 'No tienes permisos de instructor para esta sala.' })
    router.push('/instructor/capacitaciones')
    return
  } finally {
    loading.value = false
  }
  
  const domain = 'meet.jit.si'
  const options = {
    roomName: actualRoomName,
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

async function endCall() {
  try {
    if (!confirm('¿Estás seguro de que deseas finalizar la clase? Esto desconectará a todos los estudiantes.')) return
    
    let url = `/instructor/videocall/${cursoId}/end`
    if (scheduleId) {
      url += `?schedule_id=${scheduleId}`
    }
    await api.post(url)
    
    iziToast.success({ title: 'Éxito', message: 'La videollamada ha sido finalizada.' })
    if (jitsiApi) { jitsiApi.dispose(); jitsiApi = null; }
    showConclusionModal.value = true
  } catch (e: any) {
    iziToast.error({ title: 'Error', message: e.response?.data?.error || 'No se pudo finalizar la videollamada.' })
  }
}

onBeforeUnmount(() => {
  if (jitsiApi) jitsiApi.dispose()
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
      <h2>Panel del Instructor - Videollamada</h2>
      <div class="actions">
        <button class="btn btn-outline" @click="router.push('/instructor/capacitaciones')">Salir (sin finalizar)</button>
        <button class="btn btn-danger" @click="endCall">Finalizar Videollamada para todos</button>
      </div>
    </div>
    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>Validando sala y conectando...</p>
    </div>
    <div ref="jitsiContainer" class="jitsi-wrapper"></div>

    <!-- Modal de Conclusión y DC-3 -->
    <div v-if="showConclusionModal" class="modal-overlay">
      <div class="modal-card">
        <div class="modal-icon">✅</div>
        <h3>Videollamada Finalizada con Éxito</h3>
        <p class="modal-desc">
          Se ha notificado automáticamente por correo electrónico al <strong>Representante Legal / Comprador</strong> de las licencias con las instrucciones y el enlace prellenado para la emisión de constancias DC-3.
        </p>
        <div class="dc3-box">
          <p>¿Eres también el representante legal o deseas verificar el formulario para la generación de constancias?</p>
          <a :href="dc3FormUrl" target="_blank" class="btn btn-dc3">
            📋 Abrir Trámite y Emisión DC-3
          </a>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="router.push('/instructor/capacitaciones')">Volver a Mis Capacitaciones</button>
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
.loading-overlay {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #111;
}
.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 255, 255, 0.1);
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}
@keyframes spin {
  to { transform: rotate(360deg); }
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
