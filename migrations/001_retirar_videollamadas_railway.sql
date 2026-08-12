-- ============================================================================
-- Retiro del flujo de videollamadas — versión para la consola de Railway
--
-- Pega esto en:  Railway → tu servicio Postgres → pestaña Data → Query
--
-- Es SQL puro (no usa pg_dump, que es un binario de terminal y no corre en una
-- consola SQL). El respaldo se hace copiando las tablas dentro de la misma base.
--
-- Ejecuta los pasos EN ORDEN y revisa el resultado de cada uno antes de seguir.
-- ============================================================================


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 1 — Respaldo                                                        ║
-- ║ Copia las tablas a respaldo_*. Idempotente: si ya existen, no hace nada. ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

DO $$
BEGIN
  IF to_regclass('public.instructor_schedules') IS NOT NULL
     AND to_regclass('public.respaldo_instructor_schedules') IS NULL THEN
    CREATE TABLE respaldo_instructor_schedules AS
      SELECT * FROM instructor_schedules;
  END IF;

  IF to_regclass('public.videocall_tickets') IS NOT NULL
     AND to_regclass('public.respaldo_videocall_tickets') IS NULL THEN
    CREATE TABLE respaldo_videocall_tickets AS
      SELECT * FROM videocall_tickets;
  END IF;

  -- La relación invitación → ticket se pierde al soltar la columna: se guarda
  -- aparte para no quedarse sin la trazabilidad de quién recibió qué código.
  IF to_regclass('public.licencia_invitaciones') IS NOT NULL
     AND to_regclass('public.respaldo_invitacion_ticket') IS NULL
     AND EXISTS (
       SELECT 1 FROM information_schema.columns
        WHERE table_name = 'licencia_invitaciones' AND column_name = 'ticket_id'
     ) THEN
    CREATE TABLE respaldo_invitacion_ticket AS
      SELECT id AS invitacion_id, licencia_id, email, codigo, ticket_id
        FROM licencia_invitaciones;
  END IF;
END $$;

-- Verifica que el respaldo tenga las mismas filas que el original.
--
-- Se cuenta vía query_to_xml en lugar de un SELECT directo para que la consulta
-- no falle si una tabla ya no existe (al reejecutar el script, o en una base
-- que nunca tuvo videollamadas). Un SELECT normal abortaría con "does not exist".
SELECT tabla, (xpath('/row/c/text()', conteo))[1]::text::int AS filas
  FROM (
    SELECT t AS tabla,
           query_to_xml(format('SELECT count(*) AS c FROM %I', t), false, true, '') AS conteo
      FROM unnest(ARRAY[
             'instructor_schedules', 'respaldo_instructor_schedules',
             'videocall_tickets',    'respaldo_videocall_tickets'
           ]) AS t
     WHERE to_regclass('public.' || t) IS NOT NULL
  ) s
 ORDER BY tabla;

-- ⚠️  NO SIGAS si los conteos de una tabla y su respaldo no coinciden.
--     Si no aparece ninguna fila, esta base nunca tuvo videollamadas: puedes
--     saltarte el paso 2 salvo por el UPDATE de despublicación.


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 2 — Retiro                                                          ║
-- ║ Todo en una transacción: o se aplica completo, o no se aplica nada.      ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

BEGIN;

-- Soltar la referencia desde licencia_invitaciones. Sin esto el DROP de
-- videocall_tickets falla por la llave foránea.
ALTER TABLE IF EXISTS licencia_invitaciones DROP COLUMN IF EXISTS ticket_id;

-- Despublicar los cursos que quedaron sin flujo utilizable. Se conserva el
-- contenido, las inscripciones y el historial de compras: solo dejan de
-- ofrecerse en la tienda y el catálogo.
UPDATE capacitaciones
   SET is_public = false
 WHERE type = 'videocall'
   AND is_public = true;

-- videocall_tickets primero: referencia a instructor_schedules.
DROP TABLE IF EXISTS videocall_tickets;
DROP TABLE IF EXISTS instructor_schedules;

ALTER TABLE capacitaciones DROP COLUMN IF EXISTS videocall_status;

COMMIT;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ PASO 3 — Comprobación                                                    ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

SELECT
  to_regclass('public.videocall_tickets')     IS NULL AS tickets_eliminados,
  to_regclass('public.instructor_schedules')  IS NULL AS horarios_eliminados,
  to_regclass('public.respaldo_videocall_tickets') IS NOT NULL AS respaldo_intacto,
  NOT EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name = 'capacitaciones' AND column_name = 'videocall_status'
  ) AS columna_eliminada,
  (SELECT count(*) FROM capacitaciones WHERE type = 'videocall' AND is_public) AS vc_aun_publicos;

-- Los cuatro booleanos deben salir en `t` y vc_aun_publicos en 0.


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ OPCIONAL                                                                 ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- A) Reclasificar los cursos retirados como material autoformativo, para poder
--    volver a publicarlos. Revísalos uno por uno antes:
--
--    SELECT id, title, is_public FROM capacitaciones WHERE type = 'videocall';
--    UPDATE capacitaciones SET type = 'video' WHERE type = 'videocall';

-- B) Cuando estés seguro de que no necesitas el histórico, libera el espacio:
--
--    DROP TABLE IF EXISTS respaldo_videocall_tickets;
--    DROP TABLE IF EXISTS respaldo_instructor_schedules;
--    DROP TABLE IF EXISTS respaldo_invitacion_ticket;
