package handlers

import (
	"github.com/el-bulk/backend/utils/render"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
)

type AdminHandler struct {
	Service *service.AdminService
}

func NewAdminHandler(s *service.AdminService) *AdminHandler {
	return &AdminHandler{Service: s}
}

// POST /api/admin/login
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering AdminHandler.Login")
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode login request: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	admin, err := h.Service.GetByUsername(req.Username)
	if err != nil {
		// Differentiate between user not found and other errors (like DB connection)
		if strings.Contains(err.Error(), "no rows") {
			logger.Warn("Failed login attempt: user '%s' not found", req.Username)
		} else {
			logger.Error("Database error during login for '%s': %v", req.Username, err)
		}
		render.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
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
		render.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	isSecure := strings.HasPrefix(os.Getenv("FRONTEND_ORIGIN"), "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    signed,
		Path:     "/",
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
	logger.Trace("Entering AdminHandler.Logout")
	isSecure := strings.HasPrefix(os.Getenv("FRONTEND_ORIGIN"), "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Path:     "/",
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

	logger.Info("Log level hot-changed to: %s", newLevel.String())
	render.Success(w, map[string]string{
		"message": "Log level updated successfully",
		"level":   newLevel.String(),
	})
}
