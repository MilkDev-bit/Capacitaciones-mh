package handler

// Capa gRPC de las constancias DC-3.

import (
	"context"
	"errors"
	"log/slog"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CursosHandler) GetDatosDC3(ctx context.Context, req *cursospb.DatosDC3Request) (*cursospb.DatosDC3Response, error) {
	resp, err := h.svc.GetDatosDC3(ctx, req)
	switch {
	case errors.Is(err, service.ErrDC3NoHabilitado):
		// FailedPrecondition y no InvalidArgument: la petición es correcta, es
		// el curso el que no emite constancias. El Gateway lo traduce en un
		// mensaje al instructor, no en un error de validación al alumno.
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrNotFound):
		return nil, status.Error(codes.NotFound, "capacitación no encontrada")
	case err != nil:
		slog.Error("GetDatosDC3", "user_id", req.UserId, "capacitacion_id", req.CapacitacionId, "error", err)
		return nil, status.Error(codes.Internal, "error obteniendo los datos de la constancia")
	}
	return resp, nil
}

func (h *CursosHandler) GuardarDatosTrabajador(ctx context.Context, req *cursospb.DatosTrabajadorRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.GuardarDatosTrabajador(ctx, req); err != nil {
		// Las validaciones de este servicio son de formato (CURP de 18, campos
		// obligatorios) y el mensaje se le muestra al alumno tal cual, así que
		// viaja como InvalidArgument en lugar de un Internal opaco.
		slog.Warn("GuardarDatosTrabajador rechazado", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) RegistrarConstanciaDC3(ctx context.Context, req *cursospb.RegistrarConstanciaRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.RegistrarConstanciaDC3(ctx, req); err != nil {
		slog.Error("RegistrarConstanciaDC3", "user_id", req.UserId,
			"capacitacion_id", req.CapacitacionId, "error", err)
		return nil, status.Error(codes.Internal, "error registrando la constancia")
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) GetEmpresaInstructor(ctx context.Context, req *cursospb.UserRequest) (*cursospb.DatosEmpresaDC3, error) {
	e, err := h.svc.GetEmpresaInstructor(ctx, req.UserId)
	if err != nil {
		slog.Error("GetEmpresaInstructor", "instructor_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "error obteniendo los datos de la empresa")
	}
	return e, nil
}

func (h *CursosHandler) GuardarEmpresaInstructor(ctx context.Context, req *cursospb.EmpresaInstructorRequest) (*cursospb.EmptyResponse, error) {
	if err := h.svc.GuardarEmpresaInstructor(ctx, req); err != nil {
		// El mensaje se le muestra al instructor tal cual, así que viaja como
		// InvalidArgument y no como un Internal opaco.
		slog.Warn("GuardarEmpresaInstructor rechazado", "instructor_id", req.InstructorId, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &cursospb.EmptyResponse{}, nil
}

func (h *CursosHandler) ListMisConstancias(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListConstanciasResponse, error) {
	resp, err := h.svc.ListMisConstancias(ctx, req.UserId)
	if err != nil {
		slog.Error("ListMisConstancias", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "error listando constancias")
	}
	return resp, nil
}

// VerificarConstancia atiende la consulta pública por folio.
//
// El folio NO se registra en el log. Es el dato que permite consultar una
// constancia sin sesión, así que dejarlo escrito en los registros lo convierte
// en algo que se filtra con ellos.
func (h *CursosHandler) VerificarConstancia(ctx context.Context, req *cursospb.VerificarConstanciaRequest) (*cursospb.VerificarConstanciaResponse, error) {
	resp, err := h.svc.VerificarConstancia(ctx, req.Folio)
	if err != nil {
		slog.Error("VerificarConstancia", "error", err)
		return nil, status.Error(codes.Internal, "error verificando la constancia")
	}
	return resp, nil
}
