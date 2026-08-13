package service

// Métodos de pago admitidos en las Checkout Sessions de pago único.
//
// Vive aparte porque la decisión "¿este importe admite OXXO?" se repite en los
// tres constructores de sesión (curso individual, licencia B2B y carrito) y
// equivocarse en uno solo produce un checkout que Stripe rechaza al crearse.

import (
	"os"
	"strconv"

	"github.com/stripe/stripe-go/v78"
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

// metodosDePago devuelve los métodos aplicables a un importe y sus opciones.
//
// OXXO solo entra en pagos únicos en MXN y dentro de su rango de importe. Las
// suscripciones NO deben usar esta función: OXXO no admite cobros recurrentes
// ni modo setup, así que ahí la lista se queda en tarjeta.
func metodosDePago(importeCentavos int64) ([]*string, *stripe.CheckoutSessionPaymentMethodOptionsParams) {
	metodos := []string{"card"}

	if !oxxoAplica(importeCentavos) {
		return stripe.StringSlice(metodos), nil
	}

	metodos = append(metodos, "oxxo")
	opciones := &stripe.CheckoutSessionPaymentMethodOptionsParams{
		OXXO: &stripe.CheckoutSessionPaymentMethodOptionsOXXOParams{
			ExpiresAfterDays: stripe.Int64(oxxoDiasVigencia()),
		},
	}
	return stripe.StringSlice(metodos), opciones
}
