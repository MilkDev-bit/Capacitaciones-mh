package model

import "time"

// User es el modelo interno del auth service.
// Solo contiene los campos que este servicio necesita.
type User struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	TokenVersion int       `db:"token_version"`
	CreatedAt    time.Time `db:"created_at"`
}
