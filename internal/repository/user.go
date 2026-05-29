// Package repository contiene las implementaciones de acceso a datos.
// Es el ÚNICO lugar del proyecto que sabe que existe PostgreSQL.
// Los handlers y servicios dependen de interfaces, no de implementaciones concretas,
// lo que permite hacer mocking en tests sin necesitar una base de datos levantada.
package repository

import (
	"context"
	"database/sql"

	"Prueba-Go/internal/models"

	"github.com/jmoiron/sqlx"
)

// UserRepository define las operaciones de persistencia para usuarios.
// Cualquier handler o servicio debe depender de esta interfaz, nunca de la
// implementación concreta postgresUserRepository.
type UserRepository interface {
	// FindByEmail devuelve el usuario con ese email o sql.ErrNoRows si no existe.
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// FindByID devuelve el usuario con ese ID o sql.ErrNoRows si no existe.
	FindByID(ctx context.Context, id string) (*models.User, error)
	// Create inserta un usuario nuevo y devuelve su UUID generado.
	Create(ctx context.Context, name, email, passwordHash, role string) (string, error)
	// UpdateTokenVersion incrementa token_version invalidando todos los JWT activos.
	UpdateTokenVersion(ctx context.Context, id string) error
}

// postgresUserRepository es la implementación real con PostgreSQL.
type postgresUserRepository struct {
	db *sqlx.DB
}

// NewUserRepository construye el repositorio con la conexión sqlx proporcionada.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, name, email, password_hash, role, token_version
		   FROM users
		  WHERE email = $1`,
		email,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, name, email, password_hash, role, token_version
		   FROM users
		  WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, name, email, passwordHash, role string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users(name, email, password_hash, role)
		 VALUES($1, $2, $3, $4)
		 RETURNING id`,
		name, email, passwordHash, role,
	).Scan(&id)
	return id, err
}

func (r *postgresUserRepository) UpdateTokenVersion(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET token_version = token_version + 1 WHERE id = $1`,
		id,
	)
	return err
}

// Compile-time check: postgresUserRepository implementa UserRepository.
var _ UserRepository = (*postgresUserRepository)(nil)

// ErrNotFound es un alias semántico exportado para que los servicios puedan
// distinguir "no encontrado" sin importar database/sql directamente.
var ErrNotFound = sql.ErrNoRows
