import { ref, onMounted, onUnmounted, type Directive, type DirectiveBinding, type Ref } from 'vue'

/**
 * v-reveal — aparición al entrar en viewport.
 *
 * Un solo IntersectionObserver compartido por toda la app en lugar de uno por
 * elemento: con 40+ tarjetas en la tienda, crear un observer por nodo satura el
 * hilo principal justo cuando el usuario empieza a hacer scroll.
 *
 * El elemento se deja de observar en cuanto aparece: es una animación de
 * entrada, no un estado que deba revertirse al salir de pantalla.
 *
 * Uso:
 *   <h2 v-reveal>…</h2>          → sin retardo
 *   <p v-reveal="2">…</p>        → 2 pasos de retardo (escalonado)
 */

const PASO_MS = 90

/** El sistema pide menos movimiento: se muestra todo de golpe, sin animar. */
function prefiereMenosMovimiento() {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  )
}

let observer: IntersectionObserver | null = null

function obtenerObserver(): IntersectionObserver | null {
  if (typeof IntersectionObserver === 'undefined') return null
  if (observer) return observer
  observer = new IntersectionObserver(
    (entradas) => {
      for (const e of entradas) {
        if (!e.isIntersecting) continue
        e.target.classList.add('is-revealed')
        observer?.unobserve(e.target)
      }
    },
    // 12% del elemento basta: en móvil una tarjeta alta nunca llegaría al 50%
    // sin que el usuario ya la esté leyendo.
    { threshold: 0.12, rootMargin: '0px 0px -40px 0px' }
  )
  return observer
}

export const vReveal: Directive<HTMLElement, number | undefined> = {
  mounted(el: HTMLElement, binding: DirectiveBinding<number | undefined>) {
    // Sin IntersectionObserver (o con reduced-motion) el contenido debe verse
    // igual: la animación es decoración, nunca un requisito para leer.
    if (prefiereMenosMovimiento()) {
      el.classList.add('is-revealed')
      return
    }

    const pasos = Number(binding.value) || 0
    if (pasos > 0) el.style.transitionDelay = `${pasos * PASO_MS}ms`

    el.classList.add('reveal')

    const obs = obtenerObserver()
    if (!obs) {
      el.classList.add('is-revealed')
      return
    }
    obs.observe(el)
  },
  unmounted(el: HTMLElement) {
    observer?.unobserve(el)
  },
}

/**
 * Devuelve una referencia reactiva al progreso de scroll (0–1) y a la sección
 * visible, calculadas en un solo rAF compartido.
 *
 * Deliberadamente NO toca window.scrollTo: el scroll sigue siendo del
 * navegador. Solo se lee, nunca se sobrescribe, para no romper trackpad,
 * teclado, barra de desplazamiento ni lectores de pantalla.
 */
export function useScrollProgress(idsSecciones: string[] = []) {
  const progreso = ref(0)
  const seccionActiva = ref(idsSecciones[0] ?? '')

  let pendiente = false
  let vivo = true

  function medir() {
    const doc = document.documentElement
    const alturaScroll = Math.max(1, doc.scrollHeight - window.innerHeight)
    progreso.value = Math.min(1, Math.max(0, window.scrollY / alturaScroll))

    if (idsSecciones.length) {
      // Se considera "activa" la última sección cuyo inicio quedó por encima
      // del centro de la pantalla: coincide con lo que el ojo está leyendo.
      const centro = window.scrollY + window.innerHeight * 0.5
      let actual = seccionActiva.value || idsSecciones[0] || ''
      for (const id of idsSecciones) {
        const el = document.getElementById(id)
        if (!el) continue
        if (el.getBoundingClientRect().top + window.scrollY <= centro) actual = id
      }
      seccionActiva.value = actual
    }
    pendiente = false
  }

  function alHacerScroll() {
    if (pendiente || !vivo) return
    pendiente = true
    requestAnimationFrame(medir)
  }

  onMounted(() => {
    vivo = true
    medir()
    window.addEventListener('scroll', alHacerScroll, { passive: true })
    window.addEventListener('resize', alHacerScroll, { passive: true })
  })

  onUnmounted(() => {
    vivo = false
    window.removeEventListener('scroll', alHacerScroll)
    window.removeEventListener('resize', alHacerScroll)
  })

  return { progreso, seccionActiva } as {
    progreso: Ref<number>
    seccionActiva: Ref<string>
  }
}

/**
 * Desplazamiento vertical amortiguado para efectos de parallax.
 *
 * Devuelve el scrollY crudo; cada capa decide su propio factor. Se corta en
 * seco con reduced-motion devolviendo siempre 0, de modo que las capas quedan
 * quietas sin necesidad de condicionales en cada plantilla.
 */
export function useParallax() {
  const desplazamiento = ref(0)
  const activo = !prefiereMenosMovimiento()

  let pendiente = false

  function medir() {
    desplazamiento.value = window.scrollY
    pendiente = false
  }

  function alHacerScroll() {
    if (pendiente) return
    pendiente = true
    requestAnimationFrame(medir)
  }

  onMounted(() => {
    if (!activo) return
    medir()
    window.addEventListener('scroll', alHacerScroll, { passive: true })
  })

  onUnmounted(() => window.removeEventListener('scroll', alHacerScroll))

  return { desplazamiento, activo }
}
