package repository

import (
	"context"
	"database/sql"
	"time"

	"Prueba-Go/services/auth/internal/model"

	"github.com/jmoiron/sqlx"
)

// ErrNotFound se devuelve cuando no se encuentra ningún registro.
var ErrNotFound = sql.ErrNoRows

// UserRepository define las operaciones de acceso a datos del auth service.
// La interfaz permite inyectar fakes/mocks en los tests unitarios.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	// UpdateTokenVersion incrementa token_version del usuario en 1.
	UpdateTokenVersion(ctx context.Context, userID string) error
	// StorePasswordResetToken guarda el token de reset con su fecha de expiración.
	StorePasswordResetToken(ctx context.Context, userID, token string, expiry time.Time) error
	// FindByResetToken busca el usuario asociado a un token de reset válido (no expirado).
	FindByResetToken(ctx context.Context, token string) (*model.User, error)
	// UpdatePassword actualiza el hash de la contraseña.
	UpdatePassword(ctx context.Context, userID, hashedPassword string) error
	// ClearPasswordResetToken elimina el token de reset del usuario.
	ClearPasswordResetToken(ctx context.Context, userID string) error
}

// postgresUserRepository implementa UserRepository usando PostgreSQL + sqlx.
type postgresUserRepository struct {
	db *sqlx.DB
}

// NewUserRepository crea un repositorio PostgreSQL.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.GetContext(ctx, u,
		`SELECT id, name, email, password_hash, role, token_version, created_at
		   FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.GetContext(ctx, u,
		`SELECT id, name, email, password_hash, role, token_version, created_at
		   FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, password_hash, role, token_version, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.TokenVersion)
	return err
}

func (r *postgresUserRepository) UpdateTokenVersion(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET token_version = token_version + 1 WHERE id = $1`, userID)
	return err
}

func (r *postgresUserRepository) StorePasswordResetToken(ctx context.Context, userID, token string, expiry time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET reset_token = $1, reset_token_expires = $2 WHERE id = $3`,
		token, expiry, userID)
	return err
}

func (r *postgresUserRepository) FindByResetToken(ctx context.Context, token string) (*model.User, error) {
	u := &model.User{}
	err := r.db.GetContext(ctx, u,
		`SELECT id, name, email, password_hash, role, token_version, created_at
		   FROM users
		  WHERE reset_token = $1
		    AND reset_token_expires > NOW()`, token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *postgresUserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`, hashedPassword, userID)
	return err
}

func (r *postgresUserRepository) ClearPasswordResetToken(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET reset_token = NULL, reset_token_expires = NULL WHERE id = $1`, userID)
	return err
}
