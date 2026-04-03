package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/el-bulk/backend/models"
)

func TestNoticeHandler_List(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewNoticeHandler(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "title", "slug"}).AddRow("n1", "Notice 1", "notice-1")
	mock.ExpectQuery("SELECT \\* FROM notice WHERE is_published = true").WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/api/notices", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNoticeHandler_GetBySlug(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewNoticeHandler(sqlxDB)

	mock.ExpectQuery("SELECT \\* FROM notice WHERE slug = \\$1").WithArgs("notice-1").WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow("n1", "Notice 1"))

	r := chi.NewRouter()
	r.Get("/api/notices/{slug}", h.GetBySlug)
	req := httptest.NewRequest("GET", "/api/notices/notice-1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNoticeHandler_AdminList(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewNoticeHandler(sqlxDB)

	mock.ExpectQuery("SELECT \\* FROM notice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("n1"))

	req := httptest.NewRequest("GET", "/api/admin/notices", nil)
	rr := httptest.NewRecorder()
	h.AdminList(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNoticeHandler_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewNoticeHandler(sqlxDB)

	input := models.NoticeInput{
		Title:       "Test",
		Slug:        "test",
		ContentHTML: "<p>Test</p>",
		IsPublished: false,
	}
	body, _ := json.Marshal(input)

	mock.ExpectQuery("INSERT INTO notice").WithArgs(input.Title, input.Slug, input.ContentHTML, nil, input.IsPublished).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow("n1", "Test"))

	req := httptest.NewRequest("POST", "/api/admin/notices", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNoticeHandler_Update(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewNoticeHandler(sqlxDB)

	input := models.NoticeInput{
		Title:       "Updated",
		Slug:        "updated",
		ContentHTML: "<p>Updated</p>",
		IsPublished: false,
	}
	body, _ := json.Marshal(input)

	mock.ExpectQuery("UPDATE notice").WithArgs(input.Title, input.Slug, input.ContentHTML, nil, input.IsPublished, "n1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow("n1", "Updated"))

	r := chi.NewRouter()
	r.Put("/api/admin/notices/{id}", h.Update)
	req := httptest.NewRequest("PUT", "/api/admin/notices/n1", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNoticeHandler_Delete(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewNoticeHandler(sqlxDB)

	mock.ExpectExec("DELETE FROM notice").WithArgs("n1").WillReturnResult(sqlmock.NewResult(1, 1))

	r := chi.NewRouter()
	r.Delete("/api/admin/notices/{id}", h.Delete)
	req := httptest.NewRequest("DELETE", "/api/admin/notices/n1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
