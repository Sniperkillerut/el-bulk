package handlers

import (
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
)

// testBountyHandler creates a BountyHandler wired through Store→Service for testing.
func testBountyHandler(db *sqlx.DB) *BountyHandler {
	return NewBountyHandler(service.NewBountyService(store.NewBountyStore(db)))
}

// testAdminHandler creates an AdminHandler wired through Store→Service for testing.
func testAdminHandler(db *sqlx.DB) *AdminHandler {
	return NewAdminHandler(service.NewAdminService(store.NewAdminStore(db)))
}

// testNewsletterHandler creates a NewsletterHandler wired through Store→Service for testing.
func testNewsletterHandler(db *sqlx.DB) *NewsletterHandler {
	return NewNewsletterHandler(service.NewNewsletterService(store.NewNewsletterStore(db)))
}

// testStorageHandler creates a StorageHandler wired through Store→Service for testing.
func testStorageHandler(db *sqlx.DB) *StorageHandler {
	return NewStorageHandler(service.NewStorageLocationService(store.NewStorageLocationStore(db)))
}

// testHealthHandler creates a HealthHandler wired through Store→Service for testing.
func testHealthHandler(db *sqlx.DB) *HealthHandler {
	return NewHealthHandler(service.NewHealthService(store.NewHealthStore(db)))
}

// testRefreshService creates a RefreshService for testing.
func testRefreshService(db *sqlx.DB) *service.RefreshService {
	return service.NewRefreshService(store.NewRefreshStore(db))
}

// testRefreshHandler creates a RefreshHandler wired through Store→Service for testing.
func testRefreshHandler(db *sqlx.DB) *RefreshHandler {
	return NewRefreshHandler(testRefreshService(db))
}
