// lecciones-service: gestiona lecciones, progreso y preguntas intermedias.
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	leccionespb "Prueba-Go/gen/lecciones"
	"Prueba-Go/services/lecciones/internal/handler"
	"Prueba-Go/services/lecciones/internal/repository"
	"Prueba-Go/services/lecciones/internal/service"

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

	repo := repository.NewLeccionesRepository(db)
	svc := service.NewLeccionesService(repo)
	h := handler.NewLeccionesHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50054"))
	srv := grpc.NewServer()
	leccionespb.RegisterLeccionesServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("lecciones-service iniciado", "port", getEnvOr("GRPC_PORT", "50054"))
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
		`CREATE TABLE IF NOT EXISTS lecciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			capacitacion_id UUID NOT NULL,
			title VARCHAR(200) NOT NULL,
			description TEXT DEFAULT '',
			type VARCHAR(20) NOT NULL DEFAULT 'video',
			file_path TEXT DEFAULT '',
			content TEXT DEFAULT '',
			orden INT NOT NULL DEFAULT 0,
			duracion_min INT DEFAULT 0,
			deleted_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS progreso_lecciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			leccion_id UUID NOT NULL REFERENCES lecciones(id) ON DELETE CASCADE,
			completado_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, leccion_id)
		)`,
		`CREATE TABLE IF NOT EXISTS preguntas_intermedias (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			capacitacion_id UUID NOT NULL,
			despues_de_leccion_id UUID REFERENCES lecciones(id) ON DELETE SET NULL,
			texto TEXT NOT NULL,
			tipo VARCHAR(30) NOT NULL DEFAULT 'multiple_choice',
			orden INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS opciones_intermedias (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			pregunta_id UUID NOT NULL REFERENCES preguntas_intermedias(id) ON DELETE CASCADE,
			texto TEXT NOT NULL,
			es_correcta BOOLEAN NOT NULL DEFAULT false
		)`,
		`CREATE TABLE IF NOT EXISTS respuestas_intermedias (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			capacitacion_id UUID NOT NULL,
			pregunta_id UUID NOT NULL REFERENCES preguntas_intermedias(id) ON DELETE CASCADE,
			opcion_id UUID REFERENCES opciones_intermedias(id) ON DELETE SET NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, pregunta_id)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración fallida: %w", err)
		}
	}
	// Columnas que pueden faltar en BDs existentes
	alters := []string{
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`,
	}
	for _, s := range alters {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración alter fallida: %w", err)
		}
	}
	slog.Info("lecciones: migraciones aplicadas")
	return nil
}
