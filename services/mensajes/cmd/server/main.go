// mensajes-service: mensajería directa entre usuarios.
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	mensajespb "Prueba-Go/gen/mensajes"
	"Prueba-Go/services/mensajes/internal/handler"
	"Prueba-Go/services/mensajes/internal/repository"
	"Prueba-Go/services/mensajes/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	db, err := sqlx.Connect("pgx", requireEnv("DATABASE_URL"))
	if err != nil {
		slog.Error("DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		slog.Error("Migraciones fallidas", "error", err)
		os.Exit(1)
	}

	repo := repository.NewMensajesRepository(db)
	svc := service.NewMensajesService(repo)
	h := handler.NewMensajesHandler(svc)

	port := getEnvOr("GRPC_PORT", "50057")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("net.Listen", "error", err)
		os.Exit(1)
	}
	srv := grpc.NewServer()
	mensajespb.RegisterMensajesServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("mensajes-service iniciado", "port", port)
	if err := srv.Serve(lis); err != nil {
		slog.Error("Serve", "error", err)
		os.Exit(1)
	}
}

func runMigrations(db *sqlx.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS mensajes (
			id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			emisor_id     UUID        NOT NULL,
			emisor_name   TEXT        NOT NULL DEFAULT '',
			receptor_id   UUID        NOT NULL,
			receptor_name TEXT        NOT NULL DEFAULT '',
			contenido     TEXT        NOT NULL,
			leido         BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_mensajes_emisor
			ON mensajes(emisor_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_mensajes_receptor
			ON mensajes(receptor_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_mensajes_noleidos
			ON mensajes(receptor_id) WHERE leido = FALSE`,
	}
	for _, q := range migrations {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}
	return nil
}

func requireEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		slog.Error("variable requerida", "key", k)
		os.Exit(1)
	}
	return v
}

func getEnvOr(k, fb string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fb
}
