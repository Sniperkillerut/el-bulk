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

func TestNoticeHandler_AdminList(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := NewNoticeHandler(service.NewNoticeService(store.NewNoticeStore(sqlxDB)))

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "content_html", "is_published", "created_at"}).
			AddRow("1", "Test Notice", "Content", true, time.Now())
		mock.ExpectQuery("(?i)SELECT \\* FROM notice ORDER BY created_at DESC").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/admin/notices", nil)
		rr := httptest.NewRecorder()
		h.AdminList(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectQuery("(?i)SELECT \\* FROM notice").WillReturnError(fmt.Errorf("db error"))
		req, _ := http.NewRequest("GET", "/api/admin/notices", nil)
		rr := httptest.NewRecorder()
		h.AdminList(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNoticeHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := NewNoticeHandler(service.NewNoticeService(store.NewNoticeStore(sqlxDB)))

	t.Run("Success", func(t *testing.T) {
		input := models.NoticeInput{Title: "New Notice", ContentHTML: "Body"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("INSERT INTO notice").WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow("1", "New Notice"))

		req, _ := http.NewRequest("POST", "/api/admin/notices", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/admin/notices", bytes.NewBuffer([]byte("{invalid}")))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestNoticeHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := NewNoticeHandler(service.NewNoticeService(store.NewNoticeStore(sqlxDB)))

	t.Run("Success", func(t *testing.T) {
		input := models.NoticeInput{Title: "Updated Notice", ContentHTML: "New Body"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("UPDATE notice").WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow("1", "Updated Notice"))

		r := chi.NewRouter()
		r.Put("/api/admin/notices/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/notices/1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		r := chi.NewRouter()
		r.Put("/api/admin/notices/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/notices/1", bytes.NewBuffer([]byte("{invalid}")))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestNoticeHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := NewNoticeHandler(service.NewNoticeService(store.NewNoticeStore(sqlxDB)))

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM notice").WillReturnResult(sqlmock.NewResult(0, 1))

		r := chi.NewRouter()
		r.Delete("/api/admin/notices/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/notices/1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		r := chi.NewRouter()
		r.Delete("/api/admin/notices/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/notices/abc", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
