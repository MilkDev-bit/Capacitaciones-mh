package handler

import (
	"context"

	mensajespb "Prueba-Go/gen/mensajes"
	"Prueba-Go/services/mensajes/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MensajesHandler struct {
	mensajespb.UnimplementedMensajesServiceServer
	svc *service.MensajesService
}

func NewMensajesHandler(svc *service.MensajesService) *MensajesHandler {
	return &MensajesHandler{svc: svc}
}

func (h *MensajesHandler) SendMensaje(ctx context.Context, req *mensajespb.SendMensajeRequest) (*mensajespb.MensajeResponse, error) {
	if req.EmisorId == "" || req.ReceptorId == "" {
		return nil, status.Error(codes.InvalidArgument, "emisor_id y receptor_id son requeridos")
	}
	return h.svc.Send(ctx, req)
}

func (h *MensajesHandler) GetMensajes(ctx context.Context, req *mensajespb.GetMensajesRequest) (*mensajespb.GetMensajesResponse, error) {
	if req.UserId == "" || req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id y peer_id son requeridos")
	}
	return h.svc.GetMensajes(ctx, req)
}

func (h *MensajesHandler) ListConversaciones(ctx context.Context, req *mensajespb.ListConversacionesRequest) (*mensajespb.ListConversacionesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id es requerido")
	}
	return h.svc.ListConversaciones(ctx, req)
}

func (h *MensajesHandler) NoLeidos(ctx context.Context, req *mensajespb.NoLeidosRequest) (*mensajespb.NoLeidosResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id es requerido")
	}
	return h.svc.NoLeidos(ctx, req)
}
