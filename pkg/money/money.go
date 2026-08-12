// Package money representa importes como enteros en la unidad menor de la
// divisa (centavos para MXN), nunca como float64.
//
// Motivo: los precios vivían como float64 y el monto para Stripe se calculaba
// con int64(precio * 100). Eso trunca hacia abajo y cobra de menos, porque
// muchos decimales de dos cifras no son representables en binario:
//
//	int64(8.20 * 100)    == 819   (debería ser 820)
//	int64(20.15 * 100)   == 2014  (debería ser 2015)
//	int64(1234.35 * 100) == 123434 (debería ser 123435)
//
// Cuatro de cada siete precios realistas se cobraban con un centavo de menos.
// Con Cents como tipo propio, la conversión ocurre una sola vez —en el borde
// que lee la base de datos— y el resto del sistema trabaja con enteros.
package money

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Cents es un importe en centavos. Puede ser negativo (reembolsos, notas de
// crédito), pero nunca debe usarse para porcentajes ni cantidades.
type Cents int64

// Currency es un código ISO 4217. Se mantiene junto al importe porque un número
// sin divisa no significa nada.
type Currency string

const (
	MXN Currency = "MXN"
	USD Currency = "USD"
)

// exponentes por divisa: cuántos decimales tiene su unidad menor.
// JPY y CLP no tienen decimales; un importe "100" son 100 yenes, no 1.
var exponentes = map[Currency]int{
	MXN:   2,
	USD:   2,
	"JPY": 0,
	"CLP": 0,
}

// Exponent devuelve los decimales de la divisa (2 si no se conoce).
func (c Currency) Exponent() int {
	if e, ok := exponentes[c]; ok {
		return e
	}
	return 2
}

// Amount empareja importe y divisa. Es el tipo que debe cruzar fronteras
// (proto, JSON, funciones de negocio).
type Amount struct {
	Cents    Cents
	Currency Currency
}

func New(c Cents, cur Currency) Amount { return Amount{Cents: c, Currency: cur} }

// MXNAmount es el atajo para el caso normal del producto.
func MXNAmount(c Cents) Amount { return Amount{Cents: c, Currency: MXN} }

// ── Conversión desde la base de datos ────────────────────────────────────────

// ErrPrecioInvalido indica un valor que no puede representarse en centavos.
var ErrPrecioInvalido = errors.New("precio inválido")

// FromFloat convierte un NUMERIC(10,2) leído como float64 a centavos.
//
// Usa math.Round, NO una conversión directa: int64(8.20*100) da 819 porque
// 8.20 se almacena como 8.199999999999999289. Redondear al entero más cercano
// recupera el valor que el usuario escribió.
//
// Solo debe llamarse en el borde que lee la BD legacy. El código nuevo trabaja
// directamente con Cents.
func FromFloat(v float64) (Cents, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%w: %v no es un número finito", ErrPrecioInvalido, v)
	}
	// 2^53 centavos ya es un importe absurdo; más allá float64 pierde enteros.
	if math.Abs(v) > 9e13 {
		return 0, fmt.Errorf("%w: %v excede el rango representable", ErrPrecioInvalido, v)
	}
	return Cents(math.Round(v * 100)), nil
}

// MustFromFloat es FromFloat devolviendo 0 en caso de error. Para rutas donde
// un precio corrupto no debe tumbar la petición, pero sí bloquear la compra
// (un importe de 0 hace que el checkout rechace el curso por "sin precio").
func MustFromFloat(v float64) Cents {
	c, err := FromFloat(v)
	if err != nil {
		return 0
	}
	return c
}

// Float devuelve el importe como float64. Úsalo SOLO para escribir en columnas
// NUMERIC legacy o para respuestas JSON de compatibilidad — nunca para calcular.
func (a Amount) Float() float64 {
	return float64(a.Cents) / math.Pow10(a.Currency.Exponent())
}

// ── Aritmética ───────────────────────────────────────────────────────────────

// Mul multiplica por una cantidad entera (número de asientos, lugares).
// Devuelve error en desbordamiento en vez de envolver silenciosamente: un
// importe negativo por overflow sería un abono al cliente.
func (a Amount) Mul(n int64) (Amount, error) {
	if n == 0 {
		return Amount{Cents: 0, Currency: a.Currency}, nil
	}
	prod := a.Cents * Cents(n)
	if prod/Cents(n) != a.Cents {
		return Amount{}, fmt.Errorf("%w: %d × %d desborda int64", ErrPrecioInvalido, a.Cents, n)
	}
	return Amount{Cents: prod, Currency: a.Currency}, nil
}

// Add suma dos importes de la misma divisa.
func (a Amount) Add(b Amount) (Amount, error) {
	if a.Currency != b.Currency {
		return Amount{}, fmt.Errorf("%w: no se puede sumar %s con %s", ErrPrecioInvalido, a.Currency, b.Currency)
	}
	suma := a.Cents + b.Cents
	// Desbordamiento: si ambos sumandos tienen el mismo signo y el resultado cambia de signo.
	if (a.Cents > 0 && b.Cents > 0 && suma < 0) || (a.Cents < 0 && b.Cents < 0 && suma > 0) {
		return Amount{}, fmt.Errorf("%w: %d + %d desborda int64", ErrPrecioInvalido, a.Cents, b.Cents)
	}
	return Amount{Cents: suma, Currency: a.Currency}, nil
}

// Sum suma una lista de importes. Devuelve un importe en `cur` si la lista
// viene vacía, para no obligar al llamador a inventarse una divisa.
func Sum(cur Currency, importes ...Amount) (Amount, error) {
	total := Amount{Cents: 0, Currency: cur}
	for _, im := range importes {
		var err error
		if total, err = total.Add(im); err != nil {
			return Amount{}, err
		}
	}
	return total, nil
}

// IsZero indica si el importe es exactamente cero (curso gratuito).
func (a Amount) IsZero() bool { return a.Cents == 0 }

// IsPositive indica si hay algo que cobrar.
func (a Amount) IsPositive() bool { return a.Cents > 0 }

// ── Formato ──────────────────────────────────────────────────────────────────

// String da "1,234.35 MXN". Para el usuario final formatea el frontend con
// Intl.NumberFormat; esto es para logs, correos y depuración.
func (a Amount) String() string {
	exp := a.Currency.Exponent()
	neg := a.Cents < 0
	n := a.Cents
	if neg {
		n = -n
	}

	div := Cents(math.Pow10(exp))
	entero := int64(n / div)
	frac := int64(n % div)

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteString(agruparMiles(entero))
	if exp > 0 {
		b.WriteByte('.')
		b.WriteString(fmt.Sprintf("%0*d", exp, frac))
	}
	b.WriteByte(' ')
	b.WriteString(string(a.Currency))
	return b.String()
}

func agruparMiles(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, d := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(d)
	}
	return b.String()
}

// StripeAmount devuelve el importe tal como lo espera la API de Stripe:
// entero en la unidad menor de la divisa.
func (a Amount) StripeAmount() int64 { return int64(a.Cents) }

// StripeCurrency devuelve el código en minúsculas que exige Stripe.
func (a Amount) StripeCurrency() string { return strings.ToLower(string(a.Currency)) }
