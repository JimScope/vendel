package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

type entry struct {
	count    int
	resetAt  time.Time
}

// RateLimiter provides IP-based rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*entry
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a rate limiter with the given request limit per window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*entry),
		limit:   limit,
		window:  window,
	}
	// Cleanup stale entries every window period
	go func() {
		for {
			time.Sleep(window)
			rl.mu.Lock()
			now := time.Now()
			for ip, e := range rl.entries {
				if now.After(e.resetAt) {
					delete(rl.entries, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

// Check tests the rate limit and returns a 429 response if exceeded, or nil if allowed.
func (rl *RateLimiter) Check(e *core.RequestEvent) error {
	ip := e.RealIP()

	rl.mu.Lock()
	now := time.Now()
	ent, ok := rl.entries[ip]
	if !ok || now.After(ent.resetAt) {
		ent = &entry{count: 0, resetAt: now.Add(rl.window)}
		rl.entries[ip] = ent
	}
	ent.count++
	exceeded := ent.count > rl.limit
	rl.mu.Unlock()

	if exceeded {
		return e.JSON(http.StatusTooManyRequests, map[string]string{
			"detail": "Rate limit exceeded. Try again later.",
		})
	}

	return nil
}
