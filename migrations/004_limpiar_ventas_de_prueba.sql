-- ============================================================================
-- Limpieza de ventas de prueba
--
-- ⚠️  BORRA DATOS DE FORMA PERMANENTE. Léelo entero antes de ejecutar nada.
--
-- Pensado para el momento previo a salir a producción: todas las compras que
-- hay en la base son pruebas y hay que dejar los contadores a cero.
--
-- Pega esto en:  Railway → servicio Postgres → pestaña Data → Query
--
-- Ejecuta los pasos EN ORDEN, uno a uno, y mira el resultado de cada uno.
-- El PASO 1 no borra nada: solo te enseña qué se va a llevar por delante.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- LO QUE **NO** TOCA, a propósito:
--   · usuarios, cursos, lecciones, exámenes, foros ni mensajes
--   · el progreso de los alumnos
--   · las constancias DC-3 ya emitidas
--
-- Si además quieres dejar la plataforma como recién instalada (sin usuarios ni
-- cursos), eso es otra cosa y NO está aquí: es mucho más destructivo y conviene
-- decidirlo aparte.
-- ============================================================================


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 1 — Inventario (NO BORRA NADA)                                      ║
-- ║ Mira estos números y confirma que todo lo que sale es de prueba.         ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT 'ordenes'               AS tabla, count(*) AS filas FROM ordenes
UNION ALL SELECT 'orden_items',            count(*) FROM orden_items
UNION ALL SELECT 'suscripcion_facturas',   count(*) FROM suscripcion_facturas
UNION ALL SELECT 'suscripcion_asientos',   count(*) FROM suscripcion_asientos
UNION ALL SELECT 'suscripciones',          count(*) FROM suscripciones
UNION ALL SELECT 'stripe_eventos_procesados', count(*) FROM stripe_eventos_procesados
UNION ALL SELECT 'licencias vendidas (curso_licencias con comprador)',
                 count(*) FROM curso_licencias WHERE comprador_id IS NOT NULL
UNION ALL SELECT 'licencia_invitaciones',  count(*) FROM licencia_invitaciones
UNION ALL SELECT 'inscripciones por licencia',
                 count(*) FROM inscripciones WHERE licencia_id IS NOT NULL;

-- Detalle de lo cobrado, por si reconoces alguna compra que SÍ sea real.
SELECT o.id, o.estado, o.total_centavos, o.moneda,
       COALESCE(o.pagada_at, o.created_at) AS fecha, u.email
  FROM ordenes o
  LEFT JOIN users u ON u.id = o.user_id
 ORDER BY fecha DESC;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 2 — Borrar el historial de cobros                                   ║
-- ║                                                                          ║
-- ║ Este es el bloque que deja Finanzas y el Panel de Administración a cero. ║
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
-- ║ se desligan las inscripciones (el alumno CONSERVA el acceso y su avance) ║
-- ║ y solo después se borra la licencia.                                     ║
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
-- ║ Descomenta y sustituye los correos por los tuyos de prueba.              ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- BEGIN;
-- DELETE FROM progreso_lecciones
--  WHERE user_id IN (SELECT id FROM users WHERE email IN ('prueba1@ejemplo.mx','prueba2@ejemplo.mx'));
-- DELETE FROM inscripciones
--  WHERE user_id IN (SELECT id FROM users WHERE email IN ('prueba1@ejemplo.mx','prueba2@ejemplo.mx'));
-- COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 5 — Comprobar que quedó a cero                                      ║
-- ║ Las tres primeras deben dar 0.                                           ║
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
