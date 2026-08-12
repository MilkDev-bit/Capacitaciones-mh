package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Suscripciones
//
// El ciclo de vida lo manda Stripe, no nosotros: cada estado llega por webhook
// y aquí solo se aplica. Inferirlo localmente (por ejemplo, "si periodo_fin ya
// pasó entonces está cancelada") desincroniza el sistema del PSP en cuanto hay
// un reintento de cobro exitoso.
//
//	incompleta ──3DS/acción──▶ en_prueba ──fin de prueba──▶ activa
//	                                                          │  ▲
//	                                    cobro falla ──────────┘  │ el cobro se recupera
//	                                          ▼                  │
//	                                       vencida ──────────────┘
//	                                          │ dunning agotado
//	                                          ▼
//	                                    impagada / cancelada
//
// 'vencida' (past_due en Stripe) NO revoca el acceso: es el periodo de gracia
// durante el que corren los reintentos. Cortar ahí genera churn involuntario
// de gente que solo tenía la tarjeta vencida.
// ─────────────────────────────────────────────────────────────────────────────

// Estados de una suscripción.
const (
	SuscIncompleta = "incompleta"
	SuscEnPrueba   = "en_prueba"
	SuscActiva     = "activa"
	SuscVencida    = "vencida"
	SuscCancelada  = "cancelada"
	SuscImpagada   = "impagada"
)

// estadosConAcceso son aquellos en los que el usuario sigue entrando.
// 'vencida' está incluido a propósito: es el periodo de gracia.
var estadosConAcceso = []string{SuscEnPrueba, SuscActiva, SuscVencida}

// EstadoDaAcceso indica si un estado concede acceso al contenido.
func EstadoDaAcceso(estado string) bool {
	for _, e := range estadosConAcceso {
		if e == estado {
			return true
		}
	}
	return false
}

// EstadoDesdeStripe traduce el estado del PSP al vocabulario del dominio.
func EstadoDesdeStripe(s string) string {
	switch s {
	case "trialing":
		return SuscEnPrueba
	case "active":
		return SuscActiva
	case "past_due":
		return SuscVencida
	case "unpaid":
		return SuscImpagada
	case "canceled", "incomplete_expired":
		return SuscCancelada
	case "incomplete":
		return SuscIncompleta
	default:
		return SuscIncompleta
	}
}

var (
	ErrPlanNoEncontrado        = errors.New("plan no encontrado")
	ErrSuscripcionNoEncontrada = errors.New("suscripción no encontrada")
	ErrSinAsientosLibres       = errors.New("no quedan asientos libres en la suscripción")
	ErrYaTieneSuscripcion      = errors.New("ya tienes una suscripción activa")
)

// Plan es un producto de suscripción.
type Plan struct {
	ID             string  `db:"id"`
	Codigo         string  `db:"codigo"`
	Nombre         string  `db:"nombre"`
	Descripcion    string  `db:"descripcion"`
	Modalidad      string  `db:"modalidad"`
	Intervalo      string  `db:"intervalo"`
	PrecioCentavos int64   `db:"precio_centavos"`
	Moneda         string  `db:"moneda"`
	StripePriceID  *string `db:"stripe_price_id"`
	DiasPrueba     int32   `db:"dias_prueba"`
	Activo         bool    `db:"activo"`
	Orden          int32   `db:"orden"`
}

// Suscripcion es el contrato vivo de un titular.
type Suscripcion struct {
	ID                   string     `db:"id"`
	UserID               string     `db:"user_id"`
	PlanID               string     `db:"plan_id"`
	Estado               string     `db:"estado"`
	Asientos             int32      `db:"asientos"`
	StripeSubscriptionID *string    `db:"stripe_subscription_id"`
	StripeCustomerID     *string    `db:"stripe_customer_id"`
	PeriodoInicio        *time.Time `db:"periodo_inicio"`
	PeriodoFin           *time.Time `db:"periodo_fin"`
	PruebaFin            *time.Time `db:"prueba_fin"`
	CancelarAlTerminar   bool       `db:"cancelar_al_terminar"`
	CanceladaAt          *time.Time `db:"cancelada_at"`

	// Derivados del JOIN con planes.
	PlanNombre         *string `db:"plan_nombre"`
	PlanModalidad      *string `db:"plan_modalidad"`
	PlanIntervalo      *string `db:"plan_intervalo"`
	PlanPrecioCentavos *int64  `db:"plan_precio_centavos"`
}

// Asiento es un lugar de una suscripción corporativa.
type Asiento struct {
	ID            string     `db:"id"`
	SuscripcionID string     `db:"suscripcion_id"`
	UserID        *string    `db:"user_id"`
	Email         string     `db:"email"`
	Estado        string     `db:"estado"`
	InvitadoAt    time.Time  `db:"invitado_at"`
	ActivadoAt    *time.Time `db:"activado_at"`
}

// ── Contrato ─────────────────────────────────────────────────────────────────

// SuscripcionesRepository agrupa el acceso a datos de planes y suscripciones.
type SuscripcionesRepository interface {
	ListPlanesActivos(ctx context.Context) ([]*Plan, error)
	FindPlanPorCodigo(ctx context.Context, codigo string) (*Plan, error)

	FindSuscripcionDeUsuario(ctx context.Context, userID string) (*Suscripcion, error)
	FindSuscripcionPorStripeID(ctx context.Context, stripeSubID string) (*Suscripcion, error)
	FindSuscripcionPorID(ctx context.Context, id string) (*Suscripcion, error)
	// UpsertSuscripcion aplica el estado que reporta Stripe.
	UpsertSuscripcion(ctx context.Context, s *Suscripcion, planCodigo string) error
	RegistrarFactura(ctx context.Context, suscripcionID, stripeInvoiceID, estado string, totalCentavos int64, moneda string, intento int32, urlPDF string, ini, fin *time.Time) error

	// AccesoPorSuscripcion resuelve si el usuario entra por titularidad o asiento.
	AccesoPorSuscripcion(ctx context.Context, userID string) (tiene bool, origen, suscripcionID, estado string, err error)

	AsignarAsientos(ctx context.Context, suscripcionID, titularID string, participantes []Participante) ([]*Asiento, error)
	ListAsientos(ctx context.Context, suscripcionID, titularID string) ([]*Asiento, error)
	RevocarAsiento(ctx context.Context, suscripcionID, titularID, email string) error
	ContarAsientosOcupados(ctx context.Context, suscripcionID string) (int32, error)
}

// ── Planes ───────────────────────────────────────────────────────────────────

func (r *postgresCursosRepository) ListPlanesActivos(ctx context.Context) ([]*Plan, error) {
	var planes []*Plan
	return planes, r.db.SelectContext(ctx, &planes,
		`SELECT id, codigo, nombre, descripcion, modalidad, intervalo, precio_centavos,
		        moneda, stripe_price_id, dias_prueba, activo, orden
		   FROM planes WHERE activo = true ORDER BY orden, precio_centavos`)
}

func (r *postgresCursosRepository) FindPlanPorCodigo(ctx context.Context, codigo string) (*Plan, error) {
	p := &Plan{}
	err := r.db.GetContext(ctx, p,
		`SELECT id, codigo, nombre, descripcion, modalidad, intervalo, precio_centavos,
		        moneda, stripe_price_id, dias_prueba, activo, orden
		   FROM planes WHERE codigo = $1`, codigo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPlanNoEncontrado
	}
	return p, err
}

// ── Suscripciones ────────────────────────────────────────────────────────────

const selectSuscripcion = `
	SELECT s.id, s.user_id, s.plan_id, s.estado, s.asientos,
	       s.stripe_subscription_id, s.stripe_customer_id,
	       s.periodo_inicio, s.periodo_fin, s.prueba_fin,
	       s.cancelar_al_terminar, s.cancelada_at,
	       p.nombre AS plan_nombre, p.modalidad AS plan_modalidad,
	       p.intervalo AS plan_intervalo, p.precio_centavos AS plan_precio_centavos
	  FROM suscripciones s
	  JOIN planes p ON p.id = s.plan_id`

// FindSuscripcionDeUsuario devuelve la suscripción viva del titular.
// Devuelve nil sin error si no tiene ninguna: no tener suscripción es normal.
func (r *postgresCursosRepository) FindSuscripcionDeUsuario(ctx context.Context, userID string) (*Suscripcion, error) {
	s := &Suscripcion{}
	err := r.db.GetContext(ctx, s, selectSuscripcion+`
	 WHERE s.user_id = $1
	   AND s.estado IN ('incompleta','en_prueba','activa','vencida')
	 ORDER BY s.created_at DESC LIMIT 1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

func (r *postgresCursosRepository) FindSuscripcionPorStripeID(ctx context.Context, stripeSubID string) (*Suscripcion, error) {
	s := &Suscripcion{}
	err := r.db.GetContext(ctx, s, selectSuscripcion+` WHERE s.stripe_subscription_id = $1`, stripeSubID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSuscripcionNoEncontrada
	}
	return s, err
}

func (r *postgresCursosRepository) FindSuscripcionPorID(ctx context.Context, id string) (*Suscripcion, error) {
	s := &Suscripcion{}
	err := r.db.GetContext(ctx, s, selectSuscripcion+` WHERE s.id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSuscripcionNoEncontrada
	}
	return s, err
}

// UpsertSuscripcion aplica el estado que llega de Stripe.
//
// Se identifica por stripe_subscription_id, no por user_id: un usuario puede
// haber tenido varias suscripciones a lo largo del tiempo y solo la que Stripe
// nombra debe moverse.
func (r *postgresCursosRepository) UpsertSuscripcion(ctx context.Context, s *Suscripcion, planCodigo string) error {
	if s.StripeSubscriptionID == nil || *s.StripeSubscriptionID == "" {
		return errors.New("falta stripe_subscription_id")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op tras Commit

	var existenteID string
	err = tx.GetContext(ctx, &existenteID,
		`SELECT id FROM suscripciones WHERE stripe_subscription_id = $1 FOR UPDATE`,
		*s.StripeSubscriptionID)

	switch {
	case err == nil:
		// Actualización: solo se toca lo que Stripe reporta.
		if _, err := tx.ExecContext(ctx,
			`UPDATE suscripciones
			    SET estado=$2, asientos=$3, periodo_inicio=$4, periodo_fin=$5,
			        prueba_fin=$6, cancelar_al_terminar=$7,
			        cancelada_at = CASE WHEN $2 = 'cancelada' THEN COALESCE(cancelada_at, NOW()) ELSE NULL END,
			        stripe_customer_id = COALESCE($8, stripe_customer_id),
			        updated_at=NOW()
			  WHERE id=$1`,
			existenteID, s.Estado, s.Asientos, s.PeriodoInicio, s.PeriodoFin,
			s.PruebaFin, s.CancelarAlTerminar, s.StripeCustomerID); err != nil {
			return err
		}
		s.ID = existenteID
		return tx.Commit()

	case errors.Is(err, sql.ErrNoRows):
		// Alta: hace falta saber a qué plan y titular pertenece.
		if planCodigo == "" || s.UserID == "" {
			return errors.New("alta de suscripción sin plan o titular")
		}
		var planID string
		if err := tx.GetContext(ctx, &planID,
			`SELECT id FROM planes WHERE codigo = $1`, planCodigo); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrPlanNoEncontrado
			}
			return err
		}

		// El índice único parcial impide dos suscripciones vivas por usuario.
		// Si el alta choca, es que ya tenía una: se reporta como tal.
		err := tx.GetContext(ctx, &s.ID,
			`INSERT INTO suscripciones
			   (user_id, plan_id, estado, asientos, stripe_subscription_id, stripe_customer_id,
			    periodo_inicio, periodo_fin, prueba_fin, cancelar_al_terminar)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
			s.UserID, planID, s.Estado, s.Asientos, s.StripeSubscriptionID, s.StripeCustomerID,
			s.PeriodoInicio, s.PeriodoFin, s.PruebaFin, s.CancelarAlTerminar)
		if err != nil {
			if strings.Contains(err.Error(), "uq_suscripcion_viva_por_usuario") {
				return ErrYaTieneSuscripcion
			}
			return err
		}
		return tx.Commit()

	default:
		return err
	}
}

func (r *postgresCursosRepository) RegistrarFactura(
	ctx context.Context, suscripcionID, stripeInvoiceID, estado string,
	totalCentavos int64, moneda string, intento int32, urlPDF string, ini, fin *time.Time,
) error {
	// ON CONFLICT: Stripe reenvía el mismo invoice al reintentar el cobro, con
	// el contador de intentos actualizado.
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO suscripcion_facturas
		   (suscripcion_id, stripe_invoice_id, estado, total_centavos, moneda,
		    intento_cobro, url_pdf, periodo_inicio, periodo_fin)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 ON CONFLICT (stripe_invoice_id) DO UPDATE
		    SET estado=EXCLUDED.estado,
		        total_centavos=EXCLUDED.total_centavos,
		        intento_cobro=EXCLUDED.intento_cobro,
		        url_pdf=COALESCE(EXCLUDED.url_pdf, suscripcion_facturas.url_pdf)`,
		suscripcionID, stripeInvoiceID, estado, totalCentavos, moneda, intento, nilSiVacio(urlPDF), ini, fin)
	return err
}

func nilSiVacio(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// ── Acceso ───────────────────────────────────────────────────────────────────

// AccesoPorSuscripcion resuelve el acceso en una sola consulta.
//
// Se comprueba primero la titularidad y luego los asientos: si alguien es
// titular de un plan y además ocupa un asiento ajeno, manda su propio plan.
func (r *postgresCursosRepository) AccesoPorSuscripcion(ctx context.Context, userID string) (bool, string, string, string, error) {
	var fila struct {
		Origen string `db:"origen"`
		ID     string `db:"id"`
		Estado string `db:"estado"`
	}
	err := r.db.GetContext(ctx, &fila,
		`SELECT 'titular' AS origen, s.id, s.estado
		   FROM suscripciones s
		  WHERE s.user_id = $1 AND s.estado IN ('en_prueba','activa','vencida')
		 UNION ALL
		 SELECT 'asiento', s.id, s.estado
		   FROM suscripcion_asientos a
		   JOIN suscripciones s ON s.id = a.suscripcion_id
		  WHERE a.user_id = $1 AND a.estado = 'activo'
		    AND s.estado IN ('en_prueba','activa','vencida')
		 LIMIT 1`, userID)

	if errors.Is(err, sql.ErrNoRows) {
		return false, "", "", "", nil
	}
	if err != nil {
		return false, "", "", "", err
	}
	return true, fila.Origen, fila.ID, fila.Estado, nil
}

// ── Asientos ─────────────────────────────────────────────────────────────────

func (r *postgresCursosRepository) ContarAsientosOcupados(ctx context.Context, suscripcionID string) (int32, error) {
	var n int32
	err := r.db.GetContext(ctx, &n,
		`SELECT count(*) FROM suscripcion_asientos
		  WHERE suscripcion_id = $1 AND estado <> 'revocado'`, suscripcionID)
	return n, err
}

// AsignarAsientos reparte lugares validando el cupo dentro de la transacción.
//
// El SELECT ... FOR UPDATE sobre la suscripción serializa los repartos: sin él,
// dos peticiones simultáneas podrían rebasar el número de asientos pagados.
func (r *postgresCursosRepository) AsignarAsientos(
	ctx context.Context, suscripcionID, titularID string, participantes []Participante,
) ([]*Asiento, error) {
	if len(participantes) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op tras Commit

	var s struct {
		Asientos int32  `db:"asientos"`
		UserID   string `db:"user_id"`
		Estado   string `db:"estado"`
	}
	if err := tx.GetContext(ctx, &s,
		`SELECT asientos, user_id, estado FROM suscripciones WHERE id=$1 FOR UPDATE`,
		suscripcionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSuscripcionNoEncontrada
		}
		return nil, err
	}
	// Solo el titular reparte los asientos que paga.
	if s.UserID != titularID {
		return nil, errForbidden
	}
	if !EstadoDaAcceso(s.Estado) {
		return nil, errors.New("la suscripción no está activa")
	}

	// Los correos ya invitados no consumen un asiento nuevo.
	yaInvitados := map[string]bool{}
	var previos []string
	if err := tx.SelectContext(ctx, &previos,
		`SELECT email FROM suscripcion_asientos
		  WHERE suscripcion_id=$1 AND estado <> 'revocado'`, suscripcionID); err != nil {
		return nil, err
	}
	for _, e := range previos {
		yaInvitados[strings.ToLower(e)] = true
	}

	nuevos := 0
	for _, p := range participantes {
		if !yaInvitados[strings.ToLower(p.Email)] {
			nuevos++
		}
	}
	if disponibles := int(s.Asientos) - len(previos); nuevos > disponibles {
		return nil, ErrSinAsientosLibres
	}

	for _, p := range participantes {
		email := strings.ToLower(strings.TrimSpace(p.Email))
		if email == "" {
			continue
		}
		// Al reinvitar a alguien revocado se reactiva su asiento.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO suscripcion_asientos (suscripcion_id, email, estado)
			 VALUES ($1,$2,'invitado')
			 ON CONFLICT (suscripcion_id, email) DO UPDATE
			    SET estado = CASE WHEN suscripcion_asientos.estado = 'revocado'
			                      THEN 'invitado' ELSE suscripcion_asientos.estado END,
			        invitado_at = NOW()`,
			suscripcionID, email); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.ListAsientos(ctx, suscripcionID, titularID)
}

func (r *postgresCursosRepository) ListAsientos(ctx context.Context, suscripcionID, titularID string) ([]*Asiento, error) {
	var propietario string
	if err := r.db.GetContext(ctx, &propietario,
		`SELECT user_id FROM suscripciones WHERE id=$1`, suscripcionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSuscripcionNoEncontrada
		}
		return nil, err
	}
	if propietario != titularID {
		return nil, errForbidden
	}

	var asientos []*Asiento
	return asientos, r.db.SelectContext(ctx, &asientos,
		`SELECT id, suscripcion_id, user_id, email, estado, invitado_at, activado_at
		   FROM suscripcion_asientos
		  WHERE suscripcion_id=$1 AND estado <> 'revocado'
		  ORDER BY invitado_at`, suscripcionID)
}

// RevocarAsiento libera un lugar. No se borra la fila: se marca 'revocado' para
// conservar el historial de quién tuvo acceso y cuándo.
func (r *postgresCursosRepository) RevocarAsiento(ctx context.Context, suscripcionID, titularID, email string) error {
	var propietario string
	if err := r.db.GetContext(ctx, &propietario,
		`SELECT user_id FROM suscripciones WHERE id=$1`, suscripcionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSuscripcionNoEncontrada
		}
		return err
	}
	if propietario != titularID {
		return errForbidden
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE suscripcion_asientos SET estado='revocado', user_id=NULL
		  WHERE suscripcion_id=$1 AND LOWER(email)=LOWER($2)`, suscripcionID, email)
	return err
}
