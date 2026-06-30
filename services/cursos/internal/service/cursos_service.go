package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	cursospb "Prueba-Go/gen/cursos"
	mensajespb "Prueba-Go/gen/mensajes"
	"Prueba-Go/services/cursos/internal/repository"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"google.golang.org/grpc/metadata"
)

// Errores de dominio.
var (
	ErrNotFound  = errors.New("no encontrado")
	ErrForbidden = errors.New("sin permisos")
	ErrConflict  = errors.New("ya inscrito")
)

// CursosService contiene la lógica de negocio del servicio de cursos.
type CursosService struct {
	repo     repository.CursosRepository
	mensajes mensajespb.MensajesServiceClient // nil-safe: called if set
}

func NewCursosService(repo repository.CursosRepository, mensajes mensajespb.MensajesServiceClient) *CursosService {
	return &CursosService{repo: repo, mensajes: mensajes}
}

func (s *CursosService) ListPublicos(ctx context.Context) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListPublicos(ctx)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) PreviewCurso(ctx context.Context, codigo string) (*cursospb.CursoResponse, error) {
	c, err := s.repo.FindByCodigo(ctx, codigo)
	if err != nil {
		return nil, ErrNotFound
	}
	return c.ToProto(), nil
}

func (s *CursosService) GetCursoPublico(ctx context.Context, cursoID string) (*cursospb.CursoResponse, error) {
	c, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if !c.IsPublic {
		return nil, ErrForbidden
	}
	return c.ToProto(), nil
}

func (s *CursosService) ListMisCapacitaciones(ctx context.Context, userID string) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	protos := toProtoSlice(cursos)
	// Para los cursos que son videollamadas, adjuntar su ticket_codigo en codigo_acceso
	for _, p := range protos {
		if p.Type == "videocall" {
			ticket, _ := s.repo.GetTicketForUserAndCourse(ctx, userID, p.Id)
			if ticket != nil {
				p.CodigoAcceso = ticket.Codigo
			}
		}
	}
	
	return protos, nil
}

func (s *CursosService) GetCurso(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	enrolled, err := s.repo.IsEnrolled(ctx, userID, cursoID)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	// Si no está inscrito, solo puede ver si es público.
	if !enrolled && !c.IsPublic {
		return nil, ErrForbidden
	}
	return c.ToProto(), nil
}

func (s *CursosService) Inscribirse(ctx context.Context, userID, cursoID string) error {
	enrolled, _ := s.repo.IsEnrolled(ctx, userID, cursoID)
	if enrolled {
		return ErrConflict
	}
	curso, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return err
	}
	if curso.Precio > 0 {
		return errors.New("este curso es de pago, usa el flujo de compra")
	}
	return s.repo.Inscribirse(ctx, userID, cursoID)
}

func (s *CursosService) UnirseConCodigo(ctx context.Context, userID, codigo string) error {
	_, err := s.repo.UnirseConCodigo(ctx, userID, codigo)
	return err
}

// ── Instructor ────────────────────────────────────────────────────────────────

func (s *CursosService) InstructorListCapacitaciones(ctx context.Context, instructorID string) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) InstructorCreate(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorUpdate(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	// Verificar que el instructor es dueño del curso.
	existing, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != req.UserId {
		return nil, ErrForbidden
	}
	c, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorDelete(ctx context.Context, cursoID, userID string) error {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, cursoID)
}

func (s *CursosService) InstructorTogglePublic(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return nil, ErrForbidden
	}
	c, err := s.repo.TogglePublic(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorResetCodigo(ctx context.Context, cursoID, userID string) (*cursospb.CursoResponse, error) {
	existing, err := s.repo.FindByID(ctx, cursoID)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.InstructorID == nil || *existing.InstructorID != userID {
		return nil, ErrForbidden
	}
	c, err := s.repo.ResetCodigo(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) InstructorListEstudiantes(ctx context.Context, instructorID, cursoID string) ([]*cursospb.EstudianteInfo, error) {
	rows, err := s.repo.ListEstudiantes(ctx, instructorID, cursoID)
	if err != nil {
		return nil, err
	}
	result := make([]*cursospb.EstudianteInfo, 0, len(rows))
	for _, r := range rows {
		result = append(result, &cursospb.EstudianteInfo{
			UserId: r.ID, Name: r.Name, Email: r.Email,
			AssignedAt: r.AssignedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

func (s *CursosService) InstructorAsignar(ctx context.Context, instructorID, userID, cursoID string) error {
	return s.repo.InstructorAsignar(ctx, instructorID, userID, cursoID)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

func (s *CursosService) AdminList(ctx context.Context) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
}

func (s *CursosService) AdminCreate(ctx context.Context, req *cursospb.CreateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) AdminUpdate(ctx context.Context, req *cursospb.UpdateCursoRequest) (*cursospb.CursoResponse, error) {
	c, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *CursosService) AdminDelete(ctx context.Context, cursoID string) error {
	return s.repo.Delete(ctx, cursoID)
}

func (s *CursosService) AdminListAsignaciones(ctx context.Context) ([]*cursospb.AsignacionInfo, error) {
	asigs, err := s.repo.ListAsignaciones(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*cursospb.AsignacionInfo, 0, len(asigs))
	for _, a := range asigs {
		result = append(result, a.ToProto())
	}
	return result, nil
}

func (s *CursosService) AdminAsignar(ctx context.Context, userID, cursoID string) error {
	return s.repo.AdminAsignar(ctx, userID, cursoID)
}

func (s *CursosService) AdminDesAsignar(ctx context.Context, asignacionID string) error {
	return s.repo.DesAsignar(ctx, asignacionID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toProtoSlice(cursos []*repository.Curso) []*cursospb.CursoResponse {
	result := make([]*cursospb.CursoResponse, 0, len(cursos))
	for _, c := range cursos {
		result = append(result, c.ToProto())
	}
	return result
}

// ── Licencias ─────────────────────────────────────────────────────────────────

func (s *CursosService) CreateLicencia(ctx context.Context, req *cursospb.CreateLicenciaRequest) (*cursospb.Licencia, error) {
	lic, err := s.repo.CreateLicencia(ctx, req)
	if err != nil {
		return nil, err
	}
	// Auto-create cohort group in mensajes-service
	if s.mensajes != nil {
		_, _ = s.mensajes.CreateGroupForLicencia(ctx, &mensajespb.CreateGroupForLicenciaRequest{
			LicenciaId: lic.ID,
			Nombre:     lic.Nombre + " — Grupo de Cohorte",
			AdminId:    req.InstructorId,
		})
	}
	return lic.ToProto(), nil
}

func (s *CursosService) UpdateLicencia(ctx context.Context, req *cursospb.UpdateLicenciaRequest) (*cursospb.Licencia, error) {
	lic, err := s.repo.UpdateLicencia(ctx, req)
	if err != nil {
		return nil, err
	}
	return lic.ToProto(), nil
}

func (s *CursosService) DeleteLicencia(ctx context.Context, id string) error {
	return s.repo.DeleteLicencia(ctx, id)
}

func (s *CursosService) ListLicencias(ctx context.Context, cursoID string) ([]*cursospb.Licencia, error) {
	lics, err := s.repo.ListLicencias(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	res := make([]*cursospb.Licencia, len(lics))
	for i, l := range lics {
		res[i] = l.ToProto()
	}
	return res, nil
}

func (s *CursosService) UnirseConLicencia(ctx context.Context, userID, capID, codigoAcceso string) error {
	lic, err := s.repo.FindLicenciaByCodigo(ctx, codigoAcceso)
	if err != nil {
		return errors.New("código de acceso inválido")
	}
	if lic.CapacitacionID != capID {
		return errors.New("el código no corresponde a esta capacitación")
	}
	if lic.CapacidadMaxima > 0 && lic.Usadas >= lic.CapacidadMaxima {
		return errors.New("la licencia ha alcanzado su capacidad máxima")
	}
	err = s.repo.InscribirseConLicencia(ctx, userID, capID, lic.ID)
	if err == nil {
		_ = s.repo.IncrementarUsoLicencia(ctx, lic.ID)
		// Auto-enrol in cohort group
		if s.mensajes != nil {
			_, _ = s.mensajes.EnrollInLicenciaGroup(ctx, &mensajespb.EnrollInLicenciaGroupRequest{
				LicenciaId: lic.ID,
				UserId:     userID,
			})
		}
	}
	return err
}

func (s *CursosService) WebhookEnroll(ctx context.Context, req *cursospb.WebhookEnrollRequest) (*cursospb.EmptyResponse, error) {
	err := s.repo.InscribirseConLicencia(ctx, req.UserId, req.CapacitacionId, req.LicenciaId)
	if err == nil && req.LicenciaId != "" {
		_ = s.repo.IncrementarUsoLicencia(ctx, req.LicenciaId)
	}
	if err == nil && req.ScheduleId != "" {
		_, _ = s.repo.UpdateSchedule(ctx, &cursospb.UpdateScheduleRequest{
			ScheduleId: req.ScheduleId,
			Status: "booked",
		})
	}

	// Fix B2C: Si es un curso en vivo, generar su ticket personal si no tiene licencia
	if err == nil && req.LicenciaId == "" {
		curso, errC := s.repo.FindByID(ctx, req.CapacitacionId)
		if errC == nil && curso.Type == "videocall" {
			var schID *string
			if req.ScheduleId != "" {
				schID = &req.ScheduleId
			}
			tickets, errT := s.repo.CreateVideocallTickets(ctx, req.CapacitacionId, nil, schID, 1)
			if errT == nil && len(tickets) > 0 {
				_ = s.repo.AssignTicketToUser(ctx, tickets[0].ID, req.UserId)
			}
		}
	}

	return &cursospb.EmptyResponse{}, err
}

func (s *CursosService) CreateCheckoutSession(ctx context.Context, req *cursospb.CheckoutSessionRequest) (*cursospb.CheckoutSessionResponse, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	var productName string
	var amount int64
	var clientRef string

	if req.CursoId != "" {
		// B2C Course Purchase
		curso, err := s.repo.FindByID(ctx, req.CursoId)
		if err != nil {
			return nil, err
		}
		if curso.Precio <= 0 {
			return nil, errors.New("el curso no tiene precio")
		}
		productName = curso.Title
		amount = int64(curso.Precio * 100)
		clientRef = "curso||" + req.UserId + "||" + curso.ID
		if req.ScheduleId != "" {
			clientRef += "||" + req.ScheduleId
		}
	} else {
		// B2B License Purchase
		lic, err := s.repo.FindLicenciaByID(ctx, req.LicenciaId)
		if err != nil {
			return nil, err
		}
		if lic.CapacidadMaxima > 0 && lic.Usadas >= lic.CapacidadMaxima {
			return nil, errors.New("licencia agotada")
		}
		productName = lic.Nombre
		amount = int64(lic.Precio * 100)
		clientRef = req.UserId + "||" + lic.CapacitacionID + "||" + lic.ID
	}

	// Crear sesión
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("mxn"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(productName),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(req.SuccessUrl),
		CancelURL:  stripe.String(req.CancelUrl),
		ClientReferenceID: stripe.String(clientRef),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled: stripe.Bool(true),
		},
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
		TaxIDCollection: &stripe.CheckoutSessionTaxIDCollectionParams{
			Enabled: stripe.Bool(true),
		},
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-user-email"); len(vals) > 0 && vals[0] != "" {
			params.CustomerEmail = stripe.String(vals[0])
		}
	}

	if stripe.Key == "" {
		return nil, errors.New("STRIPE_SECRET_KEY no está configurada en el servidor (cursos-service)")
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error de Stripe al crear sesión: %v", err)
		return nil, fmt.Errorf("error al conectar con el procesador de pagos: %v", err)
	}
	return &cursospb.CheckoutSessionResponse{Url: sess.URL}, nil
}

func (s *CursosService) WebhookComprarLicencia(ctx context.Context, req *cursospb.WebhookComprarLicenciaRequest) (*cursospb.EmptyResponse, error) {
	err := s.repo.AsignarCompradorLicencia(ctx, req.LicenciaId, req.UserId)
	return &cursospb.EmptyResponse{}, err
}

func (s *CursosService) GetLicenciaPublica(ctx context.Context, req *cursospb.LicenciaIDRequest) (*cursospb.LicenciaPublicaResponse, error) {
	lic, err := s.repo.FindLicenciaByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	curso, err := s.repo.FindByID(ctx, lic.CapacitacionID)
	if err != nil {
		return nil, err
	}
	return &cursospb.LicenciaPublicaResponse{
		Id:                   lic.ID,
		Nombre:               lic.Nombre,
		Precio:               lic.Precio,
		CapacidadMaxima:      lic.CapacidadMaxima,
		CapacitacionId:       curso.ID,
		CapacitacionTitulo:   curso.Title,
		CapacitacionThumbnail: curso.ThumbnailURL,
	}, nil
}

func (s *CursosService) ListLicenciasCompradas(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListLicenciasResponse, error) {
	lics, err := s.repo.ListLicenciasCompradas(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	var res []*cursospb.Licencia
	for _, l := range lics {
		res = append(res, l.ToProto())
	}
	return &cursospb.ListLicenciasResponse{Licencias: res}, nil
}

func (s *CursosService) CreateCheckoutSessionB2BDirect(ctx context.Context, req *cursospb.CreateCheckoutSessionB2BDirectRequest) (*cursospb.CheckoutSessionResponse, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	curso, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return nil, err
	}
	if curso.Precio <= 0 {
		return nil, errors.New("el curso no tiene precio")
	}

	productName := "Licencias Corporativas: " + curso.Title
	amount := int64(curso.Precio * float64(req.Cantidad) * 100)
	clientRef := "b2b_direct||" + req.UserId + "||" + curso.ID + "||" + fmt.Sprintf("%d", req.Cantidad)
	if req.ScheduleId != "" {
		clientRef += "||" + req.ScheduleId
	}

	// Crear sesión
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("mxn"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(productName),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(req.SuccessUrl),
		CancelURL:  stripe.String(req.CancelUrl),
		ClientReferenceID: stripe.String(clientRef),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled: stripe.Bool(true),
		},
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
		TaxIDCollection: &stripe.CheckoutSessionTaxIDCollectionParams{
			Enabled: stripe.Bool(true),
		},
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-user-email"); len(vals) > 0 && vals[0] != "" {
			params.CustomerEmail = stripe.String(vals[0])
		}
	}

	if stripe.Key == "" {
		return nil, errors.New("STRIPE_SECRET_KEY no está configurada en el servidor (cursos-service)")
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error de Stripe al crear sesión B2B directa: %v", err)
		return nil, fmt.Errorf("error al conectar con el procesador de pagos: %v", err)
	}
	return &cursospb.CheckoutSessionResponse{Url: sess.URL}, nil
}

func (s *CursosService) WebhookComprarB2BDirect(ctx context.Context, req *cursospb.WebhookComprarB2BDirectRequest) (*cursospb.EmptyResponse, error) {
	// Verificar que el curso existe
	curso, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return nil, err
	}
	
	precioTotal := curso.Precio * float64(req.Cantidad)
	
	// Crear licencia
	lic, err := s.repo.CreateLicenciaB2BDirect(ctx, req, precioTotal)
	if err != nil {
		return nil, err
	}
	
	// Si es videollamada, generar N tickets únicos en lugar de usar un solo código de licencia
	if curso.Type == "videocall" {
		var schID *string
		if req.ScheduleId != "" {
			schID = &req.ScheduleId
		}
		_, err = s.repo.CreateVideocallTickets(ctx, req.CursoId, &lic.ID, schID, int(req.Cantidad))
		if err != nil {
			return nil, err
		}
		if req.ScheduleId != "" {
			_, _ = s.repo.UpdateSchedule(ctx, &cursospb.UpdateScheduleRequest{
				ScheduleId: req.ScheduleId,
				Status: "booked",
			})
		}
	}
	
	return &cursospb.EmptyResponse{}, nil
}

func (s *CursosService) GetAdminDashboardStats(ctx context.Context) (*cursospb.AdminDashboardStatsResponse, error) {
	return s.repo.GetAdminDashboardStats(ctx)
}

func (s *CursosService) CreateCheckoutSessionCart(ctx context.Context, req *cursospb.CheckoutCartRequest) (*cursospb.CheckoutSessionResponse, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	if len(req.Items) == 0 {
		return nil, errors.New("el carrito está vacío")
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams
	metadataMap := make(map[string]string)

	for i, item := range req.Items {
		curso, err := s.repo.FindByID(ctx, item.CursoId)
		if err != nil {
			return nil, fmt.Errorf("curso no encontrado: %s", item.CursoId)
		}
		if curso.Precio <= 0 {
			return nil, fmt.Errorf("el curso %s no tiene precio", curso.Title)
		}

		var productName string
		var amount int64
		var quantity int64
		var itemMeta string

		if item.Type == "b2c" {
			productName = curso.Title
			amount = int64(curso.Precio * 100)
			quantity = 1
			itemMeta = fmt.Sprintf("b2c||%s||%s", curso.ID, item.ScheduleId)
		} else if item.Type == "b2b_direct" {
			productName = "Licencia Corporativa: " + curso.Title
			amount = int64(curso.Precio * 100)
			quantity = int64(item.Cantidad)
			itemMeta = fmt.Sprintf("b2b_direct||%s||%d||%s", curso.ID, item.Cantidad, item.ScheduleId)
		} else {
			return nil, fmt.Errorf("tipo de ítem no válido: %s", item.Type)
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("mxn"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(productName),
				},
				UnitAmount: stripe.Int64(amount),
			},
			Quantity: stripe.Int64(quantity),
		})

		// Stripe metadata limit is 50 keys
		metadataMap[fmt.Sprintf("item_%d", i)] = itemMeta
	}

	clientRef := "cart||" + req.UserId

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String(req.SuccessUrl),
		CancelURL:          stripe.String(req.CancelUrl),
		ClientReferenceID:  stripe.String(clientRef),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled: stripe.Bool(true),
		},
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
		TaxIDCollection: &stripe.CheckoutSessionTaxIDCollectionParams{
			Enabled: stripe.Bool(true),
		},
	}

	for k, v := range metadataMap {
		params.AddMetadata(k, v)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-user-email"); len(vals) > 0 && vals[0] != "" {
			params.CustomerEmail = stripe.String(vals[0])
		}
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error de Stripe al crear sesión de carrito: %v", err)
		return nil, fmt.Errorf("error al conectar con Stripe: %v", err)
	}

	return &cursospb.CheckoutSessionResponse{Url: sess.URL}, nil
}
