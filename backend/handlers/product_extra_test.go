package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_BulkCreate_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("SP Error", func(t *testing.T) {
		inputs := []models.ProductInput{{Name: "P1", TCG: "mtg", Category: "singles"}}
		reqBody := struct {
			Products []models.ProductInput `json:"products"`
		}{Products: inputs}
		body, _ := json.Marshal(reqBody)
		mock.ExpectQuery("SELECT upserted_id FROM fn_bulk_upsert_product").WillReturnError(fmt.Errorf("db error"))

		req, _ := http.NewRequest("POST", "/api/admin/products/bulk", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BulkCreate(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestProductHandler_UpdateStorage_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Tx Begin Error", func(t *testing.T) {
		updates := []models.ProductStorage{{StorageID: "loc1", Quantity: 5}}
		body, _ := json.Marshal(updates)
		mock.ExpectBegin().WillReturnError(fmt.Errorf("tx error"))

		r := chi.NewRouter()
		r.Put("/api/admin/products/{id}/storage", h.UpdateStorage)
		req, _ := http.NewRequest("PUT", "/api/admin/products/p1/storage", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Delete Error", func(t *testing.T) {
		updates := []models.ProductStorage{{StorageID: "loc1", Quantity: 5}}
		body, _ := json.Marshal(updates)
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM product_storage").WillReturnError(fmt.Errorf("db error"))
		mock.ExpectRollback()

		r := chi.NewRouter()
		r.Put("/api/admin/products/{id}/storage", h.UpdateStorage)
		req, _ := http.NewRequest("PUT", "/api/admin/products/p1/storage", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestProductHandler_Delete_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM product").WillReturnError(fmt.Errorf("db error"))
		r := chi.NewRouter()
		r.Delete("/api/admin/products/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/products/p1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestProductHandler_ListTCGs(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "is_active", "item_count"}).
			AddRow("mtg", "Magic", true, 10)
		mock.ExpectQuery("(?is)SELECT t.*, COUNT\\(p.id\\) as item_count FROM tcg t LEFT JOIN product p ON t.id = p.tcg WHERE t.is_active = true GROUP BY t.id ORDER BY t.name").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/tcgs?active_only=true", nil)
		rr := httptest.NewRecorder()
		h.ListTCGs(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("List Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM tcg").WillReturnError(fmt.Errorf("db error"))
		req, _ := http.NewRequest("GET", "/api/tcgs", nil)
		rr := httptest.NewRecorder()
		h.ListTCGs(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestProductHandler_BulkCreate_Extra(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	ps := service.NewProductService(store.NewProductStore(sqlxDB), store.NewTCGStore(sqlxDB), settingsService, &NopAuditer{})
	h := &ProductHandler{Service: ps, DB: sqlxDB}

	t.Run("Skip Invalid", func(t *testing.T) {
		inputs := []models.ProductInput{{Name: ""}, {Name: "P1", TCG: "mtg", Category: "singles"}}
		reqBody := struct {
			Products []models.ProductInput `json:"products"`
		}{Products: inputs}
		body, _ := json.Marshal(reqBody)

		mock.ExpectQuery("SELECT upserted_id FROM fn_bulk_upsert_product").
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"upserted_id"}).AddRow("p-new"))

		req, _ := http.NewRequest("POST", "/api/admin/products/bulk", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BulkCreate(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Insert Error", func(t *testing.T) {
		inputs := []models.ProductInput{{Name: "P1", TCG: "mtg", Category: "singles"}}
		reqBody := struct {
			Products []models.ProductInput `json:"products"`
		}{Products: inputs}
		body, _ := json.Marshal(reqBody)

		mock.ExpectQuery("SELECT upserted_id FROM fn_bulk_upsert_product").WillReturnError(fmt.Errorf("db error"))

		req, _ := http.NewRequest("POST", "/api/admin/products/bulk", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BulkCreate(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
