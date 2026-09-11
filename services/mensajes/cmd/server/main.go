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
	// contactos resuelve "quién puede escribirle a quién" consultando las
	// tablas de inscripción del servicio de cursos (misma base de datos).
	contactos := repository.NewContactosRepository(db)
	svc := service.NewMensajesService(repo, contactos)
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
			id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			emisor_id       UUID        NOT NULL,
			emisor_name     TEXT        NOT NULL DEFAULT '',
			receptor_id     UUID        NOT NULL,
			receptor_name   TEXT        NOT NULL DEFAULT '',
			contenido       TEXT        NOT NULL DEFAULT '',
			leido           BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			attachment_url  TEXT        NOT NULL DEFAULT '',
			attachment_type TEXT        NOT NULL DEFAULT ''
		)`,
		// Migraciones incrementales para tablas existentes
		`ALTER TABLE mensajes ADD COLUMN IF NOT EXISTS attachment_url  TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mensajes ADD COLUMN IF NOT EXISTS attachment_type TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mensajes ADD COLUMN IF NOT EXISTS is_group        BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE mensajes ALTER COLUMN contenido SET DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_mensajes_emisor
			ON mensajes(emisor_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_mensajes_receptor
			ON mensajes(receptor_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_mensajes_noleidos
			ON mensajes(receptor_id) WHERE leido = FALSE`,
		`CREATE TABLE IF NOT EXISTS grupos (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			nombre     TEXT NOT NULL,
			admin_id   UUID NOT NULL,
			licencia_id UUID DEFAULT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE grupos ADD COLUMN IF NOT EXISTS licencia_id UUID DEFAULT NULL`,
		`CREATE TABLE IF NOT EXISTS grupo_miembros (
			grupo_id   UUID NOT NULL,
			usuario_id UUID NOT NULL,
			joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (grupo_id, usuario_id)
		)`,

		// ── Borrado de mensajes y conversaciones ─────────────────────────
		//
		// Nada se borra de verdad. Una plataforma de capacitación laboral
		// tiene que poder responder a una queja de acoso o a un instructor
		// que niega haber escrito algo, y un DELETE deja la investigación
		// sin nada que consultar. Lo que se guarda es QUIÉN dejó de verlo.
		//
		// "Para todos": se marca la fila. El texto se queda en la base pero
		// el servidor ya no lo manda al cliente.
		`ALTER TABLE mensajes ADD COLUMN IF NOT EXISTS eliminado_at  TIMESTAMPTZ DEFAULT NULL`,
		`ALTER TABLE mensajes ADD COLUMN IF NOT EXISTS eliminado_por UUID        DEFAULT NULL`,

		// "Para mí": una fila por mensaje y persona que lo ocultó. El otro
		// conserva su copia intacta.
		`CREATE TABLE IF NOT EXISTS mensajes_ocultos (
			mensaje_id UUID        NOT NULL,
			usuario_id UUID        NOT NULL,
			oculto_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (mensaje_id, usuario_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_mensajes_ocultos_usuario
			ON mensajes_ocultos(usuario_id)`,

		// Conversación oculta: no se guarda "está borrada" sino DESDE CUÁNDO.
		//
		// Así se comporta igual que WhatsApp sin tener que tocar un solo
		// mensaje: el historial anterior desaparece de tu lista, y en cuanto
		// la otra persona escribe, el chat vuelve con lo nuevo y nada de lo
		// viejo. Marcarla con un booleano obligaría a decidir al recibir cada
		// mensaje si hay que "desborrarla", y a borrar filas de mensajes.
		`CREATE TABLE IF NOT EXISTS conversaciones_ocultas (
			usuario_id   UUID        NOT NULL,
			peer_id      UUID        NOT NULL,
			is_group     BOOLEAN     NOT NULL DEFAULT FALSE,
			oculta_desde TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (usuario_id, peer_id)
		)`,
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
