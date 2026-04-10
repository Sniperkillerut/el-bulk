package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/crypto"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
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
	logger.Trace("Entering UserAuthHandler.Login | Provider: %s", provider)
	config := getProviderConfig(provider)
	if config == nil {
		render.Error(w, "Unsupported auth provider", http.StatusBadRequest)
		return
	}
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" || !strings.HasPrefix(redirectURL, "/") ||
		strings.Contains(redirectURL, "..") || strings.Contains(redirectURL, "//") {
		redirectURL = "/"
	}

	state := "elbulk-oauth-state|" + provider + "|" + redirectURL
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GET /api/auth/{provider}/callback
func (h *UserAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	logger.Trace("Entering UserAuthHandler.Callback | Provider: %s", provider)
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

	// Fetch and Normalize User Info based on provider
	client := config.Client(context.Background(), token)
	var finalFirstName, finalLastName, finalEmail, finalAvatar, finalID string

	switch provider {
	case "google":
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			logger.Error("Failed to get Google user info: %v", err)
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=userinfo_failed", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

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

	case "facebook":
		resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,first_name,last_name,email,picture.type(large)")
		if err != nil {
			logger.Error("Failed to get Facebook user info: %v", err)
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=userinfo_failed", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

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

	default:
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=unsupported_provider", http.StatusTemporaryRedirect)
		return
	}

	// 1. Check if identity is already linked to SOMEONE
	var linkedCustomerID string
	err = h.DB.Get(&linkedCustomerID, "SELECT customer_id FROM customer_auth WHERE provider = $1 AND provider_id = $2", provider, finalID)
	
	currentUserID, _ := r.Context().Value(middleware.UserIDKey).(string)
	if currentUserID != "" {
		var exists bool
		err := h.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM customer WHERE id = $1)", currentUserID)
		if err != nil || !exists {
			// User has a valid token but does not exist in DB (stale session after data wipe)
			logger.Warn("Stale session for deleted user %s. Treating as guest.", currentUserID)
			currentUserID = "" // Revert to guest flow
		}
	}

	if currentUserID != "" {
		// LINKING FLOW
		if err == nil {
			if linkedCustomerID == currentUserID {
				// Already linked to this user, just return
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+redirectURL, http.StatusFound)
				return
			}
			// Already linked to ANOTHER user
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/profile?error=already_linked", http.StatusTemporaryRedirect)
			return
		}
		
		// Link it
		_, linkErr := h.DB.Exec("INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, $2, $3)", currentUserID, provider, finalID)
		if linkErr != nil {
			logger.Error("Failed to link account: %v", linkErr)
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/profile?error=link_failed", http.StatusTemporaryRedirect)
			return
		}
		
		// Success linking
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+redirectURL, http.StatusFound)
		return
	}

	// LOGIN / SIGNUP FLOW
	var customer models.Customer
	found := false
	if err == nil {
		// Found via customer_auth
		err = h.DB.Get(&customer, "SELECT * FROM customer WHERE id = $1", linkedCustomerID)
		if err == nil {
			found = true
		} else {
			// Orphaned auth record (possibly due to developer deleting users but not auth records)
			// Treat as if not found and let it fall through to signup/email check
			logger.Warn("Customer linked in auth but missing in database: %s. Cleaning up orphan record.", linkedCustomerID)
			h.DB.Exec("DELETE FROM customer_auth WHERE provider = $1 AND provider_id = $2", provider, finalID)
		}
	}

	if !found {
		// Not found in customer_auth (or record was orphan), try email lookup
		err = h.DB.Get(&customer, "SELECT * FROM customer WHERE email = $1", finalEmail)
		if err != nil {
			// First time user, create account
			tx, _ := h.DB.Beginx()
			defer tx.Rollback()

			query := `
				INSERT INTO customer (first_name, last_name, email, avatar_url)
				VALUES ($1, $2, $3, $4)
				RETURNING *
			`
			err = tx.Get(&customer, query, finalFirstName, finalLastName, finalEmail, finalAvatar)
			if err != nil {
				logger.Error("Failed to create user: %v", err)
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=db_error", http.StatusTemporaryRedirect)
				return
			}

			// Add auth link
			_, err = tx.Exec("INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, $2, $3)", customer.ID, provider, finalID)
			if err != nil {
				logger.Error("Failed to create auth link: %v", err)
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=db_error", http.StatusTemporaryRedirect)
				return
			}
			tx.Commit()
		} else {
			// Found by email, link this provider to it
			_, err = h.DB.Exec("INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", customer.ID, provider, finalID)
			if err != nil {
				logger.Error("Failed to link provider to existing user by email: %v", err)
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
	logger.Trace("Entering UserAuthHandler.Me")
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

	// Fetch linked providers
	var providers []string
	err = h.DB.Select(&providers, "SELECT provider FROM customer_auth WHERE customer_id = $1", userID)
	if err == nil {
		customer.LinkedProviders = providers
	} else {
		customer.LinkedProviders = []string{}
	}

	// Decrypt sensitive fields
	customer.Phone = crypto.DecryptSafe(customer.Phone)
	customer.IDNumber = crypto.DecryptSafe(customer.IDNumber)
	customer.Address = crypto.DecryptSafe(customer.Address)

	render.Success(w, customer)
}

// PUT /api/auth/me
func (h *UserAuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering UserAuthHandler.UpdateMe")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	var input struct {
		Phone    string `json:"phone"`
		IDNumber string `json:"id_number"`
		Address  string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Encrypt sensitive fields before saving
	encPhone, _ := crypto.Encrypt(input.Phone)
	encIDNumber, _ := crypto.Encrypt(input.IDNumber)
	encAddress, _ := crypto.Encrypt(input.Address)

	_, err := h.DB.Exec(`
		UPDATE customer SET phone = $1, id_number = $2, address = $3
		WHERE id = $4
	`, encPhone, encIDNumber, encAddress, userID)

	if err != nil {
		logger.Error("Failed to update user %s: %v", userID, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.Me(w, r)
}

// POST /api/auth/logout
func (h *UserAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering UserAuthHandler.Logout")
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
