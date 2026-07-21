package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	leccionespb "Prueba-Go/gen/lecciones"

	"github.com/jmoiron/sqlx"
)

// ─────────────────────────────────────────────────────────────────────────────
// Modelos internos
// ─────────────────────────────────────────────────────────────────────────────

type Modulo struct {
	ID             string    `db:"id"`
	CapacitacionID string    `db:"capacitacion_id"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	Orden          int32     `db:"orden"`
	CreatedAt      time.Time `db:"created_at"`
}

func (m *Modulo) ToProto() *leccionespb.ModuloResponse {
	return &leccionespb.ModuloResponse{
		Id:          m.ID,
		CursoId:     m.CapacitacionID,
		Title:       m.Title,
		Description: m.Description,
		Orden:       m.Orden,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type Submodulo struct {
	ID          string    `db:"id"`
	ModuloID    string    `db:"modulo_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Orden       int32     `db:"orden"`
	CreatedAt   time.Time `db:"created_at"`
}

func (s *Submodulo) ToProto() *leccionespb.SubmoduloResponse {
	return &leccionespb.SubmoduloResponse{
		Id:          s.ID,
		ModuloId:    s.ModuloID,
		Title:       s.Title,
		Description: s.Description,
		Orden:       s.Orden,
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type Leccion struct {
	ID             string    `db:"id"`
	CapacitacionID string    `db:"capacitacion_id"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	Type           string    `db:"type"` // almacena el nombre del enum, ej. "LESSON_TYPE_VIDEO"
	FilePath       string    `db:"file_path"`
	Content        string    `db:"content"`
	Orden          int32     `db:"orden"`
	DuracionMin    int32     `db:"duracion_min"`
	Completada     bool      `db:"completada"`
	CreatedAt      time.Time `db:"created_at"`
	// Jerarquía (nullable)
	ModuloID       *string `db:"modulo_id"`
	SubmoduloID    *string `db:"submodulo_id"`
	// Gamificación
	GameConfigJSON string `db:"game_config_json"`
	PointsReward   int32  `db:"points_reward"`
	SegundosVistos int32  `db:"segundos_vistos"`
	// Tareas/Actividades
	FechaInicio *time.Time `db:"fecha_inicio"`
	FechaCierre *time.Time `db:"fecha_cierre"`
	MiEntrega   *EntregaActividad
}

func parseLessonType(t string) leccionespb.LessonType {
	if val, ok := leccionespb.LessonType_value[t]; ok {
		return leccionespb.LessonType(val)
	}
	if n, err := strconv.Atoi(t); err == nil {
		return leccionespb.LessonType(n)
	}
	switch strings.ToLower(t) {
	case "video":
		return leccionespb.LessonType_LESSON_TYPE_VIDEO
	case "text", "texto":
		return leccionespb.LessonType_LESSON_TYPE_TEXT
	case "pdf":
		return leccionespb.LessonType_LESSON_TYPE_PDF
	case "quiz":
		return leccionespb.LessonType_LESSON_TYPE_QUIZ
	case "memory", "memorama", "5":
		return leccionespb.LessonType_LESSON_TYPE_GAME_MEMORY
	case "dragdrop", "clasificar", "6":
		return leccionespb.LessonType_LESSON_TYPE_GAME_DRAGDROP
	case "wordsearch", "sopa", "7":
		return leccionespb.LessonType_LESSON_TYPE_GAME_WORDSEARCH
	case "fillblank", "completar", "8":
		return leccionespb.LessonType_LESSON_TYPE_GAME_FILLBLANK
	case "order", "ordenar", "9":
		return leccionespb.LessonType_LESSON_TYPE_GAME_ORDER
	case "activity", "tarea", "actividad", "10":
		return leccionespb.LessonType_LESSON_TYPE_ACTIVITY
	}
	return leccionespb.LessonType_LESSON_TYPE_UNSPECIFIED
}

func (l *Leccion) ToProto() *leccionespb.LeccionResponse {
	resp := &leccionespb.LeccionResponse{
		Id:             l.ID,
		CursoId:        l.CapacitacionID,
		Title:          l.Title,
		Description:    l.Description,
		LessonType:     parseLessonType(l.Type),
		FilePath:       l.FilePath,
		Content:        l.Content,
		Orden:          l.Orden,
		DuracionMin:    l.DuracionMin,
		Completada:     l.Completada,
		CreatedAt:      l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		GameConfigJson: l.GameConfigJSON,
		PointsReward:   l.PointsReward,
		SegundosVistos: l.SegundosVistos,
	}
	if l.ModuloID != nil {
		resp.ModuloId = *l.ModuloID
	}
	if l.SubmoduloID != nil {
		resp.SubmoduloId = *l.SubmoduloID
	}
	if l.FechaInicio != nil {
		resp.FechaInicio = l.FechaInicio.UTC().Format(time.RFC3339)
	}
	if l.FechaCierre != nil {
		resp.FechaCierre = l.FechaCierre.UTC().Format(time.RFC3339)
	}
	if l.MiEntrega != nil {
		resp.MiEntrega = l.MiEntrega.ToProto()
	}
	return resp
}

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

// GameScore registra el resultado de un intento de minijuego.
type GameScore struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	LeccionID      string    `db:"leccion_id"`
	CapacitacionID string    `db:"capacitacion_id"`
	Points         int32     `db:"points"`
	TimeSecs       int32     `db:"time_secs"`
	ScoredAt       time.Time `db:"scored_at"`
}

// LeaderboardRow es el resultado de la query de leaderboard.
type LeaderboardRow struct {
	UserID    string `db:"user_id"`
	UserName  string `db:"user_name"`
	AvatarURL string `db:"avatar_url"`
	Points    int32  `db:"points"`
}

type EntregaActividad struct {
	ID             string    `db:"id"`
	LeccionID      string    `db:"leccion_id"`
	CapacitacionID string    `db:"capacitacion_id"`
	UserID         string    `db:"user_id"`
	UserName       string    `db:"user_name"`
	UserEmail      string    `db:"user_email"`
	AvatarURL      string    `db:"avatar_url"`
	CursoTitle     string    `db:"curso_title"`
	LeccionTitle   string    `db:"leccion_title"`
	FilePath       string    `db:"file_path"`
	FileName       string    `db:"file_name"`
	FileSize       int64     `db:"file_size"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func (e *EntregaActividad) ToProto() *leccionespb.EntregaResponse {
	return &leccionespb.EntregaResponse{
		Id:             e.ID,
		LeccionId:      e.LeccionID,
		CapacitacionId: e.CapacitacionID,
		UserId:         e.UserID,
		UserName:       e.UserName,
		UserEmail:      e.UserEmail,
		AvatarUrl:      e.AvatarURL,
		CursoTitle:     e.CursoTitle,
		LeccionTitle:   e.LeccionTitle,
		FilePath:       e.FilePath,
		FileName:       e.FileName,
		FileSize:       e.FileSize,
		CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      e.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Interface del repositorio
// ─────────────────────────────────────────────────────────────────────────────

type LeccionesRepository interface {
	// ── Módulos ───────────────────────────────────────────────────────────────
	ListModulos(ctx context.Context, cursoID string) ([]*Modulo, error)
	FindModuloByID(ctx context.Context, moduloID string) (*Modulo, error)
	CreateModulo(ctx context.Context, req *leccionespb.CreateModuloRequest) (*Modulo, error)
	UpdateModulo(ctx context.Context, req *leccionespb.UpdateModuloRequest) (*Modulo, error)
	DeleteModulo(ctx context.Context, moduloID string) error
	ReorderModulos(ctx context.Context, cursoID string, ids []string) error

	// ── Submódulos ────────────────────────────────────────────────────────────
	ListSubmodulos(ctx context.Context, moduloID string) ([]*Submodulo, error)
	FindSubmoduloByID(ctx context.Context, submoduloID string) (*Submodulo, error)
	CreateSubmodulo(ctx context.Context, req *leccionespb.CreateSubmoduloRequest) (*Submodulo, error)
	UpdateSubmodulo(ctx context.Context, req *leccionespb.UpdateSubmoduloRequest) (*Submodulo, error)
	DeleteSubmodulo(ctx context.Context, submoduloID string) error
	ReorderSubmodulos(ctx context.Context, moduloID string, ids []string) error

	// ── Lecciones ─────────────────────────────────────────────────────────────
	ListByCurso(ctx context.Context, cursoID string) ([]*Leccion, error)
	ListByCursoConProgreso(ctx context.Context, cursoID, userID string) ([]*Leccion, error)
	ListByModulo(ctx context.Context, moduloID string) ([]*Leccion, error)
	ListBySubmodulo(ctx context.Context, submoduloID string) ([]*Leccion, error)
	FindByID(ctx context.Context, leccionID string) (*Leccion, error)
	Create(ctx context.Context, req *leccionespb.CreateLeccionRequest) (*Leccion, error)
	Update(ctx context.Context, req *leccionespb.UpdateLeccionRequest) (*Leccion, error)
	Delete(ctx context.Context, leccionID string) error
	Reorder(ctx context.Context, cursoID string, ids []string) error
	MarcarCompleta(ctx context.Context, leccionID, userID string) error
	GuardarProgresoVideo(ctx context.Context, leccionID, userID string, segundos int32) error

	// ── Preguntas intermedias ─────────────────────────────────────────────────
	ListPreguntas(ctx context.Context, cursoID string) ([]*PreguntaIntermedia, error)
	GetOpciones(ctx context.Context, preguntaID string) ([]*OpcionIntermedia, error)
	CreatePregunta(ctx context.Context, req *leccionespb.CreateIntermediaRequest) (*PreguntaIntermedia, error)
	DeletePregunta(ctx context.Context, preguntaID string) error
	SubmitRespuestas(ctx context.Context, cursoID, userID string, respuestas []*leccionespb.Respuesta) (correctas, total int32, err error)

	// ── Árbol del curso ───────────────────────────────────────────────────────
	// BuildCursoTree ensambla el árbol Módulo → Submódulo → Lección.
	// Si userID es vacío, completada siempre será false.
	BuildCursoTree(ctx context.Context, cursoID, userID string) (*leccionespb.CursoTreeResponse, error)

	// ── Gamificación ──────────────────────────────────────────────────────────
	// Registra puntos de un intento y devuelve el total acumulado del usuario en el curso.
	InsertGameScore(ctx context.Context, score *GameScore) (totalPoints int32, err error)
	// Devuelve si el usuario ya había completado la lección (para evitar puntos dobles en MarcarCompleta).
	IsLeccionCompletada(ctx context.Context, leccionID, userID string) (bool, error)
	// Leaderboard top-N para un curso.
	GetLeaderboard(ctx context.Context, cursoID string, topN int) ([]*LeaderboardRow, error)
	// Puntos totales del usuario en un curso.
	GetUserCoursePoints(ctx context.Context, userID, cursoID string) (int32, error)
	// Suma de puntos en todos los cursos (para el perfil).
	GetUserTotalPoints(ctx context.Context, userID string) (int32, error)
	// Actualiza la columna desnormalizada users.points_total.
	UpdateUserTotalPoints(ctx context.Context, userID string, total int32) error
	// Intenta otorgar una insignia. Devuelve (true, nil) si se insertó (era nueva).
	// Devuelve (false, nil) si el usuario ya la tenía (ON CONFLICT DO NOTHING).
	TryAwardBadge(ctx context.Context, userID, badgeSlug string) (bool, error)
	// Devuelve todos los slugs de insignias desbloqueadas por un usuario.
	GetUserBadgeSlugs(ctx context.Context, userID string) ([]string, error)

	// ── Entregas de Tareas / Actividades ──────────────────────────────────────
	SubmitEntrega(ctx context.Context, req *leccionespb.SubmitEntregaRequest) (*EntregaActividad, error)
	GetEntregaUsuario(ctx context.Context, leccionID, userID string) (*EntregaActividad, error)
	GetEntregasPorUsuarioCurso(ctx context.Context, cursoID, userID string) (map[string]*EntregaActividad, error)
	InstructorListEntregas(ctx context.Context, cursoID, leccionID string) ([]*EntregaActividad, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementación PostgreSQL
// ─────────────────────────────────────────────────────────────────────────────

// PostgresLeccionesRepository es el tipo concreto exportado.
// Se exporta para que el service pueda usarlo en type assertions si fuera
// necesario; en la práctica toda la interacción ocurre vía la interfaz.
type PostgresLeccionesRepository struct{ db *sqlx.DB }

// Alias interno para legibilidad en este archivo.
type postgresLeccionesRepository = PostgresLeccionesRepository

func NewLeccionesRepository(db *sqlx.DB) LeccionesRepository {
	return &postgresLeccionesRepository{db: db}
}

// ── Constantes de SELECT ──────────────────────────────────────────────────────

const selectModulo = `
	SELECT id, capacitacion_id, title,
	       COALESCE(description,'') description,
	       orden, created_at
	FROM modulos`

const selectSubmodulo = `
	SELECT id, modulo_id, title,
	       COALESCE(description,'') description,
	       orden, created_at
	FROM submodulos`

// selectLeccion base — la columna completada se inyecta en cada query específica.
const selectLeccionCols = `
	l.id, l.capacitacion_id, l.title,
	COALESCE(l.description,'')     AS description,
	l.type,
	COALESCE(l.file_path,'')       AS file_path,
	COALESCE(l.content,'')         AS content,
	l.orden,
	COALESCE(l.duracion_min,0)     AS duracion_min,
	l.created_at,
	l.modulo_id, l.submodulo_id,
	COALESCE(l.game_config_json,'') AS game_config_json,
	COALESCE(l.points_reward,0)    AS points_reward,
	l.fecha_inicio, l.fecha_cierre`

// ── Módulos ───────────────────────────────────────────────────────────────────

func (r *postgresLeccionesRepository) ListModulos(ctx context.Context, cursoID string) ([]*Modulo, error) {
	var ms []*Modulo
	return ms, r.db.SelectContext(ctx, &ms,
		selectModulo+` WHERE capacitacion_id=$1 AND deleted_at IS NULL ORDER BY orden`, cursoID)
}

func (r *postgresLeccionesRepository) FindModuloByID(ctx context.Context, moduloID string) (*Modulo, error) {
	m := &Modulo{}
	return m, r.db.GetContext(ctx, m,
		selectModulo+` WHERE id=$1 AND deleted_at IS NULL`, moduloID)
}

func (r *postgresLeccionesRepository) CreateModulo(ctx context.Context, req *leccionespb.CreateModuloRequest) (*Modulo, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO modulos(capacitacion_id, title, description, orden)
		 VALUES($1,$2,$3,$4) RETURNING id`,
		req.CursoId, req.Title, req.Description, req.Orden,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindModuloByID(ctx, id)
}

func (r *postgresLeccionesRepository) UpdateModulo(ctx context.Context, req *leccionespb.UpdateModuloRequest) (*Modulo, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE modulos SET title=$1, description=$2, orden=$3 WHERE id=$4`,
		req.Title, req.Description, req.Orden, req.ModuloId)
	if err != nil {
		return nil, err
	}
	return r.FindModuloByID(ctx, req.ModuloId)
}

func (r *postgresLeccionesRepository) DeleteModulo(ctx context.Context, moduloID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE modulos SET deleted_at=NOW() WHERE id=$1`, moduloID)
	return err
}

func (r *postgresLeccionesRepository) ReorderModulos(ctx context.Context, cursoID string, ids []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE modulos SET orden=$1 WHERE id=$2 AND capacitacion_id=$3`,
			i+1, id, cursoID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ── Submódulos ────────────────────────────────────────────────────────────────

func (r *postgresLeccionesRepository) ListSubmodulos(ctx context.Context, moduloID string) ([]*Submodulo, error) {
	var ss []*Submodulo
	return ss, r.db.SelectContext(ctx, &ss,
		selectSubmodulo+` WHERE modulo_id=$1 AND deleted_at IS NULL ORDER BY orden`, moduloID)
}

func (r *postgresLeccionesRepository) FindSubmoduloByID(ctx context.Context, submoduloID string) (*Submodulo, error) {
	s := &Submodulo{}
	return s, r.db.GetContext(ctx, s,
		selectSubmodulo+` WHERE id=$1 AND deleted_at IS NULL`, submoduloID)
}

func (r *postgresLeccionesRepository) CreateSubmodulo(ctx context.Context, req *leccionespb.CreateSubmoduloRequest) (*Submodulo, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO submodulos(modulo_id, title, description, orden)
		 VALUES($1,$2,$3,$4) RETURNING id`,
		req.ModuloId, req.Title, req.Description, req.Orden,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindSubmoduloByID(ctx, id)
}

func (r *postgresLeccionesRepository) UpdateSubmodulo(ctx context.Context, req *leccionespb.UpdateSubmoduloRequest) (*Submodulo, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE submodulos SET title=$1, description=$2, orden=$3 WHERE id=$4`,
		req.Title, req.Description, req.Orden, req.SubmoduloId)
	if err != nil {
		return nil, err
	}
	return r.FindSubmoduloByID(ctx, req.SubmoduloId)
}

func (r *postgresLeccionesRepository) DeleteSubmodulo(ctx context.Context, submoduloID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE submodulos SET deleted_at=NOW() WHERE id=$1`, submoduloID)
	return err
}

func (r *postgresLeccionesRepository) ReorderSubmodulos(ctx context.Context, moduloID string, ids []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE submodulos SET orden=$1 WHERE id=$2 AND modulo_id=$3`,
			i+1, id, moduloID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ── Lecciones ─────────────────────────────────────────────────────────────────

func (r *postgresLeccionesRepository) ListByCurso(ctx context.Context, cursoID string) ([]*Leccion, error) {
	query := `SELECT ` + selectLeccionCols + `, false AS completada, 0 AS segundos_vistos
	          FROM lecciones l
	          WHERE l.capacitacion_id=$1 AND l.deleted_at IS NULL
	          ORDER BY l.orden`
	var lecs []*Leccion
	return lecs, r.db.SelectContext(ctx, &lecs, query, cursoID)
}

func (r *postgresLeccionesRepository) ListByCursoConProgreso(ctx context.Context, cursoID, userID string) ([]*Leccion, error) {
	query := `SELECT ` + selectLeccionCols + `,
	          CASE WHEN p.id IS NOT NULL AND p.completado = true THEN true ELSE false END AS completada,
	          COALESCE(p.segundos_vistos, 0) AS segundos_vistos
	          FROM lecciones l
	          LEFT JOIN progreso_lecciones p ON p.leccion_id = l.id AND p.user_id = $2
	          WHERE l.capacitacion_id = $1 AND l.deleted_at IS NULL
	          ORDER BY l.orden`
	var lecs []*Leccion
	if err := r.db.SelectContext(ctx, &lecs, query, cursoID, userID); err != nil {
		return nil, err
	}
	entregasMap, _ := r.GetEntregasPorUsuarioCurso(ctx, cursoID, userID)
	if entregasMap != nil {
		for _, l := range lecs {
			if ent, ok := entregasMap[l.ID]; ok {
				l.MiEntrega = ent
			}
		}
	}
	return lecs, nil
}

func (r *postgresLeccionesRepository) ListByModulo(ctx context.Context, moduloID string) ([]*Leccion, error) {
	query := `SELECT ` + selectLeccionCols + `, false AS completada, 0 AS segundos_vistos
	          FROM lecciones l
	          WHERE l.modulo_id=$1 AND l.submodulo_id IS NULL AND l.deleted_at IS NULL
	          ORDER BY l.orden`
	var lecs []*Leccion
	return lecs, r.db.SelectContext(ctx, &lecs, query, moduloID)
}

func (r *postgresLeccionesRepository) ListBySubmodulo(ctx context.Context, submoduloID string) ([]*Leccion, error) {
	query := `SELECT ` + selectLeccionCols + `, false AS completada, 0 AS segundos_vistos
	          FROM lecciones l
	          WHERE l.submodulo_id=$1 AND l.deleted_at IS NULL
	          ORDER BY l.orden`
	var lecs []*Leccion
	return lecs, r.db.SelectContext(ctx, &lecs, query, submoduloID)
}

func (r *postgresLeccionesRepository) FindByID(ctx context.Context, leccionID string) (*Leccion, error) {
	query := `SELECT ` + selectLeccionCols + `, false AS completada, 0 AS segundos_vistos
	          FROM lecciones l
	          WHERE l.id=$1 AND l.deleted_at IS NULL`
	l := &Leccion{}
	return l, r.db.GetContext(ctx, l, query, leccionID)
}

func parseOptionalDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty date string")
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid date format")
}

func (r *postgresLeccionesRepository) Create(ctx context.Context, req *leccionespb.CreateLeccionRequest) (*Leccion, error) {
	// Convertir enum a string para almacenar en BD
	lessonTypeStr := req.LessonType.String()
	if lessonTypeStr == "" {
		lessonTypeStr = "LESSON_TYPE_VIDEO"
	}

	var moduloID, submoduloID *string
	if req.ModuloId != "" {
		moduloID = &req.ModuloId
	}
	if req.SubmoduloId != "" {
		submoduloID = &req.SubmoduloId
	}

	var fechaInicio, fechaCierre *time.Time
	if req.FechaInicio != "" {
		if t, err := parseOptionalDate(req.FechaInicio); err == nil {
			fechaInicio = &t
		}
	}
	if req.FechaCierre != "" {
		if t, err := parseOptionalDate(req.FechaCierre); err == nil {
			fechaCierre = &t
		}
	}

	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO lecciones(
			capacitacion_id, title, description, type, file_path, content,
			orden, duracion_min, modulo_id, submodulo_id, game_config_json, points_reward,
			fecha_inicio, fecha_cierre
		) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`,
		req.CursoId, req.Title, req.Description, lessonTypeStr,
		req.FilePath, req.Content, req.Orden, req.DuracionMin,
		moduloID, submoduloID, req.GameConfigJson, req.PointsReward,
		fechaInicio, fechaCierre,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *postgresLeccionesRepository) Update(ctx context.Context, req *leccionespb.UpdateLeccionRequest) (*Leccion, error) {
	lessonTypeStr := req.LessonType.String()
	if lessonTypeStr == "" {
		lessonTypeStr = "LESSON_TYPE_VIDEO"
	}

	var moduloID, submoduloID *string
	if req.ModuloId != "" {
		moduloID = &req.ModuloId
	}
	if req.SubmoduloId != "" {
		submoduloID = &req.SubmoduloId
	}

	var fechaInicio, fechaCierre *time.Time
	if req.FechaInicio != "" {
		if t, err := parseOptionalDate(req.FechaInicio); err == nil {
			fechaInicio = &t
		}
	}
	if req.FechaCierre != "" {
		if t, err := parseOptionalDate(req.FechaCierre); err == nil {
			fechaCierre = &t
		}
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE lecciones SET
			title=$1, description=$2, type=$3, file_path=$4, content=$5,
			orden=$6, duracion_min=$7, modulo_id=$8, submodulo_id=$9,
			game_config_json=$10, points_reward=$11, fecha_inicio=$12, fecha_cierre=$13
		WHERE id=$14`,
		req.Title, req.Description, lessonTypeStr, req.FilePath, req.Content,
		req.Orden, req.DuracionMin, moduloID, submoduloID,
		req.GameConfigJson, req.PointsReward, fechaInicio, fechaCierre, req.LeccionId)
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
		`INSERT INTO progreso_lecciones(leccion_id, user_id, completado, completado_at)
		 VALUES($1, $2, true, NOW())
		 ON CONFLICT (user_id, leccion_id)
		 DO UPDATE SET completado = true, completado_at = NOW()`,
		leccionID, userID)
	return err
}

func (r *postgresLeccionesRepository) GuardarProgresoVideo(ctx context.Context, leccionID, userID string, segundos int32) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO progreso_lecciones(leccion_id, user_id, segundos_vistos, completado)
		 VALUES($1, $2, $3, false)
		 ON CONFLICT (user_id, leccion_id)
		 DO UPDATE SET segundos_vistos = GREATEST(progreso_lecciones.segundos_vistos, EXCLUDED.segundos_vistos)`,
		leccionID, userID, segundos)
	return err
}

func (r *postgresLeccionesRepository) IsLeccionCompletada(ctx context.Context, leccionID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM progreso_lecciones WHERE leccion_id=$1 AND user_id=$2 AND completado=true`,
		leccionID, userID).Scan(&count)
	return count > 0, err
}

// ── Preguntas intermedias ─────────────────────────────────────────────────────

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
	for _, o := range req.Opciones {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO opciones_intermedias(pregunta_id,texto,es_correcta) VALUES($1,$2,$3)`,
			id, o.Texto, o.EsCorrecta)
		if err != nil {
			return nil, err
		}
	}
	pq := &PreguntaIntermedia{
		ID: id, CapacitacionID: req.CursoId, DespuesDeLeccionID: dlid,
		Texto: req.Texto, Tipo: req.Tipo, Orden: req.Orden,
	}
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

// ── Gamificación ──────────────────────────────────────────────────────────────

// InsertGameScore persiste el intento y devuelve el total acumulado del usuario en el curso.
// No hay límite de intentos: se suman todos los puntos (puedes cambiar a MAX si prefieres).
func (r *postgresLeccionesRepository) InsertGameScore(ctx context.Context, score *GameScore) (int32, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO game_scores(user_id, leccion_id, capacitacion_id, points, time_secs)
		 VALUES($1,$2,$3,$4,$5)`,
		score.UserID, score.LeccionID, score.CapacitacionID, score.Points, score.TimeSecs)
	if err != nil {
		return 0, err
	}
	return r.GetUserCoursePoints(ctx, score.UserID, score.CapacitacionID)
}

func (r *postgresLeccionesRepository) GetUserCoursePoints(ctx context.Context, userID, cursoID string) (int32, error) {
	var total int32
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(max_points),0) FROM (
             SELECT MAX(points) as max_points
             FROM game_scores
             WHERE user_id=$1 AND capacitacion_id=$2
             GROUP BY leccion_id
         ) sub`,
		userID, cursoID).Scan(&total)
	return total, err
}

func (r *postgresLeccionesRepository) GetUserTotalPoints(ctx context.Context, userID string) (int32, error) {
	var total int32
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(max_points),0) FROM (
             SELECT MAX(points) as max_points
             FROM game_scores
             WHERE user_id=$1
             GROUP BY leccion_id
         ) sub`,
		userID).Scan(&total)
	return total, err
}

func (r *postgresLeccionesRepository) UpdateUserTotalPoints(ctx context.Context, userID string, total int32) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_points (user_id, points_total, updated_at)
		 VALUES ($2, $1, NOW())
		 ON CONFLICT (user_id) DO UPDATE
		 SET points_total = EXCLUDED.points_total, updated_at = NOW()`, total, userID)
	return err
}

// GetLeaderboard devuelve los top-N usuarios por puntos en el curso (juegos + lecciones completadas).
func (r *postgresLeccionesRepository) GetLeaderboard(ctx context.Context, cursoID string, topN int) ([]*LeaderboardRow, error) {
	if topN <= 0 {
		topN = 5
	}
	query := `
		WITH combined_points AS (
			SELECT user_id, MAX(points) AS pts, MIN(scored_at) AS scored_at
			FROM game_scores
			WHERE capacitacion_id = $1
			GROUP BY user_id, leccion_id
			UNION ALL
			SELECT pl.user_id, COALESCE(NULLIF(l.points_reward, 0), 15) AS pts, NOW() AS scored_at
			FROM progreso_lecciones pl
			JOIN lecciones l ON l.id = pl.leccion_id
			WHERE l.capacitacion_id = $1 AND pl.completado = true
			GROUP BY pl.user_id, pl.leccion_id, l.points_reward
		)
		SELECT
			c.user_id,
			'' AS user_name,
			'' AS avatar_url,
			SUM(c.pts)::INT AS points
		FROM combined_points c
		GROUP BY c.user_id
		ORDER BY points DESC, MIN(c.scored_at) ASC
		LIMIT $2`
	var rows []*LeaderboardRow
	return rows, r.db.SelectContext(ctx, &rows, query, cursoID, topN)
}

// ── Insignias ─────────────────────────────────────────────────────────────────

// TryAwardBadge intenta insertar una nueva insignia para el usuario.
// Usa ON CONFLICT DO NOTHING para ser idempotente.
// Devuelve (true, nil) si se insertó, (false, nil) si ya existía.
func (r *postgresLeccionesRepository) TryAwardBadge(ctx context.Context, userID, badgeSlug string) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO user_badges(user_id, badge_slug)
		 VALUES($1, $2) ON CONFLICT(user_id, badge_slug) DO NOTHING`,
		userID, badgeSlug)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// GetUserBadgeSlugs devuelve los slugs de todas las insignias del usuario.
func (r *postgresLeccionesRepository) GetUserBadgeSlugs(ctx context.Context, userID string) ([]string, error) {
	var slugs []string
	err := r.db.SelectContext(ctx, &slugs,
		`SELECT badge_slug FROM user_badges WHERE user_id=$1 ORDER BY unlocked_at DESC`,
		userID)
	return slugs, err
}

// ── Entregas de Tareas / Actividades ──────────────────────────────────────────

const selectEntregaCols = `
	e.id, e.leccion_id, e.capacitacion_id, e.user_id,
	'' AS user_name,
	'' AS user_email,
	'' AS avatar_url,
	'' AS curso_title,
	COALESCE(l.title, '') AS leccion_title,
	e.file_path, e.file_name,
	COALESCE(e.file_size, 0) AS file_size,
	e.created_at, e.updated_at`

const fromEntregaJoin = `
	FROM entregas_actividad e
	LEFT JOIN lecciones l ON l.id = e.leccion_id`

func (r *postgresLeccionesRepository) SubmitEntrega(ctx context.Context, req *leccionespb.SubmitEntregaRequest) (*EntregaActividad, error) {
	var id string
	query := `
		INSERT INTO entregas_actividad (
			leccion_id, capacitacion_id, user_id, file_path, file_name, file_size, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (leccion_id, user_id) DO UPDATE SET
			file_path  = EXCLUDED.file_path,
			file_name  = EXCLUDED.file_name,
			file_size  = EXCLUDED.file_size,
			updated_at = NOW()
		RETURNING id`
	err := r.db.QueryRowContext(ctx, query,
		req.LeccionId, req.CapacitacionId, req.UserId, req.FilePath, req.FileName, req.FileSize,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetEntregaUsuario(ctx, req.LeccionId, req.UserId)
}

func (r *postgresLeccionesRepository) GetEntregaUsuario(ctx context.Context, leccionID, userID string) (*EntregaActividad, error) {
	query := `SELECT ` + selectEntregaCols + fromEntregaJoin + ` WHERE e.leccion_id=$1 AND e.user_id=$2`
	e := &EntregaActividad{}
	return e, r.db.GetContext(ctx, e, query, leccionID, userID)
}

func (r *postgresLeccionesRepository) GetEntregasPorUsuarioCurso(ctx context.Context, cursoID, userID string) (map[string]*EntregaActividad, error) {
	query := `SELECT ` + selectEntregaCols + fromEntregaJoin + ` WHERE e.capacitacion_id=$1 AND e.user_id=$2`
	var list []*EntregaActividad
	if err := r.db.SelectContext(ctx, &list, query, cursoID, userID); err != nil {
		return nil, err
	}
	m := make(map[string]*EntregaActividad, len(list))
	for _, e := range list {
		m[e.LeccionID] = e
	}
	return m, nil
}

func (r *postgresLeccionesRepository) InstructorListEntregas(ctx context.Context, cursoID, leccionID string) ([]*EntregaActividad, error) {
	var list []*EntregaActividad
	var query string
	var args []interface{}
	if leccionID != "" {
		query = `SELECT ` + selectEntregaCols + fromEntregaJoin + ` WHERE e.leccion_id=$1 ORDER BY e.updated_at DESC`
		args = append(args, leccionID)
	} else if cursoID != "" {
		query = `SELECT ` + selectEntregaCols + fromEntregaJoin + ` WHERE e.capacitacion_id=$1 ORDER BY e.updated_at DESC`
		args = append(args, cursoID)
	} else {
		query = `SELECT ` + selectEntregaCols + fromEntregaJoin + ` ORDER BY e.updated_at DESC`
	}
	return list, r.db.SelectContext(ctx, &list, query, args...)
}

