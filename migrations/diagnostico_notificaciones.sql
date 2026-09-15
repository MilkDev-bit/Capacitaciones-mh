-- ============================================================================
-- Verificación del arreglo de CreateNotificacion (SQLSTATE 42P08)
-- ============================================================================
--
-- Ejecutar en la base de USUARIOS. No escribe nada: solo prepara sentencias.
--
-- EL FALLO
--
--   ERROR: inconsistent types deduced for parameter $2 (SQLSTATE 42P08)
--
-- En un `INSERT ... SELECT` —a diferencia de un `INSERT ... VALUES`— el SELECT
-- se analiza por su cuenta: los tipos de las columnas destino NO se propagan a
-- los parámetros. Así que `$2` en la lista del SELECT se deducía como `text`,
-- mientras que `tipo = $2` en la subconsulta de deduplicación lo deducía como
-- `character varying`, que es el tipo de la columna. Dos deducciones para el
-- mismo parámetro y Postgres rechaza la sentencia.
--
-- Falla en Parse, antes de ejecutar nada. Por eso NINGUNA notificación se creó
-- nunca: ni de mensaje, ni de compra, ni de inscripción, ni de constancia.
--
-- IMPORTANTE: este error solo aparece con parámetros ($1, $2…), no con valores
-- literales. Una versión anterior de este script insertaba con literales y
-- habría pasado limpio, señalando en la dirección equivocada. Por eso aquí se
-- usa PREPARE, que es lo que hace el driver por debajo.
-- ============================================================================

-- ── PASO 1: reproducir el fallo ────────────────────────────────────────────
--
-- Esta es la consulta TAL COMO ESTABA. Debe fallar con 42P08.
-- Si falla, el diagnóstico está confirmado.
PREPARE notif_rota (uuid, text, text, text, text, int) AS
	INSERT INTO notificaciones (user_id, tipo, titulo, mensaje, enlace)
	SELECT $1::uuid, $2, $3, $4, NULLIF($5, '')
	WHERE $6 <= 0 OR NOT EXISTS (
		SELECT 1 FROM notificaciones
		 WHERE user_id = $1::uuid
		   AND tipo    = $2
		   AND titulo  = $3
		   AND mensaje = $4
		   AND COALESCE(enlace, '') = $5
		   AND leida = false
		   AND created_at > NOW() - make_interval(secs => $6::double precision)
	)
	RETURNING id;

-- ── PASO 2: comprobar el arreglo ───────────────────────────────────────────
--
-- La misma consulta con cast explícito en CADA aparición de cada parámetro.
-- Debe preparar sin error: es lo único que cambia.
PREPARE notif_arreglada (uuid, text, text, text, text, int) AS
	INSERT INTO notificaciones (user_id, tipo, titulo, mensaje, enlace)
	SELECT $1::uuid, $2::text, $3::text, $4::text, NULLIF($5::text, '')
	WHERE $6::int <= 0 OR NOT EXISTS (
		SELECT 1 FROM notificaciones
		 WHERE user_id = $1::uuid
		   AND tipo    = $2::text
		   AND titulo  = $3::text
		   AND mensaje = $4::text
		   AND COALESCE(enlace, '') = $5::text
		   AND leida = false
		   AND created_at > NOW() - make_interval(secs => $6::int::double precision)
	)
	RETURNING id;

DEALLOCATE notif_arreglada;

-- ── PASO 3: ¿hay avisos guardados? ─────────────────────────────────────────
--
-- Debería salir vacío para todos los usuarios: la tabla nunca recibió una fila
-- por esta vía. Después de desplegar, repetirlo confirma que ya entran.
SELECT tipo, count(*) AS total, max(created_at) AS ultima
  FROM notificaciones
 GROUP BY tipo
 ORDER BY total DESC;
