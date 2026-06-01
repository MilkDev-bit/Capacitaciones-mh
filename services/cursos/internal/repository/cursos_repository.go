package repository

import (
	"context"
	"time"

	cursospb "Prueba-Go/gen/cursos"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/metadata"
)

// Curso es el modelo interno del servicio de cursos.
type Curso struct {
	ID             string    `db:"id"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	Type           string    `db:"type"`
	FilePath       string    `db:"file_path"`
	Content        string    `db:"content"`
	InstructorID   *string   `db:"instructor_id"`
	IsPublic       bool      `db:"is_public"`
	CodigoAcceso   string    `db:"codigo_acceso"`
	WelcomeMessage string    `db:"welcome_message"`
	ThumbnailURL   string    `db:"thumbnail_url"`
	Color          string    `db:"color"`
	CreatedAt      time.Time `db:"created_at"`
}

func (c *Curso) ToProto() *cursospb.CursoResponse {
	r := &cursospb.CursoResponse{
		Id: c.ID, Title: c.Title, Description: c.Description, Type: c.Type,
		FilePath: c.FilePath, Content: c.Content, IsPublic: c.IsPublic,
		CodigoAcceso: c.CodigoAcceso, WelcomeMessage: c.WelcomeMessage,
		ThumbnailUrl: c.ThumbnailURL, Color: c.Color,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.InstructorID != nil {
		r.InstructorId = *c.InstructorID
	}
	return r
}

// Asignacion representa la tabla de asignaciones.
type Asignacion struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	UserName       string    `db:"user_name"`
	UserEmail      string    `db:"user_email"`
	CapacitacionID *string   `db:"capacitacion_id"`
	AssignedAt     time.Time `db:"assigned_at"`
}

func (a *Asignacion) ToProto() *cursospb.AsignacionInfo {
	r := &cursospb.AsignacionInfo{
		Id: a.ID, UserId: a.UserID,
		AssignedAt: a.AssignedAt.Format("2006-01-02T15:04:05Z"),
	}
	if a.CapacitacionID != nil {
		r.CapacitacionId = *a.CapacitacionID
	}
	return r
}

// EstudianteRow para listar estudiantes de un curso.
type EstudianteRow struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Email      string    `db:"email"`
	AssignedAt time.Time `db:"assigned_at"`
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

// CursosRepository define el contrato de acceso a datos.
type CursosRepository interface {
	List(ctx context.Context) ([]*Curso, error)
	ListPublicos(ctx context.Context) ([]*Curso, error)
	ListByUser(ctx context.Context, userID string) ([]*Curso, error)
	ListByInstructor(ctx context.Context, instructorID string) ([]*Curso, error)
	FindByID(ctx context.Context, cursoID string) (*Curso, error)
	FindByCodigo(ctx context.Context, codigo string) (*Curso, error)
	Create(ctx context.Context, req *cursospb.CreateCursoRequest) (*Curso, error)
	Update(ctx context.Context, req *cursospb.UpdateCursoRequest) (*Curso, error)
	Delete(ctx context.Context, cursoID string) error
	TogglePublic(ctx context.Context, cursoID string) (*Curso, error)
	ResetCodigo(ctx context.Context, cursoID string) (*Curso, error)

	IsEnrolled(ctx context.Context, userID, cursoID string) (bool, error)
	Inscribirse(ctx context.Context, userID, cursoID string) error
	UnirseConCodigo(ctx context.Context, userID, codigo string) (*Curso, error)

	ListEstudiantes(ctx context.Context, instructorID, cursoID string) ([]*EstudianteRow, error)
	InstructorAsignar(ctx context.Context, instructorID, userID, cursoID string) error

	ListAsignaciones(ctx context.Context) ([]*Asignacion, error)
	AdminAsignar(ctx context.Context, userID, cursoID string) error
	DesAsignar(ctx context.Context, asignacionID string) error
}

type postgresCursosRepository struct{ db *sqlx.DB }

func NewCursosRepository(db *sqlx.DB) CursosRepository {
	return &postgresCursosRepository{db: db}
}

const selectCurso = `SELECT id, title, COALESCE(description,'') description, type,
	COALESCE(file_path,'') file_path, COALESCE(content,'') content,
	instructor_id, is_public, COALESCE(codigo_acceso,'') codigo_acceso,
	COALESCE(welcome_message,'') welcome_message, COALESCE(thumbnail_url,'') thumbnail_url,
	COALESCE(color,'#f97316') color, created_at FROM capacitaciones`

func (r *postgresCursosRepository) List(ctx context.Context) ([]*Curso, error) {
	var cursos []*Curso
	return cursos, r.db.SelectContext(ctx, &cursos,
		selectCurso+` WHERE deleted_at IS NULL ORDER BY created_at DESC`)
}

func (r *postgresCursosRepository) ListPublicos(ctx context.Context) ([]*Curso, error) {
	var cursos []*Curso
	return cursos, r.db.SelectContext(ctx, &cursos,
		selectCurso+` WHERE deleted_at IS NULL AND is_public=true ORDER BY created_at DESC`)
}

func (r *postgresCursosRepository) ListByUser(ctx context.Context, userID string) ([]*Curso, error) {
	var cursos []*Curso
	return cursos, r.db.SelectContext(ctx, &cursos,
		selectCurso+` WHERE deleted_at IS NULL AND id IN
		(SELECT capacitacion_id FROM asignaciones WHERE user_id=$1 AND capacitacion_id IS NOT NULL)
		ORDER BY created_at DESC`, userID)
}

func (r *postgresCursosRepository) ListByInstructor(ctx context.Context, instructorID string) ([]*Curso, error) {
	var cursos []*Curso
	return cursos, r.db.SelectContext(ctx, &cursos,
		selectCurso+` WHERE deleted_at IS NULL AND instructor_id=$1 ORDER BY created_at DESC`, instructorID)
}

func (r *postgresCursosRepository) FindByID(ctx context.Context, cursoID string) (*Curso, error) {
	c := &Curso{}
	return c, r.db.GetContext(ctx, c, selectCurso+` WHERE id=$1 AND deleted_at IS NULL`, cursoID)
}

func (r *postgresCursosRepository) FindByCodigo(ctx context.Context, codigo string) (*Curso, error) {
	c := &Curso{}
	return c, r.db.GetContext(ctx, c, selectCurso+` WHERE codigo_acceso=$1 AND deleted_at IS NULL`, codigo)
}

func (r *postgresCursosRepository) Create(ctx context.Context, req *cursospb.CreateCursoRequest) (*Curso, error) {
	color := req.Color
	if color == "" {
		color = "#f97316"
	}
	var instructorID *string
	if req.UserId != "" {
		instructorID = &req.UserId
	}
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO capacitaciones(title,description,type,file_path,content,instructor_id,
		 is_public,welcome_message,thumbnail_url,color)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		req.Title, req.Description, req.Type, req.FilePath, req.Content, instructorID,
		req.IsPublic, req.WelcomeMessage, req.ThumbnailUrl, color,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *postgresCursosRepository) Update(ctx context.Context, req *cursospb.UpdateCursoRequest) (*Curso, error) {
	color := req.Color
	if color == "" {
		color = "#f97316"
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE capacitaciones SET title=$1,description=$2,type=$3,file_path=$4,content=$5,
		 is_public=$6,welcome_message=$7,thumbnail_url=$8,color=$9 WHERE id=$10`,
		req.Title, req.Description, req.Type, req.FilePath, req.Content,
		req.IsPublic, req.WelcomeMessage, req.ThumbnailUrl, color, req.CursoId,
	)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, req.CursoId)
}

func (r *postgresCursosRepository) Delete(ctx context.Context, cursoID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE capacitaciones SET deleted_at=NOW() WHERE id=$1`, cursoID)
	return err
}

func (r *postgresCursosRepository) TogglePublic(ctx context.Context, cursoID string) (*Curso, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE capacitaciones SET is_public = NOT is_public WHERE id=$1`, cursoID)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, cursoID)
}

func (r *postgresCursosRepository) ResetCodigo(ctx context.Context, cursoID string) (*Curso, error) {
	newCode := uuid.New().String()[:8]
	_, err := r.db.ExecContext(ctx,
		`UPDATE capacitaciones SET codigo_acceso=$1 WHERE id=$2`, newCode, cursoID)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, cursoID)
}

func (r *postgresCursosRepository) IsEnrolled(ctx context.Context, userID, cursoID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM asignaciones WHERE user_id=$1 AND capacitacion_id=$2`, userID, cursoID,
	).Scan(&count)
	return count > 0, err
}

func (r *postgresCursosRepository) Inscribirse(ctx context.Context, userID, cursoID string) error {
	userName := metaVal(ctx, "x-user-name")
	userEmail := metaVal(ctx, "x-user-email")
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO asignaciones(user_id,user_name,user_email,capacitacion_id)
		 VALUES($1,$2,$3,$4) ON CONFLICT (user_id, capacitacion_id) DO UPDATE
		 SET user_name=EXCLUDED.user_name, user_email=EXCLUDED.user_email`,
		userID, userName, userEmail, cursoID)
	return err
}

func (r *postgresCursosRepository) UnirseConCodigo(ctx context.Context, userID, codigo string) (*Curso, error) {
	curso, err := r.FindByCodigo(ctx, codigo)
	if err != nil {
		return nil, err
	}
	if err := r.Inscribirse(ctx, userID, curso.ID); err != nil {
		return nil, err
	}
	return curso, nil
}

func (r *postgresCursosRepository) ListEstudiantes(ctx context.Context, instructorID, cursoID string) ([]*EstudianteRow, error) {
	var rows []*EstudianteRow
	if cursoID == "" {
		// Sin filtro de curso: todos los estudiantes de todos los cursos del instructor.
		return rows, r.db.SelectContext(ctx, &rows,
			`SELECT DISTINCT a.user_id id,
			        COALESCE(a.user_name,'') name,
			        COALESCE(a.user_email,'') email,
			        a.assigned_at
			   FROM asignaciones a
			   JOIN capacitaciones c ON c.id = a.capacitacion_id
			  WHERE c.instructor_id = $1
			  ORDER BY a.assigned_at DESC`, instructorID)
	}
	return rows, r.db.SelectContext(ctx, &rows,
		`SELECT user_id id,
		        COALESCE(user_name,'') name,
		        COALESCE(user_email,'') email,
		        assigned_at
		   FROM asignaciones
		  WHERE capacitacion_id=$1
		    AND EXISTS(SELECT 1 FROM capacitaciones c WHERE c.id=$1 AND c.instructor_id=$2)
		  ORDER BY assigned_at DESC`, cursoID, instructorID)
}

func (r *postgresCursosRepository) InstructorAsignar(ctx context.Context, instructorID, userID, cursoID string) error {
	var owner string
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, cursoID,
	).Scan(&owner); err != nil || owner != instructorID {
		return errForbidden
	}
	return r.Inscribirse(ctx, userID, cursoID)
}

func (r *postgresCursosRepository) ListAsignaciones(ctx context.Context) ([]*Asignacion, error) {
	var asigs []*Asignacion
	return asigs, r.db.SelectContext(ctx, &asigs,
		`SELECT id, user_id,
		        COALESCE(user_name,'') user_name,
		        COALESCE(user_email,'') user_email,
		        capacitacion_id, assigned_at
		   FROM asignaciones
		  WHERE capacitacion_id IS NOT NULL ORDER BY assigned_at DESC`)
}

func (r *postgresCursosRepository) AdminAsignar(ctx context.Context, userID, cursoID string) error {
	return r.Inscribirse(ctx, userID, cursoID)
}

func (r *postgresCursosRepository) DesAsignar(ctx context.Context, asignacionID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM asignaciones WHERE id=$1`, asignacionID)
	return err
}

// errForbidden es un error de dominio para acceso denegado.
var errForbidden = &forbiddenError{}

type forbiddenError struct{}

func (e *forbiddenError) Error() string { return "forbidden" }
