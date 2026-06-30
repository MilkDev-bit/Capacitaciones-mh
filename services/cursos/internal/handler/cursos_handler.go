package handler

import (
	"context"
	"errors"
	"log/slog"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CursosHandler implementa cursospb.CursosServiceServer.
type CursosHandler struct {
	cursospb.UnimplementedCursosServiceServer
	svc *service.CursosService
}

func NewCursosHandler(svc *service.CursosService) *CursosHandler {
	return &CursosHandler{svc: svc}
}

// ── Público ────────────────────────────────────────────────────────────────────

func (h *CursosHandler) PreviewCurso(ctx context.Context, req *cursospb.CodigoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.PreviewCurso(ctx, req.Codigo)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) GetCursoPublico(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.GetCursoPublico(ctx, req.CursoId)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) ListCursosPublicos(ctx context.Context, _ *cursospb.EmptyRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.ListPublicos(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

// ── Usuario ────────────────────────────────────────────────────────────────────

func (h *CursosHandler) ListMisCapacitaciones(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.ListMisCapacitaciones(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

func (h *CursosHandler) GetCurso(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.GetCurso(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) Inscribirse(ctx context.Context, req *cursospb.InscribirseRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.Inscribirse(ctx, req.UserId, req.CursoId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) UnirseConCodigo(ctx context.Context, req *cursospb.UnirseRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.UnirseConCodigo(ctx, req.UserId, req.Codigo); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) UnirseConLicencia(ctx context.Context, req *cursospb.UnirseConLicenciaRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.UnirseConLicencia(ctx, req.UserId, req.CapacitacionId, req.CodigoAcceso); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) WebhookEnroll(ctx context.Context, req *cursospb.WebhookEnrollRequest) (*cursospb.EmptyResponse, error) {
	resp, err := h.svc.WebhookEnroll(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) WebhookComprarLicencia(ctx context.Context, req *cursospb.WebhookComprarLicenciaRequest) (*cursospb.EmptyResponse, error) {
	return h.svc.WebhookComprarLicencia(ctx, req)
}

func (h *CursosHandler) WebhookComprarB2BDirect(ctx context.Context, req *cursospb.WebhookComprarB2BDirectRequest) (*cursospb.EmptyResponse, error) {
	return h.svc.WebhookComprarB2BDirect(ctx, req)
}

func (h *CursosHandler) CreateCheckoutSession(ctx context.Context, req *cursospb.CheckoutSessionRequest) (*cursospb.CheckoutSessionResponse, error) {
	resp, err := h.svc.CreateCheckoutSession(ctx, req)
	if err != nil {
		slog.Error("CreateCheckoutSession error", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (h *CursosHandler) CreateCheckoutSessionCart(ctx context.Context, req *cursospb.CheckoutCartRequest) (*cursospb.CheckoutSessionResponse, error) {
	resp, err := h.svc.CreateCheckoutSessionCart(ctx, req)
	if err != nil {
		slog.Error("CreateCheckoutSessionCart error", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (h *CursosHandler) CreateCheckoutSessionB2BDirect(ctx context.Context, req *cursospb.CreateCheckoutSessionB2BDirectRequest) (*cursospb.CheckoutSessionResponse, error) {
	resp, err := h.svc.CreateCheckoutSessionB2BDirect(ctx, req)
	if err != nil {
		slog.Error("CreateCheckoutSessionB2BDirect error", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (h *CursosHandler) ListLicenciasCompradas(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListLicenciasResponse, error) {
	resp, err := h.svc.ListLicenciasCompradas(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (h *CursosHandler) ListLicenciaTickets(ctx context.Context, req *cursospb.LicenciaIDRequest) (*cursospb.ListLicenciaTicketsResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "licencia_id es requerido")
	}

	tickets, err := h.svc.ListTicketsByLicencia(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error al obtener tickets: %v", err)
	}

	var pbTickets []*cursospb.VideocallTicket
	for _, t := range tickets {
		pbTickets = append(pbTickets, t.ToProto())
	}

	return &cursospb.ListLicenciaTicketsResponse{Tickets: pbTickets}, nil
}

func (h *CursosHandler) InstructorListCapacitaciones(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.InstructorListCapacitaciones(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

func (h *CursosHandler) InstructorCreateCapacitacion(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorCreate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorUpdateCapacitacion(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorUpdate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorDeleteCapacitacion(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.InstructorDelete(ctx, req.CursoId, req.UserId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) InstructorTogglePublic(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorTogglePublic(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorResetCodigo(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.InstructorResetCodigo(ctx, req.CursoId, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) InstructorListEstudiantes(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListEstudiantesResponse, error) {
	// req.UserId es el instructor; el curso viene como parámetro adicional en
	// el gateway (por convención lo pasamos en user_id para este RPC específico).
	// En la práctica el gateway envía instructor_id y el curso_id en campos separados.
	// Como el proto define solo UserRequest, el gateway lo combina; aquí recibimos
	// el instructorID. Sin curso_id en este proto RPC, listamos todos sus cursos de estudiantes.
	list, err := h.svc.InstructorListEstudiantes(ctx, req.UserId, "")
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListEstudiantesResponse{Estudiantes: list}, nil
}

func (h *CursosHandler) InstructorAsignar(ctx context.Context, req *cursospb.AsignarRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.InstructorAsignar(ctx, req.RequesterId, req.TargetUserId, req.CapacitacionId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

// ── Admin ──────────────────────────────────────────────────────────────────────

func (h *CursosHandler) AdminListCapacitaciones(ctx context.Context, _ *cursospb.EmptyRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.AdminList(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

func (h *CursosHandler) AdminCreateCapacitacion(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.AdminCreate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) AdminUpdateCapacitacion(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := h.svc.AdminUpdate(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (h *CursosHandler) AdminDeleteCapacitacion(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.AdminDelete(ctx, req.CursoId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) GetAdminDashboardStats(ctx context.Context, req *cursospb.EmptyRequest) (*cursospb.AdminDashboardStatsResponse, error) {
	stats, err := h.svc.GetAdminDashboardStats(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return stats, nil
}

func (h *CursosHandler) AdminListAsignaciones(ctx context.Context, _ *cursospb.EmptyRequest) (*cursospb.ListAsignacionesResponse, error) {
	list, err := h.svc.AdminListAsignaciones(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.ListAsignacionesResponse{Asignaciones: list}, nil
}

func (h *CursosHandler) AdminAsignar(ctx context.Context, req *cursospb.AsignarRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.AdminAsignar(ctx, req.TargetUserId, req.CapacitacionId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) AdminDesAsignar(ctx context.Context, req *cursospb.AsignacionIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.AdminDesAsignar(ctx, req.AsignacionId); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

// ── error mapper ──────────────────────────────────────────────────────────────

// ── Videocalls ────────────────────────────────────────────────────────────

func (h *CursosHandler) JoinVideocall(ctx context.Context, req *cursospb.JoinVideocallRequest) (*cursospb.JoinVideocallResponse, error) {
	res, err := h.svc.JoinVideocall(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return res, nil
}

func (h *CursosHandler) LeaveVideocall(ctx context.Context, req *cursospb.LeaveVideocallRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.LeaveVideocall(ctx, req); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) EndVideocall(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.EndVideocall(ctx, req); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) GetMyVideocallTicket(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.VideocallTicketResponse, error) {
	resp, err := h.svc.GetMyVideocallTicket(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) InstructorGetCurrentRoom(ctx context.Context, req *cursospb.CursoIDRequest) (*cursospb.CurrentRoomResponse, error) {
	resp, err := h.svc.InstructorGetCurrentRoom(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

// ── Horarios Instructores ────────────────────────────────────────────────

func (h *CursosHandler) ListPublicSchedules(ctx context.Context, req *cursospb.ListPublicSchedulesRequest) (*cursospb.ListPublicSchedulesResponse, error) {
	list, err := h.svc.ListPublicSchedules(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return list, nil
}

func (h *CursosHandler) AdminListSchedules(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListSchedulesResponse, error) {
	list, err := h.svc.AdminListSchedules(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return list, nil
}

func (h *CursosHandler) AdminCreateSchedule(ctx context.Context, req *cursospb.CreateScheduleRequest) (*cursospb.InstructorSchedule, error) {
	s, err := h.svc.AdminCreateSchedule(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return s, nil
}

func (h *CursosHandler) AdminUpdateSchedule(ctx context.Context, req *cursospb.UpdateScheduleRequest) (*cursospb.InstructorSchedule, error) {
	s, err := h.svc.AdminUpdateSchedule(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return s, nil
}

func (h *CursosHandler) AdminDeleteSchedule(ctx context.Context, req *cursospb.ScheduleIDRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.AdminDeleteSchedule(ctx, req); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		slog.Error("cursos: error interno", "error", err)
		return status.Error(codes.Internal, "error interno del servidor")
	}
}

func (h *CursosHandler) GetLicenciaPublica(ctx context.Context, req *cursospb.LicenciaIDRequest) (*cursospb.LicenciaPublicaResponse, error) {
	return nil, status.Error(codes.Unimplemented, "obsolete")
}

func (h *CursosHandler) InstructorCreateLicencia(ctx context.Context, req *cursospb.CreateLicenciaRequest) (*cursospb.Licencia, error) {
	return nil, status.Error(codes.Unimplemented, "obsolete")
}

func (h *CursosHandler) InstructorUpdateLicencia(ctx context.Context, req *cursospb.UpdateLicenciaRequest) (*cursospb.Licencia, error) {
	return nil, status.Error(codes.Unimplemented, "obsolete")
}

func (h *CursosHandler) InstructorDeleteLicencia(ctx context.Context, req *cursospb.LicenciaIDRequest) (*cursospb.EmptyResponse, error) {
	return nil, status.Error(codes.Unimplemented, "obsolete")
}
