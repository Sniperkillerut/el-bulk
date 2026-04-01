package handlers

import (
"github.com/el-bulk/backend/utils/render"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"golang.org/x/oauth2/facebook"
)

type UserAuthHandler struct {
	DB *sqlx.DB
}

func NewUserAuthHandler(db *sqlx.DB) *UserAuthHandler {
	return &UserAuthHandler{DB: db}
}

func getProviderConfig(provider string) *oauth2.Config {
	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("FRONTEND_ORIGIN") + "/api/auth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	case "facebook":
		return &oauth2.Config{
			ClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
			ClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("FRONTEND_ORIGIN") + "/api/auth/facebook/callback",
			Scopes:       []string{"email", "public_profile"},
			Endpoint:     facebook.Endpoint,
		}
	default:
		return nil
	}
}

// GET /api/auth/{provider}/login
func (h *UserAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	config := getProviderConfig(provider)
	if config == nil {
		render.Error(w, "Unsupported auth provider", http.StatusBadRequest)
		return
	}
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	state := "elbulk-oauth-state|" + provider + "|" + redirectURL
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GET /api/auth/{provider}/callback
func (h *UserAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state := r.FormValue("state")
	if !strings.HasPrefix(state, "elbulk-oauth-state|"+provider+"|") {
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	redirectURL := strings.TrimPrefix(state, "elbulk-oauth-state|"+provider+"|")
	if redirectURL == "" {
		redirectURL = "/"
	}

	code := r.FormValue("code")
	config := getProviderConfig(provider)
	if config == nil {
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=unsupported_provider", http.StatusTemporaryRedirect)
		return
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		logger.Error("OAuth exchange failed: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	// Fetch user info based on provider
	client := config.Client(context.Background(), token)
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	if provider == "facebook" {
		userInfoURL = "https://graph.facebook.com/me?fields=id,name,first_name,last_name,email,picture.type(large)"
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		logger.Error("Failed to get user info: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=userinfo_failed", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	// Normalize User Info
	var finalFirstName, finalLastName, finalEmail, finalAvatar, finalID string

	if provider == "google" {
		var gInfo struct {
			ID         string `json:"id"`
			Email      string `json:"email"`
			GivenName  string `json:"given_name"`
			FamilyName string `json:"family_name"`
			Picture    string `json:"picture"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&gInfo); err != nil {
			logger.Error("Failed to parse Google user info: %v", err)
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=parse_failed", http.StatusTemporaryRedirect)
			return
		}
		finalID = gInfo.ID
		finalFirstName = gInfo.GivenName
		finalLastName = gInfo.FamilyName
		finalEmail = gInfo.Email
		finalAvatar = gInfo.Picture
	} else if provider == "facebook" {
		var fbInfo struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Picture   struct {
				Data struct {
					URL string `json:"url"`
				} `json:"data"`
			} `json:"picture"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&fbInfo); err != nil {
			logger.Error("Failed to parse Facebook user info: %v", err)
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=parse_failed", http.StatusTemporaryRedirect)
			return
		}
		finalID = fbInfo.ID
		finalFirstName = fbInfo.FirstName
		finalLastName = fbInfo.LastName
		finalEmail = fbInfo.Email
		finalAvatar = fbInfo.Picture.Data.URL
	}

	// Upsert customer in database
	var customer models.Customer
	err = h.DB.Get(&customer, "SELECT * FROM customer WHERE auth_provider_id = $1 AND auth_provider = $2", finalID, provider)

	if err != nil {
		// Not found by Provider ID, try email
		err = h.DB.Get(&customer, "SELECT * FROM customer WHERE email = $1", finalEmail)
		if err != nil {
			// First time user, create account
			query := `
				INSERT INTO customer (first_name, last_name, email, auth_provider, auth_provider_id, avatar_url)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING *
			`
			err = h.DB.Get(&customer, query, finalFirstName, finalLastName, finalEmail, provider, finalID, finalAvatar)
			if err != nil {
				logger.Error("Failed to create user: %v", err)
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=db_error", http.StatusTemporaryRedirect)
				return
			}
		} else {
			// Update existing user with Auth info
			query := `
				UPDATE customer SET auth_provider = $1, auth_provider_id = $2, avatar_url = $3
				WHERE id = $4
				RETURNING *
			`
			err = h.DB.Get(&customer, query, provider, finalID, finalAvatar, customer.ID)
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
	isSecure := strings.HasPrefix(os.Getenv("FRONTEND_ORIGIN"), "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "user_token",
		Value:    signed,
		Path:     "/",
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect back to frontend
	http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+redirectURL, http.StatusFound)
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
	isSecure := strings.HasPrefix(os.Getenv("FRONTEND_ORIGIN"), "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "user_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
	render.Success(w, map[string]string{"message": "Logged out"})
}
