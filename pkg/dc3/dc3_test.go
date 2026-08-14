package dc3

import (
	"strings"
	"testing"
)

func datosCompletos() Datos {
	return Datos{
		NombreTrabajador:    "Torruco Bautista Javier",
		CURP:                "TOBJ940921HTSRTV02",
		OcupacionEspecifica: "04.6 Supervisores en la construcción",
		Puesto:              "Supervisor",
		RazonSocial:         "MH Soluciones Empresariales",
		RFC:                 "MHS150101AB1",
		NombrePatron:        "Juan Pérez",
		NombreRepresentante: "María López",
		NombreCurso:         "Trabajos en altura",
		DuracionHoras:       "8",
		AreaTematica:        "6000",
		NombreCapacitador:   "Instructor Certificado",
		FechaInicio:         "2026-08-10",
		FechaFin:            "2026-08-12",
	}
}

func TestFaltantes(t *testing.T) {
	if faltan := datosCompletos().Faltantes(); len(faltan) != 0 {
		t.Fatalf("datos completos no deberían reportar faltantes, se obtuvo %v", faltan)
	}

	var vacio Datos
	// Los 14 campos obligatorios. Si alguien añade uno nuevo sin actualizar el
	// formulario del alumno, este número cambia y la prueba lo delata.
	if faltan := vacio.Faltantes(); len(faltan) != 14 {
		t.Fatalf("esperaba 14 faltantes en unos datos vacíos, se obtuvo %d: %v", len(faltan), faltan)
	}

	d := datosCompletos()
	d.CURP = "   " // solo espacios: debe contar como ausente
	faltan := d.Faltantes()
	if len(faltan) != 1 || faltan[0] != "CURP" {
		t.Fatalf("un campo en blanco debe reportarse como faltante, se obtuvo %v", faltan)
	}
}

func TestGenerarRechazaDatosIncompletos(t *testing.T) {
	d := datosCompletos()
	d.Puesto = ""

	_, err := Generar(d)
	if err == nil {
		t.Fatal("generar sin puesto debería fallar")
	}
	// El mensaje tiene que nombrar el campo: la generación es automática y el
	// alumno solo ve este texto para saber qué corregir.
	if !strings.Contains(err.Error(), "puesto") {
		t.Fatalf("el error debe nombrar el campo faltante, se obtuvo: %v", err)
	}
}

func TestRFCConGuiones(t *testing.T) {
	casos := []struct {
		nombre string
		entra  string
		espera string
	}{
		// Persona moral: 12 caracteres, con espacio inicial para alinear en la
		// rejilla de 15 casillas.
		{"moral", "MHS150101AB1", " MHS-150101-AB1"},
		// Persona física: 13.
		{"fisica", "TOBJ940921HT2", "TOBJ-940921-HT2"},
		{"ya con guiones", "MHS-150101-AB1", " MHS-150101-AB1"},
		{"con espacios", "MHS 150101 AB1", " MHS-150101-AB1"},
		// Longitud inesperada: se deja tal cual en vez de recortar y producir
		// una clave fiscal inventada.
		{"longitud rara", "ABC123", "ABC123"},
		{"vacio", "", ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := rfcConGuiones(c.entra); got != c.espera {
				t.Fatalf("rfcConGuiones(%q) = %q, esperaba %q", c.entra, got, c.espera)
			}
		})
	}
}

func TestCodigoOcupacion(t *testing.T) {
	casos := map[string]string{
		"04.6 Supervisores en la construcción": "04.6",
		"4.6 Supervisores":                     "04.6", // se normaliza a dos dígitos
		"1234 Algo":                            "1234",
		"SinEspacios":                          "SinEspacios",
		"":                                     "",
	}
	for entra, espera := range casos {
		if got := codigoOcupacion(entra); got != espera {
			t.Fatalf("codigoOcupacion(%q) = %q, esperaba %q", entra, got, espera)
		}
	}
}

func TestConSufijoHoras(t *testing.T) {
	casos := map[string]string{
		"8":        "8 HRS",
		"8 HRS":    "8 HRS",
		"8 hrs":    "8 HRS",
		"20 HORAS": "20 HORAS",
		"":         "",
	}
	for entra, espera := range casos {
		if got := conSufijoHoras(entra); got != espera {
			t.Fatalf("conSufijoHoras(%q) = %q, esperaba %q", entra, got, espera)
		}
	}
}

func TestRepartirRellenaCasillasSobrantes(t *testing.T) {
	m := marcadores(datosCompletos())

	// La CURP de la prueba tiene 18 caracteres: todas las casillas ocupadas.
	if m["C1"] != "T" || m["C18"] != "2" {
		t.Fatalf("la CURP no se repartió bien: C1=%v C18=%v", m["C1"], m["C18"])
	}

	// Un CURP corto debe dejar las casillas restantes VACÍAS, no sin sustituir:
	// un marcador sin reemplazar se imprimiría literal como "{C18}".
	d := datosCompletos()
	d.CURP = "ABC"
	m = marcadores(d)
	for _, casilla := range []string{"C4", "C10", "C18"} {
		if m[casilla] != "" {
			t.Fatalf("%s debería quedar vacía, se obtuvo %q", casilla, m[casilla])
		}
	}
}

func TestRepartirFecha(t *testing.T) {
	m := marcadores(datosCompletos())

	// Inicio 2026-08-10.
	if m["IY1"] != "2" || m["IY4"] != "6" || m["IM1"] != "0" || m["IM2"] != "8" ||
		m["ID1"] != "1" || m["ID2"] != "0" {
		t.Fatalf("fecha de inicio mal repartida: %v %v %v %v %v %v",
			m["IY1"], m["IY4"], m["IM1"], m["IM2"], m["ID1"], m["ID2"])
	}
	// Fin 2026-08-12.
	if m["ED1"] != "1" || m["ED2"] != "2" {
		t.Fatalf("fecha de fin mal repartida: ED1=%v ED2=%v", m["ED1"], m["ED2"])
	}

	// Una fecha inválida no debe dejar marcadores sin sustituir.
	d := datosCompletos()
	d.FechaInicio = "2026"
	m = marcadores(d)
	for _, casilla := range []string{"IY1", "IM1", "ID2"} {
		if m[casilla] != "" {
			t.Fatalf("con fecha inválida %s debería quedar vacía, se obtuvo %q", casilla, m[casilla])
		}
	}
}

func TestMarcadoresCubrenLaPlantilla(t *testing.T) {
	// La plantilla oficial tiene 59 marcadores. Si el mapa deja de cubrirlos
	// todos, los que falten se imprimen literalmente en la constancia.
	m := marcadores(datosCompletos())
	if len(m) != 59 {
		t.Fatalf("esperaba 59 marcadores, se generaron %d", len(m))
	}
}

func TestNombreArchivo(t *testing.T) {
	n := NombreArchivo("Torruco Bautista Javier", "Trabajos en altura")
	if !strings.HasPrefix(n, "DC3-Torruco-Bautista-Javier-Trabajos-en-altura-") {
		t.Fatalf("nombre inesperado: %s", n)
	}
	if !strings.HasSuffix(n, ".docx") {
		t.Fatalf("debe terminar en .docx: %s", n)
	}

	// Acentos y signos se descartan: el nombre viaja en una URL de R2.
	if got := NombreArchivo("Peña Ñandú", "Curso #1 (básico)"); strings.ContainsAny(got, "ñÑ#()áé") {
		t.Fatalf("el nombre no debería llevar caracteres especiales: %s", got)
	}
}

func TestAreasTematicas(t *testing.T) {
	// Las nueve del reverso del formato oficial. Si alguien añade o quita una
	// sin que la STPS publique un formato nuevo, la constancia sale con una
	// clave que no existe en el catálogo impreso al dorso.
	if len(AreasTematicas) != 9 {
		t.Fatalf("el catálogo debe tener 9 áreas, tiene %d", len(AreasTematicas))
	}
	for _, clave := range []string{"1000", "2000", "3000", "4000", "5000", "6000", "7000", "8000", "9000"} {
		if !AreaValida(clave) {
			t.Fatalf("la clave %s debería ser válida", clave)
		}
	}
	if AreasTematicas["6000"] != "Seguridad" {
		t.Fatalf("6000 debe ser Seguridad, es %q", AreasTematicas["6000"])
	}
}

func TestAreaValidaRechazaBasura(t *testing.T) {
	for _, clave := range []string{"", "600", "0", "10000", "Seguridad", "6000 Seguridad", "abc"} {
		if AreaValida(clave) {
			t.Fatalf("%q no debería aceptarse como área temática", clave)
		}
	}
	// Los espacios sobrantes de un copiar y pegar no invalidan una clave buena.
	if !AreaValida("  6000 ") {
		t.Fatal("la clave debe aceptarse con espacios alrededor")
	}
}

func TestDecodificarLogo(t *testing.T) {
	// Con prefijo data: como lo manda un navegador.
	if _, ok := decodificarLogo("data:image/png;base64,aGVsbG8="); !ok {
		t.Fatal("debería aceptar el prefijo data:")
	}
	if _, ok := decodificarLogo("aGVsbG8="); !ok {
		t.Fatal("debería aceptar base64 pelado")
	}
	if _, ok := decodificarLogo(""); ok {
		t.Fatal("vacío no es logo")
	}
	if _, ok := decodificarLogo("no-es-base64-válido!!"); ok {
		t.Fatal("base64 inválido debe descartarse en silencio, no romper la constancia")
	}
}
