import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SeguimientoAlumnos from './SeguimientoAlumnos.vue'

vi.mock('../api', () => ({ default: { get: vi.fn() } }))
vi.mock('../utils/toast', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

import api from '../api'
const get = api.get as unknown as ReturnType<typeof vi.fn>

const ALUMNOS = {
  inscritos: [
    {
      user_id: 'u1', nombre: 'Ana López', email: 'ana@ejemplo.mx', avatar_url: '',
      inscrito_at: '2026-08-01T10:00:00Z', por_licencia: false,
      completadas: 3, total_lecciones: 5, ultima_actividad: '2026-08-20T10:00:00Z', puntos: 120,
    },
    {
      // Inscrito que nunca empezó: la razón de ser de la pantalla.
      user_id: 'u2', nombre: 'Beto Ruiz', email: 'beto@ejemplo.mx', avatar_url: '',
      inscrito_at: '2026-08-02T10:00:00Z', por_licencia: true,
      completadas: 0, total_lecciones: 5, ultima_actividad: '', puntos: 0,
    },
  ],
  total_lecciones: 5,
}

function montar(cursoId: string | null = 'c1') {
  return mount(SeguimientoAlumnos, {
    props: { cursoId, cursoTitulo: 'Trabajos en alturas' },
    global: { stubs: { Teleport: true } },
  })
}

describe('SeguimientoAlumnos', () => {
  beforeEach(() => vi.clearAllMocks())

  it('no consulta nada mientras no haya curso', async () => {
    montar(null)
    await flushPromises()
    expect(get).not.toHaveBeenCalled()
  })

  it('lista a los inscritos con su avance', async () => {
    get.mockResolvedValue({ data: ALUMNOS })
    const w = montar()
    await flushPromises()

    expect(w.text()).toContain('Ana López')
    expect(w.text()).toContain('3 / 5 lecciones')
    expect(w.text()).toContain('120 pts')
  })

  /**
   * Quien se inscribió y no ha abierto una sola lección es justo a quien hay
   * que perseguir: si se filtrara por "tiene progreso" sería invisible.
   */
  it('incluye a quien no ha empezado', async () => {
    get.mockResolvedValue({ data: ALUMNOS })
    const w = montar()
    await flushPromises()

    expect(w.text()).toContain('Beto Ruiz')
    expect(w.text()).toContain('0 / 5 lecciones')
  })

  it('cuenta cuántos completaron el curso', async () => {
    get.mockResolvedValue({
      data: {
        inscritos: [
          { ...ALUMNOS.inscritos[0], completadas: 5 },
          ALUMNOS.inscritos[1],
        ],
        total_lecciones: 5,
      },
    })
    const w = montar()
    await flushPromises()
    expect(w.text()).toContain('1')
    expect(w.text()).toContain('completaron')
  })

  /**
   * Un curso sin lecciones no acredita nada. Contar 0/0 como completado haría
   * que la pantalla dijera que todo el mundo terminó un curso vacío.
   */
  it('un curso sin lecciones no cuenta a nadie como completado', async () => {
    get.mockResolvedValue({
      data: {
        inscritos: [{ ...ALUMNOS.inscritos[0], completadas: 0, total_lecciones: 0 }],
        total_lecciones: 0,
      },
    })
    const w = montar()
    await flushPromises()

    const resumen = w.find('.sa-resumen').text()
    expect(resumen).toMatch(/0\s*completaron/)
  })

  it('filtra por nombre y por correo', async () => {
    get.mockResolvedValue({ data: ALUMNOS })
    const w = montar()
    await flushPromises()

    await w.find('.sa-buscar').setValue('beto@')
    expect(w.text()).toContain('Beto Ruiz')
    expect(w.text()).not.toContain('Ana López')
  })

  it('avisa cuando no hay nadie inscrito', async () => {
    get.mockResolvedValue({ data: { inscritos: [], total_lecciones: 5 } })
    const w = montar()
    await flushPromises()
    expect(w.text()).toContain('Todavía no hay nadie inscrito')
  })

  // Misma trampa que tumbó la pantalla de Entregas: el backend puede responder
  // `{}` cuando la lista viene vacía.
  it('sobrevive a una respuesta sin la clave esperada', async () => {
    get.mockResolvedValue({ data: {} })
    const w = montar()
    await flushPromises()
    expect(w.text()).toContain('Todavía no hay nadie inscrito')
  })

  it('abre el detalle del alumno al pulsarlo', async () => {
    get.mockImplementation((url: string) =>
      url.includes('/u1')
        ? Promise.resolve({
            data: {
              user_id: 'u1', nombre: 'Ana López', email: 'ana@ejemplo.mx',
              completado: false, completadas: 3, total_lecciones: 5,
              ultima_actividad: '2026-08-20T10:00:00Z', puntos: 120,
              lecciones: [
                { id: 'l1', title: 'Introducción', tipo: 'VIDEO', orden: 1, duracion_min: 10, completada: true },
                { id: 'l2', title: 'Arneses', tipo: 'DOCUMENT', orden: 2, duracion_min: 15, completada: false },
              ],
              examenes: [
                { id: 'e1', title: 'Evaluación final', presentado: true, puntaje: 8, porcentaje: 80, submitted_at: '2026-08-20T11:00:00Z' },
              ],
            },
          })
        : Promise.resolve({ data: ALUMNOS })
    )
    const w = montar()
    await flushPromises()

    await w.findAll('.sa-fila')[0].trigger('click')
    await flushPromises()

    const texto = w.text()
    expect(texto).toContain('Introducción')
    expect(texto).toContain('Pendiente')
    expect(texto).toContain('Evaluación final')
    expect(texto).toContain('80%')
  })
})
