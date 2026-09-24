import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useWSStore = defineStore('ws', () => {
  const socket = ref<WebSocket | null>(null)
  const conectado = ref(false)
  
  // Lista de callbacks para escuchar mensajes entrantes
  const oyentes = ref<((data: any) => void)[]>([])

  function conectar() {
    if (socket.value && (socket.value.readyState === WebSocket.OPEN || socket.value.readyState === WebSocket.CONNECTING)) {
      return
    }

    // Ajusta la URL del WS a la ruta real de tu Gateway
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/ws`

    socket.value = new WebSocket(wsUrl)

    socket.value.onopen = () => {
      conectado.value = true
    }

    socket.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        // Notificar a todos los suscriptores (como useLlamadas)
        oyentes.value.forEach((cb) => cb(data))
      } catch (e) {
        console.error('Error al procesar mensaje WS:', e)
      }
    }

    socket.value.onclose = () => {
      conectado.value = false
      socket.value = null
      // Reintentar conexión automáticamente tras 3 segundos
      setTimeout(() => {
        conectar()
      }, 3000)
    }

    socket.value.onerror = (err) => {
      console.error('Error en WebSocket:', err)
      socket.value?.close()
    }
  }

  function enviar(payload: Record<string, unknown>) {
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      socket.value.send(JSON.stringify(payload))
    } else {
      console.warn('WebSocket no conectado. No se pudo enviar:', payload)
    }
  }

  function onMensaje(callback: (data: any) => void) {
    oyentes.value.push(callback)
  }

  return {
    socket,
    conectado,
    conectar,
    enviar,
    onMensaje
  }
})