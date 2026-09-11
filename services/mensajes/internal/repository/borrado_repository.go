package repository

import (
	"context"
	"database/sql"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Borrado de mensajes y conversaciones
//
// Nada se borra de verdad. Lo que se guarda es quién dejó de ver qué:
//
//   · mensajes.eliminado_at    → borrado para todos (deja lápida)
//   · mensajes_ocultos         → borrado para mí, mensaje a mensaje
//   · conversaciones_ocultas   → borrado para mí, de toda la conversación
//
// La fila original nunca se toca, porque una plataforma de capacitación tiene
// que poder responder a una queja meses después.
// ─────────────────────────────────────────────────────────────────────────────

// MensajeParaBorrar son los datos mínimos para decidir quién puede borrar qué.
// No trae el contenido: quien autoriza no necesita leerlo.
type MensajeParaBorrar struct {
	EmisorID   string       `db:"emisor_id"`
	ReceptorID string       `db:"receptor_id"`
	IsGroup    bool         `db:"is_group"`
	CreatedAt  time.Time    `db:"created_at"`
	Eliminado  sql.NullTime `db:"eliminado_at"`
}

type BorradoRepository interface {
	// BuscarParaBorrar devuelve sql.ErrNoRows si el mensaje no existe.
	BuscarParaBorrar(ctx context.Context, msgID string) (*MensajeParaBorrar, error)
	// OcultarMensaje lo esconde solo para userID. Es idempotente.
	OcultarMensaje(ctx context.Context, msgID, userID string) error
	// EliminarParaTodos marca el mensaje. Devuelve false si ya estaba marcado.
	EliminarParaTodos(ctx context.Context, msgID, userID string) (bool, error)
	// OcultarConversacion corta el historial de userID con peerID en NOW().
	OcultarConversacion(ctx context.Context, userID, peerID string, isGroup bool) error
}

func (r *postgresMensajesRepository) BuscarParaBorrar(ctx context.Context, msgID string) (*MensajeParaBorrar, error) {
	m := &MensajeParaBorrar{}
	err := r.db.GetContext(ctx, m, `
SELECT emisor_id::text, receptor_id::text, is_group, created_at, eliminado_at
  FROM mensajes WHERE id = $1::uuid`, msgID)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *postgresMensajesRepository) OcultarMensaje(ctx context.Context, msgID, userID string) error {
	// ON CONFLICT DO NOTHING y no un error: pulsar dos veces "eliminar para mí"
	// —o hacerlo desde dos pestañas— debe dar el mismo resultado que una.
	_, err := r.db.ExecContext(ctx, `
INSERT INTO mensajes_ocultos (mensaje_id, usuario_id)
VALUES ($1::uuid, $2::uuid)
ON CONFLICT DO NOTHING`, msgID, userID)
	return err
}

func (r *postgresMensajesRepository) EliminarParaTodos(ctx context.Context, msgID, userID string) (bool, error) {
	// El WHERE eliminado_at IS NULL evita pisar quién lo borró primero. Con dos
	// pestañas abiertas la segunda no cambia nada y devuelve false, que el
	// servicio trata como éxito: el mensaje ya está borrado, que es lo pedido.
	res, err := r.db.ExecContext(ctx, `
UPDATE mensajes
   SET eliminado_at = NOW(), eliminado_por = $2::uuid
 WHERE id = $1::uuid AND eliminado_at IS NULL`, msgID, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *postgresMensajesRepository) OcultarConversacion(ctx context.Context, userID, peerID string, isGroup bool) error {
	// Upsert con oculta_desde = NOW(): si ya la había borrado antes, el corte
	// se mueve hacia adelante y vuelve a quedar vacía. Sin el DO UPDATE, borrar
	// una conversación por segunda vez no haría nada.
	_, err := r.db.ExecContext(ctx, `
INSERT INTO conversaciones_ocultas (usuario_id, peer_id, is_group, oculta_desde)
VALUES ($1::uuid, $2::uuid, $3, NOW())
ON CONFLICT (usuario_id, peer_id)
DO UPDATE SET oculta_desde = NOW(), is_group = EXCLUDED.is_group`,
		userID, peerID, isGroup)
	return err
}
