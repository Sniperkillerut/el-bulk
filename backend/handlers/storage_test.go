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
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestStorageHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &StorageHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM storage_location").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "item_count"}).AddRow("1", "Shelf", 10))
		req, _ := http.NewRequest("GET", "/api/admin/storage", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM storage_location").WillReturnError(fmt.Errorf("db error"))
		req, _ := http.NewRequest("GET", "/api/admin/storage", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestStorageHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &StorageHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.StoredIn{Name: "New Box"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("INSERT INTO stored_in").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("5"))

		req, _ := http.NewRequest("POST", "/api/admin/storage", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		input := models.StoredIn{Name: "New Box"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("INSERT INTO stored_in").WillReturnError(fmt.Errorf("conflict"))

		req, _ := http.NewRequest("POST", "/api/admin/storage", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestStorageHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &StorageHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.StoredIn{Name: "Updated Name"}
		body, _ := json.Marshal(input)
		mock.ExpectExec("UPDATE stored_in").WillReturnResult(sqlmock.NewResult(0, 1))

		r := chi.NewRouter()
		r.Put("/api/admin/storage/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/storage/1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		input := models.StoredIn{Name: "Updated Name"}
		body, _ := json.Marshal(input)
		mock.ExpectExec("UPDATE stored_in").WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Put("/api/admin/storage/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/storage/1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestStorageHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &StorageHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM storage_location").WillReturnResult(sqlmock.NewResult(0, 1))
		r := chi.NewRouter()
		r.Delete("/api/admin/storage/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/storage/1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM storage_location").WillReturnError(fmt.Errorf("db error"))
		r := chi.NewRouter()
		r.Delete("/api/admin/storage/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/storage/1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
