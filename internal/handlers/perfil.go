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
		SELECT id, name, email, role, COALESCE(bio,''), COALESCE(avatar_url,''), COALESCE(phone,''), created_at
		FROM users WHERE id=$1`, userID,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Bio, &u.AvatarURL, &u.Phone, &u.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func UpdatePerfil(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var body struct {
		Name     string `json:"name"`
		Bio      string `json:"bio"`
		Phone    string `json:"phone"`
		Password string `json:"password"` // opcional: nueva contraseña
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Name != "" {
		db.DB.Exec(`UPDATE users SET name=$1 WHERE id=$2`, body.Name, userID)
	}
	db.DB.Exec(`UPDATE users SET bio=$1, phone=$2 WHERE id=$3`, body.Bio, body.Phone, userID)

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
