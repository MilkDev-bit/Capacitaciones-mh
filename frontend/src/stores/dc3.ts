import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * Estado del modal de la constancia DC-3.
 *
 * Los botones de "Tramitar constancia" viven en cinco vistas distintas
 * (capacitaciones, exámenes, resultado de examen, licencias y el detalle del
 * curso). Antes cada una construía por su cuenta una URL a un formulario
 * externo, con lo que el trámite quedaba fuera de la plataforma y cada copia
 * mandaba datos ligeramente distintos —una enviaba `area_tematica` del curso,
 * las otras `6000` fijo—.
 *
 * Centralizarlo aquí garantiza que los cinco botones abran exactamente el
 * mismo formulario y que añadir un sexto no reintroduzca el enlace externo:
 * lo único que necesita saber quien lo llame es el id de la capacitación.
 */
export const useDC3Store = defineStore('dc3', () => {
  const cursoId = ref('')
  const titulo = ref('')
  const abierto = ref(false)

  /**
   * @param id     id de la capacitación (no del examen ni de la licencia).
   * @param nombre título a mostrar en la cabecera mientras carga el estado.
   */
  function abrir(id: string, nombre = '') {
    if (!id) return
    cursoId.value = id
    titulo.value = nombre
    abierto.value = true
  }

  function cerrar() {
    abierto.value = false
  }

  return { cursoId, titulo, abierto, abrir, cerrar }
})
