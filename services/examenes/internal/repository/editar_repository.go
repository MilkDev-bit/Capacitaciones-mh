package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	examenespb "Prueba-Go/gen/examenes"

	"github.com/jmoiron/sqlx"
)

// ─────────────────────────────────────────────────────────────────────────────
// Edición de exámenes
//
// La regla que gobierna todo este archivo está en el esquema:
//
//	respuestas_examen.pregunta_id REFERENCES preguntas(id) ON DELETE CASCADE
//	respuestas_examen.opcion_id   REFERENCES opciones(id)  ON DELETE SET NULL
//
// Borrar una pregunta borra las respuestas de todos los que ya la contestaron.
// Por eso guardar NO es "borrar todo y reinsertar", que es lo fácil: es
// comparar con lo que había y tocar solo lo que cambió.
//
//	pregunta con id que sigue en la lista → UPDATE, conserva sus respuestas
//	pregunta con id que ya no viene       → DELETE, y se lleva sus respuestas
//	pregunta sin id                       → INSERT
//
// Las opciones siguen la misma lógica dentro de cada pregunta.
// ─────────────────────────────────────────────────────────────────────────────

// ErrNoEsTuyo lo devuelve la actualización cuando el examen es de otro
// instructor. Va aquí y no en `service` porque es este paquete el que hace la
// comprobación, dentro de la misma transacción que escribe.
var ErrNoEsTuyo = errors.New("el examen no es de este instructor")

type EditarRepository interface {
	// Update aplica los cambios conservando lo que no cambió.
	Update(ctx context.Context, req *examenespb.UpdateExamenRequest) (*Examen, error)
	// RespuestasPorPregunta cuenta cuánta gente respondió cada pregunta, para
	// poder avisar antes de que el instructor borre algo con historial.
	RespuestasPorPregunta(ctx context.Context, examenID string) (map[string]int32, error)
}

func (r *postgresExamenesRepository) RespuestasPorPregunta(ctx context.Context, examenID string) (map[string]int32, error) {
	type fila struct {
		PreguntaID string `db:"pregunta_id"`
		Total      int32  `db:"total"`
	}
	var filas []fila
	if err := r.db.SelectContext(ctx, &filas, `
		SELECT pregunta_id::text AS pregunta_id, COUNT(*)::int AS total
		  FROM respuestas_examen
		 WHERE examen_id = $1
		 GROUP BY pregunta_id`, examenID); err != nil {
		return nil, err
	}
	conteo := make(map[string]int32, len(filas))
	for _, f := range filas {
		conteo[f.PreguntaID] = f.Total
	}
	return conteo, nil
}

func (r *postgresExamenesRepository) Update(ctx context.Context, req *examenespb.UpdateExamenRequest) (*Examen, error) {
	if req.ExamenId == "" {
		return nil, errors.New("examen_id vacío")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Propiedad dentro de la MISMA transacción que escribe. Comprobarla antes,
	// en una consulta aparte, dejaría una ventana entre el permiso y el cambio.
	var dueno sql.NullString
	err = tx.GetContext(ctx, &dueno,
		`SELECT instructor_id::text FROM examenes
		  WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, req.ExamenId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	// Un examen sin instructor es de los antiguos o lo creó un admin; solo un
	// admin puede editarlo, y el servicio decide eso pasando user_id vacío.
	if req.UserId != "" && (!dueno.Valid || dueno.String != req.UserId) {
		return nil, ErrNoEsTuyo
	}

	var capID *string
	if req.CapacitacionId != "" {
		capID = &req.CapacitacionId
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE examenes SET title=$1, description=$2, capacitacion_id=$3 WHERE id=$4`,
		req.Title, req.Description, capID, req.ExamenId); err != nil {
		return nil, err
	}

	if err := sincronizarPreguntas(ctx, tx, req.ExamenId, req.Preguntas); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, req.ExamenId)
}

// sincronizarPreguntas deja la tabla `preguntas` igual a la lista recibida,
// tocando lo mínimo.
func sincronizarPreguntas(ctx context.Context, tx *sqlx.Tx, examenID string, entrada []*examenespb.PreguntaInput) error {
	// Ids que sobreviven. Se recogen ANTES de tocar nada para poder borrar en
	// bloque lo que no está en la lista.
	conservar := make([]string, 0, len(entrada))
	for _, p := range entrada {
		if p.Id != "" {
			conservar = append(conservar, p.Id)
		}
	}

	// Borrado de las preguntas retiradas. El `NOT IN` se construye con sqlx.In
	// para no concatenar identificadores en el SQL.
	//
	// Con la lista vacía —el instructor quitó todas— no se puede usar
	// `NOT IN ()`, que es sintaxis inválida: se borran todas y ya.
	if len(conservar) == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM preguntas WHERE examen_id = $1`, examenID); err != nil {
			return err
		}
	} else {
		q, args, err := sqlx.In(
			`DELETE FROM preguntas WHERE examen_id = ? AND id NOT IN (?)`, examenID, conservar)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(q), args...); err != nil {
			return err
		}
	}

	for _, p := range entrada {
		tipo := p.Tipo
		if tipo == "" {
			tipo = "multiple_choice"
		}

		preguntaID := p.Id
		if preguntaID == "" {
			if err := tx.QueryRowContext(ctx,
				`INSERT INTO preguntas(examen_id,texto,tipo,valor,orden)
				 VALUES($1,$2,$3,$4,$5) RETURNING id`,
				examenID, p.Texto, tipo, p.Valor, p.Orden,
			).Scan(&preguntaID); err != nil {
				return err
			}
		} else {
			// El `examen_id = $5` no sobra: impide que alguien mande el id de
			// una pregunta de OTRO examen y la reescriba desde aquí.
			res, err := tx.ExecContext(ctx,
				`UPDATE preguntas SET texto=$1, tipo=$2, valor=$3, orden=$4
				  WHERE id=$5 AND examen_id=$6`,
				p.Texto, tipo, p.Valor, p.Orden, preguntaID, examenID)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return fmt.Errorf("la pregunta %s no pertenece a este examen", preguntaID)
			}
		}

		if err := sincronizarOpciones(ctx, tx, preguntaID, p.Opciones); err != nil {
			return err
		}
	}
	return nil
}

// sincronizarOpciones hace lo mismo dentro de una pregunta.
//
// Importa tanto como el de arriba: `respuestas_examen.opcion_id` apunta a
// `opciones`, y aunque su ON DELETE es SET NULL y no CASCADE, borrar y
// reinsertar una opción convierte la respuesta de un alumno en un NULL que ya
// no se puede calificar.
func sincronizarOpciones(ctx context.Context, tx *sqlx.Tx, preguntaID string, entrada []*examenespb.OpcionInput) error {
	conservar := make([]string, 0, len(entrada))
	for _, o := range entrada {
		if o.Id != "" {
			conservar = append(conservar, o.Id)
		}
	}

	if len(conservar) == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM opciones WHERE pregunta_id = $1`, preguntaID); err != nil {
			return err
		}
	} else {
		q, args, err := sqlx.In(
			`DELETE FROM opciones WHERE pregunta_id = ? AND id NOT IN (?)`, preguntaID, conservar)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(q), args...); err != nil {
			return err
		}
	}

	for _, o := range entrada {
		if o.Id == "" {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO opciones(pregunta_id,texto,es_correcta) VALUES($1,$2,$3)`,
				preguntaID, o.Texto, o.EsCorrecta); err != nil {
				return err
			}
			continue
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE opciones SET texto=$1, es_correcta=$2 WHERE id=$3 AND pregunta_id=$4`,
			o.Texto, o.EsCorrecta, o.Id, preguntaID); err != nil {
			return err
		}
	}
	return nil
}
