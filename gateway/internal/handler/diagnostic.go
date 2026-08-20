package handler

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"Prueba-Go/gateway/internal/pdf"
	"Prueba-Go/pkg/dc3"

	"github.com/gin-gonic/gin"
)

// DiagnosticoGotenberg comprueba de extremo a extremo la generación del PDF.
//
// Existe porque diagnosticar esto de otra forma salió caro: el error de red
// viajaba truncado en el panel de logs, y la fuente que acaba usando LibreOffice
// solo se ve abriendo el PDF ya emitido. Cada hipótesis costaba un despliegue.
// Aquí la respuesta llega al navegador, entera y al instante.
//
// Solo admin: revela la topología interna —nombre del servicio, puerto, si
// responde—, que es justo lo que buscaría alguien tanteando la red.
func (h *DC3Handler) DiagnosticoGotenberg(ctx *gin.Context) {
	base := pdf.URL()
	if base == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"configurado": false,
			"detalle":     "GOTENBERG_URL está vacía; el gateway no intentará convertir nada.",
			"siguiente":   "Añade GOTENBERG_URL al servicio del gateway, por ejemplo http://gotenberg.railway.internal:3000",
		})
		return
	}

	// Timeout holgado: la primera conversión tras arrancar Gotenberg carga el
	// perfil de LibreOffice y puede pasar de diez segundos.
	c, cancel := context.WithTimeout(ctx.Request.Context(), 90*time.Second)
	defer cancel()

	inicio := time.Now()
	// Se convierte la plantilla REAL, con datos de relleno. Una prueba con un
	// documento cualquiera no serviría: lo que se quiere saber es qué fuente
	// resuelve para Arial Narrow, y eso depende de esta plantilla concreta.
	doc, err := dc3.Generar(datosDePrueba())
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"configurado": true, "url": base,
			"alcanzable": false,
			"detalle":    "no se pudo armar el documento de prueba: " + err.Error(),
		})
		return
	}

	salida, err := pdf.Convertir(c, "diagnostico.docx", doc)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"configurado": true, "url": base,
			"alcanzable": false,
			"detalle":    err.Error(),
			"pistas": gin.H{
				"no such host":       "el nombre del servicio no coincide con el hostname interno de Railway",
				"connection refused": "Gotenberg escucha en otra interfaz o el servicio está caído",
				"i/o timeout":        "puerto equivocado o el servicio no está arrancado",
			},
		})
		return
	}

	// LibreOffice escribe los nombres de fuente en las entradas /BaseFont, que
	// van sin comprimir, así que basta buscarlos en los bytes.
	tiene := func(s string) bool { return bytes.Contains(salida, []byte(s)) }
	narrow := tiene("Narrow")

	resp := gin.H{
		"configurado": true,
		"url":         base,
		"alcanzable":  true,
		"tardo_ms":    time.Since(inicio).Milliseconds(),
		"pdf_bytes":   len(salida),
		"fuentes": gin.H{
			"arial_narrow_sustituida": narrow,
			"cayo_a_noto_sans":        tiene("NotoSans"),
			"liberation_presente":     tiene("Liberation"),
		},
	}
	if narrow {
		resp["veredicto"] = "OK: la plantilla se convierte con una fuente estrecha; la maquetación debería respetarse."
	} else {
		// Este es el estado en el que se emitieron las primeras constancias:
		// convertía sin error, pero con una fuente de ancho normal donde la
		// plantilla espera una condensada, así que los campos salían cortados.
		resp["veredicto"] = "FALLO: Arial Narrow no resuelve a una fuente estrecha. " +
			"El servicio de Gotenberg no está construido desde gotenberg/Dockerfile. " +
			"Revisa que su Config-as-code apunte a gotenberg/railway.toml."
	}
	ctx.JSON(http.StatusOK, resp)
}

// datosDePrueba rellena los campos obligatorios con valores evidentes.
//
// No se usan datos de nadie: este endpoint no debe poder servir para extraer la
// constancia de un alumno real.
func datosDePrueba() dc3.Datos {
	return dc3.Datos{
		NombreTrabajador:    "DIAGNOSTICO DE FUENTES",
		CURP:                "XXXX000000XXXXXX00",
		OcupacionEspecifica: "00.0 Prueba",
		Puesto:              "Prueba",
		RazonSocial:         "Prueba",
		RFC:                 "XXX000000XX0",
		NombrePatron:        "Prueba",
		NombreRepresentante: "Prueba",
		NombreCurso:         "PRUEBA DE CONVERSION",
		DuracionHoras:       "1",
		AreaTematica:        "6000",
		NombreCapacitador:   "Prueba",
		FechaInicio:         "2026-01-01",
		FechaFin:            "2026-01-01",
	}
}
