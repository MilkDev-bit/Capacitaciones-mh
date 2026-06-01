// Package db provee la inicialización de la base de datos de mensajes directos.
// El gateway se conecta a la misma instancia PostgreSQL que los microservicios
// para gestionar mensajes sin necesitar un servicio gRPC adicional.
package db

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenMensajes abre la conexión y ejecuta las migraciones de la tabla mensajes.
func OpenMensajes(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(context.Background()); err != nil {
		return nil, err
	}
	if err := migrateMensajes(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateMensajes(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS mensajes (
			id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			emisor_id    UUID        NOT NULL,
			emisor_name  TEXT        NOT NULL DEFAULT '',
			receptor_id  UUID        NOT NULL,
			receptor_name TEXT       NOT NULL DEFAULT '',
			contenido    TEXT        NOT NULL,
			leido        BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_mensajes_emisor
			ON mensajes(emisor_id,   created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_mensajes_receptor
			ON mensajes(receptor_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_mensajes_noleidos
			ON mensajes(receptor_id) WHERE leido = FALSE;
	`)
	return err
}
