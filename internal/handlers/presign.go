package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"Prueba-Go/internal/storage"

	"github.com/gin-gonic/gin"
)

// allowedPresignPrefixes whitelists upload destinations to prevent path traversal.
var allowedPresignPrefixes = map[string]bool{
	"videos":     true,
	"documents":  true,
	"thumbnails": true,
	"avatars":    true,
	"covers":     true,
	"foro":       true,
}

// allowedPresignExts whitelists file extensions for presigned uploads.
var allowedPresignExts = map[string]bool{
	".mp4":  true,
	".webm": true,
	".mov":  true,
	".avi":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".pptx": true,
	".xlsx": true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// PresignUpload generates a presigned PUT URL so the client can upload directly to R2.
// Query params: prefix (e.g. "videos") and ext (e.g. "mp4" or ".mp4").
// The URL expires in 15 minutes.
func PresignUpload(c *gin.Context) {
	prefix := c.Query("prefix")
	ext := strings.ToLower(c.Query("ext"))
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	if !allowedPresignPrefixes[prefix] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prefijo no permitido"})
		return
	}
	if !allowedPresignExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "extensión no permitida"})
		return
	}

	uploadURL, finalURL, err := storage.GeneratePresignedURL(c.Request.Context(), prefix, ext, 15*time.Minute)
	if err != nil {
		slog.Error("PresignUpload", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": uploadURL,
		"final_url":  finalURL,
	})
}
