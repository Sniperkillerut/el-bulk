package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/authutil"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
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

func redirectWithError(w http.ResponseWriter, r *http.Request, path, errorCode string) {
	http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+path+"?error="+errorCode, http.StatusTemporaryRedirect)
}

func getProviderConfig(provider string) *oauth2.Config {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		// Resilience: Fallback to SITE_URL if API_URL is missing
		apiURL = os.Getenv("SITE_URL")
	}
	if apiURL == "" {
		apiURL = "http://localhost:3000"
	}

	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  apiURL + "/api/auth/google/callback",
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
			RedirectURL:  apiURL + "/api/auth/facebook/callback",
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
	logger.TraceCtx(r.Context(), "Entering UserAuthHandler.Login | Provider: %s", provider)
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
	logger.TraceCtx(r.Context(), "Entering UserAuthHandler.Callback | Provider: %s", provider)
	state := r.FormValue("state")
	if !strings.HasPrefix(state, "elbulk-oauth-state|"+provider+"|") {
		redirectWithError(w, r, "/", "invalid_state")
		return
	}

	redirectURL := strings.TrimPrefix(state, "elbulk-oauth-state|"+provider+"|")
	if redirectURL == "" {
		redirectURL = "/"
	}

	code := r.FormValue("code")
	config := getProviderConfig(provider)
	if config == nil {
		redirectWithError(w, r, "/", "unsupported_provider")
		return
	}

	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		logger.ErrorCtx(r.Context(), "OAuth exchange failed: %v", err)
		redirectWithError(w, r, "/", "exchange_failed")
		return
	}

	// Fetch and Normalize User Info based on provider
	client := config.Client(r.Context(), token)
	var finalFirstName, finalLastName, finalEmail, finalAvatar, finalID string

	switch provider {
	case "google":
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			logger.ErrorCtx(r.Context(), "Failed to get Google user info: %v", err)
			redirectWithError(w, r, "/", "userinfo_failed")
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
			logger.ErrorCtx(r.Context(), "Failed to parse Google user info: %v", err)
			redirectWithError(w, r, "/", "parse_failed")
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
			logger.ErrorCtx(r.Context(), "Failed to get Facebook user info: %v", err)
			redirectWithError(w, r, "/", "userinfo_failed")
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
			logger.ErrorCtx(r.Context(), "Failed to parse Facebook user info: %v", err)
			redirectWithError(w, r, "/", "parse_failed")
			return
		}
		finalID = fbInfo.ID
		finalFirstName = fbInfo.FirstName
		finalLastName = fbInfo.LastName
		finalEmail = fbInfo.Email
		finalAvatar = fbInfo.Picture.Data.URL

	default:
		redirectWithError(w, r, "/", "unsupported_provider")
		return
	}

	// 1. Check if identity is already linked to SOMEONE
	linkedCustomerID, linkErr := h.Service.FindLinkedCustomerID(r.Context(), provider, finalID)

	currentUserID, _ := r.Context().Value(middleware.UserIDKey).(string)
	if currentUserID != "" {
		exists, err := h.Service.CustomerExists(r.Context(), currentUserID)
		if err != nil || !exists {
			logger.WarnCtx(r.Context(), "Stale session for deleted user %s. Treating as guest.", currentUserID)
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
			redirectWithError(w, r, "/profile", "already_linked")
			return
		}

		if err := h.Service.LinkProvider(r.Context(), currentUserID, provider, finalID); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to link account: %v", err)
			redirectWithError(w, r, "/profile", "link_failed")
			return
		}

		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+redirectURL, http.StatusFound)
		return
	}

	// LOGIN / SIGNUP FLOW
	var customerID string
	found := false
	if linkErr == nil {
		customer, err := h.Service.GetCustomerByID(r.Context(), linkedCustomerID)
		if err == nil {
			customerID = customer.ID
			found = true
		} else {
			logger.WarnCtx(r.Context(), "Customer linked in auth but missing in database: %s. Cleaning up orphan record.", linkedCustomerID)
			h.Service.CleanOrphanAuth(r.Context(), provider, finalID)
		}
	}

	if !found {
		customer, err := h.Service.GetCustomerByEmail(r.Context(), finalEmail)
		if err != nil {
			// First time user, create account
			customer, err = h.Service.CreateCustomerWithAuth(r.Context(), finalFirstName, finalLastName, finalEmail, finalAvatar, provider, finalID)
			if err != nil {
				logger.ErrorCtx(r.Context(), "Failed to create user: %v", err)
				redirectWithError(w, r, "/", "db_error")
				return
			}
			customerID = customer.ID
		} else {
			// Found by email, link this provider to it
			if err := h.Service.LinkProviderIfNotExists(r.Context(), customer.ID, provider, finalID); err != nil {
				logger.ErrorCtx(r.Context(), "Failed to link provider to existing user by email: %v", err)
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
		logger.ErrorCtx(r.Context(), "Failed to sign user token: %v", err)
		redirectWithError(w, r, "/", "token_failed")
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
	logger.TraceCtx(r.Context(), "Entering UserAuthHandler.Me")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	customer, err := h.Service.GetMe(r.Context(), userID)
	if err != nil {
		render.Error(w, "User not found", http.StatusNotFound)
		return
	}

	render.Success(w, customer)
}

// PUT /api/auth/me
func (h *UserAuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering UserAuthHandler.UpdateMe")
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

	if err := h.Service.UpdateProfile(r.Context(), userID, input.Phone, input.IDNumber, input.Address); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to update user %s: %v", userID, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.Me(w, r)
}

// POST /api/auth/logout
func (h *UserAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering UserAuthHandler.Logout")
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
