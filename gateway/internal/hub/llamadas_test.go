package hub

import (
	"sync"
	"testing"
	"time"
)

// notificadorFalso registra los eventos entregados a cada usuario y permite
// declarar quién está conectado, sin abrir un solo WebSocket.
type notificadorFalso struct {
	mu         sync.Mutex
	eventos    map[string][]Event
	conectados map[string]bool
}

func nuevoFalso(conectados ...string) *notificadorFalso {
	n := &notificadorFalso{
		eventos:    map[string][]Event{},
		conectados: map[string]bool{},
	}
	for _, c := range conectados {
		n.conectados[c] = true
	}
	return n
}

func (n *notificadorFalso) Broadcast(userID string, ev Event) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.eventos[userID] = append(n.eventos[userID], ev)
}

func (n *notificadorFalso) EstaConectado(userID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.conectados[userID]
}

// tipos devuelve los tipos de evento recibidos por un usuario.
func (n *notificadorFalso) tipos(userID string) []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	var out []string
	for _, e := range n.eventos[userID] {
		out = append(out, e.Type)
	}
	return out
}

// ultimo devuelve el último evento recibido por un usuario.
func (n *notificadorFalso) ultimo(userID string) *Event {
	n.mu.Lock()
	defer n.mu.Unlock()
	evs := n.eventos[userID]
	if len(evs) == 0 {
		return nil
	}
	return &evs[len(evs)-1]
}

func (n *notificadorFalso) recibio(userID, tipo string) bool {
	for _, t := range n.tipos(userID) {
		if t == tipo {
			return true
		}
	}
	return false
}

// idDeLlamada extrae el ID que el emisor recibió en call_ringing.
func idDeLlamada(t *testing.T, n *notificadorFalso, emisor string) string {
	t.Helper()
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, e := range n.eventos[emisor] {
		if e.Type == EvRinging && e.Call != nil {
			return e.Call.ID
		}
	}
	t.Fatalf("el emisor %q nunca recibió %s", emisor, EvRinging)
	return ""
}

func TestLlamadaDirectaAceptadaEntregaLaSalaAAmbos(t *testing.T) {
	n := nuevoFalso("ana", "beto")
	g := NewGestorLlamadas(n)

	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})

	if !n.recibio("beto", EvOffer) {
		t.Fatalf("Beto no recibió el timbre; recibió %v", n.tipos("beto"))
	}
	// La oferta no debe llevar la sala: el destinatario solo la obtiene al
	// aceptar, de modo que rechazar no le deja una credencial de entrada.
	oferta := n.ultimo("beto")
	if oferta.Call.Sala != "" {
		t.Errorf("la oferta filtró la sala %q antes de aceptar", oferta.Call.Sala)
	}

	callID := idDeLlamada(t, n, "ana")
	g.Aceptar(callID, "beto", "Beto")

	for _, u := range []string{"ana", "beto"} {
		if !n.recibio(u, EvAccepted) {
			t.Fatalf("%s no recibió %s; recibió %v", u, EvAccepted, n.tipos(u))
		}
	}
	salaAna := ""
	salaBeto := ""
	n.mu.Lock()
	for _, e := range n.eventos["ana"] {
		if e.Type == EvAccepted {
			salaAna = e.Call.Sala
		}
	}
	for _, e := range n.eventos["beto"] {
		if e.Type == EvAccepted {
			salaBeto = e.Call.Sala
		}
	}
	n.mu.Unlock()

	if salaAna == "" || salaAna != salaBeto {
		t.Fatalf("las salas no coinciden: ana=%q beto=%q", salaAna, salaBeto)
	}
}

func TestSalaEsImpredecibleYNoDerivaDeLosIDs(t *testing.T) {
	vistas := map[string]bool{}
	for i := 0; i < 50; i++ {
		n := nuevoFalso("ana", "beto")
		g := NewGestorLlamadas(n)
		g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})

		sala := n.ultimo("ana").Call.Sala
		if sala == "" {
			t.Fatal("la sala vino vacía")
		}
		if vistas[sala] {
			t.Fatalf("sala repetida entre llamadas distintas: %q", sala)
		}
		vistas[sala] = true

		// Regresión del diseño anterior, donde la sala era
		// "Capacitaciones-mh-<idA>-<idB>" y cualquiera que viera esos IDs
		// podía reconstruirla y colarse en la conferencia.
		for _, filtrado := range []string{"ana", "beto"} {
			if contiene(sala, filtrado) {
				t.Fatalf("la sala %q contiene el ID de usuario %q", sala, filtrado)
			}
		}
	}
}

func TestDestinatarioDesconectadoNoDejaSonar(t *testing.T) {
	n := nuevoFalso("ana") // Beto sin conexión
	g := NewGestorLlamadas(n)

	var perdidas int
	var mu sync.Mutex
	g.OnPerdida = func(*Llamada) { mu.Lock(); perdidas++; mu.Unlock() }

	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})

	if !n.recibio("ana", EvTimeout) {
		t.Fatalf("Ana debió recibir %s de inmediato; recibió %v", EvTimeout, n.tipos("ana"))
	}
	if n.recibio("beto", EvOffer) {
		t.Error("se timbró a un usuario desconectado")
	}

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if perdidas != 1 {
		t.Errorf("llamadas perdidas registradas = %d, se esperaba 1", perdidas)
	}
}

func TestRechazoCierraLaLlamadaYNadieObtieneToken(t *testing.T) {
	n := nuevoFalso("ana", "beto")
	g := NewGestorLlamadas(n)
	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	callID := idDeLlamada(t, n, "ana")
	sala := n.ultimo("ana").Call.Sala

	g.Rechazar(callID, "beto")

	if !n.recibio("ana", EvRejected) {
		t.Fatalf("Ana no recibió %s; recibió %v", EvRejected, n.tipos("ana"))
	}
	// Tras cerrarse, ni el emisor puede pedir token: si el endpoint siguiera
	// firmando, la sala quedaría accesible indefinidamente.
	if g.PuedeEntrar(callID, sala, "ana") {
		t.Error("PuedeEntrar sigue autorizando tras el rechazo")
	}
	if g.PuedeEntrar(callID, sala, "beto") {
		t.Error("quien rechazó sigue pudiendo entrar")
	}
}

func TestPuedeEntrarRechazaSalaAjenaYExtranos(t *testing.T) {
	n := nuevoFalso("ana", "beto", "carla")
	g := NewGestorLlamadas(n)
	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	callID := idDeLlamada(t, n, "ana")
	sala := n.ultimo("ana").Call.Sala

	if !g.PuedeEntrar(callID, sala, "ana") {
		t.Error("el emisor debería poder entrar")
	}
	if !g.PuedeEntrar(callID, sala, "beto") {
		t.Error("el invitado debería poder entrar")
	}
	if g.PuedeEntrar(callID, sala, "carla") {
		t.Error("un tercero ajeno obtuvo permiso de entrada")
	}
	if g.PuedeEntrar(callID, "mh-otra-sala", "ana") {
		t.Error("se autorizó una sala distinta a la de la llamada")
	}
	if g.PuedeEntrar("call-inexistente", sala, "ana") {
		t.Error("se autorizó una llamada inexistente")
	}
}

func TestDestinatarioOcupadoRecibeBusy(t *testing.T) {
	n := nuevoFalso("ana", "beto", "carla")
	g := NewGestorLlamadas(n)

	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	g.Iniciar("carla", "Carla", "beto", "Beto", false, []string{"beto"})

	if !n.recibio("carla", EvBusy) {
		t.Fatalf("Carla debió recibir %s; recibió %v", EvBusy, n.tipos("carla"))
	}
}

func TestTimeoutCierraLaLlamadaYRegistraPerdida(t *testing.T) {
	n := nuevoFalso("ana", "beto")
	g := NewGestorLlamadas(n)

	hecho := make(chan struct{}, 1)
	g.OnPerdida = func(*Llamada) { hecho <- struct{}{} }

	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	callID := idDeLlamada(t, n, "ana")

	// Se dispara el vencimiento directamente en vez de esperar 30 s reales.
	g.expirar(callID)

	for _, u := range []string{"ana", "beto"} {
		if !n.recibio(u, EvTimeout) {
			t.Errorf("%s no recibió %s; recibió %v", u, EvTimeout, n.tipos(u))
		}
	}
	select {
	case <-hecho:
	case <-time.After(time.Second):
		t.Fatal("no se registró la llamada perdida")
	}
}

func TestLlamadaDeGrupoTimbraATodosYElPrimeroEnAceptarAbreLaSala(t *testing.T) {
	n := nuevoFalso("ana", "beto", "carla")
	g := NewGestorLlamadas(n)

	g.Iniciar("ana", "Ana", "grupo-1", "Equipo A", true, []string{"ana", "beto", "carla"})

	for _, u := range []string{"beto", "carla"} {
		if !n.recibio(u, EvOffer) {
			t.Fatalf("%s no recibió el timbre de grupo; recibió %v", u, n.tipos(u))
		}
	}
	// El emisor no debe hacerse sonar a sí mismo aunque sea miembro.
	if n.recibio("ana", EvOffer) {
		t.Error("el emisor recibió su propio timbre")
	}

	callID := idDeLlamada(t, n, "ana")
	g.Aceptar(callID, "beto", "Beto")

	if !n.recibio("beto", EvAccepted) || !n.recibio("ana", EvAccepted) {
		t.Fatal("la primera aceptación no metió a emisor y receptor en la sala")
	}
	// A Carla se le retira el timbre, pero la llamada sigue viva por si quiere
	// sumarse: el motivo se lo dice.
	if !n.recibio("carla", EvEnded) {
		t.Fatalf("a Carla no se le retiró el timbre; recibió %v", n.tipos("carla"))
	}
	if u := n.ultimo("carla"); u.Call.Motivo != "en_curso" {
		t.Errorf("motivo para Carla = %q, se esperaba \"en_curso\"", u.Call.Motivo)
	}

	// Carla todavía puede entrar: sigue invitada.
	sala := ""
	n.mu.Lock()
	for _, e := range n.eventos["ana"] {
		if e.Type == EvAccepted {
			sala = e.Call.Sala
		}
	}
	n.mu.Unlock()
	if !g.PuedeEntrar(callID, sala, "carla") {
		t.Error("Carla debería poder unirse a la llamada de grupo en curso")
	}
}

func TestRechazoEnGrupoNoCortaAlResto(t *testing.T) {
	n := nuevoFalso("ana", "beto", "carla")
	g := NewGestorLlamadas(n)
	g.Iniciar("ana", "Ana", "grupo-1", "Equipo A", true, []string{"beto", "carla"})
	callID := idDeLlamada(t, n, "ana")

	g.Rechazar(callID, "beto")

	if n.recibio("carla", EvEnded) || n.recibio("carla", EvRejected) {
		t.Error("el rechazo de Beto cortó el timbre de Carla")
	}
	// Carla sigue pudiendo aceptar.
	g.Aceptar(callID, "carla", "Carla")
	if !n.recibio("carla", EvAccepted) {
		t.Fatalf("Carla no pudo aceptar; recibió %v", n.tipos("carla"))
	}
}

func TestLimiteDeLlamadasSimultaneasPorUsuario(t *testing.T) {
	n := nuevoFalso("ana", "b1", "b2", "b3", "b4")
	g := NewGestorLlamadas(n)

	for _, d := range []string{"b1", "b2", "b3"} {
		g.Iniciar("ana", "Ana", d, d, false, []string{d})
	}
	g.Iniciar("ana", "Ana", "b4", "b4", false, []string{"b4"})

	if !n.recibio("ana", EvError) {
		t.Fatalf("no se aplicó el tope de %d llamadas; ana recibió %v",
			MaxLlamadasPorUsuario, n.tipos("ana"))
	}
	if n.recibio("b4", EvOffer) {
		t.Error("se timbró pese a superar el tope")
	}
}

func TestCancelarAntesDeContestarCuentaComoPerdida(t *testing.T) {
	n := nuevoFalso("ana", "beto")
	g := NewGestorLlamadas(n)

	hecho := make(chan struct{}, 1)
	g.OnPerdida = func(*Llamada) { hecho <- struct{}{} }

	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	callID := idDeLlamada(t, n, "ana")
	g.Cancelar(callID, "ana")

	if !n.recibio("beto", EvCancelled) {
		t.Fatalf("Beto no supo que colgaron; recibió %v", n.tipos("beto"))
	}
	select {
	case <-hecho:
	case <-time.After(time.Second):
		t.Fatal("colgar antes de contestar no se registró como llamada perdida")
	}
}

func TestCancelarTrasContestarNoCuentaComoPerdida(t *testing.T) {
	n := nuevoFalso("ana", "beto")
	g := NewGestorLlamadas(n)

	perdidas := make(chan struct{}, 1)
	g.OnPerdida = func(*Llamada) { perdidas <- struct{}{} }

	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	callID := idDeLlamada(t, n, "ana")
	g.Aceptar(callID, "beto", "Beto")
	g.Cancelar(callID, "ana")

	if !n.recibio("beto", EvEnded) {
		t.Errorf("Beto no recibió %s; recibió %v", EvEnded, n.tipos("beto"))
	}
	select {
	case <-perdidas:
		t.Fatal("una llamada contestada se registró como perdida")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestSoloElEmisorPuedeCancelar(t *testing.T) {
	n := nuevoFalso("ana", "beto")
	g := NewGestorLlamadas(n)
	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	callID := idDeLlamada(t, n, "ana")
	sala := n.ultimo("ana").Call.Sala

	g.Cancelar(callID, "beto") // Beto no es el emisor

	if !g.PuedeEntrar(callID, sala, "ana") {
		t.Error("un no-emisor consiguió cancelar la llamada")
	}
}

func TestAceptarUnaLlamadaAjenaNoConcedeAcceso(t *testing.T) {
	n := nuevoFalso("ana", "beto", "carla")
	g := NewGestorLlamadas(n)
	g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
	callID := idDeLlamada(t, n, "ana")
	sala := n.ultimo("ana").Call.Sala

	g.Aceptar(callID, "carla", "Carla") // Carla nunca fue invitada

	if n.recibio("carla", EvAccepted) {
		t.Fatal("un extraño consiguió entrar a la llamada")
	}
	if !n.recibio("carla", EvError) {
		t.Errorf("no se rechazó explícitamente; carla recibió %v", n.tipos("carla"))
	}
	if g.PuedeEntrar(callID, sala, "carla") {
		t.Error("el extraño quedó autorizado para pedir token")
	}
}

func TestAccesoConcurrenteNoRompeElEstado(t *testing.T) {
	n := nuevoFalso("ana", "beto")
	g := NewGestorLlamadas(n)

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Iniciar("ana", "Ana", "beto", "Beto", false, []string{"beto"})
		}()
	}
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.LimpiarUsuario("ana")
		}()
	}
	wg.Wait()
	// Sin panic ni carrera detectada por -race, la prueba pasa.
}

func contiene(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
