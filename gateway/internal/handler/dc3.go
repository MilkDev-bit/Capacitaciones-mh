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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/clients"
	"Prueba-Go/gateway/internal/middleware"
	"Prueba-Go/gateway/internal/pdf"
	"Prueba-Go/gateway/internal/storage"
	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/pkg/dc3"

	"github.com/gin-gonic/gin"
)

// tipoPDF es el MIME de la constancia que se entrega. R2 lo devuelve como
// Content-Type y es lo que decide si el navegador la muestra o la descarga.
//
// El .docx ya no sale de aquí: entregarlo equivalía a dar la plantilla oficial
// con los campos sustituidos, lista para cambiarle el nombre y la empresa sin
// haber comprado ningún curso.
const tipoPDF = "application/pdf"

// alfabetoFolio evita los caracteres que se confunden al teclear un código
// leído en papel: I/1, O/0, S/5, B/8.
const alfabetoFolio = "ACDEFGHJKLMNPQRTUVWXYZ2346789"

// nuevoFolio genera el código único que identifica la constancia.
//
// Usa crypto/rand y no math/rand: un folio adivinable permitiría montar una
// constancia falsa y darle un código que la página de verificación aceptase,
// que es exactamente lo que este mecanismo debe impedir.
//
// 12 caracteres sobre un alfabeto de 29 son unos 58 bits: de sobra para que
// probar folios al azar no sea una vía práctica.
func nuevoFolio() string {
	const n = 12
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Sin entropía no se emite. Un folio predecible es peor que ninguno,
		// porque el documento parecería verificable sin serlo.
		return ""
	}
	out := make([]byte, n)
	for i, v := range b {
		out[i] = alfabetoFolio[int(v)%len(alfabetoFolio)]
	}
	return "MH-" + string(out[:4]) + "-" + string(out[4:8]) + "-" + string(out[8:])
}

// pieDeFolio arma la línea que se imprime al final de la constancia.
//
// Se incluye la URL completa para que quien reciba el papel sepa dónde
// comprobarlo sin tener que preguntar. Si no hay URL pública configurada se
// imprime solo el folio: mejor un código sin instrucciones que una dirección
// inventada que no lleve a ninguna parte.
func pieDeFolio(folio string) string {
	if folio == "" {
		return ""
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/")
	if base == "" {
		return "Folio de verificación: " + folio
	}
	return "Verifica esta constancia en " + base + "/verificar  ·  Folio: " + folio
}

// ErrConstanciaIncompleta se devuelve cuando faltan datos por capturar.
var ErrConstanciaIncompleta = errors.New("faltan datos para emitir la constancia")

// ErrIncompleta detalla QUÉ falta y a quién le toca ponerlo.
//
// El error plano no bastaba y el precio fue alto: `emitir` lo devolvía tanto
// cuando faltaban datos del alumno como cuando faltaba algo del curso, y quien
// lo recibía tenía que inventarse la causa. La respuesta 202 acabó afirmando
// "tu instructor debe completar los datos de la empresa" en casos donde la
// empresa estaba completa y lo que faltaba era otra cosa, mandando a la gente a
// revisar justo donde no estaba el problema.
//
// Ahora la lista viaja hasta el cliente y nadie tiene que adivinar.
type ErrIncompleta struct {
	// Faltan son los campos vacíos, con el nombre que entiende el usuario.
	Faltan []string
	// DeAlumno indica si lo que falta lo puede resolver el propio alumno.
	DeAlumno bool
}

func (e *ErrIncompleta) Error() string {
	return "faltan datos para emitir la constancia: " + strings.Join(e.Faltan, ", ")
}

// Is permite seguir usando errors.Is(err, ErrConstanciaIncompleta) en las
// ramas que solo necesitan saber que faltó algo, sin romper a quien ya lo hacía.
func (e *ErrIncompleta) Is(target error) bool { return target == ErrConstanciaIncompleta }

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
		// Para que el formulario pueda mostrar el logotipo ya guardado en vez de
		// pedirlo otra vez en cada corrección.
		"logo_patron_url": datos.LogoPatronUrl,
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
		// Logotipo de su empresa, ya subido a R2. Solo se aplica si declara
		// empresa propia: es el lado izquierdo del documento, el del patrón.
		LogoURL string `json:"logo_url"`
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
			LogoUrl:             body.LogoURL,
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

	// reemitir=true: el alumno acaba de enviar sus datos, así que si ya había
	// constancia hay que rehacerla con lo recién capturado. Devolverle la
	// anterior sería ignorar en silencio la corrección que acaba de hacer.
	url, err := h.emitir(ctx.Request.Context(), userID, cursoID,
		ctx.GetString(middleware.CtxUserName), true)
	var incompleta *ErrIncompleta
	if errors.As(err, &incompleta) {
		// 202: lo del alumno quedó guardado, pero el documento no salió.
		//
		// El mensaje ya no presume la causa. Antes decía siempre "tu instructor
		// debe completar los datos de la empresa", incluso cuando la empresa
		// estaba completa y lo que faltaba era la duración del curso: el alumno
		// reclamaba por algo que su instructor ya había hecho.
		ctx.JSON(http.StatusAccepted, gin.H{
			"guardado": true,
			"faltan":   incompleta.Faltan,
			"mensaje": "Guardamos tus datos, pero aún no se puede emitir la constancia. Falta: " +
				strings.Join(incompleta.Faltan, "; ") + ".",
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
// VerificarConstancia expone la comprobación pública de un folio.
//
// Es el ÚNICO endpoint DC-3 sin sesión, así que devuelve lo mínimo: a nombre de
// quién, de qué curso y de cuándo. Sin CURP, sin RFC, sin correo y sin enlace al
// documento —quien teclea un folio ajeno no debe llevarse el expediente de esa
// persona ni una descarga directa—.
func (h *DC3Handler) VerificarConstancia(ctx *gin.Context) {
	resp, err := h.c.Cursos.VerificarConstancia(ctx.Request.Context(),
		&cursospb.VerificarConstanciaRequest{Folio: ctx.Param("folio")})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	// 200 también cuando no es válida: que el folio no exista es una respuesta
	// legítima de este endpoint, no un error de la petición.
	if !resp.Valida {
		ctx.JSON(http.StatusOK, gin.H{"valida": false})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"valida":              true,
		"nombre_trabajador":   resp.NombreTrabajador,
		"capacitacion_titulo": resp.CapacitacionTitulo,
		"razon_social":        resp.RazonSocial,
		"generada_at":         resp.GeneradaAt,
	})
}

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

		// reemitir=false: la emisión automática se dispara al completar el curso
		// y puede repetirse —reintentos del webhook, el alumno rehaciendo una
		// lección—. Regenerar en cada disparo subiría un documento nuevo a R2
		// cada vez, con folio distinto, invalidando el que ya tenga en la mano.
		url, err := h.emitir(ctx, userID, cursoID, nombre, false)
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
// emitir arma la constancia y la sube. Devuelve la URL del documento.
//
// `reemitir` decide qué hacer si ya existe una:
//
//	false → se devuelve la existente. Es lo que protege la emisión automática
//	        de los reintentos del webhook y de que el alumno recorra el curso
//	        otra vez: sin esto se acumularían documentos duplicados.
//	true  → se genera de nuevo. Lo pide quien acaba de cambiar algo que va
//	        impreso en el papel.
//
// La distinción faltaba y el precio era alto en los dos sentidos. Un alumno que
// corregía una CURP mal tecleada seguía descargando la constancia con la CURP
// equivocada —y una DC-3 con la CURP mal es inválida—. Y al arreglar la
// plantilla, las constancias ya emitidas se quedaban congeladas con el formato
// roto, sin ninguna forma de regenerarlas.
func (h *DC3Handler) emitir(ctx context.Context, userID, cursoID, nombreTrabajador string, reemitir bool) (string, error) {
	datos, err := h.c.Cursos.GetDatosDC3(ctx, &cursospb.DatosDC3Request{
		UserId:         userID,
		CapacitacionId: cursoID,
	})
	if err != nil {
		return "", err
	}
	if datos.ConstanciaUrl != "" && !reemitir {
		return datos.ConstanciaUrl, nil
	}
	if !datos.EmpresaCompleta || !datos.TrabajadorCompleto {
		var faltan []string
		if !datos.TrabajadorCompleto {
			faltan = append(faltan, "tus datos (CURP, puesto y ocupación)")
		}
		if !datos.EmpresaCompleta {
			faltan = append(faltan, "los datos de la empresa y el área temática del curso")
		}
		return "", &ErrIncompleta{Faltan: faltan, DeAlumno: !datos.TrabajadorCompleto}
	}

	// El folio se genera antes que el documento porque va impreso dentro de él.
	folio := nuevoFolio()

	d := dc3.Datos{
		NombreTrabajador:    nombreTrabajador,
		CURP:                datos.Trabajador.Curp,
		OcupacionEspecifica: datos.Trabajador.OcupacionEspecifica,
		Puesto:              datos.Trabajador.Puesto,

		RazonSocial:         datos.Empresa.RazonSocial,
		RFC:                 datos.Empresa.Rfc,
		NombrePatron:        datos.Empresa.NombrePatron,
		NombreRepresentante: datos.Empresa.RepresentanteTrabajadores,

		// La precedencia entre alumno e instructor ya la resolvió cursos-service;
		// aquí solo se descargan las dos imágenes que decidió.
		LogoPatron:      h.descargarLogo(ctx, datos.LogoPatronUrl),
		LogoCapacitador: h.descargarLogo(ctx, datos.LogoCapacitadorUrl),

		NombreCurso:       datos.NombreCurso,
		DuracionHoras:     datos.DuracionHoras,
		AreaTematica:      datos.AreaTematica,
		NombreCapacitador: datos.Empresa.NombreCapacitador,
		FechaInicio:       datos.FechaInicio,
		FechaFin:          datos.FechaFin,

		Folio: pieDeFolio(folio),
	}

	// La segunda validación no es redundante: los banderines de arriba miran los
	// datos capturados, y esto mira el documento. Un curso sin duración registrada
	// pasa los banderines y aun así no puede emitir constancia.
	if faltan := d.Faltantes(); len(faltan) > 0 {
		slog.Info("DC-3: datos insuficientes", "user_id", userID, "curso_id", cursoID, "faltan", faltan)
		// Estos campos no los captura el alumno: salen del curso y del perfil del
		// instructor. Devolver la lista es lo que evita el mensaje genérico que
		// mandaba a revisar la empresa cuando el hueco estaba en otro sitio.
		return "", &ErrIncompleta{Faltan: faltan}
	}

	doc, err := dc3.Generar(d)
	if err != nil {
		return "", err
	}

	// El .docx no llega a salir de aquí: se convierte y solo se sube el PDF.
	//
	// Si la conversión falla se aborta la emisión en lugar de subir el Word.
	// Un respaldo silencioso reintroduciría el problema justo cuando nadie mira,
	// y de forma intermitente, que es la peor manera de tener un agujero.
	pdfBytes, err := pdf.Convertir(ctx, "constancia.docx", doc)
	if err != nil {
		// Se registra la URL de destino junto al error.
		//
		// Sin ella, "no se pudo convertir" no distingue entre un nombre de
		// servicio que no resuelve, un puerto equivocado y una conversión que
		// falló de verdad. Los tres se ven igual en el log y cada uno se
		// arregla en un sitio distinto.
		slog.Error("DC-3: no se pudo convertir a PDF",
			"user_id", userID, "curso_id", cursoID,
			"gotenberg_url", pdf.URL(), "error", err)
		return "", fmt.Errorf("convirtiendo la constancia a PDF: %w", err)
	}

	nombreArchivo := dc3.NombreArchivo(nombreTrabajador, datos.NombreCurso)
	url, err := storage.Default().UploadFile(ctx,
		"constancias/"+userID+"/"+nombreArchivo, tipoPDF, bytes.NewReader(pdfBytes), int64(len(pdfBytes)))
	if err != nil {
		return "", err
	}

	if _, err := h.c.Cursos.RegistrarConstanciaDC3(ctx, &cursospb.RegistrarConstanciaRequest{
		UserId:         userID,
		CapacitacionId: cursoID,
		ArchivoUrl:     url,
		Folio:          folio,
		// Se guardan los mismos valores que se acaban de imprimir, no los que
		// tenga el perfil cuando alguien verifique el documento meses después.
		NombreTrabajador: nombreTrabajador,
		RazonSocial:      datos.Empresa.RazonSocial,
	}); err != nil {
		// El documento ya está en R2. Se devuelve la URL igualmente para no
		// perderla, pero se registra el fallo: sin la fila en base, el alumno no
		// la vería en su listado y se regeneraría en el siguiente intento.
		slog.Error("DC-3: subida correcta pero no se registró", "url", url, "error", err)
		return url, nil
	}
	return url, nil
}
