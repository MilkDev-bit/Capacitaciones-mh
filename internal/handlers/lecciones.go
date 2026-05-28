package handlers

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"
	"Prueba-Go/internal/storage"

	"github.com/gin-gonic/gin"
)

func InstructorListLecciones(c *gin.Context) {
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

	rows, err := db.DB.Query(`
		SELECT id, capacitacion_id, title, COALESCE(description,''), type,
		       COALESCE(file_path,''), COALESCE(content,''), orden, COALESCE(duracion_min,0), created_at
		FROM lecciones WHERE capacitacion_id=$1 AND deleted_at IS NULL ORDER BY orden`, capID)
	if err != nil {
		log.Printf("[ERROR] InstructorListLecciones: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
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
	if fu := c.PostForm("file_url"); fu != "" {
		filePath = fu
	} else if file, ferr := c.FormFile("file"); ferr == nil {
		prefix := "documents"
		if lecType == "video" {
			prefix = "videos"
		}
		var e error
		filePath, e = storage.UploadMultipart(c.Request.Context(), file, prefix)
		if e != nil {
			slog.Error("InstructorCreateLeccion: subida archivo", "error", e)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error subiendo archivo"})
			return
		}
	}

	if orden == "" {
		orden = "0"
	}
	if duracion == "" {
		duracion = "0"
	}

	var id string
	err := db.DB.QueryRow(`
		INSERT INTO lecciones(capacitacion_id, title, description, type, file_path, content, orden, duracion_min)
		VALUES($1,$2,$3,$4,$5,$6,$7::int,$8::int) RETURNING id`,
		capID, title, description, lecType, filePath, content, orden, duracion,
	).Scan(&id)
	if err != nil {
		slog.Error("InstructorCreateLeccion: INSERT", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func InstructorUpdateLeccion(c *gin.Context) {
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

	var currentFilePath string
	err := db.DB.QueryRow(`SELECT COALESCE(file_path,'') FROM lecciones WHERE id=$1 AND capacitacion_id=$2`, lecID, capID).Scan(&currentFilePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lección no encontrada"})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	lecType := c.PostForm("type")
	content := c.PostForm("content")
	duracion := c.PostForm("duracion_min")

	if title == "" || lecType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title y type son requeridos"})
		return
	}

	if fu := c.PostForm("file_url"); fu != "" {
		currentFilePath = fu
	} else if file, ferr := c.FormFile("file"); ferr == nil {
		prefix := "documents"
		if lecType == "video" {
			prefix = "videos"
		}
		if u, e := storage.UploadMultipart(c.Request.Context(), file, prefix); e == nil {
			currentFilePath = u
		} else {
			slog.Warn("InstructorUpdateLeccion: subida archivo", "error", e)
		}
	}

	if duracion == "" {
		duracion = "0"
	}

	_, err = db.DB.Exec(
		`UPDATE lecciones SET title=$1, description=$2, type=$3, content=$4, duracion_min=$5::int, file_path=$6 WHERE id=$7 AND capacitacion_id=$8`,
		title, description, lecType, content, duracion, currentFilePath, lecID, capID,
	)
	if err != nil {
		slog.Error("InstructorUpdateLeccion: UPDATE", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
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

	db.DB.Exec(`UPDATE lecciones SET deleted_at=NOW() WHERE id=$1 AND capacitacion_id=$2 AND deleted_at IS NULL`, lecID, capID)
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
		bindError(c, err)
		return
	}

	tx, _ := db.DB.Begin()
	for _, o := range orden {
		tx.Exec(`UPDATE lecciones SET orden=$1 WHERE id=$2 AND capacitacion_id=$3`, o.Orden, o.ID, capID)
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetLeccionesConProgreso(c *gin.Context) {
	capID := c.Param("id")
	userID, _ := c.Get("user_id")

	rows, err := db.DB.Query(`
		SELECT l.id, l.capacitacion_id, l.title, COALESCE(l.description,''), l.type,
		       COALESCE(l.file_path,''), COALESCE(l.content,''), l.orden, COALESCE(l.duracion_min,0), l.created_at,
		       CASE WHEN p.leccion_id IS NOT NULL THEN true ELSE false END AS completada
		FROM lecciones l
		LEFT JOIN progreso_lecciones p ON p.leccion_id = l.id AND p.user_id = $2
		WHERE l.capacitacion_id = $1 AND l.deleted_at IS NULL
		ORDER BY l.orden`, capID, userID)
	if err != nil {
		slog.Error("GetLeccionesConProgreso", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()
	result := []models.Leccion{}
	for rows.Next() {
		var l models.Leccion
		rows.Scan(&l.ID, &l.CapacitacionID, &l.Title, &l.Description, &l.Type,
			&l.FilePath, &l.Content, &l.Orden, &l.DuracionMin, &l.CreatedAt, &l.Completada)
		// Reemplazar la URL pública por una URL firmada con TTL de 2 horas para video/documento
		if (l.Type == "video" || l.Type == "document") && l.FilePath != "" {
			key := storage.ExtractKeyFromURL(l.FilePath)
			if key != "" {
				if signed, err := storage.GeneratePresignedGetURL(c.Request.Context(), key, 2*time.Hour); err == nil {
					l.FilePath = signed
				} else {
					log.Printf("[WARN] GetLeccionesConProgreso presign %s: %v", key, err)
				}
			}
		}
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
