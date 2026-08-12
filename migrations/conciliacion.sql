-- ============================================================================
-- Conciliación diaria: órdenes ↔ Stripe
--
-- Un test verde prueba el camino del código; solo la conciliación prueba el
-- dinero. Cualquier resultado distinto de cero en la consulta 1 o 2 es un
-- incidente, no una curiosidad.
--
-- ⚠ SE CORREN EN LA BASE DEL SERVICIO DE CURSOS (`cursos` en Railway).
--   No hacen JOIN con users: esa tabla vive en la base de `auth`, en otro
--   servicio, y Postgres no puede cruzar bases. Se muestra el user_id; para
--   resolver el correo, búscalo en la base de auth con ese UUID.
--
-- Uso: pégalas en Railway → Postgres (cursos) → Data → Query.
-- ============================================================================


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 1. Órdenes cobradas que nunca se cumplieron                              ║
-- ║                                                                          ║
-- ║ El cliente pagó pero no recibió sus accesos. Es la falla más cara:       ║
-- ║ tienes su dinero y no tiene su curso. Debe ser SIEMPRE cero.             ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
SELECT
    o.id                AS orden_id,
    o.user_id,          -- resuelve el correo en la base de auth con este UUID
    o.total_centavos,
    o.moneda,
    o.stripe_session_id,
    o.pagada_at,
    NOW() - o.pagada_at AS lleva_sin_cumplirse
  FROM ordenes o
 WHERE o.estado = 'pagada'
   -- 15 minutos de margen: el webhook y verify-checkout-session compiten y
   -- uno de los dos suele tardar unos segundos.
   AND o.pagada_at < NOW() - INTERVAL '15 minutes'
 ORDER BY o.pagada_at;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 2. Desviación entre el total de la orden y la suma de sus renglones      ║
-- ║                                                                          ║
-- ║ Detecta errores de cálculo de precio. Con dinero en centavos debe ser    ║
-- ║ cero siempre; si aparece algo, hay un bug en la construcción del carrito.║
-- ╚══════════════════════════════════════════════════════════════════════════╝
SELECT
    o.id AS orden_id,
    o.total_centavos                              AS total_orden,
    COALESCE(SUM(i.subtotal_centavos), 0)         AS suma_renglones,
    o.total_centavos - COALESCE(SUM(i.subtotal_centavos), 0) AS desviacion
  FROM ordenes o
  LEFT JOIN orden_items i ON i.orden_id = o.id
 WHERE o.estado IN ('pagada','cumplida')
 GROUP BY o.id, o.total_centavos
HAVING o.total_centavos <> COALESCE(SUM(i.subtotal_centavos), 0)
 ORDER BY ABS(o.total_centavos - COALESCE(SUM(i.subtotal_centavos), 0)) DESC;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 3. Órdenes pendientes viejas                                            ║
-- ║                                                                          ║
-- ║ Sesiones de Stripe abandonadas. No son un problema en sí (el usuario     ║
-- ║ cerró la pestaña), pero un salto en el volumen delata un checkout roto.  ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
WITH por_dia AS (
    SELECT DATE(created_at) AS dia,
           count(*)                                             AS total,
           count(*) FILTER (
               WHERE estado = 'pendiente' AND created_at < NOW() - INTERVAL '2 hours'
           )                                                    AS abandonadas,
           SUM(total_centavos) FILTER (
               WHERE estado = 'pendiente' AND created_at < NOW() - INTERVAL '2 hours'
           )                                                    AS valor_perdido_centavos
      FROM ordenes
     GROUP BY DATE(created_at)
)
SELECT dia,
       abandonadas,
       COALESCE(valor_perdido_centavos, 0)                      AS valor_perdido_centavos,
       ROUND(100.0 * abandonadas / NULLIF(total, 0), 1)          AS pct_del_dia
  FROM por_dia
 WHERE abandonadas > 0
 ORDER BY dia DESC
 LIMIT 30;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 4. Resumen diario de ingresos                                            ║
-- ║                                                                          ║
-- ║ Compáralo contra el panel de Stripe. Si no cuadra, empieza por la        ║
-- ║ consulta 1: lo más probable es que falten cumplimientos, no cobros.      ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
SELECT
    DATE(pagada_at)                                     AS dia,
    count(*)                                            AS ordenes,
    SUM(total_centavos)                                 AS bruto_centavos,
    -- Comisión aproximada de Stripe México: 3.6% + $3.00 MXN por transacción.
    -- Es una estimación para tener una referencia, no el dato contable.
    SUM(ROUND(total_centavos * 0.036) + 300)            AS comision_estimada_centavos,
    SUM(total_centavos) - SUM(ROUND(total_centavos * 0.036) + 300) AS neto_estimado_centavos
  FROM ordenes
 WHERE estado IN ('pagada','cumplida')
   AND pagada_at >= NOW() - INTERVAL '30 days'
 GROUP BY DATE(pagada_at)
 ORDER BY dia DESC;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 5. Salud de las suscripciones                                            ║
-- ║                                                                          ║
-- ║ 'vencida' es el estado de gracia: el cobro falló pero el acceso sigue    ║
-- ║ vivo mientras corre el dunning. Un crecimiento sostenido ahí es          ║
-- ║ churn involuntario que aún se puede recuperar.                           ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
SELECT
    p.nombre                            AS plan,
    s.estado,
    count(*)                            AS suscripciones,
    SUM(s.asientos)                     AS asientos_totales,
    SUM(p.precio_centavos * s.asientos) AS mrr_centavos
  FROM suscripciones s
  JOIN planes p ON p.id = s.plan_id
 GROUP BY p.nombre, s.estado
 ORDER BY p.nombre, s.estado;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 6. Cobros fallidos en dunning                                            ║
-- ║                                                                          ║
-- ║ Cada fila es dinero recuperable. El intento de cobro indica en qué       ║
-- ║ punto del reintento va Stripe.                                           ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
SELECT
    s.user_id,                        -- resuelve el correo en la base de auth
    p.nombre                          AS plan,
    s.estado,
    f.total_centavos,
    f.intento_cobro,
    f.created_at                      AS ultimo_intento,
    s.periodo_fin                     AS acceso_hasta
  FROM suscripcion_facturas f
  JOIN suscripciones s ON s.id = f.suscripcion_id
  JOIN planes p ON p.id = s.plan_id
 WHERE f.estado = 'fallida'
   AND s.estado IN ('vencida','impagada')
 ORDER BY f.created_at DESC;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ 7. Webhooks con error                                                    ║
-- ║                                                                          ║
-- ║ Un evento marcado 'error' significa que Stripe informó algo que no       ║
-- ║ pudimos aplicar. Requiere reproceso manual.                              ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
SELECT event_id, tipo, procesado_at, detalle
  FROM stripe_eventos_procesados
 WHERE resultado <> 'ok'
 ORDER BY procesado_at DESC
 LIMIT 100;
