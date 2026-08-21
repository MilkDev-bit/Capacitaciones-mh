package repository

import (
	"context"
	"database/sql"
	"errors"
)

// ─────────────────────────────────────────────────────────────────────────────
// Clientes de Stripe
//
// Existe por la transferencia bancaria (SPEI). La documentación de Stripe es
// explícita: habilitar transferencias en Checkout exige indicar el `customer`
// en la sesión. No basta con `customer_creation`, que crea el cliente DESPUÉS
// para compras de invitado: la CLABE de referencia se emite a nombre de un
// cliente que ya debe existir cuando la sesión se crea.
//
// El identificador se guarda para REUTILIZARLO. Crear un Customer por compra
// dejaría al mismo alumno con un cliente distinto en cada pedido, y como las
// transferencias se acreditan contra el saldo del cliente, sus fondos quedarían
// repartidos entre varios y la conciliación sería imposible.
// ─────────────────────────────────────────────────────────────────────────────

type ClientesStripeRepository interface {
	// ClienteStripeDe devuelve el cliente guardado, o "" si no tiene.
	ClienteStripeDe(ctx context.Context, userID string) (string, error)
	// GuardarClienteStripe asocia un cliente a un usuario.
	//
	// Si el usuario ya tenía uno, se CONSERVA el anterior y se devuelve ese.
	// Dos peticiones simultáneas del mismo usuario pueden crear dos clientes en
	// Stripe; que gane siempre el primero evita que el segundo pisara al
	// primero y dejara saldos huérfanos en el cliente descartado.
	GuardarClienteStripe(ctx context.Context, userID, customerID string) (string, error)
}

func (r *postgresCursosRepository) ClienteStripeDe(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", errors.New("user_id vacío")
	}
	var id string
	err := r.db.GetContext(ctx, &id,
		`SELECT stripe_customer_id FROM stripe_clientes WHERE user_id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *postgresCursosRepository) GuardarClienteStripe(ctx context.Context, userID, customerID string) (string, error) {
	if userID == "" || customerID == "" {
		return "", errors.New("user_id o customer_id vacíos")
	}

	// DO UPDATE con el valor existente en vez de DO NOTHING: así el RETURNING
	// devuelve fila siempre, y quien llama recibe el cliente que quedó —el
	// suyo o el que ya estaba— sin tener que consultar otra vez.
	var guardado string
	err := r.db.GetContext(ctx, &guardado, `
		INSERT INTO stripe_clientes (user_id, stripe_customer_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET stripe_customer_id = stripe_clientes.stripe_customer_id
		RETURNING stripe_customer_id`, userID, customerID)
	if err != nil {
		return "", err
	}
	return guardado, nil
}
