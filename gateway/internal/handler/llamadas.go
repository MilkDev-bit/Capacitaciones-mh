package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/config"
	"Prueba-Go/gateway/internal/hub"
	mw "Prueba-Go/gateway/internal/middleware"
	mensajespb "Prueba-Go/gen/mensajes"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegistrarPerdidas conecta el gestor de llamadas con el registro de llamadas
// perdidas. Se hace por inyección para que el paquete hub no dependa de los
// clientes gRPC del gateway.
func (h *LlamadasHandler) RegistrarPerdidas() {
	h.call.OnPerdida = h.avisarLlamadaPerdida
}

// LlamadasHandler expone lo que la señalización necesita del mundo HTTP:
// emitir el token con el que el cliente entra a la sala de Jitsi.
type LlamadasHandler struct {
	c    *clients.Clients
	cfg  *config.Config
	hub  *hub.Hub
	call *hub.GestorLlamadas
}

func NewLlamadasHandler(c *clients.Clients, cfg *config.Config, h *hub.Hub, g *hub.GestorLlamadas) *LlamadasHandler {
	return &LlamadasHandler{c: c, cfg: cfg, hub: h, call: g}
}

// jitsiClaims son los claims que espera el módulo `token_verification` de
// Prosody. La estructura no es negociable: viene definida por Jitsi.
//
//	aud/iss  → identifican la aplicación (JITSI_APP_ID)
//	sub      → dominio del servidor Jitsi
//	room     → sala concreta; es lo que impide reutilizar un token en otra sala
//	context  → datos del usuario que Jitsi muestra en la conferencia
type jitsiClaims struct {
	Room    string       `json:"room"`
	Context jitsiContext `json:"context"`
	jwt.RegisteredClaims
}

type jitsiContext struct {
	User jitsiUser `json:"user"`
}

type jitsiUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Moderator string `json:"moderator"`
}

// POST /api/llamadas/token
//
// Devuelve el token y la configuración que el frontend necesita para abrir
// Jitsi. El token se emite POR SALA: sin él, el servidor self-hosted rechaza
// la conexión, así que adivinar el nombre de sala no sirve de nada.
func (h *LlamadasHandler) Token(ctx *gin.Context) {
	userID := ctx.GetString(mw.CtxUserID)
	userName := ctx.GetString(mw.CtxUserName)

	var body struct {
		CallID string `json:"call_id" binding:"required"`
		Sala   string `json:"sala"    binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "faltan datos de la llamada"})
		return
	}

	// La comprobación decisiva: el gestor confirma que este usuario fue
	// invitado a ESA llamada y que la sala corresponde. Sin ella, el endpoint
	// sería una máquina de firmar tokens para cualquier sala que se pidiera.
	if !h.call.PuedeEntrar(body.CallID, body.Sala, userID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "no participas en esta llamada"})
		return
	}

	if h.cfg.JitsiAppSecret == "" {
		// Sin secreto no hay JWT posible. Se responde explícito en vez de
		// firmar con cadena vacía, que produciría un token que Prosody
		// rechaza con un error indescifrable en el navegador.
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "las videollamadas no están configuradas en el servidor",
		})
		return
	}

	ahora := time.Now()
	claims := jitsiClaims{
		Room: body.Sala,
		Context: jitsiContext{User: jitsiUser{
			ID:   userID,
			Name: userName,
			// Todos entran como moderadores: en una llamada entre dos
			// compañeros no hay jerarquía que imponer, y marcar a uno solo
			// haría que la sala muriera si es el primero en colgar.
			Moderator: "true",
		}},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.cfg.JitsiAppID,
			Audience:  jwt.ClaimStrings{h.cfg.JitsiAppID},
			Subject:   h.cfg.JitsiSubject(),
			IssuedAt:  jwt.NewNumericDate(ahora),
			NotBefore: jwt.NewNumericDate(ahora.Add(-30 * time.Second)),
			// Vida corta: el token solo tiene que durar lo que tarda el
			// cliente en conectarse. Una vez dentro, la sesión la mantiene
			// Jitsi, no el JWT.
			ExpiresAt: jwt.NewNumericDate(ahora.Add(4 * time.Hour)),
		},
	}

	firmado, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(h.cfg.JitsiAppSecret))
	if err != nil {
		slog.Error("llamadas: firmar token de Jitsi", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo preparar la llamada"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token":   firmado,
		"dominio": h.cfg.JitsiDomain,
		"sala":    body.Sala,
	})
}

// Iniciar arranca una llamada tras comprobar que está permitida.
//
// La autorización se delega en mensajes-service enviando un mensaje real de
// "llamada iniciada" ANTES de timbrar. No es un rodeo: ese servicio ya es el
// dueño de la regla "solo puedes contactar a gente de tus capacitaciones", y
// reimplementarla aquí crearía dos fuentes de verdad que se desincronizarían.
// Como efecto secundario deseable, la llamada queda registrada en el hilo con
// la misma marca de tiempo que verá el destinatario.
func (h *LlamadasHandler) Iniciar(emisorID, emisorName, peerID, peerName string, isGroup bool) {
	ctx, cancel := contextoCorto()
	defer cancel()

	if _, err := h.c.Mensajes.SendMensaje(ctx, &mensajespb.SendMensajeRequest{
		EmisorId:   emisorID,
		EmisorName: emisorName,
		ReceptorId: peerID,
		Contenido:  "📞 Videollamada",
		IsGroup:    isGroup,
	}); err != nil {
		// PermissionDenied llega aquí cuando el destinatario no comparte
		// capacitación. Se traduce a un evento que el cliente ya sabe pintar.
		motivo := "no se pudo iniciar la llamada"
		if st, ok := status.FromError(err); ok && st.Code() == codes.PermissionDenied {
			motivo = st.Message()
		}
		h.hub.Broadcast(emisorID, hub.Event{
			Type: hub.EvError,
			Call: &hub.CallPayload{PeerID: peerID, PeerName: peerName, IsGroup: isGroup, Motivo: motivo},
		})
		return
	}

	destinos, err := h.destinatarios(ctx, peerID, isGroup)
	if err != nil {
		slog.Error("llamadas: resolver destinatarios", "peer", peerID, "error", err)
		h.hub.Broadcast(emisorID, hub.Event{
			Type: hub.EvError,
			Call: &hub.CallPayload{PeerID: peerID, Motivo: "no se pudo iniciar la llamada"},
		})
		return
	}

	h.call.Iniciar(emisorID, emisorName, peerID, peerName, isGroup, destinos)
}

// destinatarios resuelve a quién hay que hacer sonar: en 1 a 1 es el propio
// peer; en grupo, todos sus miembros.
func (h *LlamadasHandler) destinatarios(ctx context.Context, peerID string, isGroup bool) ([]string, error) {
	if !isGroup {
		return []string{peerID}, nil
	}
	resp, err := h.c.Mensajes.GetGroupMembers(ctx,
		&mensajespb.GetGroupMembersRequest{GrupoId: peerID})
	if err != nil {
		return nil, err
	}
	return resp.UserIds, nil
}

// avisarLlamadaPerdida deja constancia en el hilo de la conversación.
//
// Es lo que convierte una llamada sin respuesta en algo que el destinatario
// puede ver cuando vuelva, en lugar de un timbre que se perdió en el aire.
func (h *LlamadasHandler) avisarLlamadaPerdida(l *hub.Llamada) {
	if l == nil || l.EmisorID == "" || l.PeerID == "" {
		return
	}
	ctx, cancel := contextoCorto()
	defer cancel()

	texto := "📞 Llamada perdida"
	if l.IsGroup {
		texto = "📞 Llamada de grupo sin respuesta"
	}

	_, err := h.c.Mensajes.SendMensaje(ctx, &mensajespb.SendMensajeRequest{
		EmisorId:   l.EmisorID,
		EmisorName: l.EmisorName,
		ReceptorId: l.PeerID,
		Contenido:  texto,
		IsGroup:    l.IsGroup,
	})
	if err != nil {
		slog.Warn("llamadas: no se pudo registrar la llamada perdida",
			"call_id", l.ID, "error", err)
	}

	// La constancia en el hilo solo se ve si el destinatario entra a esa
	// conversación. La campana es lo que se lo dice desde cualquier pantalla.
	//
	// Solo en llamadas directas: en las de grupo, PeerID es el ID del grupo, y
	// notificaciones.user_id tiene una FK contra users(id). Escribir ahí el
	// grupo no daría un aviso raro, daría un error de integridad. Repartir el
	// aviso entre los miembros exigiría resolver la membresía aquí; el mensaje
	// que ya quedó en el hilo cubre ese caso.
	if l.IsGroup {
		return
	}
	quien := l.EmisorName
	if quien == "" {
		quien = "Alguien"
	}
	notificar(h.c, aviso{
		UserID:  l.PeerID,
		Tipo:    TipoLlamadaPerdida,
		Titulo:  "Llamada perdida de " + quien,
		Mensaje: "No alcanzaste a responder la videollamada.",
		Enlace:  "/usuario/mensajes/" + l.EmisorID,
		Ventana: ventanaConversacion,
	})
}

// contextoCorto acota las llamadas gRPC que se hacen fuera de una petición
// HTTP (el registro de llamada perdida ocurre en una goroutine disparada por
// un temporizador, sin contexto de petición del que colgar).
func contextoCorto() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
