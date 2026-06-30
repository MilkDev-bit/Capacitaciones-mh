import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface CartItem {
  curso_id: string
  title: string
  thumbnail?: string
  precio: number
  cantidad: number
  type: 'b2c' | 'b2b_direct'
  schedule_id?: string
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
    // Si ya existe un item idéntico, sumamos cantidad
    const existing = items.value.find(
      (i) => i.curso_id === newItem.curso_id && i.type === newItem.type && i.schedule_id === newItem.schedule_id
    )
    if (existing) {
      existing.cantidad += newItem.cantidad
    } else {
      items.value.push(newItem)
    }
    saveToStorage()
  }

  function removeItem(index: number) {
    items.value.splice(index, 1)
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
    clearCart,
    totalItems,
    totalPrice,
    isDrawerOpen,
    openDrawer,
    closeDrawer
  }
})
