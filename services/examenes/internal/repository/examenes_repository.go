package repository

import (
	"context"
	"time"

	examenespb "Prueba-Go/gen/examenes"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/metadata"
)

type Examen struct {
	ID             string    `db:"id"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	InstructorID   *string   `db:"instructor_id"`
	CapacitacionID *string   `db:"capacitacion_id"`
	CreatedAt      time.Time `db:"created_at"`
}

type Pregunta struct {
	ID       string  `db:"id"`
	ExamenID string  `db:"examen_id"`
	Texto    string  `db:"texto"`
	Tipo     string  `db:"tipo"`
	Valor    float64 `db:"valor"`
	Orden    int32   `db:"orden"`
}

type Opcion struct {
	ID         string `db:"id"`
	PreguntaID string `db:"pregunta_id"`
	Texto      string `db:"texto"`
	EsCorrecta bool   `db:"es_correcta"`
}

type ResultadoRow struct {
	UserID     string  `db:"user_id"`
	UserName   string  `db:"user_name"`
	Puntaje    float64 `db:"puntaje"`
	Porcentaje float64 `db:"porcentaje"`
}

// ExamenesRepository contrato.
type ExamenesRepository interface {
	List(ctx context.Context) ([]*Examen, error)
	ListByInstructor(ctx context.Context, instructorID string) ([]*Examen, error)
	ListByUser(ctx context.Context, userID string) ([]*Examen, error)
	FindByID(ctx context.Context, examenID string) (*Examen, error)
	GetPreguntas(ctx context.Context, examenID string) ([]*Pregunta, error)
	GetOpciones(ctx context.Context, preguntaID string) ([]*Opcion, error)
	Create(ctx context.Context, req *examenespb.CreateExamenRequest) (*Examen, error)
	Delete(ctx context.Context, examenID string) error
	SubmitRespuestas(ctx context.Context, examenID, userID string, respuestas []*examenespb.RespuestaInput) (*examenespb.ResultadoResponse, error)
	GetResultados(ctx context.Context, examenID string) ([]*ResultadoRow, error)
	GetRespuestasUsuario(ctx context.Context, examenID, userID string) (*examenespb.RespuestasResponse, error)
}

type postgresExamenesRepository struct{ db *sqlx.DB }

func NewExamenesRepository(db *sqlx.DB) ExamenesRepository {
	return &postgresExamenesRepository{db: db}
}

// metaVal extrae un valor del gRPC incoming metadata.
func metaVal(ctx context.Context, key string) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(key); len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}

func (r *postgresExamenesRepository) List(ctx context.Context) ([]*Examen, error) {
	var e []*Examen
	return e, r.db.SelectContext(ctx, &e,
		`SELECT id, title, COALESCE(description,'') description, instructor_id,
		        capacitacion_id, created_at FROM examenes WHERE deleted_at IS NULL ORDER BY created_at DESC`)
}

func (r *postgresExamenesRepository) ListByInstructor(ctx context.Context, instructorID string) ([]*Examen, error) {
	var e []*Examen
	return e, r.db.SelectContext(ctx, &e,
		`SELECT id, title, COALESCE(description,'') description, instructor_id,
		        capacitacion_id, created_at FROM examenes
		  WHERE deleted_at IS NULL AND instructor_id=$1 ORDER BY created_at DESC`, instructorID)
}

func (r *postgresExamenesRepository) ListByUser(ctx context.Context, userID string) ([]*Examen, error) {
	var e []*Examen
	return e, r.db.SelectContext(ctx, &e,
		`SELECT DISTINCT ex.id, ex.title, COALESCE(ex.description,'') description, ex.instructor_id,
		        ex.capacitacion_id, ex.created_at
		   FROM examenes ex
		   LEFT JOIN asignaciones_examen ae ON ae.examen_id = ex.id AND ae.user_id = $1
		   LEFT JOIN inscripciones i ON i.capacitacion_id = ex.capacitacion_id AND i.user_id = $1
		   LEFT JOIN asignaciones a ON a.capacitacion_id = ex.capacitacion_id AND a.user_id = $1
		   LEFT JOIN capacitaciones c ON c.id = ex.capacitacion_id
		  WHERE ex.deleted_at IS NULL
		    AND (
		        ae.user_id IS NOT NULL OR
		        i.user_id IS NOT NULL OR
		        a.user_id IS NOT NULL OR
		        COALESCE(c.is_public, false) = true OR
		        EXISTS (
		            SELECT 1 FROM progreso_lecciones pl
		            JOIN lecciones l ON l.id = pl.leccion_id
		            JOIN modulos m ON m.id = l.modulo_id
		            WHERE m.capacitacion_id = ex.capacitacion_id AND pl.user_id = $1
		        )
		    )
		  ORDER BY ex.created_at DESC`, userID)
}

func (r *postgresExamenesRepository) FindByID(ctx context.Context, examenID string) (*Examen, error) {
	e := &Examen{}
	return e, r.db.GetContext(ctx, e,
		`SELECT id, title, COALESCE(description,'') description, instructor_id,
		        capacitacion_id, created_at FROM examenes WHERE id=$1 AND deleted_at IS NULL`, examenID)
}

func (r *postgresExamenesRepository) GetPreguntas(ctx context.Context, examenID string) ([]*Pregunta, error) {
	var pqs []*Pregunta
	return pqs, r.db.SelectContext(ctx, &pqs,
		`SELECT id, examen_id, texto, tipo, valor, orden FROM preguntas
		  WHERE examen_id=$1 ORDER BY orden`, examenID)
}

func (r *postgresExamenesRepository) GetOpciones(ctx context.Context, preguntaID string) ([]*Opcion, error) {
	var opts []*Opcion
	return opts, r.db.SelectContext(ctx, &opts,
		`SELECT id, pregunta_id, texto, es_correcta FROM opciones WHERE pregunta_id=$1`, preguntaID)
}

func (r *postgresExamenesRepository) Create(ctx context.Context, req *examenespb.CreateExamenRequest) (*Examen, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var instructorID *string
	if req.UserId != "" {
		instructorID = &req.UserId
	}
	var capID *string
	if req.CapacitacionId != "" {
		capID = &req.CapacitacionId
	}

	var examenID string
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO examenes(title,description,instructor_id,capacitacion_id)
		 VALUES($1,$2,$3,$4) RETURNING id`,
		req.Title, req.Description, instructorID, capID,
	).Scan(&examenID); err != nil {
		return nil, err
	}

	for _, p := range req.Preguntas {
		tipo := p.Tipo
		if tipo == "" {
			tipo = "multiple_choice"
		}
		var preguntaID string
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO preguntas(examen_id,texto,tipo,valor,orden) VALUES($1,$2,$3,$4,$5) RETURNING id`,
			examenID, p.Texto, tipo, p.Valor, p.Orden,
		).Scan(&preguntaID); err != nil {
			return nil, err
		}
		for _, o := range p.Opciones {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO opciones(pregunta_id,texto,es_correcta) VALUES($1,$2,$3)`,
				preguntaID, o.Texto, o.EsCorrecta); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, examenID)
}

func (r *postgresExamenesRepository) Delete(ctx context.Context, examenID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE examenes SET deleted_at=NOW() WHERE id=$1`, examenID)
	return err
}

func (r *postgresExamenesRepository) SubmitRespuestas(ctx context.Context, examenID, userID string, respuestas []*examenespb.RespuestaInput) (*examenespb.ResultadoResponse, error) {
	userName := metaVal(ctx, "x-user-name")

	// Obtener preguntas del examen para calcular puntaje.
	preguntas, err := r.GetPreguntas(ctx, examenID)
	if err != nil {
		return nil, err
	}

	// Map preguntaID → pregunta.
	pqMap := make(map[string]*Pregunta, len(preguntas))
	var puntajeMax float64
	for _, p := range preguntas {
		pqMap[p.ID] = p
		puntajeMax += p.Valor
	}

	var puntaje float64
	for _, resp := range respuestas {
		pq, ok := pqMap[resp.PreguntaId]
		if !ok || pq.Tipo == "open_text" {
			continue
		}
		var esCorrecta bool
		r.db.QueryRowContext(ctx,
			`SELECT es_correcta FROM opciones WHERE id=$1 AND pregunta_id=$2`,
			resp.OpcionId, resp.PreguntaId,
		).Scan(&esCorrecta)
		if esCorrecta {
			puntaje += pq.Valor
		}
		// Guardar respuesta con user_name denormalizado.
		r.db.ExecContext(ctx,
			`INSERT INTO respuestas_examen(examen_id,user_id,user_name,pregunta_id,opcion_id,respuesta_texto)
			 VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(user_id,pregunta_id) DO UPDATE
			 SET opcion_id=EXCLUDED.opcion_id, respuesta_texto=EXCLUDED.respuesta_texto,
			     user_name=EXCLUDED.user_name`,
			examenID, userID, userName, resp.PreguntaId, resp.OpcionId, resp.RespuestaTexto)
	}

	// Registrar en asignaciones_examen para que ListByUser funcione.
	r.db.ExecContext(ctx,
		`INSERT INTO asignaciones_examen(examen_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		examenID, userID)

	porcentaje := 0.0
	if puntajeMax > 0 {
		porcentaje = puntaje / puntajeMax * 100
	}

	examen, _ := r.FindByID(ctx, examenID)
	return &examenespb.ResultadoResponse{
		ExamenId: examenID, Titulo: examen.Title,
		Puntaje: puntaje, PuntajeMax: puntajeMax, Porcentaje: porcentaje,
	}, nil
}

func (r *postgresExamenesRepository) GetResultados(ctx context.Context, examenID string) ([]*ResultadoRow, error) {
	var rows []*ResultadoRow
	return rows, r.db.SelectContext(ctx, &rows,
		`SELECT re.user_id,
		        COALESCE(re.user_name,'') user_name,
		        COALESCE(SUM(CASE WHEN o.es_correcta THEN p.valor ELSE 0 END),0) puntaje,
		        CASE WHEN SUM(p.valor)>0
		             THEN SUM(CASE WHEN o.es_correcta THEN p.valor ELSE 0 END)/SUM(p.valor)*100
		             ELSE 0 END porcentaje
		   FROM respuestas_examen re
		   JOIN preguntas p ON p.id = re.pregunta_id
		   LEFT JOIN opciones o ON o.id = re.opcion_id
		  WHERE re.examen_id = $1
		  GROUP BY re.user_id, re.user_name ORDER BY porcentaje DESC`, examenID)
}

func (r *postgresExamenesRepository) GetRespuestasUsuario(ctx context.Context, examenID, userID string) (*examenespb.RespuestasResponse, error) {
	type row struct {
		PreguntaID     string  `db:"pregunta_id"`
		PreguntaTexto  string  `db:"pregunta_texto"`
		OpcionID       string  `db:"opcion_id"`
		RespuestaTexto string  `db:"respuesta_texto"`
		EsCorrecta     bool    `db:"es_correcta"`
		Valor          float64 `db:"valor"`
	}
	var rows []*row
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT re.pregunta_id, p.texto pregunta_texto,
		        COALESCE(re.opcion_id::text,'') opcion_id,
		        COALESCE(re.respuesta_texto,'') respuesta_texto,
		        COALESCE(o.es_correcta, false) es_correcta,
		        p.valor
		   FROM respuestas_examen re
		   JOIN preguntas p ON p.id = re.pregunta_id
		   LEFT JOIN opciones o ON o.id = re.opcion_id
		  WHERE re.examen_id=$1 AND re.user_id=$2`, examenID, userID); err != nil {
		return nil, err
	}
	var puntaje float64
	var puntajeMax float64
	details := make([]*examenespb.RespuestaDetalle, 0, len(rows))
	for _, rw := range rows {
		puntajeMax += rw.Valor
		if rw.EsCorrecta {
			puntaje += rw.Valor
		}
		details = append(details, &examenespb.RespuestaDetalle{
			PreguntaId: rw.PreguntaID, PreguntaTexto: rw.PreguntaTexto,
			OpcionId: rw.OpcionID, RespuestaTexto: rw.RespuestaTexto,
			EsCorrecta: rw.EsCorrecta,
		})
	}
	porcentaje := 0.0
	if puntajeMax > 0 {
		porcentaje = puntaje / puntajeMax * 100
	}
	return &examenespb.RespuestasResponse{
		Respuestas: details, Puntaje: puntaje, Porcentaje: porcentaje,
	}, nil
}
