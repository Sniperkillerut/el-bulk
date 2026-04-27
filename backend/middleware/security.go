package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// SecurityHeaders adds common security headers to the response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME-sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable browser XSS filter
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS (Strict-Transport-Security)
		origin := os.Getenv("FRONTEND_ORIGIN")
		if strings.HasPrefix(origin, "https://") {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Content Security Policy (Basic)
		// We allow:
		// - self: our own domain
		// - fonts.googleapis.com, fonts.gstatic.com: for Google Fonts
		// - www.googletagmanager.com: if you use analytics
		// - connect.facebook.net, t.contentsquare.net: Third party marketing scripts
		// - images from our own domain + any external card images (standardized later)
		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com https://connect.facebook.net https://t.contentsquare.net",
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
			"font-src 'self' https://fonts.gstatic.com data:",
			fmt.Sprintf("img-src 'self' data: https: %s", origin),
			"connect-src 'self' https://www.facebook.com https://connect.facebook.net https://t.contentsquare.net " + origin,
			"frame-ancestors 'none'",
		}
		w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))

		next.ServeHTTP(w, r)
	})
}
