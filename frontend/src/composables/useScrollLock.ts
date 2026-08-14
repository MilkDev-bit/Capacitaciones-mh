import { watch, onUnmounted, type Ref } from 'vue'

/**
 * Bloquea el scroll del documento mientras hay una capa superpuesta abierta.
 *
 * Sin esto, con el cajón de navegación abierto en móvil el dedo desplaza la
 * página de detrás: al cerrarlo, el usuario aparece en un punto distinto del
 * que dejó. Es especialmente confuso en listados largos.
 *
 * Se guarda y restaura el valor previo de `overflow` en lugar de asignar
 * cadena vacía: dos capas abiertas a la vez —cajón y modal— se desbloquearían
 * entre ellas al cerrar la primera.
 */
export function useScrollLock(activo: Ref<boolean>) {
  let anterior: string | null = null

  function bloquear() {
    if (anterior !== null) return // ya bloqueado por esta capa
    anterior = document.body.style.overflow
    document.body.style.overflow = 'hidden'
  }

  function liberar() {
    if (anterior === null) return
    document.body.style.overflow = anterior
    anterior = null
  }

  watch(activo, (val) => (val ? bloquear() : liberar()), { immediate: true })

  // Si el componente desaparece con la capa abierta —una navegación desde el
  // propio menú, que es el caso habitual— el scroll debe volver igualmente.
  onUnmounted(liberar)
}
