package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/el-bulk/backend/utils/logger"
)

type client struct {
	lastSeen time.Time
	count    int
}

var (
	clients = make(map[string]*client)
	mu      sync.Mutex
)

// RateLimit is a simple in-memory rate limiter.
// limit: number of requests allowed
// window: time duration for the limit
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	// Cleanup goroutine to prevent memory leaks
	go func() {
		for {
			time.Sleep(window)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > window {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get real client IP — behind Caddy proxy, RemoteAddr is always the proxy IP.
			// Read X-Real-IP first (set by Caddy), fallback to X-Forwarded-For, then RemoteAddr.
			ip := r.Header.Get("X-Real-IP")
			if ip == "" {
				if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
					ip = strings.Split(fwd, ",")[0]
					ip = strings.TrimSpace(ip)
				}
			}
			if ip == "" {
				ip = r.RemoteAddr
			}

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				clients[ip] = &client{lastSeen: time.Now(), count: 1}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if time.Since(c.lastSeen) > window {
				c.count = 1
				c.lastSeen = time.Now()
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if c.count >= limit {
				mu.Unlock()
				logger.Warn("Rate limit exceeded for IP: %s", ip)
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				http.Error(w, `{"error":"Rate limit exceeded. Please try again later."}`, http.StatusTooManyRequests)
				return
			}

			c.count++
			c.lastSeen = time.Now()
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
