package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"
	"Prueba-Go/internal/sanitize"
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
		       COALESCE(BOOL_OR(fl.user_id = $2::uuid), false) AS user_liked,
		       COALESCE((
				SELECT jsonb_agg(jsonb_build_object('emoji', r.emoji, 'count', r.cnt))
				FROM (
					SELECT emoji, COUNT(*) as cnt 
					FROM foro_post_reactions 
					WHERE post_id = p.id 
					GROUP BY emoji
				) r
			   ), '[]'::jsonb) AS reactions
		FROM foro_posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN foro_likes fl ON fl.post_id = p.id
		WHERE p.leccion_id = $1
		GROUP BY p.id, u.name
		ORDER BY p.created_at DESC`, leccionID, userID)
	if err != nil {
		slog.Error("ListForoPosts", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	result := []models.ForoPost{}
	for rows.Next() {
		var p models.ForoPost
		var reactionsJSON []byte
		rows.Scan(&p.ID, &p.LeccionID, &p.UserID, &p.UserName, &p.Titulo, &p.Contenido, &p.MediaURL, &p.MediaType, &p.CreatedAt, &p.LikeCount, &p.UserLiked, &reactionsJSON)
		json.Unmarshal(reactionsJSON, &p.Reactions)
		result = append(result, p)
	}
	c.JSON(http.StatusOK, result)
}

func CreateForoPost(c *gin.Context) {
	leccionID := c.Param("leccion_id")
	userID, _ := c.Get("user_id")

	titulo := sanitize.Text(strings.TrimSpace(c.PostForm("titulo")))
	contenido := sanitize.HTML(strings.TrimSpace(c.PostForm("contenido")))
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
			slog.Error("CreateForoPost: subida archivo", "error", err)
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
		slog.Error("CreateForoPost: INSERT", "error", err)
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
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	rows, err := db.DB.Query(`
		SELECT fc.id, fc.post_id, fc.parent_id, fc.user_id, u.name, fc.contenido, 
		       fc.media_url, fc.media_type, fc.is_private, fc.created_at,
			   p.user_id AS post_author_id,
			   COALESCE(parent_com.user_id, p.user_id) AS parent_author_id,
			   COALESCE((
				SELECT jsonb_agg(jsonb_build_object('emoji', r.emoji, 'count', r.cnt))
				FROM (
					SELECT emoji, COUNT(*) as cnt 
					FROM foro_comentario_reactions 
					WHERE comentario_id = fc.id 
					GROUP BY emoji
				) r
			   ), '[]'::jsonb) AS reactions
		FROM foro_comentarios fc
		JOIN users u ON u.id = fc.user_id
		JOIN foro_posts p ON p.id = fc.post_id
		LEFT JOIN foro_comentarios parent_com ON parent_com.id = fc.parent_id
		WHERE fc.post_id = $1
		ORDER BY fc.created_at ASC`, postID)
	if err != nil {
		slog.Error("ListForoComentarios", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	defer rows.Close()

	result := []models.ForoComentario{}
	for rows.Next() {
		var fc models.ForoComentario
		var postAuthorID, parentAuthorID string
		var reactionsJSON []byte
		rows.Scan(&fc.ID, &fc.PostID, &fc.ParentID, &fc.UserID, &fc.UserName, &fc.Contenido, &fc.MediaURL, &fc.MediaType, &fc.IsPrivate, &fc.CreatedAt, &postAuthorID, &parentAuthorID, &reactionsJSON)
		json.Unmarshal(reactionsJSON, &fc.Reactions)
		
		// Filtrar privacidad
		if fc.IsPrivate {
			if role != "admin" && role != "instructor" && userID != fc.UserID && userID != parentAuthorID {
				continue // Ocultar
			}
		}

		result = append(result, fc)
	}
	c.JSON(http.StatusOK, result)
}

func CreateForoComentario(c *gin.Context) {
	postID := c.Param("post_id")
	userID, _ := c.Get("user_id")

	contenido := sanitize.HTML(strings.TrimSpace(c.PostForm("contenido")))
	if contenido == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el contenido es requerido"})
		return
	}

	parentID := c.PostForm("parent_id")
	isPrivate := c.PostForm("is_private") == "true"

	var parentIDPtr *string
	if parentID != "" {
		parentIDPtr = &parentID
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "solo se permiten imágenes o videos"})
			return
		}
		var u string
		u, err = storage.UploadMultipart(c.Request.Context(), file, "foro_comentarios")
		if err != nil {
			slog.Error("CreateForoComentario: subida archivo", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error subiendo archivo"})
			return
		}
		mediaURL = u
	}

	var id string
	err = db.DB.QueryRow(`
		INSERT INTO foro_comentarios(post_id, parent_id, user_id, contenido, media_url, media_type, is_private)
		VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		postID, parentIDPtr, userID, contenido, mediaURL, mediaType, isPrivate,
	).Scan(&id)
	if err != nil {
		slog.Error("CreateForoComentario: INSERT", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	// Trigger notificaciones async
	go func() {
		if parentIDPtr != nil {
			var parentAuthorID string
			db.DB.QueryRow(`SELECT user_id FROM foro_comentarios WHERE id=$1`, *parentIDPtr).Scan(&parentAuthorID)
			if parentAuthorID != "" && parentAuthorID != userID.(string) {
				CrearNotificacion(parentAuthorID, "foro_reply", "Nueva respuesta a tu comentario", "Alguien respondió a tu comentario en el foro.", "/usuario/capacitaciones")
			}
		} else {
			var postAuthorID string
			db.DB.QueryRow(`SELECT user_id FROM foro_posts WHERE id=$1`, postID).Scan(&postAuthorID)
			if postAuthorID != "" && postAuthorID != userID.(string) {
				CrearNotificacion(postAuthorID, "foro_comment", "Nuevo comentario en tu publicación", "Alguien comentó en tu publicación del foro.", "/usuario/capacitaciones")
			}
		}
	}()

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

func ToggleForoPostReaction(c *gin.Context) {
	postID := c.Param("post_id")
	userID, _ := c.Get("user_id")

	var body struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exists bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM foro_post_reactions WHERE post_id=$1 AND user_id=$2 AND emoji=$3)`, postID, userID, body.Emoji).Scan(&exists)

	if exists {
		db.DB.Exec(`DELETE FROM foro_post_reactions WHERE post_id=$1 AND user_id=$2 AND emoji=$3`, postID, userID, body.Emoji)
	} else {
		db.DB.Exec(`INSERT INTO foro_post_reactions(post_id, user_id, emoji) VALUES($1, $2, $3)`, postID, userID, body.Emoji)
		
		// Notificación (opcional)
		go func() {
			var postAuthorID string
			db.DB.QueryRow(`SELECT user_id FROM foro_posts WHERE id=$1`, postID).Scan(&postAuthorID)
			if postAuthorID != "" && postAuthorID != userID.(string) {
				CrearNotificacion(postAuthorID, "foro_reaction", "Nueva reacción", "Alguien reaccionó con " + body.Emoji + " a tu publicación.", "/usuario/capacitaciones")
			}
		}()
	}

	rows, _ := db.DB.Query(`SELECT emoji, COUNT(*) as cnt FROM foro_post_reactions WHERE post_id=$1 GROUP BY emoji`, postID)
	var reactions []models.ReactionCount
	for rows.Next() {
		var r models.ReactionCount
		rows.Scan(&r.Emoji, &r.Count)
		reactions = append(reactions, r)
	}
	rows.Close()

	c.JSON(http.StatusOK, gin.H{"status": "ok", "reactions": reactions})
}

func ToggleForoComentarioReaction(c *gin.Context) {
	comentarioID := c.Param("comentario_id")
	userID, _ := c.Get("user_id")

	var body struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exists bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM foro_comentario_reactions WHERE comentario_id=$1 AND user_id=$2 AND emoji=$3)`, comentarioID, userID, body.Emoji).Scan(&exists)

	if exists {
		db.DB.Exec(`DELETE FROM foro_comentario_reactions WHERE comentario_id=$1 AND user_id=$2 AND emoji=$3`, comentarioID, userID, body.Emoji)
	} else {
		db.DB.Exec(`INSERT INTO foro_comentario_reactions(comentario_id, user_id, emoji) VALUES($1, $2, $3)`, comentarioID, userID, body.Emoji)
		
		// Notificación (opcional)
		go func() {
			var authorID string
			db.DB.QueryRow(`SELECT user_id FROM foro_comentarios WHERE id=$1`, comentarioID).Scan(&authorID)
			if authorID != "" && authorID != userID.(string) {
				CrearNotificacion(authorID, "foro_reaction", "Nueva reacción", "Alguien reaccionó con " + body.Emoji + " a tu comentario.", "/usuario/capacitaciones")
			}
		}()
	}

	rows, _ := db.DB.Query(`SELECT emoji, COUNT(*) as cnt FROM foro_comentario_reactions WHERE comentario_id=$1 GROUP BY emoji`, comentarioID)
	var reactions []models.ReactionCount
	for rows.Next() {
		var r models.ReactionCount
		rows.Scan(&r.Emoji, &r.Count)
		reactions = append(reactions, r)
	}
	rows.Close()

	c.JSON(http.StatusOK, gin.H{"status": "ok", "reactions": reactions})
}

func ListForoPostReactions(c *gin.Context) {
	postID := c.Param("post_id")
	rows, err := db.DB.Query(`SELECT emoji, user_id FROM foro_post_reactions WHERE post_id=$1`, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	defer rows.Close()

	reactions := make(map[string][]string) // emoji -> []user_id
	for rows.Next() {
		var emoji, userID string
		rows.Scan(&emoji, &userID)
		reactions[emoji] = append(reactions[emoji], userID)
	}
	c.JSON(http.StatusOK, reactions)
}

func ListForoComentarioReactions(c *gin.Context) {
	comentarioID := c.Param("comentario_id")
	rows, err := db.DB.Query(`SELECT emoji, user_id FROM foro_comentario_reactions WHERE comentario_id=$1`, comentarioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	defer rows.Close()

	reactions := make(map[string][]string)
	for rows.Next() {
		var emoji, userID string
		rows.Scan(&emoji, &userID)
		reactions[emoji] = append(reactions[emoji], userID)
	}
	c.JSON(http.StatusOK, reactions)
}
