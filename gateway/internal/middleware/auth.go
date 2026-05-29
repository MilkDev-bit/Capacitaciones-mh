// Package middleware implementa el middleware de autenticación del Gateway.
// El Gateway valida el JWT localmente (firma + expiración) y luego llama al
// auth service para verificar token_version (revocación).
package middleware

import (
	"net/http"
	"strings"

	"Prueba-Go/gateway/internal/clients"
	authpb "Prueba-Go/gen/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CtxUserID       = "userID"
	CtxUserEmail    = "userEmail"
	CtxUserRole     = "userRole"
	CtxTokenVersion = "tokenVersion"
)

// AuthRequired extrae el JWT de la cookie `auth_token`, lo valida llamando al
// auth service y almacena los claims en el contexto de Gin.
func AuthRequired(c *clients.Clients) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := extractToken(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
			return
		}

		claims, err := c.Auth.ValidateToken(ctx.Request.Context(), &authpb.ValidateTokenRequest{Token: token})
		if err != nil {
			code := grpcErrToHTTP(err)
			ctx.AbortWithStatusJSON(code, gin.H{"error": "sesión inválida o expirada"})
			return
		}

		ctx.Set(CtxUserID, claims.UserId)
		ctx.Set(CtxUserEmail, claims.Email)
		ctx.Set(CtxUserRole, claims.Role)
		ctx.Set(CtxTokenVersion, claims.TokenVersion)
		ctx.Next()
	}
}

// InstructorRequired aplica AuthRequired y además verifica rol instructor/admin.
func InstructorRequired(c *clients.Clients) gin.HandlerFunc {
	auth := AuthRequired(c)
	return func(ctx *gin.Context) {
		auth(ctx)
		if ctx.IsAborted() {
			return
		}
		role := ctx.GetString(CtxUserRole)
		if role != "instructor" && role != "admin" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "se requiere rol instructor"})
		}
	}
}

// AdminRequired aplica AuthRequired y verifica rol admin.
func AdminRequired(c *clients.Clients) gin.HandlerFunc {
	auth := AuthRequired(c)
	return func(ctx *gin.Context) {
		auth(ctx)
		if ctx.IsAborted() {
			return
		}
		if ctx.GetString(CtxUserRole) != "admin" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "se requiere rol admin"})
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func extractToken(ctx *gin.Context) (string, error) {
	// 1. Cookie HttpOnly (fuente principal)
	if cookie, err := ctx.Cookie("auth_token"); err == nil && cookie != "" {
		return cookie, nil
	}
	// 2. Header Authorization: Bearer <token> (fallback para herramientas)
	if h := ctx.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer "), nil
	}
	return "", http.ErrNoCookie
}

func grpcErrToHTTP(err error) int {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
