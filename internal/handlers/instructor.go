package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"path/filepath"
	"strings"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// generateCode devuelve un código alfanumérico en mayúsculas de n caracteres.
func generateCode(n int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}

// uniqueCode genera un código que no exista ya en la tabla.
func uniqueCode() string {
	for {
		code := generateCode(6)
		var existing string
		err := db.DB.QueryRow(`SELECT id FROM capacitaciones WHERE codigo_acceso=$1`, code).Scan(&existing)
		if err != nil {
			// no existe → válido
			return code
		}
	}
}

// ── Capacitaciones del instructor ─────────────────────────────────────────────

func InstructorListCapacitaciones(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT id, title, description, type,
		       COALESCE(file_path,''), COALESCE(content,''),
		       instructor_id, is_public, COALESCE(codigo_acceso,''),
		       COALESCE(welcome_message,''), COALESCE(thumbnail_url,''), created_at
		FROM capacitaciones
		WHERE instructor_id = $1
		ORDER BY created_at DESC
	`, instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Capacitacion{}
	for rows.Next() {
		var cap models.Capacitacion
		rows.Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type,
			&cap.FilePath, &cap.Content, &cap.InstructorID, &cap.IsPublic, &cap.CodigoAcceso,
			&cap.WelcomeMessage, &cap.ThumbnailURL, &cap.CreatedAt)
		result = append(result, cap)
	}
	c.JSON(http.StatusOK, result)
}

func InstructorCreateCapacitacion(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	title := c.PostForm("title")
	description := c.PostForm("description")
	capType := c.PostForm("type")
	content := c.PostForm("content")
	isPublic := c.PostForm("is_public") == "true"
	welcomeMessage := c.PostForm("welcome_message")

	if title == "" || capType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title y type son requeridos"})
		return
	}

	var filePath string
	file, err := c.FormFile("file")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		newName := uuid.NewString() + ext
		var dest string
		if capType == "video" {
			dest = filepath.Join("uploads", "videos", newName)
		} else {
			dest = filepath.Join("uploads", "documents", newName)
		}
		if err := c.SaveUploadedFile(file, dest); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error guardando archivo"})
			return
		}
		filePath = "/" + filepath.ToSlash(dest)
	}

	var thumbnailPath string
	thumbFile, err := c.FormFile("thumbnail")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(thumbFile.Filename))
		newName := uuid.NewString() + ext
		dest := filepath.Join("uploads", "thumbnails", newName)
		if err := c.SaveUploadedFile(thumbFile, dest); err == nil {
			thumbnailPath = "/" + filepath.ToSlash(dest)
		}
	}

	var id string
	err = db.DB.QueryRow(
		`INSERT INTO capacitaciones(title, description, type, file_path, content, instructor_id, is_public, codigo_acceso, welcome_message, thumbnail_url)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		title, description, capType, filePath, content, instructorID, isPublic, uniqueCode(), welcomeMessage, thumbnailPath,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Devolver id y código
	var codigo string
	db.DB.QueryRow(`SELECT COALESCE(codigo_acceso,'') FROM capacitaciones WHERE id=$1`, id).Scan(&codigo)
	c.JSON(http.StatusCreated, gin.H{"id": id, "codigo_acceso": codigo})
}

func InstructorDeleteCapacitacion(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")
	res, err := db.DB.Exec(`DELETE FROM capacitaciones WHERE id=$1 AND instructor_id=$2`, id, instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado o no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func InstructorTogglePublic(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")

	var current bool
	err := db.DB.QueryRow(`SELECT is_public FROM capacitaciones WHERE id=$1 AND instructor_id=$2`, id, instructorID).Scan(&current)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado o sin permisos"})
		return
	}
	_, err = db.DB.Exec(`UPDATE capacitaciones SET is_public=$1 WHERE id=$2`, !current, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_public": !current})
}

// ── Exámenes del instructor ────────────────────────────────────────────────────

func InstructorListExamenes(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT id, title, description, created_at
		FROM examenes WHERE instructor_id=$1 ORDER BY created_at DESC
	`, instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Examen{}
	for rows.Next() {
		var e models.Examen
		rows.Scan(&e.ID, &e.Title, &e.Description, &e.CreatedAt)
		// Cargar capacitacion enlazada
		db.DB.QueryRow(`SELECT e.capacitacion_id, COALESCE(c.title,'') FROM examenes e LEFT JOIN capacitaciones c ON c.id=e.capacitacion_id WHERE e.id=$1`, e.ID).
			Scan(&e.CapacitacionID, &e.CapacitacionNombre)
		result = append(result, e)
	}
	c.JSON(http.StatusOK, result)
}

func InstructorCreateExamen(c *gin.Context) {
	instructorID, _ := c.Get("user_id")

	var req createExamenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var examenID string
	err = tx.QueryRow(
		`INSERT INTO examenes(title, description, instructor_id, capacitacion_id) VALUES($1,$2,$3,$4) RETURNING id`,
		req.Title, req.Description, instructorID, req.CapacitacionID,
	).Scan(&examenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, p := range req.Preguntas {
		tipo := p.Tipo
		if tipo == "" {
			tipo = "multiple_choice"
		}
		var preguntaID string
		err = tx.QueryRow(
			`INSERT INTO preguntas(examen_id, texto, tipo, valor, orden) VALUES($1,$2,$3,$4,$5) RETURNING id`,
			examenID, p.Texto, tipo, p.Valor, p.Orden,
		).Scan(&preguntaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if tipo != "open_text" {
			for _, o := range p.Opciones {
				_, err = tx.Exec(
					`INSERT INTO opciones(pregunta_id, texto, es_correcta) VALUES($1,$2,$3)`,
					preguntaID, o.Texto, o.EsCorrecta,
				)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": examenID})
}

func InstructorDeleteExamen(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")
	res, err := db.DB.Exec(`DELETE FROM examenes WHERE id=$1 AND instructor_id=$2`, id, instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado o no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Estudiantes del instructor ─────────────────────────────────────────────────

func InstructorListEstudiantes(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	// Usuarios inscritos en capacitaciones propias del instructor
	rows, err := db.DB.Query(`
		SELECT DISTINCT u.id, u.name, u.email, u.role, u.created_at
		FROM users u
		INNER JOIN inscripciones i ON i.user_id = u.id
		INNER JOIN capacitaciones c ON c.id = i.capacitacion_id
		WHERE c.instructor_id = $1
		ORDER BY u.name
	`, instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.User{}
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt)
		result = append(result, u)
	}
	c.JSON(http.StatusOK, result)
}

// InstructorAsignar asigna un usuario a un curso/examen del instructor
func InstructorAsignar(c *gin.Context) {
	instructorID, _ := c.Get("user_id")

	var req asignarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CapacitacionID == nil && req.ExamenID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere capacitacion_id o examen_id"})
		return
	}

	// Verificar que la capacitacion/examen pertenece al instructor
	if req.CapacitacionID != nil {
		var owner string
		err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, *req.CapacitacionID).Scan(&owner)
		if err != nil || owner != instructorID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado para esta capacitación"})
			return
		}
	}
	if req.ExamenID != nil {
		var owner string
		err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM examenes WHERE id=$1`, *req.ExamenID).Scan(&owner)
		if err != nil || owner != instructorID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado para este examen"})
			return
		}
	}

	var id string
	err := db.DB.QueryRow(
		`INSERT INTO asignaciones(user_id, capacitacion_id, examen_id) VALUES($1,$2,$3)
		 ON CONFLICT DO NOTHING RETURNING id`,
		req.UserID, req.CapacitacionID, req.ExamenID,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ── Cursos públicos (cualquier usuario autenticado) ───────────────────────────

func ListCursosPublicos(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT c.id, c.title, c.description, c.type,
		       COALESCE(c.file_path,''), COALESCE(c.content,''),
		       c.instructor_id, c.is_public, c.created_at,
		       EXISTS(SELECT 1 FROM inscripciones i WHERE i.capacitacion_id=c.id AND i.user_id=$1) as inscrito
		FROM capacitaciones c
		WHERE c.is_public = true
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type CursoPublico struct {
		models.Capacitacion
		Inscrito bool `json:"inscrito"`
	}

	result := []CursoPublico{}
	for rows.Next() {
		var cp CursoPublico
		rows.Scan(&cp.ID, &cp.Title, &cp.Description, &cp.Type,
			&cp.FilePath, &cp.Content, &cp.InstructorID, &cp.IsPublic, &cp.CreatedAt, &cp.Inscrito)
		result = append(result, cp)
	}
	c.JSON(http.StatusOK, result)
}

func Inscribirse(c *gin.Context) {
	userID, _ := c.Get("user_id")
	capID := c.Param("id")

	// Verificar que el curso es público
	var isPublic bool
	err := db.DB.QueryRow(`SELECT is_public FROM capacitaciones WHERE id=$1`, capID).Scan(&isPublic)
	if err != nil || !isPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "el curso no existe o no es público"})
		return
	}

	_, err = db.DB.Exec(
		`INSERT INTO inscripciones(user_id, capacitacion_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		userID, capID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Unirse por código (cualquier usuario autenticado) ─────────────────────────

func UnirseConCodigo(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var body struct {
		Codigo string `json:"codigo" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere el campo 'codigo'"})
		return
	}

	// Buscar el curso por código (case-insensitive, trimmed)
	code := strings.ToUpper(strings.TrimSpace(body.Codigo))
	var cap models.Capacitacion
	err := db.DB.QueryRow(
		`SELECT id, title, description, type FROM capacitaciones WHERE codigo_acceso=$1`, code,
	).Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "código inválido o curso no encontrado"})
		return
	}

	_, err = db.DB.Exec(
		`INSERT INTO inscripciones(user_id, capacitacion_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		userID, cap.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"id":          cap.ID,
		"title":       cap.Title,
		"description": cap.Description,
		"type":        cap.Type,
	})
}

// InstructorResetCodigo genera un nuevo código de acceso para un curso del instructor
func InstructorResetCodigo(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")

	// Verificar propiedad
	var existing string
	err := db.DB.QueryRow(`SELECT id FROM capacitaciones WHERE id=$1 AND instructor_id=$2`, id, instructorID).Scan(&existing)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado o no encontrado"})
		return
	}

	newCode := uniqueCode()
	_, err = db.DB.Exec(`UPDATE capacitaciones SET codigo_acceso=$1 WHERE id=$2`, newCode, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"codigo_acceso": newCode})
}

// PreviewCurso devuelve info pública de un curso por código (sin autenticación)
func PreviewCurso(c *gin.Context) {
	code := strings.ToUpper(strings.TrimSpace(c.Param("codigo")))
	var cap models.Capacitacion
	err := db.DB.QueryRow(
		`SELECT id, title, description, type FROM capacitaciones WHERE codigo_acceso=$1`, code,
	).Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "código inválido o curso no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          cap.ID,
		"title":       cap.Title,
		"description": cap.Description,
		"type":        cap.Type,
	})
}
