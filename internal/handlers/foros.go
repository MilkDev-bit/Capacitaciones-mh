package handlers

import (
	"net/http"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
)

func ListForoPosts(c *gin.Context) {
	leccionID := c.Param("leccion_id")
	userID, _ := c.Get("user_id")

	rows, err := db.DB.Query(`
		SELECT p.id, p.leccion_id, p.user_id, u.name, p.titulo, p.contenido, p.created_at,
		       COUNT(DISTINCT fl.id) AS like_count,
		       COALESCE(BOOL_OR(fl.user_id = $2::uuid), false) AS user_liked
		FROM foro_posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN foro_likes fl ON fl.post_id = p.id
		WHERE p.leccion_id = $1
		GROUP BY p.id, u.name
		ORDER BY p.created_at DESC`, leccionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	result := []models.ForoPost{}
	for rows.Next() {
		var p models.ForoPost
		rows.Scan(&p.ID, &p.LeccionID, &p.UserID, &p.UserName, &p.Titulo, &p.Contenido, &p.CreatedAt, &p.LikeCount, &p.UserLiked)
		result = append(result, p)
	}
	c.JSON(http.StatusOK, result)
}

func CreateForoPost(c *gin.Context) {
	leccionID := c.Param("leccion_id")
	userID, _ := c.Get("user_id")

	var body struct {
		Titulo    string `json:"titulo" binding:"required"`
		Contenido string `json:"contenido" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id string
	err := db.DB.QueryRow(`
		INSERT INTO foro_posts(leccion_id, user_id, titulo, contenido)
		VALUES($1,$2,$3,$4) RETURNING id`,
		leccionID, userID, body.Titulo, body.Contenido,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
