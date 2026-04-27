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

	mock.ExpectQuery("SELECT c.*").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name"}).AddRow("c1", "John", "Doe"))

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
		mock.ExpectQuery("SELECT \\* FROM customer WHERE id = \\$1").WithArgs("c1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).AddRow("c1", "John", "Doe", "john@example.com"))
		mock.ExpectQuery("SELECT \\* FROM \"order\" WHERE customer_id = \\$1").WithArgs("c1").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT n.*, a.username as admin_name FROM customer_note").WithArgs("c1").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT EXISTS").WithArgs("c1", "john@example.com").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		mock.ExpectQuery("SELECT \\* FROM client_request").WithArgs("c1", "john@example.com").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT o.*, b.name as bounty_name FROM bounty_offer").WithArgs("c1").WillReturnRows(sqlmock.NewRows([]string{"id"}))

		r := chi.NewRouter()
		r.Get("/admin/customers/{id}", h.GetCustomerDetail)
		req := httptest.NewRequest("GET", "/admin/customers/c1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM customer WHERE id = \\$1").WithArgs("nonexistent").WillReturnError(assert.AnError)

		r := chi.NewRouter()
		r.Get("/admin/customers/{id}", h.GetCustomerDetail)
		req := httptest.NewRequest("GET", "/admin/customers/nonexistent", nil)
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
		mock.ExpectExec("INSERT INTO customer_note").WithArgs("c1", nil, "Test note", nil).WillReturnResult(sqlmock.NewResult(1, 1))

		body, _ := json.Marshal(map[string]string{"content": "Test note"})
		r := chi.NewRouter()
		r.Post("/admin/customers/{id}/notes", h.AddNote)
		req := httptest.NewRequest("POST", "/admin/customers/c1/notes", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}
