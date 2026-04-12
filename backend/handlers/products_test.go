package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	now := time.Now()

	t.Run("Basic List", func(t *testing.T) {
		settingsService.ResetCache()

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		rows := sqlmock.NewRows([]string{"id", "name", "tcg", "category", "price_source", "price_reference", "stock", "created_at", "updated_at", "stored_in_json", "categories_json"}).
			AddRow("p1", "Product 1", "mtg", "singles", "tcgplayer", 1.0, 10, now, now, []byte("[]"), []byte("[]"))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p").WillReturnRows(rows)

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		// Enrichment
		mock.ExpectQuery("(?is)SELECT .* FROM product_storage ps").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))
		mock.ExpectQuery("(?is)SELECT .* FROM product_category pc").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}))
		mock.ExpectQuery("(?is)SELECT .* FROM \"order\" o").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("p1", 0))
		mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

		// Facets
		mock.ExpectQuery("(?is)SELECT fn_get_product_facets").WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

		req, _ := http.NewRequest("GET", "/api/products", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Found", func(t *testing.T) {
		settingsService.ResetCache()
		mock.ExpectQuery("(?is)SELECT fn_get_product_detail").WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow([]byte(`{"id":"p1","name":"P1"}`)))

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		// Enrichment (public GetByID)
		mock.ExpectQuery("(?is)SELECT .* FROM \"order\" o").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("p1", 0))
		mock.ExpectQuery("(?is)SELECT .* FROM product_category pc").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("p1", "cat1", "Cat 1", "cat1", true, true, true))

		r := chi.NewRouter()
		r.Get("/api/products/{id}", h.GetByID)
		req, _ := http.NewRequest("GET", "/api/products/p1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		settingsService.ResetCache()
		input := models.ProductInput{Name: "New Product", TCG: "mtg", Category: "singles"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("(?i)INSERT INTO product").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p-new"))
		
		mock.ExpectExec("(?i)DELETE FROM product_category").WillReturnResult(sqlmock.NewResult(0, 0)) // Categories DELETE
		mock.ExpectExec("(?i)DELETE FROM product_storage").WillReturnResult(sqlmock.NewResult(0, 0))  // Storage DELETE

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		// Enrichment
		mock.ExpectQuery("(?is)SELECT .* FROM product_storage ps").WithArgs("p-new").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))
		mock.ExpectQuery("(?is)SELECT .* FROM product_category pc").WithArgs("p-new").WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}))
		mock.ExpectQuery("(?is)SELECT .* FROM \"order\" o").WithArgs("p-new").WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("p-new", 0))
		mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WithArgs("p-new").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

		req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		settingsService.ResetCache()
		input := models.ProductInput{Name: "Updated Name"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("(?i)UPDATE product").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p1"))
		
		mock.ExpectExec("(?i)DELETE FROM product_category").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("(?i)DELETE FROM product_storage").WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		// Enrichment
		mock.ExpectQuery("(?is)SELECT .* FROM product_storage ps").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))
		mock.ExpectQuery("(?is)SELECT .* FROM product_category pc").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}))
		mock.ExpectQuery("(?is)SELECT .* FROM \"order\" o").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("p1", 0))
		mock.ExpectQuery("(?is)SELECT product_id FROM order_item").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

		r := chi.NewRouter()
		r.Put("/api/admin/products/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/products/p1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_GetStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("(?is)SELECT .* FROM storage_location").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"storage_id", "name", "quantity"}).AddRow("loc1", "Box 1", 10))
		r := chi.NewRouter()
		r.Get("/api/admin/products/{id}/storage", h.GetStorage)
		req, _ := http.NewRequest("GET", "/api/admin/products/p1/storage", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_UpdateStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, nil)
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		updates := []models.ProductStorage{{StorageID: "loc1", Quantity: 5}}
		body, _ := json.Marshal(updates)

		mock.ExpectBegin()
		mock.ExpectExec("(?i)DELETE FROM product_storage").WithArgs("p1").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("(?i)INSERT INTO product_storage").WithArgs("p1", "loc1", 5).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		mock.ExpectQuery("(?is)SELECT .* FROM storage_location").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"storage_id", "name", "quantity"}).AddRow("loc1", "Box 1", 5))

		r := chi.NewRouter()
		r.Put("/api/admin/products/{id}/storage", h.UpdateStorage)
		req, _ := http.NewRequest("PUT", "/api/admin/products/p1/storage", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
