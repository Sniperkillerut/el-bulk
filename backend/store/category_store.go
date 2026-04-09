package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type CategoryStore struct {
	*BaseStore[models.CustomCategory]
}

func NewCategoryStore(db *sqlx.DB) *CategoryStore {
	return &CategoryStore{
		BaseStore: NewBaseStore[models.CustomCategory](db, "custom_category"),
	}
}

func (s *CategoryStore) Create(input models.CustomCategoryInput) (*models.CustomCategory, error) {
	// Custom implementation for Create to handle specific logic like slug generation if needed,
	// or just mapping the input struct to the DB.
	var cat models.CustomCategory
	
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

	query := `
		INSERT INTO custom_category (name, slug, is_active, show_badge, searchable, bg_color, text_color, icon) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING *`
	
	err := s.DB.QueryRowx(query, 
		input.Name, input.Slug, isActive, showBadge, searchable, input.BgColor, input.TextColor, input.Icon,
	).StructScan(&cat)
	
	return &cat, err
}

func (s *CategoryStore) ListWithCount(isAdmin bool) ([]models.CustomCategory, error) {
	var categories []models.CustomCategory
	
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
	
	err := s.DB.Select(&categories, query)
	return categories, err
}
