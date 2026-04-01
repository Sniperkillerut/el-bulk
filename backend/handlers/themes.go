package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type ThemeHandler struct {
	DB *sqlx.DB
}

func NewThemeHandler(db *sqlx.DB) *ThemeHandler {
	return &ThemeHandler{DB: db}
}

// List returns all themes.
// GET /api/themes
func (h *ThemeHandler) List(w http.ResponseWriter, r *http.Request) {
	var themes []models.Theme
	if err := h.DB.Select(&themes, "SELECT * FROM theme ORDER BY is_system DESC, name ASC"); err != nil {
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

	query := `
		INSERT INTO theme (
			name, bg_page, bg_header, bg_surface, text_main, text_secondary, text_muted, text_on_accent,
			accent_primary, accent_primary_hover, border_main, status_nm, status_lp, status_mp, status_hp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING *
	`
	var theme models.Theme
	err := h.DB.Get(&theme, query,
		input.Name, input.BgPage, input.BgHeader, input.BgSurface,
		input.TextMain, input.TextSecondary, input.TextMuted, input.TextOnAccent,
		input.AccentPrimary, input.AccentPrimaryHover, input.BorderMain,
		input.StatusNM, input.StatusLP, input.StatusMP, input.StatusHP,
	)
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

	query := `
		UPDATE theme SET
			name = $1, bg_page = $2, bg_header = $3, bg_surface = $4,
			text_main = $5, text_secondary = $6, text_muted = $7, text_on_accent = $8,
			accent_primary = $9, accent_primary_hover = $10, border_main = $11,
			status_nm = $12, status_lp = $13, status_mp = $14, status_hp = $15,
			updated_at = now()
		WHERE id = $16
		RETURNING *
	`
	var theme models.Theme
	err := h.DB.Get(&theme, query,
		input.Name, input.BgPage, input.BgHeader, input.BgSurface,
		input.TextMain, input.TextSecondary, input.TextMuted, input.TextOnAccent,
		input.AccentPrimary, input.AccentPrimaryHover, input.BorderMain,
		input.StatusNM, input.StatusLP, input.StatusMP, input.StatusHP,
		id,
	)
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
	
	// Check if system theme
	var isSystem bool
	err := h.DB.Get(&isSystem, "SELECT is_system FROM theme WHERE id = $1", id)
	if err != nil {
		render.Error(w, "theme not found", http.StatusNotFound)
		return
	}
	if isSystem {
		render.Error(w, "cannot delete system theme", http.StatusForbidden)
		return
	}

	_, err = h.DB.Exec("DELETE FROM theme WHERE id = $1", id)
	if err != nil {
		render.Error(w, "failed to delete theme", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "theme deleted"})
}
