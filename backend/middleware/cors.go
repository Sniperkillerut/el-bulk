package middleware

import (
	"net/http"
	"os"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("FRONTEND_ORIGIN")
		if origin == "" {
			origin = "http://localhost:3000"
		}

		// Allow multiple origins
		requestOrigin := r.Header.Get("Origin")
		allowedOrigins := strings.Split(origin, ",")
		
		allowed := false
		if requestOrigin != "" {
			for _, o := range allowedOrigins {
				if strings.TrimSpace(o) == requestOrigin {
					allowed = true
					break
				}
			}
			// Fallback for dev: allow 127.0.0.1 and localhost variations if not strictly matched
			// Only apply this in non-production environments for better security
			if !allowed && os.Getenv("APP_ENV") != "production" {
				if requestOrigin == "http://localhost" ||
					requestOrigin == "http://127.0.0.1" ||
					strings.HasPrefix(requestOrigin, "http://localhost:") ||
					strings.HasPrefix(requestOrigin, "http://127.0.0.1:") {
					allowed = true
				}
			}
		}

		if allowed || requestOrigin == "" {
			if requestOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
