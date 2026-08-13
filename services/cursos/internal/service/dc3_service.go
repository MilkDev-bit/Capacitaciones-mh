package service

// Lógica de las constancias DC-3.
//
// El servicio NO genera el documento: solo reúne y valida los datos. El armado
// del .docx vive en pkg/dc3 y lo invoca el Gateway, que es quien tiene acceso a
// R2 para guardarlo. Así este servicio sigue sin saber nada de ofimática ni de
// almacenamiento de objetos.

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/repository"
)

// ErrDC3NoHabilitado indica que la capacitación no emite constancias.
var ErrDC3NoHabilitado = errors.New("esta capacitación no emite constancia DC-3")

// GetDatosDC3 reúne los tres orígenes de la constancia y diagnostica qué falta.
//
// Devuelve datos incompletos en lugar de error cuando el instructor o el alumno
// no han capturado lo suyo: la generación es automática al terminar el curso y
// quien llama necesita distinguir "falta que el alumno ponga su CURP" de "falta
// que el instructor configure la empresa". Un error genérico dejaría al alumno
// mirando un mensaje que no puede resolver.
func (s *CursosService) GetDatosDC3(ctx context.Context, req *cursospb.DatosDC3Request) (*cursospb.DatosDC3Response, error) {
	curso, err := s.repo.FindByID(ctx, req.CapacitacionId)
	if err != nil {
		return nil, ErrNotFound
	}
	if !curso.DC3Enabled {
		return nil, ErrDC3NoHabilitado
	}

	trabajador, err := s.repo.FindDatosTrabajador(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	// Precedencia: manda el patrón que declara el alumno. Es el legalmente
	// correcto —el patrón es quien lo emplea— y el instructor solo cubre a quien
	// se capacita por su cuenta y no tiene empresa que lo respalde.
	//
	// La elección es en BLOQUE, nunca campo a campo: media empresa del alumno
	// más media del instructor daría un documento que no corresponde a ninguna
	// entidad real.
	empresa, origen := trabajador.EmpresaDelAlumno(), "alumno"
	if empresa == nil {
		instructorID := ""
		if curso.InstructorID != nil {
			instructorID = *curso.InstructorID
		}
		respaldo, errE := s.repo.FindEmpresaInstructor(ctx, instructorID)
		if errE != nil {
			return nil, errE
		}
		empresa, origen = respaldo.ToProto(), "instructor"
	} else {
		// El capacitador y el logo son SIEMPRE del instructor, incluso cuando el
		// patrón es el del alumno: quien imparte no cambia porque el trabajador
		// tenga empleador propio.
		if curso.InstructorID != nil {
			if respaldo, _ := s.repo.FindEmpresaInstructor(ctx, *curso.InstructorID); respaldo != nil {
				empresa.NombreCapacitador = respaldo.NombreCapacitador
				empresa.LogoBase64 = respaldo.LogoBase64
			}
		}
	}

	resp := &cursospb.DatosDC3Response{
		Empresa:            empresa,
		EmpresaOrigen:      origen,
		AreaTematica:       curso.DC3AreaTematica,
		Trabajador:         trabajador.ToProto(),
		NombreCurso:        curso.Title,
		DuracionHoras:      horasDeMinutos(curso.Duration),
		EmpresaCompleta:    empresaCompleta(empresa) && strings.TrimSpace(curso.DC3AreaTematica) != "",
		TrabajadorCompleto: trabajadorCompleto(trabajador),
	}

	// Periodo declarado: de la inscripción a hoy. Si no hay inscripción —el
	// acceso vino por asignación de RR.HH. o por suscripción— se usa la fecha
	// de creación del curso como inicio, que es lo más cercano defendible.
	inicio, err := s.repo.FechaInscripcion(ctx, req.UserId, req.CapacitacionId)
	if err != nil {
		return nil, err
	}
	if inicio.IsZero() {
		inicio = curso.CreatedAt
	}
	resp.FechaInicio = inicio.Format("2006-01-02")
	resp.FechaFin = time.Now().Format("2006-01-02")

	// Si ya se emitió, se devuelve la URL: quien llama no debe regenerarla y el
	// alumno ve la misma constancia siempre.
	constancia, err := s.repo.FindConstancia(ctx, req.UserId, req.CapacitacionId)
	if err != nil {
		return nil, err
	}
	if constancia != nil {
		resp.ConstanciaUrl = constancia.ArchivoURL
		// Se conserva la fecha con la que se emitió: reemitir el mismo documento
		// no debe cambiar el periodo que ya declaró la constancia entregada.
		resp.FechaFin = constancia.GeneradaAt.Format("2006-01-02")
	}

	return resp, nil
}

// GuardarDatosTrabajador valida y persiste lo que captura el alumno.
func (s *CursosService) GuardarDatosTrabajador(ctx context.Context, req *cursospb.DatosTrabajadorRequest) error {
	if req.Datos == nil {
		return errors.New("faltan los datos del trabajador")
	}
	curp := strings.ToUpper(strings.TrimSpace(req.Datos.Curp))
	if len(curp) != 18 {
		// Se valida la longitud y no el formato completo: la CURP tiene reglas
		// de dígito verificador que cambian con el tiempo, y rechazar una válida
		// dejaría al alumno sin constancia sin forma de saltárselo.
		return fmt.Errorf("la CURP debe tener 18 caracteres, se recibieron %d", len(curp))
	}
	puesto := strings.TrimSpace(req.Datos.Puesto)
	ocupacion := strings.TrimSpace(req.Datos.OcupacionEspecifica)
	if puesto == "" || ocupacion == "" {
		return errors.New("el puesto y la ocupación específica son obligatorios")
	}

	d := &repository.DatosTrabajadorDC3{
		UserID:              req.UserId,
		CURP:                curp,
		Puesto:              puesto,
		OcupacionEspecifica: ocupacion,
	}

	// El patrón del alumno se acepta entero o no se acepta. Guardar dos de los
	// cuatro campos dejaría un bloque que TieneEmpresa() rechaza igualmente, y
	// el alumno creería que ya declaró su empresa cuando no.
	if e := req.Empresa; e != nil {
		d.RazonSocial = strings.TrimSpace(e.RazonSocial)
		d.RFC = strings.ToUpper(strings.TrimSpace(e.Rfc))
		d.NombrePatron = strings.TrimSpace(e.NombrePatron)
		d.RepTrabajadores = strings.TrimSpace(e.RepresentanteTrabajadores)

		algunoLleno := d.RazonSocial != "" || d.RFC != "" || d.NombrePatron != "" || d.RepTrabajadores != ""
		if algunoLleno && !d.TieneEmpresa() {
			return errors.New("si declaras empresa, completa razón social, RFC, patrón y representante de los trabajadores")
		}
	}

	return s.repo.GuardarDatosTrabajador(ctx, d)
}

// GetEmpresaInstructor devuelve el respaldo configurado por el instructor.
func (s *CursosService) GetEmpresaInstructor(ctx context.Context, instructorID string) (*cursospb.DatosEmpresaDC3, error) {
	e, err := s.repo.FindEmpresaInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	return e.ToProto(), nil
}

// GuardarEmpresaInstructor valida y persiste el respaldo del instructor.
//
// Aquí sí se exige el bloque completo: es lo que va a firmar las constancias de
// todo alumno sin empresa propia, y guardarlo a medias produciría documentos
// inválidos sin que nadie se entere hasta que alguien los presente.
func (s *CursosService) GuardarEmpresaInstructor(ctx context.Context, req *cursospb.EmpresaInstructorRequest) error {
	if req.Empresa == nil {
		return errors.New("faltan los datos de la empresa")
	}
	e := &repository.EmpresaInstructorDC3{
		InstructorID:      req.InstructorId,
		RazonSocial:       strings.TrimSpace(req.Empresa.RazonSocial),
		RFC:               strings.ToUpper(strings.TrimSpace(req.Empresa.Rfc)),
		NombrePatron:      strings.TrimSpace(req.Empresa.NombrePatron),
		RepTrabajadores:   strings.TrimSpace(req.Empresa.RepresentanteTrabajadores),
		NombreCapacitador: strings.TrimSpace(req.Empresa.NombreCapacitador),
		LogoBase64:        req.Empresa.LogoBase64,
	}
	if e.RazonSocial == "" || e.RFC == "" || e.NombrePatron == "" ||
		e.RepTrabajadores == "" || e.NombreCapacitador == "" {
		return errors.New("razón social, RFC, patrón, representante de los trabajadores y capacitador son obligatorios")
	}
	return s.repo.GuardarEmpresaInstructor(ctx, e)
}

// RegistrarConstanciaDC3 guarda la URL del documento que generó el Gateway.
func (s *CursosService) RegistrarConstanciaDC3(ctx context.Context, req *cursospb.RegistrarConstanciaRequest) error {
	if req.ArchivoUrl == "" {
		return errors.New("la URL de la constancia es obligatoria")
	}
	return s.repo.RegistrarConstancia(ctx, req.UserId, req.CapacitacionId, req.ArchivoUrl)
}

// ListMisConstancias devuelve las constancias emitidas de un alumno.
func (s *CursosService) ListMisConstancias(ctx context.Context, userID string) (*cursospb.ListConstanciasResponse, error) {
	cs, err := s.repo.ListConstancias(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := &cursospb.ListConstanciasResponse{Constancias: make([]*cursospb.ConstanciaDC3, 0, len(cs))}
	for _, c := range cs {
		resp.Constancias = append(resp.Constancias, c.ToProto())
	}
	return resp, nil
}

// ── Ayudas ───────────────────────────────────────────────────────────────────

// empresaCompleta dice si la empresa resuelta sirve para emitir.
//
// El logo queda fuera a propósito: es opcional, la plantilla trae uno de fábrica.
func empresaCompleta(e *cursospb.DatosEmpresaDC3) bool {
	if e == nil {
		return false
	}
	for _, v := range []string{
		e.RazonSocial, e.Rfc, e.NombrePatron,
		e.RepresentanteTrabajadores, e.NombreCapacitador,
	} {
		if strings.TrimSpace(v) == "" {
			return false
		}
	}
	return true
}

func trabajadorCompleto(d *repository.DatosTrabajadorDC3) bool {
	return d != nil &&
		strings.TrimSpace(d.CURP) != "" &&
		strings.TrimSpace(d.Puesto) != "" &&
		strings.TrimSpace(d.OcupacionEspecifica) != ""
}

// horasDeMinutos traduce la duración del curso a las horas que declara la
// constancia, redondeando hacia arriba.
//
// Se redondea al alza porque la DC-3 declara horas de capacitación impartidas y
// un curso de 90 minutos son 2 horas de formación, no 1. Un curso sin duración
// registrada devuelve cadena vacía, que Faltantes() detecta y convierte en un
// aviso al instructor.
func horasDeMinutos(minutos int32) string {
	if minutos <= 0 {
		return ""
	}
	horas := (int(minutos) + 59) / 60
	return strconv.Itoa(horas)
}
