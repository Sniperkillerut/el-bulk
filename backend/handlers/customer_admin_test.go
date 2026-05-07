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
)

func TestCustomerAdminHandler_ListCustomers(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := &CustomerAdminHandler{DB: sqlxDB}

	mock.ExpectQuery("SELECT c.*").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name"}).AddRow("550e8400-e29b-41d4-a716-446655440010", "John", "Doe"))

	req := httptest.NewRequest("GET", "/admin/customers", nil)
	rr := httptest.NewRecorder()
	h.ListCustomers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCustomerAdminHandler_GetCustomerDetail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := &CustomerAdminHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		customerID := "550e8400-e29b-41d4-a716-446655440010"
		mock.ExpectQuery("SELECT \\* FROM customer WHERE id = \\$1").WithArgs(customerID).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).AddRow(customerID, "John", "Doe", "john@example.com"))
		mock.ExpectQuery("SELECT \\* FROM \"order\" WHERE customer_id = \\$1").WithArgs(customerID).WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT n.*, a.username as admin_name FROM customer_note").WithArgs(customerID).WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT EXISTS").WithArgs(customerID, "john@example.com").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		mock.ExpectQuery("SELECT \\* FROM client_request").WithArgs(customerID, "john@example.com").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT o.*, b.name as bounty_name FROM bounty_offer").WithArgs(customerID).WillReturnRows(sqlmock.NewRows([]string{"id"}))

		r := chi.NewRouter()
		r.Get("/admin/customers/{id}", h.GetCustomerDetail)
		req := httptest.NewRequest("GET", "/admin/customers/"+customerID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		customerID := "550e8400-e29b-41d4-a716-446655449999"
		mock.ExpectQuery("SELECT \\* FROM customer WHERE id = \\$1").WithArgs(customerID).WillReturnError(assert.AnError)

		r := chi.NewRouter()
		r.Get("/admin/customers/{id}", h.GetCustomerDetail)
		req := httptest.NewRequest("GET", "/admin/customers/"+customerID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestCustomerAdminHandler_AddNote(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	h := &CustomerAdminHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		customerID := "550e8400-e29b-41d4-a716-446655440010"
		mock.ExpectExec("INSERT INTO customer_note").WithArgs(customerID, nil, "Test note", nil).WillReturnResult(sqlmock.NewResult(1, 1))

		body, _ := json.Marshal(map[string]string{"content": "Test note"})
		r := chi.NewRouter()
		r.Post("/admin/customers/{id}/notes", h.AddNote)
		req := httptest.NewRequest("POST", "/admin/customers/"+customerID+"/notes", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}
