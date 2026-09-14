package repository

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

// ─────────────────────────────────────────────────────────────────────────────
// Compañeros de curso
//
// Responde a "¿quién comparte capacitación con este usuario?", que es la base
// de dos reglas de la plataforma: a quién puedes encontrar en el buscador y a
// quién puedes escribir.
//
// La consulta estaba duplicada en usuarios-service y en mensajes-service, que
// leían `inscripciones`, `asignaciones` y `capacitaciones` directamente. Esas
// tablas son de este servicio. En desarrollo colaba porque docker-compose da el
// mismo DATABASE_URL a todos los contenedores; en producción cada servicio
// tiene su base y la consulta fallaba con "relation does not exist".
// ─────────────────────────────────────────────────────────────────────────────

// pertenencia define qué significa "estar en un curso". Se usa dos veces —para
// el solicitante y para el otro— y por eso está escrita una sola vez, con el
// usuario como bindvar `?`.
//
// Un usuario pertenece a un curso si:
//
//   - está inscrito           → inscripciones
//   - lo asignó RR.HH.        → asignaciones
//   - lo imparte              → capacitaciones.instructor_id
//
// Incluir al instructor es lo que permite que un alumno le escriba sin una
// excepción aparte: comparte curso con él por definición.
const pertenencia = `
	    SELECT capacitacion_id FROM inscripciones  WHERE user_id = ?
	    UNION
	    SELECT capacitacion_id FROM asignaciones   WHERE user_id = ? AND capacitacion_id IS NOT NULL
	    UNION
	    SELECT id              FROM capacitaciones WHERE instructor_id = ? AND deleted_at IS NULL`

type CompanerosRepository interface {
	// CompanerosDeCurso devuelve los usuarios que comparten capacitación con
	// userID, sin incluirlo a él.
	//
	// Con candidatos, restringe la respuesta a ese conjunto. Sin ellos,
	// devuelve todos, hasta `limite` (0 = sin límite).
	CompanerosDeCurso(ctx context.Context, userID string, candidatos []string, limite int32) ([]string, error)
}

func (r *postgresCursosRepository) CompanerosDeCurso(
	ctx context.Context, userID string, candidatos []string, limite int32,
) ([]string, error) {
	if userID == "" {
		return nil, errors.New("user_id vacío")
	}

	// El conjunto se calcula una sola vez y se cruza consigo mismo por curso.
	// No se devuelven nombres: `users` vive en la base de auth, no en esta.
	q := `
		WITH mis_cursos AS (` + pertenencia + `
		)
		SELECT DISTINCT otro.user_id::text
		  FROM (
		      SELECT user_id, capacitacion_id FROM inscripciones
		      UNION
		      SELECT user_id, capacitacion_id FROM asignaciones WHERE capacitacion_id IS NOT NULL
		      UNION
		      SELECT instructor_id, id        FROM capacitaciones
		       WHERE instructor_id IS NOT NULL AND deleted_at IS NULL
		  ) otro
		  JOIN mis_cursos m ON m.capacitacion_id = otro.capacitacion_id
		 WHERE otro.user_id <> ?`

	// Tres veces userID: la CTE de pertenencia lo usa en sus tres ramas.
	args := []any{userID, userID, userID, userID}

	if len(candidatos) > 0 {
		q += ` AND otro.user_id IN (?)`
		args = append(args, candidatos)
	}
	if limite > 0 {
		q += ` LIMIT ?`
		args = append(args, limite)
	}

	// sqlx.In expande el IN a marcadores normales. Se llama siempre, con o sin
	// candidatos, porque también es quien aplana los argumentos; Rebind traduce
	// los `?` a los $1, $2… que espera Postgres.
	consulta, argsIn, err := sqlx.In(q, args...)
	if err != nil {
		return nil, err
	}

	var ids []string
	if err := r.db.SelectContext(ctx, &ids, r.db.Rebind(consulta), argsIn...); err != nil {
		return nil, err
	}
	return ids, nil
}
