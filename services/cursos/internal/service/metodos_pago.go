package service

// Métodos de pago admitidos en las Checkout Sessions de pago único.
//
// Vive aparte porque la decisión "¿este importe admite OXXO?" se repite en los
// tres constructores de sesión (curso individual, licencia B2B y carrito) y
// equivocarse en uno solo produce un checkout que Stripe rechaza al crearse.

import (
	"os"
	"strconv"

	"github.com/stripe/stripe-go/v86"
)

// Límites de importe de OXXO, según la documentación de Stripe:
// mínimo 10.00 MXN, máximo 10,000.00 MXN.
//
// Importa acertar en los DOS extremos: si se manda "oxxo" en una sesión fuera
// de rango, Stripe NO ignora el método — falla la creación de la sesión entera
// y el comprador se queda sin poder pagar ni siquiera con tarjeta. Por eso se
// filtra por importe en lugar de ofrecerlo siempre.
//
// Son configurables porque los fija Stripe y pueden cambiar sin que nos
// enteremos; ajustar una variable de entorno es más rápido que desplegar.
const (
	oxxoMinimoCentavosPorDefecto int64 = 1_000     // $10.00 MXN
	oxxoTopeCentavosPorDefecto   int64 = 1_000_000 // $10,000.00 MXN
)

// Vigencia de la ficha. Stripe admite de 1 a 7 días; su valor por defecto es 5.
const (
	oxxoDiasVigenciaPorDefecto int64 = 5
	oxxoDiasMinimo             int64 = 1
	oxxoDiasMaximo             int64 = 7
)

func oxxoHabilitado() bool {
	// Apagado explícito para poder desactivar OXXO en caliente si la cuenta de
	// Stripe pierde el método o hay una incidencia con el proveedor.
	return os.Getenv("OXXO_DESHABILITADO") != "1"
}

func enteroDeEnv(clave string, porDefecto int64) int64 {
	v, err := strconv.ParseInt(os.Getenv(clave), 10, 64)
	if err != nil || v <= 0 {
		return porDefecto
	}
	return v
}

// oxxoAplica dice si un importe cae dentro del rango que OXXO admite.
func oxxoAplica(importeCentavos int64) bool {
	if !oxxoHabilitado() {
		return false
	}
	minimo := enteroDeEnv("OXXO_MONTO_MINIMO_CENTAVOS", oxxoMinimoCentavosPorDefecto)
	tope := enteroDeEnv("OXXO_MONTO_MAXIMO_CENTAVOS", oxxoTopeCentavosPorDefecto)
	return importeCentavos >= minimo && importeCentavos <= tope
}

// oxxoDiasVigencia acota el plazo al rango que Stripe acepta.
//
// El recorte no es paranoia: un valor fuera de 1-7 no se ignora, hace fallar la
// creación de la sesión. Es un error de configuración que dejaría la tienda sin
// checkout, así que se prefiere un plazo distinto al pedido antes que caerse.
func oxxoDiasVigencia() int64 {
	dias := enteroDeEnv("OXXO_DIAS_VIGENCIA", oxxoDiasVigenciaPorDefecto)
	if dias < oxxoDiasMinimo {
		return oxxoDiasMinimo
	}
	if dias > oxxoDiasMaximo {
		return oxxoDiasMaximo
	}
	return dias
}

// ─────────────────────────────────────────────────────────────────────────────
// Transferencia bancaria (SPEI)
//
// En Stripe el método NO se llama "spei": es `customer_balance` con
// `bank_transfer.type = mx_bank_transfer`. Stripe le asigna al comprador una
// CLABE de referencia y acredita el pago cuando el dinero llega.
//
// Es de notificación diferida, igual que OXXO: la sesión se completa cuando el
// comprador recibe los datos bancarios, no cuando paga. El acceso al curso lo
// otorga `checkout.session.async_payment_succeeded`, que el webhook ya maneja.
//
// Apagado por defecto: exige que el método esté activado en la cuenta de Stripe
// y que las sesiones creen un Customer. Encenderlo sin lo primero no degrada el
// checkout, lo ROMPE —Stripe rechaza la creación de la sesión entera y el
// comprador se queda sin poder pagar ni con tarjeta—, así que la activación es
// un acto deliberado y no el estado por defecto.
// ─────────────────────────────────────────────────────────────────────────────

func speiHabilitado() bool {
	return os.Getenv("SPEI_HABILITADO") == "1"
}

// speiAplica dice si un importe puede pagarse por transferencia.
//
// No tiene el tope de $10,000 de OXXO, pero sí un mínimo: una transferencia por
// unos pocos pesos no tiene sentido y la comisión se come el importe.
const speiMinimoCentavosPorDefecto int64 = 1_000 // $10.00 MXN

func speiAplica(importeCentavos int64) bool {
	if !speiHabilitado() {
		return false
	}
	return importeCentavos >= enteroDeEnv("SPEI_MONTO_MINIMO_CENTAVOS", speiMinimoCentavosPorDefecto)
}

// RequiereCliente dice si conviene resolver el Customer antes de la sesión.
//
// La transferencia no funciona sin uno: la documentación de Stripe exige
// indicar el `customer` en la sesión cuando se ofrece `customer_balance`,
// porque la CLABE de referencia se emite a nombre de un cliente concreto.
//
// Se consulta ANTES de armar los métodos para no crear un Customer en cada
// compra con tarjeta, que no lo necesita.
func RequiereCliente(importeCentavos int64) bool {
	return speiAplica(importeCentavos)
}

// metodosDePago devuelve los métodos aplicables a un importe y sus opciones.
//
// `tieneCliente` es determinante, no informativo: sin Customer la transferencia
// NO se ofrece. Mandar `customer_balance` sin cliente hace fallar la creación de
// la sesión entera, y entonces el comprador se queda sin poder pagar ni con
// tarjeta. Ante la duda se prefiere perder el método de pago antes que la venta.
//
// OXXO y SPEI solo entran en pagos únicos en MXN. Las suscripciones NO deben
// usar esta función: ninguno de los dos admite cobros recurrentes, y Stripe
// tampoco admite transferencia en Checkout en modo suscripción.
func metodosDePago(importeCentavos int64, tieneCliente bool) ([]*string, *stripe.CheckoutSessionPaymentMethodOptionsParams) {
	metodos := []string{"card"}
	opciones := &stripe.CheckoutSessionPaymentMethodOptionsParams{}
	hayOpciones := false

	if oxxoAplica(importeCentavos) {
		metodos = append(metodos, "oxxo")
		opciones.OXXO = &stripe.CheckoutSessionPaymentMethodOptionsOXXOParams{
			ExpiresAfterDays: stripe.Int64(oxxoDiasVigencia()),
		}
		hayOpciones = true
	}

	if speiAplica(importeCentavos) && tieneCliente {
		metodos = append(metodos, "customer_balance")
		opciones.CustomerBalance = &stripe.CheckoutSessionPaymentMethodOptionsCustomerBalanceParams{
			FundingType: stripe.String("bank_transfer"),
			BankTransfer: &stripe.CheckoutSessionPaymentMethodOptionsCustomerBalanceBankTransferParams{
				Type: stripe.String("mx_bank_transfer"),
			},
		}
		hayOpciones = true
	}

	// nil y no una estructura vacía: mandar `payment_method_options: {}` no es
	// lo mismo que no mandarlo, y conviene no cambiar lo que ya funcionaba en
	// las sesiones que solo admiten tarjeta.
	if !hayOpciones {
		return stripe.StringSlice(metodos), nil
	}
	return stripe.StringSlice(metodos), opciones
}
