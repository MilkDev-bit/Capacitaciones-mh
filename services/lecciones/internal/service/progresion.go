package service

import (
	"errors"
	"fmt"

	leccionespb "Prueba-Go/gen/lecciones"
)

// ─────────────────────────────────────────────────────────────────────────────
// Avance secuencial — validación en el servidor.
//
// El bloqueo del navegador es un guardarraíl para el alumno honesto: se lo
// salta cualquiera con las herramientas de desarrollo llamando directo a
// POST /lecciones/:id/completar. Aquí está la comprobación que sí cuenta,
// porque de esta constancia depende un documento con valor ante la STPS.
//
// El orden se deriva del MISMO árbol que consume el frontend, no de una
// consulta paralela: si cliente y servidor ordenaran por su cuenta, un día
// discreparían y el alumno vería desbloqueada una lección que el servidor
// rechaza, sin ninguna forma de avanzar.
// ─────────────────────────────────────────────────────────────────────────────

// ErrLeccionBloqueada: se intenta completar saltándose una lección previa.
var ErrLeccionBloqueada = errors.New("debes completar la lección anterior")

// AplanarArbol devuelve las lecciones en el orden en que se recorre el curso:
// por cada módulo, primero sus lecciones directas y luego las de sus
// submódulos; al final, las lecciones sueltas.
func AplanarArbol(tree *leccionespb.CursoTreeResponse) []*leccionespb.LeccionResponse {
	if tree == nil {
		return nil
	}
	// Se dimensiona a ojo con las sueltas para evitar el primer par de
	// realocaciones; el árbol típico ronda las decenas de lecciones.
	salida := make([]*leccionespb.LeccionResponse, 0, len(tree.Lecciones)+8)

	for _, mod := range tree.Modulos {
		if mod == nil {
			continue
		}
		salida = append(salida, mod.Lecciones...)
		for _, sub := range mod.Submodulos {
			if sub == nil {
				continue
			}
			salida = append(salida, sub.Lecciones...)
		}
	}
	salida = append(salida, tree.Lecciones...)
	return salida
}

// PrimeraPendiente devuelve el índice de la primera lección sin completar, o
// len(lecciones) si ya están todas. Es la frontera del avance.
func PrimeraPendiente(lecciones []*leccionespb.LeccionResponse) int {
	for i, l := range lecciones {
		if l != nil && !l.Completada {
			return i
		}
	}
	return len(lecciones)
}

// ValidarOrden comprueba que la lección se puede completar ahora.
//
// Devuelve nil si es la que toca o una ya completada —repetir una lección
// terminada es inofensivo y el servicio ya evita otorgar puntos dos veces—.
// Si está más adelante, devuelve ErrLeccionBloqueada nombrando la que falta,
// para que el alumno sepa exactamente a dónde volver.
func ValidarOrden(lecciones []*leccionespb.LeccionResponse, leccionID string) error {
	indice := -1
	for i, l := range lecciones {
		if l != nil && l.Id == leccionID {
			indice = i
			break
		}
	}
	// Fuera del árbol no se opina: puede ser una lección recién creada o un
	// curso a medio migrar, y bloquear por no encontrarla dejaría al alumno
	// encerrado por un problema de datos que no causó él.
	if indice == -1 {
		return nil
	}

	frontera := PrimeraPendiente(lecciones)
	if indice <= frontera {
		return nil
	}

	pendiente := lecciones[frontera]
	if pendiente != nil && pendiente.Title != "" {
		return fmt.Errorf("%w: termina primero \"%s\"", ErrLeccionBloqueada, pendiente.Title)
	}
	return ErrLeccionBloqueada
}
