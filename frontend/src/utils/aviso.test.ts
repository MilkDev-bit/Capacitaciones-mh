import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { AVISO_VERSION, bannerAceptado, aceptarBanner, usuarioDebeAceptar } from './aviso'

describe('usuarioDebeAceptar', () => {
  it('exige aceptar cuando nunca aceptó', () => {
    expect(usuarioDebeAceptar('')).toBe(true)
    expect(usuarioDebeAceptar(null)).toBe(true)
    expect(usuarioDebeAceptar(undefined)).toBe(true)
  })

  it('no exige nada si ya aceptó la versión vigente', () => {
    expect(usuarioDebeAceptar(AVISO_VERSION)).toBe(false)
    expect(usuarioDebeAceptar(` ${AVISO_VERSION} `)).toBe(false)
  })

  // La razón de ser de guardar la versión y no un booleano: al cambiar el texto
  // del aviso, quien aceptó el anterior tiene que volver a verlo.
  it('vuelve a exigir aceptación cuando el aviso cambia de versión', () => {
    expect(usuarioDebeAceptar('2020-01-01')).toBe(true)
  })
})

describe('banner del visitante anónimo', () => {
  beforeEach(() => localStorage.clear())
  afterEach(() => vi.restoreAllMocks())

  it('arranca sin aceptar y recuerda tras aceptar', () => {
    expect(bannerAceptado()).toBe(false)
    aceptarBanner()
    expect(bannerAceptado()).toBe(true)
  })

  it('ignora una aceptación de una versión anterior', () => {
    localStorage.setItem('aviso_privacidad_banner', '2020-01-01')
    expect(bannerAceptado()).toBe(false)
  })

  // Modo privado / almacenamiento bloqueado: no debe romper la carga de la app
  // ni dar por buena una aceptación que no se puede comprobar.
  it('no revienta si localStorage lanza', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('bloqueado')
    })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('bloqueado')
    })
    expect(() => aceptarBanner()).not.toThrow()
    expect(bannerAceptado()).toBe(false)
  })
})
