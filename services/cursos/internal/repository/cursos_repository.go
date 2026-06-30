package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	Color          string     `db:"color"`
	Precio         float64    `db:"precio"`
	ScheduledAt    *time.Time `db:"scheduled_at"`
	Duration       int32      `db:"duration"`
	VideocallStatus *string   `db:"videocall_status"`
	CreatedAt      time.Time  `db:"created_at"`
}

func (c *Curso) ToProto() *cursospb.CursoResponse {
	r := &cursospb.CursoResponse{
		Id: c.ID, Title: c.Title, Description: c.Description, Type: c.Type,
		FilePath: c.FilePath, Content: c.Content, IsPublic: c.IsPublic,
		CodigoAcceso: c.CodigoAcceso, WelcomeMessage: c.WelcomeMessage,
		ThumbnailUrl: c.ThumbnailURL, Color: c.Color,
		Precio:    c.Precio,
		Duration:  c.Duration,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.InstructorID != nil {
		r.InstructorId = *c.InstructorID
	}
	if c.ScheduledAt != nil {
		r.ScheduledAt = c.ScheduledAt.Format("2006-01-02T15:04:05Z")
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

type Licencia struct {
	ID              string    `db:"id"`
	CapacitacionID  string    `db:"capacitacion_id"`
	Nombre          string    `db:"nombre"`
	Precio          float64   `db:"precio"`
	CapacidadMaxima int32     `db:"capacidad_maxima"`
	Usadas          int32     `db:"usadas"`
	CodigoAcceso    *string   `db:"codigo_acceso"`
	StripeProductID *string   `db:"stripe_product_id"`
	StripePriceID   *string   `db:"stripe_price_id"`
	CompradorID     *string   `db:"comprador_id"`
	CreatedAt       time.Time `db:"created_at"`
	CursoType       *string   `db:"curso_type"`
	CursoDuracion   *int32    `db:"curso_duracion"`
}

func (l *Licencia) ToProto() *cursospb.Licencia {
	r := &cursospb.Licencia{
		Id:              l.ID,
		CapacitacionId:  l.CapacitacionID,
		Nombre:          l.Nombre,
		Precio:          l.Precio,
		CapacidadMaxima: l.CapacidadMaxima,
		Usadas:          l.Usadas,
		CreatedAt:       l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if l.CursoType != nil {
		r.CursoType = *l.CursoType
	}
	if l.CursoDuracion != nil {
		r.CursoDuracion = *l.CursoDuracion
	}
	if l.CodigoAcceso != nil {
		r.CodigoAcceso = *l.CodigoAcceso
	}
	if l.StripeProductID != nil {
		r.StripeProductId = *l.StripeProductID
	}
	if l.StripePriceID != nil {
		r.StripePriceId = *l.StripePriceID
	}
	return r
}

// InstructorSchedule representa el horario disponible de un instructor
type InstructorSchedule struct {
	ID           string    `db:"id"`
	InstructorID string    `db:"instructor_id"`
	StartTime    time.Time `db:"start_time"`
	EndTime      time.Time `db:"end_time"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
}

func (s *InstructorSchedule) ToProto() *cursospb.InstructorSchedule {
	return &cursospb.InstructorSchedule{
		Id:           s.ID,
		InstructorId: s.InstructorID,
		StartTime:    s.StartTime.Format(time.RFC3339),
		EndTime:      s.EndTime.Format(time.RFC3339),
		Status:       s.Status,
		CreatedAt:    s.CreatedAt.Format(time.RFC3339),
	}
}

// VideocallTicket representa un acceso concurrente a una videollamada
type VideocallTicket struct {
	ID             string    `db:"id"`
	CapacitacionID string    `db:"capacitacion_id"`
	LicenciaID     *string   `db:"licencia_id"`
	Codigo         string    `db:"codigo"`
	InUseByUserID  *string   `db:"in_use_by_user_id"`
	IsValid        bool      `db:"is_valid"`
	CreatedAt      time.Time `db:"created_at"`
}

func (t *VideocallTicket) ToProto() *cursospb.VideocallTicket {
	r := &cursospb.VideocallTicket{
		Id:             t.ID,
		CapacitacionId: t.CapacitacionID,
		Codigo:         t.Codigo,
		IsValid:        t.IsValid,
	}
	if t.LicenciaID != nil {
		r.LicenciaId = *t.LicenciaID
	}
	if t.InUseByUserID != nil {
		r.InUseByUser = *t.InUseByUserID
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

	// Licencias
	CreateLicencia(ctx context.Context, req *cursospb.CreateLicenciaRequest) (*Licencia, error)
	CreateLicenciaB2BDirect(ctx context.Context, req *cursospb.WebhookComprarB2BDirectRequest, precioTotal float64) (*Licencia, error)
	UpdateLicencia(ctx context.Context, req *cursospb.UpdateLicenciaRequest) (*Licencia, error)
	DeleteLicencia(ctx context.Context, licenciaID string) error
	ListLicencias(ctx context.Context, cursoID string) ([]*Licencia, error)
	FindLicenciaByID(ctx context.Context, licenciaID string) (*Licencia, error)
	FindLicenciaByCodigo(ctx context.Context, codigo string) (*Licencia, error)
	

	// Horarios y Videollamadas
	ListSchedules(ctx context.Context, instructorID *string) ([]*InstructorSchedule, error)
	CreateSchedule(ctx context.Context, req *cursospb.CreateScheduleRequest) (*InstructorSchedule, error)
	UpdateSchedule(ctx context.Context, req *cursospb.UpdateScheduleRequest) (*InstructorSchedule, error)
	DeleteSchedule(ctx context.Context, scheduleID string) error
	
	CreateVideocallTickets(ctx context.Context, capacitacionID string, licenciaID *string, count int) ([]*VideocallTicket, error)
	JoinVideocall(ctx context.Context, codigo, userID string) (string, error) // retorna el room
	LeaveVideocall(ctx context.Context, codigo, userID string) error
	EndVideocall(ctx context.Context, cursoID string) error
	ListTicketsByLicencia(ctx context.Context, licenciaID string) ([]*VideocallTicket, error)
	
	IncrementarUsoLicencia(ctx context.Context, licenciaID string) error
	InscribirseConLicencia(ctx context.Context, userID, cursoID, licenciaID string) error
	ListLicenciasCompradas(ctx context.Context, userID string) ([]*Licencia, error)
	AsignarCompradorLicencia(ctx context.Context, licenciaID, userID string) error

	GetAdminDashboardStats(ctx context.Context) (*cursospb.AdminDashboardStatsResponse, error)
}

type postgresCursosRepository struct{ db *sqlx.DB }

func NewCursosRepository(db *sqlx.DB) CursosRepository {
	return &postgresCursosRepository{db: db}
}

const selectCurso = `SELECT id, title, COALESCE(description,'') description, type,
	COALESCE(file_path,'') file_path, COALESCE(content,'') content,
	instructor_id, is_public, COALESCE(codigo_acceso,'') codigo_acceso,
	COALESCE(welcome_message,'') welcome_message, COALESCE(thumbnail_url,'') thumbnail_url,
	COALESCE(color,'#f97316') color, precio, duration, created_at FROM capacitaciones`

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
		`INSERT INTO capacitaciones(title, description, type, file_path, content, instructor_id, is_public, welcome_message, thumbnail_url, color, precio, duration)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		req.Title, req.Description, req.Type, req.FilePath, req.Content, instructorID,
		req.IsPublic, req.WelcomeMessage, req.ThumbnailUrl, color, req.Precio, req.Duration,
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
		`UPDATE capacitaciones SET title=$1, description=$2, type=$3, file_path=$4, content=$5, is_public=$6, welcome_message=$7, thumbnail_url=$8, color=$9, precio=$10, duration=$11 WHERE id=$12`,
		req.Title, req.Description, req.Type, req.FilePath, req.Content,
		req.IsPublic, req.WelcomeMessage, req.ThumbnailUrl, color, req.Precio, req.Duration, req.CursoId,
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
		 VALUES($1,$2,$3,$4)
		 ON CONFLICT DO NOTHING`,
		userID, userName, userEmail, cursoID)
	return err
}

func (r *postgresCursosRepository) UnirseConCodigo(ctx context.Context, userID, codigo string) (*Curso, error) {
	curso, err := r.FindByCodigo(ctx, codigo)
	if err != nil {
		return nil, err
	}
	if curso.Precio > 0 {
		return nil, errors.New("este curso es de pago, el código de invitación general no es válido")
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

// ── Licencias ─────────────────────────────────────────────────────────────────

func (r *postgresCursosRepository) CreateLicencia(ctx context.Context, req *cursospb.CreateLicenciaRequest) (*Licencia, error) {
	codigo := uuid.New().String()[:12]
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO curso_licencias(capacitacion_id, nombre, precio, capacidad_maxima, codigo_acceso)
		 VALUES($1,$2,$3,$4,$5) RETURNING id`,
		req.CapacitacionId, req.Nombre, req.Precio, req.CapacidadMaxima, codigo,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindLicenciaByID(ctx, id)
}

func (r *postgresCursosRepository) CreateLicenciaB2BDirect(ctx context.Context, req *cursospb.WebhookComprarB2BDirectRequest, precioTotal float64) (*Licencia, error) {
	codigo := uuid.New().String()[:12]
	var id string
	nombre := "Licencia Corporativa"
	// Guardamos el stripe_session_id en stripe_product_id para recuperar la factura después
	var stripeSessionID *string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-stripe-session-id"); len(vals) > 0 && vals[0] != "" {
			stripeSessionID = &vals[0]
		}
	}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO curso_licencias(capacitacion_id, nombre, precio, capacidad_maxima, codigo_acceso, comprador_id, stripe_product_id)
		 SELECT $1,$2,$3,$4,$5,$6,$7
		 WHERE NOT EXISTS (
		    SELECT 1 FROM curso_licencias WHERE stripe_product_id = $7 AND $7 IS NOT NULL
		 )
		 RETURNING id`,
		req.CursoId, nombre, precioTotal, req.Cantidad, codigo, req.UserId, stripeSessionID,
	).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			// Ya se insertó por el webhook o la sesión, simplemente la buscamos y la retornamos
			err2 := r.db.QueryRowContext(ctx, `SELECT id FROM curso_licencias WHERE stripe_product_id = $1 LIMIT 1`, stripeSessionID).Scan(&id)
			if err2 != nil {
				return nil, fmt.Errorf("licencia duplicada detectada pero no se pudo recuperar: %w", err2)
			}
			return r.FindLicenciaByID(ctx, id)
		}
		return nil, err
	}
	return r.FindLicenciaByID(ctx, id)
}

func (r *postgresCursosRepository) UpdateLicencia(ctx context.Context, req *cursospb.UpdateLicenciaRequest) (*Licencia, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE curso_licencias SET nombre=$1, precio=$2, capacidad_maxima=$3 WHERE id=$4`,
		req.Nombre, req.Precio, req.CapacidadMaxima, req.Id,
	)
	if err != nil {
		return nil, err
	}
	return r.FindLicenciaByID(ctx, req.Id)
}

func (r *postgresCursosRepository) DeleteLicencia(ctx context.Context, licenciaID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM curso_licencias WHERE id=$1`, licenciaID)
	return err
}

func (r *postgresCursosRepository) ListLicencias(ctx context.Context, cursoID string) ([]*Licencia, error) {
	var lics []*Licencia
	return lics, r.db.SelectContext(ctx, &lics,
		`SELECT id, capacitacion_id, nombre, precio, capacidad_maxima, usadas, codigo_acceso, stripe_product_id, stripe_price_id, comprador_id, created_at FROM curso_licencias WHERE capacitacion_id=$1 ORDER BY created_at DESC`, cursoID)
}

func (r *postgresCursosRepository) FindLicenciaByID(ctx context.Context, licenciaID string) (*Licencia, error) {
	l := &Licencia{}
	return l, r.db.GetContext(ctx, l, `SELECT id, capacitacion_id, nombre, precio, capacidad_maxima, usadas, codigo_acceso, stripe_product_id, stripe_price_id, comprador_id, created_at FROM curso_licencias WHERE id=$1`, licenciaID)
}

func (r *postgresCursosRepository) FindLicenciaByCodigo(ctx context.Context, codigo string) (*Licencia, error) {
	l := &Licencia{}
	return l, r.db.GetContext(ctx, l, `SELECT id, capacitacion_id, nombre, precio, capacidad_maxima, usadas, codigo_acceso, stripe_product_id, stripe_price_id, comprador_id, created_at FROM curso_licencias WHERE codigo_acceso=$1`, codigo)
}

func (r *postgresCursosRepository) IncrementarUsoLicencia(ctx context.Context, licenciaID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE curso_licencias SET usadas = usadas + 1 WHERE id=$1`, licenciaID)
	return err
}

func (r *postgresCursosRepository) InscribirseConLicencia(ctx context.Context, userID, cursoID, licenciaID string) error {
	var licID interface{}
	if licenciaID == "" {
		licID = nil
	} else {
		licID = licenciaID
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO inscripciones(user_id, capacitacion_id, licencia_id)
		 VALUES($1,$2,$3)
		 ON CONFLICT DO NOTHING`,
		userID, cursoID, licID)
	return err
}

func (r *postgresCursosRepository) ListLicenciasCompradas(ctx context.Context, userID string) ([]*Licencia, error) {
	var lics []*Licencia
	return lics, r.db.SelectContext(ctx, &lics,
		`SELECT l.id, l.capacitacion_id, l.nombre, l.precio, l.capacidad_maxima, l.usadas, l.codigo_acceso, l.stripe_product_id, l.stripe_price_id, l.comprador_id, l.created_at, c.type as curso_type, c.duration as curso_duracion
		 FROM curso_licencias l
		 LEFT JOIN capacitaciones c ON c.id = l.capacitacion_id
		 WHERE l.comprador_id=$1 
		 ORDER BY l.created_at DESC`, userID)
}

func (r *postgresCursosRepository) AsignarCompradorLicencia(ctx context.Context, licenciaID, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE curso_licencias SET comprador_id=$1 WHERE id=$2`, userID, licenciaID)
	return err
}

// errForbidden es un error de dominio para acceso denegado.
var errForbidden = &forbiddenError{}

type forbiddenError struct{}

func (e *forbiddenError) Error() string { return "forbidden" }

func (r *postgresCursosRepository) GetAdminDashboardStats(ctx context.Context) (*cursospb.AdminDashboardStatsResponse, error) {
	stats := &cursospb.AdminDashboardStatsResponse{}

	// B2B Licencias vendidas (comprador_id no nulo)
	var ventasB2B struct {
		Total float64 `db:"total"`
		Count int32   `db:"count"`
	}
	err := r.db.GetContext(ctx, &ventasB2B, `SELECT COALESCE(SUM(precio), 0) as total, COUNT(*) as count FROM curso_licencias WHERE comprador_id IS NOT NULL`)
	if err != nil {
		return nil, err
	}

	// B2C Inscripciones directas (licencia_id IS NULL) cruzando con capacitaciones.precio
	var ventasB2C struct {
		Total float64 `db:"total"`
		Count int32   `db:"count"`
	}
	err = r.db.GetContext(ctx, &ventasB2C, `
		SELECT COALESCE(SUM(c.precio), 0) as total, COUNT(*) as count
		FROM inscripciones i
		JOIN capacitaciones c ON c.id = i.capacitacion_id
		WHERE i.licencia_id IS NULL AND c.precio > 0
	`)
	if err != nil {
		return nil, err
	}

	stats.TotalVentasBrutas = float32(ventasB2B.Total + ventasB2C.Total)
	
	// Calcular netas aproximadas (Bruto - 3.6% - 3 MXN por transacción)
	totalTransacciones := ventasB2B.Count + ventasB2C.Count
	var totalNeto float64 = 0
	if stats.TotalVentasBrutas > 0 {
		totalNeto = float64(stats.TotalVentasBrutas) - (float64(stats.TotalVentasBrutas)*0.036 + (float64(totalTransacciones) * 3.0))
		if totalNeto < 0 {
			totalNeto = 0
		}
	}

	stats.TotalVentasNetas = float32(totalNeto)
	stats.TotalTransacciones = totalTransacciones
	stats.LicenciasVendidas = ventasB2B.Count
	stats.ComprasIndividuales = ventasB2C.Count

	return stats, nil
}
