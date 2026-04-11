package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name            string
		frontendOrigin  string
		requestOrigin   string
		requestMethod   string
		expectedOrigin  string
		expectedStatus  int
	}{
		{
			name:           "Allowed Origin from ENV",
			frontendOrigin: "https://example.com",
			requestOrigin:  "https://example.com",
			requestMethod:  "GET",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disallowed Origin",
			frontendOrigin: "https://example.com",
			requestOrigin:  "https://attacker.com",
			requestMethod:  "GET",
			expectedOrigin: "", // Should not set Access-Control-Allow-Origin
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allow localhost by default",
			frontendOrigin: "https://example.com",
			requestOrigin:  "http://localhost:3000",
			requestMethod:  "GET",
			expectedOrigin: "http://localhost:3000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allow 127.0.0.1 by default",
			frontendOrigin: "https://example.com",
			requestOrigin:  "http://127.0.0.1:8080",
			requestMethod:  "GET",
			expectedOrigin: "http://127.0.0.1:8080",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS Preflight",
			frontendOrigin: "https://example.com",
			requestOrigin:  "https://example.com",
			requestMethod:  "OPTIONS",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Empty Origin (Same origin or direct access)",
			frontendOrigin: "https://example.com",
			requestOrigin:  "",
			requestMethod:  "GET",
			expectedOrigin: "", // No ACAO header when no origin is sent (server-to-server/same-origin)
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Multiple Allowed Origins",
			frontendOrigin: "https://a.com,https://b.com",
			requestOrigin:  "https://b.com",
			requestMethod:  "GET",
			expectedOrigin: "https://b.com",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("FRONTEND_ORIGIN", tt.frontendOrigin)
			defer os.Unsetenv("FRONTEND_ORIGIN")

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handlerToTest := CORS(nextHandler)

			req := httptest.NewRequest(tt.requestMethod, "/", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
		})
	}
}
