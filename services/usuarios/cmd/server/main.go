package main

import (
	"log/slog"
	"net"
	"os"

	usuariospb "Prueba-Go/gen/usuarios"
	"Prueba-Go/services/usuarios/internal/handler"
	"Prueba-Go/services/usuarios/internal/repository"
	"Prueba-Go/services/usuarios/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	dbURL := requireEnv("DATABASE_URL")
	grpcPort := getEnvOr("GRPC_PORT", "50052")

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		slog.Error("DB connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Garantizar que las tablas requeridas existan
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'alumno',
			bio TEXT NOT NULL DEFAULT '',
			avatar_url TEXT NOT NULL DEFAULT '',
			cover_url TEXT NOT NULL DEFAULT '',
			phone VARCHAR(50) NOT NULL DEFAULT '',
			specialty VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS notificaciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			tipo VARCHAR(50) NOT NULL,
			titulo VARCHAR(200) NOT NULL,
			mensaje TEXT NOT NULL,
			leida BOOLEAN NOT NULL DEFAULT false,
			enlace TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`ALTER TABLE notificaciones ADD COLUMN IF NOT EXISTS enlace TEXT`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS points_total INT NOT NULL DEFAULT 0`,
		`ALTER TABLE inscripciones ADD COLUMN IF NOT EXISTS licencia_id UUID`,
		`CREATE INDEX IF NOT EXISTS idx_notificaciones_user_id ON notificaciones(user_id)`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			slog.Error("migración fallida", "error", err, "sql", m[:60])
			os.Exit(1)
		}
	}

	// DI: Repository → Service → Handler
	repo := repository.NewUsuarioRepository(db)
	svc := service.NewUsuariosService(repo)
	h := handler.NewUsuariosHandler(svc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("net.Listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	usuariospb.RegisterUsuariosServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("usuarios-service iniciado", "port", grpcPort)
	if err := srv.Serve(lis); err != nil {
		slog.Error("gRPC Serve", "error", err)
		os.Exit(1)
	}
}

func requireEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		slog.Error("variable de entorno requerida", "key", k)
		os.Exit(1)
	}
	return v
}

func getEnvOr(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
