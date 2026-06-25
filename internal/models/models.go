package models

import "time"

type User struct {
	ID           string    `json:"id"                   db:"id"`
	Name         string    `json:"name"                 db:"name"`
	Email        string    `json:"email"                db:"email"`
	PasswordHash string    `json:"-"                    db:"password_hash"`
	Role         string    `json:"role"                 db:"role"`
	TokenVersion int       `json:"-"                    db:"token_version"`
	Bio          string    `json:"bio,omitempty"        db:"bio"`
	AvatarURL    string    `json:"avatar_url,omitempty" db:"avatar_url"`
	CoverURL     string    `json:"cover_url,omitempty"  db:"cover_url"`
	Phone        string    `json:"phone,omitempty"      db:"phone"`
	Specialty    string    `json:"specialty,omitempty"  db:"specialty"`
	CreatedAt    time.Time `json:"created_at"           db:"created_at"`
}

type Capacitacion struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Type           string    `json:"type"`
	FilePath       string    `json:"file_path,omitempty"`
	Content        string    `json:"content,omitempty"`
	InstructorID   *string   `json:"instructor_id,omitempty"`
	IsPublic       bool      `json:"is_public"`
	CodigoAcceso   string    `json:"codigo_acceso,omitempty"`
	WelcomeMessage string    `json:"welcome_message,omitempty"`
	ThumbnailURL         string    `json:"thumbnail_url,omitempty"`
	Color                string    `json:"color,omitempty"`
	TotalLecciones       int       `json:"total_lecciones,omitempty"`
	LeccionesCompletadas int       `json:"lecciones_completadas,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

type Leccion struct {
	ID             string    `json:"id"`
	CapacitacionID string    `json:"capacitacion_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	Type           string    `json:"type"`
	FilePath       string    `json:"file_path,omitempty"`
	Content        string    `json:"content,omitempty"`
	Orden          int       `json:"orden"`
	DuracionMin    int       `json:"duracion_min,omitempty"`
	Completada     bool      `json:"completada,omitempty"`
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
	Tipo     string   `json:"tipo"`
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
	Tipo               string             `json:"tipo"`
	Orden              int                `json:"orden"`
	Opciones           []OpcionIntermedia `json:"opciones,omitempty"`
}

type OpcionIntermedia struct {
	ID         string `json:"id"`
	PreguntaID string `json:"pregunta_id"`
	Texto      string `json:"texto"`
	EsCorrecta bool   `json:"es_correcta,omitempty"`
}

type ReactionCount struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

type ForoPost struct {
	ID          string           `json:"id"`
	LeccionID   string           `json:"leccion_id"`
	UserID      string           `json:"user_id"`
	UserName    string           `json:"user_name,omitempty"`
	Titulo      string           `json:"titulo"`
	Contenido   string           `json:"contenido"`
	MediaURL    string           `json:"media_url,omitempty"`
	MediaType   string           `json:"media_type,omitempty"`
	LikeCount   int              `json:"like_count"`
	UserLiked   bool             `json:"user_liked"`
	Reactions   []ReactionCount  `json:"reactions,omitempty"`
	Comentarios []ForoComentario `json:"comentarios,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

type ForoComentario struct {
	ID        string          `json:"id"`
	PostID    string          `json:"post_id"`
	ParentID  *string         `json:"parent_id,omitempty"`
	UserID    string          `json:"user_id"`
	UserName  string          `json:"user_name,omitempty"`
	Contenido string          `json:"contenido"`
	MediaURL  string          `json:"media_url,omitempty"`
	MediaType string          `json:"media_type,omitempty"`
	IsPrivate bool            `json:"is_private"`
	Reactions []ReactionCount `json:"reactions,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}
