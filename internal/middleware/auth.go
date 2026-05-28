package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"Prueba-Go/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	gocache "github.com/patrickmn/go-cache"
)

var jwtSecret []byte

// tokenVersionCache almacena en memoria la token_version de cada usuario.
// TTL 5 min, purga cada 10 min — evita un SELECT por cada petición autenticada.
var tokenVersionCache = gocache.New(5*time.Minute, 10*time.Minute)

func SetSecret(s []byte) {
	jwtSecret = s
}

// InvalidateTokenVersionCache elimina la entrada cacheada de un usuario
// (debe llamarse al cambiar contraseña o banear al usuario).
func InvalidateTokenVersionCache(userID string) {
	tokenVersionCache.Delete(userID)
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Leer token desde cookie HttpOnly; fallback a Bearer para clientes API.
		tokenStr, err := c.Cookie("auth_token")
		if err != nil || tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
			return
		}

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

		userID, _ := claims["sub"].(string)
		claimVer, _ := claims["ver"].(float64)

		// Verificar token_version con caché para evitar un SELECT por petición.
		var dbVer int
		cacheKey := fmt.Sprintf("tv:%s", userID)
		if cached, found := tokenVersionCache.Get(cacheKey); found {
			dbVer = cached.(int)
		} else {
			if err := db.DB.QueryRow(`SELECT token_version FROM users WHERE id=$1`, userID).Scan(&dbVer); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
				return
			}
			tokenVersionCache.Set(cacheKey, dbVer, gocache.DefaultExpiration)
		}

		if int(claimVer) != dbVer {
			tokenVersionCache.Delete(cacheKey) // forzar refresco en próximo intento
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
	return os.Getenv("JWT_SECRET")
}
