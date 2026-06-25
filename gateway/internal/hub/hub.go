package hub

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// MsgPayload define la estructura de un mensaje entrante via WebSocket.
type MsgPayload struct {
	ID             string `json:"id"`
	EmisorID       string `json:"emisor_id"`
	EmisorName     string `json:"emisor_name"`
	ReceptorID     string `json:"receptor_id"`
	ReceptorName   string `json:"receptor_name"`
	Contenido      string `json:"contenido"`
	CreatedAt      string `json:"created_at"`
	Leido          bool   `json:"leido"`
	AttachmentUrl  string `json:"attachment_url,omitempty"`
	AttachmentType string `json:"attachment_type,omitempty"`
	IsGroup        bool   `json:"is_group"`
}

// Event es el mensaje JSON que el servidor envía al cliente WebSocket.
type Event struct {
	Type     string      `json:"type"`
	Msg      *MsgPayload `json:"msg,omitempty"`
	PeerID   string      `json:"peer_id,omitempty"`
	PeerName string      `json:"peer_name,omitempty"`
	Count    int32       `json:"count,omitempty"`
}

// Client representa una conexión WebSocket activa.
type Client struct {
	Hub    *Hub
	UserID string
	Name   string
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub mantiene el conjunto de clientes activos y broadcast seguro.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
}

// New devuelve un Hub inicializado.
func New() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
	}
}

// Register añade un cliente al hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.UserID] == nil {
		h.clients[c.UserID] = make(map[*Client]struct{})
	}
	h.clients[c.UserID][c] = struct{}{}
}

// Unregister elimina un cliente del hub y cierra su canal Send.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[c.UserID]; ok {
		if _, exists := set[c]; exists {
			delete(set, c)
			close(c.Send)
			if len(set) == 0 {
				delete(h.clients, c.UserID)
			}
		}
	}
}

// Broadcast envía un evento a todas las conexiones de un usuario.
func (h *Hub) Broadcast(userID string, ev Event) {
	data, err := json.Marshal(ev)
	if err != nil {
		return
	}
	h.mu.RLock()
	set := h.clients[userID]
	h.mu.RUnlock()

	for c := range set {
		select {
		case c.Send <- data:
		default:
			// Canal lleno: descartar para no bloquear
		}
	}
}
