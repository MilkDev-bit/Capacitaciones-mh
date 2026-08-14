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

/**
 * Catálogo de cursos impartidos, con su área temática y duración oficiales.
 *
 * El nombre que va en la constancia NO es el título comercial de la plataforma:
 * "Curso Trabajos en alturas" se vende así, pero en la DC-3 debe figurar
 * "SEGURIDAD EN TRABAJOS ALTURAS", que es como está registrado. Por eso el
 * editor ofrece este catálogo aparte del título del curso.
 *
 * Transcrito literalmente del proyecto dc3 anterior, ERRATAS INCLUIDAS
 * ("TRABAJOSN", "CONDICIONES SE SEGURIDAD"). No se corrigen a la ligera: si
 * estos nombres coinciden con registros o acreditaciones ya emitidas, cambiar
 * una letra rompe la correspondencia. Conviene revisarlos, pero es una decisión
 * de negocio, no una limpieza de código.
 *
 * Es una sugerencia, no una lista cerrada: el campo admite texto libre para
 * cursos que no estén aquí.
 */
export interface CursoCatalogoDC3 {
  nombre: string
  /** Clave del área temática, ya normalizada a los 4 dígitos. */
  area: string
  /** Duración oficial en horas. */
  horas: number
}

export const CURSOS_CATALOGO_DC3: CursoCatalogoDC3[] = [
  { nombre: 'ENFOQUE DE COMPETENCIAS', area: '3000', horas: 3 },
  { nombre: 'ATENCION DE INSPECCIONES Y AUDITORIAS SEGURIDAD INDUSTRIAL', area: '6000', horas: 8 },
  { nombre: 'SEGURIDAD EN TRABAJOS ALTURAS', area: '6000', horas: 8 },
  { nombre: 'CALCULO DE LA PRIMA DE RIESGO DE TRABAJO', area: '6000', horas: 2 },
  { nombre: 'SEGURIDAD PARA TRABAJOS EN ESPACIOS CONFINADOS', area: '6000', horas: 16 },
  { nombre: 'SISTEMA DE ADMINISTRACION DE SEGURIDAD Y SALUD EN EL TRABAJO', area: '6000', horas: 8 },
  { nombre: 'ANALISIS CAUSA RAIZ DE INVESTIGACION DE ACCIDENTES', area: '6000', horas: 8 },
  { nombre: 'NORMATIVIDAD EN SEGURIDAD INDUSTRIAL', area: '6000', horas: 8 },
  { nombre: 'VERIFICACION DE LAS CONDICIONES SE SEGURIDAD E HIGIENE EN LOS CENTROS DE TRABAJO BAJO EL ESTANDAR CONOCER ECO 391', area: '6000', horas: 8 },
  { nombre: 'SEGURIDAD INDUSTRIAL', area: '6000', horas: 16 },
  { nombre: 'SEGURIDAD PARA TRABAJOSN EN ENERGIA ELECTRICA', area: '6000', horas: 14 },
  { nombre: 'LOTO CANDADEO Y ETIQUETADO', area: '6000', horas: 6 },
  { nombre: 'MANEJO SEGURO DE MONTACARGAS', area: '6000', horas: 6 },
  { nombre: 'PLATAFORMAS DE ELEVACION', area: '6000', horas: 6 },
  { nombre: 'APLICACION DE LA NOM-027-STPS-2008, ACTIVIDADES DE SOLDADURA Y CORTE, CONDICIONES DE SEGURIDAD', area: '6000', horas: 6 },
  { nombre: 'USO Y MANEJO DE HERRAMIENTA MANUAL Y DE PODER', area: '6000', horas: 16 },
  { nombre: 'SISTEMA GLOBALMENTE ARMONIZADO - IDENTIFICACION Y COMUNICACION DE RIESGOS POR SUSTANCIAS QUIMICAS', area: '6000', horas: 8 },
  { nombre: 'SEGURIDAD EN MANIOBRAS E IZAJES', area: '6000', horas: 8 },
  { nombre: 'NOM-031-STPS-2011 CONDICIONES DE SEGURIDAD Y SALUD EN LA CONSTRUCCION', area: '6000', horas: 8 },
  { nombre: 'NOM-002-STPS-2010 CONDICIONES DE SEGURIDAD - PREVENCION Y PROTECCION CONTRA INCENDIOS EN LOS CENTROS DE TRABAJO', area: '6000', horas: 8 },
  { nombre: 'SEGURIDAD EN EXCAVACIONES Y ZANJAS', area: '6000', horas: 8 },
  { nombre: 'MANEJO INTEGRAL DE SUSTANCIAS QUIMICAS', area: '6000', horas: 8 },
  { nombre: 'SELECCION, USO Y MANEJO DEL EQUIPO DE PROTECCION PERSONAL', area: '6000', horas: 8 },
  { nombre: 'ISO 14001-2015 SISTEMA DE GESTION AMBIENTAL', area: '1000', horas: 12 },
]

/** Busca un curso del catálogo por nombre exacto. */
export function cursoDelCatalogo(nombre: string): CursoCatalogoDC3 | undefined {
  const n = nombre.trim().toUpperCase()
  return CURSOS_CATALOGO_DC3.find(c => c.nombre.toUpperCase() === n)
}
