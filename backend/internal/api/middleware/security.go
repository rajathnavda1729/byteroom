package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// SecurityHeaders injects hardened HTTP security headers on every response.
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-XSS-Protection", "1; mode=block")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Content-Security-Policy",
				"default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
					"style-src 'self' 'unsafe-inline'; img-src 'self' data: blob: https:; "+
					"connect-src 'self' ws: wss:; font-src 'self' data:;")
			next.ServeHTTP(w, r)
		})
	}
}

// ipLimiter stores a per-IP rate limiter.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func newIPLimiter(rps int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    rps * 2,
	}
}

func (il *ipLimiter) get(ip string) *rate.Limiter {
	il.mu.Lock()
	defer il.mu.Unlock()
	if l, ok := il.limiters[ip]; ok {
		return l
	}
	l := rate.NewLimiter(il.rps, il.burst)
	il.limiters[ip] = l
	return l
}

// RateLimiter returns a middleware that limits each client IP to rps
// requests per second (with burst = 2×rps).
func RateLimiter(rps int) func(http.Handler) http.Handler {
	il := newIPLimiter(rps)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
				ip = xf
			}
			if !il.get(ip).Allow() {
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
