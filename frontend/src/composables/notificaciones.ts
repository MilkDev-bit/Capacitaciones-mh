/**
 * Reglas de presentación de las notificaciones del header.
 *
 * Vive fuera de NotificationBell.vue por una razón práctica: son las dos
 * decisiones del componente que pueden equivocarse en silencio —ocultar un
 * aviso que debía verse y llevar a una ruta que no existe— y ambas se prueban
 * mucho mejor como funciones puras que montando el componente con un router y
 * un mock de axios.
 */

export type Perfil = 'usuario' | 'instructor' | 'admin'

/**
 * Tipos relevantes para cada layout.
 *
 * Todas las notificaciones van dirigidas al usuario, pero no todas pertenecen
 * al sombrero que lleva puesto: un instructor en su panel de docencia no
 * necesita el acuse de la compra que hizo como alumno — eso vive en /usuario.
 *
 * Debe coincidir con las constantes Tipo* de gateway/internal/handler/notificar.go
 * y con `tiposNotificacion` del usuarios-service.
 */
export const TIPOS_POR_PERFIL: Record<Perfil, string[]> = {
  usuario:    ['compra', 'inscripcion', 'mensaje', 'llamada_perdida', 'foro_respuesta'],
  instructor: ['nuevo_alumno', 'foro_respuesta', 'mensaje', 'llamada_perdida'],
  admin:      ['nuevo_alumno', 'compra', 'mensaje', 'llamada_perdida'],
}

export const TODOS_LOS_TIPOS = [...new Set(Object.values(TIPOS_POR_PERFIL).flat())]

/** Prefijos de ruta por perfil, para reescribir el enlace al layout correcto. */
export const BASES_PERFIL = ['/admin', '/instructor', '/usuario'] as const

export function perfilDesdeRol(role: string | undefined | null): Perfil {
  if (role === 'admin') return 'admin'
  if (role === 'instructor') return 'instructor'
  return 'usuario'
}

export function baseDePerfil(perfil: Perfil): string {
  return perfil === 'usuario' ? '/usuario' : `/${perfil}`
}

/**
 * Decide si un tipo se muestra en el layout dado.
 *
 * Un tipo que no aparezca en ninguna lista se muestra SIEMPRE. Es deliberado:
 * si el backend emite un tipo nuevo y nadie actualiza este mapa, el aviso se ve
 * de más —un fallo visible y corregible—. Ocultarlo produciría una notificación
 * que se crea, se guarda y no aparece nunca, que es exactamente el defecto que
 * tenía la campana antes de este cambio.
 */
export function esTipoVisible(tipo: string, perfil: Perfil): boolean {
  if (!TODOS_LOS_TIPOS.includes(tipo)) return true
  return TIPOS_POR_PERFIL[perfil].includes(tipo)
}

/**
 * Traduce el enlace guardado al layout de quien lo abre.
 *
 * El backend emite rutas con prefijo /usuario porque es el caso mayoritario,
 * pero un instructor debe aterrizar en /instructor/mensajes/… y no salir de su
 * panel. Si la ruta reescrita no existe (el panel de admin no tiene mensajes)
 * se cae al enlace original; si ese tampoco resuelve, devuelve null y quien
 * llama debe limitarse a marcar como leída: mejor no navegar que llevar a una
 * pantalla en blanco.
 *
 * `existeRuta` se inyecta para no depender de una instancia de router.
 */
export function resolverEnlace(
  enlace: string,
  perfil: Perfil,
  existeRuta: (path: string) => boolean,
): string | null {
  if (!enlace || !enlace.startsWith('/')) return null

  const base = baseDePerfil(perfil)
  const original = BASES_PERFIL.find(b => enlace === b || enlace.startsWith(`${b}/`))

  const candidatos = original && original !== base
    ? [base + enlace.slice(original.length), enlace]
    : [enlace]

  for (const candidato of candidatos) {
    try {
      if (existeRuta(candidato)) return candidato
    } catch {
      /* ruta no resoluble: se prueba la siguiente */
    }
  }
  return null
}
