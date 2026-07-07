package repository

import (
	"context"

	leccionespb "Prueba-Go/gen/lecciones"
)

// ─────────────────────────────────────────────────────────────────────────────
// CursoTreeBuilder construye el árbol completo Módulo → Submódulo → Lección.
// Se separa en su propio archivo para mantener el repositorio principal legible.
// ─────────────────────────────────────────────────────────────────────────────

// BuildCursoTree ensambla el árbol completo de un curso para un usuario dado.
// Si userID es vacío, completada siempre será false (vista de instructor/admin).
func (r *postgresLeccionesRepository) BuildCursoTree(ctx context.Context, cursoID, userID string) (*leccionespb.CursoTreeResponse, error) {
	// 1. Obtener todas las lecciones del curso de una sola vez para evitar N+1.
	var allLecciones []*Leccion
	var err error
	if userID != "" {
		allLecciones, err = r.ListByCursoConProgreso(ctx, cursoID, userID)
	} else {
		allLecciones, err = r.ListByCurso(ctx, cursoID)
	}
	if err != nil {
		return nil, err
	}

	// 2. Indexar lecciones por (moduloID, submoduloID) para lookup O(1).
	//    clave → lista de lecciones
	type leccionKey struct{ moduloID, submoduloID string }
	leccionesByKey := make(map[leccionKey][]*Leccion)
	for _, l := range allLecciones {
		m := ""
		s := ""
		if l.ModuloID != nil {
			m = *l.ModuloID
		}
		if l.SubmoduloID != nil {
			s = *l.SubmoduloID
		}
		k := leccionKey{m, s}
		leccionesByKey[k] = append(leccionesByKey[k], l)
	}

	// 3. Lecciones sueltas (sin módulo ni submódulo).
	sueltasRaw := leccionesByKey[leccionKey{"", ""}]
	sueltas := make([]*leccionespb.LeccionResponse, 0, len(sueltasRaw))
	for _, l := range sueltasRaw {
		sueltas = append(sueltas, l.ToProto())
	}

	// 4. Obtener módulos del curso.
	modulos, err := r.ListModulos(ctx, cursoID)
	if err != nil {
		return nil, err
	}

	modulosProto := make([]*leccionespb.ModuloResponse, 0, len(modulos))
	for _, mod := range modulos {
		modProto := mod.ToProto()

		// 4a. Lecciones directo del módulo (sin submódulo).
		lecsMod := leccionesByKey[leccionKey{mod.ID, ""}]
		for _, l := range lecsMod {
			modProto.Lecciones = append(modProto.Lecciones, l.ToProto())
		}

		// 4b. Obtener submódulos del módulo.
		submodulos, err := r.ListSubmodulos(ctx, mod.ID)
		if err != nil {
			return nil, err
		}

		for _, sub := range submodulos {
			subProto := sub.ToProto()

			// Lecciones del submódulo.
			lecsSub := leccionesByKey[leccionKey{mod.ID, sub.ID}]
			for _, l := range lecsSub {
				subProto.Lecciones = append(subProto.Lecciones, l.ToProto())
			}

			modProto.Submodulos = append(modProto.Submodulos, subProto)
		}

		modulosProto = append(modulosProto, modProto)
	}

	return &leccionespb.CursoTreeResponse{
		CursoId:   cursoID,
		Modulos:   modulosProto,
		Lecciones: sueltas,
	}, nil
}
