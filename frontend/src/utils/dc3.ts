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

/** Estados posibles del panel de constancia de un alumno. */
export type EstadoDC3 =
  | 'cargando'
  | 'lista'
  | 'faltan-mios'
  | 'falta-empresa'
  | 'sin-emitir'

export type EntradaEstadoDC3 = {
  /** Ya respondió `GET /capacitaciones/:id/dc3`. */
  consultado: boolean
  /** URL del documento emitido, vacía si aún no existe. */
  constanciaUrl: string
  trabajadorCompleto: boolean
  empresaCompleta: boolean
  /** El alumno pidió volver al formulario para corregir datos ya guardados. */
  editando: boolean
}

/**
 * Decide qué le toca ver al alumno.
 *
 * Vive fuera del componente porque es donde estuvo el fallo que se reportó: los
 * dos últimos casos estaban fundidos en `falta-empresa`, así que a un alumno
 * con todo en regla se le decía que esperase a su instructor cuando el
 * instructor ya había hecho su parte y lo que había fallado era el armado del
 * documento. Aislada aquí, la distinción queda cubierta por tests.
 */
export function estadoDC3(e: EntradaEstadoDC3): EstadoDC3 {
  if (!e.consultado) return 'cargando'
  // `editando` gana a `lista`: corregir la CURP de una constancia ya emitida es
  // justamente el caso que no tenía salida.
  if (e.constanciaUrl && !e.editando) return 'lista'
  if (!e.trabajadorCompleto || e.editando) return 'faltan-mios'
  if (!e.empresaCompleta) return 'falta-empresa'
  return 'sin-emitir'
}

/** Una entrada del Catálogo Nacional de Ocupaciones. */
export type OcupacionCNO = { clave: string; denominacion: string }

/**
 * Áreas y subáreas ocupacionales del Catálogo Nacional de Ocupaciones.
 *
 * Transcritas del REVERSO de la plantilla oficial. Son 66 y NO son lo mismo que
 * las áreas temáticas: el reverso del formato lleva dos catálogos distintos y
 * cada uno alimenta una casilla distinta del anverso.
 *
 *   Ocupación específica    -> este catálogo (01, 01.1, 05.1, 11.3 ...)
 *   Área temática del curso -> el otro catálogo (1000 ... 9000)
 *
 * La casilla pide la CLAVE, no el nombre del puesto. Mientras fue un campo de
 * texto libre se emitieron constancias con "SUPERVISOR" escrito ahí, que es lo
 * que la persona hace pero no una ocupación del catálogo.
 *
 * Debe coincidir con `Ocupaciones` de pkg/dc3/dc3.go.
 */
export const OCUPACIONES_CNO: OcupacionCNO[] = [
  { clave: '01', denominacion: 'Cultivo, crianza y aprovechamiento' },
  { clave: '01.1', denominacion: 'Agricultura y silvicultura' },
  { clave: '01.2', denominacion: 'Ganadería' },
  { clave: '01.3', denominacion: 'Pesca y acuacultura' },
  { clave: '02', denominacion: 'Extracción y suministro' },
  { clave: '02.1', denominacion: 'Exploración' },
  { clave: '02.2', denominacion: 'Extracción' },
  { clave: '02.3', denominacion: 'Refinación y beneficio' },
  { clave: '02.4', denominacion: 'Provisión de energía' },
  { clave: '02.5', denominacion: 'Provisión de agua' },
  { clave: '03', denominacion: 'Construcción' },
  { clave: '03.1', denominacion: 'Planeación y dirección de obras' },
  { clave: '03.2', denominacion: 'Edificación y urbanización' },
  { clave: '03.3', denominacion: 'Acabado' },
  { clave: '03.4', denominacion: 'Instalación y mantenimiento' },
  { clave: '04', denominacion: 'Tecnología' },
  { clave: '04.1', denominacion: 'Mecánica' },
  { clave: '04.2', denominacion: 'Electricidad' },
  { clave: '04.3', denominacion: 'Electrónica' },
  { clave: '04.4', denominacion: 'Informática' },
  { clave: '04.5', denominacion: 'Telecomunicaciones' },
  { clave: '04.6', denominacion: 'Procesos industriales' },
  { clave: '05', denominacion: 'Procesamiento y fabricación' },
  { clave: '05.1', denominacion: 'Minerales no metálicos' },
  { clave: '05.2', denominacion: 'Metales' },
  { clave: '05.3', denominacion: 'Alimentos y bebidas' },
  { clave: '05.4', denominacion: 'Textiles y prendas de vestir' },
  { clave: '05.5', denominacion: 'Materia orgánica' },
  { clave: '05.6', denominacion: 'Productos químicos' },
  { clave: '05.7', denominacion: 'Productos metálicos y de hule y plástico' },
  { clave: '05.8', denominacion: 'Productos eléctricos y electrónicos' },
  { clave: '05.9', denominacion: 'Productos impresos' },
  { clave: '06', denominacion: 'Transporte' },
  { clave: '06.1', denominacion: 'Ferroviario' },
  { clave: '06.2', denominacion: 'Autotransporte' },
  { clave: '06.3', denominacion: 'Aéreo' },
  { clave: '06.4', denominacion: 'Marítimo y fluvial' },
  { clave: '06.5', denominacion: 'Servicios de apoyo' },
  { clave: '07', denominacion: 'Provisión de bienes y servicios' },
  { clave: '07.1', denominacion: 'Comercio' },
  { clave: '07.2', denominacion: 'Alimentación y hospedaje' },
  { clave: '07.3', denominacion: 'Turismo' },
  { clave: '07.4', denominacion: 'Deporte y esparcimiento' },
  { clave: '07.5', denominacion: 'Servicios personales' },
  { clave: '07.6', denominacion: 'Reparación de artículos de uso doméstico y personal' },
  { clave: '07.7', denominacion: 'Limpieza' },
  { clave: '07.8', denominacion: 'Servicio postal y mensajería' },
  { clave: '08', denominacion: 'Gestión y soporte administrativo' },
  { clave: '08.1', denominacion: 'Bolsa, banca y seguros' },
  { clave: '08.2', denominacion: 'Administración' },
  { clave: '08.3', denominacion: 'Servicios legales' },
  { clave: '09', denominacion: 'Salud y protección social' },
  { clave: '09.1', denominacion: 'Servicios médicos' },
  { clave: '09.2', denominacion: 'Inspección sanitaria y del medio ambiente' },
  { clave: '09.3', denominacion: 'Seguridad social' },
  { clave: '09.4', denominacion: 'Protección de bienes y/o personas' },
  { clave: '10', denominacion: 'Comunicación' },
  { clave: '10.1', denominacion: 'Publicación' },
  { clave: '10.2', denominacion: 'Radio, cine, televisión y teatro' },
  { clave: '10.3', denominacion: 'Interpretación artística' },
  { clave: '10.4', denominacion: 'Traducción e interpretación lingüística' },
  { clave: '10.5', denominacion: 'Publicidad, propaganda y relaciones públicas' },
  { clave: '11', denominacion: 'Desarrollo y extensión del conocimiento' },
  { clave: '11.1', denominacion: 'Investigación' },
  { clave: '11.2', denominacion: 'Enseñanza' },
  { clave: '11.3', denominacion: 'Difusión cultural' },
]

/** Busca una ocupación por su clave. */
export function ocupacionCNO(clave: string): OcupacionCNO | undefined {
  const c = clave.trim()
  return OCUPACIONES_CNO.find(o => o.clave === c)
}
