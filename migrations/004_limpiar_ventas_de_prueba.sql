-- ============================================================================
-- Limpieza de ventas de prueba
--
-- ⚠️  BORRA DATOS DE FORMA PERMANENTE. Léelo entero antes de ejecutar nada.
--
-- ┌────────────────────────────────────────────────────────────────────────┐
-- │  ▶ EJECUTAR EN LA BASE **cursos**  (Railway → grupo Cursos → `cursos`) │
-- │                                                                        │
-- │  NO en `auth`, NO en `lecciones`, NO en la de ningún otro servicio.    │
-- └────────────────────────────────────────────────────────────────────────┘
--
-- Cada microservicio tiene su propio Postgres. Todo lo relacionado con cobros
-- —ordenes, orden_items, suscripciones, curso_licencias— vive en la base de
-- **cursos**. La tabla `users` está en la de **auth** y aquí no se toca.
--
-- Pega esto en: Railway → base `cursos` → pestaña Data → Query.
-- Ejecuta los pasos EN ORDEN. El PASO 0 y el 1 no borran nada.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- LO QUE **NO** TOCA, a propósito:
--   · usuarios (están en otra base)
--   · cursos, lecciones, exámenes, foros ni mensajes
--   · el progreso de los alumnos
--   · las constancias DC-3 ya emitidas
-- ============================================================================


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 0 — ¿Estoy en la base correcta?                                     ║
-- ║                                                                          ║
-- ║ `ordenes` y `capacitaciones` deben salir con nombre, y `users` en NULL.  ║
-- ║ Si `users` NO es NULL, estás en la base de auth: DETENTE.                ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT current_database()                     AS base,
       to_regclass('public.ordenes')          AS tabla_ordenes,
       to_regclass('public.capacitaciones')   AS tabla_capacitaciones,
       to_regclass('public.users')            AS tabla_users_debe_ser_null;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 1 — Inventario (NO BORRA NADA)                                      ║
-- ║ Mira estos números y confirma que todo lo que sale es de prueba.         ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT 'ordenes'                  AS tabla, count(*) AS filas FROM ordenes
UNION ALL SELECT 'orden_items',              count(*) FROM orden_items
UNION ALL SELECT 'suscripcion_facturas',     count(*) FROM suscripcion_facturas
UNION ALL SELECT 'suscripcion_asientos',     count(*) FROM suscripcion_asientos
UNION ALL SELECT 'suscripciones',            count(*) FROM suscripciones
UNION ALL SELECT 'stripe_eventos_procesados', count(*) FROM stripe_eventos_procesados
UNION ALL SELECT 'licencias vendidas',
                 count(*) FROM curso_licencias WHERE comprador_id IS NOT NULL
UNION ALL SELECT 'licencia_invitaciones',    count(*) FROM licencia_invitaciones
UNION ALL SELECT 'inscripciones por licencia',
                 count(*) FROM inscripciones WHERE licencia_id IS NOT NULL;

-- Detalle de lo cobrado, por si alguna compra SÍ fuera real. No hay correo:
-- los usuarios están en otra base. El user_id sirve para cotejarlo en `auth`.
SELECT id, estado, total_centavos, moneda, user_id,
       COALESCE(pagada_at, created_at) AS fecha
  FROM ordenes
 ORDER BY fecha DESC;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 2 — Borrar el historial de cobros                                   ║
-- ║                                                                          ║
-- ║ Este bloque deja Finanzas y el Panel de Administración a cero.           ║
-- ║ orden_items cae solo por ON DELETE CASCADE, pero se borra explícito para ║
-- ║ que no dependa de que la restricción esté como creemos.                  ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

BEGIN;

DELETE FROM orden_items;
DELETE FROM ordenes;

DELETE FROM suscripcion_facturas;
DELETE FROM suscripcion_asientos;
DELETE FROM suscripciones;

-- Deduplicación de webhooks de Stripe. Se vacía para que, si reenvías un
-- evento de prueba desde el panel de Stripe, no se descarte por "ya procesado".
DELETE FROM stripe_eventos_procesados;

COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 3 — Licencias vendidas a empresas (OPCIONAL)                        ║
-- ║                                                                          ║
-- ║ Solo si las 5 licencias B2B del panel también son de prueba.             ║
-- ║                                                                          ║
-- ║ OJO: si alguien entró a un curso con el código de esas licencias, su     ║
-- ║ inscripción quedaría apuntando a una licencia borrada. Por eso primero   ║
-- ║ se desligan las inscripciones (el alumno CONSERVA acceso y avance) y     ║
-- ║ solo después se borra la licencia.                                       ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

BEGIN;

UPDATE inscripciones SET licencia_id = NULL WHERE licencia_id IS NOT NULL;

DELETE FROM licencia_invitaciones;

-- Solo las vendidas. Las que no tienen comprador son plantillas que el
-- instructor dejó creadas y no se han vendido: esas se conservan.
DELETE FROM curso_licencias WHERE comprador_id IS NOT NULL;

COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 4 — Inscripciones regaladas en pruebas (OPCIONAL)                   ║
-- ║                                                                          ║
-- ║ Ejecútalo SOLO si quieres que los alumnos de prueba pierdan el acceso.   ║
-- ║ Borra su avance en esos cursos, y eso no se recupera.                    ║
-- ║                                                                          ║
-- ║ Aquí NO se puede filtrar por correo: `users` está en la base de auth.    ║
-- ║ Saca los UUID con esta consulta EN LA BASE `auth`:                       ║
-- ║                                                                          ║
-- ║   SELECT id, email FROM users WHERE email IN ('prueba@ejemplo.mx');      ║
-- ║                                                                          ║
-- ║ y pégalos abajo.                                                         ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- BEGIN;
-- DELETE FROM progreso_lecciones
--  WHERE user_id IN ('00000000-0000-0000-0000-000000000000'::uuid);
-- DELETE FROM inscripciones
--  WHERE user_id IN ('00000000-0000-0000-0000-000000000000'::uuid);
-- COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 5 — Comprobar que quedó a cero                                      ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT 'ordenes'             AS tabla, count(*) AS filas FROM ordenes
UNION ALL SELECT 'suscripcion_facturas', count(*) FROM suscripcion_facturas
UNION ALL SELECT 'licencias vendidas',   count(*) FROM curso_licencias WHERE comprador_id IS NOT NULL;

-- Lo que verá el panel de Finanzas: todo en cero.
SELECT COALESCE(SUM(total_centavos), 0)    AS bruto_centavos,
       COALESCE(SUM(comision_centavos), 0) AS comision_centavos,
       count(*)                            AS transacciones
  FROM ordenes
 WHERE estado IN ('pagada', 'cumplida');
