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

	// ── Verificación de correo ────────────────────────────────────────────────
	// StoreEmailVerification guarda el hash del código y su expiración,
	// reiniciando el contador de intentos.
	StoreEmailVerification(ctx context.Context, userID, codeHash string, expiry time.Time) error
	// GetEmailVerification devuelve el estado del código de un usuario.
	GetEmailVerification(ctx context.Context, userID string) (*model.EmailVerification, error)
	// IncrementVerificationAttempts suma 1 al contador de intentos fallidos.
	IncrementVerificationAttempts(ctx context.Context, userID string) error
	// MarkEmailVerified marca la cuenta como verificada y limpia el código.
	MarkEmailVerified(ctx context.Context, userID string) error
}

// postgresUserRepository implementa UserRepository usando PostgreSQL + sqlx.
type postgresUserRepository struct {
	db *sqlx.DB
}

// NewUserRepository crea un repositorio PostgreSQL.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

// userColumns es la proyección compartida por FindByEmail/FindByID/FindByResetToken.
// Centralizarla evita que una columna nueva se olvide en una de las tres consultas.
const userColumns = `id, name, email, password_hash, role, token_version,
	COALESCE(email_verified, true) AS email_verified, created_at`

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.GetContext(ctx, u,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.GetContext(ctx, u,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, password_hash, role, token_version, email_verified, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.TokenVersion, u.EmailVerified)
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
		`SELECT `+userColumns+`
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

// ── Verificación de correo ────────────────────────────────────────────────────

func (r *postgresUserRepository) StoreEmailVerification(ctx context.Context, userID, codeHash string, expiry time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users
		    SET email_verification_hash     = $1,
		        email_verification_expires  = $2,
		        email_verification_attempts = 0,
		        email_verification_sent_at  = NOW()
		  WHERE id = $3`, codeHash, expiry, userID)
	return err
}

func (r *postgresUserRepository) GetEmailVerification(ctx context.Context, userID string) (*model.EmailVerification, error) {
	v := &model.EmailVerification{}
	err := r.db.GetContext(ctx, v,
		`SELECT id,
		        email_verification_hash,
		        email_verification_expires,
		        COALESCE(email_verification_attempts, 0) AS email_verification_attempts,
		        email_verification_sent_at
		   FROM users WHERE id = $1`, userID)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *postgresUserRepository) IncrementVerificationAttempts(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email_verification_attempts = COALESCE(email_verification_attempts, 0) + 1
		  WHERE id = $1`, userID)
	return err
}

func (r *postgresUserRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users
		    SET email_verified              = true,
		        email_verification_hash     = NULL,
		        email_verification_expires  = NULL,
		        email_verification_attempts = 0
		  WHERE id = $1`, userID)
	return err
}
