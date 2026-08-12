import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// ── Dobles del entorno ──────────────────────────────────────────────────────
// jsdom no implementa IntersectionObserver ni matchMedia; ambos se sustituyen
// por dobles controlables para poder disparar la intersección a mano.

let observados: Element[] = []
let disparar: ((entries: { isIntersecting: boolean; target: Element }[]) => void) | null = null
let reducirMovimiento = false

class IOFalso {
  constructor(cb: (e: { isIntersecting: boolean; target: Element }[]) => void) {
    disparar = cb
  }
  observe(el: Element) {
    observados.push(el)
  }
  unobserve(el: Element) {
    observados = observados.filter((o) => o !== el)
  }
  disconnect() {
    observados = []
  }
}

beforeEach(() => {
  observados = []
  disparar = null
  reducirMovimiento = false
  vi.stubGlobal('IntersectionObserver', IOFalso)
  vi.stubGlobal(
    'matchMedia',
    vi.fn((q: string) => ({
      matches: q.includes('reduced-motion') ? reducirMovimiento : false,
      media: q,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }))
  )
  // El módulo cachea un observer a nivel de módulo: hay que recargarlo por test.
  vi.resetModules()
})

afterEach(() => vi.unstubAllGlobals())

async function montarDirectiva(valor?: number) {
  const { vReveal } = await import('./useReveal')
  const el = document.createElement('div')
  // @ts-expect-error — se invoca el hook directamente, sin instancia de Vue
  vReveal.mounted(el, { value: valor })
  return { el, vReveal }
}

describe('directiva v-reveal', () => {
  it('deja el elemento oculto hasta que entra en viewport', async () => {
    const { el } = await montarDirectiva()

    expect(el.classList.contains('reveal')).toBe(true)
    expect(el.classList.contains('is-revealed')).toBe(false)
    expect(observados).toContain(el)

    disparar?.([{ isIntersecting: true, target: el }])
    expect(el.classList.contains('is-revealed')).toBe(true)
  })

  it('deja de observar en cuanto aparece', async () => {
    const { el } = await montarDirectiva()
    disparar?.([{ isIntersecting: true, target: el }])
    // Es una animación de entrada: reobservar solo gastaría trabajo por frame.
    expect(observados).not.toContain(el)
  })

  it('ignora las entradas que aún no intersecan', async () => {
    const { el } = await montarDirectiva()
    disparar?.([{ isIntersecting: false, target: el }])
    expect(el.classList.contains('is-revealed')).toBe(false)
  })

  it('escalona con transition-delay', async () => {
    const { el } = await montarDirectiva(2)
    expect(el.style.transitionDelay).toBe('180ms')
  })

  it('no aplica retardo sin argumento', async () => {
    const { el } = await montarDirectiva()
    expect(el.style.transitionDelay).toBe('')
  })

  // Lo más importante de todo el módulo: si la animación no puede correr, el
  // contenido tiene que verse igual. Nunca debe quedar en opacity 0.
  it('muestra el contenido de inmediato con prefers-reduced-motion', async () => {
    reducirMovimiento = true
    const { el } = await montarDirectiva()

    expect(el.classList.contains('is-revealed')).toBe(true)
    expect(el.classList.contains('reveal')).toBe(false)
    expect(observados).not.toContain(el)
  })

  it('muestra el contenido si el navegador no trae IntersectionObserver', async () => {
    vi.stubGlobal('IntersectionObserver', undefined)
    vi.resetModules()
    const { vReveal } = await import('./useReveal')
    const el = document.createElement('div')
    // @ts-expect-error — se invoca el hook directamente, sin instancia de Vue
    vReveal.mounted(el, { value: undefined })

    expect(el.classList.contains('is-revealed')).toBe(true)
  })
})
