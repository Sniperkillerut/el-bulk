package handlers

import (
"github.com/el-bulk/backend/utils/render"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
)

type UserAuthHandler struct {
	DB *sqlx.DB
}

func NewUserAuthHandler(db *sqlx.DB) *UserAuthHandler {
	return &UserAuthHandler{DB: db}
}

func getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("FRONTEND_ORIGIN") + "/api/auth/google/callback", // Backend handles this directly or via reverse proxy
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GET /api/auth/google/login
func (h *UserAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	config := getOAuthConfig()
	// Create a state token (should ideally be CSRF protected and saved in a short-lived cookie)
	state := "elbulk-oauth-state"
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GET /api/auth/google/callback
func (h *UserAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != "elbulk-oauth-state" {
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	config := getOAuthConfig()

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		logger.Error("OAuth exchange failed: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	// Fetch user info
	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.Error("Failed to get user info: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=userinfo_failed", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Error("Failed to parse user info: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=parse_failed", http.StatusTemporaryRedirect)
		return
	}

	// Upsert customer in database
	var customer models.Customer
	err = h.DB.Get(&customer, "SELECT * FROM customer WHERE auth_provider_id = $1 AND auth_provider = 'google'", userInfo.ID)

	if err != nil {
		// Not found by Google ID, try email
		err = h.DB.Get(&customer, "SELECT * FROM customer WHERE email = $1", userInfo.Email)
		if err != nil {
			// First time user, create account
			query := `
				INSERT INTO customer (first_name, last_name, email, auth_provider, auth_provider_id, avatar_url)
				VALUES ($1, $2, $3, 'google', $4, $5)
				RETURNING *
			`
			err = h.DB.Get(&customer, query, userInfo.GivenName, userInfo.FamilyName, userInfo.Email, userInfo.ID, userInfo.Picture)
			if err != nil {
				logger.Error("Failed to create user: %v", err)
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=db_error", http.StatusTemporaryRedirect)
				return
			}
		} else {
			// Update existing user with Google Auth info
			query := `
				UPDATE customer SET auth_provider = 'google', auth_provider_id = $1, avatar_url = $2
				WHERE id = $3
				RETURNING *
			`
			err = h.DB.Get(&customer, query, userInfo.ID, userInfo.Picture, customer.ID)
			if err != nil {
				logger.Error("Failed to update user: %v", err)
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=db_error", http.StatusTemporaryRedirect)
				return
			}
		}
	}

	// Generate JWT Token
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub": customer.ID,
		"exp": time.Now().Add(24 * 7 * time.Hour).Unix(), // 7 days
		"iat": time.Now().Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		logger.Error("Failed to sign user token: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=token_failed", http.StatusTemporaryRedirect)
		return
	}

	// Set Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_token",
		Value:    signed,
		Path:     "/",
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Set to true in prod
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect back to frontend
	http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/", http.StatusFound)
}

// GET /api/auth/me
func (h *UserAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	var customer models.Customer
	err := h.DB.Get(&customer, "SELECT * FROM customer WHERE id = $1", userID)
	if err != nil {
		render.Error(w, "User not found", http.StatusNotFound)
		return
	}

	render.Success(w, customer)
}

// POST /api/auth/logout
func (h *UserAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	render.Success(w, map[string]string{"message": "Logged out"})
}
