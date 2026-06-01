package handler

import (
	"net/http"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/config"
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
}

// PresignHandler genera URLs pre-firmadas para subidas directas al bucket R2.
type PresignHandler struct {
	svc *storage.StorageService
}

func NewPresignHandler(cfg *config.Config) *PresignHandler {
	svc := storage.New(cfg.R2Bucket, cfg.R2Endpoint, cfg.R2AccessKey, cfg.R2SecretKey, cfg.R2PublicURL)
	return &PresignHandler{svc: svc}
}

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

	contentType, ok := storage.MimeTypes[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "extensión no permitida"})
		return
	}

	uploadURL, finalURL, err := h.svc.GeneratePresignedURL(c.Request.Context(), prefix, ext, contentType, 15*time.Minute)
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
