package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
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

// codeFromUUID deriva un código de acceso de 8 caracteres a partir de un UUID.
// Usa los primeros 8 bytes del UUID mapeados al charset de 32 chars.
// 256 % 32 == 0 → distribución perfectamente uniforme, sin sesgo.
// Dos UUIDs distintos (únicos por definición) producen códigos distintos con
// probabilidad abrumadoramente alta sin necesitar consulta a la BD.
func codeFromUUID(id string) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	u, err := uuid.Parse(id)
	if err != nil {
		return generateCode(8) // fallback improbable
	}
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[u[i]%32]
	}
	return string(b)
}

// uniqueCode genera un código aleatorio único (usado solo en reset de código).
func uniqueCode() (string, error) {
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		code := generateCode(8)
		var existing string
		err := db.DB.QueryRow(`SELECT id FROM capacitaciones WHERE codigo_acceso=$1`, code).Scan(&existing)
		if err == sql.ErrNoRows {
			return code, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("no se pudo generar un código único después de %d intentos", maxAttempts)
}

// ── Capacitaciones del instructor ─────────────────────────────────────────────

func InstructorListCapacitaciones(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT id, title, description, type,
		       COALESCE(file_path,''), COALESCE(content,''),
		       instructor_id, is_public, COALESCE(codigo_acceso,''),
		       COALESCE(welcome_message,''), COALESCE(thumbnail_url,''),
		       COALESCE(color,'#f97316'), created_at
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
			&cap.WelcomeMessage, &cap.ThumbnailURL, &cap.Color, &cap.CreatedAt)
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
	color := c.PostForm("color")
	if color == "" {
		color = "#f97316"
	}

	if title == "" || capType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title y type son requeridos"})
		return
	}

	// A06: validar tipos de archivo
	allowedContent := map[string]map[string]bool{
		"video":    {".mp4": true, ".webm": true, ".mov": true},
		"document": {".pdf": true, ".doc": true, ".docx": true, ".pptx": true, ".xlsx": true},
	}
	var filePath string
	file, err := c.FormFile("file")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if allowed, ok := allowedContent[capType]; ok && !allowed[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tipo de archivo no permitido para el formato seleccionado"})
			return
		}
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

	allowedThumb := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	var thumbnailPath string
	thumbFile, err := c.FormFile("thumbnail")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(thumbFile.Filename))
		if !allowedThumb[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "miniatura: formato no permitido (jpg, png, webp)"})
			return
		}
		newName := uuid.NewString() + ext
		dest := filepath.Join("uploads", "thumbnails", newName)
		if err := c.SaveUploadedFile(thumbFile, dest); err == nil {
			thumbnailPath = "/" + filepath.ToSlash(dest)
		}
	}

	// Generar UUID en Go → derivar código determinísticamente (sin consulta a BD)
	id := uuid.NewString()
	codigo := codeFromUUID(id)
	_, err = db.DB.Exec(
		`INSERT INTO capacitaciones(id, title, description, type, file_path, content, instructor_id, is_public, codigo_acceso, welcome_message, thumbnail_url, color)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, title, description, capType, filePath, content, instructorID, isPublic, codigo, welcomeMessage, thumbnailPath, color,
	)
	if err != nil {
		log.Printf("[ERROR] InstructorCreateCapacitacion: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar la capacitación"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "codigo_acceso": codigo})
}

func InstructorUpdateCapacitacion(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")

	var currentFilePath, currentThumbPath string
	err := db.DB.QueryRow(`SELECT COALESCE(file_path,''), COALESCE(thumbnail_url,'') FROM capacitaciones WHERE id=$1 AND instructor_id=$2`, id, instructorID).Scan(&currentFilePath, &currentThumbPath)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado o no encontrado"})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	capType := c.PostForm("type")
	content := c.PostForm("content")
	isPublic := c.PostForm("is_public") == "true"
	welcomeMessage := c.PostForm("welcome_message")
	color := c.PostForm("color")
	if color == "" {
		color = "#f97316"
	}

	if title == "" || capType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title y type son requeridos"})
		return
	}

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
		if err := c.SaveUploadedFile(file, dest); err == nil {
			currentFilePath = "/" + filepath.ToSlash(dest)
		}
	}

	thumbFile, err := c.FormFile("thumbnail")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(thumbFile.Filename))
		newName := uuid.NewString() + ext
		dest := filepath.Join("uploads", "thumbnails", newName)
		if err := c.SaveUploadedFile(thumbFile, dest); err == nil {
			currentThumbPath = "/" + filepath.ToSlash(dest)
		}
	} else if c.PostForm("remove_thumbnail") == "true" {
		currentThumbPath = ""
	}

	_, err = db.DB.Exec(
		`UPDATE capacitaciones SET title=$1, description=$2, type=$3, file_path=$4, content=$5, is_public=$6, welcome_message=$7, thumbnail_url=$8, color=$9 WHERE id=$10`,
		title, description, capType, currentFilePath, content, isPublic, welcomeMessage, currentThumbPath, color, id,
	)
	if err != nil {
		log.Printf("[ERROR] InstructorUpdateCapacitacion: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar la capacitación"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func InstructorDeleteCapacitacion(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")
	res, err := db.DB.Exec(`DELETE FROM capacitaciones WHERE id=$1 AND instructor_id=$2`, id, instructorID)
	if err != nil {
		log.Printf("[ERROR] InstructorDeleteCapacitacion: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
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
		       COALESCE(c.thumbnail_url,''),
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
			&cp.FilePath, &cp.Content, &cp.ThumbnailURL, &cp.InstructorID, &cp.IsPublic, &cp.CreatedAt, &cp.Inscrito)
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

	newCode, err := uniqueCode()
	if err != nil {
		log.Printf("[ERROR] uniqueCode reset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar nuevo código"})
		return
	}
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
