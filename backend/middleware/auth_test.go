package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAdminAuth(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	// Helper to create a token
	createToken := func(adminID string, exp time.Duration) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": adminID,
			"exp": time.Now().Add(exp).Unix(),
		})
		s, _ := token.SignedString([]byte(secret))
		return s
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedAdmin  string
		method         string
		csrfHeader     string
	}{
		{
			name:           "Valid Token",
			authHeader:     "Bearer " + createToken("admin123", time.Hour),
			expectedStatus: http.StatusOK,
			expectedAdmin:  "admin123",
		},
		{
			name:           "Missing Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Prefix",
			authHeader:     "Basic abc",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Expired Token",
			authHeader:     "Bearer " + createToken("admin123", -time.Hour),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "POST - Missing CSRF Header",
			authHeader:     "Bearer " + createToken("admin123", time.Hour),
			method:         "POST",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST - Valid CSRF Header",
			authHeader:     "Bearer " + createToken("admin123", time.Hour),
			method:         "POST",
			csrfHeader:     "XMLHttpRequest",
			expectedStatus: http.StatusOK,
			expectedAdmin:  "admin123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				adminID := r.Context().Value(AdminContextKey).(string)
				assert.Equal(t, tt.expectedAdmin, adminID)
				w.WriteHeader(http.StatusOK)
			})

			handlerToTest := AdminAuth(nextHandler)

			method := "GET"
			if tt.method != "" {
				method = tt.method
			}

			req := httptest.NewRequest(method, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.csrfHeader != "" {
				req.Header.Set("X-Requested-With", tt.csrfHeader)
			}
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestAdminAuth_MissingSecret(t *testing.T) {
	// Ensure internal server error if JWT_SECRET is not set
	os.Unsetenv("JWT_SECRET")
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "admin-test",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte("some-secret"))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := AdminAuth(nextHandler)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
