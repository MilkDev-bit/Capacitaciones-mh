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
	slog.Info("cursos: iniciando migraciones...")
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
		`CREATE TABLE IF NOT EXISTS lecciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
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
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS duration INT NOT NULL DEFAULT 0`,
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS dc3_enabled BOOLEAN NOT NULL DEFAULT true`,
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
		`ALTER TABLE curso_licencias ADD COLUMN IF NOT EXISTS stripe_product_id VARCHAR(100)`,
		`ALTER TABLE curso_licencias ADD COLUMN IF NOT EXISTS stripe_price_id VARCHAR(100)`,
		`ALTER TABLE curso_licencias ADD COLUMN IF NOT EXISTS comprador_id UUID`,
		`CREATE TABLE IF NOT EXISTS inscripciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
			licencia_id UUID REFERENCES curso_licencias(id) ON DELETE SET NULL,
			inscrito_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, capacitacion_id)
		)`,
		`ALTER TABLE inscripciones ADD COLUMN IF NOT EXISTS licencia_id UUID REFERENCES curso_licencias(id) ON DELETE SET NULL`,
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
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ`,
		// Las videollamadas se retiraron: los cursos que quedaron con ese tipo se
		// despublican para que no aparezcan en la tienda ni en el catálogo.
		`UPDATE capacitaciones SET is_public = false WHERE type = 'videocall' AND is_public = true`,
		// Reparto de accesos corporativos: una fila por participante invitado.
		// La crea este servicio porque es quien la lee y escribe.
		`CREATE TABLE IF NOT EXISTS licencia_invitaciones (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			licencia_id UUID NOT NULL REFERENCES curso_licencias(id) ON DELETE CASCADE,
			nombre VARCHAR(120) NOT NULL DEFAULT '',
			email VARCHAR(200) NOT NULL,
			codigo VARCHAR(50) NOT NULL,
			estado VARCHAR(20) NOT NULL DEFAULT 'enviado',
			enviado_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(licencia_id, email)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_licencia_invitaciones_licencia_id ON licencia_invitaciones(licencia_id)`,
		// Despliegues previos crearon ticket_id apuntando a videocall_tickets.
		// Se suelta la columna para poder retirar esa tabla sin romper la FK.
		`ALTER TABLE licencia_invitaciones DROP COLUMN IF EXISTS ticket_id`,
		// Aviso DC-3: una fila por (licencia, curso). La PK compuesta deduplica el
		// correo al representante aunque terminen todos los participantes.
		`CREATE TABLE IF NOT EXISTS dc3_avisos (
			licencia_id UUID NOT NULL REFERENCES curso_licencias(id) ON DELETE CASCADE,
			capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
			enviado_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (licencia_id, capacitacion_id)
		)`,
		// ── Constancias DC-3 ────────────────────────────────────────────────
		//
		// El área temática es del CURSO: "trabajos en altura" y "primeros
		// auxilios" no comparten clave del catálogo STPS.
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS dc3_area_tematica VARCHAR(100)`,
		// Nombre oficial para la constancia; distinto del título comercial.
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS dc3_nombre_curso VARCHAR(250)`,
		// Duración oficial en horas que declara la constancia.
		//
		// No se deriva de `duration`. Esa columna son minutos de contenido y
		// además el editor de cursos nunca la envía, así que valía 0 en todas
		// las capacitaciones: la constancia se quedaba sin "duración en horas" y
		// no llegaba a generarse nunca. Aunque tuviera valor, tampoco serviría:
		// 95 minutos de vídeo pueden ser 8 horas oficiales de capacitación.
		`ALTER TABLE capacitaciones ADD COLUMN IF NOT EXISTS dc3_duracion_horas INT NOT NULL DEFAULT 0`,

		// Los datos de empresa vivieron un momento aquí, por capacitación. Se
		// retiran: el patrón es de quien emplea al trabajador, no del curso.
		// Ahora los declara el alumno y el instructor pone el respaldo.
		//
		// El DROP es seguro porque estas columnas nunca llegaron a escribirse:
		// el INSERT y el UPDATE de capacitaciones jamás las incluyeron.
		`ALTER TABLE capacitaciones DROP COLUMN IF EXISTS dc3_razon_social`,
		`ALTER TABLE capacitaciones DROP COLUMN IF EXISTS dc3_rfc`,
		`ALTER TABLE capacitaciones DROP COLUMN IF EXISTS dc3_nombre_patron`,
		`ALTER TABLE capacitaciones DROP COLUMN IF EXISTS dc3_representante_trabajadores`,
		`ALTER TABLE capacitaciones DROP COLUMN IF EXISTS dc3_nombre_capacitador`,
		`ALTER TABLE capacitaciones DROP COLUMN IF EXISTS dc3_logo_base64`,

		// Empresa por defecto del instructor. Es el respaldo para el alumno que
		// no declara patrón propio: un particular que se capacita por su cuenta
		// recibe la constancia a nombre de quien la imparte.
		// La columna nació como logo_base64 y pasó a URL de R2 antes de tener datos.
		`ALTER TABLE dc3_empresa_instructor DROP COLUMN IF EXISTS logo_base64`,
		`ALTER TABLE dc3_empresa_instructor ADD COLUMN IF NOT EXISTS logo_url TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS dc3_empresa_instructor (
			instructor_id UUID PRIMARY KEY,
			razon_social VARCHAR(200) NOT NULL DEFAULT '',
			rfc VARCHAR(20) NOT NULL DEFAULT '',
			nombre_patron VARCHAR(200) NOT NULL DEFAULT '',
			representante_trabajadores VARCHAR(200) NOT NULL DEFAULT '',
			nombre_capacitador VARCHAR(200) NOT NULL DEFAULT '',
			logo_url TEXT NOT NULL DEFAULT '',
			actualizado_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Datos del trabajador: los captura el alumno la primera vez y se
		// reutilizan en todas sus constancias. Uno por usuario, de ahí la PK.
		//
		// Sin FK contra users: esa tabla vive en la base de auth/usuarios.
		`CREATE TABLE IF NOT EXISTS dc3_datos_trabajador (
			user_id UUID PRIMARY KEY,
			curp VARCHAR(18) NOT NULL,
			puesto VARCHAR(150) NOT NULL,
			ocupacion_especifica VARCHAR(150) NOT NULL,
			actualizado_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// Patrón del alumno. Opcional en bloque: o están los cuatro o ninguno.
		// Mezclar la razón social del alumno con el representante del instructor
		// daría un documento que no corresponde a ninguna empresa real.
		`ALTER TABLE dc3_datos_trabajador ADD COLUMN IF NOT EXISTS razon_social VARCHAR(200) NOT NULL DEFAULT ''`,
		`ALTER TABLE dc3_datos_trabajador ADD COLUMN IF NOT EXISTS rfc VARCHAR(20) NOT NULL DEFAULT ''`,
		`ALTER TABLE dc3_datos_trabajador ADD COLUMN IF NOT EXISTS nombre_patron VARCHAR(200) NOT NULL DEFAULT ''`,
		`ALTER TABLE dc3_datos_trabajador ADD COLUMN IF NOT EXISTS representante_trabajadores VARCHAR(200) NOT NULL DEFAULT ''`,
		// Logotipo de la empresa del alumno, para el lado izquierdo del documento.
		//
		// Solo se usa si declara empresa propia: ese lado corresponde al patrón.
		// Quien no la declara recibe el del instructor en ambos lados.
		`ALTER TABLE dc3_datos_trabajador ADD COLUMN IF NOT EXISTS logo_url TEXT NOT NULL DEFAULT ''`,

		// Constancia emitida. La PK compuesta es lo que hace idempotente la
		// generación automática: si el alumno vuelve a completar el curso o el
		// webhook se repite, se reemplaza la fila en vez de duplicar el archivo.
		`CREATE TABLE IF NOT EXISTS dc3_constancias (
			user_id UUID NOT NULL,
			capacitacion_id UUID NOT NULL REFERENCES capacitaciones(id) ON DELETE CASCADE,
			archivo_url TEXT NOT NULL,
			generada_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (user_id, capacitacion_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_dc3_constancias_user ON dc3_constancias(user_id, generada_at DESC)`,
		// Folio impreso en el documento, para la verificación pública.
		//
		// Nullable a propósito: las constancias emitidas antes de esto no lo
		// tienen y no se pueden inventar hacia atrás sin reemitirlas.
		`ALTER TABLE dc3_constancias ADD COLUMN IF NOT EXISTS folio VARCHAR(32)`,
		// Nombre y empresa TAL COMO SALIERON IMPRESOS.
		//
		// Se copian en vez de consultarse al verificar. Por un lado, la tabla de
		// usuarios vive en la base de auth y este servicio tiene la suya, así que
		// un JOIN no resolvería. Por otro, aunque resolviera sería incorrecto: si
		// el alumno cambia de empleo, la verificación debe seguir confirmando lo
		// que dice el papel que tiene delante quien lo está comprobando.
		`ALTER TABLE dc3_constancias ADD COLUMN IF NOT EXISTS nombre_trabajador VARCHAR(250)`,
		`ALTER TABLE dc3_constancias ADD COLUMN IF NOT EXISTS razon_social VARCHAR(250)`,
		// El índice es UNIQUE porque el folio es la identidad del documento: dos
		// constancias con el mismo código harían ambiguo el resultado de la
		// verificación. Parcial, para que las filas antiguas sin folio no choquen
		// entre ellas —en Postgres los NULL no colisionan, pero dejarlo explícito
		// evita sorpresas si algún día se rellenan con cadena vacía—.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_dc3_constancias_folio
		   ON dc3_constancias(folio) WHERE folio IS NOT NULL AND folio <> ''`,

		`ALTER TABLE curso_licencias ADD COLUMN IF NOT EXISTS curso_type VARCHAR(20)`,
		`ALTER TABLE curso_licencias ADD COLUMN IF NOT EXISTS curso_duracion INT`,
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS fecha_inicio TIMESTAMPTZ`,
		`ALTER TABLE lecciones ADD COLUMN IF NOT EXISTS fecha_cierre TIMESTAMPTZ`,
		`CREATE TABLE IF NOT EXISTS entregas_actividad (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			leccion_id UUID NOT NULL REFERENCES lecciones(id) ON DELETE CASCADE,
			capacitacion_id UUID NOT NULL,
			user_id UUID NOT NULL,
			file_path TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_size BIGINT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(leccion_id, user_id)
		)`,

		// ── Comisiones de Stripe ──────────────────────────────────────────
		//
		// Lo que Stripe se queda por cada cobro no se guardaba en ninguna
		// parte. El panel enseñaba una "venta neta" calculada con 3.6% + 3 MXN,
		// que es la tarifa de lista de tarjeta nacional y no lo que Stripe cobra
		// de verdad: OXXO, tarjeta internacional y MSI tienen tarifas distintas,
		// y falta el IVA. La cifra buena viene en el BalanceTransaction del
		// cobro, que trae `fee` y `net` al centavo.
		//
		// NULL y 0 no son lo mismo: NULL es "aún no consultado a Stripe" y 0 es
		// "consultado, y no hubo comisión". Tratarlos igual haría que el
		// histórico sin rellenar pareciera libre de comisiones.
		// Clientes de Stripe, necesarios para la transferencia bancaria (SPEI):
		// la CLABE de referencia se emite a nombre de un cliente que ya debe
		// existir al crear la sesión. Se guarda para REUTILIZARLO; crear uno por
		// compra repartiría los saldos entre varios clientes y haría imposible
		// conciliar las transferencias.
		//
		// Sin FK a users: vive en la base de auth, no en esta.
		`CREATE TABLE IF NOT EXISTS stripe_clientes (
			user_id            UUID PRIMARY KEY,
			stripe_customer_id VARCHAR(255) NOT NULL UNIQUE,
			created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Envuelto en un DO condicional y no un ALTER a secas: `ordenes` y
		// `suscripcion_facturas` las crea migrations/002, que se aplica a mano.
		// Un ALTER sobre una tabla que todavía no existe aborta las migraciones
		// y deja el servicio sin arrancar; así, si no está, no se hace nada y
		// las columnas las pondrá 003 junto con la tabla.
		`DO $$
		BEGIN
			IF to_regclass('public.ordenes') IS NOT NULL THEN
				ALTER TABLE ordenes ADD COLUMN IF NOT EXISTS comision_centavos BIGINT;
				ALTER TABLE ordenes ADD COLUMN IF NOT EXISTS neto_centavos BIGINT;
				ALTER TABLE ordenes ADD COLUMN IF NOT EXISTS balance_transaction_id VARCHAR(255);

				-- Índice parcial: en cuanto una orden tiene su comisión deja de
				-- aparecer, así que se mantiene pequeño para siempre.
				CREATE INDEX IF NOT EXISTS idx_ordenes_sin_comision
				    ON ordenes (pagada_at)
				    WHERE comision_centavos IS NULL
				      AND estado IN ('pagada', 'cumplida')
				      AND stripe_payment_intent IS NOT NULL;
			END IF;

			IF to_regclass('public.suscripcion_facturas') IS NOT NULL THEN
				ALTER TABLE suscripcion_facturas ADD COLUMN IF NOT EXISTS comision_centavos BIGINT;
				ALTER TABLE suscripcion_facturas ADD COLUMN IF NOT EXISTS neto_centavos BIGINT;
				ALTER TABLE suscripcion_facturas ADD COLUMN IF NOT EXISTS balance_transaction_id VARCHAR(255);
			END IF;
		END $$`,
	}
	for i, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			slog.Error("migración fallida", "index", i, "error", err, "stmt", s[:100])
			return fmt.Errorf("migración %d fallida: %w", i, err)
		}
	}
	slog.Info("cursos: migraciones aplicadas correctamente")
	return nil
}
