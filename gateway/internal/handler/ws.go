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
	// call y llamadas soportan la señalización de videollamada. Van por el
	// mismo socket que los mensajes porque el timbre necesita exactamente lo
	// que el socket ya ofrece: entrega inmediata y saber quién está en línea.
	call     *hub.GestorLlamadas
	llamadas *LlamadasHandler
}

// NewWsHandler crea un WsHandler.
func NewWsHandler(h *hub.Hub, g *hub.GestorLlamadas, lh *LlamadasHandler) *WsHandler {
	return &WsHandler{hub: h, call: g, llamadas: lh}
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
			Type     string `json:"type"`
			PeerID   string `json:"peer_id"`
			PeerName string `json:"peer_name"`
			CallID   string `json:"call_id"`
			IsGroup  bool   `json:"is_group"`
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

		// ── Señalización de videollamada ──────────────────────────────────
		//
		// El cliente nunca elige el nombre de sala ni decide quién puede
		// entrar: solo declara intención ("llamar a X", "acepto"). Todo lo
		// demás lo resuelve el servidor, que es el único que puede hacerlo
		// sin confiar en el navegador.
		case "call_start":
			if ev.PeerID == "" || wh.llamadas == nil {
				continue
			}
			// Se lanza en goroutine para no bloquear el bucle de lectura: la
			// expansión del grupo y la comprobación de permisos hacen
			// llamadas gRPC, y un socket que deja de leer se cae por el
			// read deadline de 60 s.
			go wh.llamadas.Iniciar(userID, userName, ev.PeerID, ev.PeerName, ev.IsGroup)

		case "call_accept":
			if ev.CallID != "" && wh.call != nil {
				wh.call.Aceptar(ev.CallID, userID, userName)
			}

		case "call_reject":
			if ev.CallID != "" && wh.call != nil {
				wh.call.Rechazar(ev.CallID, userID)
			}

		case "call_cancel":
			if ev.CallID != "" && wh.call != nil {
				wh.call.Cancelar(ev.CallID, userID)
			}

		case "call_leave":
			if ev.CallID != "" && wh.call != nil {
				wh.call.Salir(ev.CallID, userID)
			}
		}
	}

	cleanupFn()
	// Al perder la conexión se cierran las llamadas del usuario. Sin esto,
	// cerrar la pestaña dejaría al otro extremo escuchando el tono hasta que
	// venciera el timeout de 30 s.
	if wh.call != nil {
		wh.call.LimpiarUsuario(userID)
	}
}
