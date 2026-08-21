-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 003 · Comisiones de Stripe                                               ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
--
-- Hasta ahora no se guardaba en ninguna parte lo que Stripe se queda por cada
-- cobro. El panel de administración enseñaba una "venta neta" calculada con una
-- tarifa fija de 3.6% + 3 MXN, que es la tarifa de lista de tarjeta nacional y
-- no coincide con lo que Stripe cobra de verdad: OXXO, tarjeta internacional y
-- meses sin intereses tienen tarifas distintas, y encima falta el IVA.
--
-- La cifra buena la da Stripe en el BalanceTransaction de cada cobro: trae `fee`
-- y `net` ya calculados, al centavo y en la moneda del saldo. Esto añade dónde
-- guardarlos.
--
-- NULL y 0 significan cosas distintas a propósito:
--   NULL → todavía no se ha consultado a Stripe (órdenes viejas, o el webhook
--          llegó antes de que el BalanceTransaction estuviera disponible).
--   0    → consultado, y la comisión fue cero (cobro de importe cero).
-- Confundirlos haría que el histórico sin rellenar apareciera como si Stripe no
-- hubiera cobrado nada, inflando la ganancia neta.

BEGIN;

ALTER TABLE ordenes
    ADD COLUMN IF NOT EXISTS comision_centavos      BIGINT,
    ADD COLUMN IF NOT EXISTS neto_centavos          BIGINT,
    ADD COLUMN IF NOT EXISTS balance_transaction_id VARCHAR(255);

ALTER TABLE suscripcion_facturas
    ADD COLUMN IF NOT EXISTS comision_centavos      BIGINT,
    ADD COLUMN IF NOT EXISTS neto_centavos          BIGINT,
    ADD COLUMN IF NOT EXISTS balance_transaction_id VARCHAR(255);

-- Ni la comisión ni el neto pueden ser negativos. El neto sí puede ser menor
-- que el total (esa es justamente la comisión), pero nunca por debajo de cero.
ALTER TABLE ordenes
    DROP CONSTRAINT IF EXISTS chk_ordenes_comision_no_negativa;
ALTER TABLE ordenes
    ADD CONSTRAINT chk_ordenes_comision_no_negativa
    CHECK (comision_centavos IS NULL OR comision_centavos >= 0);

ALTER TABLE suscripcion_facturas
    DROP CONSTRAINT IF EXISTS chk_facturas_comision_no_negativa;
ALTER TABLE suscripcion_facturas
    ADD CONSTRAINT chk_facturas_comision_no_negativa
    CHECK (comision_centavos IS NULL OR comision_centavos >= 0);

-- Índice para el relleno del histórico: busca órdenes cobradas a las que aún
-- les falta la comisión. Parcial, porque en cuanto se rellenan dejan de
-- interesar y el índice se queda pequeño para siempre.
CREATE INDEX IF NOT EXISTS idx_ordenes_sin_comision
    ON ordenes (pagada_at)
    WHERE comision_centavos IS NULL
      AND estado IN ('pagada', 'cumplida')
      AND stripe_payment_intent IS NOT NULL;

COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ VERIFICACIÓN                                                             ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- Cuánto histórico queda por rellenar. Se ejecuta el comando de relleno hasta
-- que esto dé cero.
SELECT count(*) AS ordenes_sin_comision
  FROM ordenes
 WHERE comision_centavos IS NULL
   AND estado IN ('pagada', 'cumplida')
   AND stripe_payment_intent IS NOT NULL;

-- El neto guardado tiene que cuadrar con total - comisión. Si esto devuelve
-- filas, algo se guardó mal. Salvedad: si algún día se cobra en una moneda
-- distinta a la del saldo, Stripe convierte y el neto deja de ser una resta
-- exacta; hoy todo se cobra en MXN, así que debe dar cero filas.
SELECT id, total_centavos, comision_centavos, neto_centavos
  FROM ordenes
 WHERE comision_centavos IS NOT NULL
   AND neto_centavos IS NOT NULL
   AND neto_centavos <> total_centavos - comision_centavos;
