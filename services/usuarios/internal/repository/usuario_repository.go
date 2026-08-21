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
	Search(ctx context.Context, query string, limit int, requesterID string) ([]*Usuario, error)
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
	_ = r.db.GetContext(ctx, &u.CursosInscritos, `SELECT COUNT(DISTINCT capacitacion_id) FROM inscripciones WHERE user_id=$1`, id)
	_ = r.db.GetContext(ctx, &u.LeccionesCompletadas, `SELECT COUNT(*) FROM progreso_lecciones WHERE user_id=$1`, id)
	_ = r.db.GetContext(ctx, &u.TotalLecciones, `
		SELECT COUNT(*) FROM lecciones l
		JOIN inscripciones i ON l.capacitacion_id = i.capacitacion_id
		WHERE i.user_id=$1 AND l.deleted_at IS NULL`, id)
	_ = r.db.GetContext(ctx, &u.CursosCreados, `SELECT COUNT(*) FROM capacitaciones WHERE instructor_id=$1 AND deleted_at IS NULL`, id)
	_ = r.db.GetContext(ctx, &u.EstudiantesTotal, `
		SELECT COUNT(DISTINCT i.user_id) FROM inscripciones i
		JOIN capacitaciones c ON i.capacitacion_id = c.id
		WHERE c.instructor_id=$1 AND c.deleted_at IS NULL`, id)
	_ = r.db.GetContext(ctx, &u.ExamenesCreados, `SELECT COUNT(*) FROM examenes WHERE instructor_id=$1`, id)
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

func (r *postgresUsuarioRepository) Search(ctx context.Context, query string, limit int, requesterID string) ([]*Usuario, error) {
	if limit <= 0 {
		limit = 10
	}
	// Fetch requester role if requesterID is present
	var role string
	if requesterID != "" {
		_ = r.db.GetContext(ctx, &role, `SELECT role FROM users WHERE id = $1`, requesterID)
	}

	const columnas = `u.id, u.name, u.email, u.role, COALESCE(u.bio,'') bio,
	                  COALESCE(u.avatar_url,'') avatar_url, COALESCE(u.cover_url,'') cover_url,
	                  COALESCE(u.phone,'') phone, COALESCE(u.specialty,'') specialty, u.created_at`

	// Admin e instructor buscan sin límite: necesitan dar soporte y coordinar
	// fuera de su propio grupo.
	if role == "admin" || role == "instructor" {
		q := `SELECT ` + columnas + `
		        FROM users u
		       WHERE (u.name ILIKE $1 OR u.email ILIKE $1)
		         AND u.id <> $3
		       ORDER BY u.name ASC LIMIT $2`
		var users []*Usuario
		return users, r.db.SelectContext(ctx, &users, q, "%"+query+"%", limit, requesterID)
	}

	// Sin solicitante identificado no se devuelve nada. Antes este caso —y el
	// de un `role` vacío por fallo de la consulta anterior— caía en la rama
	// sin filtro y exponía el directorio completo de la plataforma.
	if requesterID == "" {
		return nil, nil
	}

	// Alumno: solo compañeros de sus capacitaciones. La pertenencia a un curso
	// incluye inscripción, asignación por RR.HH. e impartición, de modo que el
	// instructor del curso aparece en los resultados de sus propios alumnos.
	//
	// Nota: esta consulta es solo la capa de descubrimiento. La autorización
	// real de cada mensaje vive en mensajes-service, porque un filtro de
	// búsqueda no impide llamar al endpoint de envío con un ID adivinado.
	q := `
		WITH mis_cursos AS (
		    SELECT capacitacion_id AS id FROM inscripciones  WHERE user_id = $3
		    UNION
		    SELECT capacitacion_id      FROM asignaciones   WHERE user_id = $3 AND capacitacion_id IS NOT NULL
		    UNION
		    SELECT id                   FROM capacitaciones WHERE instructor_id = $3 AND deleted_at IS NULL
		),
		companeros AS (
		    SELECT i.user_id FROM inscripciones  i JOIN mis_cursos m ON m.id = i.capacitacion_id
		    UNION
		    SELECT a.user_id FROM asignaciones   a JOIN mis_cursos m ON m.id = a.capacitacion_id
		    UNION
		    SELECT c.instructor_id FROM capacitaciones c
		      JOIN mis_cursos m ON m.id = c.id
		     WHERE c.instructor_id IS NOT NULL AND c.deleted_at IS NULL
		)
		SELECT ` + columnas + `
		  FROM users u
		 WHERE (u.name ILIKE $1 OR u.email ILIKE $1)
		   AND u.id <> $3
		   AND u.id IN (SELECT user_id FROM companeros)
		 ORDER BY u.name ASC LIMIT $2`

	var users []*Usuario
	return users, r.db.SelectContext(ctx, &users, q, "%"+query+"%", limit, requesterID)
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
func (r *postgresUsuarioRepository) CreateNotificacion(ctx context.Context, req *usuariospb.CreateNotificacionRequest) (string, bool, error) {
	const query = `
		INSERT INTO notificaciones (user_id, tipo, titulo, mensaje, enlace)
		SELECT $1::uuid, $2, $3, $4, NULLIF($5, '')
		WHERE $6 <= 0 OR NOT EXISTS (
			SELECT 1 FROM notificaciones
			 WHERE user_id = $1::uuid
			   AND tipo    = $2
			   AND titulo  = $3
			   AND mensaje = $4
			   AND COALESCE(enlace, '') = $5
			   AND leida = false
			   AND created_at > NOW() - make_interval(secs => $6::double precision)
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
