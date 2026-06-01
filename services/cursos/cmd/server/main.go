// cursos-service: gestiona capacitaciones, inscripciones y asignaciones.
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/handler"
	"Prueba-Go/services/cursos/internal/repository"
	"Prueba-Go/services/cursos/internal/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	db, err := sqlx.Connect("postgres", requireEnv("DATABASE_URL"))
	if err != nil {
		slog.Error("DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		slog.Error("Migraciones fallidas", "error", err)
		os.Exit(1)
	}

	// DI
	repo := repository.NewCursosRepository(db)
	svc := service.NewCursosService(repo)
	h := handler.NewCursosHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50053"))
	srv := grpc.NewServer()

	cursospb.RegisterCursosServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("cursos-service iniciado", "port", getEnvOr("GRPC_PORT", "50053"))
	if err := srv.Serve(lis); err != nil {
		slog.Error("Serve", "error", err)
		os.Exit(1)
	}
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

func runMigrations(db *sqlx.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS capacitaciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(200) NOT NULL,
			description TEXT DEFAULT '',
			type VARCHAR(20) NOT NULL DEFAULT 'document',
			file_path TEXT DEFAULT '',
			content TEXT DEFAULT '',
			instructor_id UUID,
			is_public BOOLEAN NOT NULL DEFAULT false,
			codigo_acceso VARCHAR(12) UNIQUE,
			welcome_message TEXT DEFAULT '',
			thumbnail_url TEXT DEFAULT '',
			color TEXT DEFAULT '#f97316',
			deleted_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS asignaciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			user_name TEXT DEFAULT '',
			user_email TEXT DEFAULT '',
			capacitacion_id UUID,
			assigned_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, capacitacion_id)
		)`,
		// Columnas que pueden faltar en BDs existentes
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`,
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS welcome_message TEXT DEFAULT ''`,
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS thumbnail_url TEXT DEFAULT ''`,
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS color TEXT DEFAULT '#f97316'`,
		`ALTER TABLE asignaciones ADD COLUMN IF NOT EXISTS user_name TEXT DEFAULT ''`,
		`ALTER TABLE asignaciones ADD COLUMN IF NOT EXISTS user_email TEXT DEFAULT ''`,
		// Ampliar color de VARCHAR(20) a TEXT para soportar valores de gradiente CSS
		`ALTER TABLE capacitaciones ALTER COLUMN color TYPE TEXT`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración fallida: %w", err)
		}
	}
	slog.Info("cursos: migraciones aplicadas")
	return nil
}
