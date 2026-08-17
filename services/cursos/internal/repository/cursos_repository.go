package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	cursospb "Prueba-Go/gen/cursos"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/metadata"
)

// Curso es el modelo interno del servicio de cursos.
type Curso struct {
	ID             string  `db:"id"`
	Title          string  `db:"title"`
	Description    string  `db:"description"`
	Type           string  `db:"type"`
	FilePath       string  `db:"file_path"`
	Content        string  `db:"content"`
	InstructorID   *string `db:"instructor_id"`
	IsPublic       bool    `db:"is_public"`
	CodigoAcceso   string  `db:"codigo_acceso"`
	WelcomeMessage string  `db:"welcome_message"`
	ThumbnailURL   string  `db:"thumbnail_url"`
	Color          string  `db:"color"`
	Precio         float64 `db:"precio"`
	PrecioCentavos       int64      `db:"precio_centavos"`
	ScheduledAt          *time.Time `db:"scheduled_at"`
	Duration             int32      `db:"duration"`
	DC3Enabled           bool       `db:"dc3_enabled"`
	CreatedAt            time.Time  `db:"created_at"`
	TotalLecciones       int32      `db:"total_lecciones"`
	LeccionesCompletadas int32      `db:"lecciones_completadas"`

	// Área temática del catálogo STPS para la constancia. Es del curso, no del
	// instructor ni del patrón: cada temario tiene su clave.
	//
	// Solo lo rellenan las consultas que usan selectCurso; ListByUser lo deja
	// vacío y no pasa nada, porque la constancia se arma desde GetDatosDC3.
	DC3AreaTematica string `db:"dc3_area_tematica"`
	// Nombre oficial del curso para la DC-3; vacío usa Title.
	DC3NombreCurso string `db:"dc3_nombre_curso"`
}

func (c *Curso) ToProto() *cursospb.CursoResponse {
	r := &cursospb.CursoResponse{
		Id: c.ID, Title: c.Title, Description: c.Description, Type: c.Type,
		FilePath: c.FilePath, Content: c.Content, IsPublic: c.IsPublic,
		CodigoAcceso: c.CodigoAcceso, WelcomeMessage: c.WelcomeMessage,
		ThumbnailUrl: c.ThumbnailURL, Color: c.Color,
		Precio:               c.Precio,
		Duration:             c.Duration,
		Dc3Enabled:           c.DC3Enabled,
		Dc3AreaTematica:      c.DC3AreaTematica,
		Dc3NombreCurso:       c.DC3NombreCurso,
		TotalLecciones:       c.TotalLecciones,
		LeccionesCompletadas: c.LeccionesCompletadas,
		CreatedAt:            c.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
	PrecioCentavos  int64     `db:"precio_centavos"`
	CapacidadMaxima int32     `db:"capacidad_maxima"`
	Usadas          int32     `db:"usadas"`
	CodigoAcceso    *string   `db:"codigo_acceso"`
	StripeProductID *string   `db:"stripe_product_id"`
	StripePriceID   *string   `db:"stripe_price_id"`
	CompradorID     *string   `db:"comprador_id"`
	CreatedAt       time.Time `db:"created_at"`
	CursoType       *string   `db:"curso_type"`
	CursoDuracion   *int32    `db:"curso_duracion"`
	// Campos derivados del JOIN en ListLicenciasCompradas.
	CapacitacionTitulo *string `db:"capacitacion_titulo"`
	AccesosEnviados    *int32  `db:"accesos_enviados"`
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
	if l.CompradorID != nil {
		r.CompradorId = *l.CompradorID
	}
	if l.CapacitacionTitulo != nil {
		r.CapacitacionTitulo = *l.CapacitacionTitulo
	}
	if l.AccesosEnviados != nil {
		r.AccesosEnviados = *l.AccesosEnviados
	}
	return r
}

// InvitacionLicencia es un acceso de una licencia corporativa ya enviado por
// correo a un participante concreto.
type InvitacionLicencia struct {
	ID         string    `db:"id"`
	LicenciaID string    `db:"licencia_id"`
	Nombre     string    `db:"nombre"`
	Email      string    `db:"email"`
	Codigo     string    `db:"codigo"`
	Estado     string    `db:"estado"`
	EnviadoAt  time.Time `db:"enviado_at"`
}

func (i *InvitacionLicencia) ToProto() *cursospb.InvitacionLicencia {
	return &cursospb.InvitacionLicencia{
		Id:         i.ID,
		LicenciaId: i.LicenciaID,
		Nombre:     i.Nombre,
		Email:      i.Email,
		Codigo:     i.Codigo,
		Estado:     i.Estado,
		EnviadoAt:  i.EnviadoAt.Format(time.RFC3339),
	}
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
	// Órdenes, idempotencia y deduplicación de webhooks.
	OrdenesRepository
	// Planes, suscripciones y asientos.
	SuscripcionesRepository

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

	IncrementarUsoLicencia(ctx context.Context, licenciaID string) error
	InscribirseConLicencia(ctx context.Context, userID, cursoID, licenciaID string) error
	ListLicenciasCompradas(ctx context.Context, userID string) ([]*Licencia, error)
	AsignarCompradorLicencia(ctx context.Context, licenciaID, userID string) error

	// Reparto de accesos corporativos por correo
	ListInvitacionesLicencia(ctx context.Context, licenciaID string) ([]*InvitacionLicencia, error)
	// AsignarAccesos registra una invitación por participante en una sola
	// transacción, validando el cupo de la licencia.
	AsignarAccesos(ctx context.Context, licenciaID, codigoCompartido string, participantes []Participante) ([]*InvitacionLicencia, error)

	// FindLicenciaDeInscripcion devuelve la licencia con la que un usuario se
	// inscribió a un curso. Devuelve nil si fue una compra individual.
	FindLicenciaDeInscripcion(ctx context.Context, userID, cursoID string) (*Licencia, error)
	// RegistrarAvisoDC3 marca el aviso como enviado y devuelve true solo la
	// primera vez, para no spamear al representante en cada participante.
	RegistrarAvisoDC3(ctx context.Context, licenciaID, cursoID string) (bool, error)

	// ── Constancias DC-3 ──────────────────────────────────────────────────
	// Find* devuelven (nil, nil) cuando el dato aún no existe: en este flujo
	// "todavía no lo ha capturado" es el caso normal, no un error.
	FindDatosTrabajador(ctx context.Context, userID string) (*DatosTrabajadorDC3, error)
	GuardarDatosTrabajador(ctx context.Context, d *DatosTrabajadorDC3) error
	FindEmpresaInstructor(ctx context.Context, instructorID string) (*EmpresaInstructorDC3, error)
	GuardarEmpresaInstructor(ctx context.Context, e *EmpresaInstructorDC3) error
	FindConstancia(ctx context.Context, userID, capacitacionID string) (*ConstanciaDC3, error)
	RegistrarConstancia(ctx context.Context, userID, capacitacionID, archivoURL string) error
	ListConstancias(ctx context.Context, userID string) ([]*ConstanciaDC3, error)
	FechaInscripcion(ctx context.Context, userID, capacitacionID string) (time.Time, error)

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
	COALESCE(color,'#f97316') color, precio, COALESCE(precio_centavos, 0) precio_centavos, scheduled_at, duration, COALESCE(dc3_enabled, true) dc3_enabled, created_at,
	COALESCE(dc3_area_tematica,'') dc3_area_tematica,
	COALESCE(dc3_nombre_curso,'') dc3_nombre_curso,
	0 as total_lecciones,
	0 as lecciones_completadas
	FROM capacitaciones`

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
	query := `
		SELECT DISTINCT c.id, c.title, COALESCE(c.description,'') description, c.type,
		       COALESCE(c.file_path,'') file_path, COALESCE(c.content,'') content,
		       c.instructor_id, c.is_public, COALESCE(c.codigo_acceso,'') codigo_acceso,
		       COALESCE(c.welcome_message,'') welcome_message, COALESCE(c.thumbnail_url,'') thumbnail_url,
		       COALESCE(c.color,'#f97316') color, c.precio, COALESCE(c.precio_centavos, 0) precio_centavos, c.scheduled_at, c.duration, COALESCE(c.dc3_enabled, true) dc3_enabled, c.created_at,
		       0 as total_lecciones,
		       0 as lecciones_completadas
		FROM capacitaciones c
		LEFT JOIN asignaciones a ON a.capacitacion_id = c.id AND a.user_id = $1
		LEFT JOIN inscripciones i ON i.capacitacion_id = c.id AND i.user_id = $1
		WHERE (a.user_id = $1 OR i.user_id = $1) AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC`
	var cursos []*Curso
	return cursos, r.db.SelectContext(ctx, &cursos, query, userID)
}

func (r *postgresCursosRepository) ListByInstructor(ctx context.Context, instructorID string) ([]*Curso, error) {
	var cursos []*Curso
	return cursos, r.db.SelectContext(ctx, &cursos,
		selectCurso+` WHERE deleted_at IS NULL AND (instructor_id=$1 OR instructor_id IS NULL OR COALESCE(is_public, false)=true) ORDER BY created_at DESC`, instructorID)
}

func (r *postgresCursosRepository) FindByID(ctx context.Context, cursoID string) (*Curso, error) {
	c := &Curso{}
	return c, r.db.GetContext(ctx, c, selectCurso+` WHERE id=$1 AND deleted_at IS NULL`, cursoID)
}

func (r *postgresCursosRepository) FindByCodigo(ctx context.Context, codigo string) (*Curso, error) {
	c := &Curso{}
	cleanCode := strings.ToUpper(strings.TrimSpace(codigo))
	return c, r.db.GetContext(ctx, c, selectCurso+` WHERE UPPER(TRIM(codigo_acceso))=$1 AND deleted_at IS NULL`, cleanCode)
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
	codigoAcceso := strings.ToUpper(uuid.New().String()[:8])
	var id string
	err := r.db.QueryRowContext(ctx,
		// precio_centavos se deriva en SQL: ROUND sobre NUMERIC es exacto en
		// decimal, a diferencia de int64(precio*100) en Go, que truncaba.
		`INSERT INTO capacitaciones(title, description, type, file_path, content, instructor_id, is_public, welcome_message, thumbnail_url, color, precio, precio_centavos, duration, dc3_enabled, codigo_acceso, dc3_area_tematica, dc3_nombre_curso)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,ROUND($11::NUMERIC*100)::BIGINT,$12,$13,$14,$15,$16) RETURNING id`,
		req.Title, req.Description, req.Type, req.FilePath, req.Content, instructorID,
		req.IsPublic, req.WelcomeMessage, req.ThumbnailUrl, color, req.Precio, req.Duration, req.Dc3Enabled, codigoAcceso,
		req.Dc3AreaTematica, req.Dc3NombreCurso,
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
	// Los marcadores van numerados a mano, así que añadir una columna obliga a
	// recorrer los índices de scheduled_at y del WHERE. Es la razón por la que
	// dc3_area_tematica se coloca al final del bloque fijo, en $13.
	query := `UPDATE capacitaciones SET title=$1, description=$2, type=$3, file_path=$4, content=$5, is_public=$6, welcome_message=$7, thumbnail_url=$8, color=$9, precio=$10, precio_centavos=ROUND($10::NUMERIC*100)::BIGINT, duration=$11, dc3_enabled=$12, dc3_area_tematica=$13, dc3_nombre_curso=$14`
	args := []interface{}{
		req.Title, req.Description, req.Type, req.FilePath, req.Content,
		req.IsPublic, req.WelcomeMessage, req.ThumbnailUrl, color,
		req.Precio, req.Duration, req.Dc3Enabled, req.Dc3AreaTematica, req.Dc3NombreCurso,
	}
	if req.ScheduledAt != "" {
		if t, err := time.Parse(time.RFC3339, req.ScheduledAt); err == nil {
			query += `, scheduled_at=$15 WHERE id=$16`
			args = append(args, t, req.CursoId)
		} else {
			query += ` WHERE id=$15`
			args = append(args, req.CursoId)
		}
	} else {
		query += `, scheduled_at=NULL WHERE id=$15`
		args = append(args, req.CursoId)
	}
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, req.CursoId)
}

func (r *postgresCursosRepository) Delete(ctx context.Context, cursoID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE capacitaciones SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, cursoID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *postgresCursosRepository) DeleteByInstructor(ctx context.Context, instructorID, cursoID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE capacitaciones SET deleted_at=NOW() WHERE id=$1 AND instructor_id=$2 AND deleted_at IS NULL`, cursoID, instructorID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("not found or forbidden")
	}
	return nil
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
	newCode := strings.ToUpper(uuid.New().String()[:8])
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
		`INSERT INTO curso_licencias(capacitacion_id, nombre, precio, precio_centavos, capacidad_maxima, codigo_acceso)
		 VALUES($1,$2,$3,ROUND($3::NUMERIC*100)::BIGINT,$4,$5) RETURNING id`,
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
		 SELECT $1,$2,$3,$4,$5,$6, CAST($7 AS VARCHAR)
		 WHERE NOT EXISTS (
		    SELECT 1 FROM curso_licencias WHERE stripe_product_id = CAST($7 AS VARCHAR) AND $7 IS NOT NULL
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
		`UPDATE curso_licencias SET nombre=$1, precio=$2, precio_centavos=ROUND($2::NUMERIC*100)::BIGINT, capacidad_maxima=$3 WHERE id=$4`,
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
		`SELECT id, capacitacion_id, nombre, precio, COALESCE(precio_centavos, 0) precio_centavos, capacidad_maxima, usadas, codigo_acceso, stripe_product_id, stripe_price_id, comprador_id, created_at FROM curso_licencias WHERE capacitacion_id=$1 ORDER BY created_at DESC`, cursoID)
}

func (r *postgresCursosRepository) FindLicenciaByID(ctx context.Context, licenciaID string) (*Licencia, error) {
	l := &Licencia{}
	return l, r.db.GetContext(ctx, l, `SELECT id, capacitacion_id, nombre, precio, COALESCE(precio_centavos, 0) precio_centavos, capacidad_maxima, usadas, codigo_acceso, stripe_product_id, stripe_price_id, comprador_id, created_at FROM curso_licencias WHERE id=$1`, licenciaID)
}

func (r *postgresCursosRepository) FindLicenciaByCodigo(ctx context.Context, codigo string) (*Licencia, error) {
	l := &Licencia{}
	cleanCode := strings.ToUpper(strings.TrimSpace(codigo))
	return l, r.db.GetContext(ctx, l, `SELECT id, capacitacion_id, nombre, precio, COALESCE(precio_centavos, 0) precio_centavos, capacidad_maxima, usadas, codigo_acceso, stripe_product_id, stripe_price_id, comprador_id, created_at FROM curso_licencias WHERE UPPER(TRIM(codigo_acceso))=$1`, cleanCode)
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
		`SELECT l.id, l.capacitacion_id, l.nombre, l.precio, COALESCE(l.precio_centavos, 0) precio_centavos, l.capacidad_maxima, l.usadas,
		        l.codigo_acceso, l.stripe_product_id, l.stripe_price_id, l.comprador_id, l.created_at,
		        c.type AS curso_type, c.duration AS curso_duracion, c.title AS capacitacion_titulo,
		        (SELECT COUNT(*) FROM licencia_invitaciones i WHERE i.licencia_id = l.id)::int AS accesos_enviados
		 FROM curso_licencias l
		 LEFT JOIN capacitaciones c ON c.id = l.capacitacion_id
		 WHERE l.comprador_id=$1
		 ORDER BY l.created_at DESC`, userID)
}

func (r *postgresCursosRepository) AsignarCompradorLicencia(ctx context.Context, licenciaID, userID string) error {
	var stripeSessionID *string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-stripe-session-id"); len(vals) > 0 && vals[0] != "" {
			stripeSessionID = &vals[0]
		}
	}

	if stripeSessionID != nil {
		_, err := r.db.ExecContext(ctx, `UPDATE curso_licencias SET comprador_id=$1, stripe_product_id=CAST($2 AS VARCHAR) WHERE id=$3`, userID, stripeSessionID, licenciaID)
		return err
	}

	_, err := r.db.ExecContext(ctx, `UPDATE curso_licencias SET comprador_id=$1 WHERE id=$2`, userID, licenciaID)
	return err
}

// ── Reparto de accesos corporativos ───────────────────────────────────────────

func (r *postgresCursosRepository) ListInvitacionesLicencia(ctx context.Context, licenciaID string) ([]*InvitacionLicencia, error) {
	var invs []*InvitacionLicencia
	return invs, r.db.SelectContext(ctx, &invs,
		`SELECT i.id, i.licencia_id, i.nombre, i.email, i.codigo, i.estado, i.enviado_at
		   FROM licencia_invitaciones i
		  WHERE i.licencia_id = $1
		  ORDER BY i.enviado_at DESC`, licenciaID)
}

// ErrSinAccesosDisponibles indica que la licencia ya repartió todos sus lugares.
var ErrSinAccesosDisponibles = errors.New("no quedan accesos disponibles en esta licencia")

// Participante es la entrada mínima para repartir un acceso.
type Participante struct {
	Nombre string
	Email  string
}

// AsignarAccesos reparte accesos de forma atómica.
//
// Todos los participantes comparten el código de acceso de la licencia; lo que
// se controla aquí es el cupo. Va en una transacción porque el conteo de
// invitaciones y la inserción tienen que ver la misma foto: sin ella, dos
// envíos simultáneos podrían rebasar capacidad_maxima entre el SELECT y el
// INSERT. El SELECT ... FOR UPDATE sobre la licencia serializa esos envíos.
func (r *postgresCursosRepository) AsignarAccesos(ctx context.Context, licenciaID, codigoCompartido string, participantes []Participante) ([]*InvitacionLicencia, error) {
	if len(participantes) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op si ya se hizo Commit

	var capacidad int32
	if err := tx.GetContext(ctx, &capacidad,
		`SELECT capacidad_maxima FROM curso_licencias WHERE id = $1 FOR UPDATE`, licenciaID); err != nil {
		return nil, err
	}

	// Un correo que ya tiene invitación conserva su código: reenviar no consume
	// un lugar nuevo de la licencia.
	yaInvitados := map[string]*InvitacionLicencia{}
	var previas []*InvitacionLicencia
	if err := tx.SelectContext(ctx, &previas,
		`SELECT id, licencia_id, nombre, email, codigo, estado, enviado_at
		   FROM licencia_invitaciones WHERE licencia_id = $1`, licenciaID); err != nil {
		return nil, err
	}
	for _, p := range previas {
		yaInvitados[strings.ToLower(p.Email)] = p
	}

	nuevos := 0
	for _, p := range participantes {
		if _, ok := yaInvitados[strings.ToLower(p.Email)]; !ok {
			nuevos++
		}
	}

	// capacidad_maxima <= 0 significa licencia sin límite.
	if capacidad > 0 {
		disponibles := int(capacidad) - len(previas)
		if nuevos > disponibles {
			return nil, fmt.Errorf("%w: quedan %d de %d solicitados",
				ErrSinAccesosDisponibles, max(disponibles, 0), nuevos)
		}
	}

	resultado := make([]*InvitacionLicencia, 0, len(participantes))
	for _, p := range participantes {
		codigo := codigoCompartido
		if prev, ok := yaInvitados[strings.ToLower(p.Email)]; ok {
			codigo = prev.Codigo // reenvío: se respeta el código ya entregado
		}

		inv := &InvitacionLicencia{
			LicenciaID: licenciaID,
			Nombre:     p.Nombre,
			Email:      p.Email,
			Codigo:     codigo,
			Estado:     "enviado",
		}
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO licencia_invitaciones (licencia_id, nombre, email, codigo, estado)
			 VALUES ($1, $2, $3, $4, 'enviado')
			 ON CONFLICT (licencia_id, email) DO UPDATE
			    SET nombre     = EXCLUDED.nombre,
			        codigo     = EXCLUDED.codigo,
			        enviado_at = NOW()
			 RETURNING id, enviado_at`,
			inv.LicenciaID, inv.Nombre, inv.Email, inv.Codigo,
		).Scan(&inv.ID, &inv.EnviadoAt); err != nil {
			return nil, err
		}
		resultado = append(resultado, inv)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return resultado, nil
}

// ── Aviso DC-3 ────────────────────────────────────────────────────────────────

func (r *postgresCursosRepository) FindLicenciaDeInscripcion(ctx context.Context, userID, cursoID string) (*Licencia, error) {
	l := &Licencia{}
	err := r.db.GetContext(ctx, l,
		`SELECT l.id, l.capacitacion_id, l.nombre, l.precio, COALESCE(l.precio_centavos, 0) precio_centavos, l.capacidad_maxima, l.usadas,
		        l.codigo_acceso, l.stripe_product_id, l.stripe_price_id, l.comprador_id, l.created_at
		   FROM inscripciones i
		   JOIN curso_licencias l ON l.id = i.licencia_id
		  WHERE i.user_id = $1 AND i.capacitacion_id = $2`, userID, cursoID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // compra individual: no hay representante a quien avisar
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

// RegistrarAvisoDC3 se apoya en la clave primaria compuesta de dc3_avisos: el
// ON CONFLICT DO NOTHING hace la deduplicación atómica sin necesidad de leer
// primero, así que dos participantes que terminan a la vez no generan dos avisos.
func (r *postgresCursosRepository) RegistrarAvisoDC3(ctx context.Context, licenciaID, cursoID string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO dc3_avisos (licencia_id, capacitacion_id)
		 VALUES ($1, $2)
		 ON CONFLICT (licencia_id, capacitacion_id) DO NOTHING`, licenciaID, cursoID)
	if err != nil {
		return false, err
	}
	filas, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return filas > 0, nil
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
