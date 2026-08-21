package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	cursospb "Prueba-Go/gen/cursos"
)

// ─────────────────────────────────────────────────────────────────────────────
// Panel financiero
//
// TODO sale de `ordenes` y `suscripcion_facturas`, que es lo que realmente se
// cobró. El panel anterior sumaba `capacitaciones.precio` cruzando con
// `inscripciones`, y eso está mal por tres motivos:
//
//   1. Cuenta como venta las altas que no pagaron nada: las que entran por
//      suscripción o por código de licencia aparecían al precio de catálogo.
//   2. Reescribe el pasado: subir el precio de un curso cambiaba la facturación
//      histórica, porque leía el precio de HOY y no el que se cobró entonces.
//   3. No sabe de reembolsos ni de cobros fallidos, que no descuenta.
//
// `ordenes` congela el importe al comprar y tiene estado, así que ninguna de
// las tres cosas ocurre aquí.
// ─────────────────────────────────────────────────────────────────────────────

// mesesSerie es la ventana del gráfico de tendencia.
const mesesSerie = 6

// maxTransaccionesRecientes acota la lista de movimientos de la pantalla.
const maxTransaccionesRecientes = 12

// estadosCobrados son los estados en los que el dinero entró de verdad.
//
// 'pendiente' es una ficha de OXXO sin pagar y 'fallida' un cobro rechazado:
// ninguno es una venta. 'reembolsada' se excluye porque el dinero se devolvió.
const estadosCobrados = `('pagada','cumplida')`

type FinanzasRepository interface {
	GetFinanzasAdmin(ctx context.Context) (*cursospb.FinanzasAdminResponse, error)
	AdminListLicenciasEmpresas(ctx context.Context) (*cursospb.AdminListLicenciasEmpresasResponse, error)
	ListOrdenesSinComision(ctx context.Context, limite int32) (*cursospb.ListOrdenesSinComisionResponse, error)
	RegistrarComisionOrden(ctx context.Context, ordenID string, com Comision) error
}

func (r *postgresCursosRepository) GetFinanzasAdmin(ctx context.Context) (*cursospb.FinanzasAdminResponse, error) {
	resp := &cursospb.FinanzasAdminResponse{Moneda: "MXN"}

	// ── Totales ──────────────────────────────────────────────────────────
	//
	// La comisión se suma con COALESCE a cero, pero las órdenes que aún no la
	// tienen se cuentan aparte en `sin_comision`. Así la pantalla puede decir
	// que el neto es un piso y no la cifra definitiva, en vez de presentar como
	// exacto un número al que le faltan comisiones por descontar.
	var tot struct {
		Bruto         int64 `db:"bruto"`
		Comision      int64 `db:"comision"`
		Transacciones int32 `db:"transacciones"`
		SinComision   int32 `db:"sin_comision"`
	}
	err := r.db.GetContext(ctx, &tot, `
		SELECT COALESCE(SUM(total_centavos), 0)                             AS bruto,
		       COALESCE(SUM(comision_centavos), 0)                          AS comision,
		       COUNT(*)                                                     AS transacciones,
		       COUNT(*) FILTER (WHERE comision_centavos IS NULL)            AS sin_comision
		  FROM ordenes
		 WHERE estado IN `+estadosCobrados)
	if err != nil {
		return nil, err
	}

	var fact struct {
		Bruto         int64 `db:"bruto"`
		Comision      int64 `db:"comision"`
		Transacciones int32 `db:"transacciones"`
		SinComision   int32 `db:"sin_comision"`
	}
	err = r.db.GetContext(ctx, &fact, `
		SELECT COALESCE(SUM(total_centavos), 0)                  AS bruto,
		       COALESCE(SUM(comision_centavos), 0)               AS comision,
		       COUNT(*)                                          AS transacciones,
		       COUNT(*) FILTER (WHERE comision_centavos IS NULL) AS sin_comision
		  FROM suscripcion_facturas
		 WHERE estado = 'pagada'`)
	if err != nil {
		return nil, err
	}

	resp.BrutoCentavos = tot.Bruto + fact.Bruto
	resp.ComisionCentavos = tot.Comision + fact.Comision
	resp.NetoCentavos = resp.BrutoCentavos - resp.ComisionCentavos
	resp.Transacciones = tot.Transacciones + fact.Transacciones
	resp.SinComision = tot.SinComision + fact.SinComision

	// ── Serie mensual ────────────────────────────────────────────────────
	//
	// generate_series y no un GROUP BY a secas: un mes sin ventas debe salir en
	// cero, no desaparecer del gráfico. Sin esto, la línea uniría marzo con
	// mayo y aparentaría una caída suave donde hubo un mes entero a cero.
	//
	// date_trunc va sobre la fecha de COBRO, no la de creación: una ficha de
	// OXXO se crea un día y se paga tres días después, y el ingreso pertenece
	// al mes en que entró el dinero.
	type filaMes struct {
		Mes      time.Time `db:"mes"`
		Bruto    int64     `db:"bruto"`
		Comision int64     `db:"comision"`
	}
	var filas []filaMes
	err = r.db.SelectContext(ctx, &filas, `
		WITH meses AS (
		    SELECT generate_series(
		        date_trunc('month', NOW()) - INTERVAL '`+itoa(mesesSerie-1)+` months',
		        date_trunc('month', NOW()),
		        INTERVAL '1 month'
		    ) AS mes
		),
		cobros AS (
		    SELECT date_trunc('month', COALESCE(pagada_at, created_at)) AS mes,
		           total_centavos,
		           COALESCE(comision_centavos, 0) AS comision_centavos
		      FROM ordenes
		     WHERE estado IN `+estadosCobrados+`
		    UNION ALL
		    SELECT date_trunc('month', created_at) AS mes,
		           total_centavos,
		           COALESCE(comision_centavos, 0) AS comision_centavos
		      FROM suscripcion_facturas
		     WHERE estado = 'pagada'
		)
		SELECT m.mes                                        AS mes,
		       COALESCE(SUM(c.total_centavos), 0)           AS bruto,
		       COALESCE(SUM(c.comision_centavos), 0)        AS comision
		  FROM meses m
		  LEFT JOIN cobros c ON c.mes = m.mes
		 GROUP BY m.mes
		 ORDER BY m.mes`)
	if err != nil {
		return nil, err
	}
	for _, f := range filas {
		resp.Serie = append(resp.Serie, &cursospb.PuntoMensual{
			Mes:              f.Mes.Format("2006-01"),
			BrutoCentavos:    f.Bruto,
			ComisionCentavos: f.Comision,
			NetoCentavos:     f.Bruto - f.Comision,
		})
	}

	// ── Movimientos recientes ────────────────────────────────────────────
	//
	// El nombre del cliente se resuelve contra `users`, que vive en la misma
	// base aunque la sirva otro servicio. El LEFT JOIN es deliberado: una
	// cuenta borrada no debe hacer desaparecer su compra del histórico
	// financiero.
	type filaMov struct {
		ID       string         `db:"id"`
		Cliente  sql.NullString `db:"cliente"`
		Email    sql.NullString `db:"email"`
		Fecha    time.Time      `db:"fecha"`
		Bruto    int64          `db:"bruto"`
		Comision sql.NullInt64  `db:"comision"`
		Origen   string         `db:"origen"`
		Concepto sql.NullString `db:"concepto"`
	}
	var movs []filaMov
	err = r.db.SelectContext(ctx, &movs, `
		SELECT o.id::text                       AS id,
		       u.name                           AS cliente,
		       u.email                          AS email,
		       COALESCE(o.pagada_at, o.created_at) AS fecha,
		       o.total_centavos                 AS bruto,
		       o.comision_centavos              AS comision,
		       'curso'                          AS origen,
		       (SELECT string_agg(c.title, ', ')
		          FROM orden_items oi
		          LEFT JOIN capacitaciones c ON c.id = oi.capacitacion_id
		         WHERE oi.orden_id = o.id)      AS concepto
		  FROM ordenes o
		  LEFT JOIN users u ON u.id = o.user_id
		 WHERE o.estado IN `+estadosCobrados+`
		 ORDER BY COALESCE(o.pagada_at, o.created_at) DESC
		 LIMIT $1`, maxTransaccionesRecientes)
	if err != nil {
		return nil, err
	}
	for _, m := range movs {
		comision := int64(0)
		if m.Comision.Valid {
			comision = m.Comision.Int64
		}
		resp.TransaccionesRecientes = append(resp.TransaccionesRecientes, &cursospb.TransaccionFin{
			Id:               m.ID,
			Cliente:          textoOr(m.Cliente, "Cuenta eliminada"),
			Email:            textoOr(m.Email, ""),
			Fecha:            m.Fecha.Format(time.RFC3339),
			BrutoCentavos:    m.Bruto,
			ComisionCentavos: comision,
			NetoCentavos:     m.Bruto - comision,
			ComisionConocida: m.Comision.Valid,
			Origen:           m.Origen,
			Concepto:         textoOr(m.Concepto, "Compra"),
		})
	}

	return resp, nil
}

func (r *postgresCursosRepository) AdminListLicenciasEmpresas(ctx context.Context) (*cursospb.AdminListLicenciasEmpresasResponse, error) {
	type fila struct {
		ID              string         `db:"id"`
		Nombre          string         `db:"nombre"`
		CapacitacionID  string         `db:"capacitacion_id"`
		CursoTitulo     sql.NullString `db:"capacitacion_titulo"`
		CompradorID     sql.NullString `db:"comprador_id"`
		CompradorNombre sql.NullString `db:"comprador_nombre"`
		CompradorEmail  sql.NullString `db:"comprador_email"`
		PrecioCentavos  int64          `db:"precio_centavos"`
		Capacidad       int32          `db:"capacidad_maxima"`
		Usadas          int32          `db:"usadas"`
		Codigo          sql.NullString `db:"codigo_acceso"`
		CreatedAt       time.Time      `db:"created_at"`
	}

	var filas []fila
	// Solo licencias con comprador: son las vendidas a una empresa. Las que no
	// lo tienen son plantillas que el instructor dejó creadas y todavía no ha
	// vendido nadie, y en una pantalla de "Empresas" solo harían ruido.
	err := r.db.SelectContext(ctx, &filas, `
		SELECT l.id::text                        AS id,
		       l.nombre                          AS nombre,
		       l.capacitacion_id::text           AS capacitacion_id,
		       c.title                           AS capacitacion_titulo,
		       l.comprador_id::text              AS comprador_id,
		       u.name                            AS comprador_nombre,
		       u.email                           AS comprador_email,
		       COALESCE(l.precio_centavos, 0)    AS precio_centavos,
		       l.capacidad_maxima                AS capacidad_maxima,
		       l.usadas                          AS usadas,
		       l.codigo_acceso                   AS codigo_acceso,
		       l.created_at                      AS created_at
		  FROM curso_licencias l
		  LEFT JOIN capacitaciones c ON c.id = l.capacitacion_id
		  LEFT JOIN users u          ON u.id = l.comprador_id
		 WHERE l.comprador_id IS NOT NULL
		 ORDER BY l.created_at DESC`)
	if err != nil {
		return nil, err
	}

	resp := &cursospb.AdminListLicenciasEmpresasResponse{}
	for _, f := range filas {
		resp.Licencias = append(resp.Licencias, &cursospb.AdminLicenciaEmpresa{
			Id:                 f.ID,
			Nombre:             f.Nombre,
			CapacitacionId:     f.CapacitacionID,
			CapacitacionTitulo: textoOr(f.CursoTitulo, "Curso eliminado"),
			CompradorId:        textoOr(f.CompradorID, ""),
			CompradorNombre:    textoOr(f.CompradorNombre, "Cuenta eliminada"),
			CompradorEmail:     textoOr(f.CompradorEmail, ""),
			PrecioCentavos:     f.PrecioCentavos,
			CapacidadMaxima:    f.Capacidad,
			Usadas:             f.Usadas,
			CodigoAcceso:       textoOr(f.Codigo, ""),
			CreatedAt:          f.CreatedAt.Format(time.RFC3339),
		})
		resp.TotalAsientos += f.Capacidad
		resp.AsientosUsados += f.Usadas
	}
	resp.TotalLicencias = int32(len(resp.Licencias))
	return resp, nil
}

func (r *postgresCursosRepository) ListOrdenesSinComision(ctx context.Context, limite int32) (*cursospb.ListOrdenesSinComisionResponse, error) {
	resp := &cursospb.ListOrdenesSinComisionResponse{}

	type fila struct {
		ID     string `db:"id"`
		Intent string `db:"stripe_payment_intent"`
	}
	var filas []fila
	// Las más antiguas primero: si el relleno se corta a medias, lo que queda
	// pendiente es lo más reciente, que es lo que el siguiente lote alcanza
	// antes.
	err := r.db.SelectContext(ctx, &filas, `
		SELECT id::text AS id, stripe_payment_intent
		  FROM ordenes
		 WHERE comision_centavos IS NULL
		   AND estado IN `+estadosCobrados+`
		   AND stripe_payment_intent IS NOT NULL
		 ORDER BY pagada_at ASC NULLS LAST
		 LIMIT $1`, limite)
	if err != nil {
		return nil, err
	}
	for _, f := range filas {
		resp.Ordenes = append(resp.Ordenes, &cursospb.OrdenSinComision{
			Id:                  f.ID,
			StripePaymentIntent: f.Intent,
		})
	}

	err = r.db.GetContext(ctx, &resp.Restantes, `
		SELECT COUNT(*)
		  FROM ordenes
		 WHERE comision_centavos IS NULL
		   AND estado IN `+estadosCobrados+`
		   AND stripe_payment_intent IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *postgresCursosRepository) RegistrarComisionOrden(ctx context.Context, ordenID string, com Comision) error {
	if ordenID == "" {
		return errors.New("orden_id vacío")
	}
	var tx interface{}
	if com.BalanceTxID != "" {
		tx = com.BalanceTxID
	}
	// Solo rellena lo que falta: el WHERE deja fuera las órdenes que ya la
	// tienen, de modo que llamar dos veces al relleno no reescribe nada.
	_, err := r.db.ExecContext(ctx, `
		UPDATE ordenes
		   SET comision_centavos      = $2,
		       neto_centavos          = $3,
		       balance_transaction_id = COALESCE($4, balance_transaction_id),
		       updated_at             = NOW()
		 WHERE id = $1 AND comision_centavos IS NULL`,
		ordenID, com.Centavos, com.NetoCentavos, tx)
	return err
}

// textoOr devuelve el valor de una columna que admite NULL, o un texto de
// respaldo. Existe para que una cuenta o un curso borrados no dejen huecos en
// la pantalla ni hagan desaparecer el movimiento del histórico.
func textoOr(v sql.NullString, respaldo string) string {
	if v.Valid && v.String != "" {
		return v.String
	}
	return respaldo
}

// itoa evita arrastrar strconv solo para interpolar una constante en el SQL.
// El valor nunca viene del usuario: es `mesesSerie`.
func itoa(n int) string {
	if n <= 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
