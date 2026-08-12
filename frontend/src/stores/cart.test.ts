import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useCartStore, type CartItem } from './cart'

const item = (over: Partial<CartItem> = {}): CartItem => ({
  curso_id: 'c1',
  title: 'Curso de prueba',
  precio: 500,
  cantidad: 1,
  type: 'b2c',
  ...over,
})

beforeEach(() => {
  localStorage.clear()
  setActivePinia(createPinia())
})

describe('useCartStore – edición de cantidades', () => {
  it('actualiza el número de lugares de un ítem', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct', cantidad: 2 }))

    cart.setCantidad(0, 7)

    expect(cart.items[0]!.cantidad).toBe(7)
    expect(cart.totalPrice).toBe(3500)
  })

  it('ignora cantidades inválidas en lugar de romper el total', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct', cantidad: 3 }))

    cart.setCantidad(0, 0)
    cart.setCantidad(0, -5)
    cart.setCantidad(0, NaN)

    expect(cart.items[0]!.cantidad).toBe(3)
  })

  it('limita la cantidad máxima por ítem', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct' }))

    cart.setCantidad(0, 9999)

    expect(cart.items[0]!.cantidad).toBe(500)
  })

  it('persiste el cambio en localStorage', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct' }))

    cart.setCantidad(0, 4)

    expect(JSON.parse(localStorage.getItem('cart')!)[0].cantidad).toBe(4)
  })
})

describe('useCartStore – cambio de tipo de compra', () => {
  it('al pasar a individual fuerza un solo lugar', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct', cantidad: 10 }))

    cart.setTipo(0, 'b2c')

    expect(cart.items[0]!.type).toBe('b2c')
    expect(cart.items[0]!.cantidad).toBe(1)
  })

  it('al pasar a corporativa arranca en 2 lugares', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2c', cantidad: 1 }))

    cart.setTipo(0, 'b2b_direct')

    expect(cart.items[0]!.type).toBe('b2b_direct')
    expect(cart.items[0]!.cantidad).toBe(2)
  })

  it('no altera la cantidad si el tipo ya era el mismo', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct', cantidad: 8 }))

    cart.setTipo(0, 'b2b_direct')

    expect(cart.items[0]!.cantidad).toBe(8)
  })
})

describe('useCartStore – totales', () => {
  it('suma lugares y precios de varios ítems', () => {
    const cart = useCartStore()
    cart.addItem(item({ curso_id: 'c1', precio: 500, cantidad: 1, type: 'b2c' }))
    cart.addItem(item({ curso_id: 'c2', precio: 300, cantidad: 4, type: 'b2b_direct' }))

    expect(cart.totalItems).toBe(5)
    expect(cart.totalPrice).toBe(1700)
  })
})
