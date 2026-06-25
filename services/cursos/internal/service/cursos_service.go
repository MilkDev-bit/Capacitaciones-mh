package service

import (
	"context"
	"errors"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/repository"
	"os"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

// Errores de dominio.
var (
	ErrNotFound  = errors.New("no encontrado")
	ErrForbidden = errors.New("sin permisos")
	ErrConflict  = errors.New("ya inscrito")
)

// CursosService contiene la lógica de negocio del servicio de cursos.
type CursosService struct {
	repo repository.CursosRepository
}

func NewCursosService(repo repository.CursosRepository) *CursosService {
	return &CursosService{repo: repo}
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

func (s *CursosService) ListMisCapacitaciones(ctx context.Context, userID string) ([]*cursospb.CursoResponse, error) {
	cursos, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toProtoSlice(cursos), nil
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

func (s *CursosService) UnirseConLicencia(ctx context.Context, userID, capacitacionID, codigo string) error {
	lic, err := s.repo.FindLicenciaByCodigo(ctx, codigo)
	if err != nil {
		return err
	}
	if lic.CapacitacionID != capacitacionID {
		return ErrNotFound
	}
	if lic.CapacidadMaxima > 0 && lic.Usadas >= lic.CapacidadMaxima {
		return errors.New("licencia agotada")
	}
	err = s.repo.InscribirseConLicencia(ctx, userID, capacitacionID, lic.ID)
	if err == nil {
		_ = s.repo.IncrementarUsoLicencia(ctx, lic.ID)
	}
	return err
}

func (s *CursosService) WebhookEnroll(ctx context.Context, userID, capacitacionID, licenciaID string) error {
	// The webhook already verified payment, so we just enroll them.
	err := s.repo.InscribirseConLicencia(ctx, userID, capacitacionID, licenciaID)
	if err == nil {
		_ = s.repo.IncrementarUsoLicencia(ctx, licenciaID)
	}
	return err
}

func (s *CursosService) CreateCheckoutSession(ctx context.Context, req *cursospb.CheckoutSessionRequest) (*cursospb.CheckoutSessionResponse, error) {
	lic, err := s.repo.FindLicenciaByID(ctx, req.LicenciaId)
	if err != nil {
		return nil, err
	}
	if lic.CapacidadMaxima > 0 && lic.Usadas >= lic.CapacidadMaxima {
		return nil, errors.New("licencia agotada")
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Crear sesión
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(lic.Nombre),
					},
					UnitAmount: stripe.Int64(int64(lic.Precio * 100)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(req.SuccessUrl),
		CancelURL:  stripe.String(req.CancelUrl),
		ClientReferenceID: stripe.String(req.UserId + "||" + lic.CapacitacionID + "||" + lic.ID),
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, err
	}
	return &cursospb.CheckoutSessionResponse{Url: sess.URL}, nil
}
