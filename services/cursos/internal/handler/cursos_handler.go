package handler

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/repository"
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
		slog.Error("ListCursosPublicos error", "error", err)
		return nil, mapErr(err)
	}
	return &cursospb.ListCursosResponse{Cursos: list}, nil
}

// ── Usuario ────────────────────────────────────────────────────────────────────

func (h *CursosHandler) ListMisCapacitaciones(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListCursosResponse, error) {
	list, err := h.svc.ListMisCapacitaciones(ctx, req.UserId)
	if err != nil {
		slog.Error("ListMisCapacitaciones error", "error", err, "userId", req.UserId)
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

func (h *CursosHandler) UnirseConCodigo(ctx context.Context, req *cursospb.UnirseRequest) (*cursospb.CursoResponse, error) {
	curso, err := h.svc.UnirseConCodigo(ctx, req.UserId, req.Codigo)
	if err != nil {
		return nil, mapErr(err)
	}
	return curso, nil
}

func (h *CursosHandler) UnirseConLicencia(ctx context.Context, req *cursospb.UnirseConLicenciaRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.UnirseConLicencia(ctx, req.UserId, req.CapacitacionId, req.CodigoAcceso); err != nil {
		return nil, mapErr(err)
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) WebhookEnroll(ctx context.Context, req *cursospb.WebhookEnrollRequest) (*cursospb.EnrollResponse, error) {
	resp, err := h.svc.WebhookEnroll(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) WebhookComprarLicencia(ctx context.Context, req *cursospb.WebhookComprarLicenciaRequest) (*cursospb.EmptyResponse, error) {
	return h.svc.WebhookComprarLicencia(ctx, req)
}

func (h *CursosHandler) WebhookComprarB2BDirect(ctx context.Context, req *cursospb.WebhookComprarB2BDirectRequest) (*cursospb.ComprarB2BDirectResponse, error) {
	return h.svc.WebhookComprarB2BDirect(ctx, req)
}

// AsignarAccesosLicencia reparte accesos de una licencia entre participantes.
func (h *CursosHandler) AsignarAccesosLicencia(ctx context.Context, req *cursospb.AsignarAccesosLicenciaRequest) (*cursospb.AsignarAccesosLicenciaResponse, error) {
	if req.LicenciaId == "" || req.CompradorId == "" {
		return nil, status.Error(codes.InvalidArgument, "licencia_id y comprador_id son requeridos")
	}
	resp, err := h.svc.AsignarAccesosLicencia(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

// NotificarCursoCompletado decide si toca avisar al representante sobre el DC-3.
func (h *CursosHandler) NotificarCursoCompletado(ctx context.Context, req *cursospb.CursoCompletadoRequest) (*cursospb.CursoCompletadoResponse, error) {
	if req.UserId == "" || req.CapacitacionId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id y capacitacion_id son requeridos")
	}
	resp, err := h.svc.NotificarCursoCompletado(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

// RegistrarEventoStripe deduplica los webhooks entrantes.
func (h *CursosHandler) RegistrarEventoStripe(ctx context.Context, req *cursospb.EventoStripeRequest) (*cursospb.EventoStripeResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id es requerido")
	}
	resp, err := h.svc.RegistrarEventoStripe(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

// ActualizarEstadoOrden mueve la orden por su máquina de estados.
func (h *CursosHandler) ActualizarEstadoOrden(ctx context.Context, req *cursospb.ActualizarEstadoOrdenRequest) (*cursospb.EmptyResponse, error) {
	if req.StripeSessionId == "" || req.Estado == "" {
		return nil, status.Error(codes.InvalidArgument, "stripe_session_id y estado son requeridos")
	}
	resp, err := h.svc.ActualizarEstadoOrden(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

// ListInvitacionesLicencia devuelve el estado de entrega de los accesos.
func (h *CursosHandler) ListInvitacionesLicencia(ctx context.Context, req *cursospb.LicenciaIDRequest) (*cursospb.ListInvitacionesLicenciaResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "licencia_id es requerido")
	}
	resp, err := h.svc.ListInvitacionesLicencia(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) CreateCheckoutSession(ctx context.Context, req *cursospb.CheckoutSessionRequest) (*cursospb.CheckoutSessionResponse, error) {
	resp, err := h.svc.CreateCheckoutSession(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrYaInscrito) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		slog.Error("CreateCheckoutSession error", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (h *CursosHandler) CreateCheckoutSessionCart(ctx context.Context, req *cursospb.CheckoutCartRequest) (*cursospb.CheckoutSessionResponse, error) {
	resp, err := h.svc.CreateCheckoutSessionCart(ctx, req)
	if err != nil {
		// Comprar algo que ya tienes es un error del usuario, no del servidor:
		// como Internal se pintaba un 500 genérico y el mensaje útil se perdía.
		if errors.Is(err, service.ErrYaInscrito) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
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

// ── Horarios Instructores ────────────────────────────────────────────────

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, service.ErrNotFound) || errors.Is(err, sql.ErrNoRows):
		return status.Error(codes.NotFound, "recurso o código de acceso no encontrado")
	case errors.Is(err, service.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrRequierePago):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	// La licencia se quedó sin lugares por repartir: es un error del usuario,
	// no del servidor, y el frontend debe mostrar el mensaje tal cual.
	case errors.Is(err, repository.ErrOrdenYaPagada):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, repository.ErrSinAccesosDisponibles):
		return status.Error(codes.FailedPrecondition, err.Error())
	case strings.Contains(err.Error(), "no es válido") || strings.Contains(err.Error(), "inválido") || strings.Contains(err.Error(), "capacidad máxima") || strings.Contains(err.Error(), "no corresponde") || strings.Contains(err.Error(), "de pago") || strings.Contains(err.Error(), "requerido") || strings.Contains(err.Error(), "invalid input syntax"):
		return status.Error(codes.InvalidArgument, err.Error())
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
