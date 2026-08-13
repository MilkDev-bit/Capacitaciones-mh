package hub

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Este archivo implementa la señalización de llamadas: el mecanismo que
// convierte "comparte este código de sala" en "tu teléfono suena".
//
// El modelo es el de WhatsApp/Instagram y tiene tres piezas:
//
//  1. El emisor manda `call_offer`. El servidor crea una llamada en estado
//     "timbrando" y reenvía el evento a las conexiones de los destinatarios.
//  2. El destinatario responde `call_accept` o `call_reject`. Al aceptar,
//     ambos extremos reciben `call_accepted` con el nombre de sala, y recién
//     entonces el cliente abre Jitsi.
//  3. Si nadie contesta en RingTimeout, el servidor cierra la llamada por su
//     cuenta y avisa a todos con `call_timeout`.
//
// El nombre de sala lo genera el SERVIDOR, no el cliente. Es la diferencia
// entre un sistema donde entrar depende de conocer un identificador y uno
// donde entrar depende de haber sido invitado: un nombre derivado de los IDs
// de usuario (como el `Capacitaciones-mh-<uuid>-<uuid>` anterior) lo puede
// reconstruir cualquiera que vea esos IDs en una respuesta de la API.
//
// El estado vive en memoria a propósito. Una llamada dura segundos y no
// sobrevive a un reinicio del gateway en ningún caso: persistirla añadiría
// una dependencia sin comprar nada. Lo que sí se persiste es la constancia de
// la llamada perdida, y eso va por el hilo de mensajes normal.

const (
	// RingTimeout es cuánto suena antes de rendirse. 30 s es el estándar de
	// facto en apps de mensajería: suficiente para llegar al teléfono, corto
	// para no dejar a nadie escuchando un tono muerto.
	RingTimeout = 30 * time.Second

	// MaxLlamadasPorUsuario limita las llamadas salientes simultáneas de un
	// mismo usuario. Sin este tope, un cliente comprometido podría abrir
	// miles de llamadas y hacer sonar a toda la plataforma.
	MaxLlamadasPorUsuario = 3
)

// Estados de una llamada.
const (
	EstadoTimbrando = "timbrando"
	EstadoEnCurso   = "en_curso"
)

// Tipos de evento de señalización que viajan por el WebSocket.
const (
	EvOffer     = "call_offer"     // servidor → destinatario: te están llamando
	EvRinging   = "call_ringing"   // servidor → emisor: el destinatario ya recibió el timbre
	EvAccepted  = "call_accepted"  // servidor → ambos: entren a la sala
	EvRejected  = "call_rejected"  // servidor → emisor: colgaron
	EvCancelled = "call_cancelled" // servidor → destinatario: el emisor desistió
	EvTimeout   = "call_timeout"   // servidor → ambos: nadie contestó
	EvEnded     = "call_ended"     // servidor → resto: la llamada terminó
	EvBusy      = "call_busy"      // servidor → emisor: el destinatario ya está en otra
	EvError     = "call_error"     // servidor → emisor: no se pudo iniciar
)

// CallPayload es la parte de Event que describe una llamada.
type CallPayload struct {
	ID          string   `json:"id"`
	Sala        string   `json:"sala,omitempty"`
	EmisorID    string   `json:"emisor_id"`
	EmisorName  string   `json:"emisor_name"`
	PeerID      string   `json:"peer_id"`             // conversación: usuario o grupo
	PeerName    string   `json:"peer_name,omitempty"` // nombre a mostrar en la UI
	IsGroup     bool     `json:"is_group"`
	Destinos    []string `json:"-"` // uso interno; nunca se serializa al cliente
	Motivo      string   `json:"motivo,omitempty"`
	SegundosMax int      `json:"segundos_max,omitempty"`
}

// Llamada es el estado en memoria de una llamada activa.
type Llamada struct {
	ID         string
	Sala       string
	EmisorID   string
	EmisorName string
	PeerID     string
	PeerName   string
	IsGroup    bool
	// Destinos son los usuarios a los que se hizo sonar el timbre.
	Destinos []string
	// Participantes son quienes ya aceptaron (incluye al emisor).
	Participantes map[string]bool
	Estado        string
	CreadaEn      time.Time

	temporizador *time.Timer
}

// Notificador es lo único que el gestor necesita del transporte: poder
// entregar un evento y saber si alguien está en línea.
//
// Se declara como interfaz y no se usa *Hub directamente para poder ejercitar
// la máquina de estados (timbre, timeout, ocupado, grupo) con un doble en los
// tests, sin levantar conexiones WebSocket reales.
type Notificador interface {
	Broadcast(userID string, ev Event)
	EstaConectado(userID string) bool
}

// GestorLlamadas coordina las llamadas activas del gateway.
type GestorLlamadas struct {
	hub Notificador

	mu         sync.Mutex
	llamadas   map[string]*Llamada // id de llamada → estado
	porUsuario map[string]int      // emisor → salientes en curso

	// OnPerdida se invoca cuando una llamada termina sin que nadie contestara.
	// El handler la usa para dejar la constancia en el hilo de mensajes; se
	// inyecta como función para que este paquete no dependa de los clientes
	// gRPC del gateway.
	OnPerdida func(l *Llamada)
}

// NewGestorLlamadas crea un gestor asociado a un hub.
func NewGestorLlamadas(h Notificador) *GestorLlamadas {
	return &GestorLlamadas{
		hub:        h,
		llamadas:   make(map[string]*Llamada),
		porUsuario: make(map[string]int),
	}
}

// Iniciar crea una llamada y hace sonar a los destinos conectados.
//
// destinos debe venir ya autorizado por el llamador: este paquete no sabe de
// inscripciones ni de cursos, solo de conexiones.
func (g *GestorLlamadas) Iniciar(emisorID, emisorName, peerID, peerName string, isGroup bool, destinos []string) {
	// Se descarta al propio emisor de la lista de timbres: en un grupo, el
	// emisor también es miembro y hacerse sonar a sí mismo sería absurdo.
	var reales []string
	for _, d := range destinos {
		if d != "" && d != emisorID {
			reales = append(reales, d)
		}
	}
	if len(reales) == 0 {
		g.hub.Broadcast(emisorID, Event{
			Type: EvError,
			Call: &CallPayload{PeerID: peerID, Motivo: "no hay nadie a quien llamar"},
		})
		return
	}

	// Solo tiene sentido timbrar a quien tiene una pestaña abierta. Si no hay
	// ninguna conexión, se informa de inmediato en vez de dejar sonar 30 s
	// contra el vacío.
	var conectados []string
	for _, d := range reales {
		if g.hub.EstaConectado(d) {
			conectados = append(conectados, d)
		}
	}
	if len(conectados) == 0 {
		l := &Llamada{
			EmisorID: emisorID, EmisorName: emisorName,
			PeerID: peerID, PeerName: peerName, IsGroup: isGroup,
			Destinos: reales, CreadaEn: time.Now(),
		}
		g.hub.Broadcast(emisorID, Event{
			Type: EvTimeout,
			Call: &CallPayload{PeerID: peerID, PeerName: peerName, IsGroup: isGroup,
				Motivo: "no disponible"},
		})
		if g.OnPerdida != nil {
			go g.OnPerdida(l)
		}
		return
	}

	g.mu.Lock()
	if g.porUsuario[emisorID] >= MaxLlamadasPorUsuario {
		g.mu.Unlock()
		g.hub.Broadcast(emisorID, Event{
			Type: EvError,
			Call: &CallPayload{PeerID: peerID, Motivo: "tienes demasiadas llamadas abiertas"},
		})
		return
	}

	// En 1 a 1, si el destinatario ya está timbrando o hablando en otra
	// llamada se responde ocupado en lugar de encimar dos timbres.
	if !isGroup {
		for _, l := range g.llamadas {
			if l.participa(peerID) {
				g.mu.Unlock()
				g.hub.Broadcast(emisorID, Event{
					Type: EvBusy,
					Call: &CallPayload{PeerID: peerID, PeerName: peerName, Motivo: "ocupado"},
				})
				return
			}
		}
	}

	l := &Llamada{
		ID:            nuevoID("call"),
		Sala:          nuevoID("mh"),
		EmisorID:      emisorID,
		EmisorName:    emisorName,
		PeerID:        peerID,
		PeerName:      peerName,
		IsGroup:       isGroup,
		Destinos:      conectados,
		Participantes: map[string]bool{emisorID: true},
		Estado:        EstadoTimbrando,
		CreadaEn:      time.Now(),
	}
	l.temporizador = time.AfterFunc(RingTimeout, func() { g.expirar(l.ID) })

	g.llamadas[l.ID] = l
	g.porUsuario[emisorID]++
	g.mu.Unlock()

	// El emisor recibe la sala desde el principio para poder precalentar la
	// conexión; los destinatarios NO la reciben hasta aceptar.
	g.hub.Broadcast(emisorID, Event{
		Type: EvRinging,
		Call: &CallPayload{
			ID: l.ID, Sala: l.Sala, EmisorID: emisorID, EmisorName: emisorName,
			PeerID: peerID, PeerName: peerName, IsGroup: isGroup,
			SegundosMax: int(RingTimeout / time.Second),
		},
	})

	oferta := Event{
		Type: EvOffer,
		Call: &CallPayload{
			ID: l.ID, EmisorID: emisorID, EmisorName: emisorName,
			PeerID: peerID, PeerName: peerName, IsGroup: isGroup,
			SegundosMax: int(RingTimeout / time.Second),
		},
	}
	for _, d := range conectados {
		g.hub.Broadcast(d, oferta)
	}

	slog.Info("llamada iniciada", "call_id", l.ID, "emisor", emisorID,
		"peer", peerID, "grupo", isGroup, "destinos", len(conectados))
}

// Aceptar mete a userID en la llamada y le entrega el nombre de sala.
func (g *GestorLlamadas) Aceptar(callID, userID, userName string) {
	g.mu.Lock()
	l, ok := g.llamadas[callID]
	if !ok || !l.invitado(userID) {
		g.mu.Unlock()
		g.hub.Broadcast(userID, Event{
			Type: EvError,
			Call: &CallPayload{ID: callID, Motivo: "la llamada ya no está disponible"},
		})
		return
	}

	primeraAceptacion := l.Estado == EstadoTimbrando
	if primeraAceptacion {
		l.Estado = EstadoEnCurso
		// Se detiene el temporizador de timbre: a partir de aquí la llamada
		// vive mientras haya gente dentro, no contra reloj.
		if l.temporizador != nil {
			l.temporizador.Stop()
		}
	}
	l.Participantes[userID] = true
	// Todo lo que se necesita fuera del candado se copia dentro: leer campos
	// de l después de soltar el mutex sería una carrera con otra goroutine que
	// esté terminando la misma llamada.
	sala, emisorID, emisorName := l.Sala, l.EmisorID, l.EmisorName
	peerID, peerName, isGroup := l.PeerID, l.PeerName, l.IsGroup
	destinos := append([]string(nil), l.Destinos...)
	g.mu.Unlock()

	entrar := Event{
		Type: EvAccepted,
		Call: &CallPayload{
			ID: callID, Sala: sala, EmisorID: emisorID, EmisorName: emisorName,
			PeerID: peerID, PeerName: peerName, IsGroup: isGroup,
		},
	}
	g.hub.Broadcast(userID, entrar)

	// El emisor solo entra en la primera aceptación; en un grupo, la segunda
	// persona que acepta no debe reabrirle la ventana.
	if primeraAceptacion {
		g.hub.Broadcast(emisorID, entrar)

		// Al resto del grupo se les retira el timbre y se les deja el aviso de
		// que la llamada sigue viva por si quieren sumarse.
		if isGroup {
			enCurso := Event{
				Type: EvEnded, // el cliente lo interpreta como "deja de sonar"
				Call: &CallPayload{ID: callID, PeerID: peerID, PeerName: peerName,
					IsGroup: true, Motivo: "en_curso"},
			}
			for _, d := range destinos {
				if d != userID {
					g.hub.Broadcast(d, enCurso)
				}
			}
		}
	}
}

// Rechazar cierra la llamada (1 a 1) o retira a un miembro (grupo).
func (g *GestorLlamadas) Rechazar(callID, userID string) {
	g.mu.Lock()
	l, ok := g.llamadas[callID]
	if !ok || !l.invitado(userID) {
		g.mu.Unlock()
		return
	}
	// En grupo, que uno cuelgue no cancela el timbre de los demás.
	if l.IsGroup && l.Estado == EstadoTimbrando {
		l.quitarDestino(userID)
		quedan := len(l.Destinos)
		emisorID, peerID := l.EmisorID, l.PeerID
		g.mu.Unlock()
		if quedan == 0 {
			g.terminar(callID, EvRejected, "rechazada")
			return
		}
		g.hub.Broadcast(emisorID, Event{
			Type: EvRejected,
			Call: &CallPayload{ID: callID, PeerID: peerID, EmisorID: userID, Motivo: "un miembro rechazó"},
		})
		return
	}
	g.mu.Unlock()
	g.terminar(callID, EvRejected, "rechazada")
}

// Cancelar la ejecuta el emisor cuando cuelga antes de que contesten.
func (g *GestorLlamadas) Cancelar(callID, userID string) {
	g.mu.Lock()
	l, ok := g.llamadas[callID]
	if !ok || l.EmisorID != userID {
		g.mu.Unlock()
		return
	}
	sinRespuesta := l.Estado == EstadoTimbrando
	g.mu.Unlock()

	if sinRespuesta {
		// Colgar antes de que contesten sigue siendo una llamada perdida para
		// el otro lado: se registra igual que un timeout.
		g.terminarConPerdida(callID, EvCancelled, "cancelada")
		return
	}
	g.terminar(callID, EvEnded, "finalizada")
}

// Salir retira a un participante; si no queda nadie, cierra la llamada.
func (g *GestorLlamadas) Salir(callID, userID string) {
	g.mu.Lock()
	l, ok := g.llamadas[callID]
	if !ok {
		g.mu.Unlock()
		return
	}
	delete(l.Participantes, userID)
	vacia := len(l.Participantes) == 0
	g.mu.Unlock()

	if vacia {
		g.terminar(callID, EvEnded, "finalizada")
	}
}

// LimpiarUsuario cancela las llamadas en las que participa un usuario que
// acaba de desconectarse. Sin esto, cerrar la pestaña dejaría al otro extremo
// escuchando el tono hasta agotar el timeout.
func (g *GestorLlamadas) LimpiarUsuario(userID string) {
	g.mu.Lock()
	var afectadas []string
	for id, l := range g.llamadas {
		if l.EmisorID == userID || l.Participantes[userID] {
			afectadas = append(afectadas, id)
		}
	}
	g.mu.Unlock()

	for _, id := range afectadas {
		// Solo se cierra si al usuario no le queda ninguna otra conexión: un
		// móvil y un escritorio abiertos son el mismo usuario.
		if g.hub.EstaConectado(userID) {
			continue
		}
		g.Salir(id, userID)
		g.Cancelar(id, userID)
	}
}

// expirar cierra una llamada que nadie contestó dentro de RingTimeout.
func (g *GestorLlamadas) expirar(callID string) {
	g.terminarConPerdida(callID, EvTimeout, "sin respuesta")
}

func (g *GestorLlamadas) terminarConPerdida(callID, tipo, motivo string) {
	l := g.terminar(callID, tipo, motivo)
	if l != nil && g.OnPerdida != nil && len(l.Participantes) <= 1 {
		go g.OnPerdida(l)
	}
}

// terminar retira la llamada del registro y notifica a todos los implicados.
// Devuelve la llamada retirada, o nil si ya no existía.
func (g *GestorLlamadas) terminar(callID, tipo, motivo string) *Llamada {
	g.mu.Lock()
	l, ok := g.llamadas[callID]
	if !ok {
		g.mu.Unlock()
		return nil
	}
	delete(g.llamadas, callID)
	if l.temporizador != nil {
		l.temporizador.Stop()
	}
	if n := g.porUsuario[l.EmisorID]; n > 1 {
		g.porUsuario[l.EmisorID] = n - 1
	} else {
		delete(g.porUsuario, l.EmisorID)
	}
	g.mu.Unlock()

	ev := Event{
		Type: tipo,
		Call: &CallPayload{
			ID: l.ID, EmisorID: l.EmisorID, EmisorName: l.EmisorName,
			PeerID: l.PeerID, PeerName: l.PeerName, IsGroup: l.IsGroup, Motivo: motivo,
		},
	}
	avisados := map[string]bool{}
	for _, d := range append(append([]string(nil), l.Destinos...), l.EmisorID) {
		if d == "" || avisados[d] {
			continue
		}
		avisados[d] = true
		g.hub.Broadcast(d, ev)
	}
	for p := range l.Participantes {
		if !avisados[p] {
			avisados[p] = true
			g.hub.Broadcast(p, ev)
		}
	}

	slog.Info("llamada terminada", "call_id", l.ID, "motivo", motivo,
		"duracion_s", int(time.Since(l.CreadaEn).Seconds()))
	return l
}

// ── Helpers de Llamada ───────────────────────────────────────────────────────

func (l *Llamada) invitado(userID string) bool {
	if l.Participantes[userID] {
		return true
	}
	for _, d := range l.Destinos {
		if d == userID {
			return true
		}
	}
	return false
}

func (l *Llamada) participa(userID string) bool {
	return l.EmisorID == userID || l.invitado(userID)
}

func (l *Llamada) quitarDestino(userID string) {
	out := l.Destinos[:0]
	for _, d := range l.Destinos {
		if d != userID {
			out = append(out, d)
		}
	}
	l.Destinos = out
}

// PuedeEntrar valida que el usuario tiene derecho a un token para esa sala.
//
// Es la comprobación que respalda al endpoint de tokens: se exige que la
// llamada exista, que la sala coincida y que el usuario haya sido invitado o
// ya sea participante. Firmar un token sin esto convertiría el endpoint en un
// generador de accesos para cualquier sala que alguien quisiera nombrar.
func (g *GestorLlamadas) PuedeEntrar(callID, sala, userID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	l, ok := g.llamadas[callID]
	if !ok || l.Sala != sala {
		return false
	}
	return l.EmisorID == userID || l.invitado(userID)
}

// nuevoID genera un identificador opaco de 128 bits en hexadecimal.
//
// Se usa crypto/rand y no un contador ni un hash de los IDs de usuario: el
// nombre de sala es, junto al JWT, lo que delimita quién entra a la
// conferencia, así que tiene que ser impredecible. Hexadecimal en minúsculas
// porque Prosody normaliza el nombre de sala a minúsculas y cualquier
// codificación con mayúsculas dejaría de coincidir con el claim `room` del
// token.
func nuevoID(prefijo string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand no falla en la práctica; si lo hiciera, un identificador
		// basado en el reloj es preferible a devolver cadena vacía y dejar la
		// sala sin nombre.
		return fmt.Sprintf("%s-%d", prefijo, time.Now().UnixNano())
	}
	return prefijo + "-" + hex.EncodeToString(b)
}
