package handlers

import (
	"testing"
)

func TestCalcularPuntaje_TodoCorrecto(t *testing.T) {
	respuestas := []RespuestaInput{
		{Valor: 10, Tipo: "multiple_choice", EsCorrecta: true, Respondida: true},
		{Valor: 5, Tipo: "multiple_choice", EsCorrecta: true, Respondida: true},
	}
	res := CalcularPuntaje(respuestas)
	if res.Puntaje != 15 {
		t.Errorf("esperado puntaje=15, got %v", res.Puntaje)
	}
	if res.PuntajeMax != 15 {
		t.Errorf("esperado puntajeMax=15, got %v", res.PuntajeMax)
	}
	if res.Porcentaje != 100 {
		t.Errorf("esperado porcentaje=100, got %v", res.Porcentaje)
	}
}

func TestCalcularPuntaje_ParcialmenteCorrecto(t *testing.T) {
	respuestas := []RespuestaInput{
		{Valor: 10, Tipo: "multiple_choice", EsCorrecta: true, Respondida: true},
		{Valor: 10, Tipo: "multiple_choice", EsCorrecta: false, Respondida: true},
		{Valor: 10, Tipo: "multiple_choice", EsCorrecta: false, Respondida: false},
	}
	res := CalcularPuntaje(respuestas)
	if res.Puntaje != 10 {
		t.Errorf("esperado puntaje=10, got %v", res.Puntaje)
	}
	if res.PuntajeMax != 30 {
		t.Errorf("esperado puntajeMax=30, got %v", res.PuntajeMax)
	}
	// Comparar con tolerancia por aritmética de punto flotante
	expected := 10.0 / 30.0 * 100
	diff := res.Porcentaje - expected
	if diff < -0.001 || diff > 0.001 {
		t.Errorf("esperado porcentaje≈%v, got %v", expected, res.Porcentaje)
	}
}

func TestCalcularPuntaje_SinRespuestas(t *testing.T) {
	res := CalcularPuntaje(nil)
	if res.Puntaje != 0 || res.PuntajeMax != 0 || res.Porcentaje != 0 {
		t.Errorf("esperado todo en 0, got %+v", res)
	}
}

func TestCalcularPuntaje_OpenTextNoSuma(t *testing.T) {
	respuestas := []RespuestaInput{
		{Valor: 10, Tipo: "open_text", EsCorrecta: false, Respondida: true},
		{Valor: 10, Tipo: "multiple_choice", EsCorrecta: true, Respondida: true},
	}
	res := CalcularPuntaje(respuestas)
	// open_text suma al max pero no al puntaje automático
	if res.Puntaje != 10 {
		t.Errorf("esperado puntaje=10, got %v", res.Puntaje)
	}
	if res.PuntajeMax != 20 {
		t.Errorf("esperado puntajeMax=20, got %v", res.PuntajeMax)
	}
	if res.Porcentaje != 50 {
		t.Errorf("esperado porcentaje=50, got %v", res.Porcentaje)
	}
}

func TestCalcularPuntaje_TodosIncorrectos(t *testing.T) {
	respuestas := []RespuestaInput{
		{Valor: 5, Tipo: "multiple_choice", EsCorrecta: false, Respondida: true},
		{Valor: 5, Tipo: "multiple_choice", EsCorrecta: false, Respondida: true},
	}
	res := CalcularPuntaje(respuestas)
	if res.Puntaje != 0 {
		t.Errorf("esperado puntaje=0, got %v", res.Puntaje)
	}
	if res.Porcentaje != 0 {
		t.Errorf("esperado porcentaje=0, got %v", res.Porcentaje)
	}
}
