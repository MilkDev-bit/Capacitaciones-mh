package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Capacitacion struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // video | document | text
	FilePath    string    `json:"file_path,omitempty"`
	Content     string    `json:"content,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Examen struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Preguntas   []Pregunta `json:"preguntas,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Pregunta struct {
	ID       string   `json:"id"`
	ExamenID string   `json:"examen_id"`
	Texto    string   `json:"texto"`
	Valor    float64  `json:"valor"`
	Orden    int      `json:"orden"`
	Opciones []Opcion `json:"opciones,omitempty"`
}

type Opcion struct {
	ID         string `json:"id"`
	PreguntaID string `json:"pregunta_id"`
	Texto      string `json:"texto"`
	EsCorrecta bool   `json:"es_correcta,omitempty"` // omitido en respuestas al usuario
}

type Asignacion struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	CapacitacionID *string   `json:"capacitacion_id,omitempty"`
	ExamenID       *string   `json:"examen_id,omitempty"`
	AssignedAt     time.Time `json:"assigned_at"`
}

type Respuesta struct {
	PreguntaID string `json:"pregunta_id"`
	OpcionID   string `json:"opcion_id"`
}

type ResultadoExamen struct {
	ExamenID   string  `json:"examen_id"`
	Titulo     string  `json:"titulo"`
	Puntaje    float64 `json:"puntaje"`
	PuntajeMax float64 `json:"puntaje_max"`
	Porcentaje float64 `json:"porcentaje"`
}
