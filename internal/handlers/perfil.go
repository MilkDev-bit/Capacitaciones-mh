package handlers

import (
	"net/http"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetPerfil(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var u models.User
	err := db.DB.QueryRow(`
		SELECT id, name, email, role,
		       COALESCE(bio,''), COALESCE(avatar_url,''),
		       COALESCE(phone,''), COALESCE(specialty,''), created_at
		FROM users WHERE id=$1`, userID,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role,
		&u.Bio, &u.AvatarURL, &u.Phone, &u.Specialty, &u.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}

	// Build stats
	stats := gin.H{}

	if u.Role == "user" {
		var cursosInscritos, leccionesCompletadas, totalLecciones int
		db.DB.QueryRow(`SELECT COUNT(DISTINCT capacitacion_id) FROM inscripciones WHERE user_id=$1`, userID).Scan(&cursosInscritos)
		db.DB.QueryRow(`SELECT COUNT(*) FROM progreso_lecciones WHERE user_id=$1`, userID).Scan(&leccionesCompletadas)
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM lecciones l
			JOIN inscripciones i ON l.capacitacion_id = i.capacitacion_id
			WHERE i.user_id=$1`, userID).Scan(&totalLecciones)
		stats["cursos_inscritos"] = cursosInscritos
		stats["lecciones_completadas"] = leccionesCompletadas
		stats["total_lecciones"] = totalLecciones
	}

	if u.Role == "instructor" {
		var cursosCreados, estudiantesTotal, examenesCreados int
		db.DB.QueryRow(`SELECT COUNT(*) FROM capacitaciones WHERE instructor_id=$1`, userID).Scan(&cursosCreados)
		db.DB.QueryRow(`
			SELECT COUNT(DISTINCT i.user_id) FROM inscripciones i
			JOIN capacitaciones c ON i.capacitacion_id = c.id
			WHERE c.instructor_id=$1`, userID).Scan(&estudiantesTotal)
		db.DB.QueryRow(`SELECT COUNT(*) FROM examenes WHERE instructor_id=$1`, userID).Scan(&examenesCreados)
		stats["cursos_creados"] = cursosCreados
		stats["estudiantes_total"] = estudiantesTotal
		stats["examenes_creados"] = examenesCreados
	}

	c.JSON(http.StatusOK, gin.H{"user": u, "stats": stats})
}

func UpdatePerfil(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var body struct {
		Name      string `json:"name"`
		Bio       string `json:"bio"`
		Phone     string `json:"phone"`
		Specialty string `json:"specialty"`
		AvatarURL string `json:"avatar_url"`
		Password  string `json:"password"` // opcional: nueva contraseña
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Name != "" {
		db.DB.Exec(`UPDATE users SET name=$1 WHERE id=$2`, body.Name, userID)
	}
	db.DB.Exec(`UPDATE users SET bio=$1, phone=$2, specialty=$3, avatar_url=$4 WHERE id=$5`,
		body.Bio, body.Phone, body.Specialty, body.AvatarURL, userID)

	if body.Password != "" {
		if len(body.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "la contraseña debe tener al menos 6 caracteres"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error procesando contraseña"})
			return
		}
		db.DB.Exec(`UPDATE users SET password_hash=$1 WHERE id=$2`, string(hash), userID)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
