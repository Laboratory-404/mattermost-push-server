package server

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Laboratory-404/mattermost-push-server/pkg/config"
)

// SimpleTokenBucket for rate limiting
type tokenBucket struct {
	rate       float64
	capacity   float64
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

func newTokenBucket(rate float64, capacity float64) *tokenBucket {
	return &tokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastUpdate: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.lastUpdate = now

	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// SecurityAndLoggingMiddleware wraps handlers with security headers, logging, and rate limiting
func SecurityAndLoggingMiddleware(cfg config.ServerConfig, next http.Handler) http.Handler {
	var limiter *tokenBucket
	if cfg.EnableRateLimit {
		limiter = newTokenBucket(cfg.RateLimitPerSec, float64(cfg.RateLimitBurst))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Server", "Mattermost-Push-Proxy/2.0")

		// Rate Limiter Check
		if limiter != nil && !limiter.allow() {
			http.Error(w, `{"status":"FAIL","error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		// Panic Recovery
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %s %s: %v", r.Method, r.URL.Path, err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"status":"FAIL","error":"Internal server error"}`))
			}
		}()

		next.ServeHTTP(w, r)

		log.Printf("[HTTP] %s %s | Remote: %s | Duration: %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
