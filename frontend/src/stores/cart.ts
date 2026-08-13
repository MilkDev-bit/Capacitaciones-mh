import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface CartItem {
  curso_id: string
  title: string
  thumbnail?: string
  precio: number
  cantidad: number
  type: 'b2c' | 'b2b_direct'
}

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])

  /**
   * Consolida renglones repetidos del mismo curso y modalidad.
   *
   * Se aplica al leer de localStorage, no solo al agregar: hay carritos ya
   * guardados en los navegadores con la línea duplicada de antes del arreglo.
   * Sin este saneado, esos usuarios seguirían pagando dos veces el mismo curso
   * aunque el código nuevo ya no permita crear el duplicado.
   */
  function consolidar(lista: CartItem[]): CartItem[] {
    const salida: CartItem[] = []
    for (const it of lista) {
      if (!it || typeof it.curso_id !== 'string') continue
      const previo = salida.find((s) => s.curso_id === it.curso_id && s.type === it.type)
      if (!previo) {
        salida.push({ ...it, cantidad: Math.max(1, Math.floor(it.cantidad) || 1) })
        continue
      }
      // Individual: no se acumula, es un acceso por persona.
      if (it.type === 'b2b_direct') {
        previo.cantidad = Math.min(500, previo.cantidad + Math.max(1, it.cantidad))
      }
    }
    return salida
  }

  // Cargar del localStorage si existe
  const saved = localStorage.getItem('cart')
  if (saved) {
    try {
      const crudo = JSON.parse(saved)
      items.value = Array.isArray(crudo) ? consolidar(crudo) : []
    } catch (e) {
      console.error('Error al parsear el carrito guardado', e)
    }
  }

  function saveToStorage() {
    localStorage.setItem('cart', JSON.stringify(items.value))
  }

  /** Índice de la línea que ya representa ese curso en esa modalidad. */
  function indiceDe(cursoID: string, tipo: CartItem['type']) {
    return items.value.findIndex((i) => i.curso_id === cursoID && i.type === tipo)
  }

  /**
   * Agrega un curso al carrito, o lo consolida si ya estaba.
   *
   * Nunca hace push a ciegas. Un `push` sin comprobar permitía dos renglones
   * idénticos del mismo curso: Stripe cobraba dos veces y el backend inscribía
   * una sola —la inscripción es única por usuario y curso—, así que el segundo
   * cobro no compraba nada y acababa en un reembolso manual.
   *
   * Devuelve 'nuevo' | 'existente' | 'sumado' para que la UI pueda decir qué
   * pasó en lugar de fingir que siempre agregó algo.
   */
  function addItem(newItem: CartItem): 'nuevo' | 'existente' | 'sumado' {
    const i = indiceDe(newItem.curso_id, newItem.type)
    if (i === -1) {
      items.value.push({ ...newItem })
      saveToStorage()
      return 'nuevo'
    }

    const actual = items.value[i]!
    // Una inscripción individual es una por persona: la cantidad no significa
    // nada y sumar sería cobrar de más por el mismo acceso.
    if (newItem.type === 'b2c') return 'existente'

    // Las licencias corporativas sí acumulan lugares.
    actual.cantidad = Math.min(500, actual.cantidad + Math.max(1, newItem.cantidad))
    saveToStorage()
    return 'sumado'
  }

  /** ¿Ese curso ya está en el carrito, en cualquier modalidad? */
  function contiene(cursoID: string) {
    return items.value.some((i) => i.curso_id === cursoID)
  }

  function removeItem(index: number) {
    items.value.splice(index, 1)
    saveToStorage()
  }

  /**
   * Ajusta el número de lugares de un ítem corporativo.
   * Antes había que borrar la línea y volver a la tienda para cambiar la
   * cantidad; ahora se edita en el propio carrito.
   */
  function setCantidad(index: number, cantidad: number) {
    const item = items.value[index]
    if (!item) return
    const n = Math.floor(cantidad)
    if (!Number.isFinite(n) || n < 1) return
    item.cantidad = Math.min(n, 500)
    saveToStorage()
  }

  /** Alterna entre inscripción individual (1 lugar) y compra corporativa. */
  function setTipo(index: number, tipo: CartItem['type']) {
    const item = items.value[index]
    if (!item || item.type === tipo) return

    // Cambiar de modalidad puede chocar con una línea que ya existe para ese
    // mismo curso. Sin este paso, pasar la línea b2c de un curso a b2b cuando
    // ya había una b2b del mismo curso dejaría dos renglones idénticos: el
    // mismo duplicado que se evita en addItem, entrando por otra puerta.
    const destino = indiceDe(item.curso_id, tipo)
    if (destino !== -1 && destino !== index) {
      const otro = items.value[destino]!
      if (tipo === 'b2b_direct') {
        otro.cantidad = Math.min(500, otro.cantidad + Math.max(1, item.cantidad))
      }
      items.value.splice(index, 1)
      saveToStorage()
      return
    }

    item.type = tipo
    if (tipo === 'b2c') item.cantidad = 1
    else if (item.cantidad < 2) item.cantidad = 2
    saveToStorage()
  }

  function clearCart() {
    items.value = []
    saveToStorage()
  }

  const totalItems = computed(() => {
    return items.value.reduce((acc, curr) => acc + curr.cantidad, 0)
  })

  const totalPrice = computed(() => {
    return items.value.reduce((acc, curr) => acc + (curr.precio * curr.cantidad), 0)
  })

  // Controlar la visibilidad del drawer globalmente
  const isDrawerOpen = ref(false)
  function openDrawer() {
    isDrawerOpen.value = true
  }
  function closeDrawer() {
    isDrawerOpen.value = false
  }

  return {
    items,
    addItem,
    contiene,
    removeItem,
    setCantidad,
    setTipo,
    clearCart,
    totalItems,
    totalPrice,
    isDrawerOpen,
    openDrawer,
    closeDrawer
  }
})
