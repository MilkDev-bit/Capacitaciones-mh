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
	}
}

type Conversacion struct {
	PeerID      string    `db:"peer_id"`
	PeerName    string    `db:"peer_name"`
	LastMessage string    `db:"last_message"`
	LastTime    time.Time `db:"last_time"`
	UnreadCount int32     `db:"unread_count"`
}

func (c *Conversacion) ToProto() *mensajespb.ConversacionResponse {
	return &mensajespb.ConversacionResponse{
		PeerId:      c.PeerID,
		PeerName:    c.PeerName,
		LastMessage: c.LastMessage,
		LastTime:    c.LastTime.UTC().Format("2006-01-02T15:04:05Z"),
		UnreadCount: c.UnreadCount,
	}
}

type MensajesRepository interface {
	Send(ctx context.Context, m *Mensaje) (*Mensaje, error)
	// limit=0 uses default (50). Returns (messages, hasMore, error).
	GetConversacion(ctx context.Context, userID, peerID string, limit int, beforeID string) ([]*Mensaje, bool, error)
	MarcarLeidos(ctx context.Context, userID, peerID string) error
	// MarcarLeido marks a single message as read. Returns emisorID (or "" if already read/not found).
	MarcarLeido(ctx context.Context, msgID, userID string) (string, error)
	ListConversaciones(ctx context.Context, userID string) ([]*Conversacion, error)
	NoLeidos(ctx context.Context, userID string) (int32, error)
}

type postgresMensajesRepository struct{ db *sqlx.DB }

func NewMensajesRepository(db *sqlx.DB) MensajesRepository {
	return &postgresMensajesRepository{db: db}
}

func (r *postgresMensajesRepository) Send(ctx context.Context, m *Mensaje) (*Mensaje, error) {
	out := &Mensaje{}
	err := r.db.QueryRowxContext(ctx, `
INSERT INTO mensajes (emisor_id, emisor_name, receptor_id, receptor_name, contenido, attachment_url, attachment_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at, attachment_url, attachment_type
`, m.EmisorID, m.EmisorName, m.ReceptorID, m.ReceptorName, m.Contenido, m.AttachmentUrl, m.AttachmentType).StructScan(out)
	return out, err
}

func (r *postgresMensajesRepository) GetConversacion(ctx context.Context, userID, peerID string, limit int, beforeID string) ([]*Mensaje, bool, error) {
	if limit <= 0 {
		limit = 50
	}
	fetch := limit + 1 // one extra to detect has_more

	var msgs []*Mensaje
	var err error

	if beforeID == "" {
		err = r.db.SelectContext(ctx, &msgs, `
SELECT * FROM (
SELECT id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at, attachment_url, attachment_type
FROM mensajes
WHERE (emisor_id = $1 AND receptor_id = $2)
   OR (emisor_id = $2 AND receptor_id = $1)
ORDER BY created_at DESC
LIMIT $3
) t ORDER BY t.created_at ASC
`, userID, peerID, fetch)
	} else {
		err = r.db.SelectContext(ctx, &msgs, `
SELECT * FROM (
SELECT id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at, attachment_url, attachment_type
FROM mensajes
WHERE ((emisor_id = $1 AND receptor_id = $2) OR (emisor_id = $2 AND receptor_id = $1))
  AND created_at < (SELECT created_at FROM mensajes WHERE id = $4::uuid)
ORDER BY created_at DESC
LIMIT $3
) t ORDER BY t.created_at ASC
`, userID, peerID, fetch, beforeID)
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

func (r *postgresMensajesRepository) MarcarLeidos(ctx context.Context, userID, peerID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE mensajes SET leido = TRUE WHERE receptor_id = $1 AND emisor_id = $2 AND leido = FALSE`,
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
WITH ranked AS (
SELECT
CASE WHEN emisor_id = $1 THEN receptor_id   ELSE emisor_id   END AS peer_id,
CASE WHEN emisor_id = $1 THEN receptor_name ELSE emisor_name END AS peer_name,
contenido   AS last_message,
created_at  AS last_time,
ROW_NUMBER() OVER (
PARTITION BY
LEAST   (emisor_id::text, receptor_id::text),
GREATEST(emisor_id::text, receptor_id::text)
ORDER BY created_at DESC
) AS rn
FROM mensajes
WHERE emisor_id = $1 OR receptor_id = $1
)
SELECT
r.peer_id,
r.peer_name,
r.last_message,
r.last_time,
(
SELECT COUNT(*)::int
FROM mensajes
WHERE receptor_id = $1 AND emisor_id = r.peer_id AND leido = FALSE
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
		`SELECT COUNT(*)::int FROM mensajes WHERE receptor_id = $1 AND leido = FALSE`,
		userID,
	).Scan(&count)
	return count, err
}
