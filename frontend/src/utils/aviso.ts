/**
 * Aviso de privacidad: versión vigente y consentimiento del visitante anónimo.
 *
 * La VERSIÓN es lo que hace que esto sirva de algo con el tiempo. Un simple
 * booleano «ya aceptó» se queda obsoleto en cuanto se cambia una línea del
 * aviso, y entonces la constancia guardada no dice a qué texto dio su
 * conformidad la persona. Al guardar la versión, actualizar el aviso es subir
 * este número: quien tenga una anterior vuelve a verlo.
 */

/**
 * Fecha de la última revisión del aviso, en formato AAAA-MM-DD.
 *
 * SUBIR ESTE VALOR al cambiar el texto de AvisoPrivacidad.vue. Si no se sube,
 * los usuarios existentes no vuelven a aceptar y quedará registrado que
 * aceptaron una versión que ya no es la que está publicada.
 *
 * El sufijo de letra permite más de una revisión el mismo día. Aquí hizo falta
 * porque el texto se corrigió horas después de publicarse: citaba la LFPDPPP de
 * 2010, que está abrogada, y al INAI, que ya no existe. Quien hubiera aceptado
 * la versión anterior debe volver a ver el aviso, porque lo que aceptó decía
 * otra cosa.
 */
export const AVISO_VERSION = '2026-08-21b'

/** Clave del consentimiento de quien navega SIN cuenta. */
const CLAVE_BANNER = 'aviso_privacidad_banner'

/**
 * ¿El visitante anónimo ya cerró el banner con la versión vigente?
 *
 * Vive en localStorage porque no hay a quién asociarlo: no hay cuenta. Eso
 * significa que NO sirve como prueba —se borra al limpiar el navegador y no
 * viaja entre dispositivos—, y por eso el consentimiento que cuenta es el que
 * se guarda en la base al registrarse. Este banner solo cumple la función de
 * informar antes de que alguien empiece a navegar.
 */
export function bannerAceptado(): boolean {
  try {
    return localStorage.getItem(CLAVE_BANNER) === AVISO_VERSION
  } catch {
    // Modo privado o almacenamiento bloqueado: se muestra el banner cada vez.
    // Es preferible a asumir una aceptación que no se puede comprobar.
    return false
  }
}

export function aceptarBanner(): void {
  try {
    localStorage.setItem(CLAVE_BANNER, AVISO_VERSION)
  } catch {
    // Sin almacenamiento no se puede recordar; el banner reaparecerá. No se
    // interrumpe la navegación por ello.
  }
}

/**
 * ¿Este usuario debe (re)aceptar el aviso?
 *
 * Se compara la versión guardada en su cuenta con la vigente. Vacía significa
 * que nunca aceptó: es el caso de todo el que se registró antes de que esto
 * existiera.
 */
export function usuarioDebeAceptar(versionGuardada: string | null | undefined): boolean {
  return (versionGuardada || '').trim() !== AVISO_VERSION
}
