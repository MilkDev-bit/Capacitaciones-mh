package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Este archivo concentra la regla de visibilidad entre usuarios:
// "solo puedes escribir a quien comparte curso contigo".
//
// Por qué vive aquí y no en el gateway: la regla es una restricción de
// dominio, no de transporte. Si se validara solo en el gateway o en el
// frontend, cualquiera podría hacer POST /api/mensajes/<uuid-arbitrario> y
// saltársela — el filtro de búsqueda de usuarios que ya existía era
// exactamente eso, cosmético. Al validarla en el servicio, ninguna ruta de
// entrada puede evitarla.
//
// El servicio de mensajes comparte base de datos con el de cursos (todos los
// contenedores reciben el mismo DATABASE_URL), igual que el servicio de
// usuarios ya consulta `inscripciones` para el perfil público. Se sigue esa
// convención en lugar de introducir una llamada gRPC extra en la ruta caliente
// de cada mensaje enviado.

// cursosDelSolicitante lista los cursos del usuario pasado como bindvar `?`.
// cursosDelDestinatario hace lo mismo correlacionando con `t.id` de la query
// externa, por lo que no consume bindvars.
//
// Un usuario pertenece a un curso si:
//
//   - está inscrito           → inscripciones
//   - fue asignado por RR.HH. → asignaciones
//   - lo imparte              → capacitaciones.instructor_id
//
// Incluir al instructor es lo que permite que un alumno pueda escribirle sin
// necesidad de una excepción aparte: comparte curso con él por definición.
const cursosDelSolicitante = `
	           SELECT capacitacion_id FROM inscripciones  WHERE user_id = ?
	           UNION
	           SELECT capacitacion_id FROM asignaciones   WHERE user_id = ? AND capacitacion_id IS NOT NULL
	           UNION
	           SELECT id              FROM capacitaciones WHERE instructor_id = ? AND deleted_at IS NULL`

const cursosDelDestinatario = `
	           SELECT capacitacion_id FROM inscripciones  WHERE user_id = t.id
	           UNION
	           SELECT capacitacion_id FROM asignaciones   WHERE user_id = t.id AND capacitacion_id IS NOT NULL
	           UNION
	           SELECT id              FROM capacitaciones WHERE instructor_id = t.id AND deleted_at IS NULL`

// ContactosRepository resuelve a quién puede contactar un usuario.
type ContactosRepository interface {
	// FiltrarContactables devuelve el subconjunto de targetIDs que el
	// solicitante tiene permitido contactar. Nunca devuelve al propio
	// solicitante.
	FiltrarContactables(ctx context.Context, requesterID string, targetIDs []string) ([]string, error)
	// PuedeContactar es el caso de un solo destinatario.
	PuedeContactar(ctx context.Context, requesterID, targetID string) (bool, error)
	// EsMiembroDeGrupo indica si el usuario pertenece al grupo.
	EsMiembroDeGrupo(ctx context.Context, userID, grupoID string) (bool, error)
	// AdminDeGrupo devuelve el admin_id del grupo, o "" si no existe.
	AdminDeGrupo(ctx context.Context, grupoID string) (string, error)
}

type postgresContactosRepository struct{ db *sqlx.DB }

func NewContactosRepository(db *sqlx.DB) ContactosRepository {
	return &postgresContactosRepository{db: db}
}

func (r *postgresContactosRepository) FiltrarContactables(ctx context.Context, requesterID string, targetIDs []string) ([]string, error) {
	if requesterID == "" || len(targetIDs) == 0 {
		return nil, nil
	}

	// Se usa sqlx.In en lugar de un array de Postgres porque el operador
	// `= ANY($1::uuid[])` obliga a pasar por pq.Array/pgtype, y este módulo
	// se conecta con pgx stdlib: el IN expandido funciona igual con cualquier
	// driver y evita acoplar el repositorio al que esté configurado.
	//
	// Orden de los bindvars `?`: IN, t.id <>, subconsulta de rol, y los tres
	// del bloque cursosDelSolicitante.
	q := `
		SELECT t.id
		  FROM users t
		 WHERE t.id IN (?)
		   AND t.id <> ?
		   AND (
		         (SELECT role FROM users WHERE id = ?) IN ('admin', 'instructor')
		      OR EXISTS (
		           SELECT 1
		             FROM (` + cursosDelSolicitante + `
		             ) yo
		             JOIN (` + cursosDelDestinatario + `
		             ) otro ON yo.capacitacion_id = otro.capacitacion_id
		         )
		       )`

	query, args, err := sqlx.In(q, targetIDs, requesterID, requesterID, requesterID, requesterID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("contactos: construir consulta: %w", err)
	}

	var permitidos []string
	if err := r.db.SelectContext(ctx, &permitidos, r.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("contactos: consultar permitidos: %w", err)
	}
	return permitidos, nil
}

func (r *postgresContactosRepository) PuedeContactar(ctx context.Context, requesterID, targetID string) (bool, error) {
	permitidos, err := r.FiltrarContactables(ctx, requesterID, []string{targetID})
	if err != nil {
		return false, err
	}
	return len(permitidos) == 1, nil
}

func (r *postgresContactosRepository) EsMiembroDeGrupo(ctx context.Context, userID, grupoID string) (bool, error) {
	var existe bool
	err := r.db.GetContext(ctx, &existe,
		`SELECT EXISTS (SELECT 1 FROM grupo_miembros WHERE grupo_id = $1 AND usuario_id = $2)`,
		grupoID, userID)
	if err != nil {
		return false, fmt.Errorf("contactos: verificar membresía: %w", err)
	}
	return existe, nil
}

func (r *postgresContactosRepository) AdminDeGrupo(ctx context.Context, grupoID string) (string, error) {
	var adminID string
	err := r.db.GetContext(ctx, &adminID, `SELECT admin_id FROM grupos WHERE id = $1`, grupoID)
	if err != nil {
		return "", fmt.Errorf("contactos: obtener admin del grupo: %w", err)
	}
	return adminID, nil
}
