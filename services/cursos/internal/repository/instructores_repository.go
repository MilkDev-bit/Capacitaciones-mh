package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Instructores asignados a un curso
//
// Hay DOS relaciones entre un instructor y un curso, y mezclarlas sería el
// error que este archivo existe para evitar:
//
//	capacitaciones.instructor_id → el DUEÑO. Lo creó. Puede editarlo y borrarlo.
//	curso_instructores           → REVISORES. Ven el avance de los alumnos y los
//	                               resultados de los exámenes. Nada más.
//
// El administrador suma revisores para que alguien dé seguimiento sin entregarle
// la capacidad de destruir el material. Si un día hiciera falta que un revisor
// edite, eso es una columna nueva en esta tabla, no quitarle el candado a las
// funciones de abajo.
// ─────────────────────────────────────────────────────────────────────────────

type InstructorAsignadoRow struct {
	InstructorID string    `db:"instructor_id"`
	AsignadoAt   time.Time `db:"asignado_at"`
}

type InstructoresRepository interface {
	// AsignarInstructor da acceso de revisión. Es idempotente.
	AsignarInstructor(ctx context.Context, cursoID, instructorID, asignadoPor string) error
	// QuitarInstructor retira el acceso de revisión.
	QuitarInstructor(ctx context.Context, cursoID, instructorID string) error
	// ListInstructoresDeCurso devuelve los revisores de un curso.
	ListInstructoresDeCurso(ctx context.Context, cursoID string) ([]*InstructorAsignadoRow, error)
	// PuedeRevisar indica si el usuario es dueño O revisor del curso.
	//
	// Es el permiso de LECTURA de seguimiento. No autoriza a editar nada: para
	// eso el repositorio sigue comprobando `instructor_id` directamente.
	PuedeRevisar(ctx context.Context, cursoID, instructorID string) (bool, error)
}

func (r *postgresCursosRepository) AsignarInstructor(ctx context.Context, cursoID, instructorID, asignadoPor string) error {
	if cursoID == "" || instructorID == "" {
		return errors.New("capacitacion_id e instructor_id son obligatorios")
	}

	// No se deja asignar al propio dueño: ya puede todo, y aparecer también
	// como revisor haría creer al administrador que le dio un permiso que en
	// realidad no cambió nada.
	var dueno sql.NullString
	if err := r.db.GetContext(ctx, &dueno,
		`SELECT instructor_id::text FROM capacitaciones
		  WHERE id = $1 AND deleted_at IS NULL`, cursoID); err != nil {
		return err
	}
	if dueno.Valid && dueno.String == instructorID {
		return errYaEsDueno
	}

	var asignador any
	if asignadoPor != "" {
		asignador = asignadoPor
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO curso_instructores (capacitacion_id, instructor_id, asignado_por)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (capacitacion_id, instructor_id) DO NOTHING`,
		cursoID, instructorID, asignador)
	return err
}

func (r *postgresCursosRepository) QuitarInstructor(ctx context.Context, cursoID, instructorID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM curso_instructores
		  WHERE capacitacion_id = $1::uuid AND instructor_id = $2::uuid`,
		cursoID, instructorID)
	return err
}

func (r *postgresCursosRepository) ListInstructoresDeCurso(ctx context.Context, cursoID string) ([]*InstructorAsignadoRow, error) {
	var filas []*InstructorAsignadoRow
	return filas, r.db.SelectContext(ctx, &filas, `
		SELECT instructor_id::text AS instructor_id, asignado_at
		  FROM curso_instructores
		 WHERE capacitacion_id = $1::uuid
		 ORDER BY asignado_at ASC`, cursoID)
}

func (r *postgresCursosRepository) PuedeRevisar(ctx context.Context, cursoID, instructorID string) (bool, error) {
	if cursoID == "" || instructorID == "" {
		return false, nil
	}
	var puede bool
	// Una sola consulta para las dos vías. Separarlas en dos viajes dejaría una
	// ventana entre comprobar la titularidad y comprobar la asignación.
	err := r.db.GetContext(ctx, &puede, `
		SELECT EXISTS (
		    SELECT 1 FROM capacitaciones
		     WHERE id = $1::uuid AND instructor_id = $2::uuid AND deleted_at IS NULL
		    UNION ALL
		    SELECT 1 FROM curso_instructores
		     WHERE capacitacion_id = $1::uuid AND instructor_id = $2::uuid
		)`, cursoID, instructorID)
	return puede, err
}

// errYaEsDueno distingue "no hice nada porque ya podía" de un fallo real, para
// que el panel lo explique en vez de decir que la asignación falló.
var errYaEsDueno = errors.New("ese instructor ya es el dueño del curso")

// ErrYaEsDueno se exporta para que el servicio lo traduzca a un código gRPC.
var ErrYaEsDueno error = errYaEsDueno

// Compilación: el tipo debe seguir cumpliendo la interfaz.
var _ InstructoresRepository = (*postgresCursosRepository)(nil)

// cursosDelRevisor lista los cursos que un instructor puede revisar sin ser su
// dueño. Se usa para que esos cursos aparezcan en su panel.
func (r *postgresCursosRepository) cursosDelRevisor(ctx context.Context, instructorID string) ([]string, error) {
	var ids []string
	return ids, r.db.SelectContext(ctx, &ids,
		`SELECT capacitacion_id::text FROM curso_instructores WHERE instructor_id = $1::uuid`,
		instructorID)
}
