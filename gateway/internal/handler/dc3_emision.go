package handler

import (
	"context"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Presupuesto de tiempo y concurrencia de la emisión DC-3
//
// Emitir una constancia es, con diferencia, la operación más pesada del
// Gateway. Encadena, en orden:
//
//	gRPC a cursos-service para reunir los datos
//	dos descargas de logo desde R2
//	generación del .docx sobre la plantilla
//	conversión a PDF en Gotenberg, que levanta Chromium
//	subida del PDF a R2
//	gRPC a cursos-service para registrar el folio
//
// Este archivo existe porque esa operación no puede regirse por las mismas
// reglas que una notificación.
// ─────────────────────────────────────────────────────────────────────────────

// tiempoEmision es el presupuesto para toda la cadena de arriba.
//
// Antes se usaba contextoCorto(), que son 10 segundos. Se escribió para
// notificaciones y señalización de llamadas —gRPC baratos contra un servicio
// vecino— y acabó gobernando una conversión en Chromium. El cliente HTTP de
// Gotenberg declara 90 segundos, pero no servían de nada: el contexto cancelaba
// antes, y la constancia fallaba con "no se pudo emitir automáticamente" de
// forma intermitente, que es la peor manera de tener un agujero.
//
// El caso lento es real y está documentado en el propio código del conversor:
// la primera conversión tras arrancar Gotenberg carga Chromium desde cero.
const tiempoEmision = 2 * time.Minute

// emisionesSimultaneas es cuántas constancias pueden estar convirtiéndose a la
// vez.
//
// La emisión se dispara sola al completar un curso, y se dispara una vez por
// alumno. Una empresa que cierra una cohorte de doscientas personas —que es
// exactamente para lo que se vende esta plataforma— lanzaría doscientas
// goroutines, y cada una pide a Gotenberg que levante un Chromium. El
// contenedor se queda sin memoria y caen TODAS, incluidas las que habrían
// salido bien de haber esperado su turno.
//
// Tres es conservador a propósito: Gotenberg corre en su propio contenedor con
// la memoria que le dé Railway, y es preferible que la constancia número
// doscientos tarde unos minutos a que ninguna se emita. Si algún día sobra
// capacidad, esto se sube; mientras tanto la cola es invisible para el alumno,
// porque la emisión ya era asíncrona y termina con una notificación.
const emisionesSimultaneas = 3

// turnoDeEmision limita cuántas emisiones corren a la vez.
//
// Es un canal y no un sync.Mutex porque hace falta un contador, no exclusión, y
// porque permite rendirse si el contexto se cancela mientras se espera turno.
var turnoDeEmision = make(chan struct{}, emisionesSimultaneas)

// contextoEmision devuelve el contexto de una emisión en segundo plano.
//
// Cuelga de context.Background() y no del de gin: la emisión automática vive en
// una goroutine que sobrevive al handler que la lanzó, y gin recicla su contexto
// en cuanto el handler retorna.
func contextoEmision() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), tiempoEmision)
}

// contextoEmisionDe es la variante para la emisión que ocurre dentro de una
// petición HTTP, cuando el alumno acaba de mandar sus datos.
//
// Cuelga del contexto de la petición para que cerrar la pestaña cancele el
// trabajo y devuelva el turno, en vez de seguir ocupando un hueco de Gotenberg
// para un PDF que ya no espera nadie.
func contextoEmisionDe(padre context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(padre, tiempoEmision)
}

// esperarTurno bloquea hasta que haya un hueco libre o el contexto se agote.
//
// Devuelve la función que libera el hueco. El error solo aparece si se agotó el
// presupuesto esperando, y en ese caso no hay hueco que liberar.
func esperarTurno(ctx context.Context) (func(), error) {
	select {
	case turnoDeEmision <- struct{}{}:
		return func() { <-turnoDeEmision }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
