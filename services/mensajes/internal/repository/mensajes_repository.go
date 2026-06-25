package repository

import (
	"context"
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
}

func (m *Mensaje) ToProto() *mensajespb.MensajeResponse {
	return &mensajespb.MensajeResponse{
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
	}
}

type Conversacion struct {
	PeerID      string    `db:"peer_id"`
	PeerName    string    `db:"peer_name"`
	LastMessage string    `db:"last_message"`
	LastTime    time.Time `db:"last_time"`
	UnreadCount int32     `db:"unread_count"`
	IsGroup     bool      `db:"is_group"`
}

func (c *Conversacion) ToProto() *mensajespb.ConversacionResponse {
	return &mensajespb.ConversacionResponse{
		PeerId:      c.PeerID,
		PeerName:    c.PeerName,
		LastMessage: c.LastMessage,
		LastTime:    c.LastTime.UTC().Format("2006-01-02T15:04:05Z"),
		UnreadCount: c.UnreadCount,
		IsGroup:     c.IsGroup,
	}
}

type MensajesRepository interface {
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
	
	condition := `((emisor_id = $1 AND receptor_id = $2) OR (emisor_id = $2 AND receptor_id = $1)) AND is_group = false`
	if isGroup {
		condition = `receptor_id = $2 AND is_group = true`
	}

	if beforeID == "" {
		query := fmt.Sprintf(`
SELECT * FROM (
SELECT id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at, attachment_url, attachment_type, is_group
FROM mensajes
WHERE %s
ORDER BY created_at DESC
LIMIT $3
) t ORDER BY t.created_at ASC
`, condition)
		err = r.db.SelectContext(ctx, &msgs, query, userID, peerID, fetch)
	} else {
		query := fmt.Sprintf(`
SELECT * FROM (
SELECT id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at, attachment_url, attachment_type, is_group
FROM mensajes
WHERE %s
  AND created_at < (SELECT created_at FROM mensajes WHERE id = $4::uuid)
ORDER BY created_at DESC
LIMIT $3
) t ORDER BY t.created_at ASC
`, condition)
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
ranked AS (
SELECT
CASE WHEN is_group THEN receptor_id ELSE (CASE WHEN emisor_id = $1 THEN receptor_id ELSE emisor_id END) END AS peer_id,
CASE WHEN is_group THEN receptor_name ELSE (CASE WHEN emisor_id = $1 THEN receptor_name ELSE emisor_name END) END AS peer_name,
contenido   AS last_message,
created_at  AS last_time,
is_group,
ROW_NUMBER() OVER (
PARTITION BY
CASE WHEN is_group THEN receptor_id::text ELSE
LEAST   (emisor_id::text, receptor_id::text) || GREATEST(emisor_id::text, receptor_id::text) END
ORDER BY created_at DESC
) AS rn
FROM mensajes
WHERE (is_group = false AND (emisor_id = $1 OR receptor_id = $1))
   OR (is_group = true AND receptor_id IN (SELECT grupo_id FROM group_memberships))
)
SELECT
r.peer_id,
r.peer_name,
r.last_message,
r.last_time,
r.is_group,
(
CASE WHEN r.is_group THEN 0 ELSE (
SELECT COUNT(*)::int
FROM mensajes
WHERE receptor_id = $1 AND emisor_id = r.peer_id AND leido = FALSE AND is_group = FALSE
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
	err := r.db.QueryRowxContext(ctx,
		`SELECT COUNT(*)::int FROM mensajes WHERE receptor_id = $1 AND leido = FALSE AND is_group = FALSE`,
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
