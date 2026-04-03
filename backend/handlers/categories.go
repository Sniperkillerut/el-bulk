package handlers

import (
	"github.com/el-bulk/backend/utils/render"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
)

type CategoriesHandler struct {
	DB *sqlx.DB
}

func NewCategoriesHandler(db *sqlx.DB) *CategoriesHandler {
	return &CategoriesHandler{DB: db}
}

// GET /api/admin/categories
// (Also used by frontend public clients via /api/categories if needed)
func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	var categories []models.CustomCategory
	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)
	
	query := `
		SELECT c.id, c.name, c.slug, c.is_active, c.show_badge, c.searchable, c.bg_color, c.text_color, c.icon, c.created_at, COUNT(pc.product_id) as item_count
		FROM custom_category c
		LEFT JOIN product_category pc ON c.id = pc.category_id
	`
	if !isAdmin {
		query += " WHERE c.is_active = true OR c.searchable = true OR c.show_badge = true "
	}
	query += `
		GROUP BY c.id, c.name, c.slug, c.is_active, c.show_badge, c.searchable, c.bg_color, c.text_color, c.icon, c.created_at
	`
	if !isAdmin {
		query += " HAVING COUNT(pc.product_id) > 0 "
	}
	query += ` ORDER BY c.name `
	
	err := h.DB.Select(&categories, query)
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if categories == nil {
		categories = []models.CustomCategory{}
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
	var input models.CustomCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		render.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	showBadge := true
	if input.ShowBadge != nil {
		showBadge = *input.ShowBadge
	}
	searchable := true
	if input.Searchable != nil {
		searchable = *input.Searchable
	}

	var cat models.CustomCategory
	err := h.DB.QueryRowx(
		"INSERT INTO custom_category (name, slug, is_active, show_badge, searchable, bg_color, text_color, icon) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *",
		input.Name, slug, isActive, showBadge, searchable, input.BgColor, input.TextColor, input.Icon,
	).StructScan(&cat)

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
	var input models.CustomCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	var cat models.CustomCategory
	// We use COALESCE or similar if we wanted partial updates, but here we assume full input for simple CRUD
	query := `UPDATE custom_category SET name = $1, slug = $2`
	args := []interface{}{input.Name, slug}
	
	idx := 3
	if input.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", idx)
		args = append(args, *input.IsActive)
		idx++
	}
	if input.ShowBadge != nil {
		query += fmt.Sprintf(", show_badge = $%d", idx)
		args = append(args, *input.ShowBadge)
		idx++
	}
	if input.Searchable != nil {
		query += fmt.Sprintf(", searchable = $%d", idx)
		args = append(args, *input.Searchable)
		idx++
	}
	if input.BgColor != nil {
		query += fmt.Sprintf(", bg_color = $%d", idx)
		args = append(args, input.BgColor)
		idx++
	}
	if input.TextColor != nil {
		query += fmt.Sprintf(", text_color = $%d", idx)
		args = append(args, input.TextColor)
		idx++
	}
	if input.Icon != nil {
		query += fmt.Sprintf(", icon = $%d", idx)
		args = append(args, input.Icon)
		idx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", idx)
	args = append(args, id)
	
	query += ` RETURNING *`
	
	err := h.DB.QueryRowx(query, args...).StructScan(&cat)

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
	res, err := h.DB.Exec("DELETE FROM custom_category WHERE id = $1", id)
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		render.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	render.Success(w, map[string]string{"message": "Category deleted"})
}
