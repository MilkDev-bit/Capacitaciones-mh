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
	// Si hay adjunto, el texto puede estar vacío; si no hay adjunto, debe tener texto
	if req.AttachmentUrl == "" {
		if n := utf8.RuneCountInString(contenido); n < 1 || n > 5000 {
			return nil, status.Error(codes.InvalidArgument, "el mensaje debe tener entre 1 y 5000 caracteres")
		}
	} else if n := utf8.RuneCountInString(contenido); n > 5000 {
		return nil, status.Error(codes.InvalidArgument, "el texto del mensaje es demasiado largo")
	}
	m, err := s.repo.Send(ctx, &repository.Mensaje{
		EmisorID:       req.EmisorId,
		EmisorName:     req.EmisorName,
		ReceptorID:     req.ReceptorId,
		ReceptorName:   req.ReceptorName,
		Contenido:      contenido,
		AttachmentUrl:  req.AttachmentUrl,
		AttachmentType: req.AttachmentType,
		IsGroup:        req.IsGroup,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "error guardando mensaje")
	}
	return m.ToProto(), nil
}

func (s *MensajesService) GetMensajes(ctx context.Context, req *mensajespb.GetMensajesRequest) (*mensajespb.GetMensajesResponse, error) {
// Solo marcar como leidos en la carga inicial (sin cursor de paginacion)
if req.BeforeId == "" {
_ = s.repo.MarcarLeidos(ctx, req.UserId, req.PeerId, req.IsGroup)
}

msgs, hasMore, err := s.repo.GetConversacion(ctx, req.UserId, req.PeerId, int(req.Limit), req.BeforeId, req.IsGroup)
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

func (s *MensajesService) MarcarLeidos(ctx context.Context, req *mensajespb.MarcarLeidosRequest) (*mensajespb.Empty, error) {
if err := s.repo.MarcarLeidos(ctx, req.UserId, req.PeerId, req.IsGroup); err != nil {
return nil, status.Error(codes.Internal, "error marcando leidos")
}
return &mensajespb.Empty{}, nil
}

func (s *MensajesService) CreateGroup(ctx context.Context, req *mensajespb.CreateGroupRequest) (*mensajespb.CreateGroupResponse, error) {
grupoID, err := s.repo.CreateGroup(ctx, req.Nombre, req.AdminId)
if err != nil {
return nil, status.Error(codes.Internal, "error creando grupo")
}
if len(req.Members) > 0 {
if err := s.repo.AddGroupMembers(ctx, grupoID, req.Members); err != nil {
return nil, status.Error(codes.Internal, "error añadiendo miembros")
}
}
return &mensajespb.CreateGroupResponse{GrupoId: grupoID, Nombre: req.Nombre}, nil
}

func (s *MensajesService) AddGroupMembers(ctx context.Context, req *mensajespb.AddGroupMembersRequest) (*mensajespb.Empty, error) {
if err := s.repo.AddGroupMembers(ctx, req.GrupoId, req.UserIds); err != nil {
return nil, status.Error(codes.Internal, "error añadiendo miembros")
}
return &mensajespb.Empty{}, nil
}

func (s *MensajesService) GetGroupMembers(ctx context.Context, req *mensajespb.GetGroupMembersRequest) (*mensajespb.GetGroupMembersResponse, error) {
members, err := s.repo.GetGroupMembers(ctx, req.GrupoId)
if err != nil {
return nil, status.Error(codes.Internal, "error obteniendo miembros")
}
return &mensajespb.GetGroupMembersResponse{UserIds: members}, nil
}

func (s *MensajesService) CreateGroupForLicencia(ctx context.Context, req *mensajespb.CreateGroupForLicenciaRequest) (*mensajespb.CreateGroupResponse, error) {
	// Idempotent: if a group already exists for this licencia, return it
	existingID, err := s.repo.GetGroupIDByLicencia(ctx, req.LicenciaId)
	if err == nil && existingID != "" {
		return &mensajespb.CreateGroupResponse{GrupoId: existingID, Nombre: req.Nombre}, nil
	}
	grupoID, err := s.repo.CreateGroupForLicencia(ctx, req.Nombre, req.AdminId, req.LicenciaId)
	if err != nil {
		return nil, status.Error(codes.Internal, "error creando grupo de cohorte")
	}
	// Auto-add admin as member
	_ = s.repo.AddGroupMembers(ctx, grupoID, []string{req.AdminId})
	return &mensajespb.CreateGroupResponse{GrupoId: grupoID, Nombre: req.Nombre}, nil
}

func (s *MensajesService) EnrollInLicenciaGroup(ctx context.Context, req *mensajespb.EnrollInLicenciaGroupRequest) (*mensajespb.Empty, error) {
	grupoID, err := s.repo.GetGroupIDByLicencia(ctx, req.LicenciaId)
	if err != nil || grupoID == "" {
		// No group yet — silently skip (group created on license creation)
		return &mensajespb.Empty{}, nil
	}
	if err := s.repo.AddGroupMembers(ctx, grupoID, []string{req.UserId}); err != nil {
		return nil, status.Error(codes.Internal, "error añadiendo al grupo de cohorte")
	}
	return &mensajespb.Empty{}, nil
}
