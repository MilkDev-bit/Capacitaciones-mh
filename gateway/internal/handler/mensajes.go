package handler

import (
"net/http"
"time"

"Prueba-Go/gateway/internal/clients"
mw "Prueba-Go/gateway/internal/middleware"
mensajespb "Prueba-Go/gen/mensajes"

"github.com/gin-gonic/gin"
)

// MensajesHandler delega al microservicio mensajes via gRPC.
type MensajesHandler struct {
client mensajespb.MensajesServiceClient
}

func NewMensajesHandler(svc *clients.Clients) *MensajesHandler {
return &MensajesHandler{client: svc.Mensajes}
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

type conversacion struct {
PeerID      string    `json:"peer_id"`
PeerName    string    `json:"peer_name"`
LastMessage string    `json:"last_message"`
LastTime    time.Time `json:"last_time"`
UnreadCount int32     `json:"unread_count"`
}

convs := make([]conversacion, 0, len(resp.Conversaciones))
for _, cv := range resp.Conversaciones {
t, _ := time.Parse(time.RFC3339, cv.LastTime)
convs = append(convs, conversacion{
PeerID:      cv.PeerId,
PeerName:    cv.PeerName,
LastMessage: cv.LastMessage,
LastTime:    t,
UnreadCount: cv.UnreadCount,
})
}
c.JSON(http.StatusOK, convs)
}

// GetMensajes devuelve la conversacion entre el usuario autenticado y un peer.
func (h *MensajesHandler) GetMensajes(c *gin.Context) {
userID := c.GetString(mw.CtxUserID)
peerID := c.Param("peer_id")

resp, err := h.client.GetMensajes(c.Request.Context(), &mensajespb.GetMensajesRequest{
UserId: userID,
PeerId: peerID,
})
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "error cargando mensajes"})
return
}

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

msgs := make([]mensajeDTO, 0, len(resp.Mensajes))
for _, m := range resp.Mensajes {
t, _ := time.Parse(time.RFC3339, m.CreatedAt)
msgs = append(msgs, mensajeDTO{
ID:           m.Id,
EmisorID:     m.EmisorId,
EmisorName:   m.EmisorName,
ReceptorID:   m.ReceptorId,
ReceptorName: m.ReceptorName,
Contenido:    m.Contenido,
Leido:        m.Leido,
CreatedAt:    t,
})
}
c.JSON(http.StatusOK, msgs)
}

// SendMensaje envia un mensaje al peer indicado.
func (h *MensajesHandler) SendMensaje(c *gin.Context) {
userID := c.GetString(mw.CtxUserID)
userName := c.GetString(mw.CtxUserName)
peerID := c.Param("peer_id")

var body struct {
Contenido string `json:"contenido"`
PeerName  string `json:"peer_name"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "cuerpo invalido"})
return
}

resp, err := h.client.SendMensaje(c.Request.Context(), &mensajespb.SendMensajeRequest{
EmisorId:     userID,
EmisorName:   userName,
ReceptorId:   peerID,
ReceptorName: body.PeerName,
Contenido:    body.Contenido,
})
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

t, _ := time.Parse(time.RFC3339, resp.CreatedAt)
c.JSON(http.StatusCreated, gin.H{
"id":            resp.Id,
"emisor_id":     resp.EmisorId,
"emisor_name":   resp.EmisorName,
"receptor_id":   resp.ReceptorId,
"receptor_name": resp.ReceptorName,
"contenido":     resp.Contenido,
"leido":         resp.Leido,
"created_at":    t,
})
}
