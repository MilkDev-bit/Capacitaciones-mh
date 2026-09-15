package handler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// El semáforo de emisión es de los pocos sitios del Gateway donde un error
// bloquea de verdad: un turno que no se libera deja la cola parada para
// siempre, y las constancias dejan de emitirse sin un solo error en el log.

func TestEsperarTurnoLiberaElHueco(t *testing.T) {
	vaciarTurnos(t)

	liberar, err := esperarTurno(context.Background())
	if err != nil {
		t.Fatalf("no se consiguió turno estando libre: %v", err)
	}
	liberar()

	// Si liberar() no devolviera el hueco, tras agotar el aforo esta llamada
	// se quedaría colgada.
	if libres := emisionesSimultaneas - len(turnoDeEmision); libres != emisionesSimultaneas {
		t.Fatalf("quedaron %d huecos ocupados tras liberar", emisionesSimultaneas-libres)
	}
}

func TestEsperarTurnoLimitaLaConcurrencia(t *testing.T) {
	vaciarTurnos(t)

	// Se ocupa el aforo completo sin liberar.
	liberadores := make([]func(), 0, emisionesSimultaneas)
	for i := 0; i < emisionesSimultaneas; i++ {
		liberar, err := esperarTurno(context.Background())
		if err != nil {
			t.Fatalf("turno %d: %v", i, err)
		}
		liberadores = append(liberadores, liberar)
	}

	// El siguiente NO debe pasar. Es justo lo que protege a Gotenberg: la
	// constancia número cuatro espera en vez de pedir otro Chromium.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := esperarTurno(ctx); err == nil {
		t.Fatal("se concedió un turno por encima del aforo")
	}

	for _, l := range liberadores {
		l()
	}

	// Liberado el aforo, vuelve a conceder.
	liberar, err := esperarTurno(context.Background())
	if err != nil {
		t.Fatalf("no se recuperó el aforo tras liberar: %v", err)
	}
	liberar()
}

// Nunca deben correr más de `emisionesSimultaneas` a la vez, que es el número
// que decide si el contenedor de Gotenberg aguanta o se queda sin memoria.
func TestEsperarTurnoNuncaExcedeElAforo(t *testing.T) {
	vaciarTurnos(t)

	var enVuelo, maximo int64
	var wg sync.WaitGroup

	for i := 0; i < emisionesSimultaneas*10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			liberar, err := esperarTurno(context.Background())
			if err != nil {
				return
			}
			defer liberar()

			n := atomic.AddInt64(&enVuelo, 1)
			for {
				m := atomic.LoadInt64(&maximo)
				if n <= m || atomic.CompareAndSwapInt64(&maximo, m, n) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&enVuelo, -1)
		}()
	}
	wg.Wait()

	if maximo > int64(emisionesSimultaneas) {
		t.Fatalf("llegaron a correr %d emisiones a la vez; el aforo es %d", maximo, emisionesSimultaneas)
	}
}

// Un contexto ya cancelado no debe consumir turno: si lo hiciera, una tanda de
// peticiones abandonadas dejaría la cola llena de huecos que nadie libera.
func TestEsperarTurnoConContextoCanceladoNoConsumeHueco(t *testing.T) {
	vaciarTurnos(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// El select de esperarTurno puede elegir cualquiera de las dos ramas cuando
	// las dos están listas, así que se repite hasta observar el rechazo.
	var rechazado bool
	for i := 0; i < 100; i++ {
		liberar, err := esperarTurno(ctx)
		if err != nil {
			rechazado = true
			break
		}
		liberar()
	}
	if !rechazado {
		t.Fatal("un contexto cancelado nunca fue rechazado")
	}
	if len(turnoDeEmision) != 0 {
		t.Fatalf("quedaron %d huecos ocupados por turnos rechazados", len(turnoDeEmision))
	}
}

func TestPresupuestoDeEmisionCubreLaConversion(t *testing.T) {
	// El cliente HTTP de Gotenberg declara 90 segundos. Si el presupuesto fuera
	// menor, ese timeout no serviría de nada porque el contexto cancelaría
	// antes: es exactamente el fallo que tenía contextoCorto() con sus 10s.
	const timeoutGotenberg = 90 * time.Second
	if tiempoEmision <= timeoutGotenberg {
		t.Fatalf("el presupuesto de emisión (%v) no cubre el timeout de Gotenberg (%v)",
			tiempoEmision, timeoutGotenberg)
	}
}

// vaciarTurnos deja el semáforo limpio entre pruebas: es una variable de
// paquete y los tests comparten proceso.
func vaciarTurnos(t *testing.T) {
	t.Helper()
	for len(turnoDeEmision) > 0 {
		<-turnoDeEmision
	}
}
