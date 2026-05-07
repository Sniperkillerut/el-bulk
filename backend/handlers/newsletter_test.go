package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNewsletterHandler_Subscribe(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := testNewsletterHandler(sqlxDB)

	t.Run("New Subscription", func(t *testing.T) {
		email := "test@example.com"
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM newsletter_subscriber").WithArgs(email).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery("SELECT id FROM customer WHERE email = \\$1").WithArgs(email).WillReturnError(assert.AnError) // Customer not found
		mock.ExpectExec("INSERT INTO newsletter_subscriber").WithArgs(email, nil).WillReturnResult(sqlmock.NewResult(1, 1))

		body, _ := json.Marshal(map[string]string{"email": email})
		req := httptest.NewRequest("POST", "/newsletter/subscribe", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Subscribe(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Already Subscribed", func(t *testing.T) {
		email := "test@example.com"
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM newsletter_subscriber").WithArgs(email).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		body, _ := json.Marshal(map[string]string{"email": email})
		req := httptest.NewRequest("POST", "/newsletter/subscribe", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Subscribe(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Subscribed with Customer Match", func(t *testing.T) {
		email := "cust@example.com"
		custID := "550e8400-e29b-41d4-a716-446655440010"
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM newsletter_subscriber").WithArgs(email).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery("SELECT id FROM customer").WithArgs(email).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(custID))
		mock.ExpectExec("INSERT INTO newsletter_subscriber").WithArgs(email, custID).WillReturnResult(sqlmock.NewResult(1, 1))

		body, _ := json.Marshal(map[string]string{"email": email})
		req := httptest.NewRequest("POST", "/newsletter/subscribe", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Subscribe(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestNewsletterHandler_Unsubscribe(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := testNewsletterHandler(sqlxDB)

	email := "test@example.com"
	mock.ExpectExec("DELETE FROM newsletter_subscriber").WithArgs(email).WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest("POST", "/newsletter/unsubscribe", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.Unsubscribe(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewsletterHandler_AdminGetSubscribers(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := testNewsletterHandler(sqlxDB)

	mock.ExpectQuery("SELECT n.*, c.first_name, c.last_name").WillReturnRows(sqlmock.NewRows([]string{"email", "customer_id", "first_name", "last_name"}).
		AddRow("test@example.com", "550e8400-e29b-41d4-a716-446655440010", "John", "Doe"))

	req := httptest.NewRequest("GET", "/admin/newsletter/subscribers", nil)
	rr := httptest.NewRecorder()
	h.AdminGetSubscribers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
