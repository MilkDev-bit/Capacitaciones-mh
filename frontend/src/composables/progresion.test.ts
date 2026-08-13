import { describe, it, expect } from 'vitest'
import {
  aplanarArbol,
  frontera,
  desbloqueada,
  desbloqueadaEnIndice,
  idsDesbloqueados,
  idsModulosDesbloqueados,
  siguienteAbierta,
  anterior,
  requisitoDe,
  type ArbolCurso,
} from './progresion'

const lec = (id: string, completada = false) => ({ id, title: `Lección ${id}`, completada })

/** Módulo con dos lecciones, un submódulo con una, y una lección suelta. */
const arbol: ArbolCurso = {
  modulos: [
    {
      id: 'm1',
      lecciones: [lec('a'), lec('b')],
      submodulos: [{ id: 's1', lecciones: [lec('c')] }],
    },
    { id: 'm2', lecciones: [lec('d')], submodulos: [] },
  ],
  lecciones: [lec('z')],
}

describe('aplanarArbol', () => {
  // El orden es lo que decide qué está bloqueado: si no coincide con lo que
  // pinta la barra lateral, se bloquean lecciones equivocadas.
  it('respeta el orden visual: módulos, luego submódulos, luego sueltas', () => {
    expect(aplanarArbol(arbol).map((l) => l.id)).toEqual(['a', 'b', 'c', 'd', 'z'])
  })

  it('etiqueta cada lección con su módulo y deja vacías las sueltas', () => {
    const p = aplanarArbol(arbol)
    expect(p.find((l) => l.id === 'c')!.moduloID).toBe('m1')
    expect(p.find((l) => l.id === 'd')!.moduloID).toBe('m2')
    expect(p.find((l) => l.id === 'z')!.moduloID).toBe('')
  })

  it('no revienta con un árbol vacío o nulo', () => {
    expect(aplanarArbol(null)).toEqual([])
    expect(aplanarArbol({})).toEqual([])
    expect(aplanarArbol({ modulos: [{ id: 'm', lecciones: [], submodulos: [] }] })).toEqual([])
  })
})

describe('frontera del avance', () => {
  it('es la primera lección sin completar', () => {
    const p = aplanarArbol({
      modulos: [{ id: 'm1', lecciones: [lec('a', true), lec('b', true), lec('c')] }],
    })
    expect(frontera(p)).toBe(2)
  })

  it('con el curso terminado nada queda bloqueado', () => {
    const p = aplanarArbol({
      modulos: [{ id: 'm1', lecciones: [lec('a', true), lec('b', true)] }],
    })
    expect(frontera(p)).toBe(2)
    expect(desbloqueada(p, 'a')).toBe(true)
    expect(desbloqueada(p, 'b')).toBe(true)
  })

  it('en un curso recién empezado solo se abre la primera', () => {
    const p = aplanarArbol(arbol)
    expect(desbloqueada(p, 'a')).toBe(true)
    expect(desbloqueada(p, 'b')).toBe(false)
    expect(desbloqueada(p, 'z')).toBe(false)
  })
})

describe('desbloqueo', () => {
  it('abre lo anterior, la actual, y cierra lo que sigue', () => {
    const p = aplanarArbol({
      modulos: [{ id: 'm1', lecciones: [lec('a', true), lec('b', true), lec('c'), lec('d')] }],
    })
    const abiertas = idsDesbloqueados(p)

    expect([...abiertas].sort()).toEqual(['a', 'b', 'c'])
    expect(abiertas.has('d')).toBe(false)
  })

  it('una lección inexistente nunca se abre', () => {
    expect(desbloqueada(aplanarArbol(arbol), 'fantasma')).toBe(false)
    expect(desbloqueadaEnIndice(aplanarArbol(arbol), -1)).toBe(false)
    expect(desbloqueadaEnIndice(aplanarArbol(arbol), 99)).toBe(false)
  })

  // Si alguien completó una lección de en medio (datos antiguos, un examen
  // que marcó de más), lo anterior debe seguir accesible en vez de encerrar
  // al alumno detrás de un hueco.
  it('tolera lecciones completadas fuera de orden', () => {
    const p = aplanarArbol({
      modulos: [{ id: 'm1', lecciones: [lec('a'), lec('b', true), lec('c')] }],
    })
    expect(desbloqueada(p, 'a')).toBe(true)
    expect(desbloqueada(p, 'b')).toBe(false)
  })
})

describe('módulos', () => {
  it('el segundo módulo queda cerrado hasta terminar el primero', () => {
    const p = aplanarArbol(arbol)
    const abiertos = idsModulosDesbloqueados(arbol, p)

    expect(abiertos.has('m1')).toBe(true)
    expect(abiertos.has('m2')).toBe(false)
  })

  it('se abre cuando todas las lecciones previas están completas', () => {
    const terminado: ArbolCurso = {
      modulos: [
        {
          id: 'm1',
          lecciones: [lec('a', true), lec('b', true)],
          submodulos: [{ id: 's1', lecciones: [lec('c', true)] }],
        },
        { id: 'm2', lecciones: [lec('d')], submodulos: [] },
      ],
      lecciones: [],
    }
    const abiertos = idsModulosDesbloqueados(terminado, aplanarArbol(terminado))
    expect(abiertos.has('m2')).toBe(true)
  })

  // Un módulo vacío no se puede completar nunca: si bloqueara, el curso
  // quedaría muerto a partir de ahí.
  it('un módulo vacío nunca bloquea', () => {
    const conHueco: ArbolCurso = {
      modulos: [
        { id: 'm1', lecciones: [lec('a')], submodulos: [] },
        { id: 'vacio', lecciones: [], submodulos: [] },
      ],
      lecciones: [],
    }
    const abiertos = idsModulosDesbloqueados(conHueco, aplanarArbol(conHueco))
    expect(abiertos.has('vacio')).toBe(true)
  })
})

describe('navegación', () => {
  it('no ofrece siguiente si está bloqueada', () => {
    const p = aplanarArbol(arbol)
    expect(siguienteAbierta(p, 'a')).toBeNull()
  })

  it('ofrece siguiente en cuanto la actual se completa', () => {
    const p = aplanarArbol({
      modulos: [{ id: 'm1', lecciones: [lec('a', true), lec('b')] }],
    })
    expect(siguienteAbierta(p, 'a')?.id).toBe('b')
  })

  it('no hay siguiente al final del curso', () => {
    const p = aplanarArbol({ modulos: [{ id: 'm1', lecciones: [lec('a', true)] }] })
    expect(siguienteAbierta(p, 'a')).toBeNull()
  })

  it('atrás siempre se puede, aunque lo de delante esté cerrado', () => {
    const p = aplanarArbol({
      modulos: [{ id: 'm1', lecciones: [lec('a', true), lec('b')] }],
    })
    expect(anterior(p, 'b')?.id).toBe('a')
    expect(anterior(p, 'a')).toBeNull()
  })

  it('indica qué lección falta para desbloquear', () => {
    const p = aplanarArbol(arbol)
    expect(requisitoDe(p, 'z')?.id).toBe('a')
    // La primera no tiene requisito previo.
    expect(requisitoDe(p, 'a')).toBeNull()
  })
})
