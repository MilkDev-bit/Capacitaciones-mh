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

// ─────────────────────────────────────────────────────────────────────────────
// Regresión de un cobro real: se pagaron 2×$400 por el mismo curso y el backend
// inscribió una sola vez, porque la inscripción es única por usuario y curso.
// El segundo renglón no compró nada. Estas pruebas cubren cada puerta por la
// que se colaba el duplicado.
// ─────────────────────────────────────────────────────────────────────────────
describe('useCartStore – nunca dos veces el mismo curso', () => {
  it('no duplica una inscripción individual y avisa que ya estaba', () => {
    const cart = useCartStore()

    expect(cart.addItem(item())).toBe('nuevo')
    expect(cart.addItem(item())).toBe('existente')

    expect(cart.items).toHaveLength(1)
    expect(cart.totalPrice).toBe(500)
  })

  it('suma lugares cuando la compra es corporativa', () => {
    const cart = useCartStore()

    cart.addItem(item({ type: 'b2b_direct', cantidad: 3 }))
    expect(cart.addItem(item({ type: 'b2b_direct', cantidad: 4 }))).toBe('sumado')

    expect(cart.items).toHaveLength(1)
    expect(cart.items[0]!.cantidad).toBe(7)
  })

  // El mismo curso como inscripción propia Y como licencias para el equipo es
  // legítimo: son dos compras distintas, no un duplicado.
  it('permite el mismo curso en las dos modalidades a la vez', () => {
    const cart = useCartStore()

    cart.addItem(item({ type: 'b2c' }))
    cart.addItem(item({ type: 'b2b_direct', cantidad: 5 }))

    expect(cart.items).toHaveLength(2)
  })

  it('contiene() detecta el curso en cualquier modalidad', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct', cantidad: 2 }))

    expect(cart.contiene('c1')).toBe(true)
    expect(cart.contiene('otro')).toBe(false)
  })

  it('cambiar de modalidad no crea un renglón repetido', () => {
    const cart = useCartStore()
    cart.addItem(item({ type: 'b2b_direct', cantidad: 4 }))
    cart.addItem(item({ type: 'b2c' }))

    // Pasar el individual a corporativo choca con el que ya existe.
    cart.setTipo(1, 'b2b_direct')

    expect(cart.items).toHaveLength(1)
    expect(cart.items[0]!.type).toBe('b2b_direct')
    expect(cart.items[0]!.cantidad).toBe(5)
  })

  // Hay carritos ya guardados en los navegadores con el duplicado dentro. El
  // arreglo tiene que sanearlos al leerlos, no solo evitar crear otros nuevos.
  it('consolida un carrito duplicado que venía de localStorage', () => {
    localStorage.setItem(
      'cart',
      JSON.stringify([item(), item(), item({ type: 'b2b_direct', cantidad: 2 })])
    )
    setActivePinia(createPinia())
    const cart = useCartStore()

    expect(cart.items).toHaveLength(2)
    expect(cart.totalPrice).toBe(1500)
  })

  it('sobrevive a un carrito guardado corrupto', () => {
    localStorage.setItem('cart', JSON.stringify([{ nada: true }, null, item()]))
    setActivePinia(createPinia())
    const cart = useCartStore()

    expect(cart.items).toHaveLength(1)
    expect(cart.items[0]!.curso_id).toBe('c1')
  })

  it('no guarda la referencia del objeto recibido', () => {
    const cart = useCartStore()
    const original = item({ type: 'b2b_direct', cantidad: 2 })
    cart.addItem(original)

    cart.setCantidad(0, 9)

    // Mutar el carrito no debe alterar el objeto de quien llamó.
    expect(original.cantidad).toBe(2)
  })
})
