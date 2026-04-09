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
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService)
			h := &ProductHandler{Service: ps, DB: sqlxDB}
			settingsService.ResetCache()

			// 1. Total count
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			// 2. Main list
			mock.ExpectQuery("SELECT .* FROM view_product_enriched p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

			// Settings (now called after main query in List)
			mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))
			
			// Note: populatePrices NO LONGER calls loadSettings because we passed it from handler
			
			// 4. Enrichment: populateStorage
			mock.ExpectQuery("(?i)SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("1", "loc1", "Box 1", 10))
			
			// 5. Enrichment: populateCategories
			mock.ExpectQuery("(?i)SELECT .* FROM product_category pc JOIN custom_category c").
				WithArgs("1").
				WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("1", "cat1", "Category 1", "cat1", true, true, true))

			// 5b. Enrichment: populateCartCounts
			mock.ExpectQuery("(?i)SELECT .* FROM \"order\" o").
				WithArgs("1").
				WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("1", 0))
			mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

			// 6. Facets (1 call)
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
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService)
		h := &ProductHandler{Service: ps, DB: sqlxDB}
		settingsService.ResetCache()

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p LEFT JOIN tcg t").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p LEFT JOIN tcg t").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
		
		// Settings (now called after main query)
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))
		
		// Enrichment (populatePrices)
		mock.ExpectQuery("(?i)SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("1", "loc1", "Box 1", 10))
		mock.ExpectQuery("(?i)SELECT .* FROM product_category pc JOIN custom_category c").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("1", "cat1", "Cat 1", "cat1", true, true, true))
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
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService)
		h := &ProductHandler{Service: ps, DB: sqlxDB}
		settingsService.ResetCache()

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
		
		// Settings (now called after main query)
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))
		
		// Enrichment (populatePrices)
		mock.ExpectQuery("(?i)SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("1", "loc1", "Box 1", 10))
		mock.ExpectQuery("(?i)SELECT .* FROM product_category pc JOIN custom_category c").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("1", "cat1", "Cat 1", "cat1", true, true, true))
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
