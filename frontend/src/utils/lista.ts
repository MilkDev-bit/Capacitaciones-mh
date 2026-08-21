/**
 * Extrae un arreglo de una respuesta de la API, pase lo que pase.
 *
 * Existe por un fallo concreto. El idioma habitual era:
 *
 *     entregas.value = res.data?.entregas || res.data || []
 *
 * y parece defensivo, pero no lo es: **un objeto vacío es `truthy` en
 * JavaScript**. Cuando el backend responde `{}` —que es lo que devuelve
 * protobuf-go al serializar un repeated vacío, porque el `omitempty` de la
 * etiqueta JSON borra el campo entero— la primera opción es `undefined`, la
 * segunda es `{}` y gana. La variable acaba valiendo `{}` en vez de `[]`, y a
 * la primera llamada a `.filter()` o `.length` la vista revienta y se queda en
 * blanco. El `|| []` del final no llega a ejecutarse nunca.
 *
 * Aquí se comprueba que sea un arreglo de verdad, no que sea "algo".
 */
export function listaDe<T = any>(datos: any, clave?: string): T[] {
  if (clave && Array.isArray(datos?.[clave])) return datos[clave] as T[]
  if (Array.isArray(datos)) return datos as T[]
  return []
}
