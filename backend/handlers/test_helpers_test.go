package handlers

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
)

type NopAuditer struct{}

func (n *NopAuditer) LogAction(ctx context.Context, action, resourceType, resourceID string, details models.JSONB) {
}
func (n *NopAuditer) List(ctx context.Context, page, pageSize int, adminID, action, resourceType string) ([]models.AuditLog, int, error) {
	return nil, 0, nil
}

// testBountyHandler creates a BountyHandler wired through Store→Service for testing.
func testBountyHandler(db *sqlx.DB) *BountyHandler {
	return NewBountyHandler(service.NewBountyService(store.NewBountyStore(db)))
}

// testAdminHandler creates an AdminHandler wired through Store→Service for testing.
func testAdminHandler(db *sqlx.DB) *AdminHandler {
	as := store.NewAuditStore(db)
	aud := &NopAuditer{}
	ps := store.NewProductStore(db)
	pSvc := service.NewProductService(ps, store.NewTCGStore(db), nil, aud)
	cSvc := service.NewCategoryService(store.NewCategoryStore(db), aud)
	slSvc := service.NewStorageLocationService(store.NewStorageLocationStore(db), aud)
	sSvc := service.NewSettingsService(store.NewSettingsStore(db), aud)
	
	rs := service.NewRevertService(as, aud, ps, pSvc, cSvc, store.NewCategoryStore(db), slSvc, store.NewStorageLocationStore(db), sSvc)
	return NewAdminHandler(service.NewAdminService(store.NewAdminStore(db)), aud, rs)
}

// testNewsletterHandler creates a NewsletterHandler wired through Store→Service for testing.
func testNewsletterHandler(db *sqlx.DB) *NewsletterHandler {
	return NewNewsletterHandler(service.NewNewsletterService(store.NewNewsletterStore(db)))
}

// testStorageHandler creates a StorageHandler wired through Store→Service for testing.
func testStorageHandler(db *sqlx.DB) *StorageHandler {
	return NewStorageHandler(service.NewStorageLocationService(store.NewStorageLocationStore(db), &NopAuditer{}))
}

func testHealthHandler(db *sqlx.DB) *HealthHandler {
	return NewHealthHandler(service.NewHealthService(store.NewHealthStore(db)), "test-version")
}

// testRefreshService creates a RefreshService for testing.
func testRefreshService(db *sqlx.DB) *service.RefreshService {
	return service.NewRefreshService(store.NewRefreshStore(db), service.NewSettingsService(store.NewSettingsStore(db), &NopAuditer{}))
}

// testRefreshHandler creates a RefreshHandler wired through Store→Service for testing.
func testRefreshHandler(db *sqlx.DB) *RefreshHandler {
	return NewRefreshHandler(testRefreshService(db))
}
