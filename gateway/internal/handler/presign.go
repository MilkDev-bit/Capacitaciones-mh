package handler

import (
	"net/http"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/storage"

	"github.com/gin-gonic/gin"
)

var allowedPrefixes = map[string]bool{
	"videos":     true,
	"documents":  true,
	"thumbnails": true,
	"avatars":    true,
	"covers":     true,
	"foro":       true,
	"mensajes":   true,
	"entregas":   true,
	// Logotipo de empresa de la cabecera de la constancia DC-3. El servidor lo
	// vuelve a descargar al emitir, y solo acepta URLs que cuelguen del bucket
	// propio, así que el prefijo tiene que estar permitido aquí o la subida
	// muere con un 400 antes de empezar.
	"dc3-logos": true,
}

// vigencia devuelve cuánto dura la URL prefirmada según lo que se vaya a subir.
//
// Los 15 minutos que había para todo son suficientes para un avatar y
// claramente insuficientes para un vídeo: una grabación de 250 MB necesita
// sostener 2,3 Mbps durante los 900 segundos completos para entrar en esa
// ventana, y eso deja fuera a cualquier conexión modesta. Al caducar a mitad,
// R2 responde 403 y el navegador lo enseña como un error de CORS — así que el
// instructor ve "bloqueado por CORS" cuando lo que pasó fue que se le acabó el
// tiempo.
//
// El plazo es el tiempo que alguien puede EMPEZAR a usar la URL y seguir
// subiendo; alargarlo no deja el bucket más expuesto de forma apreciable,
// porque la URL ya solo permite escribir en una clave concreta que elige el
// servidor.
func vigencia(prefix string) time.Duration {
	switch prefix {
	case "videos":
		// Una clase grabada puede pasar del medio giga.
		return 2 * time.Hour
	case "documents", "entregas":
		// Manuales y entregas: grandes, pero rara vez como un vídeo.
		return 45 * time.Minute
	default:
		// Avatares, portadas, logos y adjuntos de mensaje: pesan poco y el plazo
		// corto limita la ventana en que una URL filtrada sirve de algo.
		return 15 * time.Minute
	}
}

// PresignHandler genera URLs pre-firmadas para subidas directas al bucket R2.
type PresignHandler struct{}

func NewPresignHandler() *PresignHandler { return &PresignHandler{} }

// PresignUpload godoc
// GET /api/presign?prefix=videos&ext=mp4
func (h *PresignHandler) PresignUpload(c *gin.Context) {
	prefix := c.Query("prefix")
	ext := strings.ToLower(c.Query("ext"))
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}

	if !allowedPrefixes[prefix] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prefijo no permitido"})
		return
	}

	if prefix == "entregas" && storage.BlockedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se permiten archivos ejecutables o scripts"})
		return
	}

	contentType, ok := storage.MimeTypes[ext]
	if !ok {
		if prefix == "entregas" && !storage.BlockedExtensions[ext] {
			contentType = "application/octet-stream"
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "extensión no permitida"})
			return
		}
	}

	svc := storage.Default()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "almacenamiento no configurado"})
		return
	}

	uploadURL, finalURL, err := svc.GeneratePresignedURL(c.Request.Context(), prefix, ext, contentType, vigencia(prefix))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo generar la URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url":   uploadURL,
		"final_url":    finalURL,
		"content_type": contentType,
	})
}
