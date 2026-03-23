package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

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
	isAdmin := strings.Contains(r.URL.Path, "/admin/")
	
	query := `
		SELECT c.id, c.name, c.slug, c.is_active, c.created_at, COUNT(pc.product_id) as item_count
		FROM custom_categories c
		LEFT JOIN product_categories pc ON c.id = pc.category_id
	`
	if !isAdmin {
		query += " WHERE c.is_active = true "
	}
	query += `
		GROUP BY c.id, c.name, c.slug, c.is_active, c.created_at
		ORDER BY c.name
	`
	
	err := h.DB.Select(&categories, query)
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if categories == nil {
		categories = []models.CustomCategory{}
	}
	jsonOK(w, categories)
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
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
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

	var cat models.CustomCategory
	err := h.DB.QueryRowx(
		"INSERT INTO custom_categories (name, slug, is_active) VALUES ($1, $2, $3) RETURNING *",
		input.Name, slug, isActive,
	).StructScan(&cat)

	if err != nil {
		// handle unique constraint violation
		if strings.Contains(err.Error(), "unique constraint") {
			jsonError(w, "Category name or slug already exists", http.StatusConflict)
			return
		}
		jsonError(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, cat)
}

// PUT /api/admin/categories/:id
func (h *CategoriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.CustomCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid input", http.StatusBadRequest)
		return
	}

	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	var cat models.CustomCategory
	// We use COALESCE or similar if we wanted partial updates, but here we assume full input for simple CRUD
	query := `UPDATE custom_categories SET name = $1, slug = $2`
	args := []interface{}{input.Name, slug}
	
	if input.IsActive != nil {
		query += `, is_active = $3 WHERE id = $4`
		args = append(args, *input.IsActive, id)
	} else {
		query += ` WHERE id = $3`
		args = append(args, id)
	}
	
	query += ` RETURNING *`
	
	err := h.DB.QueryRowx(query, args...).StructScan(&cat)

	if err == sql.ErrNoRows {
		jsonError(w, "Category not found", http.StatusNotFound)
		return
	} else if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			jsonError(w, "Category name or slug already exists", http.StatusConflict)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	jsonOK(w, cat)
}

// DELETE /api/admin/categories/:id
func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.DB.Exec("DELETE FROM custom_categories WHERE id = $1", id)
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		jsonError(w, "Category not found", http.StatusNotFound)
		return
	}

	jsonOK(w, map[string]string{"message": "Category deleted"})
}
