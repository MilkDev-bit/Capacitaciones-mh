package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ── Instructor: CRUD de lecciones ─────────────────────────────────────────────

func InstructorListLecciones(c *gin.Context) {
	capID := c.Param("id")
	instructorID, _ := c.Get("user_id")

	// Verificar pertenencia
	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, capID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, capacitacion_id, title, COALESCE(description,''), type,
		       COALESCE(file_path,''), COALESCE(content,''), orden, COALESCE(duracion_min,0), created_at
		FROM lecciones WHERE capacitacion_id=$1 ORDER BY orden`, capID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Leccion{}
	for rows.Next() {
		var l models.Leccion
		rows.Scan(&l.ID, &l.CapacitacionID, &l.Title, &l.Description, &l.Type,
			&l.FilePath, &l.Content, &l.Orden, &l.DuracionMin, &l.CreatedAt)
		result = append(result, l)
	}
	c.JSON(http.StatusOK, result)
}

func InstructorCreateLeccion(c *gin.Context) {
	capID := c.Param("id")
	instructorID, _ := c.Get("user_id")

	// Verificar pertenencia
	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, capID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	title := c.PostForm("title")
	lecType := c.PostForm("type")
	description := c.PostForm("description")
	content := c.PostForm("content")
	orden := c.PostForm("orden")
	duracion := c.PostForm("duracion_min")

	if title == "" || lecType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title y type son requeridos"})
		return
	}

	var filePath string
	file, err := c.FormFile("file")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		newName := uuid.NewString() + ext
		var dest string
		if lecType == "video" {
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

	if orden == "" {
		orden = "0"
	}
	if duracion == "" {
		duracion = "0"
	}

	var id string
	err = db.DB.QueryRow(`
		INSERT INTO lecciones(capacitacion_id, title, description, type, file_path, content, orden, duracion_min)
		VALUES($1,$2,$3,$4,$5,$6,$7::int,$8::int) RETURNING id`,
		capID, title, description, lecType, filePath, content, orden, duracion,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func InstructorDeleteLeccion(c *gin.Context) {
	capID := c.Param("id")
	lecID := c.Param("leccion_id")
	instructorID, _ := c.Get("user_id")

	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, capID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	db.DB.Exec(`DELETE FROM lecciones WHERE id=$1 AND capacitacion_id=$2`, lecID, capID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func InstructorReorderLecciones(c *gin.Context) {
	capID := c.Param("id")
	instructorID, _ := c.Get("user_id")

	var owner string
	if err := db.DB.QueryRow(`SELECT COALESCE(instructor_id::text,'') FROM capacitaciones WHERE id=$1`, capID).Scan(&owner); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
		return
	}
	if owner != instructorID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permisos"})
		return
	}

	var orden []struct {
		ID    string `json:"id"`
		Orden int    `json:"orden"`
	}
	if err := c.ShouldBindJSON(&orden); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, _ := db.DB.Begin()
	for _, o := range orden {
		tx.Exec(`UPDATE lecciones SET orden=$1 WHERE id=$2 AND capacitacion_id=$3`, o.Orden, o.ID, capID)
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Usuario: ver lecciones con progreso ───────────────────────────────────────

func GetLeccionesConProgreso(c *gin.Context) {
	capID := c.Param("id")
	userID, _ := c.Get("user_id")

	rows, err := db.DB.Query(`
		SELECT l.id, l.capacitacion_id, l.title, COALESCE(l.description,''), l.type,
		       COALESCE(l.file_path,''), COALESCE(l.content,''), l.orden, COALESCE(l.duracion_min,0), l.created_at,
		       CASE WHEN p.leccion_id IS NOT NULL THEN true ELSE false END AS completada
		FROM lecciones l
		LEFT JOIN progreso_lecciones p ON p.leccion_id = l.id AND p.user_id = $2
		WHERE l.capacitacion_id = $1
		ORDER BY l.orden`, capID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	result := []models.Leccion{}
	for rows.Next() {
		var l models.Leccion
		rows.Scan(&l.ID, &l.CapacitacionID, &l.Title, &l.Description, &l.Type,
			&l.FilePath, &l.Content, &l.Orden, &l.DuracionMin, &l.CreatedAt, &l.Completada)
		result = append(result, l)
	}
	c.JSON(http.StatusOK, result)
}

func MarcarLeccionCompleta(c *gin.Context) {
	lecID := c.Param("leccion_id")
	userID, _ := c.Get("user_id")

	db.DB.Exec(`
		INSERT INTO progreso_lecciones(user_id, leccion_id)
		VALUES($1,$2) ON CONFLICT(user_id, leccion_id) DO NOTHING`, userID, lecID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
