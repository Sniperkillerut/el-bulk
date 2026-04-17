package service

import (
	"context"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type CategoryService struct {
	Store *store.CategoryStore
	Audit Auditer
}

func NewCategoryService(s *store.CategoryStore, a Auditer) *CategoryService {
	return &CategoryService{Store: s, Audit: a}
}

func (s *CategoryService) List(ctx context.Context, isAdmin bool) ([]models.CustomCategory, error) {
	logger.TraceCtx(ctx, "Entering CategoryService.List | Admin: %v", isAdmin)
	categories, err := s.Store.ListWithCount(ctx, isAdmin)
	if err != nil {
		return nil, err
	}
	s.populateHotNew(ctx, categories)

	for i := range categories {
		categories[i].Redact(isAdmin)
	}

	return categories, nil
}

func (s *CategoryService) Create(ctx context.Context, input models.CustomCategoryInput) (*models.CustomCategory, error) {
	logger.TraceCtx(ctx, "Entering CategoryService.Create | Name: %s", input.Name)
	cat, err := s.Store.Create(ctx, input)
	if err == nil {
		s.Audit.LogAction(ctx, "CREATE_CATEGORY", "category", cat.ID, models.JSONB{"input": input})
	}
	return cat, err
}

func (s *CategoryService) Update(ctx context.Context, id string, updates map[string]interface{}) (*models.CustomCategory, error) {
	logger.TraceCtx(ctx, "Entering CategoryService.Update | ID: %s", id)
	before, _ := s.Store.GetByID(ctx, id)
	cat, err := s.Store.Update(ctx, id, updates)
	if err == nil {
		s.Audit.LogAction(ctx, "UPDATE_CATEGORY", "category", id, models.JSONB{
			"before": before,
			"after":  cat, // Captured full updated category snapshot
		})
	}
	return cat, err
}

func (s *CategoryService) Delete(ctx context.Context, id string) error {
	logger.TraceCtx(ctx, "Entering CategoryService.Delete | ID: %s", id)
	before, _ := s.Store.GetByID(ctx, id)
	
	// Deep Capture: Get all product mappings before they are cascaded away
	mappings, _ := s.Store.GetProductMappings(ctx, id)
	
	err := s.Store.Delete(ctx, id)
	if err == nil {
		s.Audit.LogAction(ctx, "DELETE_CATEGORY", "category", id, models.JSONB{
			"deleted": before,
			"product_mappings": mappings,
		})
	}
	return err
}

func (s *CategoryService) populateHotNew(ctx context.Context, categories []models.CustomCategory) {
	if len(categories) == 0 {
		return
	}
	tenDaysAgo := time.Now().AddDate(0, 0, -10)
	
	ids := make([]string, 0, len(categories))
	for i := range categories {
		if categories[i].CreatedAt != nil && categories[i].CreatedAt.After(tenDaysAgo) {
			categories[i].IsNew = true
		}
		ids = append(ids, categories[i].ID)
	}

	if len(ids) == 0 {
		return
	}

	query, args, err := sqlx.In(`
		SELECT pc.category_id
		FROM order_item oi
		JOIN "order" o ON oi.order_id = o.id
		JOIN product_category pc ON oi.product_id = pc.product_id
		WHERE o.created_at > (now() - interval '7 days')
		  AND pc.category_id IN (?)
		GROUP BY pc.category_id
		HAVING SUM(oi.quantity) >= 5
	`, ids)

	if err != nil {
		return
	}

	var hotIDs []string
	if err := s.Store.DB.SelectContext(ctx, &hotIDs, s.Store.DB.Rebind(query), args...); err == nil {
		hotMap := make(map[string]bool)
		for _, id := range hotIDs {
			hotMap[id] = true
		}
		for i := range categories {
			if hotMap[categories[i].ID] {
				categories[i].IsHot = true
			}
		}
	}
}
