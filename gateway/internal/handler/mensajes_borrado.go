package handler

import (
	"net/http"

	"Prueba-Go/gateway/internal/hub"
	mw "Prueba-Go/gateway/internal/middleware"
	mensajespb "Prueba-Go/gen/mensajes"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// Borrado de mensajes y conversaciones
//
// Quién puede borrar qué se decide en mensajes-service, no aquí: el user_id
// sale del token y el gateway solo lo transporta. Lo que sí es cosa del
// gateway es avisar por WebSocket, porque es quien tiene el hub.
// ─────────────────────────────────────────────────────────────────────────────

// EliminarMensaje elimina un mensaje.
//
//	DELETE /api/mensajes/mensaje/:msg_id            → solo para mí
//	DELETE /api/mensajes/mensaje/:msg_id?todos=true → para todos
func (h *MensajesHandler) EliminarMensaje(c *gin.Context) {
	userID := c.GetString(mw.CtxUserID)
	msgID := c.Param("msg_id")
	paraTodos := c.Query("todos") == "true"

	// La respuesta trae a quién avisar. El servicio ya tenía la fila en la mano
	// para autorizar el borrado, así que devolverla sale gratis; buscarla aquí
	// después obligaría a recorrer conversaciones enteras.
	destino, err := h.client.EliminarMensaje(c.Request.Context(), &mensajespb.EliminarMensajeRequest{
		MsgId:     msgID,
		UserId:    userID,
		ParaTodos: paraTodos,
	})
	if err != nil {
		grpcToHTTP(c, err)
		return
	}

	// Solo el borrado para todos se anuncia. El borrado para mí no cambia nada
	// en la pantalla del otro, y mandárselo le haría parpadear un mensaje que
	// sigue teniendo.
	if paraTodos {
		evento := hub.Event{Type: "message_deleted", MsgID: msgID, PeerID: userID}
		if destino.IsGroup {
			// En grupo el hilo se identifica por el grupo, no por quien borró:
			// el cliente necesita saber QUÉ conversación refrescar.
			evento.PeerID = destino.ReceptorId
			h.difundirAlGrupo(c, destino.ReceptorId, userID, evento)
		} else {
			h.hub.Broadcast(destino.ReceptorId, evento)
			h.hub.Broadcast(destino.EmisorId, evento)
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// EliminarConversacion oculta una conversación entera para quien lo pide.
//
//	DELETE /api/mensajes/conversacion/:peer_id[?is_group=true]
//
// No avisa por WebSocket: es una acción privada. El otro conserva su copia y no
// tiene por qué enterarse de que alguien limpió su bandeja.
func (h *MensajesHandler) EliminarConversacion(c *gin.Context) {
	userID := c.GetString(mw.CtxUserID)

	if _, err := h.client.EliminarConversacion(c.Request.Context(), &mensajespb.EliminarConversacionRequest{
		UserId:  userID,
		PeerId:  c.Param("peer_id"),
		IsGroup: c.Query("is_group") == "true",
	}); err != nil {
		grpcToHTTP(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// difundirAlGrupo manda el evento a todos los miembros menos al que lo provocó,
// que ya actualizó su pantalla al recibir la respuesta HTTP.
func (h *MensajesHandler) difundirAlGrupo(c *gin.Context, grupoID, exceptoID string, e hub.Event) {
	miembros, err := h.client.GetGroupMembers(c.Request.Context(),
		&mensajespb.GetGroupMembersRequest{GrupoId: grupoID})
	if err != nil {
		return
	}
	for _, id := range miembros.UserIds {
		if id != exceptoID {
			h.hub.Broadcast(id, e)
		}
	}
}
