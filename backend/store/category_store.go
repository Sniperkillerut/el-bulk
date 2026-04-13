package store

import (
	"context"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
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

func (s *CategoryStore) Create(ctx context.Context, input models.CustomCategoryInput) (*models.CustomCategory, error) {
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
	
	logger.TraceCtx(ctx, "[DB] Executing CreateCategory: %s", query)
	err := s.DB.QueryRowxContext(ctx, query, 
		input.Name, input.Slug, isActive, showBadge, searchable, input.BgColor, input.TextColor, input.Icon,
	).StructScan(&cat)
	
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] CreateCategory failed: %v", err)
	}
	return &cat, err
}

func (s *CategoryStore) ListWithCount(ctx context.Context, isAdmin bool) ([]models.CustomCategory, error) {
	categories := []models.CustomCategory{}
	
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
	
	start := time.Now()
	logger.TraceCtx(ctx, "[DB] Executing ListWithCount (isAdmin=%v): %s", isAdmin, query)
	err := s.DB.SelectContext(ctx, &categories, query)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] ListWithCount failed: %v", err)
	}
	logger.DebugCtx(ctx, "[DB] ListWithCount took %v", time.Since(start))
	return categories, err
}
