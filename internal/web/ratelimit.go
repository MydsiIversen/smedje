package web

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimiter tracks per-IP request rates using token buckets.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

// visitor holds the rate limiter and last-seen time for a single IP.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// newRateLimiter creates a rateLimiter and starts a background goroutine
// that cleans up stale entries every 5 minutes.
func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
	}
	go rl.cleanup()
	return rl
}

// allow returns true if the request from ip is within the rate limit.
// Each IP gets a token bucket refilling at 1 token/second with a burst
// capacity of 60 (roughly 60 requests per minute).
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{
			limiter: rate.NewLimiter(rate.Limit(1), 60),
		}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

// cleanup removes visitors that haven't been seen for more than 10
// minutes. It runs in a loop every 5 minutes.
func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// rateLimitMiddleware returns 429 when a visitor exceeds the rate limit.
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !s.limiter.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":               "rate_limited",
				"retry_after_seconds": 30,
				"message":             "Demo rate limit reached. Install Smedje locally for unlimited use: go install github.com/MydsiIversen/smedje/cmd/smedje@latest",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// extractIP returns the client IP from the request. It checks
// X-Forwarded-For first (for requests behind a reverse proxy like
// Caddy), falling back to RemoteAddr.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP, which is the original client.
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
