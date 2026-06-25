// foros-service: gestiona posts, comentarios y likes de foros por lección.
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	forospb "Prueba-Go/gen/foros"
	"Prueba-Go/services/foros/internal/handler"
	"Prueba-Go/services/foros/internal/repository"
	"Prueba-Go/services/foros/internal/service"

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

	repo := repository.NewForosRepository(db)
	svc := service.NewForosService(repo)
	h := handler.NewForosHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50056"))
	srv := grpc.NewServer()
	forospb.RegisterForosServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("foros-service iniciado", "port", getEnvOr("GRPC_PORT", "50056"))
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
		`CREATE TABLE IF NOT EXISTS foro_posts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			leccion_id UUID NOT NULL,
			user_id UUID NOT NULL,
			user_name TEXT NOT NULL DEFAULT '',
			titulo TEXT NOT NULL,
			contenido TEXT NOT NULL,
			media_url TEXT DEFAULT '',
			media_type TEXT DEFAULT '',
			deleted_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS foro_likes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			post_id UUID NOT NULL REFERENCES foro_posts(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			UNIQUE(post_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS foro_comentarios (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			post_id UUID NOT NULL REFERENCES foro_posts(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			user_name TEXT NOT NULL DEFAULT '',
			contenido TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		// Columnas que pueden faltar en BDs existentes
		`ALTER TABLE foro_posts ADD COLUMN IF NOT EXISTS user_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE foro_comentarios ADD COLUMN IF NOT EXISTS user_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE foro_posts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`,
		`ALTER TABLE foro_comentarios ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES foro_comentarios(id) ON DELETE CASCADE`,
		`CREATE TABLE IF NOT EXISTS foro_post_reactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			post_id UUID NOT NULL REFERENCES foro_posts(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			emoji VARCHAR(20) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(post_id, user_id, emoji)
		)`,
		`CREATE TABLE IF NOT EXISTS foro_comentario_reactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			comentario_id UUID NOT NULL REFERENCES foro_comentarios(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			emoji VARCHAR(20) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(comentario_id, user_id, emoji)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración fallida: %w", err)
		}
	}
	slog.Info("foros: migraciones aplicadas")
	return nil
}
