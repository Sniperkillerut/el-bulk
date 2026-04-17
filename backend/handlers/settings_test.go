package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestSettingsHandler_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	svc := service.NewSettingsService(store.NewSettingsStore(sqlxDB), nil)
	h := NewSettingsHandler(svc)

	t.Run("Success", func(t *testing.T) {
		svc.ResetCache()
		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("usd_to_cop_rate", "4500.5").
			AddRow("contact_email", "test@example.com")

		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/admin/settings", nil)
		rr := httptest.NewRecorder()
		h.Get(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var res models.Settings
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Equal(t, 4500.5, res.USDToCOPRate)
		assert.Equal(t, "test@example.com", res.ContactEmail)
	})

	t.Run("Defaults on Error", func(t *testing.T) {
		svc.ResetCache()
		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnError(assert.AnError)

		req, _ := http.NewRequest("GET", "/api/admin/settings", nil)
		rr := httptest.NewRecorder()
		h.Get(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestSettingsHandler_PublicGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	svc := service.NewSettingsService(store.NewSettingsStore(sqlxDB), nil)
	h := NewSettingsHandler(svc)

	t.Run("Returns only public fields", func(t *testing.T) {
		svc.ResetCache()
		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("usd_to_cop_rate", "4800.0").
			AddRow("eur_to_cop_rate", "5200.0").
			AddRow("ck_to_cop_rate", "4900.0").
			AddRow("contact_email", "tienda@example.com").
			AddRow("contact_phone", "3001234567").
			AddRow("flat_shipping_fee_cop", "10000").
			AddRow("hot_sales_threshold", "5").
			AddRow("hot_days_threshold", "7").
			AddRow("new_threshold_days", "14")

		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/settings", nil)
		rr := httptest.NewRecorder()
		h.PublicGet(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Decode into a raw map to assert exactly which keys are present in the JSON
		var data map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&data)

		// Public fields should be present with correct values
		assert.Equal(t, "tienda@example.com", data["contact_email"])
		assert.Equal(t, "3001234567", data["contact_phone"])
		assert.EqualValues(t, 10000, data["flat_shipping_fee_cop"])

		// Sensitive fields must NOT appear in the JSON response
		_, hasUSD := data["usd_to_cop_rate"]
		_, hasEUR := data["eur_to_cop_rate"]
		_, hasCK := data["ck_to_cop_rate"]
		_, hasHotSales := data["hot_sales_threshold"]
		_, hasHotDays := data["hot_days_threshold"]
		_, hasNewDays := data["new_days_threshold"]
		_, hasLastSync := data["last_set_sync"]
		assert.False(t, hasUSD, "exchange rates must not leak publicly")
		assert.False(t, hasEUR, "exchange rates must not leak publicly")
		assert.False(t, hasCK, "exchange rates must not leak publicly")
		assert.False(t, hasHotSales, "algorithm thresholds must not leak publicly")
		assert.False(t, hasHotDays, "algorithm thresholds must not leak publicly")
		assert.False(t, hasNewDays, "algorithm thresholds must not leak publicly")
		assert.False(t, hasLastSync, "operational data must not leak publicly")
	})

	t.Run("Returns full settings for admin", func(t *testing.T) {
		svc.ResetCache()
		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("usd_to_cop_rate", "5000.0").
			AddRow("contact_email", "admin@example.com")

		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/settings", nil)
		// Simulate admin context
		ctx := context.WithValue(req.Context(), middleware.IsAdminKey, true)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.PublicGet(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var data map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&data)

		assert.Equal(t, "admin@example.com", data["contact_email"])
		assert.EqualValues(t, 5000, data["usd_to_cop_rate"], "admin must see exchange rates even on public route")
	})

	t.Run("Returns empty public settings on DB error", func(t *testing.T) {
		svc.ResetCache()
		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnError(assert.AnError)

		req, _ := http.NewRequest("GET", "/api/settings", nil)
		rr := httptest.NewRecorder()
		h.PublicGet(rr, req)

		// Should still return 200 with an empty (but safe) public settings object
		assert.Equal(t, http.StatusOK, rr.Code)

		var data map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&data)
		_, hasUSD := data["usd_to_cop_rate"]
		assert.False(t, hasUSD, "no exchange rates even in error fallback")
	})
}

func TestSettingsHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	svc := service.NewSettingsService(store.NewSettingsStore(sqlxDB), nil)
	h := NewSettingsHandler(svc)

	t.Run("Success", func(t *testing.T) {
		svc.ResetCache()
		rate := 4400.0
		email := "new@example.com"
		input := struct {
			USDToCOPRate *float64 `json:"usd_to_cop_rate"`
			ContactEmail *string  `json:"contact_email"`
		}{
			USDToCOPRate: &rate,
			ContactEmail: &email,
		}
		body, _ := json.Marshal(input)

		// Two upserts
		mock.ExpectExec("INSERT INTO setting").WithArgs("usd_to_cop_rate", "4400.0000").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO setting").WithArgs("contact_email", email).WillReturnResult(sqlmock.NewResult(0, 1))

		// Post-update load
		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("usd_to_cop_rate", "4400.0"))

		req, _ := http.NewRequest("PUT", "/api/admin/settings", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Update All Fields", func(t *testing.T) {
		svc.ResetCache()
		usd := 4500.0
		eur := 4900.0
		addr := "Addr"
		phone := "123"
		email := "test@test.com"
		insta := "@insta"
		hours := "9-5"
		input := struct {
			USDToCOPRate     *float64 `json:"usd_to_cop_rate"`
			EURToCOPRate     *float64 `json:"eur_to_cop_rate"`
			ContactAddress   *string  `json:"contact_address"`
			ContactPhone     *string  `json:"contact_phone"`
			ContactEmail     *string  `json:"contact_email"`
			ContactInstagram *string  `json:"contact_instagram"`
			ContactHours     *string  `json:"contact_hours"`
		}{
			USDToCOPRate: &usd, EURToCOPRate: &eur, ContactAddress: &addr,
			ContactPhone: &phone, ContactEmail: &email, ContactInstagram: &insta, ContactHours: &hours,
		}
		body, _ := json.Marshal(input)

		// 7 upserts
		for i := 0; i < 7; i++ {
			mock.ExpectExec("INSERT INTO setting").WillReturnResult(sqlmock.NewResult(0, 1))
		}
		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("usd_to_cop_rate", "4500.0"))

		req, _ := http.NewRequest("PUT", "/api/admin/settings", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/admin/settings", bytes.NewBuffer([]byte("{invalid}")))
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		rate := 4400.0
		input := struct {
			USDToCOPRate *float64 `json:"usd_to_cop_rate"`
		}{USDToCOPRate: &rate}
		body, _ := json.Marshal(input)

		mock.ExpectExec("INSERT INTO setting").WillReturnError(fmt.Errorf("db error"))

		req, _ := http.NewRequest("PUT", "/api/admin/settings", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
