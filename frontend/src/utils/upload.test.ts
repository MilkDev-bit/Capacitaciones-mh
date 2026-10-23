import { describe, it, expect } from 'vitest'
import { tamanoLegible } from './upload'

/**
 * El tamaño se le enseña a un instructor mientras espera varios minutos a que
 * suba una clase grabada. Si dice "257.2 MB" y el archivo son 2 GB, lo que
 * parece es que la subida se quedó atascada.
 */
describe('tamanoLegible', () => {
  it('usa bytes por debajo de 1 KB', () => {
    expect(tamanoLegible(0)).toBe('0 B')
    expect(tamanoLegible(999)).toBe('999 B')
  })

  it('usa KB sin decimales', () => {
    expect(tamanoLegible(1024)).toBe('1 KB')
    expect(tamanoLegible(1024 * 500)).toBe('500 KB')
  })

  it('usa MB con un decimal, que es como se mide un vídeo', () => {
    expect(tamanoLegible(1024 * 1024)).toBe('1.0 MB')
    expect(tamanoLegible(257.2 * 1024 * 1024)).toBe('257.2 MB')
  })

  it('pasa a GB al cruzar los 1024 MB', () => {
    expect(tamanoLegible(1024 * 1024 * 1024)).toBe('1.00 GB')
    expect(tamanoLegible(2.5 * 1024 * 1024 * 1024)).toBe('2.50 GB')
  })

  // Los saltos de unidad son donde se cuelan los errores de un orden de
  // magnitud: un byte menos tiene que seguir en la unidad pequeña.
  it('cambia de unidad en el límite exacto', () => {
    expect(tamanoLegible(1023)).toBe('1023 B')
    expect(tamanoLegible(1024 * 1024 - 1)).toBe('1024 KB')
    expect(tamanoLegible(1024 * 1024 * 1024 - 1)).toBe('1024.0 MB')
  })
})
