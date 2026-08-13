import { computed, ref } from 'vue'
import api from '../api'

/**
 * Máquina de estados de la videollamada en el cliente.
 *
 * El navegador NO decide nada aquí: no elige el nombre de sala, no sabe quién
 * puede entrar y no abre Jitsi por su cuenta. Solo declara intención
 * ("llamar", "acepto", "cuelgo") y reacciona a lo que el servidor responde.
 * Toda la autoridad vive en el gateway, que es el único punto donde se puede
 * comprobar la identidad sin confiar en el cliente.
 *
 * Flujo:
 *
 *   inactiva ──llamar()──► saliente ──call_accepted──► en_llamada
 *      ▲                      │
 *      │                      └──call_rejected/timeout/busy──┐
 *      │                                                     │
 *      └──colgar()───────────────────────────────────────────┘
 *
 *   inactiva ──call_offer──► entrante ──aceptar()──► en_llamada
 */

export type EstadoLlamada = 'inactiva' | 'saliente' | 'entrante' | 'en_llamada'

export interface DatosLlamada {
  id: string
  sala?: string
  emisor_id: string
  emisor_name: string
  peer_id: string
  peer_name?: string
  is_group: boolean
  motivo?: string
  segundos_max?: number
}

/** Evento de señalización tal como llega por el WebSocket. */
export interface EventoLlamada {
  type: string
  call?: DatosLlamada
}

export interface CredencialesJitsi {
  token: string
  dominio: string
  sala: string
}

export function useLlamadas(enviar: (payload: Record<string, unknown>) => void) {
  const estado = ref<EstadoLlamada>('inactiva')
  const llamada = ref<DatosLlamada | null>(null)
  const credenciales = ref<CredencialesJitsi | null>(null)
  const aviso = ref('')
  /** Segundos restantes de timbre, para la cuenta atrás de la UI. */
  const restantes = ref(0)

  let temporizador: ReturnType<typeof setInterval> | undefined

  const enCurso = computed(() => estado.value !== 'inactiva')
  const nombreOtro = computed(() => {
    if (!llamada.value) return ''
    // En una llamada entrante el "otro" es quien llama; en una saliente, el
    // destinatario de la conversación.
    return estado.value === 'entrante'
      ? llamada.value.emisor_name
      : (llamada.value.peer_name || llamada.value.emisor_name)
  })

  function iniciarCuenta(segundos: number) {
    detenerCuenta()
    restantes.value = segundos
    temporizador = setInterval(() => {
      restantes.value -= 1
      if (restantes.value <= 0) detenerCuenta()
    }, 1000)
  }

  function detenerCuenta() {
    if (temporizador) clearInterval(temporizador)
    temporizador = undefined
  }

  function reiniciar(motivo = '') {
    detenerCuenta()
    estado.value = 'inactiva'
    llamada.value = null
    credenciales.value = null
    restantes.value = 0
    aviso.value = motivo
    if (motivo) setTimeout(() => { aviso.value = '' }, 5000)
  }

  /** Pide al servidor iniciar una llamada. La sala la elige él, no nosotros. */
  function llamar(peerID: string, peerName: string, isGroup: boolean) {
    if (enCurso.value) return
    estado.value = 'saliente'
    llamada.value = {
      id: '', emisor_id: '', emisor_name: '',
      peer_id: peerID, peer_name: peerName, is_group: isGroup,
    }
    enviar({ type: 'call_start', peer_id: peerID, peer_name: peerName, is_group: isGroup })
  }

  function aceptar() {
    if (estado.value !== 'entrante' || !llamada.value) return
    enviar({ type: 'call_accept', call_id: llamada.value.id })
  }

  function rechazar() {
    if (estado.value !== 'entrante' || !llamada.value) return
    enviar({ type: 'call_reject', call_id: llamada.value.id })
    reiniciar()
  }

  /** Colgar: cancela si aún timbra, o sale si ya estaba dentro. */
  function colgar() {
    if (!llamada.value) { reiniciar(); return }
    const tipo = estado.value === 'en_llamada' ? 'call_leave' : 'call_cancel'
    enviar({ type: tipo, call_id: llamada.value.id })
    reiniciar()
  }

  /**
   * Canjea la sala por un token de Jitsi.
   *
   * Se hace en este momento y no antes porque el token es la credencial de
   * entrada: pedirlo al recibir la oferta significaría que quien rechaza la
   * llamada se queda igualmente con una llave válida de la sala.
   */
  async function pedirCredenciales(datos: DatosLlamada): Promise<boolean> {
    if (!datos.sala) return false
    try {
      const res = await api.post('/llamadas/token', { call_id: datos.id, sala: datos.sala })
      credenciales.value = {
        token: res.data.token,
        dominio: res.data.dominio,
        sala: res.data.sala,
      }
      return true
    } catch (e: any) {
      reiniciar(e.response?.data?.error || 'No se pudo conectar la llamada')
      return false
    }
  }

  /** Procesa un evento de señalización. Devuelve true si lo consumió. */
  async function manejarEvento(ev: EventoLlamada): Promise<boolean> {
    if (!ev.type.startsWith('call_')) return false
    const c = ev.call

    switch (ev.type) {
      case 'call_offer':
        // Una llamada entrante mientras hay otra en curso no interrumpe: el
        // servidor ya respondió "ocupado" al otro lado.
        if (enCurso.value || !c) break
        llamada.value = c
        estado.value = 'entrante'
        iniciarCuenta(c.segundos_max ?? 30)
        break

      case 'call_ringing':
        if (!c) break
        llamada.value = c
        estado.value = 'saliente'
        iniciarCuenta(c.segundos_max ?? 30)
        break

      case 'call_accepted': {
        if (!c) break
        llamada.value = c
        detenerCuenta()
        const ok = await pedirCredenciales(c)
        if (ok) estado.value = 'en_llamada'
        break
      }

      case 'call_rejected':
        // En grupo, que un miembro cuelgue no termina la llamada del emisor.
        if (c?.is_group && estado.value === 'saliente') break
        reiniciar('Llamada rechazada')
        break

      case 'call_cancelled':
        reiniciar('El otro usuario colgó')
        break

      case 'call_timeout':
        reiniciar(c?.motivo === 'no disponible' ? 'No disponible' : 'Sin respuesta')
        break

      case 'call_busy':
        reiniciar('Ocupado en otra llamada')
        break

      case 'call_ended':
        // motivo "en_curso" solo retira el timbre en un grupo: la llamada
        // sigue viva y el usuario puede unirse desde el hilo.
        if (c?.motivo === 'en_curso') {
          if (estado.value === 'entrante') reiniciar('')
          break
        }
        reiniciar('Llamada finalizada')
        break

      case 'call_error':
        reiniciar(c?.motivo || 'No se pudo iniciar la llamada')
        break
    }
    return true
  }

  function limpiar() {
    detenerCuenta()
  }

  return {
    estado, llamada, credenciales, aviso, restantes,
    enCurso, nombreOtro,
    llamar, aceptar, rechazar, colgar, manejarEvento, limpiar,
  }
}
