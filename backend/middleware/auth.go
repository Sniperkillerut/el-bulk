package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AdminContextKey contextKey = "admin_id"
const IsAdminKey contextKey = "is_admin"

func AdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom header for CSRF protection on state-changing requests
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
				http.Error(w, `{"error":"CSRF protection: missing or invalid X-Requested-With header"}`, http.StatusForbidden)
				return
			}
		}

		var tokenStr string
		cookie, err := r.Cookie("admin_token")
		if err == nil {
			logger.Trace("[Auth] Found admin_token cookie")
			tokenStr = cookie.Value
		} else {
			logger.Trace("[Auth] admin_token cookie NOT found: %v", err)
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenStr == "" {
			logger.Trace("[Auth] No authentication token providing for request: %s %s", r.Method, r.URL.Path)
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error":"invalid claims"}`, http.StatusUnauthorized)
			return
		}

		adminID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, `{"error":"invalid claims"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AdminContextKey, adminID)
		ctx = context.WithValue(ctx, IsAdminKey, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
