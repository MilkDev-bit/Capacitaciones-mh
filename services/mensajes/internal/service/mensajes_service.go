package service

import (
"context"
"strings"
"unicode/utf8"

mensajespb "Prueba-Go/gen/mensajes"
"Prueba-Go/services/mensajes/internal/repository"

"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

type MensajesService struct {
repo repository.MensajesRepository
}

func NewMensajesService(repo repository.MensajesRepository) *MensajesService {
return &MensajesService{repo: repo}
}

func (s *MensajesService) Send(ctx context.Context, req *mensajespb.SendMensajeRequest) (*mensajespb.MensajeResponse, error) {
if req.EmisorId == req.ReceptorId {
return nil, status.Error(codes.InvalidArgument, "no puedes enviarte mensajes a ti mismo")
}
contenido := strings.TrimSpace(req.Contenido)
if n := utf8.RuneCountInString(contenido); n < 1 || n > 5000 {
return nil, status.Error(codes.InvalidArgument, "el mensaje debe tener entre 1 y 5000 caracteres")
}
m, err := s.repo.Send(ctx, &repository.Mensaje{
EmisorID:     req.EmisorId,
EmisorName:   req.EmisorName,
ReceptorID:   req.ReceptorId,
ReceptorName: req.ReceptorName,
Contenido:    contenido,
})
if err != nil {
return nil, status.Error(codes.Internal, "error enviando mensaje")
}
return m.ToProto(), nil
}

func (s *MensajesService) GetMensajes(ctx context.Context, req *mensajespb.GetMensajesRequest) (*mensajespb.GetMensajesResponse, error) {
// Solo marcar como leidos en la carga inicial (sin cursor de paginacion)
if req.BeforeId == "" {
_ = s.repo.MarcarLeidos(ctx, req.UserId, req.PeerId)
}

msgs, hasMore, err := s.repo.GetConversacion(ctx, req.UserId, req.PeerId, int(req.Limit), req.BeforeId)
if err != nil {
return nil, status.Error(codes.Internal, "error cargando mensajes")
}
resp := &mensajespb.GetMensajesResponse{HasMore: hasMore}
for _, m := range msgs {
resp.Mensajes = append(resp.Mensajes, m.ToProto())
}
return resp, nil
}

func (s *MensajesService) ListConversaciones(ctx context.Context, req *mensajespb.ListConversacionesRequest) (*mensajespb.ListConversacionesResponse, error) {
convs, err := s.repo.ListConversaciones(ctx, req.UserId)
if err != nil {
return nil, status.Error(codes.Internal, "error cargando conversaciones")
}
resp := &mensajespb.ListConversacionesResponse{}
for _, c := range convs {
resp.Conversaciones = append(resp.Conversaciones, c.ToProto())
}
return resp, nil
}

func (s *MensajesService) NoLeidos(ctx context.Context, req *mensajespb.NoLeidosRequest) (*mensajespb.NoLeidosResponse, error) {
count, err := s.repo.NoLeidos(ctx, req.UserId)
if err != nil {
return nil, status.Error(codes.Internal, "error contando mensajes")
}
return &mensajespb.NoLeidosResponse{Count: count}, nil
}

func (s *MensajesService) MarcarLeido(ctx context.Context, req *mensajespb.MarcarLeidoRequest) (*mensajespb.MarcarLeidoResponse, error) {
emisorID, err := s.repo.MarcarLeido(ctx, req.MsgId, req.UserId)
if err != nil {
return nil, status.Error(codes.Internal, "error marcando mensaje")
}
return &mensajespb.MarcarLeidoResponse{Ok: emisorID != "", EmisorId: emisorID}, nil
}
