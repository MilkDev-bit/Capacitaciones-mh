package service

import (
	"context"
	"log/slog"
	"strings"
	"unicode/utf8"

	mensajespb "Prueba-Go/gen/mensajes"
	"Prueba-Go/services/mensajes/internal/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MensajesService struct {
	repo      repository.MensajesRepository
	contactos repository.ContactosRepository
}

func NewMensajesService(repo repository.MensajesRepository, contactos repository.ContactosRepository) *MensajesService {
	return &MensajesService{repo: repo, contactos: contactos}
}

// errNoContactable es el mensaje único para cualquier intento de escribir
// fuera del propio curso. Es deliberadamente genérico: distinguir entre "ese
// usuario no existe" y "existe pero no compartes curso" convertiría el
// endpoint en un oráculo para enumerar cuentas de la plataforma.
const errNoContactable = "solo puedes escribir a personas de tus capacitaciones"

// verificarContacto aplica la regla de visibilidad a un destinatario directo.
// Las conversaciones creadas antes de esta restricción conservan su historial:
// la regla solo se evalúa al enviar un mensaje nuevo.
func (s *MensajesService) verificarContacto(ctx context.Context, emisorID, receptorID string) error {
	if s.contactos == nil {
		return nil
	}
	ok, err := s.contactos.PuedeContactar(ctx, emisorID, receptorID)
	if err != nil {
		slog.Error("verificarContacto", "emisor", emisorID, "receptor", receptorID, "error", err)
		return status.Error(codes.Internal, "no se pudo validar el destinatario")
	}
	if !ok {
		return status.Error(codes.PermissionDenied, errNoContactable)
	}
	return nil
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

	// Autorización del destinatario. En grupo se comprueba la pertenencia
	// (los grupos de cohorte los crea el sistema y sus miembros ya son del
	// mismo curso); en directo, que compartan capacitación.
	if req.IsGroup {
		if err := s.verificarMiembroGrupo(ctx, req.EmisorId, req.ReceptorId); err != nil {
			return nil, err
		}
	} else if err := s.verificarContacto(ctx, req.EmisorId, req.ReceptorId); err != nil {
		return nil, err
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

// verificarMiembroGrupo exige que el emisor pertenezca al grupo al que escribe.
func (s *MensajesService) verificarMiembroGrupo(ctx context.Context, userID, grupoID string) error {
	if s.contactos == nil {
		return nil
	}
	ok, err := s.contactos.EsMiembroDeGrupo(ctx, userID, grupoID)
	if err != nil {
		slog.Error("verificarMiembroGrupo", "user_id", userID, "grupo_id", grupoID, "error", err)
		return status.Error(codes.Internal, "no se pudo validar el grupo")
	}
	if !ok {
		return status.Error(codes.PermissionDenied, "no perteneces a este grupo")
	}
	return nil
}

// miembrosPermitidos filtra una lista de candidatos dejando solo a quienes
// comparten curso con el titular. Devuelve error si alguno queda fuera, en
// lugar de descartarlo en silencio: un grupo creado con menos gente de la que
// el usuario seleccionó es peor que un error explícito.
func (s *MensajesService) miembrosPermitidos(ctx context.Context, titularID string, candidatos []string) ([]string, error) {
	if s.contactos == nil || len(candidatos) == 0 {
		return candidatos, nil
	}

	// Se deduplica y se excluye al titular, que se añade siempre aparte.
	vistos := make(map[string]bool, len(candidatos))
	var unicos []string
	for _, id := range candidatos {
		if id == "" || id == titularID || vistos[id] {
			continue
		}
		vistos[id] = true
		unicos = append(unicos, id)
	}
	if len(unicos) == 0 {
		return nil, nil
	}

	permitidos, err := s.contactos.FiltrarContactables(ctx, titularID, unicos)
	if err != nil {
		slog.Error("miembrosPermitidos", "titular", titularID, "error", err)
		return nil, status.Error(codes.Internal, "no se pudieron validar los miembros")
	}
	if len(permitidos) != len(unicos) {
		return nil, status.Error(codes.PermissionDenied,
			"solo puedes agregar a personas de tus capacitaciones")
	}
	return permitidos, nil
}

func (s *MensajesService) CreateGroup(ctx context.Context, req *mensajespb.CreateGroupRequest) (*mensajespb.CreateGroupResponse, error) {
	if strings.TrimSpace(req.Nombre) == "" {
		return nil, status.Error(codes.InvalidArgument, "el grupo necesita un nombre")
	}

	// Se valida ANTES de crear el grupo: si un miembro no es válido no debe
	// quedar un grupo huérfano de un solo integrante en la base de datos.
	miembros, err := s.miembrosPermitidos(ctx, req.AdminId, req.Members)
	if err != nil {
		return nil, err
	}

	grupoID, err := s.repo.CreateGroup(ctx, req.Nombre, req.AdminId)
	if err != nil {
		return nil, status.Error(codes.Internal, "error creando grupo")
	}
	if len(miembros) > 0 {
		if err := s.repo.AddGroupMembers(ctx, grupoID, miembros); err != nil {
			return nil, status.Error(codes.Internal, "error añadiendo miembros")
		}
	}
	return &mensajespb.CreateGroupResponse{GrupoId: grupoID, Nombre: req.Nombre}, nil
}

func (s *MensajesService) AddGroupMembers(ctx context.Context, req *mensajespb.AddGroupMembersRequest) (*mensajespb.Empty, error) {
	// Los candidatos se contrastan contra el admin del grupo, no contra quien
	// hace la llamada: es el admin quien define el alcance del grupo, y así la
	// regla no depende de qué ruta invoque el RPC.
	adminID, err := s.contactosAdmin(ctx, req.GrupoId)
	if err != nil {
		return nil, err
	}

	miembros, err := s.miembrosPermitidos(ctx, adminID, req.UserIds)
	if err != nil {
		return nil, err
	}
	if len(miembros) == 0 {
		return &mensajespb.Empty{}, nil
	}

	if err := s.repo.AddGroupMembers(ctx, req.GrupoId, miembros); err != nil {
		return nil, status.Error(codes.Internal, "error añadiendo miembros")
	}
	return &mensajespb.Empty{}, nil
}

func (s *MensajesService) contactosAdmin(ctx context.Context, grupoID string) (string, error) {
	if s.contactos == nil {
		return "", nil
	}
	adminID, err := s.contactos.AdminDeGrupo(ctx, grupoID)
	if err != nil {
		slog.Error("contactosAdmin", "grupo_id", grupoID, "error", err)
		return "", status.Error(codes.NotFound, "el grupo no existe")
	}
	return adminID, nil
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
