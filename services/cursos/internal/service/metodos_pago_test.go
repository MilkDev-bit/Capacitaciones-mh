package service

import "testing"

// contiene dice si la lista de métodos incluye uno concreto.
func contiene(metodos []*string, buscado string) bool {
	for _, m := range metodos {
		if m != nil && *m == buscado {
			return true
		}
	}
	return false
}

func TestSPEIApagadoPorDefecto(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "")

	metodos, _ := metodosDePago(50_000, true) // $500.00
	if contiene(metodos, "customer_balance") {
		t.Fatal("SPEI no debe ofrecerse sin SPEI_HABILITADO=1")
	}
	if !contiene(metodos, "card") {
		t.Fatal("la tarjeta debe ofrecerse siempre")
	}
}

func TestSPEISeOfreceAlEncenderlo(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "1")

	metodos, opciones := metodosDePago(50_000, true)
	if !contiene(metodos, "customer_balance") {
		t.Fatal("SPEI debe ofrecerse con SPEI_HABILITADO=1")
	}
	if opciones == nil || opciones.CustomerBalance == nil {
		t.Fatal("faltan las opciones de customer_balance")
	}
	// Sin estos dos valores Stripe rechaza la sesión entera.
	if got := *opciones.CustomerBalance.FundingType; got != "bank_transfer" {
		t.Fatalf("funding_type = %q, se esperaba bank_transfer", got)
	}
	if got := *opciones.CustomerBalance.BankTransfer.Type; got != "mx_bank_transfer" {
		t.Fatalf("bank_transfer.type = %q, se esperaba mx_bank_transfer", got)
	}
}

func TestSPEIRespetaElMinimo(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "1")
	t.Setenv("SPEI_MONTO_MINIMO_CENTAVOS", "20000") // $200.00

	if metodos, _ := metodosDePago(19_999, true); contiene(metodos, "customer_balance") {
		t.Fatal("por debajo del mínimo no debe ofrecerse SPEI")
	}
	if metodos, _ := metodosDePago(20_000, true); !contiene(metodos, "customer_balance") {
		t.Fatal("justo en el mínimo sí debe ofrecerse")
	}
}

// SPEI no tiene el tope de $10,000 de OXXO: una compra corporativa grande debe
// poder pagarse por transferencia aunque quede fuera del rango de OXXO.
func TestSPEIConvivaConOXXOYSuTope(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "1")
	t.Setenv("OXXO_DESHABILITADO", "")

	metodos, opciones := metodosDePago(5_000_000, true) // $50,000.00
	if contiene(metodos, "oxxo") {
		t.Fatal("OXXO no admite $50,000; mandarlo tumba la sesión entera")
	}
	if !contiene(metodos, "customer_balance") {
		t.Fatal("SPEI sí debe cubrir importes altos")
	}
	if opciones.OXXO != nil {
		t.Fatal("no deben viajar opciones de OXXO si OXXO no se ofrece")
	}
}

func TestAmbosMetodosEnRangoComun(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "1")
	t.Setenv("OXXO_DESHABILITADO", "")

	metodos, opciones := metodosDePago(50_000, true)
	for _, esperado := range []string{"card", "oxxo", "customer_balance"} {
		if !contiene(metodos, esperado) {
			t.Fatalf("falta el método %q", esperado)
		}
	}
	if opciones.OXXO == nil || opciones.CustomerBalance == nil {
		t.Fatal("deben viajar las opciones de ambos")
	}
}

// Con solo tarjeta se manda nil y no una estructura vacía: `{}` no es lo mismo
// que ausencia, y no conviene cambiar lo que ya funcionaba.
func TestSinMetodosDiferidosNoHayOpciones(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "")
	t.Setenv("OXXO_DESHABILITADO", "1")

	metodos, opciones := metodosDePago(50_000, true)
	if len(metodos) != 1 || !contiene(metodos, "card") {
		t.Fatal("solo debe quedar la tarjeta")
	}
	if opciones != nil {
		t.Fatal("sin métodos diferidos, payment_method_options debe ir ausente")
	}
}

func TestRequiereClienteSoloConSPEI(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "")
	if RequiereCliente(50_000) {
		t.Fatal("sin SPEI no hace falta crear Customer")
	}

	t.Setenv("SPEI_HABILITADO", "1")
	if !RequiereCliente(50_000) {
		t.Fatal("con SPEI hace falta Customer: la CLABE se emite a su nombre")
	}
	if RequiereCliente(1) {
		t.Fatal("por debajo del mínimo SPEI no aplica, así que tampoco el Customer")
	}
}

// El caso que más daño hace: SPEI encendido pero sin Customer resuelto (Stripe
// caído, o el alta del cliente falló). Mandar customer_balance sin cliente
// tumba la sesión ENTERA y el comprador no puede pagar ni con tarjeta.
func TestSinClienteNoSeOfreceSPEI(t *testing.T) {
	t.Setenv("SPEI_HABILITADO", "1")

	metodos, opciones := metodosDePago(50_000, false)
	if contiene(metodos, "customer_balance") {
		t.Fatal("sin Customer, SPEI no debe ofrecerse: tumbaría la sesión entera")
	}
	if !contiene(metodos, "card") {
		t.Fatal("la tarjeta debe seguir disponible pase lo que pase")
	}
	if opciones != nil && opciones.CustomerBalance != nil {
		t.Fatal("no deben viajar opciones de customer_balance si el método no se ofrece")
	}
}
