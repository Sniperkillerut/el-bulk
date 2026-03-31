package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
			h := &ProductHandler{DB: sqlxDB}
			ResetSettingsCache()

			// 1. Total count
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			// 2. Main list
			mock.ExpectQuery("SELECT .* FROM view_product_enriched p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
			
			// 3. Enrichment: populatePrices calls loadSettings
			mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))
			
			// 4. Enrichment: populateStorage
			mock.ExpectQuery("SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("1", "loc1", "Box 1", 10))
			
			// 5. Enrichment: populateCategories
			mock.ExpectQuery("(?i)SELECT pc\\.product_id, c\\.id, c\\.name, c\\.slug, c\\.show_badge, c\\.is_active, c\\.searchable FROM product_category pc JOIN custom_category c ON pc.category_id = c.id WHERE pc\\.product_id IN \\(\\$1\\) AND c\\.show_badge = true ORDER BY c\\.name").
				WithArgs("1").
				WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("1", "cat1", "Category 1", "cat1", true, true, true))

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
		h := &ProductHandler{DB: sqlxDB}
		ResetSettingsCache()

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p LEFT JOIN tcg t").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p LEFT JOIN tcg t").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
		
		// Enrichment
		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))
		mock.ExpectQuery("SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("1", "loc1", "Box 1", 10))
		mock.ExpectQuery("(?i)SELECT pc\\.product_id, c\\.id, c\\.name, c\\.slug, c\\.show_badge, c\\.is_active, c\\.searchable FROM product_category pc JOIN custom_category c ON pc.category_id = c.id WHERE pc\\.product_id IN \\(\\$1\\) AND c\\.show_badge = true ORDER BY c\\.name").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("1", "cat1", "Cat 1", "cat1", true, true, true))

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
		h := &ProductHandler{DB: sqlxDB}
		ResetSettingsCache()

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
		
		// Enrichment
		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))
		mock.ExpectQuery("SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("1", "loc1", "Box 1", 10))
		mock.ExpectQuery("(?i)SELECT pc\\.product_id, c\\.id, c\\.name, c\\.slug, c\\.show_badge, c\\.is_active, c\\.searchable FROM product_category pc JOIN custom_category c ON pc.category_id = c.id WHERE pc\\.product_id IN \\(\\$1\\) ORDER BY c\\.name").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("1", "cat1", "Cat 1", "cat1", true, true, true))

		// Facets
		mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
			WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

		req, _ := http.NewRequest("GET", "/api/admin/products", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
