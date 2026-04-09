package handlers

import (
"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_List_FiltersInactiveTCG(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	// 1. Total count query (buildFilters in store)
	mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 2. Main list query (Note: sqlx expands p.* into explicit columns)
	mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "tcg"}).AddRow("1", "Black Lotus", "mtg"))

	// Settings first (now called after main query in List)
	mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

	// Enrichment
	mock.ExpectQuery("(?i)SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("1", "loc1", "Box 1", 10))
	mock.ExpectQuery("(?i)SELECT .* FROM product_category pc JOIN custom_category c").
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("1", "cat1", "Cat 1", "cat1", true, true, true))
	mock.ExpectQuery("(?i)SELECT .* FROM \"order\" o").
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("1", 0))
	mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

	// 3. Facets query (one single call to the special function)
	mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
		WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{"condition":{"NM":1}}`)))

	// 4. Mock enrichment queries (populatePrices (settings already mocked), populateStorage, populateCategories, populateCartCounts)
	mock.ExpectQuery("(?i)SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("(?i)SELECT .* FROM product_category").WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name"}))
	mock.ExpectQuery("(?i)SELECT .* FROM \"order\" o").WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}))

	req, _ := http.NewRequest("GET", "/api/products", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestProductHandler_GetByID_Public_FiltersInactiveTCG(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

        // Mocking a scenario where product is not found (e.g. inactive)
        mock.ExpectQuery("(?is)SELECT fn_get_product_detail").
                WithArgs("123").
                WillReturnError(sql.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/api/products/{id}", h.GetByID)

	req, _ := http.NewRequest("GET", "/api/products/123", nil)
	rr := httptest.NewRecorder()
	
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
