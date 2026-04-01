package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	os.Setenv("FRONTEND_ORIGIN", "https://example.com")
	defer os.Unsetenv("FRONTEND_ORIGIN")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := SecurityHeaders(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	// Check headers
	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rr.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
	assert.Contains(t, rr.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	assert.Contains(t, rr.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func TestSecurityHeaders_NoHTTPS(t *testing.T) {
	os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")
	defer os.Unsetenv("FRONTEND_ORIGIN")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := SecurityHeaders(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	// HSTS should NOT be present
	assert.Empty(t, rr.Header().Get("Strict-Transport-Security"))
}
