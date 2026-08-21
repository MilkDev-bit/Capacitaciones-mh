package service

import (
	"context"
	"log/slog"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/customer"
	"google.golang.org/grpc/metadata"
)

// clienteStripeDe devuelve el Customer de Stripe del usuario, creándolo si no
// lo tenía.
//
// Hace falta para la transferencia bancaria: Checkout exige indicar el
// `customer` en la sesión cuando se ofrece `customer_balance`, porque la CLABE
// de referencia se emite a nombre de un cliente concreto.
//
// Devuelve "" si no se pudo obtener. NO es un error fatal para quien llama: la
// sesión debe crearse igual, sin transferencia, en vez de dejar al comprador
// sin poder pagar de ninguna forma. Perder la opción de transferir es
// molesto; perder el checkout entero es perder la venta.
func (s *CursosService) clienteStripeDe(ctx context.Context, userID, email, nombre string) string {
	if userID == "" {
		return ""
	}

	existente, err := s.repo.ClienteStripeDe(ctx, userID)
	if err != nil {
		slog.Warn("no se pudo leer el cliente de Stripe", "user_id", userID, "error", err)
		return ""
	}
	if existente != "" {
		return existente
	}

	params := &stripe.CustomerParams{}
	if email != "" {
		params.Email = stripe.String(email)
	}
	if nombre != "" {
		params.Name = stripe.String(nombre)
	}
	// Permite localizar al usuario desde el panel de Stripe cuando llega una
	// transferencia que hay que conciliar a mano.
	params.AddMetadata("user_id", userID)

	nuevo, err := customer.New(params)
	if err != nil || nuevo == nil {
		slog.Warn("no se pudo crear el cliente de Stripe", "user_id", userID, "error", err)
		return ""
	}

	// El guardado devuelve el cliente que quedó, que puede no ser el recién
	// creado si otra petición simultánea se adelantó. Se usa ese para que los
	// saldos de transferencia de este usuario vivan siempre en el mismo sitio.
	guardado, err := s.repo.GuardarClienteStripe(ctx, userID, nuevo.ID)
	if err != nil {
		slog.Warn("cliente de Stripe creado pero no guardado",
			"user_id", userID, "customer_id", nuevo.ID, "error", err)
		return nuevo.ID
	}
	return guardado
}

// datosDelComprador saca nombre y correo de la metadata gRPC que manda el
// gateway. En el servicio no hay JWT, así que es la única fuente disponible.
func datosDelComprador(ctx context.Context) (correo, nombre string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ""
	}
	if vals := md.Get("x-user-email"); len(vals) > 0 {
		correo = vals[0]
	}
	if vals := md.Get("x-user-name"); len(vals) > 0 {
		nombre = vals[0]
	}
	return correo, nombre
}
