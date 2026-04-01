package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/el-bulk/backend/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCSRFProtection_Integration(t *testing.T) {
	r := chi.NewRouter()

	// Dummy handler
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Using AdminAuth (which contains CSRF check)
	r.With(middleware.AdminAuth).Handle("/test-admin", successHandler)
	r.With(middleware.RequireUserAuth).Handle("/test-user", successHandler)

	tests := []struct {
		name           string
		url            string
		method         string
		csrfHeader     string
		expectedStatus int
	}{
		{
			name:           "Admin POST - Missing CSRF",
			url:            "/test-admin",
			method:         "POST",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "User POST - Missing CSRF",
			url:            "/test-user",
			method:         "POST",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Admin GET - Missing CSRF (Allowed)",
			url:            "/test-admin",
			method:         "GET",
			expectedStatus: http.StatusUnauthorized, // Fails auth, but passes CSRF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, bytes.NewBuffer([]byte("{}")))
			if tt.csrfHeader != "" {
				req.Header.Set("X-Requested-With", tt.csrfHeader)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
