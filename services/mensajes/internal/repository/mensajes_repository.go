package repository

import (
	"context"
	"fmt"
	"time"

	mensajespb "Prueba-Go/gen/mensajes"

	"github.com/jmoiron/sqlx"
)

type Mensaje struct {
	ID             string    `db:"id"`
	EmisorID       string    `db:"emisor_id"`
	EmisorName     string    `db:"emisor_name"`
	ReceptorID     string    `db:"receptor_id"`
	ReceptorName   string    `db:"receptor_name"`
	Contenido      string    `db:"contenido"`
	Leido          bool      `db:"leido"`
	CreatedAt      time.Time `db:"created_at"`
	AttachmentUrl  string    `db:"attachment_url"`
	AttachmentType string    `db:"attachment_type"`
	IsGroup        bool      `db:"is_group"`
	Eliminado      bool      `db:"eliminado"`
}

func (m *Mensaje) ToProto() *mensajespb.MensajeResponse {
	resp := &mensajespb.MensajeResponse{
		Id:             m.ID,
		EmisorId:       m.EmisorID,
		EmisorName:     m.EmisorName,
		ReceptorId:     m.ReceptorID,
		ReceptorName:   m.ReceptorName,
		Contenido:      m.Contenido,
		Leido:          m.Leido,
		CreatedAt:      m.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		AttachmentUrl:  m.AttachmentUrl,
		AttachmentType: m.AttachmentType,
		IsGroup:        m.IsGroup,
		Eliminado:      m.Eliminado,
	}

	// El texto de un mensaje borrado no sale del servidor.
	//
	// Va aquí, en la conversión a proto, y no en el SELECT: así ninguna
	// consulta futura puede olvidarse de censurarlo. La fila conserva el
	// contenido en la base para una eventual investigación, pero el cliente
	// solo recibe la lápida.
	if m.Eliminado {
		resp.Contenido = ""
		resp.AttachmentUrl = ""
		resp.AttachmentType = ""
	}
	return resp
}

type Conversacion struct {
	PeerID        string    `db:"peer_id"`
	PeerName      string    `db:"peer_name"`
	LastMessage   string    `db:"last_message"`
	LastTime      time.Time `db:"last_time"`
	UnreadCount   int32     `db:"unread_count"`
	IsGroup       bool      `db:"is_group"`
	LastEliminado bool      `db:"last_eliminado"`
}

func (c *Conversacion) ToProto() *mensajespb.ConversacionResponse {
	return &mensajespb.ConversacionResponse{
		PeerId:        c.PeerID,
		PeerName:      c.PeerName,
		LastMessage:   c.LastMessage,
		LastTime:      c.LastTime.UTC().Format("2006-01-02T15:04:05Z"),
		UnreadCount:   c.UnreadCount,
		IsGroup:       c.IsGroup,
		LastEliminado: c.LastEliminado,
	}
}

type MensajesRepository interface {
	// Borrado al estilo WhatsApp. Se embebe en vez de pasarse aparte porque lo
	// implementa el mismo tipo y el servicio ya recibe esta interfaz entera.
	BorradoRepository

	Send(ctx context.Context, m *Mensaje) (*Mensaje, error)
	// limit=0 uses default (50). Returns (messages, hasMore, error).
	GetConversacion(ctx context.Context, userID, peerID string, limit int, beforeID string, isGroup bool) ([]*Mensaje, bool, error)
	MarcarLeidos(ctx context.Context, userID, peerID string, isGroup bool) error
	// MarcarLeido marks a single message as read. Returns emisorID (or "" if already read/not found).
	MarcarLeido(ctx context.Context, msgID, userID string) (string, error)
	ListConversaciones(ctx context.Context, userID string) ([]*Conversacion, error)
	NoLeidos(ctx context.Context, userID string) (int32, error)
	CreateGroup(ctx context.Context, nombre, adminID string) (string, error)
	AddGroupMembers(ctx context.Context, grupoID string, userIDs []string) error
	GetGroupMembers(ctx context.Context, grupoID string) ([]string, error)
	// Licencia-linked groups
	CreateGroupForLicencia(ctx context.Context, nombre, adminID, licenciaID string) (string, error)
	GetGroupIDByLicencia(ctx context.Context, licenciaID string) (string, error)
}

type postgresMensajesRepository struct{ db *sqlx.DB }

func NewMensajesRepository(db *sqlx.DB) MensajesRepository {
	return &postgresMensajesRepository{db: db}
}

func (r *postgresMensajesRepository) Send(ctx context.Context, m *Mensaje) (*Mensaje, error) {
	out := &Mensaje{}
	err := r.db.QueryRowxContext(ctx, `
INSERT INTO mensajes (emisor_id, emisor_name, receptor_id, receptor_name, contenido, attachment_url, attachment_type, is_group)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at, attachment_url, attachment_type, is_group
`, m.EmisorID, m.EmisorName, m.ReceptorID, m.ReceptorName, m.Contenido, m.AttachmentUrl, m.AttachmentType, m.IsGroup).StructScan(out)
	return out, err
}

func (r *postgresMensajesRepository) GetConversacion(ctx context.Context, userID, peerID string, limit int, beforeID string, isGroup bool) ([]*Mensaje, bool, error) {
	if limit <= 0 {
		limit = 50
	}
	fetch := limit + 1 // one extra to detect has_more

	var msgs []*Mensaje
	var err error

	condition := `((m.emisor_id = $1 AND m.receptor_id = $2) OR (m.emisor_id = $2 AND m.receptor_id = $1)) AND m.is_group = false`
	if isGroup {
		condition = `m.receptor_id = $2 AND m.is_group = true`
	}

	// Los dos filtros de borrado, siempre. Van pegados a la condición y no
	// sueltos por la consulta para que no se puedan añadir ramas nuevas que se
	// olviden de uno de ellos.
	//
	//  · visible: lo que ESTE usuario ocultó uno a uno.
	//  · corte:   si borró la conversación, todo lo anterior a ese instante
	//             desapareció para él. Lo posterior vuelve a verse solo.
	const visible = `
  AND NOT EXISTS (
      SELECT 1 FROM mensajes_ocultos o
       WHERE o.mensaje_id = m.id AND o.usuario_id = $1
  )
  AND m.created_at > COALESCE(
      (SELECT c.oculta_desde FROM conversaciones_ocultas c
        WHERE c.usuario_id = $1 AND c.peer_id = $2),
      '-infinity'::timestamptz
  )`

	const columnas = `m.id, m.emisor_id, m.emisor_name, m.receptor_id, m.receptor_name,
       m.contenido, m.leido, m.created_at, m.attachment_url, m.attachment_type, m.is_group,
       (m.eliminado_at IS NOT NULL) AS eliminado`

	if beforeID == "" {
		query := fmt.Sprintf(`
SELECT * FROM (
SELECT %s
FROM mensajes m
WHERE %s%s
ORDER BY m.created_at DESC
LIMIT $3
) t ORDER BY t.created_at ASC
`, columnas, condition, visible)
		err = r.db.SelectContext(ctx, &msgs, query, userID, peerID, fetch)
	} else {
		query := fmt.Sprintf(`
SELECT * FROM (
SELECT %s
FROM mensajes m
WHERE %s%s
  AND m.created_at < (SELECT created_at FROM mensajes WHERE id = $4::uuid)
ORDER BY m.created_at DESC
LIMIT $3
) t ORDER BY t.created_at ASC
`, columnas, condition, visible)
		err = r.db.SelectContext(ctx, &msgs, query, userID, peerID, fetch, beforeID)
	}
	if err != nil {
		return nil, false, err
	}

	hasMore := len(msgs) == fetch
	if hasMore {
		msgs = msgs[1:] // remove the extra oldest element
	}
	return msgs, hasMore, nil
}

func (r *postgresMensajesRepository) MarcarLeidos(ctx context.Context, userID, peerID string, isGroup bool) error {
	if isGroup {
		// En grupos, el leido=TRUE es complejo porque hay muchos usuarios. 
		// Por simplicidad, asumimos que los mensajes de grupo siempre se leen al recibirlos o no rastreamos lectura individual aquí
		return nil
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE mensajes SET leido = TRUE WHERE receptor_id = $1 AND emisor_id = $2 AND leido = FALSE AND is_group = FALSE`,
		userID, peerID,
	)
	return err
}

func (r *postgresMensajesRepository) MarcarLeido(ctx context.Context, msgID, userID string) (string, error) {
	var emisorID string
	err := r.db.QueryRowxContext(ctx, `
UPDATE mensajes SET leido = TRUE
WHERE id = $1::uuid AND receptor_id = $2::uuid AND leido = FALSE
RETURNING emisor_id
`, msgID, userID).Scan(&emisorID)
	if err != nil {
		// No rows updated = already read or not found; not an error
		return "", nil
	}
	return emisorID, nil
}

func (r *postgresMensajesRepository) ListConversaciones(ctx context.Context, userID string) ([]*Conversacion, error) {
	var convs []*Conversacion
	err := r.db.SelectContext(ctx, &convs, `
WITH group_memberships AS (
	SELECT grupo_id FROM grupo_miembros WHERE usuario_id = $1
),
-- Todo lo que este usuario puede ver, con el interlocutor ya resuelto.
-- Calcular "peer" aquí y no en cada rama permite cruzarlo después contra
-- conversaciones_ocultas, que se indexa por peer.
mios AS (
SELECT m.id, m.emisor_id, m.receptor_id, m.contenido, m.leido, m.created_at,
       m.is_group, m.eliminado_at,
       CASE WHEN m.is_group THEN m.receptor_id
            ELSE (CASE WHEN m.emisor_id = $1 THEN m.receptor_id ELSE m.emisor_id END) END AS peer,
       CASE WHEN m.is_group THEN m.receptor_name
            ELSE (CASE WHEN m.emisor_id = $1 THEN m.receptor_name ELSE m.emisor_name END) END AS peer_nombre
FROM mensajes m
WHERE ((m.is_group = false AND (m.emisor_id = $1 OR m.receptor_id = $1))
    OR (m.is_group = true AND m.receptor_id IN (SELECT grupo_id FROM group_memberships)))
  AND NOT EXISTS (
      SELECT 1 FROM mensajes_ocultos o
       WHERE o.mensaje_id = m.id AND o.usuario_id = $1)
),
-- Corte por conversación borrada. Si no queda ningún mensaje posterior, la
-- conversación desaparece sola de la lista; no hace falta marcarla como
-- borrada ni volver a activarla cuando llegue uno nuevo.
visibles AS (
SELECT v.* FROM mios v
LEFT JOIN conversaciones_ocultas c ON c.usuario_id = $1 AND c.peer_id = v.peer
WHERE v.created_at > COALESCE(c.oculta_desde, '-infinity'::timestamptz)
),
ranked AS (
SELECT
peer        AS peer_id,
peer_nombre AS peer_name,
CASE WHEN eliminado_at IS NOT NULL THEN '' ELSE contenido END AS last_message,
(eliminado_at IS NOT NULL) AS last_eliminado,
created_at  AS last_time,
is_group,
ROW_NUMBER() OVER (PARTITION BY peer, is_group ORDER BY created_at DESC) AS rn
FROM visibles
)
SELECT
r.peer_id,
r.peer_name,
r.last_message,
r.last_eliminado,
r.last_time,
r.is_group,
(
CASE WHEN r.is_group THEN 0 ELSE (
SELECT COUNT(*)::int
FROM visibles v
WHERE v.receptor_id = $1 AND v.emisor_id = r.peer_id AND v.leido = FALSE AND v.is_group = FALSE
) END
) AS unread_count
FROM ranked r
WHERE r.rn = 1
ORDER BY r.last_time DESC
`, userID)
	return convs, err
}

func (r *postgresMensajesRepository) NoLeidos(ctx context.Context, userID string) (int32, error) {
	var count int32
	// Mismo filtro de borrado que la lista. Sin él, la campana seguiría
	// anunciando mensajes que el usuario ya no puede abrir por ningún lado, y
	// el contador no bajaría nunca. En un mensaje recibido y directo el
	// interlocutor es siempre el emisor, así que el cruce es directo.
	err := r.db.QueryRowxContext(ctx, `
SELECT COUNT(*)::int
  FROM mensajes m
  LEFT JOIN conversaciones_ocultas c
         ON c.usuario_id = $1 AND c.peer_id = m.emisor_id
 WHERE m.receptor_id = $1
   AND m.leido = FALSE
   AND m.is_group = FALSE
   AND m.created_at > COALESCE(c.oculta_desde, '-infinity'::timestamptz)
   AND NOT EXISTS (
       SELECT 1 FROM mensajes_ocultos o
        WHERE o.mensaje_id = m.id AND o.usuario_id = $1)`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *postgresMensajesRepository) CreateGroup(ctx context.Context, nombre, adminID string) (string, error) {
	var id string
	err := r.db.QueryRowxContext(ctx, `INSERT INTO grupos (nombre, admin_id) VALUES ($1, $2) RETURNING id`, nombre, adminID).Scan(&id)
	return id, err
}

func (r *postgresMensajesRepository) AddGroupMembers(ctx context.Context, grupoID string, userIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `INSERT INTO grupo_miembros (grupo_id, usuario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, uid := range userIDs {
		if _, err := stmt.ExecContext(ctx, grupoID, uid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *postgresMensajesRepository) GetGroupMembers(ctx context.Context, grupoID string) ([]string, error) {
	var userIDs []string
	err := r.db.SelectContext(ctx, &userIDs, `SELECT usuario_id FROM grupo_miembros WHERE grupo_id = $1`, grupoID)
	return userIDs, err
}

func (r *postgresMensajesRepository) CreateGroupForLicencia(ctx context.Context, nombre, adminID, licenciaID string) (string, error) {
	var id string
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO grupos (nombre, admin_id, licencia_id) VALUES ($1, $2, $3) RETURNING id`,
		nombre, adminID, licenciaID).Scan(&id)
	return id, err
}

func (r *postgresMensajesRepository) GetGroupIDByLicencia(ctx context.Context, licenciaID string) (string, error) {
	var id string
	err := r.db.QueryRowxContext(ctx,
		`SELECT id FROM grupos WHERE licencia_id = $1 LIMIT 1`, licenciaID).Scan(&id)
	return id, err
}
