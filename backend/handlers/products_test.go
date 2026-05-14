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
	t.Run("Basic List", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		mock.MatchExpectationsInOrder(false)
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
		h := &ProductHandler{Service: ps, DB: sqlxDB}

		now := time.Now()
		settingsService.ResetCache()

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))
		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		rows := sqlmock.NewRows([]string{"id", "name", "tcg", "category", "price_source", "price_reference", "stock", "created_at", "updated_at", "stored_in_json", "categories_json"}).
			AddRow("550e8400-e29b-41d4-a716-446655440001", "Product 1", "550e8400-e29b-41d4-a716-446655440003", "singles", "tcgplayer", 1.0, 10, now, now, []byte("[]"), []byte("[]"))
		mock.ExpectQuery("(?i)SELECT .* FROM .* p").WillReturnRows(rows)
		mockProductListEnrichment(mock)
		mock.ExpectQuery("(?is)SELECT fn_get_product_facets").WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_facets"}).AddRow([]byte(`{}`)))

		req, _ := http.NewRequest("GET", "/api/products", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_GetByID(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		mock.MatchExpectationsInOrder(false)
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
		h := &ProductHandler{Service: ps, DB: sqlxDB}

		productID := "550e8400-e29b-41d4-a716-446655440001"
		settingsService.ResetCache()
		mock.ExpectQuery("(?is)SELECT fn_get_product_detail").WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow([]byte(`{"id":"` + productID + `","name":"P1"}`)))
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))
		mockProductHotStatus(mock)

		r := chi.NewRouter()
		r.Get("/api/products/{id}", h.GetByID)
		req, _ := http.NewRequest("GET", "/api/products/"+productID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		mock.MatchExpectationsInOrder(false)
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
		h := &ProductHandler{Service: ps, DB: sqlxDB}

		input := models.ProductInput{
			Name:     "New Product",
			TCG:      "mtg",
			Category: "singles",
		}
		body, _ := json.Marshal(input)

		// 1. BulkUpsert
		mock.ExpectQuery("(?i)SELECT upserted_id FROM fn_bulk_upsert_product").WillReturnRows(sqlmock.NewRows([]string{"upserted_id"}).AddRow("550e8400-e29b-41d4-a716-446655440001"))
		
		// 2. GetByID (follow-up)
		mock.ExpectQuery("(?is)SELECT fn_get_product_detail").WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow([]byte(`{"id":"550e8400-e29b-41d4-a716-446655440001","name":"New Product"}`)))
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))
		
		// 3. Enrichment
		mockProductHotStatus(mock)

		req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		mock.MatchExpectationsInOrder(false)
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
		h := &ProductHandler{Service: ps, DB: sqlxDB}

		productID := "550e8400-e29b-41d4-a716-446655440001"
		input := map[string]interface{}{"name": "Updated"}
		body, _ := json.Marshal(input)

		// GetEnrichedByID (before update)
		mock.ExpectQuery("(?is)SELECT fn_get_product_detail").WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow([]byte(`{"id":"` + productID + `","name":"Old"}`)))
		
		// UpdateProduct
		mock.ExpectQuery("(?i)UPDATE product SET").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))
		
		// SaveCategories (cleanup)
		mock.ExpectExec("(?i)DELETE FROM product_category").WithArgs(productID).WillReturnResult(sqlmock.NewResult(0, 1))
		
		// SaveStorage (cleanup)
		mock.ExpectExec("(?i)DELETE FROM product_storage").WithArgs(productID).WillReturnResult(sqlmock.NewResult(0, 1))

		// GetSettings
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4000"))

		// Enrichment
		mockProductHotStatus(mock)

		r := chi.NewRouter()
		r.Put("/api/admin/products/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/products/"+productID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductHandler_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		mock.MatchExpectationsInOrder(false)
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
		h := &ProductHandler{Service: ps, DB: sqlxDB}

		productID := "550e8400-e29b-41d4-a716-446655440001"
		
		// GetEnrichedByID (before delete)
		mock.ExpectQuery("(?is)SELECT fn_get_product_detail").WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow([]byte(`{"id":"` + productID + `"}`)))
		
		mock.ExpectExec("(?i)DELETE FROM product").WithArgs(productID).WillReturnResult(sqlmock.NewResult(0, 1))

		r := chi.NewRouter()
		r.Delete("/api/admin/products/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/products/"+productID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
