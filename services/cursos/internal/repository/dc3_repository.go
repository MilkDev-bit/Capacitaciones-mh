package repository

// Acceso a datos de las constancias DC-3.
//
// Vive en cursos-service y no en usuarios-service aunque parte de los datos sean
// del trabajador: la constancia se emite POR CURSO, y aquí están la
// capacitación, la inscripción y sus fechas. Cruzar de servicio para leer un
// CURP obligaría a un viaje extra en el camino crítico de generación.

import (
	"context"
	"database/sql"
	"errors"
	"time"

	cursospb "Prueba-Go/gen/cursos"
)

// DatosTrabajadorDC3 es lo que el alumno captura una sola vez.
type DatosTrabajadorDC3 struct {
	UserID              string    `db:"user_id"`
	CURP                string    `db:"curp"`
	Puesto              string    `db:"puesto"`
	OcupacionEspecifica string    `db:"ocupacion_especifica"`
	ActualizadoAt       time.Time `db:"actualizado_at"`
}

// ConstanciaDC3 es una constancia ya emitida.
type ConstanciaDC3 struct {
	UserID             string    `db:"user_id"`
	CapacitacionID     string    `db:"capacitacion_id"`
	CapacitacionTitulo string    `db:"capacitacion_titulo"`
	ArchivoURL         string    `db:"archivo_url"`
	GeneradaAt         time.Time `db:"generada_at"`
}

// FindDatosTrabajador devuelve los datos DC-3 del alumno.
//
// Devuelve (nil, nil) cuando todavía no los ha capturado: es el caso normal la
// primera vez, no un error, y quien llama lo traduce en "pídeselos".
func (r *postgresCursosRepository) FindDatosTrabajador(ctx context.Context, userID string) (*DatosTrabajadorDC3, error) {
	d := &DatosTrabajadorDC3{}
	err := r.db.GetContext(ctx, d,
		`SELECT user_id, curp, puesto, ocupacion_especifica, actualizado_at
		   FROM dc3_datos_trabajador WHERE user_id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

// GuardarDatosTrabajador crea o actualiza los datos del alumno.
//
// Es un upsert porque el alumno puede corregirlos —un CURP mal tecleado invalida
// la constancia— y porque los reutiliza en cada curso posterior.
func (r *postgresCursosRepository) GuardarDatosTrabajador(ctx context.Context, d *DatosTrabajadorDC3) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dc3_datos_trabajador (user_id, curp, puesto, ocupacion_especifica, actualizado_at)
		 VALUES ($1::uuid, $2, $3, $4, NOW())
		 ON CONFLICT (user_id) DO UPDATE
		    SET curp = EXCLUDED.curp,
		        puesto = EXCLUDED.puesto,
		        ocupacion_especifica = EXCLUDED.ocupacion_especifica,
		        actualizado_at = NOW()`,
		d.UserID, d.CURP, d.Puesto, d.OcupacionEspecifica)
	return err
}

// FindConstancia devuelve la constancia emitida de un alumno para un curso, o
// (nil, nil) si aún no existe.
func (r *postgresCursosRepository) FindConstancia(ctx context.Context, userID, capacitacionID string) (*ConstanciaDC3, error) {
	c := &ConstanciaDC3{}
	err := r.db.GetContext(ctx, c,
		`SELECT k.user_id, k.capacitacion_id, k.archivo_url, k.generada_at,
		        COALESCE(c.title,'') capacitacion_titulo
		   FROM dc3_constancias k
		   LEFT JOIN capacitaciones c ON c.id = k.capacitacion_id
		  WHERE k.user_id = $1::uuid AND k.capacitacion_id = $2::uuid`, userID, capacitacionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// RegistrarConstancia guarda la URL del documento emitido.
//
// El upsert sobre la PK compuesta es lo que hace idempotente la generación
// automática: si el alumno vuelve a completar el curso, se sustituye la fila en
// lugar de acumular constancias duplicadas del mismo curso.
func (r *postgresCursosRepository) RegistrarConstancia(ctx context.Context, userID, capacitacionID, archivoURL string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dc3_constancias (user_id, capacitacion_id, archivo_url, generada_at)
		 VALUES ($1::uuid, $2::uuid, $3, NOW())
		 ON CONFLICT (user_id, capacitacion_id) DO UPDATE
		    SET archivo_url = EXCLUDED.archivo_url,
		        generada_at = NOW()`,
		userID, capacitacionID, archivoURL)
	return err
}

// ListConstancias devuelve las constancias de un alumno, de la más reciente a
// la más antigua.
func (r *postgresCursosRepository) ListConstancias(ctx context.Context, userID string) ([]*ConstanciaDC3, error) {
	var cs []*ConstanciaDC3
	return cs, r.db.SelectContext(ctx, &cs,
		`SELECT k.user_id, k.capacitacion_id, k.archivo_url, k.generada_at,
		        COALESCE(c.title,'') capacitacion_titulo
		   FROM dc3_constancias k
		   LEFT JOIN capacitaciones c ON c.id = k.capacitacion_id
		  WHERE k.user_id = $1::uuid
		  ORDER BY k.generada_at DESC`, userID)
}

// FechaInscripcion devuelve cuándo se inscribió el alumno al curso.
//
// Es la fecha de inicio que declara la constancia. Si no hay inscripción —el
// acceso vino por asignación de RR.HH. o por suscripción— devuelve cero y quien
// llama decide el respaldo.
func (r *postgresCursosRepository) FechaInscripcion(ctx context.Context, userID, capacitacionID string) (time.Time, error) {
	var t time.Time
	err := r.db.GetContext(ctx, &t,
		`SELECT inscrito_at FROM inscripciones
		  WHERE user_id = $1::uuid AND capacitacion_id = $2::uuid`, userID, capacitacionID)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	return t, err
}

// ToProto convierte los datos del trabajador al contrato gRPC.
func (d *DatosTrabajadorDC3) ToProto() *cursospb.DatosTrabajadorDC3 {
	if d == nil {
		return &cursospb.DatosTrabajadorDC3{}
	}
	return &cursospb.DatosTrabajadorDC3{
		Curp:                d.CURP,
		Puesto:              d.Puesto,
		OcupacionEspecifica: d.OcupacionEspecifica,
	}
}

// ToProto convierte una constancia emitida al contrato gRPC.
func (c *ConstanciaDC3) ToProto() *cursospb.ConstanciaDC3 {
	return &cursospb.ConstanciaDC3{
		CapacitacionId:     c.CapacitacionID,
		CapacitacionTitulo: c.CapacitacionTitulo,
		ArchivoUrl:         c.ArchivoURL,
		GeneradaAt:         c.GeneradaAt.Format(time.RFC3339),
	}
}
