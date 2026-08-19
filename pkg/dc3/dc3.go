// Package dc3 genera la constancia DC-3 (Constancia de Competencias o de
// Habilidades Laborales, formato STPS) rellenando la plantilla oficial.
//
// La plantilla va embebida en el binario con go:embed. Antes vivía como fichero
// en el disco de un servicio aparte, lo que obligaba a montar un volumen y a que
// ese fichero existiera en cada despliegue: si faltaba, el fallo aparecía en la
// primera constancia de un alumno real, no al arrancar. Embebida, o compila o no
// compila.
//
// El formato del documento NO es negociable: lo fija la STPS. En particular la
// CURP y el RFC se pintan carácter por carácter, una casilla por letra, y por
// eso la plantilla tiene 59 marcadores en lugar de un puñado.
package dc3

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lukasjarosch/go-docx"
)

//go:embed plantilla/dc3.docx
var plantilla []byte

// Las dos imágenes de la cabecera, en el orden en que aparecen.
//
//	image1.jpg   5198x4536 px, 1.8 MB → logotipo IZQUIERDO (patrón)
//	image2.jpeg   204x162 px,  14 KB  → logotipo DERECHO (agente capacitador)
//
// Ambas se sustituyen. Un comentario anterior aquí describía la segunda como
// "el sello del formato" que no debía tocarse nunca; era falso. Se extrajo del
// .docx y es el mismo logotipo comercial que la primera, solo que más pequeño.
// Mientras se dio por buena esa descripción, toda constancia salía con el
// logotipo de la plantilla en el lado derecho, fuese de quien fuese quien la
// emitía.
//
// El desequilibrio de tamaños no es un error de la plantilla original: la
// izquierda se guardó a resolución de cámara y se dibuja a 0.93 pulgadas. Por
// eso el .docx pesa casi 2 MB y LibreOffice tarda en convertirlo.
const (
	logoIzquierdo = "word/media/image1.jpg"
	logoDerecho   = "word/media/image2.jpeg"
)

// Datos es todo lo que necesita una constancia.
//
// Se agrupa por origen, que es lo que determina quién lo captura:
//   - Trabajador: lo llena el alumno al terminar el curso.
//   - Empresa y Curso: lo llena el instructor una vez, en la capacitación.
type Datos struct {
	// ── Trabajador ───────────────────────────────────────────────────────────
	NombreTrabajador    string
	CURP                string
	OcupacionEspecifica string
	Puesto              string

	// ── Empresa ──────────────────────────────────────────────────────────────
	RazonSocial string
	RFC         string
	// NombrePatron y NombreRepresentante son los agentes que firman: el patrón
	// o su representante legal, y el representante de los trabajadores.
	NombrePatron        string
	NombreRepresentante string

	// Logos de la cabecera, en bytes de imagen ya decodificados.
	//
	// La constancia lleva DOS: a la izquierda el del patrón —la empresa que
	// emplea al trabajador— y a la derecha el de quien imparte la capacitación.
	// Cuando el alumno no declara empresa propia coinciden, porque el patrón que
	// figura en el documento es el del instructor.
	//
	// Vacíos dejan el de la plantilla. Antes solo se sustituía el izquierdo y el
	// derecho quedaba siempre con el logotipo de la plantilla, que es de otra
	// empresa: una constancia emitida por un instructor con logo propio salía
	// con dos marcas distintas y una de ellas ajena.
	LogoPatron      []byte
	LogoCapacitador []byte

	// ── Curso ────────────────────────────────────────────────────────────────
	NombreCurso       string
	DuracionHoras     string
	AreaTematica      string
	NombreCapacitador string
	// Fechas en formato YYYY-MM-DD. Se descomponen en casillas.
	FechaInicio string
	FechaFin    string

	// ── Verificación ─────────────────────────────────────────────────────────
	//
	// Folio es el pie que se imprime al final del documento, con el código único
	// y la URL donde contrastarlo.
	//
	// No es un campo del formato oficial de la STPS: es un añadido nuestro, por
	// eso va en un párrafo suelto al final y no en una casilla. Existe porque el
	// PDF por sí solo no impide falsificar —se edita y el diseño se copia—; lo
	// que hace detectable una constancia inventada es que su folio no exista en
	// la base. Vacío deja el documento sin pie, sin romper nada.
	Folio string
}

// ErrDatosIncompletos indica que faltan campos obligatorios de la constancia.
var ErrDatosIncompletos = errors.New("faltan datos obligatorios para la DC-3")

// AreasTematicas es el catálogo oficial de áreas temáticas de los cursos.
//
// Transcrito del REVERSO de la plantilla embebida: "CLAVES Y DENOMINACIONES DEL
// CATÁLOGO DE ÁREAS TEMÁTICAS DE LOS CURSOS". Son nueve y solo cambian si la
// STPS publica un formato nuevo, en cuyo caso también cambiaría la plantilla.
//
// Debe coincidir con AREAS_TEMATICAS_DC3 de frontend/src/utils/dc3.ts.
var AreasTematicas = map[string]string{
	"1000": "Producción general",
	"2000": "Servicios",
	"3000": "Administración, contabilidad y economía",
	"4000": "Comercialización",
	"5000": "Mantenimiento y reparación",
	"6000": "Seguridad",
	"7000": "Desarrollo personal y familiar",
	"8000": "Uso de tecnologías de la información y comunicación",
	"9000": "Participación social",
}

// AreaValida comprueba que la clave pertenezca al catálogo.
//
// Se valida en el servidor además de ofrecer un select en la interfaz: la clave
// va impresa en un documento legal y una petición directa podría colar
// cualquier cadena, produciendo una constancia inválida que nadie detecta hasta
// que alguien la presenta en una inspección.
func AreaValida(clave string) bool {
	_, ok := AreasTematicas[strings.TrimSpace(clave)]
	return ok
}

// Faltantes devuelve los campos obligatorios que están vacíos.
//
// Se expone en lugar de limitarse a fallar porque quien llama necesita decirle
// al usuario QUÉ le falta: la generación es automática al terminar el curso, y
// un "no se pudo generar" a secas dejaría al alumno sin saber qué hacer.
func (d Datos) Faltantes() []string {
	obligatorios := []struct {
		nombre string
		valor  string
	}{
		{"nombre del trabajador", d.NombreTrabajador},
		{"CURP", d.CURP},
		{"ocupación específica", d.OcupacionEspecifica},
		{"puesto", d.Puesto},
		{"razón social", d.RazonSocial},
		{"RFC", d.RFC},
		{"nombre del patrón o representante legal", d.NombrePatron},
		{"representante de los trabajadores", d.NombreRepresentante},
		{"nombre del curso", d.NombreCurso},
		{"duración en horas", d.DuracionHoras},
		{"área temática", d.AreaTematica},
		{"nombre del capacitador", d.NombreCapacitador},
		{"fecha de inicio", d.FechaInicio},
		{"fecha de fin", d.FechaFin},
	}

	var faltan []string
	for _, c := range obligatorios {
		if strings.TrimSpace(c.valor) == "" {
			faltan = append(faltan, c.nombre)
		}
	}
	return faltan
}

// Generar devuelve el .docx de la constancia listo para guardar o servir.
//
// Trabaja sobre un fichero temporal porque go-docx abre y escribe por ruta, no
// por io.Reader. El temporal se borra siempre, incluso si algo falla a medias.
func Generar(d Datos) ([]byte, error) {
	if faltan := d.Faltantes(); len(faltan) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrDatosIncompletos, strings.Join(faltan, ", "))
	}

	origen, err := os.CreateTemp("", "dc3-plantilla-*.docx")
	if err != nil {
		return nil, fmt.Errorf("temporal de plantilla: %w", err)
	}
	defer os.Remove(origen.Name())

	if _, err := origen.Write(plantilla); err != nil {
		origen.Close()
		return nil, fmt.Errorf("volcando plantilla: %w", err)
	}
	if err := origen.Close(); err != nil {
		return nil, fmt.Errorf("cerrando plantilla: %w", err)
	}

	doc, err := docx.Open(origen.Name())
	if err != nil {
		return nil, fmt.Errorf("abriendo plantilla: %w", err)
	}

	if err := doc.ReplaceAll(marcadores(d)); err != nil {
		return nil, fmt.Errorf("reemplazando marcadores: %w", err)
	}

	if len(d.LogoPatron) > 0 {
		doc.SetFile(logoIzquierdo, d.LogoPatron)
	}
	if len(d.LogoCapacitador) > 0 {
		doc.SetFile(logoDerecho, d.LogoCapacitador)
	}

	destino, err := os.CreateTemp("", "dc3-salida-*.docx")
	if err != nil {
		return nil, fmt.Errorf("temporal de salida: %w", err)
	}
	destino.Close()
	defer os.Remove(destino.Name())

	if err := doc.WriteToFile(destino.Name()); err != nil {
		return nil, fmt.Errorf("escribiendo constancia: %w", err)
	}
	return os.ReadFile(destino.Name())
}

// NombreArchivo propone un nombre estable y legible para la constancia.
//
// La extensión es .pdf: lo que se entrega al alumno ya no es el .docx. Servir
// el Word significaba entregarle la plantilla oficial con los campos ya
// sustituidos, es decir, un documento en el que cambiar el nombre y la empresa
// es cuestión de dos clics y no requiere haber comprado ningún curso.
func NombreArchivo(nombreTrabajador, nombreCurso string) string {
	limpia := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.Map(func(r rune) rune {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
				return r
			case r == ' ':
				return '-'
			default:
				return -1
			}
		}, s)
		if len(s) > 40 {
			s = s[:40]
		}
		return strings.Trim(s, "-")
	}
	return fmt.Sprintf("DC3-%s-%s-%d.pdf",
		limpia(nombreTrabajador), limpia(nombreCurso), time.Now().Unix())
}

// ── Interno ──────────────────────────────────────────────────────────────────

func marcadores(d Datos) docx.PlaceholderMap {
	m := docx.PlaceholderMap{
		"NombreTrabajador":    mayus(d.NombreTrabajador),
		"OcupacionEspecifica": mayus(codigoOcupacion(d.OcupacionEspecifica)),
		"Puesto":              mayus(d.Puesto),
		"NombrePatron":        mayus(d.NombrePatron),
		"NombreRepresentante": mayus(d.NombreRepresentante),
		"RazonSocial":         mayus(d.RazonSocial),
		"NombreCurso":         mayus(d.NombreCurso),
		"DuracionHoras":       conSufijoHoras(d.DuracionHoras),
		"AreaTematica":        mayus(d.AreaTematica),
		"NombreCapacitador":   mayus(d.NombreCapacitador),
		// Sin mayus: el folio lleva una URL y forzarla a mayúsculas la haría
		// más difícil de teclear a mano desde el papel.
		"Folio": strings.TrimSpace(d.Folio),
	}

	// La plantilla NO tiene marcador para el giro principal: el formato oficial
	// no lo pide en una casilla propia. Si algún día se añade a la plantilla,
	// aquí es donde entra.

	repartir(m, "C", d.CURP, 18)
	repartir(m, "R", rfcConGuiones(d.RFC), 15)

	repartirFecha(m, d.FechaInicio, "IY", "IM", "ID")
	repartirFecha(m, d.FechaFin, "EY", "EM", "ED")

	return m
}

// repartir escribe un valor carácter por carácter en marcadores numerados
// (C1..C18, R1..R15). Los sobrantes se rellenan vacíos: sin esto, un CURP corto
// dejaría marcadores sin sustituir y saldrían impresos como "{C17}".
func repartir(m docx.PlaceholderMap, prefijo, valor string, casillas int) {
	runas := []rune(valor)
	for i := 0; i < casillas; i++ {
		c := ""
		if i < len(runas) {
			c = string(runas[i])
		}
		m[fmt.Sprintf("%s%d", prefijo, i+1)] = c
	}
}

// repartirFecha descompone una fecha YYYY-MM-DD en sus casillas.
func repartirFecha(m docx.PlaceholderMap, fecha, pfxAnio, pfxMes, pfxDia string) {
	if len(fecha) < 10 {
		// Se reparten vacíos igualmente para no dejar marcadores a la vista.
		repartir(m, pfxAnio, "", 4)
		repartir(m, pfxMes, "", 2)
		repartir(m, pfxDia, "", 2)
		return
	}
	repartir(m, pfxAnio, fecha[0:4], 4)
	repartir(m, pfxMes, fecha[5:7], 2)
	repartir(m, pfxDia, fecha[8:10], 2)
}

// codigoOcupacion extrae la clave del Catálogo Nacional de Ocupaciones.
//
// El campo suele llegar como "04.6 Supervisores..." y en la constancia va solo
// la clave. Las de la forma "4.6" se normalizan a "04.6", que es como las
// espera el formato.
func codigoOcupacion(texto string) string {
	partes := strings.Fields(texto)
	if len(partes) == 0 {
		return texto
	}
	codigo := partes[0]
	if len(codigo) == 3 && codigo[1] == '.' {
		return "0" + codigo
	}
	return codigo
}

// rfcConGuiones da al RFC el formato de la constancia. Distingue persona moral
// (12 caracteres) de física (13).
func rfcConGuiones(rfc string) string {
	rfc = strings.NewReplacer("-", "", " ", "").Replace(rfc)
	switch len(rfc) {
	case 12:
		// El espacio inicial alinea la clave de 12 en la rejilla de 15 casillas.
		return " " + rfc[:3] + "-" + rfc[3:9] + "-" + rfc[9:]
	case 13:
		return rfc[:4] + "-" + rfc[4:10] + "-" + rfc[10:]
	default:
		return rfc
	}
}

// conSufijoHoras normaliza la duración a "N HRS", que es como se lee el campo.
func conSufijoHoras(d string) string {
	d = strings.ToUpper(strings.TrimSpace(d))
	if d == "" || strings.HasSuffix(d, "HRS") || strings.HasSuffix(d, "HORAS") {
		return d
	}
	return d + " HRS"
}

func mayus(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }
