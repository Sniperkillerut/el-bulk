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
			settingsStore := store.NewSettingsStore(sqlxDB)
			settingsService := service.NewSettingsService(settingsStore, nil)
			ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
			h := &ProductHandler{Service: ps, DB: sqlxDB}
			settingsService.ResetCache()

			// 1. Settings (called first in ProductService.List)
			mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

			// 2. Total count
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			// 3. Main list
			mock.ExpectQuery("SELECT .* FROM view_product_enriched p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

			// 4. Enrichment: populateCartCounts
			mock.ExpectQuery("(?i)SELECT .* FROM \"order\" o").
				WithArgs("1").
				WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("1", 0))
			mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

			// 5. Facets (1 call)
			mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
				WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

			req, _ := http.NewRequest("GET", "/api/products"+url, nil)
			rr := httptest.NewRecorder()
			h.List(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}

	testFilters("Multiple Filters", "?tcg=mtg&category=singles&foil=foil,non_foil&rarity=rare&condition=NM,LP&language=en&search=lotus")
	testFilters("Color Filter", "?color=W,U,B")
	testFilters("Storage Filter", "?storage_id=loc1")
	testFilters("Collection Filter", "?collection=my-deck")
	testFilters("Treatment Filter", "?treatment=borderless,extendedart")
}

func TestProductHandler_AdminVsPublic(t *testing.T) {
	t.Run("Public Joins TCGs", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
		h := &ProductHandler{Service: ps, DB: sqlxDB}
		settingsService.ResetCache()

		// 1. Settings
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p LEFT JOIN tcg t").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p LEFT JOIN tcg t").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

		// Enrichment
		mock.ExpectQuery("(?i)SELECT .* FROM \"order\" o").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("1", 0))
		mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

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
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
		h := &ProductHandler{Service: ps, DB: sqlxDB}
		settingsService.ResetCache()

		// 1. Settings
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

		// Enrichment
		mock.ExpectQuery("(?i)SELECT .* FROM \"order\" o").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("1", 0))
		mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

		// Facets
		mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
			WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

		req, _ := http.NewRequest("GET", "/api/admin/products", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
