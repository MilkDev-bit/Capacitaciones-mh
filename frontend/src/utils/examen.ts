/**
 * Qué se pierde al guardar la edición de un examen.
 *
 * Vive fuera de la vista porque es la única parte del formulario cuya respuesta
 * tiene consecuencias irreversibles: el esquema declara
 *
 *     respuestas_examen.pregunta_id REFERENCES preguntas(id) ON DELETE CASCADE
 *
 * así que quitar una pregunta borra lo que contestaron los estudiantes. En una
 * plataforma que emite constancias DC-3, eso es destruir la evidencia de que
 * alguien aprobó. Que ocurra en silencio es lo que no puede pasar.
 */

/** Pregunta tal como estaba al abrir el formulario. */
export interface PreguntaOriginal {
  id: string
  texto: string
  respuestas: number
}

/** Pregunta tal como está ahora en el formulario. Sin `id` = es nueva. */
export interface PreguntaEditada {
  id?: string
}

/**
 * Preguntas que existían, ya tenían respuestas, y ya no están en el formulario.
 *
 * Solo devuelve las que tienen respuestas: quitar una pregunta que nadie
 * contestó no destruye nada y no merece una confirmación que el instructor
 * acabaría aceptando sin leer.
 */
export function preguntasQuitadas(
  originales: PreguntaOriginal[],
  actuales: PreguntaEditada[],
): PreguntaOriginal[] {
  // Se indexan los ids que siguen presentes. Comparar con `.some()` dentro del
  // filtro sería cuadrático, y un examen puede tener bastantes preguntas.
  const presentes = new Set(
    actuales.map(p => p.id).filter((id): id is string => !!id),
  )
  return originales.filter(o => o.respuestas > 0 && !presentes.has(o.id))
}

/** Total de respuestas de estudiantes que se borrarían. */
export function respuestasQueSePierden(quitadas: PreguntaOriginal[]): number {
  return quitadas.reduce((n, q) => n + q.respuestas, 0)
}

/**
 * Texto de la confirmación. Nombra el número de preguntas Y el de respuestas
 * porque son cosas distintas: quitar una sola pregunta puede borrar treinta
 * respuestas, y ese es el dato que hace dudar.
 */
export function avisoDeBorrado(quitadas: PreguntaOriginal[]): string {
  const nPreguntas = quitadas.length
  const nRespuestas = respuestasQueSePierden(quitadas)
  const preguntas = nPreguntas === 1
    ? '1 pregunta que ya fue respondida'
    : `${nPreguntas} preguntas que ya fueron respondidas`
  const respuestas = nRespuestas === 1
    ? '1 respuesta de estudiantes'
    : `${nRespuestas} respuestas de estudiantes`
  return `Vas a quitar ${preguntas}. Se borrarán ${respuestas} y no se pueden recuperar. ¿Continuar?`
}
