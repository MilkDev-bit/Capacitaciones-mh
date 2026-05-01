package handlers

import (
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
	rows, err := db.DB.Query(`SELECT id, title, description, type, COALESCE(file_path,''), COALESCE(content,''), created_at FROM capacitaciones ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Capacitacion{}
	for rows.Next() {
		var cap models.Capacitacion
		rows.Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type, &cap.FilePath, &cap.Content, &cap.CreatedAt)
		result = append(result, cap)
	}
	c.JSON(http.StatusOK, result)
}

func CreateCapacitacion(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	capType := c.PostForm("type") // video | document | text
	content := c.PostForm("content")

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

	var id string
	err = db.DB.QueryRow(
		`INSERT INTO capacitaciones(title, description, type, file_path, content) VALUES($1,$2,$3,$4,$5) RETURNING id`,
		title, description, capType, filePath, content,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func DeleteCapacitacion(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec(`DELETE FROM capacitaciones WHERE id=$1`, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Para usuario: lista solo las asignadas
func ListCapacitacionesUsuario(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`
		SELECT DISTINCT c.id, c.title, c.description, c.type,
		       COALESCE(c.file_path,''), COALESCE(c.content,''), c.created_at
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
		rows.Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type, &cap.FilePath, &cap.Content, &cap.CreatedAt)
		result = append(result, cap)
	}
	c.JSON(http.StatusOK, result)
}

func GetCapacitacion(c *gin.Context) {
	id := c.Param("id")
	var cap models.Capacitacion
	var createdAt time.Time
	err := db.DB.QueryRow(
		`SELECT id, title, description, type, COALESCE(file_path,''), COALESCE(content,''), created_at FROM capacitaciones WHERE id=$1`, id,
	).Scan(&cap.ID, &cap.Title, &cap.Description, &cap.Type, &cap.FilePath, &cap.Content, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	cap.CreatedAt = createdAt
	c.JSON(http.StatusOK, cap)
}
