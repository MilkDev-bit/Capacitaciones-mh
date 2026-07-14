package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"net/http"
	"path/filepath"
	"strings"

	"Prueba-Go/internal/cache"
	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"
	"Prueba-Go/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func generateCode(n int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}

func codeFromUUID(id string) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	u, err := uuid.Parse(id)
	if err != nil {
		return generateCode(8)
	}
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[u[i]%32]
	}
	return string(b)
}

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

func InstructorListCapacitaciones(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT id, title, description, type,
		       COALESCE(file_path,''), COALESCE(content,''),
		       instructor_id, is_public, COALESCE(codigo_acceso,''),
		       COALESCE(welcome_message,''), COALESCE(thumbnail_url,''),
		       COALESCE(color,'#f97316'), created_at
		FROM capacitaciones
		WHERE instructor_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, instructorID)
	if err != nil {
		slog.Error("InstructorListCapacitaciones", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()
	result := []models.Capacitacion{}
	for rows.Next() {
		var cap models.Capacitacion
		rows.Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type,
			&cap.FilePath, &cap.Content, &cap.InstructorID, &cap.IsPublic, &cap.CodigoAcceso,
			&cap.WelcomeMessage, &cap.ThumbnailURL, &cap.Color, &cap.CreatedAt)
		if cap.CodigoAcceso == "" {
			newCode := codeFromUUID(cap.ID)
			_, _ = db.DB.Exec(`UPDATE capacitaciones SET codigo_acceso=$1 WHERE id=$2 AND (codigo_acceso IS NULL OR codigo_acceso='')`, newCode, cap.ID)
			cap.CodigoAcceso = newCode
		}
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

	allowedContent := map[string]map[string]bool{
		"video":    {".mp4": true, ".webm": true, ".mov": true},
		"document": {".pdf": true, ".doc": true, ".docx": true, ".pptx": true, ".xlsx": true},
	}
	var filePath string
	if fu := c.PostForm("file_url"); fu != "" {
		filePath = fu
	} else if file, ferr := c.FormFile("file"); ferr == nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if allowed, ok := allowedContent[capType]; ok && !allowed[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tipo de archivo no permitido para el formato seleccionado"})
			return
		}
		prefix := "documents"
		if capType == "video" {
			prefix = "videos"
		}
		var uploadErr error
		filePath, uploadErr = storage.UploadMultipart(c.Request.Context(), file, prefix)
		if uploadErr != nil {
			slog.Error("InstructorCreateCapacitacion: subida archivo", "error", uploadErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error subiendo archivo"})
			return
		}
	}

	allowedThumb := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	var thumbnailPath string
	if tu := c.PostForm("thumbnail_url"); tu != "" {
		thumbnailPath = tu
	} else if thumbFile, ferr := c.FormFile("thumbnail"); ferr == nil {
		ext := strings.ToLower(filepath.Ext(thumbFile.Filename))
		if !allowedThumb[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "miniatura: formato no permitido (jpg, png, webp)"})
			return
		}
		if u, e := storage.UploadMultipart(c.Request.Context(), thumbFile, "thumbnails"); e == nil {
			thumbnailPath = u
		} else {
			slog.Warn("InstructorCreateCapacitacion: subida miniatura", "error", e)
		}
	}

	id := uuid.NewString()
	codigo := codeFromUUID(id)
	if _, err := db.DB.Exec(
		`INSERT INTO capacitaciones(id, title, description, type, file_path, content, instructor_id, is_public, codigo_acceso, welcome_message, thumbnail_url, color)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, title, description, capType, filePath, content, instructorID, isPublic, codigo, welcomeMessage, thumbnailPath, color,
	); err != nil {
		slog.Error("InstructorCreateCapacitacion: INSERT", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar la capacitación"})
		return
	}
	cache.C.Flush() // invalidate public course cache
	c.JSON(http.StatusCreated, gin.H{"id": id, "codigo_acceso": codigo})
}

func InstructorUpdateCapacitacion(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")

	var currentFilePath, currentThumbPath string
	err := db.DB.QueryRow(`SELECT COALESCE(file_path,''), COALESCE(thumbnail_url,'') FROM capacitaciones WHERE id=$1 AND instructor_id=$2 AND deleted_at IS NULL`, id, instructorID).Scan(&currentFilePath, &currentThumbPath)
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

	if fu := c.PostForm("file_url"); fu != "" {
		currentFilePath = fu
	} else if file, ferr := c.FormFile("file"); ferr == nil {
		prefix := "documents"
		if capType == "video" {
			prefix = "videos"
		}
		if u, e := storage.UploadMultipart(c.Request.Context(), file, prefix); e == nil {
			currentFilePath = u
		} else {
			slog.Warn("InstructorUpdateCapacitacion: subida archivo", "error", e)
		}
	}

	if tu := c.PostForm("thumbnail_url"); tu != "" {
		currentThumbPath = tu
	} else if thumbFile, ferr := c.FormFile("thumbnail"); ferr == nil {
		if u, e := storage.UploadMultipart(c.Request.Context(), thumbFile, "thumbnails"); e == nil {
			currentThumbPath = u
		} else {
			slog.Warn("InstructorUpdateCapacitacion: subida miniatura", "error", e)
		}
	} else if c.PostForm("remove_thumbnail") == "true" {
		currentThumbPath = ""
	}

	_, err = db.DB.Exec(
		`UPDATE capacitaciones SET title=$1, description=$2, type=$3, file_path=$4, content=$5, is_public=$6, welcome_message=$7, thumbnail_url=$8, color=$9 WHERE id=$10`,
		title, description, capType, currentFilePath, content, isPublic, welcomeMessage, currentThumbPath, color, id,
	)
	if err != nil {
		slog.Error("InstructorUpdateCapacitacion: UPDATE", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar la capacitación"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func InstructorDeleteCapacitacion(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")
	res, err := db.DB.Exec(`UPDATE capacitaciones SET deleted_at=NOW() WHERE id=$1 AND instructor_id=$2 AND deleted_at IS NULL`, id, instructorID)
	if err != nil {
		slog.Error("InstructorDeleteCapacitacion", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado o no encontrado"})
		return
	}
	cache.C.Flush() // invalidate public course cache
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func InstructorTogglePublic(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")

	var current bool
	err := db.DB.QueryRow(`SELECT is_public FROM capacitaciones WHERE id=$1 AND instructor_id=$2 AND deleted_at IS NULL`, id, instructorID).Scan(&current)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado o sin permisos"})
		return
	}
	_, err = db.DB.Exec(`UPDATE capacitaciones SET is_public=$1 WHERE id=$2`, !current, id)
	if err != nil {
		slog.Error("InstructorTogglePublic", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	cache.C.Flush() // invalidate public course cache
	c.JSON(http.StatusOK, gin.H{"is_public": !current})
}

func InstructorListExamenes(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT id, title, description, created_at
		FROM examenes WHERE instructor_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC
	`, instructorID)
	if err != nil {
		slog.Error("InstructorListExamenes", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
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
		bindError(c, err)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		slog.Error("InstructorCreateExamen: Begin", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer tx.Rollback()

	var examenID string
	err = tx.QueryRow(
		`INSERT INTO examenes(title, description, instructor_id, capacitacion_id) VALUES($1,$2,$3,$4) RETURNING id`,
		req.Title, req.Description, instructorID, req.CapacitacionID,
	).Scan(&examenID)
	if err != nil {
		slog.Error("InstructorCreateExamen: INSERT", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
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
			slog.Error("InstructorCreateExamen: INSERT pregunta", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
			return
		}
		if tipo != "open_text" {
			for _, o := range p.Opciones {
				_, err = tx.Exec(
					`INSERT INTO opciones(pregunta_id, texto, es_correcta) VALUES($1,$2,$3)`,
					preguntaID, o.Texto, o.EsCorrecta,
				)
				if err != nil {
					slog.Error("InstructorCreateExamen: INSERT opcion", "error", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
					return
				}
			}
		}
	}

	if err = tx.Commit(); err != nil {
		slog.Error("InstructorCreateExamen: Commit", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": examenID})
}

func InstructorDeleteExamen(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")
	res, err := db.DB.Exec(`UPDATE examenes SET deleted_at=NOW() WHERE id=$1 AND instructor_id=$2 AND deleted_at IS NULL`, id, instructorID)
	if err != nil {
		slog.Error("InstructorDeleteExamen", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado o no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func InstructorListEstudiantes(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT DISTINCT u.id, u.name, u.email, u.role, u.created_at
		FROM users u
		INNER JOIN inscripciones i ON i.user_id = u.id
		INNER JOIN capacitaciones c ON c.id = i.capacitacion_id
		WHERE c.instructor_id = $1
		ORDER BY u.name
	`, instructorID)
	if err != nil {
		slog.Error("InstructorListEstudiantes", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
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

func InstructorAsignar(c *gin.Context) {
	instructorID, _ := c.Get("user_id")

	var req asignarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}
	if req.CapacitacionID == nil && req.ExamenID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere capacitacion_id o examen_id"})
		return
	}

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

func ListCursosPublicos(c *gin.Context) {
	userID, _ := c.Get("user_id")

	type CursoPublico struct {
		models.Capacitacion
		Inscrito bool `json:"inscrito"`
	}

	// Cache the public course list per user so enrolled status is correct.
	cacheKey := "cursos_publicos:" + userID.(string)
	if cached, found := cache.C.Get(cacheKey); found {
		c.JSON(http.StatusOK, cached)
		return
	}

	rows, err := db.DB.Query(`
		SELECT c.id, c.title, c.description, c.type,
		       COALESCE(c.file_path,''), COALESCE(c.content,''),
		       COALESCE(c.thumbnail_url,''),
		       c.instructor_id, c.is_public, c.created_at,
		       EXISTS(SELECT 1 FROM inscripciones i WHERE i.capacitacion_id=c.id AND i.user_id=$1) as inscrito
		FROM capacitaciones c
		WHERE c.is_public = true AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		slog.Error("ListCursosPublicos", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	result := []CursoPublico{}
	for rows.Next() {
		var cp CursoPublico
		rows.Scan(&cp.ID, &cp.Title, &cp.Description, &cp.Type,
			&cp.FilePath, &cp.Content, &cp.ThumbnailURL, &cp.InstructorID, &cp.IsPublic, &cp.CreatedAt, &cp.Inscrito)
		result = append(result, cp)
	}
	cache.C.Set(cacheKey, result, 0) // uses default TTL (2 min)
	c.JSON(http.StatusOK, result)
}

func Inscribirse(c *gin.Context) {
	userID, _ := c.Get("user_id")
	capID := c.Param("id")
	var isPublic bool
	err := db.DB.QueryRow(`SELECT is_public FROM capacitaciones WHERE id=$1 AND deleted_at IS NULL`, capID).Scan(&isPublic)
	if err != nil || !isPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "el curso no existe o no es público"})
		return
	}

	_, err = db.DB.Exec(
		`INSERT INTO inscripciones(user_id, capacitacion_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		userID, capID,
	)
	if err != nil {
		slog.Error("Inscribirse: INSERT", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnirseConCodigo(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var body struct {
		Codigo string `json:"codigo" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere el campo 'codigo'"})
		return
	}

	code := strings.ToUpper(strings.TrimSpace(body.Codigo))
	var cap models.Capacitacion
	err := db.DB.QueryRow(
		`SELECT id, title, description, type FROM capacitaciones WHERE codigo_acceso=$1 AND deleted_at IS NULL`, code,
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
		slog.Error("UnirseConCodigo: INSERT inscripcion", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
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

func InstructorResetCodigo(c *gin.Context) {
	instructorID, _ := c.Get("user_id")
	id := c.Param("id")
	var existing string
	err := db.DB.QueryRow(`SELECT id FROM capacitaciones WHERE id=$1 AND instructor_id=$2 AND deleted_at IS NULL`, id, instructorID).Scan(&existing)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado o no encontrado"})
		return
	}

	newCode, err := uniqueCode()
	if err != nil {
		slog.Error("InstructorResetCodigo: UPDATE", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar nuevo código"})
		return
	}
	_, err = db.DB.Exec(`UPDATE capacitaciones SET codigo_acceso=$1 WHERE id=$2`, newCode, id)
	if err != nil {
		log.Printf("[ERROR] InstructorResetCodigo update: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"codigo_acceso": newCode})
}

func PreviewCurso(c *gin.Context) {
	code := strings.ToUpper(strings.TrimSpace(c.Param("codigo")))
	var cap models.Capacitacion
	err := db.DB.QueryRow(
		`SELECT id, title, description, type FROM capacitaciones WHERE codigo_acceso=$1 AND deleted_at IS NULL`, code,
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
