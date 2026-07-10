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
	// ── Tablas base (idempotentes con IF NOT EXISTS) ──────────────────────────
	tables := []string{
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
			segundos_vistos INT NOT NULL DEFAULT 0,
			completado BOOLEAN NOT NULL DEFAULT true,
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

		// ── Feature A: Módulos ───────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS modulos (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			capacitacion_id UUID NOT NULL,
			title       VARCHAR(200) NOT NULL,
			description TEXT DEFAULT '',
			orden       INT NOT NULL DEFAULT 0,
			deleted_at  TIMESTAMPTZ,
			created_at  TIMESTAMPTZ DEFAULT NOW()
		)`,

		// ── Feature A: Submódulos ────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS submodulos (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			modulo_id  UUID NOT NULL REFERENCES modulos(id) ON DELETE CASCADE,
			title      VARCHAR(200) NOT NULL,
			description TEXT DEFAULT '',
			orden      INT NOT NULL DEFAULT 0,
			deleted_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		// ── Feature B: Puntos por minijuego ──────────────────────────────────
		// Un registro por intento; para el leaderboard se agrupa con SUM.
		`CREATE TABLE IF NOT EXISTS game_scores (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     UUID NOT NULL,
			leccion_id  UUID NOT NULL REFERENCES lecciones(id) ON DELETE CASCADE,
			capacitacion_id UUID NOT NULL,
			points      INT NOT NULL DEFAULT 0,
			time_secs   INT NOT NULL DEFAULT 0,
			scored_at   TIMESTAMPTZ DEFAULT NOW()
		)`,

		// ── Feature B: Insignias ─────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS user_badges (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     UUID NOT NULL,
			badge_slug  VARCHAR(80) NOT NULL,
			unlocked_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, badge_slug)
		)`,

		// ── Feature B: Puntos globales del usuario (tabla propia del microservicio)
		`CREATE TABLE IF NOT EXISTS user_points (
			user_id      UUID PRIMARY KEY,
			points_total INT NOT NULL DEFAULT 0,
			updated_at   TIMESTAMPTZ DEFAULT NOW()
		)`,
	}

	for _, s := range tables {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración (create table) fallida: %w", err)
		}
	}

	// ── ALTER TABLE: columnas nuevas sobre tablas existentes ──────────────────
	alters := []string{
		// Soft-delete y gamificación en lecciones
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`,

		// Feature A: FK jerárquicas en lecciones (nullable → lección "suelta" si ambas son NULL)
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS modulo_id    UUID REFERENCES modulos(id)    ON DELETE SET NULL`,
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS submodulo_id UUID REFERENCES submodulos(id) ON DELETE SET NULL`,

		// Feature B: tipo de lección y configuración del minijuego
		// Cambiamos type de VARCHAR(20) a TEXT para acomodar los nuevos slugs del enum
		`ALTER TABLE lecciones ALTER COLUMN type TYPE TEXT`,
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS game_config_json TEXT DEFAULT ''`,
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS points_reward    INT  NOT NULL DEFAULT 0`,
		`ALTER TABLE progreso_lecciones ADD COLUMN IF NOT EXISTS segundos_vistos INT NOT NULL DEFAULT 0`,
		`ALTER TABLE progreso_lecciones ADD COLUMN IF NOT EXISTS completado BOOLEAN NOT NULL DEFAULT true`,
	}

	for _, s := range alters {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración (alter) fallida: %w", err)
		}
	}

	// ── Índices ───────────────────────────────────────────────────────────────
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_modulos_capacitacion    ON modulos(capacitacion_id)`,
		`CREATE INDEX IF NOT EXISTS idx_submodulos_modulo       ON submodulos(modulo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_lecciones_modulo        ON lecciones(modulo_id)       WHERE modulo_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_lecciones_submodulo     ON lecciones(submodulo_id)    WHERE submodulo_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_user        ON game_scores(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_leccion     ON game_scores(leccion_id)`,
		`CREATE INDEX IF NOT EXISTS idx_game_scores_capacitacion ON game_scores(capacitacion_id)`,
		// Índice de leaderboard: por curso → suma de puntos descendente
		`CREATE INDEX IF NOT EXISTS idx_game_scores_lb          ON game_scores(capacitacion_id, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_badges_user        ON user_badges(user_id)`,
	}

	for _, s := range indexes {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración (index) fallida: %w", err)
		}
	}

	slog.Info("lecciones: migraciones aplicadas")
	return nil
}
