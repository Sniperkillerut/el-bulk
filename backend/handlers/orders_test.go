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

		// 1. Settings
		mock.ExpectQuery("SELECT key, value FROM setting").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("usd_to_cop_rate", "4000").
				AddRow("eur_to_cop_rate", "4400"))

		// 2. Products fetch
		mock.ExpectQuery("SELECT .* FROM product WHERE id IN \\(\\$1\\)").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price_reference", "price_source", "stock"}).
				AddRow("p1", "Product 1", 1.0, "tcgplayer", 10))

		// 3. Storage fetch
		mock.ExpectQuery("SELECT ps.product_id as product_id_temp, .* FROM product_storage ps .* WHERE ps.product_id IN \\(\\$1\\)").
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id_temp", "stored_in_id", "name", "quantity"}).AddRow("p1", "loc1", "Box 1", 10))

		// 4. Place Order via SP
		mock.ExpectQuery("SELECT order_id, order_number FROM fn_place_order").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "order_number"}).
				AddRow("o1", "EB-20260325-1234"))

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
		mock.ExpectQuery("INSERT INTO customer").WillReturnError(fmt.Errorf("customer error"))
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
		mock.ExpectQuery("INSERT INTO customer").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("c1"))
		mock.ExpectQuery("SELECT .* FROM setting").WillReturnRows(sqlmock.NewRows([]string{"usd_to_cop_rate", "eur_to_cop_rate"}).AddRow(4000, 4400))
		mock.ExpectQuery("SELECT .* FROM product WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price_reference", "price_source", "stock"}).AddRow("p1", "Product 1", 1.0, "tcgplayer", 10))
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
		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM view_order_list o").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "order_number", "customer_id", "status", "payment_method", "total_cop", "notes", "created_at", "completed_at", "customer_name", "item_count"}).
			AddRow("o1", "EB-1", "c1", "pending", "whatsapp", 1000.0, nil, time.Now(), nil, "John Doe", 1)

		mock.ExpectQuery("(?i)SELECT \\* FROM view_order_list o").WillReturnRows(rows)

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
		mock.ExpectQuery("(?i)SELECT .* FROM \"order\"").
			WithArgs("o1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow("o1", "c1"))

		mock.ExpectQuery("(?i)SELECT .* FROM customer WHERE id = \\$1").
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone", "id_number", "address", "created_at"}).
				AddRow("c1", "John", "Doe", "john@example.com", "123", "ID123", "Main St", time.Now()))

		mock.ExpectQuery("(?i)SELECT \\* FROM view_order_item_enriched").
			WithArgs("o1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "product_id", "product_name", "product_set", "foil_treatment", "card_treatment", "condition", "unit_price_cop", "quantity", "stored_in_snapshot", "image_url", "stock", "stored_in"}).
				AddRow("oi1", "o1", "p1", "Product 1", "Set 1", "non_foil", "normal", "NM", 100.0, 1, "Box 1", "image.png", 10, []byte("[]")))

		r := chi.NewRouter()
		r.Get("/api/admin/orders/{id}", h.GetDetail)
		req, _ := http.NewRequest("GET", "/api/admin/orders/o1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error GetDetail", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM \"order\"").WillReturnError(fmt.Errorf("db error"))
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
		mock.ExpectExec("UPDATE \"order\" SET status = \\$1 WHERE id = \\$2").
			WithArgs("shipped", "o1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery("SELECT COALESCE.* FROM order_item").
			WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1000.0))
		mock.ExpectExec("UPDATE \"order\" SET total_cop = \\$1 WHERE id = \\$2").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		// For GetDetail call inside Update
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow("o1", "c1"))
		mock.ExpectQuery("SELECT .* FROM customer WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone"}).AddRow("c1", "John", "Doe", "123"))
		mock.ExpectQuery("SELECT \\* FROM view_order_item_enriched").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("oi1"))

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

func TestOrderHandler_Confirm(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &OrderHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.ConfirmOrderInput{
			Decrements: []models.StockDecrement{
				{ProductID: "p1", StorageID: "loc1", Quantity: 2},
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectExec("SELECT fn_confirm_order").
			WithArgs("o1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// For GetDetail follow-up
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow("o1", "c1"))
		mock.ExpectQuery("SELECT .* FROM customer WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone"}).AddRow("c1", "John", "Doe", "123"))
		mock.ExpectQuery("SELECT \\* FROM view_order_item_enriched").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("oi1"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/confirm", h.Confirm)
		req, _ := http.NewRequest("POST", "/api/admin/orders/o1/confirm", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Insufficient Stock", func(t *testing.T) {
		input := models.ConfirmOrderInput{
			Decrements: []models.StockDecrement{
				{ProductID: "p1", StorageID: "loc1", Quantity: 20}, // More than 10
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectExec("SELECT fn_confirm_order").
			WithArgs("o1", sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("Insufficient stock"))


		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/confirm", h.Confirm)
		req, _ := http.NewRequest("POST", "/api/admin/orders/o1/confirm", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Tx Error", func(t *testing.T) {
		mock.ExpectExec("SELECT fn_confirm_order").
			WithArgs("EB-123", sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/confirm", h.Confirm)
		req, _ := http.NewRequest("POST", "/api/admin/orders/EB-123/confirm", bytes.NewBuffer([]byte(`{"items":[]}`)))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
