package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/hub"
	mw "Prueba-Go/gateway/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		for _, allowed := range config.C.AllowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// WsHandler gestiona conexiones WebSocket autenticadas.
type WsHandler struct {
	hub *hub.Hub
}

// NewWsHandler crea un WsHandler.
func NewWsHandler(h *hub.Hub) *WsHandler {
	return &WsHandler{hub: h}
}

// Handle actualiza la conexión HTTP a WebSocket, registra el cliente en el hub
// y lanza las goroutines de escritura y lectura.
func (wh *WsHandler) Handle(c *gin.Context) {
	userID := c.GetString(mw.CtxUserID)
	userName := c.GetString(mw.CtxUserName)
	if userID == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &hub.Client{
		Hub:    wh.hub,
		UserID: userID,
		Name:   userName,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
	wh.hub.Register(client)

	var cleanup sync.Once
	cleanupFn := func() {
		cleanup.Do(func() {
			wh.hub.Unregister(client)
			conn.Close()
		})
	}

	// Goroutine escritora: drena el canal Send y envía pings periódicos.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer func() {
			ticker.Stop()
			cleanupFn()
		}()
		for {
			select {
			case msg, ok := <-client.Send:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// Goroutine lectora principal: recibe mensajes del cliente.
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var ev struct {
			Type   string `json:"type"`
			PeerID string `json:"peer_id"`
		}
		if err := json.Unmarshal(raw, &ev); err != nil {
			log.Printf("error unmarshal: %v", err)
			continue
		}

		switch ev.Type {
		case "typing":
			if ev.PeerID != "" {
				wh.hub.Broadcast(ev.PeerID, hub.Event{
					Type:     "typing",
					PeerID:   userID,
					PeerName: userName,
				})
			}
		}
	}

	cleanupFn()
}
