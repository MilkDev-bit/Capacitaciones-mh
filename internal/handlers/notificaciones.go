package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"Prueba-Go/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type Notificacion struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Tipo      string    `json:"tipo"`
	Titulo    string    `json:"titulo"`
	Mensaje   string    `json:"mensaje"`
	Leida     bool      `json:"leida"`
	Enlace    string    `json:"enlace"`
	CreatedAt time.Time `json:"created_at"`
}

func ListNotificaciones(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, user_id, tipo, titulo, mensaje, leida, COALESCE(enlace, ''), created_at
		FROM notificaciones
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		slog.Error("ListNotificaciones db err", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al obtener notificaciones"})
		return
	}
	defer rows.Close()

	var notificaciones []Notificacion
	for rows.Next() {
		var n Notificacion
		if err := rows.Scan(&n.ID, &n.UserID, &n.Tipo, &n.Titulo, &n.Mensaje, &n.Leida, &n.Enlace, &n.CreatedAt); err != nil {
			slog.Error("ListNotificaciones scan err", "err", err)
			continue
		}
		notificaciones = append(notificaciones, n)
	}

	if notificaciones == nil {
		notificaciones = []Notificacion{}
	}

	c.JSON(http.StatusOK, notificaciones)
}

func MarcarNotificacionesLeidas(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato inválido"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	// For simplicity, handle multiple updates in a loop or using ANY
	query := `UPDATE notificaciones SET leida = true WHERE user_id = $1 AND id = ANY($2)`
	_, err := db.DB.Exec(query, userID, pq.Array(req.IDs))
	if err != nil {
		slog.Error("MarcarNotificacionesLeidas db err", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar notificaciones"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Función utilitaria para crear notificaciones desde otros handlers
func CrearNotificacion(userID, tipo, titulo, mensaje, enlace string) error {
	_, err := db.DB.Exec(`
		INSERT INTO notificaciones (user_id, tipo, titulo, mensaje, enlace)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, tipo, titulo, mensaje, enlace)
	if err != nil {
		slog.Error("CrearNotificacion db err", "err", err)
	}
	return err
}
