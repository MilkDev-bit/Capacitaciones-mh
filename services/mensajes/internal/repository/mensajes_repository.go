package repository

import (
	"context"
	"time"

	mensajespb "Prueba-Go/gen/mensajes"

	"github.com/jmoiron/sqlx"
)

// ─── Modelos internos ─────────────────────────────────────────────────────────

type Mensaje struct {
	ID           string    `db:"id"`
	EmisorID     string    `db:"emisor_id"`
	EmisorName   string    `db:"emisor_name"`
	ReceptorID   string    `db:"receptor_id"`
	ReceptorName string    `db:"receptor_name"`
	Contenido    string    `db:"contenido"`
	Leido        bool      `db:"leido"`
	CreatedAt    time.Time `db:"created_at"`
}

func (m *Mensaje) ToProto() *mensajespb.MensajeResponse {
	return &mensajespb.MensajeResponse{
		Id:           m.ID,
		EmisorId:     m.EmisorID,
		EmisorName:   m.EmisorName,
		ReceptorId:   m.ReceptorID,
		ReceptorName: m.ReceptorName,
		Contenido:    m.Contenido,
		Leido:        m.Leido,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
		LastTime:    c.LastTime.Format("2006-01-02T15:04:05Z"),
		UnreadCount: c.UnreadCount,
	}
}

// ─── Interfaz y PostgreSQL ────────────────────────────────────────────────────

type MensajesRepository interface {
	Send(ctx context.Context, m *Mensaje) (*Mensaje, error)
	GetConversacion(ctx context.Context, userID, peerID string) ([]*Mensaje, error)
	MarcarLeidos(ctx context.Context, userID, peerID string) error
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
		INSERT INTO mensajes (emisor_id, emisor_name, receptor_id, receptor_name, contenido)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at
	`, m.EmisorID, m.EmisorName, m.ReceptorID, m.ReceptorName, m.Contenido).StructScan(out)
	return out, err
}

func (r *postgresMensajesRepository) GetConversacion(ctx context.Context, userID, peerID string) ([]*Mensaje, error) {
	var msgs []*Mensaje
	err := r.db.SelectContext(ctx, &msgs, `
		SELECT id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at
		FROM mensajes
		WHERE (emisor_id = $1 AND receptor_id = $2)
		   OR (emisor_id = $2 AND receptor_id = $1)
		ORDER BY created_at ASC
	`, userID, peerID)
	return msgs, err
}

func (r *postgresMensajesRepository) MarcarLeidos(ctx context.Context, userID, peerID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE mensajes SET leido = TRUE WHERE receptor_id = $1 AND emisor_id = $2 AND leido = FALSE`,
		userID, peerID,
	)
	return err
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
