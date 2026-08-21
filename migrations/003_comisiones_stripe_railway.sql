-- ============================================================================
-- Comisiones de Stripe — versión para la consola de Railway
--
-- Pega esto en:  Railway → servicio Postgres → pestaña Data → Query
--
-- Ejecuta los pasos EN ORDEN, uno a uno, y mira el resultado de cada uno antes
-- de seguir. Todo es idempotente: repetirlo no rompe ni duplica nada.
--
-- Si ya desplegaste cursos-service después de este cambio, los PASOS 1 y 2 ya
-- están aplicados por el arranque del servicio. Ejecútalos igual: no harán nada
-- y el PASO 3 te lo confirmará.
-- ============================================================================


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 0 — ¿Existen las tablas de pagos?                                   ║
-- ║ Si esto devuelve NULL en alguna, PARA: falta correr 002 primero.         ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT to_regclass('public.ordenes')              AS tabla_ordenes,
       to_regclass('public.suscripcion_facturas') AS tabla_facturas;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 1 — Columnas de comisión                                            ║
-- ║                                                                          ║
-- ║ NULL y 0 significan cosas distintas y hay que respetarlo:                ║
-- ║   NULL → todavía no se le ha preguntado a Stripe                         ║
-- ║   0    → preguntado, y la comisión fue cero                              ║
-- ║ Por eso las columnas NO llevan DEFAULT 0: un cero por defecto haría que  ║
-- ║ todo el histórico pareciera libre de comisiones e inflaría el neto.      ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

ALTER TABLE ordenes
    ADD COLUMN IF NOT EXISTS comision_centavos      BIGINT,
    ADD COLUMN IF NOT EXISTS neto_centavos          BIGINT,
    ADD COLUMN IF NOT EXISTS balance_transaction_id VARCHAR(255);

ALTER TABLE suscripcion_facturas
    ADD COLUMN IF NOT EXISTS comision_centavos      BIGINT,
    ADD COLUMN IF NOT EXISTS neto_centavos          BIGINT,
    ADD COLUMN IF NOT EXISTS balance_transaction_id VARCHAR(255);


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 2 — Restricciones e índice                                          ║
-- ║ El índice es parcial: en cuanto una orden tiene su comisión deja de      ║
-- ║ aparecer, así que se mantiene pequeño para siempre.                      ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

ALTER TABLE ordenes DROP CONSTRAINT IF EXISTS chk_ordenes_comision_no_negativa;
ALTER TABLE ordenes ADD  CONSTRAINT chk_ordenes_comision_no_negativa
    CHECK (comision_centavos IS NULL OR comision_centavos >= 0);

ALTER TABLE suscripcion_facturas DROP CONSTRAINT IF EXISTS chk_facturas_comision_no_negativa;
ALTER TABLE suscripcion_facturas ADD  CONSTRAINT chk_facturas_comision_no_negativa
    CHECK (comision_centavos IS NULL OR comision_centavos >= 0);

CREATE INDEX IF NOT EXISTS idx_ordenes_sin_comision
    ON ordenes (pagada_at)
    WHERE comision_centavos IS NULL
      AND estado IN ('pagada', 'cumplida')
      AND stripe_payment_intent IS NOT NULL;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 3 — Comprobar que quedó                                             ║
-- ║ Deben salir las 3 columnas por tabla, todas is_nullable = YES.           ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT table_name, column_name, data_type, is_nullable
  FROM information_schema.columns
 WHERE table_name IN ('ordenes', 'suscripcion_facturas')
   AND column_name IN ('comision_centavos', 'neto_centavos', 'balance_transaction_id')
 ORDER BY table_name, column_name;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 4 — Cuánto histórico falta por rellenar                             ║
-- ║                                                                          ║
-- ║ Este número es el que sale en el aviso del panel de Finanzas. Se baja    ║
-- ║ pulsando "Traer de Stripe" ahí, no desde aquí: la comisión hay que       ║
-- ║ pedírsela a Stripe cobro por cobro y eso no se puede hacer en SQL.       ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT count(*) AS ordenes_sin_comision
  FROM ordenes
 WHERE comision_centavos IS NULL
   AND estado IN ('pagada', 'cumplida')
   AND stripe_payment_intent IS NOT NULL;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 5 — Cuadre (ejecútalo DESPUÉS de rellenar desde el panel)           ║
-- ║ El neto guardado debe ser total − comisión. Cero filas = todo cuadra.    ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT id, total_centavos, comision_centavos, neto_centavos
  FROM ordenes
 WHERE comision_centavos IS NOT NULL
   AND neto_centavos IS NOT NULL
   AND neto_centavos <> total_centavos - comision_centavos;
