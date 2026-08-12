package handler

import (
	"context"

	cursospb "Prueba-Go/gen/cursos"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─────────────────────────────────────────────────────────────────────────────
// Suscripciones — capa de presentación gRPC.
// Solo valida entrada y traduce errores; la lógica vive en el servicio.
// ─────────────────────────────────────────────────────────────────────────────

func (h *CursosHandler) ListPlanes(ctx context.Context, _ *cursospb.EmptyRequest) (*cursospb.ListPlanesResponse, error) {
	resp, err := h.svc.ListPlanes(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) GetMiSuscripcion(ctx context.Context, req *cursospb.UserRequest) (*cursospb.SuscripcionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id es requerido")
	}
	resp, err := h.svc.GetMiSuscripcion(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) CrearCheckoutSuscripcion(ctx context.Context, req *cursospb.CheckoutSuscripcionRequest) (*cursospb.CheckoutSessionResponse, error) {
	if req.UserId == "" || req.PlanCodigo == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id y plan_codigo son requeridos")
	}
	resp, err := h.svc.CrearCheckoutSuscripcion(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) SincronizarSuscripcion(ctx context.Context, req *cursospb.SincronizarSuscripcionRequest) (*cursospb.EmptyResponse, error) {
	if req.StripeSubscriptionId == "" {
		return nil, status.Error(codes.InvalidArgument, "stripe_subscription_id es requerido")
	}
	resp, err := h.svc.SincronizarSuscripcion(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) RegistrarFacturaSuscripcion(ctx context.Context, req *cursospb.FacturaSuscripcionRequest) (*cursospb.EmptyResponse, error) {
	if req.StripeInvoiceId == "" {
		return nil, status.Error(codes.InvalidArgument, "stripe_invoice_id es requerido")
	}
	resp, err := h.svc.RegistrarFacturaSuscripcion(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) TieneAccesoPorSuscripcion(ctx context.Context, req *cursospb.UserRequest) (*cursospb.AccesoSuscripcionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id es requerido")
	}
	resp, err := h.svc.TieneAccesoPorSuscripcion(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) AsignarAsientos(ctx context.Context, req *cursospb.AsignarAsientosRequest) (*cursospb.ListAsientosResponse, error) {
	if req.SuscripcionId == "" || req.TitularId == "" {
		return nil, status.Error(codes.InvalidArgument, "suscripcion_id y titular_id son requeridos")
	}
	resp, err := h.svc.AsignarAsientos(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) ListAsientos(ctx context.Context, req *cursospb.SuscripcionIDRequest) (*cursospb.ListAsientosResponse, error) {
	if req.SuscripcionId == "" || req.TitularId == "" {
		return nil, status.Error(codes.InvalidArgument, "suscripcion_id y titular_id son requeridos")
	}
	resp, err := h.svc.ListAsientos(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (h *CursosHandler) RevocarAsiento(ctx context.Context, req *cursospb.RevocarAsientoRequest) (*cursospb.EmptyResponse, error) {
	if req.SuscripcionId == "" || req.TitularId == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "suscripcion_id, titular_id y email son requeridos")
	}
	resp, err := h.svc.RevocarAsiento(ctx, req)
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}
