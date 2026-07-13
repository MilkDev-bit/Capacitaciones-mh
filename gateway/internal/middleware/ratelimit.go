package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipVisitor struct {
	tokens    float64
	lastVisit time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*ipVisitor
	rate       float64 // tokens agregados por segundo
	burst      float64 // máximo de tokens permitidos (capacidad)
}

// NewRateLimiter crea un limitador de tasa usando Token Bucket por IP.
// requestsPerMinute define cuántas peticiones promedio se permiten por minuto.
// burst define el pico máximo instantáneo de peticiones permitidas.
func NewRateLimiter(requestsPerMinute int, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*ipVisitor),
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    float64(burst),
	}

	// Limpieza periódica de IPs inactivas (cada 3 minutos)
	go func() {
		for range time.Tick(3 * time.Minute) {
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.Sub(v.lastVisit) > 5*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

// Handler devuelve la función middleware de Gin para protección Anti-DDoS.
func (rl *RateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &ipVisitor{
				tokens:    rl.burst - 1,
				lastVisit: now,
			}
			rl.visitors[ip] = v
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Recargar tokens en base al tiempo transcurrido
		elapsed := now.Sub(v.lastVisit).Seconds()
		v.tokens += elapsed * rl.rate
		if v.tokens > rl.burst {
			v.tokens = rl.burst
		}
		v.lastVisit = now

		if v.tokens < 1 {
			rl.mu.Unlock()
			c.Header("Retry-After", "10")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "demasiadas peticiones, protección Anti-DDoS activa. Por favor espera un momento.",
			})
			return
		}

		v.tokens--
		rl.mu.Unlock()
		c.Next()
	}
}
