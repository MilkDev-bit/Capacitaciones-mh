package handler

import (
	"net/http"
	"strconv"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/hub"
	mw "Prueba-Go/gateway/internal/middleware"
	mensajespb "Prueba-Go/gen/mensajes"
	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/gin-gonic/gin"
)

// MensajesHandler delega al microservicio mensajes via gRPC.
type MensajesHandler struct {
	client         mensajespb.MensajesServiceClient
	usuariosClient usuariospb.UsuariosServiceClient
	hub            *hub.Hub
}

func NewMensajesHandler(svc *clients.Clients, h *hub.Hub) *MensajesHandler {
	return &MensajesHandler{
		client:         svc.Mensajes,
		usuariosClient: svc.Usuarios,
		hub:            h,
	}
}

// NoLeidos devuelve el total de mensajes no leidos del usuario autenticado.
func (h *MensajesHandler) NoLeidos(c *gin.Context) {
	userID := c.GetString(mw.CtxUserID)
	resp, err := h.client.NoLeidos(c.Request.Context(), &mensajespb.NoLeidosRequest{UserId: userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error contando mensajes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": resp.Count})
}

// ListConversaciones devuelve las conversaciones activas del usuario autenticado.
func (h *MensajesHandler) ListConversaciones(c *gin.Context) {
	userID := c.GetString(mw.CtxUserID)
	resp, err := h.client.ListConversaciones(c.Request.Context(), &mensajespb.ListConversacionesRequest{UserId: userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error cargando conversaciones"})
		return
	}

	type conversacionDTO struct {
		PeerID      string    `json:"peer_id"`
		PeerName    string    `json:"peer_name"`
		LastMessage string    `json:"last_message"`
		LastTime    time.Time `json:"last_time"`
		UnreadCount int32     `json:"unread_count"`
		AvatarURL   string    `json:"avatar_url"`
	}

	// Fetch avatars concurrently to minimize latency
	type avatarResult struct {
		peerID    string
		avatarURL string
	}

	n := len(resp.Conversaciones)
	ch := make(chan avatarResult, n)

	for _, cv := range resp.Conversaciones {
		go func(peerId string) {
			var avatarURL string
			peerProfile, err := h.usuariosClient.GetPublicPerfil(c.Request.Context(), &usuariospb.UserIDRequest{UserId: peerId})
			if err == nil && peerProfile != nil {
				avatarURL = peerProfile.AvatarUrl
			}
			ch <- avatarResult{peerID: peerId, avatarURL: avatarURL}
		}(cv.PeerId)
	}

	avatars := make(map[string]string)
	for i := 0; i < n; i++ {
		res := <-ch
		avatars[res.peerID] = res.avatarURL
	}

	convs := make([]conversacionDTO, 0, n)
	for _, cv := range resp.Conversaciones {
		t, _ := time.Parse("2006-01-02T15:04:05Z", cv.LastTime)
		convs = append(convs, conversacionDTO{
			PeerID:      cv.PeerId,
			PeerName:    cv.PeerName,
			LastMessage: cv.LastMessage,
			LastTime:    t,
			UnreadCount: cv.UnreadCount,
			AvatarURL:   avatars[cv.PeerId],
		})
	}
	c.JSON(http.StatusOK, convs)
}

// GetMensajes devuelve la conversacion entre el usuario autenticado y un peer.
// Soporta paginacion con ?limit=N&before_id=UUID
func (h *MensajesHandler) GetMensajes(c *gin.Context) {
userID := c.GetString(mw.CtxUserID)
peerID := c.Param("peer_id")

limitStr := c.DefaultQuery("limit", "50")
limit, err := strconv.Atoi(limitStr)
if err != nil || limit <= 0 || limit > 200 {
limit = 50
}
beforeID := c.Query("before_id")

resp, err := h.client.GetMensajes(c.Request.Context(), &mensajespb.GetMensajesRequest{
UserId:   userID,
PeerId:   peerID,
Limit:    int32(limit),
BeforeId: beforeID,
})
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "error cargando mensajes"})
return
}

// En la carga inicial (sin cursor) notificar al peer via WS que sus mensajes fueron leidos
if beforeID == "" {
h.hub.Broadcast(peerID, hub.Event{
Type:   "message_read",
PeerID: userID,
})
}

type mensajeDTO struct {
ID             string    `json:"id"`
EmisorID       string    `json:"emisor_id"`
EmisorName     string    `json:"emisor_name"`
ReceptorID     string    `json:"receptor_id"`
ReceptorName   string    `json:"receptor_name"`
Contenido      string    `json:"contenido"`
Leido          bool      `json:"leido"`
CreatedAt      time.Time `json:"created_at"`
AttachmentUrl  string    `json:"attachment_url,omitempty"`
AttachmentType string    `json:"attachment_type,omitempty"`
}

msgs := make([]mensajeDTO, 0, len(resp.Mensajes))
for _, m := range resp.Mensajes {
t, _ := time.Parse("2006-01-02T15:04:05Z", m.CreatedAt)
msgs = append(msgs, mensajeDTO{
ID:             m.Id,
EmisorID:       m.EmisorId,
EmisorName:     m.EmisorName,
ReceptorID:     m.ReceptorId,
ReceptorName:   m.ReceptorName,
Contenido:      m.Contenido,
Leido:          m.Leido,
CreatedAt:      t,
AttachmentUrl:  m.AttachmentUrl,
AttachmentType: m.AttachmentType,
})
}
c.JSON(http.StatusOK, gin.H{"mensajes": msgs, "has_more": resp.HasMore})
}

// SendMensaje envia un mensaje al peer indicado.
func (h *MensajesHandler) SendMensaje(c *gin.Context) {
userID := c.GetString(mw.CtxUserID)
userName := c.GetString(mw.CtxUserName)
peerID := c.Param("peer_id")

var body struct {
Contenido      string `json:"contenido"`
PeerName       string `json:"peer_name"`
AttachmentUrl  string `json:"attachment_url"`
AttachmentType string `json:"attachment_type"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "cuerpo invalido"})
return
}

resp, err := h.client.SendMensaje(c.Request.Context(), &mensajespb.SendMensajeRequest{
EmisorId:       userID,
EmisorName:     userName,
ReceptorId:     peerID,
ReceptorName:   body.PeerName,
Contenido:      body.Contenido,
AttachmentUrl:  body.AttachmentUrl,
AttachmentType: body.AttachmentType,
})
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

t, _ := time.Parse("2006-01-02T15:04:05Z", resp.CreatedAt)

// Notificar al receptor en tiempo real via WebSocket
h.hub.Broadcast(resp.ReceptorId, hub.Event{
Type: "new_message",
Msg: &hub.MsgPayload{
ID:             resp.Id,
EmisorID:       resp.EmisorId,
EmisorName:     resp.EmisorName,
ReceptorID:     resp.ReceptorId,
ReceptorName:   resp.ReceptorName,
Contenido:      resp.Contenido,
CreatedAt:      t.Format(time.RFC3339),
Leido:          resp.Leido,
AttachmentUrl:  resp.AttachmentUrl,
AttachmentType: resp.AttachmentType,
},
})

c.JSON(http.StatusCreated, gin.H{
"id":              resp.Id,
"emisor_id":       resp.EmisorId,
"emisor_name":     resp.EmisorName,
"receptor_id":     resp.ReceptorId,
"receptor_name":   resp.ReceptorName,
"contenido":       resp.Contenido,
"leido":           resp.Leido,
"created_at":      t,
"attachment_url":  resp.AttachmentUrl,
"attachment_type": resp.AttachmentType,
})
}

// MarcarLeido marca un mensaje individual como leido y notifica al emisor.
func (h *MensajesHandler) MarcarLeido(c *gin.Context) {
userID := c.GetString(mw.CtxUserID)
msgID := c.Param("msg_id")

resp, err := h.client.MarcarLeido(c.Request.Context(), &mensajespb.MarcarLeidoRequest{
MsgId:  msgID,
UserId: userID,
})
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "error marcando mensaje"})
return
}

if resp.EmisorId != "" {
h.hub.Broadcast(resp.EmisorId, hub.Event{
Type:   "message_read",
PeerID: userID,
})
}

c.JSON(http.StatusOK, gin.H{"ok": resp.Ok})
}
