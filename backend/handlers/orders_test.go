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

func TestOrderHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &OrderHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.CreateOrderInput{
			FirstName:     "John",
			LastName:      "Doe",
			Phone:         "1234567890",
			PaymentMethod: "whatsapp",
			Items: []models.CreateOrderItem{
				{ProductID: "p1", Quantity: 2},
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		
		// 1. Customer upsert
		mock.ExpectQuery("INSERT INTO customers").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("c1"))

		// 2. Settings
		mock.ExpectQuery("SELECT key, value FROM settings").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("usd_to_cop_rate", "4000").
				AddRow("eur_to_cop_rate", "4400"))

		// 3. Products fetch
		mock.ExpectQuery("SELECT .* FROM products WHERE id IN \\(\\$1\\)").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price_reference", "price_source", "stock"}).
				AddRow("p1", "Product 1", 1.0, "tcgplayer", 10))

		// 4. Create Order
		mock.ExpectQuery("INSERT INTO orders").
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_number", "customer_id", "status", "payment_method", "total_cop", "notes", "created_at"}).
				AddRow("o1", "EB-20260325-1234", "c1", "pending", "whatsapp", 8000.0, "", time.Now()))

		// 5. Create Order Items
		mock.ExpectQuery("SELECT ps.product_id as product_id_temp, .* FROM product_stored_in ps .* WHERE ps.product_id IN \\(\\$1\\)").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id_temp", "stored_in_id", "name", "quantity"}).AddRow("p1", "loc1", "Box 1", 10))

		mock.ExpectExec("INSERT INTO order_items").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var res map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Equal(t, "EB-20260325-1234", res["order_number"])
	})

	t.Run("Validation Failure", func(t *testing.T) {
		input := models.CreateOrderInput{FirstName: "John"} // Missing phone, payment method, items
		body, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Tx Error", func(t *testing.T) {
		input := models.CreateOrderInput{
			FirstName: "John", Phone: "12345", PaymentMethod: "WhatsApp",
			Items: []models.CreateOrderItem{{ProductID: "p1", Quantity: 1}},
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin().WillReturnError(fmt.Errorf("tx error"))

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Customer Upsert Error", func(t *testing.T) {
		input := models.CreateOrderInput{
			FirstName: "John", Phone: "12345", PaymentMethod: "WhatsApp",
			Items: []models.CreateOrderItem{{ProductID: "p1", Quantity: 1}},
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO customers").WillReturnError(fmt.Errorf("customer error"))
		mock.ExpectRollback()

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Create Order Error", func(t *testing.T) {
		input := models.CreateOrderInput{
			FirstName: "John", Phone: "12345", PaymentMethod: "WhatsApp",
			Items: []models.CreateOrderItem{{ProductID: "p1", Quantity: 1}},
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO customers").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("c1"))
		mock.ExpectQuery("SELECT .* FROM settings").WillReturnRows(sqlmock.NewRows([]string{"usd_to_cop_rate", "eur_to_cop_rate"}).AddRow(4000, 4400))
		mock.ExpectQuery("SELECT .* FROM products WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price_reference", "price_source", "stock"}).AddRow("p1", "Product 1", 1.0, "tcgplayer", 10))
		mock.ExpectQuery("INSERT INTO orders").WillReturnError(fmt.Errorf("order error"))
		mock.ExpectRollback()

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestOrderHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &OrderHandler{DB: sqlxDB}

	t.Run("List Admin", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM orders").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "order_number", "customer_id", "status", "payment_method", "total_cop", "created_at", "customer_name", "item_count"}).
			AddRow("o1", "EB-1", "c1", "pending", "whatsapp", 1000.0, time.Now(), "John Doe", 1)

		mock.ExpectQuery("SELECT o\\.*").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/admin/orders?page=1&page_size=20", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestOrderHandler_GetDetail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &OrderHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM orders WHERE id = \\$1").
			WithArgs("o1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow("o1", "c1"))

		mock.ExpectQuery("SELECT .* FROM customers WHERE id = \\$1").
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "first_name"}).AddRow("c1", "John"))

		mock.ExpectQuery("SELECT .* FROM order_items WHERE order_id = \\$1").
			WithArgs("o1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "product_name"}).AddRow("oi1", "p1", "Product 1"))

		mock.ExpectQuery("SELECT .* FROM products WHERE id IN \\(\\$1\\)").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "stock"}).AddRow("p1", 10))

		mock.ExpectQuery("SELECT ps.product_id as product_id_temp, .* FROM product_stored_in ps .* WHERE ps.product_id IN \\(\\$1\\)").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id_temp", "stored_in_id", "name", "quantity"}).AddRow("p1", "loc1", "Box 1", 10))

		r := chi.NewRouter()
		r.Get("/api/admin/orders/{id}", h.GetDetail)
		req, _ := http.NewRequest("GET", "/api/admin/orders/o1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error GetDetail", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM orders").WillReturnError(fmt.Errorf("db error"))
		r := chi.NewRouter()
		r.Get("/api/admin/orders/{id}", h.GetDetail)
		req, _ := http.NewRequest("GET", "/api/admin/orders/o1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestOrderHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &OrderHandler{DB: sqlxDB}

	t.Run("Update Status", func(t *testing.T) {
		input := map[string]interface{}{"status": "shipped"}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE orders SET status = \\$1 WHERE id = \\$2").
			WithArgs("shipped", "o1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery("SELECT COALESCE.* FROM order_items").
			WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1000.0))
		mock.ExpectExec("UPDATE orders SET total_cop = \\$1 WHERE id = \\$2").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		// For GetDetail call inside Update
		mock.ExpectQuery("SELECT .* FROM orders WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow("o1", "c1"))
		mock.ExpectQuery("SELECT .* FROM customers WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("c1"))
		mock.ExpectQuery("SELECT .* FROM order_items").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("oi1"))

		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/o1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Tx Error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(fmt.Errorf("tx error"))
		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/EB-123", bytes.NewBuffer([]byte(`{"status":"paid"}`)))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestOrderHandler_Complete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &OrderHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.CompleteOrderInput{
			Decrements: []models.StockDecrement{
				{ProductID: "p1", StoredInID: "loc1", Quantity: 2},
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("SELECT .* FROM orders WHERE id = \\$1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("o1", "pending"))

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COALESCE.* FROM product_stored_in").
			WithArgs("p1", "loc1").
			WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(10))

		mock.ExpectExec("UPDATE product_stored_in").
			WithArgs(2, "p1", "loc1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("UPDATE orders SET status = 'completed'").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		// For GetDetail
		mock.ExpectQuery("SELECT .* FROM orders WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow("o1", "c1"))
		mock.ExpectQuery("SELECT .* FROM customers WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("c1"))
		mock.ExpectQuery("SELECT .* FROM order_items").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("oi1"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/complete", h.Complete)
		req, _ := http.NewRequest("POST", "/api/admin/orders/o1/complete", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Insufficient Stock", func(t *testing.T) {
		input := models.CompleteOrderInput{
			Decrements: []models.StockDecrement{
				{ProductID: "p1", StoredInID: "loc1", Quantity: 20}, // More than 10
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("SELECT .* FROM orders WHERE id = \\$1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("o1", "pending"))

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COALESCE.* FROM product_stored_in").
			WithArgs("p1", "loc1").
			WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(10))

		mock.ExpectRollback()

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/complete", h.Complete)
		req, _ := http.NewRequest("POST", "/api/admin/orders/o1/complete", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Tx Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM orders WHERE id = \\$1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("o1", "pending"))
		mock.ExpectBegin().WillReturnError(fmt.Errorf("tx error"))
		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/complete", h.Complete)
		req, _ := http.NewRequest("POST", "/api/admin/orders/EB-123/complete", bytes.NewBuffer([]byte(`{"items":[]}`)))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
