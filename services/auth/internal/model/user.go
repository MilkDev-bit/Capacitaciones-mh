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
	// AvisoVersion es la versión del aviso de privacidad que aceptó. Vacía en
	// las cuentas anteriores a que existiera el registro de consentimiento.
	AvisoVersion string `db:"aviso_version"`
}

// EmailVerification agrupa el estado del código de 6 dígitos de un usuario.
// Se consulta aparte de User porque solo interesa durante el alta de cuenta.
// PasswordOTP es el código de un solo uso para cambiar la contraseña desde el
// perfil, con su expiración y los intentos ya gastados.
type PasswordOTP struct {
	UserID   string     `db:"id"`
	Hash     *string    `db:"pwd_otp_hash"`
	Expira   *time.Time `db:"pwd_otp_expira"`
	Intentos int        `db:"pwd_otp_intentos"`
}

type EmailVerification struct {
	UserID   string     `db:"id"`
	Hash     *string    `db:"email_verification_hash"`
	Expires  *time.Time `db:"email_verification_expires"`
	Attempts int        `db:"email_verification_attempts"`
	SentAt   *time.Time `db:"email_verification_sent_at"`
}
