package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	mensajespb "Prueba-Go/gen/mensajes"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─────────────────────────────────────────────────────────────────────────────
// Borrado de mensajes y conversaciones
//
// Dos operaciones distintas, y la diferencia importa:
//
//   "Para mí"     → lo oculta de MI vista. El otro conserva su copia intacta.
//                   Puede hacerlo cualquiera de los dos, sin límite de tiempo.
//   "Para todos"  → lo quita también al otro y deja la lápida. Solo el AUTOR,
//                   y solo dentro de la ventana de abajo.
//
// En ningún caso desaparece la fila de la base de datos.
// ─────────────────────────────────────────────────────────────────────────────

// ventanaBorrado es el plazo para retirar un mensaje ya enviado.
//
// Existe porque "eliminar para todos" es deshacer un error reciente, no
// reescribir una conversación pasada. Sin límite, un instructor podría borrar
// hoy lo que dijo hace tres meses justo cuando alguien lo cuestiona, y la otra
// persona vería evaporarse su historial sin haber hecho nada.
const ventanaBorrado = time.Hour

func (s *MensajesService) EliminarMensaje(ctx context.Context, req *mensajespb.EliminarMensajeRequest) (*mensajespb.EliminarMensajeResponse, error) {
	m, err := s.repo.BuscarParaBorrar(ctx, req.MsgId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "el mensaje no existe")
	}
	if err != nil {
		slog.Error("EliminarMensaje: buscar", "msg_id", req.MsgId, "error", err)
		return nil, status.Error(codes.Internal, "no se pudo leer el mensaje")
	}

	// Pertenencia. Sin esto bastaría con probar identificadores en la URL para
	// ir borrando mensajes de conversaciones ajenas.
	if m.IsGroup {
		if err := s.verificarMiembroGrupo(ctx, req.UserId, m.ReceptorID); err != nil {
			return nil, err
		}
	} else if req.UserId != m.EmisorID && req.UserId != m.ReceptorID {
		// NotFound y no PermissionDenied: responder "no puedes" confirmaría que
		// ese mensaje existe, y con eso se puede sondear la plataforma.
		return nil, status.Error(codes.NotFound, "el mensaje no existe")
	}

	// Los destinatarios se arman aquí, con la fila ya leída, y se devuelven
	// incluso cuando no hay nada que anunciar: quien decide si avisa es el
	// gateway, que es el que tiene el WebSocket.
	destinatarios := &mensajespb.EliminarMensajeResponse{
		EmisorId:   m.EmisorID,
		ReceptorId: m.ReceptorID,
		IsGroup:    m.IsGroup,
	}

	if !req.ParaTodos {
		if err := s.repo.OcultarMensaje(ctx, req.MsgId, req.UserId); err != nil {
			slog.Error("EliminarMensaje: ocultar", "msg_id", req.MsgId, "error", err)
			return nil, status.Error(codes.Internal, "no se pudo eliminar el mensaje")
		}
		return destinatarios, nil
	}

	if req.UserId != m.EmisorID {
		return nil, status.Error(codes.PermissionDenied,
			"solo quien escribió el mensaje puede eliminarlo para todos")
	}
	if m.Eliminado.Valid {
		// Ya estaba borrado. Es el resultado que se pedía, así que no es error.
		return destinatarios, nil
	}
	if time.Since(m.CreatedAt) > ventanaBorrado {
		return nil, status.Error(codes.FailedPrecondition,
			"ya pasó el tiempo para eliminar este mensaje para todos; puedes eliminarlo solo para ti")
	}

	if _, err := s.repo.EliminarParaTodos(ctx, req.MsgId, req.UserId); err != nil {
		slog.Error("EliminarMensaje: para todos", "msg_id", req.MsgId, "error", err)
		return nil, status.Error(codes.Internal, "no se pudo eliminar el mensaje")
	}
	return destinatarios, nil
}

func (s *MensajesService) EliminarConversacion(ctx context.Context, req *mensajespb.EliminarConversacionRequest) (*mensajespb.Empty, error) {
	// En grupo se exige pertenencia; en directo no hay nada que comprobar,
	// porque ocultar una conversación solo afecta a quien lo pide. Si el peer
	// no existe, lo único que queda es una fila inerte en su propia bandeja.
	if req.IsGroup {
		if err := s.verificarMiembroGrupo(ctx, req.UserId, req.PeerId); err != nil {
			return nil, err
		}
	}

	if err := s.repo.OcultarConversacion(ctx, req.UserId, req.PeerId, req.IsGroup); err != nil {
		slog.Error("EliminarConversacion", "user_id", req.UserId, "peer_id", req.PeerId, "error", err)
		return nil, status.Error(codes.Internal, "no se pudo eliminar la conversación")
	}
	return &mensajespb.Empty{}, nil
}
