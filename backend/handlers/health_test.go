package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Ping(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewHealthHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectPing()
		req := httptest.NewRequest("GET", "/health/ping", nil)
		rr := httptest.NewRecorder()
		h.Ping(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp map[string]string
		json.NewDecoder(rr.Body).Decode(&resp)
		assert.Equal(t, "ok", resp["status"])
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectPing().WillReturnError(assert.AnError)
		req := httptest.NewRequest("GET", "/health/ping", nil)
		rr := httptest.NewRecorder()
		h.Ping(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var resp map[string]string
		json.NewDecoder(rr.Body).Decode(&resp)
		assert.Equal(t, "error", resp["status"])
	})

	t.Run("No DB", func(t *testing.T) {
		hNoDB := NewHealthHandler(nil)
		req := httptest.NewRequest("GET", "/health/ping", nil)
		rr := httptest.NewRecorder()
		hNoDB.Ping(rr, req)

		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	})
}

func TestHealthHandler_GetStats(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := NewHealthHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT pg_size_pretty").WillReturnRows(sqlmock.NewRows([]string{"pg_size_pretty"}).AddRow("10 MB"))
		mock.ExpectQuery("SELECT CASE WHEN").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(95.5))
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_stat_activity").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"order\"").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM bounty_offer").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM client_request").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))
		mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT key\\) FROM translation").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
		mock.ExpectQuery("SELECT locale, COUNT\\(\\*\\) as count FROM translation").WillReturnRows(sqlmock.NewRows([]string{"locale", "count"}).AddRow("en", 10).AddRow("es", 8))

		req := httptest.NewRequest("GET", "/health/stats", nil)
		rr := httptest.NewRecorder()
		h.GetStats(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var stats DBStats
		json.NewDecoder(rr.Body).Decode(&stats)
		assert.Equal(t, "10 MB", stats.DatabaseSize)
		assert.Equal(t, 95.5, stats.CacheHitRatio)
		assert.Equal(t, 100, stats.TotalProducts)
		assert.Len(t, stats.TranslationProgress, 2)
	})

	t.Run("No DB", func(t *testing.T) {
		hNoDB := NewHealthHandler(nil)
		req := httptest.NewRequest("GET", "/health/stats", nil)
		rr := httptest.NewRecorder()
		hNoDB.GetStats(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp map[string]string
		json.NewDecoder(rr.Body).Decode(&resp)
		assert.Equal(t, "degraded", resp["status"])
	})
}
