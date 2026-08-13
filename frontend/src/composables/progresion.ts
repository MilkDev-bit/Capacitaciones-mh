/**
 * Avance secuencial del curso.
 *
 * Regla única: una lección se abre solo cuando todas las anteriores están
 * completadas. Los módulos no necesitan una regla propia — un módulo está
 * bloqueado cuando lo está su primera lección, que es exactamente "no terminaste
 * el módulo anterior". Dos reglas separadas acabarían contradiciéndose.
 *
 * Todo aquí es puro y se prueba sin montar componentes: el orden de las
 * lecciones es lo que decide qué está bloqueado, y equivocarlo deja al alumno
 * encerrado sin poder avanzar.
 */

export interface LeccionPlana {
  id: string
  title?: string
  completada?: boolean
  /** id del módulo que la contiene; vacío si es una lección suelta. */
  moduloID: string
}

export interface ArbolCurso {
  modulos?: {
    id: string
    lecciones?: { id: string; title?: string; completada?: boolean }[]
    submodulos?: { id: string; lecciones?: { id: string; title?: string; completada?: boolean }[] }[]
  }[]
  lecciones?: { id: string; title?: string; completada?: boolean }[]
}

/**
 * Aplana el árbol en el MISMO orden en que lo pinta la barra lateral:
 * módulos (con sus lecciones y luego sus submódulos) y al final las sueltas.
 *
 * El orden importa el doble desde que existe el bloqueo secuencial: antes una
 * discrepancia solo hacía que "Siguiente" saltara a una lección distinta de la
 * que se ve debajo en la lista; ahora además bloquearía las equivocadas.
 */
export function aplanarArbol(tree: ArbolCurso | null | undefined): LeccionPlana[] {
  const salida: LeccionPlana[] = []
  if (!tree) return salida

  for (const mod of tree.modulos ?? []) {
    for (const lec of mod.lecciones ?? []) {
      salida.push({ id: lec.id, title: lec.title, completada: !!lec.completada, moduloID: mod.id })
    }
    for (const sub of mod.submodulos ?? []) {
      for (const lec of sub.lecciones ?? []) {
        salida.push({ id: lec.id, title: lec.title, completada: !!lec.completada, moduloID: mod.id })
      }
    }
  }
  for (const lec of tree.lecciones ?? []) {
    salida.push({ id: lec.id, title: lec.title, completada: !!lec.completada, moduloID: '' })
  }
  return salida
}

/**
 * Índice de la primera lección sin completar: la frontera del avance.
 *
 * Todo lo anterior queda abierto para repasar y esta misma es la que toca
 * ahora. Si están todas completas devuelve `length`, con lo que nada queda
 * bloqueado y el alumno puede revisar el curso entero.
 */
export function frontera(lecciones: readonly LeccionPlana[]): number {
  const i = lecciones.findIndex((l) => !l.completada)
  return i === -1 ? lecciones.length : i
}

/** ¿Se puede abrir la lección que está en esa posición? */
export function desbloqueadaEnIndice(lecciones: readonly LeccionPlana[], indice: number): boolean {
  if (indice < 0 || indice >= lecciones.length) return false
  return indice <= frontera(lecciones)
}

/** ¿Se puede abrir esa lección? Una lección que no existe nunca se abre. */
export function desbloqueada(lecciones: readonly LeccionPlana[], leccionID: string): boolean {
  return desbloqueadaEnIndice(
    lecciones,
    lecciones.findIndex((l) => l.id === leccionID)
  )
}

/**
 * Conjunto de ids abiertos. La barra lateral pinta cientos de lecciones, así
 * que se calcula una vez y se consulta con `has`, en lugar de recorrer la lista
 * entera por cada fila.
 */
export function idsDesbloqueados(lecciones: readonly LeccionPlana[]): Set<string> {
  const limite = frontera(lecciones)
  const set = new Set<string>()
  lecciones.forEach((l, i) => {
    if (i <= limite) set.add(l.id)
  })
  return set
}

/**
 * Módulos abiertos: los que tienen al menos una lección abierta.
 *
 * Un módulo sin lecciones nunca bloquea; encerrar al alumno detrás de un
 * módulo vacío que jamás podrá completar sería un callejón sin salida.
 */
export function idsModulosDesbloqueados(
  tree: ArbolCurso | null | undefined,
  lecciones: readonly LeccionPlana[]
): Set<string> {
  const abiertas = idsDesbloqueados(lecciones)
  const set = new Set<string>()
  for (const mod of tree?.modulos ?? []) {
    const propias = lecciones.filter((l) => l.moduloID === mod.id)
    if (!propias.length || propias.some((l) => abiertas.has(l.id))) set.add(mod.id)
  }
  return set
}

/**
 * Siguiente lección a la que se puede saltar desde la actual, o null.
 *
 * Devuelve null en vez de la lección bloqueada para que el botón "Siguiente"
 * se deshabilite en lugar de llevar a una pantalla que rebota.
 */
export function siguienteAbierta(
  lecciones: readonly LeccionPlana[],
  actualID: string
): LeccionPlana | null {
  const i = lecciones.findIndex((l) => l.id === actualID)
  if (i === -1 || i + 1 >= lecciones.length) return null
  return desbloqueadaEnIndice(lecciones, i + 1) ? lecciones[i + 1]! : null
}

/** Lección anterior. Siempre accesible: lo ya visto se puede repasar. */
export function anterior(
  lecciones: readonly LeccionPlana[],
  actualID: string
): LeccionPlana | null {
  const i = lecciones.findIndex((l) => l.id === actualID)
  return i > 0 ? lecciones[i - 1]! : null
}

/** Título de la lección que hay que terminar para abrir la siguiente. */
export function requisitoDe(
  lecciones: readonly LeccionPlana[],
  leccionID: string
): LeccionPlana | null {
  const i = lecciones.findIndex((l) => l.id === leccionID)
  if (i <= 0) return null
  const f = frontera(lecciones)
  return lecciones[f] ?? null
}
