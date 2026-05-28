package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// C is the shared in-memory cache with a default TTL of 2 minutes
// and a cleanup interval of 5 minutes.
var C = gocache.New(2*time.Minute, 5*time.Minute)

// Invalidate removes one or more keys from the cache.
func Invalidate(keys ...string) {
	for _, k := range keys {
		C.Delete(k)
	}
}
