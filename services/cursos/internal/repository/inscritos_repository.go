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

	// Propiedad del curso. Un instructor no debe poder listar los alumnos de un
	// curso ajeno cambiando el id en la URL: el middleware solo comprueba el
	// ROL, no de quién es el curso.
	var dueno sql.NullString
	err := r.db.GetContext(ctx, &dueno,
		`SELECT instructor_id::text FROM capacitaciones WHERE id = $1 AND deleted_at IS NULL`, cursoID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !dueno.Valid || dueno.String != instructorID {
		return nil, ErrForbidden
	}

	type fila struct {
		UserID     string         `db:"user_id"`
		InscritoAt time.Time      `db:"inscrito_at"`
		LicenciaID sql.NullString `db:"licencia_id"`
	}
	var filas []fila
	err = r.db.SelectContext(ctx, &filas, `
		SELECT user_id::text     AS user_id,
		       inscrito_at       AS inscrito_at,
		       licencia_id::text AS licencia_id
		  FROM inscripciones
		 WHERE capacitacion_id = $1
		 ORDER BY inscrito_at ASC`, cursoID)
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
