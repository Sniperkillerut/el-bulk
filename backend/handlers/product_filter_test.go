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
	settingsService := service.NewSettingsService(settingsStore, nil)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	// 1. Settings (called first in ProductService.List)
	mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

	// 2. Total count query (buildFilters in store)
	mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 3. Main list query
	mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "tcg"}).AddRow("550e8400-e29b-41d4-a716-446655440001", "Black Lotus", "550e8400-e29b-41d4-a716-446655440003"))

	// 4. Enrichment (inside List)
	mock.ExpectQuery("(?i)SELECT .* FROM \"order\" o").
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("550e8400-e29b-41d4-a716-446655440001", 0))
	mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

	// 5. Facets query
	mock.ExpectQuery("(?i)SELECT fn_get_product_facets").
		WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{"condition":{"NM":1}}`)))

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
	settingsService := service.NewSettingsService(settingsStore, nil)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	// Mocking a scenario where product is not found (e.g. inactive)
	mock.ExpectQuery("(?is)SELECT fn_get_product_detail").
		WithArgs("550e8400-e29b-41d4-a716-446655440004").
		WillReturnError(sql.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/api/products/{id}", h.GetByID)

	req, _ := http.NewRequest("GET", "/api/products/550e8400-e29b-41d4-a716-446655440004", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
