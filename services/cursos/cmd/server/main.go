// cursos-service: gestiona capacitaciones, inscripciones y asignaciones.
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	cursospb "Prueba-Go/gen/cursos"
	mensajespb "Prueba-Go/gen/mensajes"
	"Prueba-Go/services/cursos/internal/handler"
	"Prueba-Go/services/cursos/internal/repository"
	"Prueba-Go/services/cursos/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// DI
	repo := repository.NewCursosRepository(db)

	// Optional: connect to mensajes-service for cohort group management
	var mensajesClient mensajespb.MensajesServiceClient
	if addr := getEnvOr("MENSAJES_ADDR", ""); addr != "" {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			mensajesClient = mensajespb.NewMensajesServiceClient(conn)
		} else {
			slog.Warn("no se pudo conectar al mensajes-service", "error", err)
		}
	}

	svc := service.NewCursosService(repo, mensajesClient)
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
			precio NUMERIC(10,2) NOT NULL DEFAULT 0.00,
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
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS precio NUMERIC(10,2) NOT NULL DEFAULT 0.00`,
		`ALTER TABLE asignaciones ADD COLUMN IF NOT EXISTS user_name TEXT DEFAULT ''`,
		`ALTER TABLE asignaciones ADD COLUMN IF NOT EXISTS user_email TEXT DEFAULT ''`,
		// Ampliar color de VARCHAR(20) a TEXT para soportar valores de gradiente CSS
		`ALTER TABLE capacitaciones ALTER COLUMN color TYPE TEXT`,
		`CREATE TABLE IF NOT EXISTS curso_licencias (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
			nombre VARCHAR(100) NOT NULL,
			precio NUMERIC(10,2) NOT NULL DEFAULT 0.00,
			capacidad_maxima INT NOT NULL DEFAULT 0,
			usadas INT NOT NULL DEFAULT 0,
			codigo_acceso VARCHAR(50) UNIQUE,
			stripe_product_id VARCHAR(100),
			stripe_price_id VARCHAR(100),
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`ALTER TABLE curso_licencias ADD COLUMN IF NOT EXISTS comprador_id UUID`,
		`CREATE TABLE IF NOT EXISTS inscripciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
			licencia_id UUID REFERENCES curso_licencias(id) ON DELETE SET NULL,
			inscrito_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, capacitacion_id)
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
		`CREATE INDEX IF NOT EXISTS idx_notificaciones_user_id ON notificaciones(user_id)`,
		// Eliminar asignaciones duplicadas conservando solo la más antigua, luego aplicar constraint único
		`DO $$ BEGIN
		   IF NOT EXISTS (
		     SELECT 1 FROM pg_constraint
		     WHERE conrelid='asignaciones'::regclass AND contype='u'
		   ) THEN
		     DELETE FROM asignaciones a USING (
		       SELECT MIN(ctid) as ctid, user_id, capacitacion_id
		       FROM asignaciones 
		       GROUP BY user_id, capacitacion_id HAVING COUNT(*) > 1
		     ) b 
		     WHERE a.user_id = b.user_id AND a.capacitacion_id = b.capacitacion_id AND a.ctid <> b.ctid;
		     
		     ALTER TABLE asignaciones ADD CONSTRAINT asignaciones_user_curso_unique
		       UNIQUE (user_id, capacitacion_id);
		   END IF;
		 END $$`,
		`CREATE TABLE IF NOT EXISTS instructor_schedules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			instructor_id UUID NOT NULL,
			start_time TIMESTAMPTZ NOT NULL,
			end_time TIMESTAMPTZ NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'available',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ`,
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS videocall_status VARCHAR(20) DEFAULT 'pending'`,
		`CREATE TABLE IF NOT EXISTS videocall_tickets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
			licencia_id UUID REFERENCES curso_licencias(id) ON DELETE CASCADE,
			codigo VARCHAR(50) UNIQUE NOT NULL,
			in_use_by_user_id UUID,
			is_valid BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migración fallida: %w", err)
		}
	}
	slog.Info("cursos: migraciones aplicadas")
	return nil
}
