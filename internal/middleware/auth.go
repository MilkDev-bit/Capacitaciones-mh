package middleware

import (
	"net/http"
	"os"
	"strings"

	"Prueba-Go/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func SetSecret(s []byte) {
	jwtSecret = s
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}
		claims, _ := token.Claims.(jwt.MapClaims)

		// Verificar token_version: si el usuario cambió su contraseña o fue baneado,
		// la versión en la BD será mayor y el token queda revocado.
		userID, _ := claims["sub"].(string)
		claimVer, _ := claims["ver"].(float64)
		var dbVer int
		if err := db.DB.QueryRow(`SELECT token_version FROM users WHERE id=$1`, userID).Scan(&dbVer); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}
		if int(claimVer) != dbVer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesión expirada, inicia sesión de nuevo"})
			return
		}

		c.Set("user_id", userID)
		c.Set("role", claims["role"])
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
			return
		}
		c.Next()
	}
}

func InstructorRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "instructor" && role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acceso denegado: se requiere rol instructor"})
			return
		}
		c.Next()
	}
}

func getSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "changeme_secret_key_32chars_long!!"
}
