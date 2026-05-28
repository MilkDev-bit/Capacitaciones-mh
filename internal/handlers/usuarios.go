package handlers

import (
	"log/slog"
	"net/http"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/models"

	"github.com/gin-gonic/gin"
)

type asignarRequest struct {
	UserID         string  `json:"user_id" binding:"required"`
	CapacitacionID *string `json:"capacitacion_id"`
	ExamenID       *string `json:"examen_id"`
}

func Asignar(c *gin.Context) {
	var req asignarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}
	if req.CapacitacionID == nil && req.ExamenID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere capacitacion_id o examen_id"})
		return
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

func DesAsignar(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec(`DELETE FROM asignaciones WHERE id=$1`, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListAsignaciones(c *gin.Context) {
	userID := c.Query("user_id")
	var rows interface{ Close() error }
	var err error

	if userID != "" {
		rows, err = db.DB.Query(`
			SELECT id, user_id, capacitacion_id, examen_id, assigned_at
			FROM asignaciones WHERE user_id=$1
		`, userID)
	} else {
		rows, err = db.DB.Query(`SELECT id, user_id, capacitacion_id, examen_id, assigned_at FROM asignaciones`)
	}
	if err != nil {
		slog.Error("ListAsignaciones", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	type rowScanner interface {
		Next() bool
		Scan(...interface{}) error
		Close() error
	}

	scanner := rows.(rowScanner)
	defer scanner.Close()

	result := []models.Asignacion{}
	for scanner.Next() {
		var a models.Asignacion
		scanner.Scan(&a.ID, &a.UserID, &a.CapacitacionID, &a.ExamenID, &a.AssignedAt)
		result = append(result, a)
	}
	c.JSON(http.StatusOK, result)
}

func ListUsers(c *gin.Context) {
	limit, offset, page := parsePagination(c)
	var total int
	db.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total)
	rows, err := db.DB.Query(`SELECT id, name, email, role, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		slog.Error("ListUsers", "error", err)
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
	c.JSON(http.StatusOK, gin.H{"data": result, "total": total, "page": page, "limit": limit})
}

// RevokeUserSessions incrementa token_version del usuario, invalidando todos sus JWTs activos.
// Solo accesible para admins.
func RevokeUserSessions(c *gin.Context) {
	userID := c.Param("id")
	res, err := db.DB.Exec(`UPDATE users SET token_version=token_version+1 WHERE id=$1`, userID)
	if err != nil {
		slog.Error("RevokeUserSessions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
