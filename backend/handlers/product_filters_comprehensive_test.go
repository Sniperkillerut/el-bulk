package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_BuildFilters_Comprehensive(t *testing.T) {
	// Helper to mock a List call with specific URL params
	testFilters := func(name string, url string) {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "postgres")
			mock.MatchExpectationsInOrder(false)
			settingsStore := store.NewSettingsStore(sqlxDB)
			settingsService := service.NewSettingsService(settingsStore, nil)
			ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
			h := &ProductHandler{Service: ps, DB: sqlxDB}
			settingsService.ResetCache()

			// 1. Settings (called first in ProductService.List)
			mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

			// 2. Total count
			mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM .* p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			productID := "550e8400-e29b-41d4-a716-446655440001"
			// 3. Main list
			mock.ExpectQuery("(?i)SELECT .* FROM .* p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))

			// 4. Enrichment
			mockProductListEnrichment(mock)

			// 5. Facets (1 call)
			mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
				WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

			req, _ := http.NewRequest("GET", "/api/products"+url, nil)
			rr := httptest.NewRecorder()
			h.List(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}

	testFilters("Multiple Filters", "?tcg=550e8400-e29b-41d4-a716-446655440003&category=singles&foil=foil,non_foil&rarity=rare&condition=NM,LP&language=en&search=lotus")
	testFilters("Color Filter", "?color=W,U,B")
	testFilters("Storage Filter", "?storage_id=550e8400-e29b-41d4-a716-446655440005")
	testFilters("Collection Filter", "?collection=my-deck")
	testFilters("Treatment Filter", "?treatment=borderless,extendedart")
}

func TestProductHandler_AdminVsPublic(t *testing.T) {
	t.Run("Public Joins TCGs", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		mock.MatchExpectationsInOrder(false)
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
		h := &ProductHandler{Service: ps, DB: sqlxDB}
		settingsService.ResetCache()

		// 1. Settings
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

		productID := "550e8400-e29b-41d4-a716-446655440001"
		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM .* p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM .* p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))

		// Enrichment
		mockProductListEnrichment(mock)

		// Facets
		mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
			WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

		req, _ := http.NewRequest("GET", "/api/products", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Admin No Joins", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		mock.MatchExpectationsInOrder(false)
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
		h := &ProductHandler{Service: ps, DB: sqlxDB}
		settingsService.ResetCache()

		// 1. Settings
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

		productID := "550e8400-e29b-41d4-a716-446655440001"
		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM .* p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM .* p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))

		// Enrichment
		mockProductListEnrichment(mock)

		// Facets
		mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
			WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

		req, _ := http.NewRequest("GET", "/api/admin/products", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
