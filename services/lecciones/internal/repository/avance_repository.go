package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	leccionespb "Prueba-Go/gen/lecciones"

	"github.com/jmoiron/sqlx"
)

// ─────────────────────────────────────────────────────────────────────────────
// Resumen de avance por curso
//
// Para la lista de seguimiento del instructor: una fila por alumno con cuántas
// lecciones lleva completadas, cuándo fue la última y qué puntos acumula.
//
// Se resuelve en UNA consulta y no llamando al detalle una vez por alumno. Con
// treinta inscritos, lo segundo son treinta viajes de red y treinta consultas
// para pintar una tabla que cabe en pantalla.
// ─────────────────────────────────────────────────────────────────────────────

type AvanceRepository interface {
	ResumenAvanceCurso(ctx context.Context, cursoID string, userIDs []string) (*leccionespb.ResumenAvanceResponse, error)
}

func (r *postgresLeccionesRepository) ResumenAvanceCurso(
	ctx context.Context, cursoID string, userIDs []string,
) (*leccionespb.ResumenAvanceResponse, error) {
	if cursoID == "" {
		return nil, errors.New("curso_id vacío")
	}

	resp := &leccionespb.ResumenAvanceResponse{}

	// Total de lecciones del curso. Viaja aparte porque es igual para todos y
	// permite pintar la barra de un alumno que aún no tiene ninguna fila de
	// progreso: sin este dato su avance sería "0 de 0", que se lee como
	// completado.
	if err := r.db.GetContext(ctx, &resp.TotalLecciones, `
		SELECT COUNT(*) FROM lecciones
		 WHERE capacitacion_id = $1 AND deleted_at IS NULL`, cursoID); err != nil {
		return nil, err
	}

	type fila struct {
		UserID      string       `db:"user_id"`
		Completadas int32        `db:"completadas"`
		Ultima      sql.NullTime `db:"ultima"`
		Puntos      int32        `db:"puntos"`
	}
	var filas []fila

	// LEFT JOIN desde progreso hacia game_scores y agregación por usuario.
	// El filtro por lección del curso va en el JOIN y no en el WHERE para no
	// perder a quien tiene puntos de juego pero ninguna lección completada.
	consulta := `
		WITH progreso AS (
		    SELECT p.user_id,
		           COUNT(*)              AS completadas,
		           MAX(p.completado_at)  AS ultima
		      FROM progreso_lecciones p
		      JOIN lecciones l ON l.id = p.leccion_id
		     WHERE l.capacitacion_id = ? AND l.deleted_at IS NULL
		     GROUP BY p.user_id
		),
		puntos AS (
		    SELECT user_id, COALESCE(SUM(points), 0) AS puntos
		      FROM game_scores
		     WHERE capacitacion_id = ?
		     GROUP BY user_id
		)
		SELECT COALESCE(pr.user_id, pt.user_id)::text AS user_id,
		       COALESCE(pr.completadas, 0)            AS completadas,
		       pr.ultima                              AS ultima,
		       COALESCE(pt.puntos, 0)::int            AS puntos
		  FROM progreso pr
		  FULL OUTER JOIN puntos pt ON pt.user_id = pr.user_id`

	// DOS veces cursoID: la consulta lo usa en las dos CTE, `progreso` y
	// `puntos`. Pasar uno solo desplaza todos los marcadores y la consulta falla
	// o, peor, filtra por el valor equivocado.
	args := []any{cursoID, cursoID}

	if len(userIDs) > 0 {
		// sqlx.In y no `= ANY($n)` con pq.Array: el driver de este servicio es
		// pgx, y lib/pq solo figura como dependencia indirecta. Traerla como
		// directa para una sola función sería pagar un driver entero por un
		// helper. sqlx ya está aquí y expande el IN a marcadores normales.
		consulta += ` WHERE COALESCE(pr.user_id, pt.user_id) IN (?)`
		expandida, argsIn, err := sqlx.In(consulta, cursoID, cursoID, userIDs)
		if err != nil {
			return nil, err
		}
		consulta, args = expandida, argsIn
	}

	// Rebind siempre, con o sin IN: la consulta está escrita con `?` y Postgres
	// espera $1, $2… Olvidarlo en la rama sin filtro dejaría un SQL inválido que
	// solo falla cuando se pide el curso completo.
	if err := r.db.SelectContext(ctx, &filas, r.db.Rebind(consulta), args...); err != nil {
		return nil, err
	}

	// Índice por usuario para poder completar con ceros a continuación.
	porUsuario := make(map[string]*leccionespb.AvanceAlumno, len(filas))
	for _, f := range filas {
		a := &leccionespb.AvanceAlumno{
			UserId:      f.UserID,
			Completadas: f.Completadas,
			Total:       resp.TotalLecciones,
			Puntos:      f.Puntos,
		}
		if f.Ultima.Valid {
			a.UltimaActividad = f.Ultima.Time.Format(time.RFC3339)
		}
		porUsuario[f.UserID] = a
	}

	// Una fila por alumno pedido, aunque no tenga nada registrado.
	//
	// Es lo que hace útil la pantalla: el inscrito que no ha abierto una sola
	// lección es justo a quien hay que perseguir, y si se omitiera por no tener
	// filas en `progreso_lecciones` sería invisible.
	if len(userIDs) > 0 {
		for _, id := range userIDs {
			if a, ok := porUsuario[id]; ok {
				resp.Avances = append(resp.Avances, a)
				continue
			}
			resp.Avances = append(resp.Avances, &leccionespb.AvanceAlumno{
				UserId: id,
				Total:  resp.TotalLecciones,
			})
		}
		return resp, nil
	}

	for _, a := range porUsuario {
		resp.Avances = append(resp.Avances, a)
	}
	return resp, nil
}
