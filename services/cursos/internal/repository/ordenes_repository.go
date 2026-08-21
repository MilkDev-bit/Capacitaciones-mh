package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ─────────────────────────────────────────────────────────────────────────────
// Órdenes
//
// La orden se crea ANTES de mandar al usuario a Stripe. Sin ella no quedaba
// rastro del intento de cobro: si el webhook fallaba, no había forma de saber
// qué se intentó cobrar ni a quién.
//
// Su idempotency_key se deriva del contenido del carrito, no de un UUID por
// petición. Así un doble clic, un reintento del navegador y un reenvío del
// cliente resuelven todos a la MISMA orden y a la MISMA sesión de Stripe.
// ─────────────────────────────────────────────────────────────────────────────

// Estados posibles de una orden.
const (
	OrdenPendiente   = "pendiente"
	OrdenPagada      = "pagada"
	OrdenCumplida    = "cumplida"
	OrdenFallida     = "fallida"
	OrdenReembolsada = "reembolsada"
)

// ventanaSesionStripe es cuánto tiempo se reutiliza la sesión de una orden
// pendiente. Stripe expira sus Checkout Sessions a las 24 h; con 23 se evita
// devolver una URL que va a morir mientras el usuario la tiene abierta.
const ventanaSesionStripe = 23 * time.Hour

// Comision es lo que Stripe se quedó de un cobro, según su BalanceTransaction.
//
// Se pasa como puntero para poder distinguir "no lo sabemos todavía" (nil) de
// "lo sabemos y fue cero". Guardar un cero cuando en realidad no se ha
// consultado haría que esos cobros parecieran libres de comisión e inflaría la
// ganancia neta del panel.
type Comision struct {
	Centavos     int64
	NetoCentavos int64
	BalanceTxID  string
}

// ErrOrdenYaPagada indica que este carrito exacto ya se cobró.
var ErrOrdenYaPagada = errors.New("esta compra ya fue pagada")

// Orden es una intención de compra.
type Orden struct {
	ID                  string     `db:"id"`
	UserID              string     `db:"user_id"`
	Estado              string     `db:"estado"`
	TotalCentavos       int64      `db:"total_centavos"`
	Moneda              string     `db:"moneda"`
	StripeSessionID     *string    `db:"stripe_session_id"`
	StripePaymentIntent *string    `db:"stripe_payment_intent"`
	IdempotencyKey      string     `db:"idempotency_key"`
	Intento             int32      `db:"intento"`
	MotivoFallo         *string    `db:"motivo_fallo"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
	PagadaAt            *time.Time `db:"pagada_at"`
	CumplidaAt          *time.Time `db:"cumplida_at"`
}

// OrdenItem es un renglón con el precio congelado al momento de comprar.
type OrdenItem struct {
	CapacitacionID         string
	Tipo                   string // "b2c" | "b2b_direct"
	Titulo                 string
	Cantidad               int32
	PrecioUnitarioCentavos int64
	SubtotalCentavos       int64
}

// HuellaCarrito produce una clave estable a partir del usuario y su carrito.
//
// Los ítems se ordenan antes de hashear para que el mismo carrito en distinto
// orden dé la misma huella: de lo contrario, reordenar el carrito generaría una
// orden nueva y una segunda sesión de pago.
func HuellaCarrito(userID string, items []OrdenItem) string {
	partes := make([]string, 0, len(items))
	for _, it := range items {
		partes = append(partes, fmt.Sprintf("%s|%s|%d|%d",
			it.CapacitacionID, it.Tipo, it.Cantidad, it.PrecioUnitarioCentavos))
	}
	sort.Strings(partes)

	h := sha256.Sum256([]byte(userID + "::" + strings.Join(partes, ";")))
	return "carrito-" + hex.EncodeToString(h[:16])
}

// ── Contrato ─────────────────────────────────────────────────────────────────

// OrdenesRepository agrupa el acceso a datos de órdenes y eventos de Stripe.
type OrdenesRepository interface {
	// CrearOAbrirOrden devuelve la orden pendiente correspondiente a este
	// carrito, creándola si no existía. reutilizada=true significa que ya había
	// una sesión de Stripe viva que debe reaprovecharse en lugar de crear otra.
	CrearOAbrirOrden(ctx context.Context, userID string, totalCentavos int64, moneda string, items []OrdenItem) (orden *Orden, reutilizada bool, err error)
	// GuardarSesionStripe asocia la sesión recién creada con la orden.
	GuardarSesionStripe(ctx context.Context, ordenID, sessionID string) error
	// ActualizarEstadoOrden aplica una transición de estado. `com` solo se
	// usa al pasar a 'pagada' y puede ser nil cuando Stripe todavía no ha
	// liquidado el cobro.
	ActualizarEstadoOrden(ctx context.Context, sessionID, estado, motivoFallo, paymentIntent string, com *Comision) error
	// RegistrarEventoStripe deduplica: devuelve false si el evento ya se procesó.
	RegistrarEventoStripe(ctx context.Context, eventID, tipo string) (primeraVez bool, err error)
}

// ── Implementación ───────────────────────────────────────────────────────────

func (r *postgresCursosRepository) CrearOAbrirOrden(
	ctx context.Context, userID string, totalCentavos int64, moneda string, items []OrdenItem,
) (*Orden, bool, error) {
	if len(items) == 0 {
		return nil, false, errors.New("la orden no tiene renglones")
	}
	key := HuellaCarrito(userID, items)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op tras Commit

	// Un carrito idéntico del mismo usuario cae siempre en la misma fila.
	// El bloqueo serializa los dobles clics concurrentes.
	var existente Orden
	err = tx.GetContext(ctx, &existente,
		`SELECT id, user_id, estado, total_centavos, moneda, stripe_session_id,
		        stripe_payment_intent, idempotency_key, intento, motivo_fallo,
		        created_at, updated_at, pagada_at, cumplida_at
		   FROM ordenes WHERE idempotency_key = $1 FOR UPDATE`, key)

	switch {
	case err == nil:
		// Ya se cobró: no se permite recomprar exactamente lo mismo sin cambiar
		// el carrito. Evita el doble cargo por volver atrás en el navegador.
		if existente.Estado == OrdenPagada || existente.Estado == OrdenCumplida {
			return nil, false, ErrOrdenYaPagada
		}
		// Pendiente con sesión viva: se reutiliza en lugar de crear otra.
		if existente.Estado == OrdenPendiente &&
			existente.StripeSessionID != nil && *existente.StripeSessionID != "" &&
			time.Since(existente.CreatedAt) < ventanaSesionStripe {
			if err := tx.Commit(); err != nil {
				return nil, false, err
			}
			return &existente, true, nil
		}
		// Fallida o con sesión caduca: se reabre como intento nuevo.
		if _, err := tx.ExecContext(ctx,
			`UPDATE ordenes
			    SET estado='pendiente', intento=intento+1, stripe_session_id=NULL,
			        motivo_fallo=NULL, total_centavos=$2, updated_at=NOW(), created_at=NOW()
			  WHERE id=$1`, existente.ID, totalCentavos); err != nil {
			return nil, false, err
		}
		if err := reemplazarItems(ctx, tx, existente.ID, items); err != nil {
			return nil, false, err
		}
		existente.Estado = OrdenPendiente
		existente.Intento++
		existente.TotalCentavos = totalCentavos
		existente.StripeSessionID = nil
		if err := tx.Commit(); err != nil {
			return nil, false, err
		}
		return &existente, false, nil

	case errors.Is(err, sql.ErrNoRows):
		// Camino normal: orden nueva.
		var nueva Orden
		if err := tx.GetContext(ctx, &nueva,
			`INSERT INTO ordenes (user_id, estado, total_centavos, moneda, idempotency_key)
			 VALUES ($1, 'pendiente', $2, $3, $4)
			 RETURNING id, user_id, estado, total_centavos, moneda, stripe_session_id,
			           stripe_payment_intent, idempotency_key, intento, motivo_fallo,
			           created_at, updated_at, pagada_at, cumplida_at`,
			userID, totalCentavos, moneda, key); err != nil {
			return nil, false, err
		}
		if err := reemplazarItems(ctx, tx, nueva.ID, items); err != nil {
			return nil, false, err
		}
		if err := tx.Commit(); err != nil {
			return nil, false, err
		}
		return &nueva, false, nil

	default:
		return nil, false, err
	}
}

// reemplazarItems deja los renglones exactamente como vienen. Se borra y se
// reinserta porque un reintento puede traer cantidades distintas.
func reemplazarItems(ctx context.Context, tx *sqlx.Tx, ordenID string, items []OrdenItem) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM orden_items WHERE orden_id=$1`, ordenID); err != nil {
		return err
	}
	for _, it := range items {
		var capID interface{}
		if it.CapacitacionID != "" {
			capID = it.CapacitacionID
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO orden_items
			   (orden_id, capacitacion_id, tipo, titulo, cantidad,
			    precio_unitario_centavos, subtotal_centavos)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			ordenID, capID, it.Tipo, it.Titulo, it.Cantidad,
			it.PrecioUnitarioCentavos, it.SubtotalCentavos); err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresCursosRepository) GuardarSesionStripe(ctx context.Context, ordenID, sessionID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ordenes SET stripe_session_id=$1, updated_at=NOW() WHERE id=$2`,
		sessionID, ordenID)
	return err
}

// ActualizarEstadoOrden aplica la transición sin retroceder de estado.
//
// El webhook y verify-checkout-session compiten y pueden llegar en cualquier
// orden. Las guardas del WHERE evitan que una entrega tardía de 'pagada'
// sobrescriba un 'cumplida' que ya se aplicó.
func (r *postgresCursosRepository) ActualizarEstadoOrden(ctx context.Context, sessionID, estado, motivoFallo, paymentIntent string, com *Comision) error {
	if sessionID == "" {
		return errors.New("stripe_session_id vacío")
	}

	var pi, motivo interface{}
	if paymentIntent != "" {
		pi = paymentIntent
	}
	if motivoFallo != "" {
		motivo = motivoFallo
	}

	// nil cuando Stripe aún no ha liquidado el cobro. Los COALESCE de abajo
	// dejan entonces las columnas como estaban —en NULL la primera vez— y el
	// relleno del histórico las recoge más tarde.
	var comCentavos, comNeto, comTx interface{}
	if com != nil {
		comCentavos, comNeto = com.Centavos, com.NetoCentavos
		if com.BalanceTxID != "" {
			comTx = com.BalanceTxID
		}
	}

	switch estado {
	case OrdenPagada:
		_, err := r.db.ExecContext(ctx,
			`UPDATE ordenes
			    SET estado='pagada', pagada_at=COALESCE(pagada_at, NOW()),
			        stripe_payment_intent=COALESCE($2, stripe_payment_intent),
			        comision_centavos=COALESCE($3, comision_centavos),
			        neto_centavos=COALESCE($4, neto_centavos),
			        balance_transaction_id=COALESCE($5, balance_transaction_id),
			        updated_at=NOW()
			  WHERE stripe_session_id=$1 AND estado='pendiente'`,
			sessionID, pi, comCentavos, comNeto, comTx)
		return err

	case OrdenCumplida:
		// Se acepta desde 'pendiente' también: si el webhook nunca llegó pero
		// verify sí cumplió la compra, la orden debe cerrarse igual.
		_, err := r.db.ExecContext(ctx,
			`UPDATE ordenes
			    SET estado='cumplida',
			        pagada_at=COALESCE(pagada_at, NOW()),
			        cumplida_at=COALESCE(cumplida_at, NOW()),
			        stripe_payment_intent=COALESCE($2, stripe_payment_intent), updated_at=NOW()
			  WHERE stripe_session_id=$1 AND estado IN ('pendiente','pagada')`, sessionID, pi)
		return err

	case OrdenFallida:
		// Una orden ya pagada nunca vuelve a 'fallida'.
		_, err := r.db.ExecContext(ctx,
			`UPDATE ordenes
			    SET estado='fallida', motivo_fallo=$2, updated_at=NOW()
			  WHERE stripe_session_id=$1 AND estado='pendiente'`, sessionID, motivo)
		return err

	case OrdenReembolsada:
		_, err := r.db.ExecContext(ctx,
			`UPDATE ordenes SET estado='reembolsada', updated_at=NOW()
			  WHERE stripe_session_id=$1 AND estado IN ('pagada','cumplida')`, sessionID)
		return err

	default:
		return fmt.Errorf("estado de orden no válido: %s", estado)
	}
}

// RegistrarEventoStripe deduplica por clave primaria: el INSERT ... ON CONFLICT
// es atómico, así que dos entregas simultáneas del mismo evento solo dejan
// pasar a una.
func (r *postgresCursosRepository) RegistrarEventoStripe(ctx context.Context, eventID, tipo string) (bool, error) {
	if eventID == "" {
		return false, errors.New("event_id vacío")
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO stripe_eventos_procesados (event_id, tipo)
		 VALUES ($1, $2) ON CONFLICT (event_id) DO NOTHING`, eventID, tipo)
	if err != nil {
		return false, err
	}
	filas, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}
