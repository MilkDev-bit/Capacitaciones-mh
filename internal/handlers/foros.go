package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"
	"Prueba-Go/internal/storage"

	"github.com/gin-gonic/gin"
)

func ListForoPosts(c *gin.Context) {
	leccionID := c.Param("leccion_id")
	userID, _ := c.Get("user_id")

	rows, err := db.DB.Query(`
		SELECT p.id, p.leccion_id, p.user_id, u.name, p.titulo, p.contenido,
		       COALESCE(p.media_url,''), COALESCE(p.media_type,''),
		       p.created_at,
		       COUNT(DISTINCT fl.id) AS like_count,
		       COALESCE(BOOL_OR(fl.user_id = $2::uuid), false) AS user_liked
		FROM foro_posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN foro_likes fl ON fl.post_id = p.id
		WHERE p.leccion_id = $1
		GROUP BY p.id, u.name
		ORDER BY p.created_at DESC`, leccionID, userID)
	if err != nil {
		log.Printf("[ERROR] ListForoPosts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	result := []models.ForoPost{}
	for rows.Next() {
		var p models.ForoPost
		rows.Scan(&p.ID, &p.LeccionID, &p.UserID, &p.UserName, &p.Titulo, &p.Contenido, &p.MediaURL, &p.MediaType, &p.CreatedAt, &p.LikeCount, &p.UserLiked)
		result = append(result, p)
	}
	c.JSON(http.StatusOK, result)
}

func CreateForoPost(c *gin.Context) {
	leccionID := c.Param("leccion_id")
	userID, _ := c.Get("user_id")

	titulo := strings.TrimSpace(c.PostForm("titulo"))
	contenido := strings.TrimSpace(c.PostForm("contenido"))
	if titulo == "" || contenido == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "titulo y contenido son requeridos"})
		return
	}

	var mediaURL, mediaType string
	file, err := c.FormFile("media")
	if err == nil {
		if file.Size > 50*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el archivo no puede superar 50 MB"})
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		imgExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
		vidExts := map[string]bool{".mp4": true, ".webm": true, ".mov": true, ".avi": true}
		if imgExts[ext] {
			mediaType = "image"
		} else if vidExts[ext] {
			mediaType = "video"
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "solo se permiten imágenes (jpg,png,webp,gif) o videos (mp4,webm,mov)"})
			return
		}
		var u string
		u, err = storage.UploadMultipart(c.Request.Context(), file, "foro")
		if err != nil {
			log.Printf("[ERROR] CreateForoPost upload: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error subiendo archivo"})
			return
		}
		mediaURL = u
	}

	var id string
	err = db.DB.QueryRow(`
		INSERT INTO foro_posts(leccion_id, user_id, titulo, contenido, media_url, media_type)
		VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		leccionID, userID, titulo, contenido, mediaURL, mediaType,
	).Scan(&id)
	if err != nil {
		log.Printf("[ERROR] CreateForoPost: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func DeleteForoPost(c *gin.Context) {
	postID := c.Param("post_id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var query string
	var args []interface{}
	if role == "admin" || role == "instructor" {
		query = `DELETE FROM foro_posts WHERE id=$1`
		args = []interface{}{postID}
	} else {
		query = `DELETE FROM foro_posts WHERE id=$1 AND user_id=$2`
		args = []interface{}{postID, userID}
	}
	db.DB.Exec(query, args...)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListForoComentarios(c *gin.Context) {
	postID := c.Param("post_id")

	rows, err := db.DB.Query(`
		SELECT fc.id, fc.post_id, fc.user_id, u.name, fc.contenido, fc.created_at
		FROM foro_comentarios fc
		JOIN users u ON u.id = fc.user_id
		WHERE fc.post_id = $1
		ORDER BY fc.created_at ASC`, postID)
	if err != nil {
		log.Printf("[ERROR] ListForoComentarios: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	result := []models.ForoComentario{}
	for rows.Next() {
		var fc models.ForoComentario
		rows.Scan(&fc.ID, &fc.PostID, &fc.UserID, &fc.UserName, &fc.Contenido, &fc.CreatedAt)
		result = append(result, fc)
	}
	c.JSON(http.StatusOK, result)
}

func CreateForoComentario(c *gin.Context) {
	postID := c.Param("post_id")
	userID, _ := c.Get("user_id")

	var body struct {
		Contenido string `json:"contenido" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id string
	err := db.DB.QueryRow(`
		INSERT INTO foro_comentarios(post_id, user_id, contenido)
		VALUES($1,$2,$3) RETURNING id`,
		postID, userID, body.Contenido,
	).Scan(&id)
	if err != nil {
		log.Printf("[ERROR] CreateForoComentario: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ToggleForoPostLike(c *gin.Context) {
	postID := c.Param("post_id")
	userID, _ := c.Get("user_id")

	var exists bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM foro_likes WHERE post_id=$1 AND user_id=$2)`,
		postID, userID).Scan(&exists)

	if exists {
		db.DB.Exec(`DELETE FROM foro_likes WHERE post_id=$1 AND user_id=$2`, postID, userID)
	} else {
		db.DB.Exec(`INSERT INTO foro_likes(post_id, user_id) VALUES($1, $2)`, postID, userID)
	}

	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM foro_likes WHERE post_id=$1`, postID).Scan(&count)
	c.JSON(http.StatusOK, gin.H{"liked": !exists, "count": count})
}
