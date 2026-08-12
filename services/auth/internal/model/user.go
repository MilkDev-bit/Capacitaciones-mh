package model

import "time"

// User es el modelo interno del auth service.
// Solo contiene los campos que este servicio necesita.
type User struct {
	ID            string    `db:"id"`
	Name          string    `db:"name"`
	Email         string    `db:"email"`
	PasswordHash  string    `db:"password_hash"`
	Role          string    `db:"role"`
	TokenVersion  int       `db:"token_version"`
	EmailVerified bool      `db:"email_verified"`
	CreatedAt     time.Time `db:"created_at"`
}

// EmailVerification agrupa el estado del código de 6 dígitos de un usuario.
// Se consulta aparte de User porque solo interesa durante el alta de cuenta.
type EmailVerification struct {
	UserID   string     `db:"id"`
	Hash     *string    `db:"email_verification_hash"`
	Expires  *time.Time `db:"email_verification_expires"`
	Attempts int        `db:"email_verification_attempts"`
	SentAt   *time.Time `db:"email_verification_sent_at"`
}
