import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import { useSuscripcionStore, precioDesdeCentavos } from './suscripcion'
import api from '../api'

const planIndividual = {
  id: 'p1',
  codigo: 'individual-mensual',
  nombre: 'Acceso total',
  modalidad: 'individual',
  intervalo: 'mes',
  precio_centavos: 49900,
  moneda: 'MXN',
  activo: true,
}
const planEquipo = {
  id: 'p2',
  codigo: 'equipo-mensual',
  nombre: 'Equipos',
  modalidad: 'asientos',
  intervalo: 'mes',
  precio_centavos: 39900,
  moneda: 'MXN',
  activo: true,
}

describe('store de suscripción', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('separa los planes por modalidad', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({ data: [planIndividual, planEquipo] } as never)
    const s = useSuscripcionStore()
    await s.cargarPlanes()

    expect(s.planesIndividuales).toHaveLength(1)
    expect(s.planesEquipo).toHaveLength(1)
    expect(s.planesIndividuales[0]?.codigo).toBe('individual-mensual')
  })

  it('no repite la consulta de planes si ya los tiene', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: [planIndividual] } as never)
    const s = useSuscripcionStore()
    await s.cargarPlanes()
    await s.cargarPlanes()
    expect(api.get).toHaveBeenCalledTimes(1)
  })

  it('trata un id vacío como "no tiene suscripción"', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({ data: {} } as never)
    const s = useSuscripcionStore()
    await s.cargarMia()

    expect(s.mia).toBeNull()
    expect(s.tieneSuscripcion).toBe(false)
    expect(s.accesoVigente).toBe(false)
  })

  it('un 401 en vista pública no rompe el store', async () => {
    vi.mocked(api.get).mockRejectedValueOnce({ response: { status: 401 } })
    const s = useSuscripcionStore()
    await s.cargarMia()

    expect(s.mia).toBeNull()
    expect(s.accesoVigente).toBe(false)
  })

  // El periodo de gracia es la regla de negocio que más fácil se rompe: un
  // cobro fallido NO debe cerrarle el catálogo al usuario mientras Stripe
  // reintenta. El backend lo decide y el front solo lo refleja.
  it('conserva el acceso durante el periodo de gracia', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: { id: 's1', estado: 'vencida', acceso_vigente: true, plan_nombre: 'Acceso total' },
    } as never)
    const s = useSuscripcionStore()
    await s.cargarMia()

    expect(s.accesoVigente).toBe(true)
    expect(s.enPeriodoDeGracia).toBe(true)
  })

  it('una suscripción cancelada deja de dar acceso', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: { id: 's1', estado: 'cancelada', acceso_vigente: false },
    } as never)
    const s = useSuscripcionStore()
    await s.cargarMia()

    expect(s.tieneSuscripcion).toBe(true)
    expect(s.accesoVigente).toBe(false)
  })

  it('limpiar borra la membresía y permite volver a consultarla', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { id: 's1', acceso_vigente: true } } as never)
    const s = useSuscripcionStore()
    await s.cargarMia()
    expect(s.accesoVigente).toBe(true)

    s.limpiar()
    expect(s.mia).toBeNull()

    await s.cargarMia()
    expect(api.get).toHaveBeenCalledTimes(2)
  })

  it('falla el checkout si Stripe no devuelve URL', async () => {
    vi.mocked(api.post).mockResolvedValueOnce({ data: {} } as never)
    const s = useSuscripcionStore()
    await expect(s.checkout('individual-mensual')).rejects.toThrow()
  })
})

describe('precioDesdeCentavos', () => {
  it('omite decimales cuando el importe es redondo', () => {
    expect(precioDesdeCentavos(49900)).toBe('$499')
  })

  it('los conserva cuando no lo es', () => {
    expect(precioDesdeCentavos(49950)).toBe('$499.50')
  })

  // El bug que motivó pkg/money: dividir entre 100 nunca debe truncar.
  it('no pierde centavos al convertir', () => {
    expect(precioDesdeCentavos(820)).toBe('$8.20')
  })
})
