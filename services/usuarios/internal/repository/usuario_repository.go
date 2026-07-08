package repository

import (
	"context"
	"time"

	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/jmoiron/sqlx"
)

// Usuario es el modelo interno de este servicio.
type Usuario struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Role      string    `db:"role"`
	Bio       string    `db:"bio"`
	AvatarURL string    `db:"avatar_url"`
	CoverURL  string    `db:"cover_url"`
	Phone     string    `db:"phone"`
	Specialty string    `db:"specialty"`
	CreatedAt time.Time `db:"created_at"`
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
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		CursosInscritos: u.CursosInscritos,
		LeccionesCompletadas: u.LeccionesCompletadas,
		TotalLecciones: u.TotalLecciones,
		CursosCreados: u.CursosCreados,
		EstudiantesTotal: u.EstudiantesTotal,
		ExamenesCreados: u.ExamenesCreados,
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
		        COALESCE(specialty,'') specialty, created_at
		   FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	if u.Role == "user" {
		_ = r.db.GetContext(ctx, &u.CursosInscritos, `SELECT COUNT(DISTINCT capacitacion_id) FROM inscripciones WHERE user_id=$1`, id)
		_ = r.db.GetContext(ctx, &u.LeccionesCompletadas, `SELECT COUNT(*) FROM progreso_lecciones WHERE user_id=$1`, id)
		_ = r.db.GetContext(ctx, &u.TotalLecciones, `
			SELECT COUNT(*) FROM lecciones l
			JOIN inscripciones i ON l.capacitacion_id = i.capacitacion_id
			WHERE i.user_id=$1 AND l.deleted_at IS NULL`, id)
	} else if u.Role == "instructor" {
		_ = r.db.GetContext(ctx, &u.CursosCreados, `SELECT COUNT(*) FROM capacitaciones WHERE instructor_id=$1 AND deleted_at IS NULL`, id)
		_ = r.db.GetContext(ctx, &u.EstudiantesTotal, `
			SELECT COUNT(DISTINCT i.user_id) FROM inscripciones i
			JOIN capacitaciones c ON i.capacitacion_id = c.id
			WHERE c.instructor_id=$1 AND c.deleted_at IS NULL`, id)
		_ = r.db.GetContext(ctx, &u.ExamenesCreados, `SELECT COUNT(*) FROM examenes WHERE instructor_id=$1`, id)
	}
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
	// Fetch requester role
	var role string
	_ = r.db.GetContext(ctx, &role, `SELECT role FROM users WHERE id = $1`, requesterID)

	var q string
	var args []any
	if role == "admin" || role == "instructor" {
		q = `SELECT id, name, email, role, COALESCE(bio,'') bio, COALESCE(avatar_url,'') avatar_url,
		            COALESCE(cover_url,'') cover_url, COALESCE(phone,'') phone,
		            COALESCE(specialty,'') specialty, created_at
		     FROM users
		     WHERE name ILIKE $1 OR email ILIKE $1
		     ORDER BY name ASC LIMIT $2`
		args = []any{"%" + query + "%", limit}
	} else {
		// Isolate to same cohort
		q = `SELECT DISTINCT u.id, u.name, u.email, u.role, COALESCE(u.bio,'') bio, COALESCE(u.avatar_url,'') avatar_url,
		            COALESCE(u.cover_url,'') cover_url, COALESCE(u.phone,'') phone,
		            COALESCE(u.specialty,'') specialty, u.created_at
		     FROM users u
		     JOIN inscripciones i ON u.id = i.user_id
		     WHERE (u.name ILIKE $1 OR u.email ILIKE $1)
		       AND EXISTS (
		           SELECT 1 FROM inscripciones i2
		           WHERE i2.user_id = $3
		             AND i2.capacitacion_id = i.capacitacion_id
		             AND i2.licencia_id IS NOT DISTINCT FROM i.licencia_id
		       )
		     ORDER BY u.name ASC LIMIT $2`
		args = []any{"%" + query + "%", limit, requesterID}
	}

	var users []*Usuario
	return users, r.db.SelectContext(ctx, &users, q, args...)
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
