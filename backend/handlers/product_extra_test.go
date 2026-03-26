package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func TestProductHandler_BulkCreate_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Tx Begin Error", func(t *testing.T) {
		inputs := []models.ProductInput{{Name: "P1", TCG: "mtg", Category: "singles"}}
		body, _ := json.Marshal(inputs)
		mock.ExpectBegin().WillReturnError(fmt.Errorf("tx error"))

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
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Tx Begin Error", func(t *testing.T) {
		updates := []models.ProductStorage{{StoredInID: "loc1", Quantity: 5}}
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
		updates := []models.ProductStorage{{StoredInID: "loc1", Quantity: 5}}
		body, _ := json.Marshal(updates)
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM product_stored_in").WillReturnError(fmt.Errorf("db error"))
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
	h := &ProductHandler{DB: sqlxDB}

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM products").WillReturnError(fmt.Errorf("db error"))
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
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "is_active", "created_at"}).
			AddRow("mtg", "Magic", true, time.Now())
		mock.ExpectQuery("SELECT id, name, is_active, created_at FROM tcgs WHERE is_active = true").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/tcgs?active_only=true", nil)
		rr := httptest.NewRecorder()
		h.ListTCGs(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("List Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT").WillReturnError(fmt.Errorf("db error"))
		req, _ := http.NewRequest("GET", "/api/admin/orders", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestProductHandler_BulkCreate_Extra(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	t.Run("Skip Invalid", func(t *testing.T) {
		inputs := []models.ProductInput{{Name: ""}, {Name: "P1", TCG: "mtg", Category: "singles"}}
		body, _ := json.Marshal(inputs)
		
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO products").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p1"))
		mock.ExpectCommit()

		req, _ := http.NewRequest("POST", "/api/admin/products/bulk", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BulkCreate(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Insert Error", func(t *testing.T) {
		inputs := []models.ProductInput{{Name: "P1", TCG: "mtg", Category: "singles"}}
		body, _ := json.Marshal(inputs)
		
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO products").WillReturnError(fmt.Errorf("db error"))
		mock.ExpectRollback()

		req, _ := http.NewRequest("POST", "/api/admin/products/bulk", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BulkCreate(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
