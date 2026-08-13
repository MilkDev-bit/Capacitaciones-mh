import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useLlamadas, type EventoLlamada } from './useLlamadas'

// api se mockea porque `call_accepted` canjea la sala por un token antes de
// pasar a 'en_llamada'.
vi.mock('../api', () => ({
  default: {
    post: vi.fn(async () => ({
      data: { token: 'jwt-falso', dominio: 'jitsi.test', sala: 'mh-abc' },
    })),
  },
}))

import api from '../api'

function crear() {
  const enviados: Record<string, unknown>[] = []
  const l = useLlamadas((p) => { enviados.push(p) })
  return { l, enviados }
}

const oferta = (over: Partial<EventoLlamada['call']> = {}): EventoLlamada => ({
  type: 'call_offer',
  call: {
    id: 'call-1', emisor_id: 'ana', emisor_name: 'Ana',
    peer_id: 'ana', peer_name: 'Ana', is_group: false, segundos_max: 30,
    ...over,
  } as never,
})

describe('useLlamadas', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.mocked(api.post).mockClear()
  })
  afterEach(() => vi.useRealTimers())

  it('arranca inactiva', () => {
    const { l } = crear()
    expect(l.estado.value).toBe('inactiva')
    expect(l.enCurso.value).toBe(false)
  })

  it('llamar pide al servidor iniciar, sin inventar sala', () => {
    const { l, enviados } = crear()
    l.llamar('beto', 'Beto', false)

    expect(l.estado.value).toBe('saliente')
    expect(enviados).toHaveLength(1)
    expect(enviados[0]).toMatchObject({
      type: 'call_start', peer_id: 'beto', is_group: false,
    })
    // La sala es competencia exclusiva del servidor.
    expect(enviados[0]).not.toHaveProperty('sala')
  })

  it('no permite iniciar una segunda llamada mientras hay una en curso', () => {
    const { l, enviados } = crear()
    l.llamar('beto', 'Beto', false)
    l.llamar('carla', 'Carla', false)
    expect(enviados).toHaveLength(1)
  })

  it('una oferta entrante pasa a estado entrante y arranca la cuenta atrás', async () => {
    const { l } = crear()
    await l.manejarEvento(oferta())

    expect(l.estado.value).toBe('entrante')
    expect(l.nombreOtro.value).toBe('Ana')
    expect(l.restantes.value).toBe(30)

    vi.advanceTimersByTime(3000)
    expect(l.restantes.value).toBe(27)
  })

  it('la oferta entrante no trae la sala: rechazar no deja credencial', async () => {
    const { l } = crear()
    await l.manejarEvento(oferta())
    expect(l.llamada.value?.sala).toBeUndefined()
    expect(l.credenciales.value).toBeNull()

    l.rechazar()
    expect(api.post).not.toHaveBeenCalled()
    expect(l.estado.value).toBe('inactiva')
  })

  it('aceptar avisa al servidor con el id de la llamada', async () => {
    const { l, enviados } = crear()
    await l.manejarEvento(oferta())
    l.aceptar()
    expect(enviados[0]).toMatchObject({ type: 'call_accept', call_id: 'call-1' })
  })

  it('call_accepted canjea la sala por un token antes de abrir Jitsi', async () => {
    const { l } = crear()
    await l.manejarEvento(oferta())
    await l.manejarEvento({
      type: 'call_accepted',
      call: { ...oferta().call!, sala: 'mh-abc' },
    })

    expect(api.post).toHaveBeenCalledWith('/llamadas/token', {
      call_id: 'call-1', sala: 'mh-abc',
    })
    expect(l.estado.value).toBe('en_llamada')
    expect(l.credenciales.value).toEqual({
      token: 'jwt-falso', dominio: 'jitsi.test', sala: 'mh-abc',
    })
  })

  it('si el token es rechazado no se entra a la llamada', async () => {
    vi.mocked(api.post).mockRejectedValueOnce({
      response: { data: { error: 'no participas en esta llamada' } },
    })
    const { l } = crear()
    await l.manejarEvento(oferta())
    await l.manejarEvento({
      type: 'call_accepted',
      call: { ...oferta().call!, sala: 'mh-abc' },
    })

    expect(l.estado.value).toBe('inactiva')
    expect(l.credenciales.value).toBeNull()
    expect(l.aviso.value).toBe('no participas en esta llamada')
  })

  it('rechazo, timeout, ocupado y error devuelven a inactiva con su aviso', async () => {
    const casos: [string, string, Record<string, unknown>?][] = [
      ['call_rejected', 'Llamada rechazada'],
      ['call_timeout', 'Sin respuesta'],
      ['call_timeout', 'No disponible', { motivo: 'no disponible' }],
      ['call_busy', 'Ocupado en otra llamada'],
      ['call_cancelled', 'El otro usuario colgó'],
      ['call_error', 'lo que diga el servidor', { motivo: 'lo que diga el servidor' }],
    ]
    for (const [tipo, esperado, extra] of casos) {
      const { l } = crear()
      l.llamar('beto', 'Beto', false)
      await l.manejarEvento({
        type: tipo,
        call: { ...oferta().call!, ...extra } as never,
      })
      expect(l.estado.value, tipo).toBe('inactiva')
      expect(l.aviso.value, tipo).toBe(esperado)
    }
  })

  it('en grupo, el rechazo de un miembro no corta la llamada saliente', async () => {
    const { l } = crear()
    l.llamar('grupo-1', 'Equipo A', true)
    await l.manejarEvento({
      type: 'call_rejected',
      call: { ...oferta().call!, is_group: true },
    })
    expect(l.estado.value).toBe('saliente')
  })

  it('call_ended con motivo en_curso solo retira el timbre entrante', async () => {
    const { l } = crear()
    await l.manejarEvento(oferta({ is_group: true }))
    expect(l.estado.value).toBe('entrante')

    await l.manejarEvento({
      type: 'call_ended',
      call: { ...oferta().call!, is_group: true, motivo: 'en_curso' },
    })
    expect(l.estado.value).toBe('inactiva')
    // Sin aviso de error: no pasó nada malo, otro contestó primero.
    expect(l.aviso.value).toBe('')
  })

  it('colgar envía cancel si timbra y leave si ya estaba dentro', async () => {
    const { l, enviados } = crear()
    await l.manejarEvento(oferta())
    l.colgar()
    expect(enviados.at(-1)).toMatchObject({ type: 'call_cancel', call_id: 'call-1' })

    const b = crear()
    await b.l.manejarEvento(oferta())
    await b.l.manejarEvento({
      type: 'call_accepted',
      call: { ...oferta().call!, sala: 'mh-abc' },
    })
    b.l.colgar()
    expect(b.enviados.at(-1)).toMatchObject({ type: 'call_leave', call_id: 'call-1' })
  })

  it('una oferta que llega durante otra llamada se ignora', async () => {
    const { l } = crear()
    await l.manejarEvento(oferta())
    await l.manejarEvento(oferta({ id: 'call-2', emisor_name: 'Carla' }))
    expect(l.llamada.value?.id).toBe('call-1')
  })

  it('ignora eventos que no son de llamada', async () => {
    const { l } = crear()
    expect(await l.manejarEvento({ type: 'new_message' })).toBe(false)
    expect(await l.manejarEvento({ type: 'typing' })).toBe(false)
    expect(l.estado.value).toBe('inactiva')
  })
})
