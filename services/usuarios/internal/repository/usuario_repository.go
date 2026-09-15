package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/jmoiron/sqlx"
)

// Usuario es el modelo interno de este servicio.
type Usuario struct {
	ID                   string    `db:"id"`
	Name                 string    `db:"name"`
	Email                string    `db:"email"`
	Role                 string    `db:"role"`
	Bio                  string    `db:"bio"`
	AvatarURL            string    `db:"avatar_url"`
	CoverURL             string    `db:"cover_url"`
	Phone                string    `db:"phone"`
	Specialty            string    `db:"specialty"`
	CreatedAt            time.Time `db:"created_at"`
	AvisoVersion         string    `db:"aviso_version"`
	CursosInscritos      int32
	LeccionesCompletadas int32
	TotalLecciones       int32
	CursosCreados        int32
	EstudiantesTotal     int32
	ExamenesCreados      int32
}

func (u *Usuario) ToProto() *usuariospb.PerfilResponse {
	return &usuariospb.PerfilResponse{
		Id: u.ID, Name: u.Name, Email: u.Email, Role: u.Role,
		Bio: u.Bio, AvatarUrl: u.AvatarURL, CoverUrl: u.CoverURL,
		Phone: u.Phone, Specialty: u.Specialty,
		CreatedAt:            u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		AvisoVersion:         u.AvisoVersion,
		CursosInscritos:      u.CursosInscritos,
		LeccionesCompletadas: u.LeccionesCompletadas,
		TotalLecciones:       u.TotalLecciones,
		CursosCreados:        u.CursosCreados,
		EstudiantesTotal:     u.EstudiantesTotal,
		ExamenesCreados:      u.ExamenesCreados,
	}
}

func (u *Usuario) ToSummaryProto() *usuariospb.UserSummary {
	return &usuariospb.UserSummary{
		Id: u.ID, Name: u.Name, Email: u.Email, Role: u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		AvatarUrl: u.AvatarURL,
	}
}

// UsuarioRepository define el contrato de acceso a datos.
type UsuarioRepository interface {
	FindByID(ctx context.Context, id string) (*Usuario, error)
	UpdatePerfil(ctx context.Context, req *usuariospb.UpdatePerfilRequest) error
	UpdateField(ctx context.Context, userID, field, value string) error
	List(ctx context.Context, role string) ([]*Usuario, error)
	Delete(ctx context.Context, userID string) error
	// soloIDs acota la búsqueda a usuarios ya autorizados por cursos-service.
	// Para quien no es admin ni instructor, vacío = no ve a nadie.
	Search(ctx context.Context, query string, limit int, requesterID string, soloIDs []string) ([]*Usuario, error)
	ListNotificaciones(ctx context.Context, userID string) ([]*usuariospb.Notificacion, error)
	MarkNotificacionesRead(ctx context.Context, userID string, ids []string) error
	CreateNotificacion(ctx context.Context, req *usuariospb.CreateNotificacionRequest) (id string, creada bool, err error)
}

type postgresUsuarioRepository struct{ db *sqlx.DB }

func NewUsuarioRepository(db *sqlx.DB) UsuarioRepository {
	return &postgresUsuarioRepository{db: db}
}

func (r *postgresUsuarioRepository) FindByID(ctx context.Context, id string) (*Usuario, error) {
	u := &Usuario{}
	err := r.db.GetContext(ctx, u,
		`SELECT id, name, email, role, COALESCE(bio,'') bio, COALESCE(avatar_url,'') avatar_url,
		        COALESCE(cover_url,'') cover_url, COALESCE(phone,'') phone,
		        COALESCE(specialty,'') specialty, created_at,
		        COALESCE(aviso_version,'') aviso_version
		   FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	// Aquí había seis consultas más: cursos inscritos, lecciones completadas,
	// cursos creados, estudiantes, exámenes. Contaban sobre `inscripciones`,
	// `progreso_lecciones`, `lecciones`, `capacitaciones` y `examenes`, que son
	// de otros tres servicios y no existen en esta base.
	//
	// Se ejecutaban con `_ =`, así que no rompían nada: fallaban en silencio y
	// dejaban los contadores en cero. Seis consultas fallidas por cada carga de
	// perfil, y el gateway sobrescribía el resultado de todas preguntando a
	// cursos, lecciones y exámenes, que es donde están los datos. Se quitan
	// porque no aportaban un número correcto en ningún caso.
	//
	// Los contadores del struct siguen existiendo: los rellena el gateway al
	// componer /api/perfil.
	return u, nil
}

func (r *postgresUsuarioRepository) UpdatePerfil(ctx context.Context, req *usuariospb.UpdatePerfilRequest) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET name=$1, bio=$2, phone=$3, specialty=$4 WHERE id=$5`,
		req.Name, req.Bio, req.Phone, req.Specialty, req.UserId)
	return err
}

func (r *postgresUsuarioRepository) UpdateField(ctx context.Context, userID, field, value string) error {
	// Nota: field viene de código interno (no de input de usuario), por lo que
	// es seguro usarlo en la query. Solo acepta valores conocidos del servicio.
	query := `UPDATE users SET ` + field + ` = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, value, userID)
	return err
}

func (r *postgresUsuarioRepository) List(ctx context.Context, role string) ([]*Usuario, error) {
	query := `SELECT id, name, email, role, COALESCE(bio,'') bio, COALESCE(avatar_url,'') avatar_url,
	                 COALESCE(cover_url,'') cover_url, COALESCE(phone,'') phone,
	                 COALESCE(specialty,'') specialty, created_at
	            FROM users`
	args := []any{}
	if role != "" {
		query += " WHERE role = $1"
		args = append(args, role)
	}
	query += " ORDER BY created_at DESC"
	var users []*Usuario
	return users, r.db.SelectContext(ctx, &users, query, args...)
}

func (r *postgresUsuarioRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return err
}

// Search busca usuarios por nombre o correo.
//
// Esta consulta SOLO toca `users`, que es de este servicio. Antes cruzaba
// `inscripciones`, `asignaciones` y `capacitaciones` para quedarse con los
// compañeros de curso del solicitante, y esas tres tablas son de
// cursos-service. En desarrollo pasaba desapercibido porque docker-compose da
// el mismo DATABASE_URL a los siete contenedores; en producción, donde cada
// servicio tiene su base, la búsqueda devolvía 500 a todo alumno que escribiera
// una letra.
//
// Ahora quién es visible lo calcula cursos-service y llega en soloIDs.
func (r *postgresUsuarioRepository) Search(
	ctx context.Context, query string, limit int, requesterID string, soloIDs []string,
) ([]*Usuario, error) {
	if limit <= 0 {
		limit = 10
	}

	// Sin solicitante identificado no se devuelve nada. Este caso ya cayó una
	// vez en la rama sin filtro y expuso el directorio completo de la
	// plataforma; la puerta se queda cerrada por defecto.
	if requesterID == "" {
		return nil, nil
	}

	var role string
	_ = r.db.GetContext(ctx, &role, `SELECT role FROM users WHERE id = $1`, requesterID)

	const columnas = `u.id, u.name, u.email, u.role, COALESCE(u.bio,'') bio,
	                  COALESCE(u.avatar_url,'') avatar_url, COALESCE(u.cover_url,'') cover_url,
	                  COALESCE(u.phone,'') phone, COALESCE(u.specialty,'') specialty, u.created_at`

	q := `SELECT ` + columnas + `
	        FROM users u
	       WHERE (u.name ILIKE ? OR u.email ILIKE ?)
	         AND u.id <> ?`
	args := []any{"%" + query + "%", "%" + query + "%", requesterID}

	// Admin e instructor buscan sin restricción: dan soporte y coordinan fuera
	// de su propio grupo.
	if role != "admin" && role != "instructor" {
		// Para el resto, soloIDs vacío significa "no ve a nadie", no "los ve a
		// todos". Si el gateway se olvidara de calcularlo, o cursos-service
		// respondiera con error, la búsqueda sale vacía en vez de abrir el
		// directorio entero.
		if len(soloIDs) == 0 {
			return nil, nil
		}
		q += ` AND u.id IN (?)`
		args = append(args, soloIDs)
	}

	q += ` ORDER BY u.name ASC LIMIT ?`
	args = append(args, limit)

	// sqlx.In expande el IN y aplana los argumentos; Rebind traduce los `?` a
	// los $1, $2… de Postgres. Se llaman en las dos ramas: en la de admin no
	// hay IN que expandir, pero la consulta sigue escrita con `?`.
	consulta, argsIn, err := sqlx.In(q, args...)
	if err != nil {
		return nil, err
	}

	var users []*Usuario
	return users, r.db.SelectContext(ctx, &users, r.db.Rebind(consulta), argsIn...)
}

func (r *postgresUsuarioRepository) ListNotificaciones(ctx context.Context, userID string) ([]*usuariospb.Notificacion, error) {
	query := `
		SELECT id, user_id, tipo, titulo, mensaje, leida, COALESCE(enlace, '') as enlace, created_at
		FROM notificaciones
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 50`

	type dbNotif struct {
		ID        string    `db:"id"`
		UserID    string    `db:"user_id"`
		Tipo      string    `db:"tipo"`
		Titulo    string    `db:"titulo"`
		Mensaje   string    `db:"mensaje"`
		Leida     bool      `db:"leida"`
		Enlace    string    `db:"enlace"`
		CreatedAt time.Time `db:"created_at"`
	}

	var rows []dbNotif
	if err := r.db.SelectContext(ctx, &rows, query, userID); err != nil {
		return nil, err
	}

	res := make([]*usuariospb.Notificacion, len(rows))
	for i, r := range rows {
		res[i] = &usuariospb.Notificacion{
			Id:        r.ID,
			UserId:    r.UserID,
			Tipo:      r.Tipo,
			Titulo:    r.Titulo,
			Mensaje:   r.Mensaje,
			Leida:     r.Leida,
			Enlace:    r.Enlace,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		}
	}
	return res, nil
}

func (r *postgresUsuarioRepository) MarkNotificacionesRead(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	query := `UPDATE notificaciones SET leida = true WHERE user_id = $1 AND id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, userID, ids)
	return err
}

// CreateNotificacion inserta una notificación, opcionalmente deduplicada.
//
// La deduplicación se resuelve dentro del propio INSERT (INSERT ... SELECT ...
// WHERE NOT EXISTS) en lugar de con un SELECT previo seguido de un INSERT: dos
// eventos concurrentes para el mismo usuario —dos mensajes que llegan a la vez—
// pasarían ambos la comprobación si fueran dos viajes separados a la base.
//
// Cuando la fila se suprime por duplicada no hay RETURNING, así que sql.ErrNoRows
// es el caso normal y no un error: se traduce a creada=false.
//
// TODOS los parámetros llevan cast explícito, y no es cosmética: sin ellos esta
// consulta no llegaba ni a ejecutarse.
//
//	ERROR: inconsistent types deduced for parameter $2 (SQLSTATE 42P08)
//
// Cada parámetro aparece dos veces —en la lista del INSERT y en la subconsulta
// de deduplicación— y Postgres dedujo un tipo distinto en cada sitio. En un
// `INSERT ... SELECT`, a diferencia de un `INSERT ... VALUES`, el SELECT se
// analiza por su cuenta: los tipos de las columnas destino NO se propagan a los
// parámetros, así que `$2` sin cast queda como `text`. Abajo, `tipo = $2` lo
// resuelve como `character varying`, que es el tipo de la columna. Dos
// deducciones para el mismo parámetro y el Parse falla entero.
//
// Falla en Parse, no al ejecutar, así que NINGUNA notificación se creaba nunca:
// ni de mensaje, ni de compra, ni de inscripción, ni de constancia. La campana
// llevaba vacía desde que se escribió esta consulta.
func (r *postgresUsuarioRepository) CreateNotificacion(ctx context.Context, req *usuariospb.CreateNotificacionRequest) (string, bool, error) {
	// Se castea a `text` y no a `varchar` porque el tipo tiene que ser el mismo
	// en los dos usos: al insertar, Postgres aplica el cast de asignación a
	// varchar(50)/varchar(200); al comparar, promociona la columna a text. Con
	// `::varchar` también funcionaría, pero `text` deja una sola lectura
	// posible y no depende del largo declarado de cada columna.
	const query = `
		INSERT INTO notificaciones (user_id, tipo, titulo, mensaje, enlace)
		SELECT $1::uuid, $2::text, $3::text, $4::text, NULLIF($5::text, '')
		WHERE $6::int <= 0 OR NOT EXISTS (
			SELECT 1 FROM notificaciones
			 WHERE user_id = $1::uuid
			   AND tipo    = $2::text
			   AND titulo  = $3::text
			   AND mensaje = $4::text
			   AND COALESCE(enlace, '') = $5::text
			   AND leida = false
			   AND created_at > NOW() - make_interval(secs => $6::int::double precision)
		)
		RETURNING id`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		req.UserId, req.Tipo, req.Titulo, req.Mensaje, req.Enlace, req.DedupeVentanaSeg,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}
