package service

import (
	"os"

	"github.com/stripe/stripe-go/v86"
)

// ─────────────────────────────────────────────────────────────────────────────
// Stripe Tax
//
// Con `automatic_tax.enabled=true`, Stripe calcula el impuesto de cada venta y
// lo desglosa en la factura.
//
// El comportamiento fiscal se decidió como IVA INCLUIDO: un curso de $215 se
// cobra en $215, de los cuales ~$185.34 son ingreso y ~$29.66 IVA. Ese ajuste
// vive en el dashboard de Stripe («Comportamiento fiscal predeterminado» →
// Incluido), no aquí. OJO: dejarlo en «Automático» produce el mismo resultado
// para MXN por casualidad —Stripe incluye impuestos en toda divisa que no sea
// USD o CAD— pero conviene fijarlo explícitamente para que no cambie solo si
// algún día se vende en dólares.
//
// Apagado por defecto, igual que SPEI. Encenderlo sin registro fiscal en la
// región del cliente no rompe el cobro, pero devuelve impuesto CERO y da la
// falsa impresión de estar recaudando IVA cuando no es así.
// ─────────────────────────────────────────────────────────────────────────────

func impuestosHabilitados() bool {
	return os.Getenv("STRIPE_TAX_HABILITADO") == "1"
}

// impuestoAutomatico devuelve el bloque de automatic_tax para la sesión.
//
// Devuelve nil cuando está apagado: mandar `automatic_tax: {enabled: false}` y
// no mandar nada acaban igual, pero omitirlo deja las sesiones exactamente como
// estaban antes de este cambio.
func impuestoAutomatico() *stripe.CheckoutSessionAutomaticTaxParams {
	if !impuestosHabilitados() {
		return nil
	}
	return &stripe.CheckoutSessionAutomaticTaxParams{
		Enabled: stripe.Bool(true),
	}
}

// actualizacionDelCliente autoriza a Checkout a escribir en el Customer los
// datos que capture durante el pago.
//
// NO depende de Stripe Tax, y esto costó un checkout caído: la sesión ya traía
// `tax_id_collection` activado de antes, y en cuanto se le pasa un `customer`
// existente Stripe exige `customer_update[name]='auto'`. Sin él responde
//
//	400: "Tax ID collection requires updating business name on the customer"
//
// y no se crea la sesión, así que el comprador no puede pagar de ninguna forma.
//
// Por eso la regla es: SIEMPRE que se mande `customer`, se manda esto. El
// `address` va en el mismo bloque porque además es lo que permite recalcular el
// impuesto si el comprador corrige su dirección; sin Tax encendido no estorba.
func actualizacionDelCliente() *stripe.CheckoutSessionCustomerUpdateParams {
	return &stripe.CheckoutSessionCustomerUpdateParams{
		Address: stripe.String("auto"),
		Name:    stripe.String("auto"),
	}
}
