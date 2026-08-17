package handler

import (
	"context"
	"net/http"
	"time"

	"Prueba-Go/gateway/internal/pdf"

	"github.com/gin-gonic/gin"
)

// DiagnosticoGotenberg comprueba la conexión con el conversor a PDF.
//
// Existe porque diagnosticar esto por los registros salió caro: el mensaje de
// error viaja como atributo de una línea de log, el panel de Railway lo trunca,
// y cada intento de averiguar la causa costaba un despliegue completo. Aquí la
// respuesta llega al navegador entera y al instante.
//
// Solo admin: revela la topología interna —el nombre del servicio, su puerto y
// si responde—, que es justo lo que buscaría alguien tanteando la red.
//
// No convierte ningún documento: pregunta a /health, que es lo que distingue
// "no llego al servicio" de "llego pero la conversión falla".
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

	// Timeout corto: aquí interesa saber si responde, no esperar a que arranque
	// LibreOffice. Un diagnóstico que tarda un minuto no se usa.
	c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	inicio := time.Now()
	req, err := http.NewRequestWithContext(c, http.MethodGet, base+"/health", nil)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"configurado": true,
			"url":         base,
			"alcanzable":  false,
			"detalle":     "La URL configurada no es válida: " + err.Error(),
		})
		return
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		// El texto del error de red es el dato que hacía falta y que el log
		// recortaba. Se devuelve tal cual, sin interpretarlo.
		ctx.JSON(http.StatusOK, gin.H{
			"configurado": true,
			"url":         base,
			"alcanzable":  false,
			"detalle":     err.Error(),
			"pistas": gin.H{
				"no such host":       "el nombre del servicio no coincide con el hostname interno de Railway",
				"connection refused": "Gotenberg escucha solo en IPv4; falta API_BIND_IP=::",
				"i/o timeout":        "puerto equivocado o el servicio no está arrancado",
			},
		})
		return
	}
	defer resp.Body.Close()

	ctx.JSON(http.StatusOK, gin.H{
		"configurado": true,
		"url":         base,
		"alcanzable":  true,
		"estado_http": resp.StatusCode,
		"tardo_ms":    time.Since(inicio).Milliseconds(),
	})
}
