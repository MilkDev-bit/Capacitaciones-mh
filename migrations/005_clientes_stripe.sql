-- ============================================================================
-- 005 · Clientes de Stripe
--
-- ┌────────────────────────────────────────────────────────────────────────┐
-- │  ▶ EJECUTAR EN LA BASE **cursos**                                      │
-- └────────────────────────────────────────────────────────────────────────┘
--
-- Necesario para aceptar transferencias bancarias (SPEI).
--
-- La documentación de Stripe es explícita: "Enabling bank transfers on the
-- checkout page requires specifying the customer in the checkout session". No
-- basta con `customer_creation: always` —que crea el Customer DESPUÉS, para
-- compras de invitado—: la CLABE de referencia se emite a nombre de un cliente
-- que ya debe existir cuando se crea la sesión.
--
-- Hasta ahora el único `stripe_customer_id` del sistema vivía en
-- `suscripciones`, así que solo lo tenían quienes se suscribieron. Esta tabla
-- lo generaliza a cualquier comprador.
--
-- Se guarda para REUTILIZARLO. Sin esta tabla habría que crear un Customer en
-- cada compra, y el mismo alumno acabaría con un cliente distinto por pedido:
-- los saldos quedarían repartidos entre varios clientes y la conciliación de
-- transferencias sería imposible.

BEGIN;

CREATE TABLE IF NOT EXISTS stripe_clientes (
    -- Sin FK: `users` vive en la base de auth, no en esta.
    user_id            UUID PRIMARY KEY,
    stripe_customer_id VARCHAR(255) NOT NULL UNIQUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rellena la tabla con los clientes que ya existen en `suscripciones`, para no
-- crear un Customer nuevo a quien ya tiene uno. ON CONFLICT DO NOTHING porque
-- un usuario puede tener varias filas de suscripción con el mismo cliente.
INSERT INTO stripe_clientes (user_id, stripe_customer_id)
SELECT DISTINCT ON (user_id) user_id, stripe_customer_id
  FROM suscripciones
 WHERE stripe_customer_id IS NOT NULL
   AND stripe_customer_id <> ''
 ORDER BY user_id, created_at DESC
ON CONFLICT DO NOTHING;

COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ VERIFICACIÓN                                                             ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT count(*) AS clientes_registrados FROM stripe_clientes;

-- Ningún usuario debe tener dos clientes distintos, ni un cliente estar
-- asignado a dos usuarios. Cero filas en ambas = todo correcto.
SELECT user_id, count(*)
  FROM stripe_clientes
 GROUP BY user_id HAVING count(*) > 1;
