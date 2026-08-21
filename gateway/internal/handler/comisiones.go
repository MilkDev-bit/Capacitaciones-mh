package handler

import (
	"log/slog"
	"net/http"

	cursospb "Prueba-Go/gen/cursos"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/paymentintent"
)

// ─────────────────────────────────────────────────────────────────────────────
// Comisiones de Stripe
//
// Lo que Stripe se queda por cada cobro NO se calcula aquí. Se le pregunta a
// Stripe, porque la tarifa real depende del método de pago —tarjeta nacional,
// tarjeta internacional, OXXO, meses sin intereses—, lleva IVA encima y cambia
// con el contrato de cada cuenta. Cualquier fórmula que escribiéramos nosotros
// sería una estimación presentada como si fuera un hecho.
//
// La cifra buena vive en el BalanceTransaction del cobro: trae `fee` y `net` ya
// calculados, en centavos y en la moneda del saldo.
// ─────────────────────────────────────────────────────────────────────────────

// comisionStripe es lo que Stripe reporta de un cobro concreto.
type comisionStripe struct {
	ComisionCentavos int64
	NetoCentavos     int64
	BalanceTxID      string
}

// comisionDeIntent consulta a Stripe la comisión de un PaymentIntent.
//
// Devuelve ok=false cuando todavía no se puede saber, que no es un error: el
// BalanceTransaction aparece cuando el cobro se liquida, y con métodos
// asíncronos eso puede tardar. Quien llama debe dejar la comisión sin registrar
// —NULL en base— para que el relleno del histórico la recoja más tarde. Guardar
// un cero aquí sería peor que no guardar nada: se leería como "Stripe no cobró
// comisión" e inflaría la ganancia neta.
func comisionDeIntent(intentID string) (comisionStripe, bool) {
	if intentID == "" {
		return comisionStripe{}, false
	}

	params := &stripe.PaymentIntentParams{}
	params.AddExpand("latest_charge.balance_transaction")

	pi, err := paymentintent.Get(intentID, params)
	if err != nil {
		slog.Warn("no se pudo consultar la comisión de Stripe",
			"payment_intent", intentID, "error", err)
		return comisionStripe{}, false
	}

	if pi.LatestCharge == nil || pi.LatestCharge.BalanceTransaction == nil {
		slog.Info("cobro sin BalanceTransaction todavía; la comisión se rellenará después",
			"payment_intent", intentID)
		return comisionStripe{}, false
	}

	bt := pi.LatestCharge.BalanceTransaction
	return comisionStripe{
		ComisionCentavos: bt.Fee,
		NetoCentavos:     bt.Net,
		BalanceTxID:      bt.ID,
	}, true
}

// loteRelleno es cuántas órdenes se rellenan por llamada.
//
// Cada una es una petición a Stripe, así que un lote grande se come el límite
// de peticiones y alarga el tiempo de respuesta. Se llama varias veces hasta
// que `restantes` llega a cero.
const loteRelleno = 50

// RellenarComisiones trae de Stripe la comisión de los cobros que no la tienen.
//
// POST /api/admin/comisiones/rellenar
//
// Existe porque las comisiones empezaron a guardarse hoy: todo lo cobrado antes
// tiene la columna en NULL. Es idempotente y se puede llamar tantas veces como
// haga falta; solo mira las órdenes a las que aún les falta.
func (h *CursosHandler) RellenarComisiones(ctx *gin.Context) {
	pendientes, err := h.c.Cursos.ListOrdenesSinComision(ctx.Request.Context(),
		&cursospb.ListOrdenesSinComisionRequest{Limite: loteRelleno})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	var rellenadas, sinDatoAun int
	for _, o := range pendientes.Ordenes {
		com, ok := comisionDeIntent(o.StripePaymentIntent)
		if !ok {
			// Aún no liquidado. Se deja en NULL para el siguiente intento.
			sinDatoAun++
			continue
		}
		if _, err := h.c.Cursos.RegistrarComisionOrden(ctx.Request.Context(),
			&cursospb.RegistrarComisionOrdenRequest{
				OrdenId:              o.Id,
				ComisionCentavos:     com.ComisionCentavos,
				NetoCentavos:         com.NetoCentavos,
				BalanceTransactionId: com.BalanceTxID,
			}); err != nil {
			slog.Error("no se pudo guardar la comisión",
				"orden", o.Id, "error", err)
			continue
		}
		rellenadas++
	}

	ctx.JSON(http.StatusOK, gin.H{
		"rellenadas":     rellenadas,
		"sin_dato_aun":   sinDatoAun,
		"restantes":      pendientes.Restantes - int32(rellenadas),
		"lote_procesado": len(pendientes.Ordenes),
	})
}

// conComision añade a la transición de estado lo que Stripe cobró, si ya se sabe.
//
// Se llama solo al pasar la orden a 'pagada': es el único momento en que existe
// un cobro liquidado del que preguntar. En 'fallida' o 'cumplida' no hay nada
// que consultar.
func conComision(req *cursospb.ActualizarEstadoOrdenRequest, intentID string) {
	com, ok := comisionDeIntent(intentID)
	if !ok {
		return
	}
	req.ComisionConocida = true
	req.ComisionCentavos = com.ComisionCentavos
	req.NetoCentavos = com.NetoCentavos
	req.BalanceTransactionId = com.BalanceTxID
}
