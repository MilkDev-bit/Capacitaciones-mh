package handler

// Emisión de constancias DC-3.
//
// El reparto de responsabilidades es deliberado:
//
//	cursos-service → reúne y valida los datos (empresa, trabajador, fechas)
//	pkg/dc3        → arma el .docx sobre la plantilla oficial
//	este fichero   → orquesta, sube a R2 y registra el resultado
//
// El Gateway hace de pegamento porque es el único con acceso a R2; cursos-service
// no sabe nada de almacenamiento de objetos y no tiene por qué.

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	"Prueba-Go/gateway/internal/storage"
	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/pkg/dc3"

	"github.com/gin-gonic/gin"
)

// tipoDocx es el MIME de un .docx. R2 lo devuelve como Content-Type y es lo que
// hace que el navegador descargue el archivo en vez de intentar mostrarlo.
const tipoDocx = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

// ErrConstanciaIncompleta se devuelve cuando faltan datos por capturar.
var ErrConstanciaIncompleta = errors.New("faltan datos para emitir la constancia")

type DC3Handler struct {
	c *clients.Clients
}

func NewDC3Handler(c *clients.Clients) *DC3Handler { return &DC3Handler{c: c} }

// GET /api/capacitaciones/:id/dc3
//
// Devuelve el estado de la constancia del alumno para ese curso: si ya existe,
// qué falta y quién debe completarlo. Es lo que alimenta el formulario.
func (h *DC3Handler) EstadoConstancia(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxUserID)
	datos, err := h.c.Cursos.GetDatosDC3(ctx.Request.Context(), &cursospb.DatosDC3Request{
		UserId:         userID,
		CapacitacionId: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"constancia_url": datos.ConstanciaUrl,
		// El frontend necesita saber a quién pedirle los datos: al alumno le
		// muestra el formulario, y si lo que falta es la empresa le dice que
		// avise a su instructor, porque él no puede hacer nada al respecto.
		"trabajador_completo": datos.TrabajadorCompleto,
		"empresa_completa":    datos.EmpresaCompleta,
		"trabajador":          datos.Trabajador,
		"nombre_curso":        datos.NombreCurso,
		// A nombre de quién saldrá la constancia y de dónde salió ese dato. Es
		// lo que necesita el alumno para decidir si declara su propio patrón.
		"empresa":        datos.Empresa,
		"empresa_origen": datos.EmpresaOrigen,
	})
}

// GET /api/instructor/dc3-empresa
//
// Datos de empresa que el instructor deja configurados como respaldo para los
// alumnos que no declaran patrón propio.
func (h *DC3Handler) GetEmpresaInstructor(ctx *gin.Context) {
	resp, err := h.c.Cursos.GetEmpresaInstructor(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// PUT /api/instructor/dc3-empresa
func (h *DC3Handler) GuardarEmpresaInstructor(ctx *gin.Context) {
	var body struct {
		RazonSocial       string `json:"razon_social"`
		RFC               string `json:"rfc"`
		NombrePatron      string `json:"nombre_patron"`
		RepTrabajadores   string `json:"representante_trabajadores"`
		NombreCapacitador string `json:"nombre_capacitador"`
		// URL del logo ya subido a R2 por el frontend, como avatares y portadas.
		LogoURL string `json:"logo_url"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "revisa los datos capturados"})
		return
	}

	if _, err := h.c.Cursos.GuardarEmpresaInstructor(ctx.Request.Context(), &cursospb.EmpresaInstructorRequest{
		InstructorId: ctx.GetString(middleware.CtxUserID),
		Empresa: &cursospb.DatosEmpresaDC3{
			RazonSocial:               body.RazonSocial,
			Rfc:                       body.RFC,
			NombrePatron:              body.NombrePatron,
			RepresentanteTrabajadores: body.RepTrabajadores,
			NombreCapacitador:         body.NombreCapacitador,
			LogoUrl:                   body.LogoURL,
		},
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// POST /api/capacitaciones/:id/dc3
//
// El alumno manda sus datos y, si con eso ya está todo, se emite la constancia
// en la misma petición. Devolver la URL al momento evita que tenga que recargar
// o esperar sin saber si funcionó.
func (h *DC3Handler) GuardarDatosYEmitir(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxUserID)
	cursoID := ctx.Param("id")

	var body struct {
		CURP                string `json:"curp"`
		Puesto              string `json:"puesto"`
		OcupacionEspecifica string `json:"ocupacion_especifica"`
		// Patrón del alumno. Opcional: quien no tenga empresa lo deja vacío y la
		// constancia sale a nombre de la que configuró el instructor.
		RazonSocial     string `json:"razon_social"`
		RFC             string `json:"rfc"`
		NombrePatron    string `json:"nombre_patron"`
		RepTrabajadores string `json:"representante_trabajadores"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "revisa los datos capturados"})
		return
	}

	if _, err := h.c.Cursos.GuardarDatosTrabajador(ctx.Request.Context(), &cursospb.DatosTrabajadorRequest{
		UserId: userID,
		Datos: &cursospb.DatosTrabajadorDC3{
			Curp:                body.CURP,
			Puesto:              body.Puesto,
			OcupacionEspecifica: body.OcupacionEspecifica,
		},
		Empresa: &cursospb.DatosEmpresaDC3{
			RazonSocial:               body.RazonSocial,
			Rfc:                       body.RFC,
			NombrePatron:              body.NombrePatron,
			RepresentanteTrabajadores: body.RepTrabajadores,
		},
	}); err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	url, err := h.emitir(ctx.Request.Context(), userID, cursoID,
		ctx.GetString(middleware.CtxUserName))
	if errors.Is(err, ErrConstanciaIncompleta) {
		// 202: lo del alumno quedó guardado, pero falta que el instructor
		// configure la empresa. No es un fallo suyo y no debe leerse como tal.
		ctx.JSON(http.StatusAccepted, gin.H{
			"guardado": true,
			"mensaje":  "Guardamos tus datos. Tu instructor debe completar los datos de la empresa para emitir la constancia.",
		})
		return
	}
	if err != nil {
		slog.Error("DC-3: fallo al emitir", "user_id", userID, "curso_id", cursoID, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo generar la constancia"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"constancia_url": url})
}

// GET /api/mis-constancias
func (h *DC3Handler) ListMisConstancias(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListMisConstancias(ctx.Request.Context(), &cursospb.UserRequest{
		UserId: ctx.GetString(middleware.CtxUserID),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Constancias)
}

// EmitirEnSegundoPlano intenta generar la constancia sin bloquear a quien llama.
//
// Es el punto de entrada desde la finalización del curso. Si faltan datos, no
// hace ruido: el alumno verá el formulario la próxima vez que abra el curso.
// Un fallo aquí NUNCA debe impedir que se marque la lección como completada.
func (h *DC3Handler) EmitirEnSegundoPlano(userID, cursoID, nombre string) {
	go func() {
		ctx, cancel := contextoCorto()
		defer cancel()

		url, err := h.emitir(ctx, userID, cursoID, nombre)
		switch {
		case errors.Is(err, ErrConstanciaIncompleta):
			slog.Info("DC-3: constancia pendiente de datos",
				"user_id", userID, "curso_id", cursoID)
		case err != nil:
			slog.Error("DC-3: no se pudo emitir automáticamente",
				"user_id", userID, "curso_id", cursoID, "error", err)
		default:
			slog.Info("DC-3: constancia emitida", "user_id", userID, "curso_id", cursoID, "url", url)
			notificar(h.c, aviso{
				UserID:  userID,
				Tipo:    TipoConstancia,
				Titulo:  "Tu constancia DC-3 está lista",
				Mensaje: "Ya puedes descargarla desde tus capacitaciones.",
				Enlace:  "/usuario/capacitaciones/" + cursoID,
				Ventana: ventanaCompra,
			})
		}
	}()
}

// logoMaxBytes acota lo que se descarga para incrustar en la constancia.
//
// El logo va DENTRO del .docx, así que su peso es el peso del documento que
// recibe cada alumno. La plantilla trae uno de ~1.8 MB; 4 MB da margen de
// sobra sin permitir que una imagen enorme infle todas las constancias.
const logoMaxBytes = 4 << 20

// descargarLogo trae el logo de R2 para incrustarlo.
//
// Devuelve nil ante cualquier problema, y eso es deliberado: sin logo la
// constancia sale con el de la plantilla, que es un documento válido. Fallar la
// emisión entera porque una imagen decorativa no cargó sería desproporcionado.
func (h *DC3Handler) descargarLogo(ctx context.Context, url string) []byte {
	if url == "" {
		return nil
	}
	// Solo se descarga de NUESTRO bucket.
	//
	// La URL llega en el cuerpo de una petición, así que un instructor podría
	// mandar cualquier dirección y convertir al gateway en su cliente HTTP
	// —incluso contra servicios internos de la red privada de Railway, que no
	// son alcanzables desde fuera—. Comprobar el prefijo lo impide.
	base := storage.Default().PublicURL()
	if base == "" || !strings.HasPrefix(url, base+"/") {
		slog.Warn("DC-3: logo con URL ajena a R2, se ignora", "url", url)
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("DC-3: no se pudo descargar el logo", "url", url, "error", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Warn("DC-3: el logo respondió con error", "url", url, "status", resp.StatusCode)
		return nil
	}

	// LimitReader y no ContentLength: la cabecera la controla el servidor y
	// puede mentir o faltar.
	img, err := io.ReadAll(io.LimitReader(resp.Body, logoMaxBytes+1))
	if err != nil || len(img) == 0 || len(img) > logoMaxBytes {
		slog.Warn("DC-3: logo inválido o demasiado grande", "url", url, "bytes", len(img))
		return nil
	}
	return img
}

// emitir hace el trabajo de verdad: reúne, genera, sube y registra.
//
// Es idempotente por diseño: si cursos-service ya tiene una constancia para ese
// par (alumno, curso), devuelve su URL sin regenerar nada. Eso evita que una
// relectura de lecciones o un reintento produzcan documentos duplicados con
// fechas distintas.
func (h *DC3Handler) emitir(ctx context.Context, userID, cursoID, nombreTrabajador string) (string, error) {
	datos, err := h.c.Cursos.GetDatosDC3(ctx, &cursospb.DatosDC3Request{
		UserId:         userID,
		CapacitacionId: cursoID,
	})
	if err != nil {
		return "", err
	}
	if datos.ConstanciaUrl != "" {
		return datos.ConstanciaUrl, nil
	}
	if !datos.EmpresaCompleta || !datos.TrabajadorCompleto {
		return "", ErrConstanciaIncompleta
	}

	d := dc3.Datos{
		NombreTrabajador:    nombreTrabajador,
		CURP:                datos.Trabajador.Curp,
		OcupacionEspecifica: datos.Trabajador.OcupacionEspecifica,
		Puesto:              datos.Trabajador.Puesto,

		RazonSocial:         datos.Empresa.RazonSocial,
		RFC:                 datos.Empresa.Rfc,
		NombrePatron:        datos.Empresa.NombrePatron,
		NombreRepresentante: datos.Empresa.RepresentanteTrabajadores,
		Logo:                h.descargarLogo(ctx, datos.Empresa.LogoUrl),

		NombreCurso:       datos.NombreCurso,
		DuracionHoras:     datos.DuracionHoras,
		AreaTematica:      datos.AreaTematica,
		NombreCapacitador: datos.Empresa.NombreCapacitador,
		FechaInicio:       datos.FechaInicio,
		FechaFin:          datos.FechaFin,
	}

	// La segunda validación no es redundante: los banderines de arriba miran los
	// datos capturados, y esto mira el documento. Un curso sin duración registrada
	// pasa los banderines y aun así no puede emitir constancia.
	if faltan := d.Faltantes(); len(faltan) > 0 {
		slog.Info("DC-3: datos insuficientes", "user_id", userID, "curso_id", cursoID, "faltan", faltan)
		return "", ErrConstanciaIncompleta
	}

	doc, err := dc3.Generar(d)
	if err != nil {
		return "", err
	}

	nombreArchivo := dc3.NombreArchivo(nombreTrabajador, datos.NombreCurso)
	url, err := storage.Default().UploadFile(ctx,
		"constancias/"+userID+"/"+nombreArchivo, tipoDocx, bytes.NewReader(doc), int64(len(doc)))
	if err != nil {
		return "", err
	}

	if _, err := h.c.Cursos.RegistrarConstanciaDC3(ctx, &cursospb.RegistrarConstanciaRequest{
		UserId:         userID,
		CapacitacionId: cursoID,
		ArchivoUrl:     url,
	}); err != nil {
		// El documento ya está en R2. Se devuelve la URL igualmente para no
		// perderla, pero se registra el fallo: sin la fila en base, el alumno no
		// la vería en su listado y se regeneraría en el siguiente intento.
		slog.Error("DC-3: subida correcta pero no se registró", "url", url, "error", err)
		return url, nil
	}
	return url, nil
}
