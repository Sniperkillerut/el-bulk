package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/el-bulk/backend/utils/authutil"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

type UserAuthHandler struct {
	Service *service.AuthService
}

func NewUserAuthHandler(s *service.AuthService) *UserAuthHandler {
	return &UserAuthHandler{Service: s}
}

func getProviderConfig(provider string) *oauth2.Config {
	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("API_URL") + "/api/auth/google/callback",
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
			RedirectURL:  os.Getenv("API_URL") + "/api/auth/facebook/callback",
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
	linkedCustomerID, linkErr := h.Service.FindLinkedCustomerID(provider, finalID)

	currentUserID, _ := r.Context().Value(middleware.UserIDKey).(string)
	if currentUserID != "" {
		exists, err := h.Service.CustomerExists(currentUserID)
		if err != nil || !exists {
			logger.Warn("Stale session for deleted user %s. Treating as guest.", currentUserID)
			currentUserID = "" // Revert to guest flow
		}
	}

	if currentUserID != "" {
		// LINKING FLOW
		if linkErr == nil {
			if linkedCustomerID == currentUserID {
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+redirectURL, http.StatusFound)
				return
			}
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/profile?error=already_linked", http.StatusTemporaryRedirect)
			return
		}

		if err := h.Service.LinkProvider(currentUserID, provider, finalID); err != nil {
			logger.Error("Failed to link account: %v", err)
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/profile?error=link_failed", http.StatusTemporaryRedirect)
			return
		}

		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+redirectURL, http.StatusFound)
		return
	}

	// LOGIN / SIGNUP FLOW
	var customerID string
	found := false
	if linkErr == nil {
		customer, err := h.Service.GetCustomerByID(linkedCustomerID)
		if err == nil {
			customerID = customer.ID
			found = true
		} else {
			logger.Warn("Customer linked in auth but missing in database: %s. Cleaning up orphan record.", linkedCustomerID)
			h.Service.CleanOrphanAuth(provider, finalID)
		}
	}

	if !found {
		customer, err := h.Service.GetCustomerByEmail(finalEmail)
		if err != nil {
			// First time user, create account
			customer, err = h.Service.CreateCustomerWithAuth(finalFirstName, finalLastName, finalEmail, finalAvatar, provider, finalID)
			if err != nil {
				logger.Error("Failed to create user: %v", err)
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/?error=db_error", http.StatusTemporaryRedirect)
				return
			}
			customerID = customer.ID
		} else {
			// Found by email, link this provider to it
			if err := h.Service.LinkProviderIfNotExists(customer.ID, provider, finalID); err != nil {
				logger.Error("Failed to link provider to existing user by email: %v", err)
			}
			customerID = customer.ID
		}
	}

	// Generate JWT Token
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub": customerID,
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
		Domain:   authutil.GetCookieDomain(),
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

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

	customer, err := h.Service.GetMe(userID)
	if err != nil {
		render.Error(w, "User not found", http.StatusNotFound)
		return
	}

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

	if err := h.Service.UpdateProfile(userID, input.Phone, input.IDNumber, input.Address); err != nil {
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
		Domain:   authutil.GetCookieDomain(),
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
	render.Success(w, map[string]string{"message": "Logged out"})
}
