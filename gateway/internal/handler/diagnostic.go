package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func (h *CursosHandler) DiagnosticLastError(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"error": "Check the logs, diagnostic endpoint active!"})
}
