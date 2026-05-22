package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipBucket struct {
	count   int
	resetAt time.Time
}

// RateLimiter es un limitador de tasa por IP en memoria.
// NOTA: en despliegues multi-instancia usar Redis.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	max     int
	window  time.Duration
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*ipBucket),
		max:     max,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.buckets {
			if now.After(b.resetAt) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware devuelve un gin.HandlerFunc que limita peticiones por IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()
		b, ok := rl.buckets[ip]
		now := time.Now()
		if !ok || now.After(b.resetAt) {
			rl.buckets[ip] = &ipBucket{count: 1, resetAt: now.Add(rl.window)}
			rl.mu.Unlock()
			c.Next()
			return
		}
		b.count++
		if b.count > rl.max {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "demasiados intentos, intenta más tarde",
			})
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}
