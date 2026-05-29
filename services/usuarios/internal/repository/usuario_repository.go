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
}

func (u *Usuario) ToProto() *usuariospb.PerfilResponse {
	return &usuariospb.PerfilResponse{
		Id: u.ID, Name: u.Name, Email: u.Email, Role: u.Role,
		Bio: u.Bio, AvatarUrl: u.AvatarURL, CoverUrl: u.CoverURL,
		Phone: u.Phone, Specialty: u.Specialty,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (u *Usuario) ToSummaryProto() *usuariospb.UserSummary {
	return &usuariospb.UserSummary{
		Id: u.ID, Name: u.Name, Email: u.Email, Role: u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// UsuarioRepository define el contrato de acceso a datos.
type UsuarioRepository interface {
	FindByID(ctx context.Context, id string) (*Usuario, error)
	UpdatePerfil(ctx context.Context, req *usuariospb.UpdatePerfilRequest) error
	UpdateField(ctx context.Context, userID, field, value string) error
	List(ctx context.Context, role string) ([]*Usuario, error)
	Delete(ctx context.Context, userID string) error
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
	return u, err
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
