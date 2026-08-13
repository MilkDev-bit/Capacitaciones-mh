package handler

// Emisión de notificaciones in-app.
//
// La campana del header se alimenta de la tabla `notificaciones`, que pertenece
// al usuarios-service. Este fichero concentra el único camino por el que el
// Gateway escribe en ella, para que añadir un aviso nuevo sea una línea y no
// una decisión de arquitectura repetida en cada handler.
//
// Dos reglas que conviene no romper:
//
//  1. Notificar NUNCA debe hacer fallar la acción que la originó. Una compra se
//     completa aunque el usuarios-service esté caído; el aviso se pierde y se
//     registra en el log, pero el pago no se revierte por eso. Por eso todo va
//     en una goroutine con su propio contexto y el error solo se loguea.
//
//  2. El contexto de gin NO puede viajar a esa goroutine: gin lo recicla en
//     cuanto el handler retorna. Se usa contextoCorto() (definido en llamadas.go),
//     que cuelga de context.Background().

import (
	"log/slog"
	"time"

	"Prueba-Go/gateway/internal/clients"
	usuariospb "Prueba-Go/gen/usuarios"
)

// Tipos de notificación. Deben coincidir con `tiposNotificacion` del
// usuarios-service y con TIPOS_POR_PERFIL de NotificationBell.vue: los tres
// puntos tienen que hablar del mismo vocabulario o el aviso se crea y nunca se
// muestra.
const (
	TipoCompra         = "compra"          // el comprador recibe el acuse de su pago
	TipoInscripcion    = "inscripcion"     // el alumno entra a una capacitación
	TipoNuevoAlumno    = "nuevo_alumno"    // el instructor recibe un alta en su curso
	TipoMensaje        = "mensaje"         // mensaje directo
	TipoLlamadaPerdida = "llamada_perdida" // videollamada sin respuesta
	TipoForoRespuesta  = "foro_respuesta"  // alguien respondió al usuario en el foro
)

// Ventanas de deduplicación por familia de evento.
//
// Sin ellas, veinte mensajes seguidos de la misma persona dejan veinte campanas
// idénticas. El valor es el tiempo durante el cual un aviso no leído idéntico
// absorbe a los siguientes.
const (
	ventanaConversacion = 6 * time.Hour // mensajes y llamadas: muy repetitivos
	ventanaForo         = 30 * time.Minute
	ventanaCompra       = 24 * time.Hour // el webhook y la pantalla de éxito llegan por separado
)

// aviso es un evento a notificar. Ventana cero desactiva la deduplicación.
type aviso struct {
	UserID  string
	Tipo    string
	Titulo  string
	Mensaje string
	Enlace  string
	Ventana time.Duration
}

// notificar emite los avisos en segundo plano.
//
// Descarta los que no tienen destinatario, para que quien llama pueda pasar
// campos que a veces vienen vacíos (un curso sin instructor, un comentario de
// primer nivel sin padre) sin tener que comprobarlo antes.
func notificar(c *clients.Clients, avisos ...aviso) {
	if c == nil || c.Usuarios == nil {
		return
	}
	for _, a := range avisos {
		if a.UserID == "" || a.Titulo == "" {
			continue
		}
		go enviarAviso(c, a)
	}
}

func enviarAviso(c *clients.Clients, a aviso) {
	ctx, cancel := contextoCorto()
	defer cancel()

	_, err := c.Usuarios.CreateNotificacion(ctx, &usuariospb.CreateNotificacionRequest{
		UserId:           a.UserID,
		Tipo:             a.Tipo,
		Titulo:           a.Titulo,
		Mensaje:          a.Mensaje,
		Enlace:           a.Enlace,
		DedupeVentanaSeg: int32(a.Ventana.Seconds()),
	})
	if err != nil {
		slog.Warn("notificar: no se pudo crear la notificación",
			"tipo", a.Tipo, "user_id", a.UserID, "error", err)
	}
}

// notificarSalvoA descarta el aviso dirigido a quien originó la acción.
//
// Es el caso más común de ruido: responderte a ti mismo en el foro o mandarte
// un mensaje a tu propio hilo no debería encender tu campana.
func notificarSalvoA(c *clients.Clients, actorID string, avisos ...aviso) {
	filtrados := make([]aviso, 0, len(avisos))
	for _, a := range avisos {
		if a.UserID == actorID {
			continue
		}
		filtrados = append(filtrados, a)
	}
	notificar(c, filtrados...)
}

// recorta deja el mensaje en algo que quepa en el desplegable de la campana.
// La columna es TEXT, así que el límite es de presentación, no de esquema.
func recorta(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
