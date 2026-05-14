package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
)

type SettingsProvider interface {
	GetSettings(ctx context.Context) (models.Settings, error)
}

// Blacklist is a middleware that rejects requests from blocked IP addresses.
func Blacklist(provider SettingsProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get real client IP
			ip := r.Header.Get("X-Real-IP")
			if ip == "" {
				if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
					ip = strings.Split(fwd, ",")[0]
					ip = strings.TrimSpace(ip)
				}
			}
			if ip == "" {
				// Strip port if present
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}

			// Get blocked IPs from settings
			settings, err := provider.GetSettings(r.Context())
			if err != nil {
				// If settings fail, we proceed rather than blocking everyone
				logger.ErrorCtx(r.Context(), "Blacklist middleware failed to get settings: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			if settings.BlockedIPs != "" {
				blockedList := strings.Split(settings.BlockedIPs, ",")
				for _, blocked := range blockedList {
					blocked = strings.TrimSpace(blocked)
					if blocked == "" {
						continue
					}
					if ip == blocked {
						logger.Warn("🚫 Request blocked from blacklisted IP: %s", ip)
						http.Error(w, `{"error":"Access denied. Your IP has been blacklisted."}`, http.StatusForbidden)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
