package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/el-bulk/backend/utils/logger"
)

type userContextKey string

const UserIDKey userContextKey = "user_id"

// OptionalUserAuth middleware verifies the JWT token from the user_token cookie.
func OptionalUserAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_token")
		if err != nil {
			// Not logged in is okay for public routes, but if this middleware is applied, it implies protection.
			// Actually, we can use this loosely - if the token is present, we set the ID, otherwise we omit it.
			// But usually, an Auth middleware rejects the request.
			// We will create two: OptionalUserAuth and RequireUserAuth
			next.ServeHTTP(w, r)
			return
		}

		tokenStr := cookie.Value
		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

		if err != nil || !token.Valid {
			// Invalid token - just ignore and treat as guest, but log it
			logger.Warn("Invalid user token in optional auth: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && claims["sub"] != nil {
			ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"].(string))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireUserAuth is strict
func RequireUserAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom header for CSRF protection on state-changing requests
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
				http.Error(w, `{"error":"CSRF protection: missing or invalid X-Requested-With header"}`, http.StatusForbidden)
				return
			}
		}

		cookie, err := r.Cookie("user_token")
		if err != nil {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

		if err != nil || !token.Valid {
			logger.Error("Invalid user token in middleware: %v", err)
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["sub"] == nil {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"].(string))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
