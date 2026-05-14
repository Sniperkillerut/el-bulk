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
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestTCGHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &TCGHandler{Service: service.NewTCGService(store.NewTCGStore(sqlxDB), nil)}

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "is_active", "created_at", "item_count"}).
			AddRow("550e8400-e29b-41d4-a716-446655440003", "Magic", true, time.Now(), 0)
		mock.ExpectQuery("(?i)SELECT t.*, COUNT\\(p.id\\) as item_count FROM tcg t LEFT JOIN product p ON t.id = p.tcg GROUP BY t.id ORDER BY t.name").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/admin/tcgs", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error", func(t *testing.T) {
		h.Service.ResetCache()
		mock.ExpectQuery("(?i)SELECT t.*, COUNT").WillReturnError(fmt.Errorf("db error"))
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
	h := &TCGHandler{Service: service.NewTCGService(store.NewTCGStore(sqlxDB), nil)}

	t.Run("Success", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440020"
		input := models.TCGInput{ID: tcgID, Name: "One Piece"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("INSERT INTO tcg").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(tcgID, "One Piece"))

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
		input := models.TCGInput{ID: "550e8400-e29b-41d4-a716-446655440020"} // Missing Name
		body, _ := json.Marshal(input)
		req, _ := http.NewRequest("POST", "/api/admin/tcgs", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440020"
		input := models.TCGInput{ID: tcgID, Name: "One Piece"}
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
	h := &TCGHandler{Service: service.NewTCGService(store.NewTCGStore(sqlxDB), nil)}

	t.Run("Success", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440003"
		input := models.TCGInput{Name: "New Name", IsActive: true}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("UPDATE tcg").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(tcgID, "New Name"))

		r := chi.NewRouter()
		r.Put("/api/admin/tcgs/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/tcgs/"+tcgID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error (Not Found)", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440003"
		input := models.TCGInput{Name: "New Name", IsActive: true}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("UPDATE tcg").WillReturnError(fmt.Errorf("not found"))

		r := chi.NewRouter()
		r.Put("/api/admin/tcgs/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/tcgs/"+tcgID, bytes.NewBuffer(body))
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
	h := &TCGHandler{Service: service.NewTCGService(store.NewTCGStore(sqlxDB), nil)}

	t.Run("Success", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440003"
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("DELETE FROM tcg").WillReturnResult(sqlmock.NewResult(0, 1))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/"+tcgID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Conflict", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440003"
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/"+tcgID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("DB Error Checking", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440003"
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/"+tcgID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("DB Error Deleting", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440003"
		mock.ExpectQuery("SELECT COUNT.* FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("DELETE FROM tcg").WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Delete("/api/admin/tcgs/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/tcgs/"+tcgID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTCGHandler_Sync(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := NewTCGHandler(service.NewTCGService(store.NewTCGStore(sqlxDB), nil), testWorkerPool(sqlxDB), &NopAuditer{})

	t.Run("SyncSets", func(t *testing.T) {
		mock.ExpectQuery("(?i)INSERT INTO job").WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

		req, _ := http.NewRequest("POST", "/api/admin/tcgs/mtg/sync-sets", nil)
		rr := httptest.NewRecorder()
		h.SyncSets(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SyncPrices", func(t *testing.T) {
		mock.ExpectQuery("(?i)INSERT INTO job").WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

		req, _ := http.NewRequest("POST", "/api/admin/tcgs/mtg/sync-prices", nil)
		rr := httptest.NewRecorder()
		h.SyncPrices(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
