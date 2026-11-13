package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	cursospb "Prueba-Go/gen/cursos"
)

// ─────────────────────────────────────────────────────────────────────────────
// Inscritos a una capacitación
//
// Para el seguimiento del instructor. No devuelve nombres: `users` vive en la
// base de auth y esta consulta corre en la de cursos, así que el gateway los
// resuelve después por gRPC. Es el mismo patrón del panel financiero.
// ─────────────────────────────────────────────────────────────────────────────

type InscritosRepository interface {
	// InstructorListInscritos devuelve los inscritos de un curso.
	//
	// Comprueba la propiedad: si el curso no es de este instructor devuelve
	// ErrForbidden. La verificación va en la misma consulta y no en una previa,
	// para que no exista una ventana entre comprobar y leer.
	InstructorListInscritos(ctx context.Context, cursoID, instructorID string) (*cursospb.ListInscritosResponse, error)
}

func (r *postgresCursosRepository) InstructorListInscritos(ctx context.Context, cursoID, instructorID string) (*cursospb.ListInscritosResponse, error) {
	if cursoID == "" || instructorID == "" {
		return nil, errors.New("curso_id o instructor_id vacíos")
	}

	// Permiso de REVISIÓN, no de propiedad. Un instructor no debe poder listar
	// los alumnos de un curso ajeno cambiando el id en la URL —el middleware
	// solo comprueba el ROL, no de quién es el curso— pero sí debe poder verlos
	// si el administrador lo asignó como revisor.
	//
	// Se comprueba que el curso existe antes que el permiso: así un id
	// inventado da "no existe" y uno real ajeno da "no puedes", que es la
	// distinción que necesita quien depura.
	var existe bool
	err := r.db.GetContext(ctx, &existe,
		`SELECT EXISTS(SELECT 1 FROM capacitaciones WHERE id = $1 AND deleted_at IS NULL)`, cursoID)
	if err != nil {
		return nil, err
	}
	// sql.ErrNoRows y errForbidden, no los errores del paquete `service`: aquí
	// estamos en `repository` y aquellos viven una capa más arriba. Es la misma
	// convención que sigue el resto de este paquete; el servicio los traduce.
	if !existe {
		return nil, sql.ErrNoRows
	}

	puede, err := r.PuedeRevisar(ctx, cursoID, instructorID)
	if err != nil {
		return nil, err
	}
	if !puede {
		return nil, errForbidden
	}

	type fila struct {
		UserID     string         `db:"user_id"`
		InscritoAt time.Time      `db:"inscrito_at"`
		LicenciaID sql.NullString `db:"licencia_id"`
	}
	var filas []fila

	// DOS tablas, no una. Un participante puede llegar por dos caminos y solo
	// mirar `inscripciones` deja fuera a media plataforma:
	//
	//   inscripciones → compró, entró por suscripción o usó un código de
	//                   licencia. Tiene `licencia_id`.
	//   asignaciones  → el instructor o el administrador lo dio de alta a mano
	//                   desde "Estudiantes". Es el camino de las capacitaciones
	//                   internas de empresa, donde nadie compra nada.
	//
	// `InstructorAsignar` escribe en ambas —llama a `Inscribirse`—, pero las
	// altas antiguas y las de administrador pueden existir solo en una. El UNION
	// las junta y deduplica quedándose con la fecha más antigua, que es cuando
	// esa persona entró de verdad al curso.
	err = r.db.SelectContext(ctx, &filas, `
		WITH participantes AS (
		    SELECT user_id, inscrito_at AS entro_at, licencia_id
		      FROM inscripciones
		     WHERE capacitacion_id = $1
		    UNION ALL
		    SELECT user_id, assigned_at AS entro_at, NULL::uuid AS licencia_id
		      FROM asignaciones
		     WHERE capacitacion_id = $1
		)
		SELECT user_id::text          AS user_id,
		       MIN(entro_at)          AS inscrito_at,
		       -- MAX(licencia_id::text) y no MAX(licencia_id): Postgres NO tiene
		       -- un agregado max() para uuid, y la consulta reventaba entera con
		       -- "function max(uuid) does not exist". Sobre texto sí existe, y
		       -- aquí solo se usa para saber si hubo licencia, no para ordenar.
		       MAX(licencia_id::text) AS licencia_id
		  FROM participantes
		 GROUP BY user_id
		 ORDER BY MIN(entro_at) ASC`, cursoID)
	if err != nil {
		return nil, err
	}

	resp := &cursospb.ListInscritosResponse{}
	for _, f := range filas {
		lic := ""
		if f.LicenciaID.Valid {
			lic = f.LicenciaID.String
		}
		resp.Inscritos = append(resp.Inscritos, &cursospb.InscritoInfo{
			UserId:     f.UserID,
			InscritoAt: f.InscritoAt.Format(time.RFC3339),
			LicenciaId: lic,
		})
	}
	return resp, nil
}
