import { describe, it, expect } from 'vitest'
import {
  TIPOS_POR_PERFIL,
  esTipoVisible,
  perfilDesdeRol,
  baseDePerfil,
  resolverEnlace,
} from './notificaciones'

/**
 * Rutas que realmente existen en el router, recortadas a lo que estas pruebas
 * necesitan. Se replican a mano en vez de importar el router porque el objetivo
 * es fijar el comportamiento ante rutas presentes y ausentes, no validar el
 * árbol de rutas: el caso interesante es justamente /admin/mensajes, que NO
 * existe.
 */
const RUTAS = [
  '/usuario/capacitaciones',
  '/usuario/mensajes',
  '/usuario/licencias',
  '/instructor/mensajes',
  '/instructor/estudiantes',
  '/admin/usuarios',
]

const existeRuta = (path: string) =>
  RUTAS.some(r => path === r || path.startsWith(`${r}/`))

describe('perfilDesdeRol', () => {
  it('mapea los roles conocidos', () => {
    expect(perfilDesdeRol('admin')).toBe('admin')
    expect(perfilDesdeRol('instructor')).toBe('instructor')
    expect(perfilDesdeRol('user')).toBe('usuario')
  })

  it('cae a usuario ante un rol ausente o desconocido', () => {
    expect(perfilDesdeRol(undefined)).toBe('usuario')
    expect(perfilDesdeRol(null)).toBe('usuario')
    expect(perfilDesdeRol('superadmin')).toBe('usuario')
  })
})

describe('baseDePerfil', () => {
  it('usa /usuario para el alumno y /rol para el resto', () => {
    expect(baseDePerfil('usuario')).toBe('/usuario')
    expect(baseDePerfil('instructor')).toBe('/instructor')
    expect(baseDePerfil('admin')).toBe('/admin')
  })
})

describe('esTipoVisible', () => {
  it('muestra al alumno su compra pero no las altas de instructor', () => {
    expect(esTipoVisible('compra', 'usuario')).toBe(true)
    expect(esTipoVisible('nuevo_alumno', 'usuario')).toBe(false)
  })

  it('muestra al instructor las altas pero no su compra personal', () => {
    expect(esTipoVisible('nuevo_alumno', 'instructor')).toBe(true)
    expect(esTipoVisible('compra', 'instructor')).toBe(false)
  })

  it('deja pasar un tipo desconocido en todos los perfiles (fail-open)', () => {
    // Si el backend emite un tipo nuevo y nadie actualiza el mapa, es preferible
    // que el aviso se vea de más a que se guarde y no aparezca nunca.
    for (const perfil of ['usuario', 'instructor', 'admin'] as const) {
      expect(esTipoVisible('tipo_que_no_existe_aun', perfil)).toBe(true)
    }
  })

  it('todo tipo listado es visible en al menos un perfil', () => {
    const todos = new Set(Object.values(TIPOS_POR_PERFIL).flat())
    for (const tipo of todos) {
      const visibleEnAlguno = (['usuario', 'instructor', 'admin'] as const)
        .some(p => esTipoVisible(tipo, p))
      expect(visibleEnAlguno, `"${tipo}" quedaría oculto en todos los perfiles`).toBe(true)
    }
  })
})

describe('resolverEnlace', () => {
  it('deja intacto el enlace cuando ya es del perfil correcto', () => {
    expect(resolverEnlace('/usuario/mensajes/abc', 'usuario', existeRuta))
      .toBe('/usuario/mensajes/abc')
  })

  it('reescribe el prefijo al layout de quien abre la notificación', () => {
    // El backend emite /usuario/... pero el instructor no debe salir de su panel.
    expect(resolverEnlace('/usuario/mensajes/abc', 'instructor', existeRuta))
      .toBe('/instructor/mensajes/abc')
  })

  it('cae al enlace original si la ruta reescrita no existe', () => {
    // El panel de admin no tiene sección de mensajes.
    expect(resolverEnlace('/usuario/mensajes/abc', 'admin', existeRuta))
      .toBe('/usuario/mensajes/abc')
  })

  it('devuelve null si ninguna variante resuelve', () => {
    expect(resolverEnlace('/usuario/seccion-borrada', 'instructor', existeRuta)).toBeNull()
  })

  it('rechaza enlaces vacíos o no absolutos', () => {
    expect(resolverEnlace('', 'usuario', existeRuta)).toBeNull()
    expect(resolverEnlace('https://externo.com/x', 'usuario', existeRuta)).toBeNull()
    expect(resolverEnlace('usuario/mensajes', 'usuario', existeRuta)).toBeNull()
  })

  it('no confunde un prefijo parcial con una base de perfil', () => {
    // "/usuarios" no es "/usuario": no debe reescribirse el prefijo.
    expect(resolverEnlace('/admin/usuarios/1', 'admin', existeRuta)).toBe('/admin/usuarios/1')
  })

  it('propaga null si existeRuta lanza en vez de devolver false', () => {
    const revienta = () => { throw new Error('ruta no resoluble') }
    expect(resolverEnlace('/usuario/mensajes/abc', 'usuario', revienta)).toBeNull()
  })
})
