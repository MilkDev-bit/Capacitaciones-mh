-- ============================================================================
-- Retiro del flujo de capacitaciones por videollamada
--
-- NO se ejecuta sola. La aplicación ya dejó de leer y escribir estas tablas,
-- así que puedes desplegar el código primero y correr esto cuando quieras.
--
-- ⚠️  DESTRUCTIVO: borra el historial de horarios y de tickets de acceso.
--     Si alguna licencia pagada tiene tickets asociados, esa trazabilidad se
--     pierde. Respalda antes:
--
--       pg_dump -t instructor_schedules -t videocall_tickets \
--               -d "$DATABASE_URL" > respaldo_videollamadas.sql
--
-- Ejecución:
--       psql "$DATABASE_URL" -f migrations/001_retirar_videollamadas.sql
--
-- ¿Sin acceso a terminal (Railway, Neon, Supabase)? Usa la variante de consola:
--       migrations/001_retirar_videollamadas_railway.sql
--   Hace el respaldo con SQL puro, sin depender de pg_dump.
-- ============================================================================

BEGIN;

-- 1. Soltar la referencia desde licencia_invitaciones.
--    Sin esto el DROP de videocall_tickets falla por la llave foránea.
ALTER TABLE IF EXISTS licencia_invitaciones DROP COLUMN IF EXISTS ticket_id;

-- 2. Despublicar los cursos que quedaron sin flujo utilizable.
--    Se conserva el contenido, las inscripciones y el historial de compras:
--    solo dejan de ofrecerse en la tienda y el catálogo.
UPDATE capacitaciones
   SET is_public = false
 WHERE type = 'videocall'
   AND is_public = true;

-- 3. Eliminar las tablas del flujo.
--    videocall_tickets primero: referencia a instructor_schedules.
DROP TABLE IF EXISTS videocall_tickets;
DROP TABLE IF EXISTS instructor_schedules;

-- 4. Columna de estado de la videollamada en capacitaciones.
ALTER TABLE capacitaciones DROP COLUMN IF EXISTS videocall_status;

COMMIT;

-- ── Opcional ────────────────────────────────────────────────────────────────
-- Reclasificar los cursos retirados como 'video' para que vuelvan a ser
-- publicables como material autoformativo. Revisa uno por uno antes:
--
--   SELECT id, title, is_public FROM capacitaciones WHERE type = 'videocall';
--   UPDATE capacitaciones SET type = 'video' WHERE type = 'videocall';
