package money

import (
	"errors"
	"math"
	"testing"
)

// TestFromFloatCorrigeElBugDeTruncado documenta el defecto exacto que motivó
// este paquete: int64(precio*100) cobraba un centavo de menos.
func TestFromFloatCorrigeElBugDeTruncado(t *testing.T) {
	casos := []struct {
		precio    float64
		esperado  Cents
		truncando Cents // lo que daba el código anterior
	}{
		{8.20, 820, 819},
		{20.15, 2015, 2014},
		{1234.35, 123435, 123434},
		{2.30, 230, 229},
		{1.10, 110, 110},       // este sí salía bien
		{349.45, 34945, 34945}, // y este también
		{1899.95, 189995, 189995},
	}

	for _, c := range casos {
		got, err := FromFloat(c.precio)
		if err != nil {
			t.Fatalf("FromFloat(%v) error = %v", c.precio, err)
		}
		if got != c.esperado {
			t.Errorf("FromFloat(%v) = %d, se esperaba %d", c.precio, got, c.esperado)
		}
		// Confirma que el método viejo era realmente incorrecto en estos casos.
		if viejo := Cents(int64(c.precio * 100)); viejo != c.truncando {
			t.Errorf("el truncado de int64(%v*100) dio %d, se documentó %d", c.precio, viejo, c.truncando)
		}
	}
}

func TestFromFloatRechazaValoresNoFinitos(t *testing.T) {
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), 1e14} {
		if _, err := FromFloat(v); !errors.Is(err, ErrPrecioInvalido) {
			t.Errorf("FromFloat(%v) debería fallar con ErrPrecioInvalido, dio %v", v, err)
		}
	}
}

func TestMulNoPierdeCentavos(t *testing.T) {
	// 3 licencias de $20.15: el cálculo con floats daba 6044 o 6045 según el orden.
	precio := MXNAmount(2015)
	total, err := precio.Mul(3)
	if err != nil {
		t.Fatalf("Mul() error = %v", err)
	}
	if total.Cents != 6045 {
		t.Errorf("3 × 2015 = %d, se esperaba 6045", total.Cents)
	}

	// 50 asientos de $1,234.35
	grande, _ := MXNAmount(123435).Mul(50)
	if grande.Cents != 6171750 {
		t.Errorf("50 × 123435 = %d, se esperaba 6171750", grande.Cents)
	}
}

func TestMulDetectaDesbordamiento(t *testing.T) {
	enorme := MXNAmount(Cents(math.MaxInt64/2 + 1))
	if _, err := enorme.Mul(3); !errors.Is(err, ErrPrecioInvalido) {
		t.Error("Mul() debería detectar el desbordamiento en vez de envolver a negativo")
	}
}

func TestMulPorCeroDaCero(t *testing.T) {
	a, err := MXNAmount(2015).Mul(0)
	if err != nil || a.Cents != 0 {
		t.Errorf("Mul(0) = %v, %v; se esperaba 0 sin error", a.Cents, err)
	}
}

func TestAddRechazaDivisasDistintas(t *testing.T) {
	if _, err := MXNAmount(100).Add(New(100, USD)); !errors.Is(err, ErrPrecioInvalido) {
		t.Error("sumar MXN con USD debería fallar")
	}
}

func TestAddDetectaDesbordamiento(t *testing.T) {
	max := MXNAmount(math.MaxInt64)
	if _, err := max.Add(MXNAmount(1)); !errors.Is(err, ErrPrecioInvalido) {
		t.Error("Add() debería detectar el desbordamiento positivo")
	}
	min := MXNAmount(math.MinInt64)
	if _, err := min.Add(MXNAmount(-1)); !errors.Is(err, ErrPrecioInvalido) {
		t.Error("Add() debería detectar el desbordamiento negativo")
	}
}

func TestSumaDeCarrito(t *testing.T) {
	// Carrito real: 1 curso de $8.20 + 3 licencias de $20.15 + 1 de $1,234.35
	licencias, _ := MXNAmount(2015).Mul(3)
	total, err := Sum(MXN, MXNAmount(820), licencias, MXNAmount(123435))
	if err != nil {
		t.Fatalf("Sum() error = %v", err)
	}
	// 820 + 6045 + 123435
	if total.Cents != 130300 {
		t.Errorf("total = %d, se esperaba 130300", total.Cents)
	}
	if total.String() != "1,303.00 MXN" {
		t.Errorf("String() = %q, se esperaba \"1,303.00 MXN\"", total.String())
	}
}

func TestSumVaciaDevuelveCeroEnLaDivisaDada(t *testing.T) {
	total, err := Sum(MXN)
	if err != nil || !total.IsZero() || total.Currency != MXN {
		t.Errorf("Sum(MXN) = %v, %v; se esperaba 0 MXN", total, err)
	}
}

func TestString(t *testing.T) {
	casos := map[Cents]string{
		0:         "0.00 MXN",
		5:         "0.05 MXN",
		820:       "8.20 MXN",
		123435:    "1,234.35 MXN",
		6171750:   "61,717.50 MXN",
		-2015:     "-20.15 MXN",
		100000000: "1,000,000.00 MXN",
	}
	for c, esperado := range casos {
		if got := MXNAmount(c).String(); got != esperado {
			t.Errorf("MXNAmount(%d).String() = %q, se esperaba %q", c, got, esperado)
		}
	}
}

// TestDivisaSinDecimales cubre el caso que rompe a casi toda integración de
// pagos: en JPY, "100" son 100 yenes, no 1.00.
func TestDivisaSinDecimales(t *testing.T) {
	yen := New(100, "JPY")
	if yen.String() != "100 JPY" {
		t.Errorf("String() = %q, se esperaba \"100 JPY\"", yen.String())
	}
	if yen.Float() != 100 {
		t.Errorf("Float() = %v, se esperaba 100", yen.Float())
	}
}

func TestFormatoParaStripe(t *testing.T) {
	a := MXNAmount(123435)
	if a.StripeAmount() != 123435 {
		t.Errorf("StripeAmount() = %d", a.StripeAmount())
	}
	if a.StripeCurrency() != "mxn" {
		t.Errorf("StripeCurrency() = %q, Stripe exige minúsculas", a.StripeCurrency())
	}
}

func TestIdaYVueltaConLaBaseLegacy(t *testing.T) {
	// Todo NUMERIC(10,2) debe sobrevivir el viaje float64 → Cents → float64.
	for _, v := range []float64{0, 0.01, 8.20, 20.15, 99.99, 1234.35, 999999.99} {
		c, err := FromFloat(v)
		if err != nil {
			t.Fatalf("FromFloat(%v) error = %v", v, err)
		}
		if vuelta := MXNAmount(c).Float(); math.Abs(vuelta-v) > 1e-9 {
			t.Errorf("ida y vuelta de %v dio %v", v, vuelta)
		}
	}
}
