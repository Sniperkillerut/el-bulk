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

func TestRequireUserAuth(t *testing.T) {
	secret := "test-user-secret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	createToken := func(userID string, exp time.Duration) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"exp": time.Now().Add(exp).Unix(),
		})
		s, _ := token.SignedString([]byte(secret))
		return s
	}

	tests := []struct {
		name           string
		cookieValue    string
		method         string
		csrfHeader     string
		expectedStatus int
		expectedUser   string
	}{
		{
			name:           "Valid Cookie",
			cookieValue:    createToken("user123", time.Hour),
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedUser:   "user123",
		},
		{
			name:           "Missing Cookie",
			cookieValue:    "",
			method:         "GET",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "POST - Missing CSRF Header",
			cookieValue:    createToken("user123", time.Hour),
			method:         "POST",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST - Valid CSRF Header",
			cookieValue:    createToken("user123", time.Hour),
			method:         "POST",
			csrfHeader:     "XMLHttpRequest",
			expectedStatus: http.StatusOK,
			expectedUser:   "user123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value(UserIDKey).(string)
				assert.Equal(t, tt.expectedUser, userID)
				w.WriteHeader(http.StatusOK)
			})

			handlerToTest := RequireUserAuth(nextHandler)

			method := tt.method
			if method == "" {
				method = "GET"
			}

			req := httptest.NewRequest(method, "/", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{Name: "user_token", Value: tt.cookieValue})
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
