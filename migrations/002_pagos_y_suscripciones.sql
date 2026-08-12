-- ============================================================================
-- Fundación de pagos: dinero en centavos, órdenes, webhooks y suscripciones
--
-- ⚠ SE CORRE EN LA BASE DEL SERVICIO DE CURSOS (la llamada `cursos` en Railway),
--   porque toca capacitaciones y curso_licencias, que viven ahí.
--
--   NO lleva llaves foráneas a users: esa tabla está en la base de `auth` y
--   Postgres no permite FK entre bases distintas. user_id se guarda como UUID
--   suelto, igual que ya hace el resto de cursos-service (asignaciones,
--   inscripciones, licencia_invitaciones).
--
-- Idempotente: puede correrse varias veces sin efecto adicional.
-- Ejecución:
--   psql "$DATABASE_URL" -f migrations/002_pagos_y_suscripciones.sql
--   o pega el contenido en Railway → Postgres → Data → Query
-- ============================================================================

BEGIN;

-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 1. DINERO EN CENTAVOS                                                    ║
-- ║                                                                          ║
-- ║ Las columnas NUMERIC(10,2) se leían como float64 y el monto de Stripe   ║
-- ║ salía de int64(precio*100), que trunca: $8.20 se cobraba como 819        ║
-- ║ centavos. Se añade la columna entera y se rellena redondeando.           ║
-- ║ Las columnas viejas se conservan durante la transición.                  ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

ALTER TABLE capacitaciones  ADD COLUMN IF NOT EXISTS precio_centavos BIGINT NOT NULL DEFAULT 0;
ALTER TABLE curso_licencias ADD COLUMN IF NOT EXISTS precio_centavos BIGINT NOT NULL DEFAULT 0;
ALTER TABLE capacitaciones  ADD COLUMN IF NOT EXISTS moneda CHAR(3) NOT NULL DEFAULT 'MXN';

-- ROUND() sobre NUMERIC es exacto: opera en decimal, no en binario, así que
-- no arrastra el error que tenía el float64 de Go.
UPDATE capacitaciones
   SET precio_centavos = ROUND(precio * 100)::BIGINT
 WHERE precio_centavos = 0 AND precio > 0;

UPDATE curso_licencias
   SET precio_centavos = ROUND(precio * 100)::BIGINT
 WHERE precio_centavos = 0 AND precio > 0;

-- Un precio negativo no tiene sentido y rompería el checkout.
ALTER TABLE capacitaciones  DROP CONSTRAINT IF EXISTS chk_capacitaciones_precio_no_negativo;
ALTER TABLE capacitaciones  ADD  CONSTRAINT chk_capacitaciones_precio_no_negativo CHECK (precio_centavos >= 0);
ALTER TABLE curso_licencias DROP CONSTRAINT IF EXISTS chk_licencias_precio_no_negativo;
ALTER TABLE curso_licencias ADD  CONSTRAINT chk_licencias_precio_no_negativo CHECK (precio_centavos >= 0);


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 2. ÓRDENES                                                               ║
-- ║                                                                          ║
-- ║ Hoy no queda rastro de la compra: si el webhook falla no hay forma de    ║
-- ║ saber qué se intentó cobrar. La orden se crea ANTES de mandar al usuario ║
-- ║ a Stripe, y su id alimenta la clave de idempotencia.                     ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

CREATE TABLE IF NOT EXISTS ordenes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Sin FK: users vive en la base de auth, en otro servicio.
    user_id UUID NOT NULL,

    -- pendiente  → creada, aún no se paga
    -- pagada     → Stripe confirmó el cobro
    -- cumplida   → accesos/licencias ya otorgados
    -- fallida    → el pago fue rechazado o la sesión expiró
    -- reembolsada→ se devolvió el dinero
    estado VARCHAR(20) NOT NULL DEFAULT 'pendiente',

    total_centavos BIGINT NOT NULL DEFAULT 0,
    moneda CHAR(3) NOT NULL DEFAULT 'MXN',

    stripe_session_id       VARCHAR(255) UNIQUE,
    stripe_payment_intent   VARCHAR(255),
    -- Deriva de la orden, no de un UUID por petición: un doble clic o un
    -- reintento del cliente resuelven a la misma sesión de Stripe.
    idempotency_key         VARCHAR(255) UNIQUE NOT NULL,
    intento                 INT NOT NULL DEFAULT 1,

    motivo_fallo TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pagada_at    TIMESTAMPTZ,
    cumplida_at  TIMESTAMPTZ,

    CONSTRAINT chk_ordenes_estado CHECK (
        estado IN ('pendiente','pagada','cumplida','fallida','reembolsada')
    ),
    CONSTRAINT chk_ordenes_total_no_negativo CHECK (total_centavos >= 0)
);
CREATE INDEX IF NOT EXISTS idx_ordenes_user   ON ordenes(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ordenes_estado ON ordenes(estado) WHERE estado IN ('pendiente','pagada');

-- Renglones de la orden. Se congela el precio al momento de comprar: si el
-- curso sube de precio mañana, la orden histórica no debe cambiar.
CREATE TABLE IF NOT EXISTS orden_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    orden_id UUID NOT NULL REFERENCES ordenes(id) ON DELETE CASCADE,
    capacitacion_id UUID REFERENCES capacitaciones(id) ON DELETE SET NULL,

    tipo VARCHAR(20) NOT NULL,              -- 'b2c' | 'b2b_direct'
    titulo VARCHAR(200) NOT NULL DEFAULT '',
    cantidad INT NOT NULL DEFAULT 1,
    precio_unitario_centavos BIGINT NOT NULL DEFAULT 0,
    subtotal_centavos BIGINT NOT NULL DEFAULT 0,

    licencia_id UUID REFERENCES curso_licencias(id) ON DELETE SET NULL,

    CONSTRAINT chk_orden_items_tipo CHECK (tipo IN ('b2c','b2b_direct')),
    CONSTRAINT chk_orden_items_cantidad CHECK (cantidad > 0)
);
CREATE INDEX IF NOT EXISTS idx_orden_items_orden ON orden_items(orden_id);


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 3. DEDUPE DE WEBHOOKS                                                    ║
-- ║                                                                          ║
-- ║ Stripe entrega al-menos-una-vez: en la práctica, dos veces. La PK sobre  ║
-- ║ event_id hace la deduplicación atómica con INSERT ... ON CONFLICT.       ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

CREATE TABLE IF NOT EXISTS stripe_eventos_procesados (
    event_id     VARCHAR(255) PRIMARY KEY,
    tipo         VARCHAR(100) NOT NULL,
    procesado_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resultado    VARCHAR(20)  NOT NULL DEFAULT 'ok',
    detalle      TEXT
);
CREATE INDEX IF NOT EXISTS idx_stripe_eventos_fecha ON stripe_eventos_procesados(procesado_at DESC);


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 4. SUSCRIPCIONES                                                         ║
-- ║                                                                          ║
-- ║ Dos modelos conviviendo:                                                 ║
-- ║   'individual' → una persona, acceso a todo el catálogo                  ║
-- ║   'asientos'   → una empresa paga N asientos para su equipo              ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

CREATE TABLE IF NOT EXISTS planes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo VARCHAR(50) UNIQUE NOT NULL,     -- 'individual_mensual', 'empresa_anual'
    nombre VARCHAR(100) NOT NULL,
    descripcion TEXT NOT NULL DEFAULT '',

    modalidad VARCHAR(20) NOT NULL,         -- 'individual' | 'asientos'
    intervalo VARCHAR(10) NOT NULL,         -- 'mes' | 'anio'
    precio_centavos BIGINT NOT NULL,        -- por asiento cuando modalidad='asientos'
    moneda CHAR(3) NOT NULL DEFAULT 'MXN',

    -- El precio real lo manda Stripe: aquí se guarda para poder pintarlo sin
    -- llamar a su API en cada carga de la página de planes.
    stripe_price_id VARCHAR(255) UNIQUE,

    dias_prueba INT NOT NULL DEFAULT 0,
    activo BOOLEAN NOT NULL DEFAULT true,
    orden INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_planes_modalidad CHECK (modalidad IN ('individual','asientos')),
    CONSTRAINT chk_planes_intervalo CHECK (intervalo IN ('mes','anio')),
    CONSTRAINT chk_planes_precio CHECK (precio_centavos >= 0)
);

CREATE TABLE IF NOT EXISTS suscripciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- El titular: la persona en 'individual', el comprador en 'asientos'.
    -- Sin FK: users vive en la base de auth.
    user_id UUID NOT NULL,
    plan_id UUID NOT NULL REFERENCES planes(id) ON DELETE RESTRICT,

    -- Refleja los estados de Stripe. past_due NO revoca el acceso: es el
    -- periodo de gracia durante el que corre el dunning.
    estado VARCHAR(20) NOT NULL DEFAULT 'incompleta',

    asientos INT NOT NULL DEFAULT 1,

    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id     VARCHAR(255),

    periodo_inicio TIMESTAMPTZ,
    periodo_fin    TIMESTAMPTZ,
    prueba_fin     TIMESTAMPTZ,
    -- Cancelación al final del periodo: sigue activa hasta periodo_fin.
    cancelar_al_terminar BOOLEAN NOT NULL DEFAULT false,
    cancelada_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_suscripciones_estado CHECK (
        estado IN ('incompleta','en_prueba','activa','vencida','cancelada','impagada')
    ),
    CONSTRAINT chk_suscripciones_asientos CHECK (asientos > 0)
);
CREATE INDEX IF NOT EXISTS idx_suscripciones_user ON suscripciones(user_id);
CREATE INDEX IF NOT EXISTS idx_suscripciones_activas ON suscripciones(estado)
    WHERE estado IN ('activa','en_prueba','vencida');

-- Una persona no puede tener dos suscripciones vivas a la vez. El índice
-- parcial único lo impide a nivel de BD, no solo en la aplicación.
CREATE UNIQUE INDEX IF NOT EXISTS uq_suscripcion_viva_por_usuario
    ON suscripciones(user_id)
    WHERE estado IN ('incompleta','en_prueba','activa','vencida');

-- Quién ocupa cada asiento de una suscripción corporativa.
CREATE TABLE IF NOT EXISTS suscripcion_asientos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    suscripcion_id UUID NOT NULL REFERENCES suscripciones(id) ON DELETE CASCADE,
    user_id UUID,  -- se rellena cuando la persona invitada crea su cuenta
    email VARCHAR(200) NOT NULL,
    estado VARCHAR(20) NOT NULL DEFAULT 'invitado',   -- 'invitado' | 'activo' | 'revocado'
    invitado_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activado_at TIMESTAMPTZ,

    UNIQUE(suscripcion_id, email),
    CONSTRAINT chk_asientos_estado CHECK (estado IN ('invitado','activo','revocado'))
);
CREATE INDEX IF NOT EXISTS idx_asientos_suscripcion ON suscripcion_asientos(suscripcion_id);
CREATE INDEX IF NOT EXISTS idx_asientos_user ON suscripcion_asientos(user_id) WHERE user_id IS NOT NULL;

-- Historial de facturación: alimenta la conciliación y la vista de recibos.
CREATE TABLE IF NOT EXISTS suscripcion_facturas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    suscripcion_id UUID NOT NULL REFERENCES suscripciones(id) ON DELETE CASCADE,
    stripe_invoice_id VARCHAR(255) UNIQUE NOT NULL,
    estado VARCHAR(20) NOT NULL,            -- 'pagada' | 'fallida' | 'abierta'
    total_centavos BIGINT NOT NULL DEFAULT 0,
    moneda CHAR(3) NOT NULL DEFAULT 'MXN',
    intento_cobro INT NOT NULL DEFAULT 1,
    url_pdf TEXT,
    periodo_inicio TIMESTAMPTZ,
    periodo_fin TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_facturas_suscripcion ON suscripcion_facturas(suscripcion_id, created_at DESC);

COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ VERIFICACIÓN                                                             ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- Ningún precio debe haberse desviado al pasar a centavos.
SELECT 'capacitaciones' AS tabla, count(*) AS precios_desviados
  FROM capacitaciones
 WHERE precio_centavos <> ROUND(precio * 100)::BIGINT
UNION ALL
SELECT 'curso_licencias', count(*)
  FROM curso_licencias
 WHERE precio_centavos <> ROUND(precio * 100)::BIGINT;

-- Ambos conteos deben salir en 0.


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PLANES SEMILLA                                                           ║
-- ║                                                                          ║
-- ║ Los precios de aquí son solo para pintar la página sin llamar a Stripe.  ║
-- ║ El cobro real usa stripe_price_id, que debes rellenar con los Price IDs  ║
-- ║ de tu panel de Stripe (Productos → Precios → price_...).                 ║
-- ║                                                                          ║
-- ║ Sin stripe_price_id el checkout de ese plan devuelve error explícito.    ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

INSERT INTO planes (codigo, nombre, descripcion, modalidad, intervalo, precio_centavos, dias_prueba, orden)
VALUES
  ('individual_mensual', 'Individual mensual',
   'Acceso a todo el catálogo para una persona. Cancela cuando quieras.',
   'individual', 'mes',  49900, 7, 1),
  ('individual_anual', 'Individual anual',
   'Acceso a todo el catálogo para una persona. Dos meses gratis frente al mensual.',
   'individual', 'anio', 499000, 7, 2),
  ('empresa_mensual', 'Empresa mensual',
   'Precio por asiento. Reparte los lugares entre tu equipo y añade más cuando crezcas.',
   'asientos', 'mes',  39900, 0, 3),
  ('empresa_anual', 'Empresa anual',
   'Precio por asiento con dos meses de descuento. Incluye constancias DC-3.',
   'asientos', 'anio', 399000, 0, 4)
ON CONFLICT (codigo) DO NOTHING;

-- Recordatorio: enlaza cada plan con su precio de Stripe antes de vender.
--   UPDATE planes SET stripe_price_id = 'price_XXXX' WHERE codigo = 'individual_mensual';
SELECT codigo, nombre, precio_centavos,
       CASE WHEN stripe_price_id IS NULL THEN '⚠ falta stripe_price_id' ELSE 'listo' END AS estado
  FROM planes ORDER BY orden;
