package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
)

type ThemeHandler struct {
	Service *service.ThemeService
}

func NewThemeHandler(s *service.ThemeService) *ThemeHandler {
	return &ThemeHandler{Service: s}
}

// List returns all themes.
// GET /api/themes
func (h *ThemeHandler) List(w http.ResponseWriter, r *http.Request) {
	themes, err := h.Service.List()
	if err != nil {
		render.Error(w, "failed to fetch themes", http.StatusInternalServerError)
		return
	}
	render.Success(w, themes)
}

// Create adds a new custom theme.
// POST /api/admin/themes
func (h *ThemeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.ThemeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	theme, err := h.Service.Create(input)
	if err != nil {
		render.Error(w, "failed to create theme: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render.Success(w, theme)
}

// Update modifies an existing theme.
// PUT /api/admin/themes/{id}
func (h *ThemeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.ThemeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	theme, err := h.Service.Update(id, input)
	if err != nil {
		render.Error(w, "failed to update theme: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render.Success(w, theme)
}

// Delete removes a theme (unless it is a system theme).
// DELETE /api/admin/themes/{id}
func (h *ThemeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	err := h.Service.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "system theme") {
			render.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		render.Error(w, "failed to delete theme", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "theme deleted"})
}
