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

	uploadURL, finalURL, err := svc.GeneratePresignedURL(c.Request.Context(), prefix, ext, contentType, 15*time.Minute)
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
