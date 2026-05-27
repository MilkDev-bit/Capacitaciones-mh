package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ListCapacitaciones(c *gin.Context) {
	rows, err := db.DB.Query(`
		SELECT id, title, description, type,
		       COALESCE(file_path,''), COALESCE(content,''),
		       COALESCE(welcome_message,''), COALESCE(thumbnail_url,''),
		       COALESCE(color,'#f97316'), is_public,
		       COALESCE(codigo_acceso,''), created_at
		FROM capacitaciones ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Capacitacion{}
	for rows.Next() {
		var cap models.Capacitacion
		var createdAt interface{}
		rows.Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type,
			&cap.FilePath, &cap.Content,
			&cap.WelcomeMessage, &cap.ThumbnailURL,
			&cap.Color, &cap.IsPublic,
			&cap.CodigoAcceso, &createdAt)
		result = append(result, cap)
	}
	c.JSON(http.StatusOK, result)
}

func CreateCapacitacion(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	capType := c.PostForm("type")
	content := c.PostForm("content")
	welcomeMsg := c.PostForm("welcome_message")
	isPublicStr := c.PostForm("is_public")
	isPublic := isPublicStr == "true"
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

	var id string
	err = db.DB.QueryRow(
		`INSERT INTO capacitaciones(title, description, type, file_path, content, welcome_message, is_public, color, thumbnail_url)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		title, description, capType, filePath, content, welcomeMsg, isPublic, color, thumbnailPath,
	).Scan(&id)
	if err != nil {
		log.Printf("[ERROR] CreateCapacitacion: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar la capacitación"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCapacitacion(c *gin.Context) {
	id := c.Param("id")
	title := c.PostForm("title")
	description := c.PostForm("description")
	welcomeMsg := c.PostForm("welcome_message")
	isPublicStr := c.PostForm("is_public")
	isPublic := isPublicStr == "true"
	color := c.PostForm("color")
	if color == "" {
		color = "#f97316"
	}
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "título es requerido"})
		return
	}

	var thumbnailPath string
	thumbFile, err := c.FormFile("thumbnail")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(thumbFile.Filename))
		newName := uuid.NewString() + ext
		dest := filepath.Join("uploads", "thumbnails", newName)
		if saveErr := c.SaveUploadedFile(thumbFile, dest); saveErr == nil {
			thumbnailPath = "/" + filepath.ToSlash(dest)
		}
	}

	var execErr error
	if thumbnailPath != "" {
		_, execErr = db.DB.Exec(
			`UPDATE capacitaciones SET title=$1, description=$2, welcome_message=$3, color=$4, is_public=$5, thumbnail_url=$6 WHERE id=$7`,
			title, description, welcomeMsg, color, isPublic, thumbnailPath, id,
		)
	} else {
		_, execErr = db.DB.Exec(
			`UPDATE capacitaciones SET title=$1, description=$2, welcome_message=$3, color=$4, is_public=$5 WHERE id=$6`,
			title, description, welcomeMsg, color, isPublic, id,
		)
	}
	if execErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": execErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCapacitacion(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec(`DELETE FROM capacitaciones WHERE id=$1`, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListCapacitacionesUsuario(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT DISTINCT c.id, c.title, c.description, c.type,
		       COALESCE(c.file_path,''), COALESCE(c.content,''),
		       COALESCE(c.thumbnail_url,''), COALESCE(c.color,'#f97316'), c.created_at
		FROM capacitaciones c
		LEFT JOIN asignaciones a ON a.capacitacion_id = c.id AND a.user_id = $1
		LEFT JOIN inscripciones i ON i.capacitacion_id = c.id AND i.user_id = $1
		WHERE a.user_id = $1 OR i.user_id = $1
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Capacitacion{}
	for rows.Next() {
		var cap models.Capacitacion
		rows.Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type, &cap.FilePath, &cap.Content, &cap.ThumbnailURL, &cap.Color, &cap.CreatedAt)
		result = append(result, cap)
	}
	c.JSON(http.StatusOK, result)
}

func GetCapacitacion(c *gin.Context) {
	id := c.Param("id")
	var cap models.Capacitacion
	var createdAt time.Time
	err := db.DB.QueryRow(
		`SELECT id, title, description, type, COALESCE(file_path,''), COALESCE(content,''),
		        COALESCE(welcome_message,''), COALESCE(thumbnail_url,''), COALESCE(color,'#f97316'), created_at
		 FROM capacitaciones WHERE id=$1`, id,
	).Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type, &cap.FilePath, &cap.Content,
		&cap.WelcomeMessage, &cap.ThumbnailURL, &cap.Color, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	cap.CreatedAt = createdAt
	c.JSON(http.StatusOK, cap)
}
