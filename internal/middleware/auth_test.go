package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func buildToken(secret []byte, claims jwt.MapClaims) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := t.SignedString(secret)
	return signed
}

// TestAuthRequired_SinToken verifica que la petición sin header devuelve 401.
func TestAuthRequired_SinToken(t *testing.T) {
	SetSecret([]byte("test-secret"))
	router := gin.New()
	router.GET("/protected", AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, got %d", w.Code)
	}
}

// TestAuthRequired_TokenMalformado verifica que un token con firma inválida devuelve 401.
func TestAuthRequired_TokenMalformado(t *testing.T) {
	SetSecret([]byte("test-secret"))
	router := gin.New()
	router.GET("/protected", AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer este.no.es.un.jwt.valido")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, got %d", w.Code)
	}
}

// TestAuthRequired_FirmaDiferente verifica que un token firmado con otra clave devuelve 401.
func TestAuthRequired_FirmaDiferente(t *testing.T) {
	SetSecret([]byte("clave-servidor"))
	tokenStr := buildToken([]byte("clave-diferente"), jwt.MapClaims{
		"sub":  "user-1",
		"role": "user",
		"ver":  float64(1),
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	router := gin.New()
	router.GET("/protected", AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, got %d", w.Code)
	}
}

// TestAuthRequired_TokenExpirado verifica que un token con exp en el pasado devuelve 401.
func TestAuthRequired_TokenExpirado(t *testing.T) {
	secret := []byte("clave-servidor")
	SetSecret(secret)
	tokenStr := buildToken(secret, jwt.MapClaims{
		"sub":  "user-1",
		"role": "user",
		"ver":  float64(1),
		"exp":  time.Now().Add(-time.Hour).Unix(), // ya expiró
	})

	router := gin.New()
	router.GET("/protected", AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, got %d", w.Code)
	}
}

// TestAdminRequired_RolUsuario verifica que un usuario con rol "user" es rechazado.
func TestAdminRequired_RolUsuario(t *testing.T) {
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	}, AdminRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("esperado 403, got %d", w.Code)
	}
}

// TestAdminRequired_RolAdmin verifica que un admin pasa el middleware.
func TestAdminRequired_RolAdmin(t *testing.T) {
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	}, AdminRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, got %d", w.Code)
	}
}
