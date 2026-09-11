/**
 * Reglas de borrado de mensajes.
 *
 * Vive fuera de la vista porque es la única parte del borrado que decide algo
 * por su cuenta: el resto es pedirle al servidor y pintar lo que responda. Aquí
 * se puede probar sin montar una pantalla de mil líneas.
 */

/**
 * Plazo para retirar un mensaje ya enviado.
 *
 * Tiene que coincidir con `ventanaBorrado` de mensajes-service
 * (services/mensajes/internal/service/borrado.go). Si aquí fuera más largo, el
 * menú ofrecería una opción que el servidor rechaza; si fuera más corto,
 * escondería una que sí funciona.
 */
export const VENTANA_BORRADO_MS = 60 * 60 * 1000

/** Lo mínimo que hace falta saber de un mensaje para decidir. */
export interface MensajeBorrable {
  emisor_id: string
  created_at: string
  eliminado?: boolean
  _status?: 'sending' | 'sent' | 'error'
  _tempId?: string
}

/**
 * ¿Se le puede ofrecer a `usuarioID` el botón de "eliminar para todos"?
 *
 * Es una comprobación de interfaz, no de seguridad: quien autoriza de verdad es
 * el servidor. Sirve para no enseñar un botón que iba a devolver un error.
 *
 * @param ahora inyectable para poder probar el límite sin esperar una hora.
 */
export function puedeBorrarParaTodos(
  msg: MensajeBorrable,
  usuarioID: string | undefined,
  ahora: number = Date.now(),
): boolean {
  if (!usuarioID || msg.emisor_id !== usuarioID) return false
  if (msg.eliminado) return false

  // Un mensaje que aún viaja —o que falló— no tiene id en el servidor, así que
  // no hay nada que borrar allí. `_status` indefinido es un mensaje que vino ya
  // guardado de la API.
  if (msg._status && msg._status !== 'sent') return false

  const enviado = new Date(msg.created_at).getTime()
  // Una fecha ilegible da NaN, y toda comparación con NaN es falsa. Se descarta
  // de forma explícita para que el motivo quede escrito y no parezca casualidad.
  if (Number.isNaN(enviado)) return false

  return ahora - enviado < VENTANA_BORRADO_MS
}
