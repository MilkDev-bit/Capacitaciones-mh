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

  // Cargar del localStorage si existe
  const saved = localStorage.getItem('cart')
  if (saved) {
    try {
      items.value = JSON.parse(saved)
    } catch (e) {
      console.error('Error al parsear el carrito guardado', e)
    }
  }

  function saveToStorage() {
    localStorage.setItem('cart', JSON.stringify(items.value))
  }

  function addItem(newItem: CartItem) {
    items.value.push(newItem)
    saveToStorage()
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
