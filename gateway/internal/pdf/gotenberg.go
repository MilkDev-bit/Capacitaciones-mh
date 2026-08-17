// Package pdf convierte documentos ofimáticos a PDF a través de Gotenberg.
//
// Gotenberg es un servicio aparte que empaqueta LibreOffice detrás de una API
// HTTP. Se eligió frente a instalar LibreOffice en esta imagen porque el
// gateway corre sobre distroless/static (unos 2 MB) y meter LibreOffice dentro
// lo llevaría a cerca de 1 GB, con el arranque en frío y los despliegues que
// eso implica.
//
// El precio de la decisión es una dependencia de red: si Gotenberg no responde,
// no se emiten constancias. Es deliberado —ver Convertir— porque la alternativa
// sería entregar el .docx, que es justo lo que se quiere evitar.
package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// ErrNoConfigurado indica que falta GOTENBERG_URL.
var ErrNoConfigurado = errors.New("conversión a PDF no configurada")

// tamMaxPDF limita lo que se acepta de vuelta del conversor.
//
// Sin tope, un Gotenberg comprometido o mal configurado podría devolver un
// cuerpo enorme y agotar la memoria del gateway.
const tamMaxPDF = 32 << 20

// clienteHTTP se reutiliza entre llamadas: crear un http.Client por conversión
// abre un pool de conexiones nuevo cada vez.
//
// El timeout es generoso porque la primera conversión tras arrancar Gotenberg
// carga el perfil de LibreOffice y puede pasar de diez segundos; las siguientes
// bajan a uno o dos.
var clienteHTTP = &http.Client{Timeout: 90 * time.Second}

// URL devuelve la base de Gotenberg configurada, sin barra final.
func URL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("GOTENBERG_URL")), "/")
}

// Disponible informa de si la conversión está configurada.
//
// Se consulta al arrancar para avisar en los logs, en lugar de descubrirlo la
// primera vez que un alumno termine un curso.
func Disponible() bool { return URL() != "" }

// Convertir transforma un documento ofimático en PDF.
//
// NO tiene camino de respaldo a propósito. Devolver el .docx original cuando la
// conversión falla dejaría al alumno con la plantilla editable en la mano —el
// escenario que motivó el cambio— y encima de forma intermitente, que es la
// manera más difícil de detectar un agujero. Si esto falla, la constancia no se
// emite y el error queda registrado.
func Convertir(ctx context.Context, nombre string, doc []byte) ([]byte, error) {
	base := URL()
	if base == "" {
		return nil, ErrNoConfigurado
	}
	if len(doc) == 0 {
		return nil, errors.New("documento vacío")
	}

	cuerpo := &bytes.Buffer{}
	w := multipart.NewWriter(cuerpo)
	// El nombre importa: Gotenberg decide el conversor por la extensión, así que
	// un fichero sin ".docx" se rechaza.
	parte, err := w.CreateFormFile("files", nombre)
	if err != nil {
		return nil, fmt.Errorf("armando la petición: %w", err)
	}
	if _, err := parte.Write(doc); err != nil {
		return nil, fmt.Errorf("escribiendo el documento: %w", err)
	}
	// No se envía `pdfa`: ese campo selecciona un perfil PDF/A concreto
	// (PDF/A-1b, PDF/A-3b…) y mandarlo vacío es un valor inválido, no un
	// "por defecto". La conversión normal ya conserva la maquetación.
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("cerrando la petición: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		base+"/forms/libreoffice/convert", cuerpo)
	if err != nil {
		return nil, fmt.Errorf("creando la petición: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := clienteHTTP.Do(req)
	if err != nil {
		// El error de red se envuelve con la URL: "no such host" y "connection
		// refused" apuntan a problemas distintos —nombre del servicio frente a
		// puerto o binding— y saber cuál es ahorra un ciclo de despliegue.
		return nil, fmt.Errorf("llamando a Gotenberg en %s: %w", base, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Se recorta el cuerpo: Gotenberg devuelve el error en texto plano, pero
		// ante una respuesta inesperada no conviene volcarla entera al log.
		detalle, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("Gotenberg respondió %d: %s",
			resp.StatusCode, strings.TrimSpace(string(detalle)))
	}

	// LimitReader y no ContentLength: la cabecera la controla el otro extremo.
	out, err := io.ReadAll(io.LimitReader(resp.Body, tamMaxPDF+1))
	if err != nil {
		return nil, fmt.Errorf("leyendo el PDF: %w", err)
	}
	if len(out) > tamMaxPDF {
		return nil, errors.New("el PDF devuelto supera el tamaño máximo")
	}
	// Comprobación barata contra una respuesta 200 que no sea un PDF: sin ella,
	// se subiría a R2 una página de error con extensión .pdf.
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		return nil, errors.New("la respuesta no es un PDF")
	}
	return out, nil
}
