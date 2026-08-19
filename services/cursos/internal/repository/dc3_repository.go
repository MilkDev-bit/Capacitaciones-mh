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
	"strings"
	"time"

	cursospb "Prueba-Go/gen/cursos"
)

// DatosTrabajadorDC3 es lo que el alumno captura una sola vez.
//
// El bloque de empresa es opcional y va TODO junto: es el patrón del alumno.
// Quien no lo declara recibe la constancia a nombre de la empresa que configuró
// el instructor.
type DatosTrabajadorDC3 struct {
	UserID              string    `db:"user_id"`
	CURP                string    `db:"curp"`
	Puesto              string    `db:"puesto"`
	OcupacionEspecifica string    `db:"ocupacion_especifica"`
	ActualizadoAt       time.Time `db:"actualizado_at"`

	RazonSocial     string `db:"razon_social"`
	RFC             string `db:"rfc"`
	NombrePatron    string `db:"nombre_patron"`
	RepTrabajadores string `db:"representante_trabajadores"`
	// Logotipo de su empresa, para el lado izquierdo de la constancia. Solo se
	// usa si declara empresa propia: ese lado es del patrón.
	LogoURL string `db:"logo_url"`
}

// TieneEmpresa dice si el alumno declaró un patrón propio.
//
// Exige el bloque completo a propósito: con media empresa saldría una constancia
// que no corresponde a ninguna entidad real, y es preferible caer al respaldo
// del instructor —que sí está completo— que emitir un híbrido.
func (d *DatosTrabajadorDC3) TieneEmpresa() bool {
	if d == nil {
		return false
	}
	for _, v := range []string{d.RazonSocial, d.RFC, d.NombrePatron, d.RepTrabajadores} {
		if strings.TrimSpace(v) == "" {
			return false
		}
	}
	return true
}

// EmpresaDelAlumno devuelve el patrón declarado por el alumno.
func (d *DatosTrabajadorDC3) EmpresaDelAlumno() *cursospb.DatosEmpresaDC3 {
	if !d.TieneEmpresa() {
		return nil
	}
	return &cursospb.DatosEmpresaDC3{
		RazonSocial:               d.RazonSocial,
		Rfc:                       d.RFC,
		NombrePatron:              d.NombrePatron,
		RepresentanteTrabajadores: d.RepTrabajadores,
	}
}

// EmpresaInstructorDC3 es el respaldo que configura el instructor.
type EmpresaInstructorDC3 struct {
	InstructorID      string    `db:"instructor_id"`
	RazonSocial       string    `db:"razon_social"`
	RFC               string    `db:"rfc"`
	NombrePatron      string    `db:"nombre_patron"`
	RepTrabajadores   string    `db:"representante_trabajadores"`
	NombreCapacitador string    `db:"nombre_capacitador"`
	LogoURL           string    `db:"logo_url"`
	ActualizadoAt     time.Time `db:"actualizado_at"`
}

func (e *EmpresaInstructorDC3) ToProto() *cursospb.DatosEmpresaDC3 {
	if e == nil {
		return &cursospb.DatosEmpresaDC3{}
	}
	return &cursospb.DatosEmpresaDC3{
		RazonSocial:               e.RazonSocial,
		Rfc:                       e.RFC,
		NombrePatron:              e.NombrePatron,
		RepresentanteTrabajadores: e.RepTrabajadores,
		NombreCapacitador:         e.NombreCapacitador,
		LogoUrl:                   e.LogoURL,
	}
}

// ConstanciaDC3 es una constancia ya emitida.
type ConstanciaDC3 struct {
	UserID             string    `db:"user_id"`
	CapacitacionID     string    `db:"capacitacion_id"`
	CapacitacionTitulo string    `db:"capacitacion_titulo"`
	ArchivoURL         string    `db:"archivo_url"`
	GeneradaAt         time.Time `db:"generada_at"`
	// Folio impreso en el documento. Vacío en las constancias emitidas antes
	// de que existiera la verificación pública.
	Folio string `db:"folio"`
}

// ConstanciaEmitida son los datos con los que se registra una constancia.
//
// Va como struct y no como seis parámetros sueltos porque cinco de los seis son
// string: una llamada posicional invita a cruzar el folio con la URL sin que el
// compilador diga nada.
type ConstanciaEmitida struct {
	UserID         string
	CapacitacionID string
	ArchivoURL     string
	Folio          string
	// NombreTrabajador y RazonSocial se guardan tal como salieron impresos.
	NombreTrabajador string
	RazonSocial      string
}

// ConstanciaVerificada es lo que devuelve una consulta pública por folio.
//
// Es un tipo aparte y no ConstanciaDC3 a propósito: al no tener UserID ni
// ArchivoURL, no hay forma de que un descuido al serializar filtre el
// identificador del alumno o un enlace directo al documento de otra persona.
type ConstanciaVerificada struct {
	NombreTrabajador   string    `db:"nombre_trabajador"`
	CapacitacionTitulo string    `db:"capacitacion_titulo"`
	RazonSocial        string    `db:"razon_social"`
	GeneradaAt         time.Time `db:"generada_at"`
}

// FindDatosTrabajador devuelve los datos DC-3 del alumno.
//
// Devuelve (nil, nil) cuando todavía no los ha capturado: es el caso normal la
// primera vez, no un error, y quien llama lo traduce en "pídeselos".
func (r *postgresCursosRepository) FindDatosTrabajador(ctx context.Context, userID string) (*DatosTrabajadorDC3, error) {
	d := &DatosTrabajadorDC3{}
	err := r.db.GetContext(ctx, d,
		`SELECT user_id, curp, puesto, ocupacion_especifica, actualizado_at,
		        razon_social, rfc, nombre_patron, representante_trabajadores,
		        COALESCE(logo_url,'') logo_url
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
		`INSERT INTO dc3_datos_trabajador
		    (user_id, curp, puesto, ocupacion_especifica,
		     razon_social, rfc, nombre_patron, representante_trabajadores, logo_url, actualizado_at)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		 ON CONFLICT (user_id) DO UPDATE
		    SET curp = EXCLUDED.curp,
		        puesto = EXCLUDED.puesto,
		        ocupacion_especifica = EXCLUDED.ocupacion_especifica,
		        razon_social = EXCLUDED.razon_social,
		        rfc = EXCLUDED.rfc,
		        nombre_patron = EXCLUDED.nombre_patron,
		        representante_trabajadores = EXCLUDED.representante_trabajadores,
		        logo_url = EXCLUDED.logo_url,
		        actualizado_at = NOW()`,
		d.UserID, d.CURP, d.Puesto, d.OcupacionEspecifica,
		d.RazonSocial, d.RFC, d.NombrePatron, d.RepTrabajadores, d.LogoURL)
	return err
}

// FindEmpresaInstructor devuelve el respaldo configurado por un instructor, o
// (nil, nil) si todavía no lo ha capturado.
func (r *postgresCursosRepository) FindEmpresaInstructor(ctx context.Context, instructorID string) (*EmpresaInstructorDC3, error) {
	e := &EmpresaInstructorDC3{}
	err := r.db.GetContext(ctx, e,
		`SELECT instructor_id, razon_social, rfc, nombre_patron,
		        representante_trabajadores, nombre_capacitador, logo_url, actualizado_at
		   FROM dc3_empresa_instructor WHERE instructor_id = $1::uuid`, instructorID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

// GuardarEmpresaInstructor crea o actualiza el respaldo del instructor.
func (r *postgresCursosRepository) GuardarEmpresaInstructor(ctx context.Context, e *EmpresaInstructorDC3) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dc3_empresa_instructor
		    (instructor_id, razon_social, rfc, nombre_patron,
		     representante_trabajadores, nombre_capacitador, logo_url, actualizado_at)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT (instructor_id) DO UPDATE
		    SET razon_social = EXCLUDED.razon_social,
		        rfc = EXCLUDED.rfc,
		        nombre_patron = EXCLUDED.nombre_patron,
		        representante_trabajadores = EXCLUDED.representante_trabajadores,
		        nombre_capacitador = EXCLUDED.nombre_capacitador,
		        logo_url = EXCLUDED.logo_url,
		        actualizado_at = NOW()`,
		e.InstructorID, e.RazonSocial, e.RFC, e.NombrePatron,
		e.RepTrabajadores, e.NombreCapacitador, e.LogoURL)
	return err
}

// FindConstancia devuelve la constancia emitida de un alumno para un curso, o
// (nil, nil) si aún no existe.
func (r *postgresCursosRepository) FindConstancia(ctx context.Context, userID, capacitacionID string) (*ConstanciaDC3, error) {
	c := &ConstanciaDC3{}
	err := r.db.GetContext(ctx, c,
		`SELECT k.user_id, k.capacitacion_id, k.archivo_url, k.generada_at,
		        COALESCE(k.folio,'') folio,
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
func (r *postgresCursosRepository) RegistrarConstancia(ctx context.Context, c *ConstanciaEmitida) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dc3_constancias (user_id, capacitacion_id, archivo_url, folio,
		                              nombre_trabajador, razon_social, generada_at)
		 VALUES ($1::uuid, $2::uuid, $3, NULLIF($4,''), $5, $6, NOW())
		 ON CONFLICT (user_id, capacitacion_id) DO UPDATE
		    SET archivo_url       = EXCLUDED.archivo_url,
		        folio             = EXCLUDED.folio,
		        nombre_trabajador = EXCLUDED.nombre_trabajador,
		        razon_social      = EXCLUDED.razon_social,
		        generada_at       = NOW()`,
		c.UserID, c.CapacitacionID, c.ArchivoURL, c.Folio, c.NombreTrabajador, c.RazonSocial)
	return err
}

// VerificarConstancia busca una constancia por su folio.
//
// Devuelve (nil, nil) cuando no existe: un folio inventado no es un error del
// servidor, es la respuesta que busca quien está comprobando un documento.
//
// El nombre y la razón social se leen de la propia fila, congelados el día de
// la emisión. No se consultan los datos actuales del alumno por dos razones:
// si cambia de empleo la constancia debe seguir diciendo lo que dice el papel,
// y este servicio tiene base de datos propia —la tabla de usuarios vive en la
// de auth, así que un JOIN contra ella ni siquiera resolvería—.
func (r *postgresCursosRepository) VerificarConstancia(ctx context.Context, folio string) (*ConstanciaVerificada, error) {
	c := &ConstanciaVerificada{}
	err := r.db.GetContext(ctx, c,
		`SELECT COALESCE(k.nombre_trabajador,'') nombre_trabajador,
		        COALESCE(k.razon_social,'')      razon_social,
		        COALESCE(cap.title,'')           capacitacion_titulo,
		        k.generada_at
		   FROM dc3_constancias k
		   LEFT JOIN capacitaciones cap ON cap.id = k.capacitacion_id
		  WHERE k.folio = $1`, folio)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListConstancias devuelve las constancias de un alumno, de la más reciente a
// la más antigua.
func (r *postgresCursosRepository) ListConstancias(ctx context.Context, userID string) ([]*ConstanciaDC3, error) {
	var cs []*ConstanciaDC3
	return cs, r.db.SelectContext(ctx, &cs,
		`SELECT k.user_id, k.capacitacion_id, k.archivo_url, k.generada_at,
		        COALESCE(k.folio,'') folio,
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
		LogoUrl:             d.LogoURL,
	}
}

// ToProto convierte una constancia emitida al contrato gRPC.
func (c *ConstanciaDC3) ToProto() *cursospb.ConstanciaDC3 {
	return &cursospb.ConstanciaDC3{
		CapacitacionId:     c.CapacitacionID,
		CapacitacionTitulo: c.CapacitacionTitulo,
		ArchivoUrl:         c.ArchivoURL,
		GeneradaAt:         c.GeneradaAt.Format(time.RFC3339),
		Folio:              c.Folio,
	}
}
