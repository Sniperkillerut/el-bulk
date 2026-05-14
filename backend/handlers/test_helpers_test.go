package handlers

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/DATA-DOG/go-sqlmock"
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

// testJobService creates a JobService for testing.
func testJobService(db *sqlx.DB) *service.JobService {
	return service.NewJobService(store.NewJobStore(db))
}

// testWorkerPool creates a WorkerPool for testing.
func testWorkerPool(db *sqlx.DB) *service.WorkerPool {
	return service.NewWorkerPool(testJobService(db), 0) // 0 workers for synchronous-like testing or just submission checks
}

// testRefreshHandler creates a RefreshHandler wired through Store→Service for testing.
func testRefreshHandler(db *sqlx.DB) *RefreshHandler {
	return NewRefreshHandler(testRefreshService(db), testWorkerPool(db), &NopAuditer{})
}
// mockProductSideData adds mocks for side-data population (parallel in SelectEnriched).
func mockProductSideData(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("(?i)SELECT ps.product_id, s.id as stored_in_id").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))
	mock.ExpectQuery("(?i)SELECT pc.product_id, c.id, c.name").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable", "bg_color", "text_color", "icon"}))
	mock.ExpectQuery("(?is)SELECT .* FROM \"order\" o").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}))
	mock.ExpectQuery("(?is)SELECT .* FROM deck_card").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "set_code", "collector_number", "quantity", "type_line", "image_url", "foil_treatment", "card_treatment", "rarity", "art_variation", "scryfall_id", "frame_effects"}))
}

// mockProductListEnrichment adds all mocks for a full Product List enrichment.
func mockProductListEnrichment(mock sqlmock.Sqlmock) {
	mockProductSideData(mock)
	mockProductHotStatus(mock)
}

// mockProductHotStatus adds mocks for hot/new status identification.
func mockProductHotStatus(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("(?is)SELECT product_id FROM order_item").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"product_id"}))
}
