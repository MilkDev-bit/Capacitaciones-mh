package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	mw "Prueba-Go/gateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

// MensajesHandler gestiona los mensajes directos entre usuarios.
// Se conecta directamente a la BD (misma instancia que los microservicios).
type MensajesHandler struct {
	db *sql.DB
}

func NewMensajesHandler(db *sql.DB) *MensajesHandler {
	return &MensajesHandler{db: db}
}

// conversacion representa el resumen de una conversación en la bandeja de entrada.
type conversacion struct {
	PeerID      string    `json:"peer_id"`
	PeerName    string    `json:"peer_name"`
	LastMessage string    `json:"last_message"`
	LastTime    time.Time `json:"last_time"`
	UnreadCount int       `json:"unread_count"`
}

// mensajeDTO es la representación de un mensaje individual.
type mensajeDTO struct {
	ID           string    `json:"id"`
	EmisorID     string    `json:"emisor_id"`
	EmisorName   string    `json:"emisor_name"`
	ReceptorID   string    `json:"receptor_id"`
	ReceptorName string    `json:"receptor_name"`
	Contenido    string    `json:"contenido"`
	Leido        bool      `json:"leido"`
	CreatedAt    time.Time `json:"created_at"`
}

// NoLeidos devuelve el total de mensajes no leídos del usuario autenticado.
func (h *MensajesHandler) NoLeidos(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, gin.H{"count": 0})
		return
	}
	userID := c.GetString(mw.CtxUserID)

	var count int
	if err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT COUNT(*) FROM mensajes WHERE receptor_id = $1::uuid AND leido = FALSE`,
		userID,
	).Scan(&count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error contando mensajes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// ListConversaciones devuelve todas las conversaciones activas del usuario autenticado,
// ordenadas por el mensaje más reciente. Incluye el conteo de no leídos por conversación.
func (h *MensajesHandler) ListConversaciones(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, []conversacion{})
		return
	}
	userID := c.GetString(mw.CtxUserID)

	rows, err := h.db.QueryContext(c.Request.Context(), `
		WITH ranked AS (
			SELECT
				CASE WHEN emisor_id = $1::uuid THEN receptor_id    ELSE emisor_id    END AS peer_id,
				CASE WHEN emisor_id = $1::uuid THEN receptor_name  ELSE emisor_name  END AS peer_name,
				contenido,
				created_at,
				ROW_NUMBER() OVER (
					PARTITION BY
						LEAST   (emisor_id::text, receptor_id::text),
						GREATEST(emisor_id::text, receptor_id::text)
					ORDER BY created_at DESC
				) AS rn
			FROM mensajes
			WHERE emisor_id = $1::uuid OR receptor_id = $1::uuid
		)
		SELECT
			r.peer_id,
			r.peer_name,
			r.contenido,
			r.created_at,
			(
				SELECT COUNT(*)
				FROM mensajes
				WHERE receptor_id = $1::uuid
				  AND emisor_id   = r.peer_id
				  AND leido       = FALSE
			) AS unread_count
		FROM ranked r
		WHERE r.rn = 1
		ORDER BY r.created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error cargando conversaciones"})
		return
	}
	defer rows.Close()

	convs := make([]conversacion, 0)
	for rows.Next() {
		var conv conversacion
		if err := rows.Scan(&conv.PeerID, &conv.PeerName, &conv.LastMessage, &conv.LastTime, &conv.UnreadCount); err != nil {
			continue
		}
		convs = append(convs, conv)
	}
	c.JSON(http.StatusOK, convs)
}

// GetMensajes devuelve todos los mensajes entre el usuario autenticado y un peer.
// También marca como leídos los mensajes recibidos del peer.
func (h *MensajesHandler) GetMensajes(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, []mensajeDTO{})
		return
	}
	userID := c.GetString(mw.CtxUserID)
	peerID := c.Param("peer_id")

	// Marcar como leídos los mensajes que nos envió el peer.
	_, _ = h.db.ExecContext(c.Request.Context(),
		`UPDATE mensajes SET leido = TRUE
		 WHERE receptor_id = $1::uuid AND emisor_id = $2::uuid AND leido = FALSE`,
		userID, peerID,
	)

	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at
		FROM mensajes
		WHERE (emisor_id = $1::uuid AND receptor_id = $2::uuid)
		   OR (emisor_id = $2::uuid AND receptor_id = $1::uuid)
		ORDER BY created_at ASC
	`, userID, peerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error cargando mensajes"})
		return
	}
	defer rows.Close()

	msgs := make([]mensajeDTO, 0)
	for rows.Next() {
		var m mensajeDTO
		if err := rows.Scan(
			&m.ID, &m.EmisorID, &m.EmisorName,
			&m.ReceptorID, &m.ReceptorName,
			&m.Contenido, &m.Leido, &m.CreatedAt,
		); err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	c.JSON(http.StatusOK, msgs)
}

// SendMensaje envía un mensaje de texto al usuario indicado por peer_id.
func (h *MensajesHandler) SendMensaje(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "mensajería no disponible"})
		return
	}
	userID := c.GetString(mw.CtxUserID)
	userName := c.GetString(mw.CtxUserName)
	peerID := c.Param("peer_id")

	if userID == peerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no puedes enviarte mensajes a ti mismo"})
		return
	}

	var body struct {
		Contenido string `json:"contenido"`
		PeerName  string `json:"peer_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cuerpo inválido"})
		return
	}

	body.Contenido = strings.TrimSpace(body.Contenido)
	if n := utf8.RuneCountInString(body.Contenido); n < 1 || n > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el mensaje debe tener entre 1 y 5000 caracteres"})
		return
	}
	body.PeerName = strings.TrimSpace(body.PeerName)

	var m mensajeDTO
	err := h.db.QueryRowContext(c.Request.Context(), `
		INSERT INTO mensajes (emisor_id, emisor_name, receptor_id, receptor_name, contenido)
		VALUES ($1::uuid, $2, $3::uuid, $4, $5)
		RETURNING id, emisor_id, emisor_name, receptor_id, receptor_name, contenido, leido, created_at
	`, userID, userName, peerID, body.PeerName, body.Contenido).Scan(
		&m.ID, &m.EmisorID, &m.EmisorName,
		&m.ReceptorID, &m.ReceptorName,
		&m.Contenido, &m.Leido, &m.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error enviando mensaje"})
		return
	}
	c.JSON(http.StatusCreated, m)
}
