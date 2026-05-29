package repository

import (
	"context"
	"time"

	leccionespb "Prueba-Go/gen/lecciones"

	"github.com/jmoiron/sqlx"
)

// Leccion es el modelo interno.
type Leccion struct {
	ID             string    `db:"id"`
	CapacitacionID string    `db:"capacitacion_id"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	Type           string    `db:"type"`
	FilePath       string    `db:"file_path"`
	Content        string    `db:"content"`
	Orden          int32     `db:"orden"`
	DuracionMin    int32     `db:"duracion_min"`
	Completada     bool      `db:"completada"`
	CreatedAt      time.Time `db:"created_at"`
}

func (l *Leccion) ToProto() *leccionespb.LeccionResponse {
	return &leccionespb.LeccionResponse{
		Id: l.ID, CursoId: l.CapacitacionID, Title: l.Title,
		Description: l.Description, Type: l.Type, FilePath: l.FilePath,
		Content: l.Content, Orden: l.Orden, DuracionMin: l.DuracionMin,
		Completada: l.Completada,
		CreatedAt:  l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// PreguntaIntermedia modelo interno.
type PreguntaIntermedia struct {
	ID                 string    `db:"id"`
	CapacitacionID     string    `db:"capacitacion_id"`
	DespuesDeLeccionID *string   `db:"despues_de_leccion_id"`
	Texto              string    `db:"texto"`
	Tipo               string    `db:"tipo"`
	Orden              int32     `db:"orden"`
	CreatedAt          time.Time `db:"created_at"`
}

type OpcionIntermedia struct {
	ID         string `db:"id"`
	PreguntaID string `db:"pregunta_id"`
	Texto      string `db:"texto"`
	EsCorrecta bool   `db:"es_correcta"`
}

// LeccionesRepository contrato.
type LeccionesRepository interface {
	ListByCurso(ctx context.Context, cursoID string) ([]*Leccion, error)
	ListByCursoConProgreso(ctx context.Context, cursoID, userID string) ([]*Leccion, error)
	FindByID(ctx context.Context, leccionID string) (*Leccion, error)
	Create(ctx context.Context, req *leccionespb.CreateLeccionRequest) (*Leccion, error)
	Update(ctx context.Context, req *leccionespb.UpdateLeccionRequest) (*Leccion, error)
	Delete(ctx context.Context, leccionID string) error
	Reorder(ctx context.Context, cursoID string, ids []string) error
	MarcarCompleta(ctx context.Context, leccionID, userID string) error

	ListPreguntas(ctx context.Context, cursoID string) ([]*PreguntaIntermedia, error)
	GetOpciones(ctx context.Context, preguntaID string) ([]*OpcionIntermedia, error)
	CreatePregunta(ctx context.Context, req *leccionespb.CreateIntermediaRequest) (*PreguntaIntermedia, error)
	DeletePregunta(ctx context.Context, preguntaID string) error
	SubmitRespuestas(ctx context.Context, cursoID, userID string, respuestas []*leccionespb.Respuesta) (correctas, total int32, err error)
}

type postgresLeccionesRepository struct{ db *sqlx.DB }

func NewLeccionesRepository(db *sqlx.DB) LeccionesRepository {
	return &postgresLeccionesRepository{db: db}
}

const selectLeccion = `SELECT id, capacitacion_id, title,
	COALESCE(description,'') description, type,
	COALESCE(file_path,'') file_path, COALESCE(content,'') content,
	orden, COALESCE(duracion_min,0) duracion_min, false AS completada, created_at
  FROM lecciones`

func (r *postgresLeccionesRepository) ListByCurso(ctx context.Context, cursoID string) ([]*Leccion, error) {
	var lecs []*Leccion
	return lecs, r.db.SelectContext(ctx, &lecs,
		selectLeccion+` WHERE capacitacion_id=$1 AND deleted_at IS NULL ORDER BY orden`, cursoID)
}

func (r *postgresLeccionesRepository) ListByCursoConProgreso(ctx context.Context, cursoID, userID string) ([]*Leccion, error) {
	query := `SELECT l.id, l.capacitacion_id, l.title,
		COALESCE(l.description,'') description, l.type,
		COALESCE(l.file_path,'') file_path, COALESCE(l.content,'') content,
		l.orden, COALESCE(l.duracion_min,0) duracion_min,
		CASE WHEN p.id IS NOT NULL THEN true ELSE false END AS completada,
		l.created_at
	  FROM lecciones l
	  LEFT JOIN progreso_lecciones p ON p.leccion_id = l.id AND p.user_id = $2
	  WHERE l.capacitacion_id = $1 AND l.deleted_at IS NULL
	  ORDER BY l.orden`
	var lecs []*Leccion
	return lecs, r.db.SelectContext(ctx, &lecs, query, cursoID, userID)
}

func (r *postgresLeccionesRepository) FindByID(ctx context.Context, leccionID string) (*Leccion, error) {
	l := &Leccion{}
	return l, r.db.GetContext(ctx, l,
		selectLeccion+` WHERE id=$1 AND deleted_at IS NULL`, leccionID)
}

func (r *postgresLeccionesRepository) Create(ctx context.Context, req *leccionespb.CreateLeccionRequest) (*Leccion, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO lecciones(capacitacion_id,title,description,type,file_path,content,orden,duracion_min)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		req.CursoId, req.Title, req.Description, req.Type,
		req.FilePath, req.Content, req.Orden, req.DuracionMin,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *postgresLeccionesRepository) Update(ctx context.Context, req *leccionespb.UpdateLeccionRequest) (*Leccion, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE lecciones SET title=$1,description=$2,type=$3,file_path=$4,
		 content=$5,orden=$6,duracion_min=$7 WHERE id=$8`,
		req.Title, req.Description, req.Type, req.FilePath,
		req.Content, req.Orden, req.DuracionMin, req.LeccionId)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, req.LeccionId)
}

func (r *postgresLeccionesRepository) Delete(ctx context.Context, leccionID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE lecciones SET deleted_at=NOW() WHERE id=$1`, leccionID)
	return err
}

func (r *postgresLeccionesRepository) Reorder(ctx context.Context, cursoID string, ids []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE lecciones SET orden=$1 WHERE id=$2 AND capacitacion_id=$3`,
			i+1, id, cursoID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *postgresLeccionesRepository) MarcarCompleta(ctx context.Context, leccionID, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO progreso_lecciones(leccion_id, user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		leccionID, userID)
	return err
}

func (r *postgresLeccionesRepository) ListPreguntas(ctx context.Context, cursoID string) ([]*PreguntaIntermedia, error) {
	var pqs []*PreguntaIntermedia
	return pqs, r.db.SelectContext(ctx, &pqs,
		`SELECT id, capacitacion_id, despues_de_leccion_id, texto, tipo, orden, created_at
		   FROM preguntas_intermedias WHERE capacitacion_id=$1 ORDER BY orden`, cursoID)
}

func (r *postgresLeccionesRepository) GetOpciones(ctx context.Context, preguntaID string) ([]*OpcionIntermedia, error) {
	var opts []*OpcionIntermedia
	return opts, r.db.SelectContext(ctx, &opts,
		`SELECT id, pregunta_id, texto, es_correcta FROM opciones_intermedias WHERE pregunta_id=$1`, preguntaID)
}

func (r *postgresLeccionesRepository) CreatePregunta(ctx context.Context, req *leccionespb.CreateIntermediaRequest) (*PreguntaIntermedia, error) {
	var id string
	var dlid *string
	if req.DespuesDeLeccionId != "" {
		dlid = &req.DespuesDeLeccionId
	}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO preguntas_intermedias(capacitacion_id,despues_de_leccion_id,texto,tipo,orden)
		 VALUES($1,$2,$3,$4,$5) RETURNING id`,
		req.CursoId, dlid, req.Texto, req.Tipo, req.Orden,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	// Insertar opciones.
	for _, o := range req.Opciones {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO opciones_intermedias(pregunta_id,texto,es_correcta) VALUES($1,$2,$3)`,
			id, o.Texto, o.EsCorrecta)
		if err != nil {
			return nil, err
		}
	}
	pq := &PreguntaIntermedia{ID: id, CapacitacionID: req.CursoId, DespuesDeLeccionID: dlid,
		Texto: req.Texto, Tipo: req.Tipo, Orden: req.Orden}
	return pq, nil
}

func (r *postgresLeccionesRepository) DeletePregunta(ctx context.Context, preguntaID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM preguntas_intermedias WHERE id=$1`, preguntaID)
	return err
}

func (r *postgresLeccionesRepository) SubmitRespuestas(ctx context.Context, cursoID, userID string, respuestas []*leccionespb.Respuesta) (correctas, total int32, err error) {
	for _, resp := range respuestas {
		total++
		var esCorrecta bool
		r.db.QueryRowContext(ctx,
			`SELECT es_correcta FROM opciones_intermedias WHERE id=$1 AND pregunta_id=$2`,
			resp.OpcionId, resp.PreguntaId,
		).Scan(&esCorrecta)
		if esCorrecta {
			correctas++
		}
	}
	return correctas, total, nil
}
