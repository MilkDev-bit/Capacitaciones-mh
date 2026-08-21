/**
 * Formateo de importes que llegan del backend en CENTAVOS.
 *
 * El backend manda enteros a propósito. Un importe en pesos como número de
 * coma flotante pierde centavos al sumarse —0.1 + 0.2 no es 0.3— y el error se
 * acumula justo en la pantalla donde la directiva mira el total del año. La
 * división entre 100 se hace una sola vez, al formatear, y nunca antes de
 * sumar.
 */

/** Divisor de MXN: dos decimales. */
const CENTAVOS_POR_PESO = 100

/**
 * "$1,234.56" a partir de 123456.
 *
 * `decimales: false` da "$1,235" para los titulares grandes, donde los
 * centavos solo estorban.
 */
export function pesos(centavos: number | null | undefined, decimales = true): string {
  const valor = (centavos || 0) / CENTAVOS_POR_PESO
  return new Intl.NumberFormat('es-MX', {
    style: 'currency',
    currency: 'MXN',
    minimumFractionDigits: decimales ? 2 : 0,
    maximumFractionDigits: decimales ? 2 : 0,
  }).format(valor)
}

/**
 * Entero con separador de miles: 15239 → "15,239".
 *
 * Los contadores del panel se pintaban en crudo. Con tres cifras da igual, pero
 * a partir de cuatro deja de leerse de un vistazo, que es lo único que se le
 * pide a una tarjeta de resumen.
 */
export function entero(n: number | null | undefined): string {
  return new Intl.NumberFormat('es-MX', { maximumFractionDigits: 0 }).format(n || 0)
}

/**
 * Porcentaje de una parte sobre un total, acotado entre 0 y 100.
 *
 * Los anillos de los indicadores dibujan este valor, y un porcentaje fuera de
 * rango los deja pintando una circunferencia de más de una vuelta. Un total de
 * cero devuelve cero en vez de dividir por cero: es lo que se ve el primer día,
 * cuando todavía no hay ninguna venta.
 */
export function porcentaje(parte: number, total: number): number {
  if (!total || total <= 0) return 0
  const p = (parte / total) * 100
  if (!Number.isFinite(p)) return 0
  return Math.min(100, Math.max(0, p))
}

/** "ago 2026" a partir de "2026-08". */
export function mesCorto(mes: string): string {
  const [anio, m] = (mes || '').split('-')
  if (!anio || !m) return mes || ''
  // Día 1 y mediodía: construir la fecha a medianoche UTC la corría al mes
  // anterior en México, que va a UTC-6.
  const d = new Date(Number(anio), Number(m) - 1, 1, 12)
  if (Number.isNaN(d.getTime())) return mes
  return d.toLocaleDateString('es-MX', { month: 'short', year: 'numeric' })
}

/** "21 ago 2026, 14:30" a partir de un ISO. */
export function fechaHora(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString('es-MX', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
