package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Bio          string    `json:"bio,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Specialty    string    `json:"specialty,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Capacitacion struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Type           string    `json:"type"` // video | document | text | course
	FilePath       string    `json:"file_path,omitempty"`
	Content        string    `json:"content,omitempty"`
	InstructorID   *string   `json:"instructor_id,omitempty"`
	IsPublic       bool      `json:"is_public"`
	CodigoAcceso   string    `json:"codigo_acceso,omitempty"`
	WelcomeMessage string    `json:"welcome_message,omitempty"`
	ThumbnailURL   string    `json:"thumbnail_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Leccion struct {
	ID             string    `json:"id"`
	CapacitacionID string    `json:"capacitacion_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	Type           string    `json:"type"` // video | document | text
	FilePath       string    `json:"file_path,omitempty"`
	Content        string    `json:"content,omitempty"`
	Orden          int       `json:"orden"`
	DuracionMin    int       `json:"duracion_min,omitempty"`
	Completada     bool      `json:"completada,omitempty"` // calculado por usuario
	CreatedAt      time.Time `json:"created_at"`
}

type Examen struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	InstructorID       *string    `json:"instructor_id,omitempty"`
	CapacitacionID     *string    `json:"capacitacion_id,omitempty"`
	CapacitacionNombre string     `json:"capacitacion_nombre,omitempty"`
	Preguntas          []Pregunta `json:"preguntas,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type Pregunta struct {
	ID       string   `json:"id"`
	ExamenID string   `json:"examen_id"`
	Texto    string   `json:"texto"`
	Tipo     string   `json:"tipo"` // multiple_choice | true_false | open_text
	Valor    float64  `json:"valor"`
	Orden    int      `json:"orden"`
	Opciones []Opcion `json:"opciones,omitempty"`
}

type Opcion struct {
	ID         string `json:"id"`
	PreguntaID string `json:"pregunta_id"`
	Texto      string `json:"texto"`
	EsCorrecta bool   `json:"es_correcta,omitempty"`
}

type Asignacion struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	CapacitacionID *string   `json:"capacitacion_id,omitempty"`
	ExamenID       *string   `json:"examen_id,omitempty"`
	AssignedAt     time.Time `json:"assigned_at"`
}

type Respuesta struct {
	PreguntaID     string `json:"pregunta_id"`
	OpcionID       string `json:"opcion_id,omitempty"`
	RespuestaTexto string `json:"respuesta_texto,omitempty"`
}

type ResultadoExamen struct {
	ExamenID   string  `json:"examen_id"`
	Titulo     string  `json:"titulo"`
	Puntaje    float64 `json:"puntaje"`
	PuntajeMax float64 `json:"puntaje_max"`
	Porcentaje float64 `json:"porcentaje"`
}

type PreguntaIntermedia struct {
	ID                 string             `json:"id"`
	CapacitacionID     string             `json:"capacitacion_id"`
	DespuesDeLeccionID *string            `json:"despues_de_leccion_id,omitempty"`
	Texto              string             `json:"texto"`
	Tipo               string             `json:"tipo"` // multiple_choice | true_false | open_text
	Orden              int                `json:"orden"`
	Opciones           []OpcionIntermedia `json:"opciones,omitempty"`
}

type OpcionIntermedia struct {
	ID         string `json:"id"`
	PreguntaID string `json:"pregunta_id"`
	Texto      string `json:"texto"`
	EsCorrecta bool   `json:"es_correcta,omitempty"`
}

type ForoPost struct {
	ID          string           `json:"id"`
	LeccionID   string           `json:"leccion_id"`
	UserID      string           `json:"user_id"`
	UserName    string           `json:"user_name,omitempty"`
	Titulo      string           `json:"titulo"`
	Contenido   string           `json:"contenido"`
	Comentarios []ForoComentario `json:"comentarios,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

type ForoComentario struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Contenido string    `json:"contenido"`
	CreatedAt time.Time `json:"created_at"`
}
