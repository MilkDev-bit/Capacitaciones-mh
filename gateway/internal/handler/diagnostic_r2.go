package handler

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/storage"

	"github.com/gin-gonic/gin"
)

// DiagnosticoR2 comprueba de extremo a extremo que el gateway puede escribir en
// el bucket.
//
// Existe por la misma razón que el de Gotenberg: diagnosticarlo desde el
// navegador es imposible. R2 responde a un PUT rechazado con un XML que explica
// exactamente qué pasa —firma incorrecta, llave inexistente, permisos
// insuficientes—, pero esa respuesta NO lleva cabeceras CORS, así que el
// navegador la descarta y solo muestra "blocked by CORS policy". El motivo real
// nunca llega a la consola.
//
// Esta prueba hace la misma escritura desde el servidor, donde CORS no existe,
// y devuelve el error de R2 íntegro.
//
// Solo admin: revela el nombre del bucket y el endpoint.
type R2Handler struct{}

func NewR2Handler() *R2Handler { return &R2Handler{} }

func (h *R2Handler) DiagnosticoR2(ctx *gin.Context) {
	svc := storage.Default()
	if svc == nil || !svc.Configurado() {
		ctx.JSON(http.StatusOK, gin.H{
			"configurado": false,
			"detalle":     "Faltan R2_BUCKET, R2_ENDPOINT, R2_ACCESS_KEY o R2_SECRET_KEY en el gateway.",
			"siguiente":   "Revisa las variables del servicio en Railway y vuelve a desplegar.",
		})
		return
	}

	c, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	// Se escribe de verdad y se borra. Una comprobación de solo lectura no
	// serviría: un token "Object Read only" la pasaría y seguiría rompiendo
	// todas las subidas, que es justo el caso que hay que distinguir.
	clave := "diagnostico/escritura-de-prueba.txt"
	cuerpo := []byte("prueba de escritura del gateway")

	inicio := time.Now()
	_, err := svc.UploadFile(c, clave, "text/plain", bytes.NewReader(cuerpo), int64(len(cuerpo)))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"configurado":    true,
			"bucket":         svc.Bucket(),
			"puede_escribir": false,
			"detalle":        err.Error(),
			// El error de R2 trae un <Code> que identifica la causa exacta. Se
			// traduce porque los cuatro se arreglan en sitios distintos y
			// confundirlos cuesta horas.
			"pistas": gin.H{
				"InvalidAccessKeyId":    "La Access Key no existe o se borró al rotar. Copia de nuevo las dos mitades del token.",
				"SignatureDoesNotMatch": "La Secret Key no corresponde a esa Access Key, o quedó con un espacio o salto de línea al pegarla en Railway.",
				"AccessDenied":          "El token existe y firma bien, pero no tiene permiso de escritura. En Cloudflare el token debe ser 'Object Read & Write', no 'Object Read only', y su alcance incluir este bucket.",
				"NoSuchBucket":          "El bucket del endpoint no coincide con R2_BUCKET. R2_ENDPOINT lleva SOLO la cuenta: https://<cuenta>.r2.cloudflarestorage.com",
			},
		})
		return
	}
	escritura := time.Since(inicio)

	// Se limpia. Si el borrado falla no se considera un fallo del diagnóstico:
	// escribir era lo que se quería comprobar, y un objeto de 31 bytes suelto
	// no molesta a nadie.
	errBorrado := svc.DeleteFile(c, clave)

	resp := gin.H{
		"configurado":    true,
		"bucket":         svc.Bucket(),
		"puede_escribir": true,
		"ms_escritura":   escritura.Milliseconds(),
		"detalle":        "El gateway escribió y leyó en el bucket sin problemas.",
	}
	if errBorrado != nil {
		resp["aviso"] = "Se pudo escribir pero no borrar el objeto de prueba (" +
			strings.TrimSpace(errBorrado.Error()) + "). El token quizá no tiene permiso de borrado."
	}
	ctx.JSON(http.StatusOK, resp)
}
