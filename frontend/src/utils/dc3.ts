/**
 * Catálogo de áreas temáticas de los cursos, formato DC-3 (STPS).
 *
 * Está transcrito del REVERSO de la propia plantilla oficial, que es la fuente
 * autorizada: "CLAVES Y DENOMINACIONES DEL CATÁLOGO DE ÁREAS TEMÁTICAS DE LOS
 * CURSOS". Son nueve y no cambian salvo que la STPS publique un formato nuevo.
 *
 * Se ofrece como select y no como texto libre a propósito: la clave va impresa
 * en un documento legal, y un "6000 Seguridad" tecleado como "600" o "Seguridad
 * industrial" produce una constancia inválida que nadie detecta hasta que
 * alguien la presenta en una inspección.
 *
 * Tampoco se deduce automáticamente del título del curso: adivinar el área a
 * partir de "Curso Trabajos en alturas" acierta a menudo y falla en silencio el
 * resto de las veces, que es el peor comportamiento posible para un dato legal.
 * El instructor la elige una vez por curso y queda.
 *
 * Debe coincidir con AreasTematicas de pkg/dc3, que la valida en el backend.
 */
export interface AreaTematica {
  clave: string
  denominacion: string
}

export const AREAS_TEMATICAS_DC3: AreaTematica[] = [
  { clave: '1000', denominacion: 'Producción general' },
  { clave: '2000', denominacion: 'Servicios' },
  { clave: '3000', denominacion: 'Administración, contabilidad y economía' },
  { clave: '4000', denominacion: 'Comercialización' },
  { clave: '5000', denominacion: 'Mantenimiento y reparación' },
  { clave: '6000', denominacion: 'Seguridad' },
  { clave: '7000', denominacion: 'Desarrollo personal y familiar' },
  { clave: '8000', denominacion: 'Uso de tecnologías de la información y comunicación' },
  { clave: '9000', denominacion: 'Participación social' },
]

/** Etiqueta legible de una clave, para mostrarla fuera del select. */
export function etiquetaAreaTematica(clave: string): string {
  const a = AREAS_TEMATICAS_DC3.find(x => x.clave === clave)
  return a ? `${a.clave} — ${a.denominacion}` : clave
}
