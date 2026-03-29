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
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestSettingsHandler_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &SettingsHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		ResetSettingsCache()
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
		ResetSettingsCache()
		mock.ExpectQuery("SELECT key, value FROM setting").WillReturnError(assert.AnError)

		req, _ := http.NewRequest("GET", "/api/admin/settings", nil)
		rr := httptest.NewRecorder()
		h.Get(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code) // loadSettings returns defaults on error
		var res models.Settings
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Equal(t, 4200.0, res.USDToCOPRate)
	})
}

func TestSettingsHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &SettingsHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		ResetSettingsCache()
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
		ResetSettingsCache()
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
