package handlers

import (
	"testing"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestConstructors(t *testing.T) {
	db := &sqlx.DB{}
	settingsStore := store.NewSettingsStore(db)
	settingsService := service.NewSettingsService(settingsStore)

	orderStore := store.NewOrderStore(db)
	customerStore := store.NewCustomerStore(db)
	orderService := service.NewOrderService(orderStore, store.NewProductStore(db), customerStore, settingsService, nil)

	assert.NotNil(t, testAdminHandler(db))
	assert.NotNil(t, NewCategoriesHandler(service.NewCategoryService(store.NewCategoryStore(db))))
	assert.NotNil(t, NewOrderHandler(orderService))

	ps := service.NewProductService(store.NewProductStore(db), store.NewTCGStore(db), settingsService, nil)
	assert.NotNil(t, NewProductHandler(ps, db))

	assert.NotNil(t, testRefreshHandler(db))
	assert.NotNil(t, NewSettingsHandler(settingsService))
	assert.NotNil(t, testStorageHandler(db))
	assert.NotNil(t, NewTCGHandler(service.NewTCGService(store.NewTCGStore(db))))
	assert.NotNil(t, NewThemeHandler(service.NewThemeService(store.NewThemeStore(db))))
	assert.NotNil(t, NewNoticeHandler(service.NewNoticeService(store.NewNoticeStore(db))))
}
