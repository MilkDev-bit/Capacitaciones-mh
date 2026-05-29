// examenes-service: gestiona exámenes, preguntas, envío y calificaciones.
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	examenespb "Prueba-Go/gen/examenes"
	"Prueba-Go/services/examenes/internal/handler"
	"Prueba-Go/services/examenes/internal/repository"
	"Prueba-Go/services/examenes/internal/service"

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

	repo := repository.NewExamenesRepository(db)
	svc := service.NewExamenesService(repo)
	h := handler.NewExamenesHandler(svc)

	lis, _ := net.Listen("tcp", ":"+getEnvOr("GRPC_PORT", "50055"))
	srv := grpc.NewServer()
	examenespb.RegisterExamenesServiceServer(srv, h)
	reflection.Register(srv)

	slog.Info("examenes-service iniciado", "port", getEnvOr("GRPC_PORT", "50055"))
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
		`CREATE TABLE IF NOT EXISTS examenes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(200) NOT NULL,
			description TEXT DEFAULT '',
			instructor_id UUID,
			capacitacion_id UUID,
			deleted_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS preguntas (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			examen_id UUID NOT NULL REFERENCES examenes(id) ON DELETE CASCADE,
			texto TEXT NOT NULL,
			tipo VARCHAR(30) NOT NULL DEFAULT 'multiple_choice',
			valor NUMERIC(5,2) NOT NULL DEFAULT 1,
			orden INT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS opciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			pregunta_id UUID NOT NULL REFERENCES preguntas(id) ON DELETE CASCADE,
			texto TEXT NOT NULL,
			es_correcta BOOLEAN NOT NULL DEFAULT false
		)`,
		`CREATE TABLE IF NOT EXISTS respuestas_examen (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			examen_id UUID NOT NULL REFERENCES examenes(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			user_name TEXT NOT NULL DEFAULT '',
			pregunta_id UUID NOT NULL REFERENCES preguntas(id) ON DELETE CASCADE,
			opcion_id UUID REFERENCES opciones(id) ON DELETE SET NULL,
			respuesta_texto TEXT,
			respondido_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, pregunta_id)
		)`,
		`CREATE TABLE IF NOT EXISTS asignaciones_examen (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			examen_id UUID NOT NULL REFERENCES examenes(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			assigned_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(examen_id, user_id)
		)`,
		// Columnas que pueden faltar en BDs existentes
		`ALTER TABLE examenes ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`,
		`ALTER TABLE respuestas_examen ADD COLUMN IF NOT EXISTS user_name TEXT NOT NULL DEFAULT ''`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración fallida: %w", err)
		}
	}
	slog.Info("examenes: migraciones aplicadas")
	return nil
}
