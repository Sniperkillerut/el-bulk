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

func TestTCGHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &TCGHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "is_active", "created_at"}).
			AddRow("mtg", "Magic", true, time.Now())
		mock.ExpectQuery("SELECT \\* FROM tcg ORDER BY name").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/admin/tcgs", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM tcg").WillReturnError(fmt.Errorf("db error"))
		req, _ := http.NewRequest("GET", "/api/admin/tcgs", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTCGHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &TCGHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.TCGInput{ID: "one", Name: "One Piece"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("INSERT INTO tcg").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("one", "One Piece"))

		req, _ := http.NewRequest("POST", "/api/admin/tcgs", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/admin/tcgs", bytes.NewBuffer([]byte("{invalid}")))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Missing Fields", func(t *testing.T) {
		input := models.TCGInput{ID: "one"} // Missing Name
		body, _ := json.Marshal(input)
		req, _ := http.NewRequest("POST", "/api/admin/tcgs", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		input := models.TCGInput{ID: "one", Name: "One Piece"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("INSERT INTO tcg").WillReturnError(fmt.Errorf("conflict"))

		req, _ := http.NewRequest("POST", "/api/admin/tcgs", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTCGHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &TCGHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.TCGInput{Name: "New Name", IsActive: true}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("UPDATE tcg").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("mtg", "New Name"))

		r := chi.NewRouter()
		r.Put("/api/admin/tcgs/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/tcgs/mtg", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error (Not Found)", func(t *testing.T) {
		input := models.TCGInput{Name: "New Name", IsActive: true}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("UPDATE tcg").WillReturnError(fmt.Errorf("not found"))

		r := chi.NewRouter()
		r.Put("/api/admin/tcgs/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/tcgs/mtg", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestTCGHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &TCGHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("DELETE FROM tcg").WillReturnResult(sqlmock.NewResult(0, 1))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/mtg", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Conflict", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/mtg", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("DB Error Checking", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/mtg", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("DB Error Deleting", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("DELETE FROM tcg").WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/mtg", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("DELETE FROM tcg").WillReturnResult(sqlmock.NewResult(0, 0))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/mtg", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
