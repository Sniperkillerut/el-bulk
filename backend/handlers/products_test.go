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
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Basic List", func(t *testing.T) {
		ResetSettingsCache()
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "name", "tcg", "category", "price_source", "price_reference", "stock", "created_at", "updated_at"}).
			AddRow("p1", "Product 1", "mtg", "singles", "tcgplayer", 1.0, 10, now, now)

		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		
		rows = sqlmock.NewRows([]string{"id", "name", "tcg", "category", "price_source", "price_reference", "stock", "created_at", "updated_at", "stored_in_json", "categories_json"}).
			AddRow("p1", "Product 1", "mtg", "singles", "tcgplayer", 1.0, 10, now, now, []byte("[]"), []byte("[]"))
		mock.ExpectQuery("(?i)SELECT .* FROM view_product_enriched p").WillReturnRows(rows)
		
		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		mock.ExpectQuery("(?i)SELECT condition, COUNT.*").WillReturnRows(sqlmock.NewRows([]string{"condition", "count"}).AddRow("NM", 1))
		mock.ExpectQuery("(?i)SELECT foil_treatment, COUNT.*").WillReturnRows(sqlmock.NewRows([]string{"foil_treatment", "count"}).AddRow("non_foil", 1))
		mock.ExpectQuery("(?i)SELECT card_treatment, COUNT.*").WillReturnRows(sqlmock.NewRows([]string{"card_treatment", "count"}).AddRow("normal", 1))
		mock.ExpectQuery("(?i)SELECT rarity, COUNT.*").WillReturnRows(sqlmock.NewRows([]string{"rarity", "count"}).AddRow("rare", 1))
		mock.ExpectQuery("(?i)SELECT language, COUNT.*").WillReturnRows(sqlmock.NewRows([]string{"language", "count"}).AddRow("en", 1))
		mock.ExpectQuery("(?i)SELECT.*FILTER.*").WillReturnRows(sqlmock.NewRows([]string{"w", "u", "b", "r", "g", "c"}).AddRow(1, 0, 0, 0, 0, 0))
		mock.ExpectQuery("(?i)SELECT c.slug, COUNT.* FROM product_category pc JOIN custom_category c").WillReturnRows(sqlmock.NewRows([]string{"slug", "count"}).AddRow("cat1", 1))

		req, _ := http.NewRequest("GET", "/api/products?page=1&pageSize=20", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var res models.ProductListResponse
		err = json.NewDecoder(rr.Body).Decode(&res)
		assert.NoError(t, err)
		assert.Len(t, res.Products, 1)
	})
}

func TestProductHandler_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Found", func(t *testing.T) {
		ResetSettingsCache()

		mock.ExpectQuery("(?i)SELECT COALESCE\\(t\\.is_active, true\\) FROM product p").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"active"}).AddRow(true))

		mock.ExpectQuery("(?i)SELECT fn_get_product_detail").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow([]byte(`{"id":"p1","name":"Product 1","tcg":"mtg","category":"singles"}`)))

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		r := chi.NewRouter()
		r.Get("/api/products/{id}", h.GetByID)
		req, _ := http.NewRequest("GET", "/api/products/p1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		ResetSettingsCache()
		mock.ExpectQuery("SELECT .* FROM product WHERE id = \\$1").
			WithArgs("p99").
			WillReturnError(http.ErrNoLocation)

		r := chi.NewRouter()
		r.Get("/api/products/{id}", h.GetByID)
		req, _ := http.NewRequest("GET", "/api/products/p99", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProductHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		ResetSettingsCache()
		input := models.ProductInput{Name: "New Product", TCG: "mtg", Category: "singles", Stock: 10}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("INSERT INTO product .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "tcg", "category"}).AddRow("p-new", "New Product", "mtg", "singles"))
		
		mock.ExpectExec("DELETE FROM product_category WHERE product_id = \\$1").WithArgs("p-new").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("DELETE FROM product_storage WHERE product_id = \\$1").WithArgs("p-new").WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		mock.ExpectQuery("(?i)SELECT ps\\.product_id, s\\.id as stored_in_id.*").
			WithArgs("p-new").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("p-new", "loc1", "Box 1", 10))

		mock.ExpectQuery("(?i)SELECT pc\\.product_id, c\\.id, c\\.name.*").
			WithArgs("p-new").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("p-new", "cat1", "Cat 1", "cat1", true, true, true))

		req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestProductHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		ResetSettingsCache()
		input := models.ProductInput{Name: "Updated Name"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("(?i)UPDATE product").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "tcg", "category"}).AddRow("p1", "Updated Name", "mtg", "singles"))

		mock.ExpectExec("DELETE FROM product_category WHERE product_id = \\$1").WithArgs("p1").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("DELETE FROM product_storage WHERE product_id = \\$1").WithArgs("p1").WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectQuery("(?i)SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop", "4000"))

		mock.ExpectQuery("(?i)SELECT ps\\.product_id, s\\.id as stored_in_id.*").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).AddRow("p1", "loc1", "Box 1", 10))

		mock.ExpectQuery("(?i)SELECT pc\\.product_id, c\\.id, c\\.name.*").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable"}).AddRow("p1", "cat1", "Cat 1", "cat1", true, true, true))

		r := chi.NewRouter()
		r.Put("/api/admin/products/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/products/p1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestProductHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM product WHERE id = \\$1").
			WithArgs("p1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := chi.NewRouter()
		r.Delete("/api/admin/products/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/products/p1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestProductHandler_BulkCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		ResetSettingsCache()
		inputs := []models.ProductInput{
			{Name: "Bulk 1", TCG: "mtg", Category: "singles"},
			{Name: "Bulk 2", TCG: "mtg", Category: "singles"},
		}
		body, _ := json.Marshal(inputs)

		mock.ExpectQuery("SELECT product_id FROM fn_bulk_upsert_product").
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"product_id"}).AddRow("b1").AddRow("b2"))

		req, _ := http.NewRequest("POST", "/api/admin/products/bulk", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BulkCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestProductHandler_GetStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		ResetSettingsCache()
		rows := sqlmock.NewRows([]string{"storage_id", "name", "quantity"}).
			AddRow("loc1", "Box 1", 10).
			AddRow("loc2", "Box 2", 0)

		mock.ExpectQuery("(?i)SELECT .* FROM storage_location").
			WithArgs("p1").
			WillReturnRows(rows)

		r := chi.NewRouter()
		r.Get("/api/admin/products/{id}/storage", h.GetStorage)
		req, _ := http.NewRequest("GET", "/api/admin/products/p1/storage", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestProductHandler_UpdateStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		ResetSettingsCache()
		updates := []models.ProductStorage{
			{StorageID: "loc1", Quantity: 5},
		}
		body, _ := json.Marshal(updates)

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM product_storage WHERE product_id = \\$1").WithArgs("p1").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO product_storage").WithArgs("p1", "loc1", 5).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// GetStorage follow-up
		mock.ExpectQuery("(?i)SELECT .* FROM storage_location").WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"storage_id", "name", "quantity"}).AddRow("loc1", "Box 1", 5))

		r := chi.NewRouter()
		r.Put("/api/admin/products/{id}/storage", h.UpdateStorage)
		req, _ := http.NewRequest("PUT", "/api/admin/products/p1/storage", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
