package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/httputil"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
)

type CategoriesHandler struct {
	Service *service.CategoryService
}

func NewCategoriesHandler(s *service.CategoryService) *CategoriesHandler {
	return &CategoriesHandler{Service: s}
}

// GET /api/admin/categories
func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering CategoriesHandler.List")
	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)
	categories, err := h.Service.List(r.Context(), isAdmin)
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, categories)
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

// POST /api/admin/categories
func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering CategoriesHandler.Create")
	var input models.CustomCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode category input: %v", err)
		render.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		render.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if input.Slug == "" {
		input.Slug = generateSlug(input.Name)
	}

	cat, err := h.Service.Create(r.Context(), input)

	if err != nil {
		// handle unique constraint violation
		if strings.Contains(err.Error(), "unique constraint") {
			render.Error(w, "Category name or slug already exists", http.StatusConflict)
			return
		}
		render.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, cat)
}

// PUT /api/admin/categories/:id
func (h *CategoriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering CategoriesHandler.Update | ID: %s", id)
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Invalid Category ID format", http.StatusBadRequest)
		return
	}
	var input models.CustomCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode category update for %s: %v", id, err)
		render.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	updates := make(map[string]interface{})
	updates["name"] = input.Name
	updates["slug"] = slug

	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.ShowBadge != nil {
		updates["show_badge"] = *input.ShowBadge
	}
	if input.Searchable != nil {
		updates["searchable"] = *input.Searchable
	}
	if input.BgColor != nil {
		updates["bg_color"] = input.BgColor
	}
	if input.TextColor != nil {
		updates["text_color"] = input.TextColor
	}
	if input.Icon != nil {
		updates["icon"] = input.Icon
	}

	cat, err := h.Service.Update(r.Context(), id, updates)

	if err == sql.ErrNoRows {
		render.Error(w, "Category not found", http.StatusNotFound)
		return
	} else if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			render.Error(w, "Category name or slug already exists", http.StatusConflict)
			return
		}
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, cat)
}

// DELETE /api/admin/categories/:id
func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering CategoriesHandler.Delete | ID: %s", id)
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Invalid Category ID format", http.StatusBadRequest)
		return
	}
	err := h.Service.Delete(r.Context(), id)
	if err != nil {
		if err.Error() == "no rows deleted" {
			render.Error(w, "Category not found", http.StatusNotFound)
			return
		}
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "Category deleted"})
}
