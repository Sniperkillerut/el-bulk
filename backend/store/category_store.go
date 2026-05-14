package store

import (
	"context"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/cache"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

const categoryCacheKeyPrefix = "categories_admin_"

type CategoryStore struct {
	*BaseStore[models.CustomCategory]
	cache *cache.Cache[[]models.CustomCategory]
}

func NewCategoryStore(db *sqlx.DB) *CategoryStore {
	return &CategoryStore{
		BaseStore: NewBaseStore[models.CustomCategory](db, "custom_category"),
		cache:     cache.New[[]models.CustomCategory](),
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
		INSERT INTO custom_category (id, name, slug, is_active, show_badge, searchable, bg_color, text_color, icon) 
		VALUES (COALESCE($1, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING *`

	logger.TraceCtx(ctx, "[DB] Executing CreateCategory: %s", query)
	err := s.DB.QueryRowxContext(ctx, query,
		input.ID, input.Name, input.Slug, isActive, showBadge, searchable, input.BgColor, input.TextColor, input.Icon,
	).StructScan(&cat)

	if err != nil {
		logger.ErrorCtx(ctx, "[DB] CreateCategory failed: %v", err)
	} else {
		s.cache.Flush()
	}
	return &cat, err
}

func (s *CategoryStore) ListWithCount(ctx context.Context, isAdmin bool) ([]models.CustomCategory, error) {
	cacheKey := categoryCacheKeyPrefix + fmt.Sprint(isAdmin)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

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
	logger.TraceCtx(ctx, "[DB] Executing ListWithCount (isAdmin=%v) (Cache Miss): %s", isAdmin, query)
	err := s.DB.SelectContext(ctx, &categories, query)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] ListWithCount failed: %v", err)
		return nil, err
	}

	s.cache.Set(cacheKey, categories, 5*time.Minute)

	logger.DebugCtx(ctx, "[DB] ListWithCount took %v", time.Since(start))
	return categories, nil
}

func (s *CategoryStore) GetProductMappings(ctx context.Context, categoryID string) ([]string, error) {
	var ids []string
	query := "SELECT product_id FROM product_category WHERE category_id = $1"
	err := s.DB.SelectContext(ctx, &ids, query, categoryID)
	return ids, err
}

func (s *CategoryStore) BatchAddProducts(ctx context.Context, categoryID string, productIDs []string) error {
	if len(productIDs) == 0 {
		return nil
	}

	// Use a transaction for safety
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, pid := range productIDs {
		if _, err := stmt.ExecContext(ctx, pid, categoryID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	s.cache.Flush()
	return nil
}
