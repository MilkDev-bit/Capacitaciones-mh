package db

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"Prueba-Go/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB es la conexión global a PostgreSQL expuesta como *sqlx.DB.
// sqlx.DB extiende database/sql.DB, por lo que todo el código existente
// (QueryRowContext, ExecContext, etc.) sigue funcionando sin cambios.
// Los nuevos handlers pueden aprovechar db.DB.GetContext / db.DB.SelectContext
// para mapear resultados directamente a structs con tags db:"...".
var DB *sqlx.DB

func Connect() {
	var dsn string

	if url := config.C.DatabaseURL; url != "" {
		dsn = url
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.C.DBHost, config.C.DBPort, config.C.DBUser,
			config.C.DBPassword, config.C.DBName,
		)
	}

	var err error
	DB, err = sqlx.Open("postgres", dsn)
	if err != nil {
		slog.Error("Error abriendo base de datos", "error", err)
		os.Exit(1)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	for i := 1; i <= 10; i++ {
		if err = DB.Ping(); err == nil {
			slog.Info("Conexión a PostgreSQL exitosa")
			return
		}
		slog.Warn("DB no disponible", "intento", i, "error", err)
		time.Sleep(3 * time.Second)
	}
	slog.Error("No se pudo conectar a la base de datos tras 10 intentos", "error", err)
	os.Exit(1)
}
