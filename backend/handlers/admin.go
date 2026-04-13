package handlers

import (
	"github.com/el-bulk/backend/utils/render"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/authutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AdminHandler struct {
	Service      *service.AdminService
	AuditService *service.AuditService
}

func NewAdminHandler(s *service.AdminService, audit *service.AuditService) *AdminHandler {
	return &AdminHandler{
		Service:      s,
		AuditService: audit,
	}
}

// POST /api/admin/login
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering AdminHandler.Login")
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode login request: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	admin, err := h.Service.GetByUsername(r.Context(), req.Username)
	if err != nil {
		// Differentiate between user not found and other errors (like DB connection)
		if strings.Contains(err.Error(), "no rows") {
			logger.WarnCtx(r.Context(), "Failed login attempt: user '%s' not found", req.Username)
		} else {
			logger.ErrorCtx(r.Context(), "Database error during login for '%s': %v", req.Username, err)
		}
		render.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if admin.PasswordHash == nil {
		render.Error(w, "Authentication method not supported for this account", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*admin.PasswordHash), []byte(req.Password)); err != nil {
		render.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		render.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	claims := jwt.MapClaims{
		"sub": admin.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to generate token: %v", err)
		render.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	isSecure := strings.HasPrefix(os.Getenv("FRONTEND_ORIGIN"), "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    signed,
		Path:     "/",
		Domain:   authutil.GetCookieDomain(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	render.Success(w, map[string]string{
		"message": "Logged in successfully",
		"token":   signed,
	})
}

// POST /api/admin/logout
func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering AdminHandler.Logout")
	isSecure := strings.HasPrefix(os.Getenv("FRONTEND_ORIGIN"), "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
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

// GET /api/admin/logs/level
func (h *AdminHandler) GetLogLevel(w http.ResponseWriter, r *http.Request) {
	level := logger.Default.GetLevel()
	render.Success(w, map[string]string{
		"level": level.String(),
	})
}

// PUT /api/admin/logs/level
func (h *AdminHandler) UpdateLogLevel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Level string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newLevel := logger.ParseLevel(req.Level)
	logger.SetLevel(newLevel)

	logger.InfoCtx(r.Context(), "Log level hot-changed to: %s", newLevel.String())
	render.Success(w, map[string]string{
		"message": "Log level updated successfully",
		"level":   newLevel.String(),
	})
}

// GET /api/admin/audit-logs
func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20
	
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	adminID := r.URL.Query().Get("admin_id")
	action := r.URL.Query().Get("action")
	resourceType := r.URL.Query().Get("resource_type")

	logs, total, err := h.AuditService.List(r.Context(), page, pageSize, adminID, action, resourceType)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list audit logs: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]interface{}{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GET /api/admin/auth/google/login
func (h *AdminHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID_ADMIN")
	if clientID == "" {
		clientID = os.Getenv("GOOGLE_CLIENT_ID")
	}
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET_ADMIN")
	if clientSecret == "" {
		clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  os.Getenv("API_URL") + "/api/admin/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	state := "elbulk-admin-oauth-state"
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GET /api/admin/auth/google/callback
func (h *AdminHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != "elbulk-admin-oauth-state" {
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID_ADMIN")
	if clientID == "" {
		clientID = os.Getenv("GOOGLE_CLIENT_ID")
	}
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET_ADMIN")
	if clientSecret == "" {
		clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	}

	code := r.FormValue("code")
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  os.Getenv("API_URL") + "/api/admin/auth/google/callback",
		Endpoint:     google.Endpoint,
	}

	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Admin OAuth exchange failed: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/login?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	client := config.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to get Google admin info: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/login?error=userinfo_failed", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var gInfo struct {
		Email     string `json:"email"`
		Name      string `json:"name"`
		Picture   string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gInfo); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to parse Google admin info: %v", err)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/login?error=parse_failed", http.StatusTemporaryRedirect)
		return
	}

	// 1. Verify Whitelist
	envEmails := os.Getenv("ADMIN_EMAILS")
	if envEmails == "" {
		logger.ErrorCtx(r.Context(), "[AdminAuth] ADMIN_EMAILS environment variable is NOT SET. All OAuth login attempts will fail.")
	}
	allowedEmails := strings.Split(envEmails, ",")
	allowed := false
	for _, e := range allowedEmails {
		if strings.TrimSpace(e) == gInfo.Email {
			allowed = true
			break
		}
	}

	if !allowed {
		logger.WarnCtx(r.Context(), "Unauthorized admin login attempt: %s", gInfo.Email)
		http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/login?error=unauthorized_email", http.StatusTemporaryRedirect)
		return
	}

	// 2. Find or Create Admin
	admin, err := h.Service.GetByEmail(r.Context(), gInfo.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create admin
			username := strings.Split(gInfo.Email, "@")[0]
			newAdmin := models.Admin{
				Username:  username,
				Email:     gInfo.Email,
				AvatarURL: &gInfo.Picture,
			}
			admin, err = h.Service.Create(r.Context(), newAdmin)
			if err != nil {
				logger.ErrorCtx(r.Context(), "[AdminAuth] Failed to create admin via OAuth for %s: %v", gInfo.Email, err)
				http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/login?error=db_error", http.StatusTemporaryRedirect)
				return
			}
			h.AuditService.LogAction(r.Context(), "OAUTH_SIGNUP", "admin", admin.ID, models.JSONB{"email": gInfo.Email})
		} else {
			logger.ErrorCtx(r.Context(), "[AdminAuth] Database error looking up admin %s: %v. Check if 'email' column exists in 'admin' table.", gInfo.Email, err)
			http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/login?error=db_error", http.StatusTemporaryRedirect)
			return
		}
	}

	// 3. Issue Token
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub": admin.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := jwtToken.SignedString([]byte(secret))

	isSecure := strings.HasPrefix(os.Getenv("FRONTEND_ORIGIN"), "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    signed,
		Path:     "/",
		Domain:   authutil.GetCookieDomain(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, os.Getenv("FRONTEND_ORIGIN")+"/admin/dashboard", http.StatusFound)
}
