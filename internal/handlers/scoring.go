package handlers

// RespuestaInput representa una respuesta a una pregunta en el cálculo de puntaje.
type RespuestaInput struct {
	Valor      float64 // valor (peso) de la pregunta
	Tipo       string  // "multiple_choice" | "open_text"
	EsCorrecta bool    // true si la opción seleccionada es correcta
	Respondida bool    // false si el usuario no seleccionó opción
}

// ResultadoPuntaje contiene el resultado del cálculo.
type ResultadoPuntaje struct {
	Puntaje    float64
	PuntajeMax float64
	Porcentaje float64
}

// CalcularPuntaje calcula el puntaje obtenido a partir de las respuestas dadas.
// Las preguntas de tipo "open_text" suman al puntaje máximo pero no al puntaje
// automático (requieren revisión manual).
func CalcularPuntaje(respuestas []RespuestaInput) ResultadoPuntaje {
	var puntaje, puntajeMax float64
	for _, r := range respuestas {
		puntajeMax += r.Valor
		if r.Tipo != "open_text" && r.Respondida && r.EsCorrecta {
			puntaje += r.Valor
		}
	}
	porcentaje := 0.0
	if puntajeMax > 0 {
		porcentaje = puntaje / puntajeMax * 100
	}
	return ResultadoPuntaje{Puntaje: puntaje, PuntajeMax: puntajeMax, Porcentaje: porcentaje}
}
