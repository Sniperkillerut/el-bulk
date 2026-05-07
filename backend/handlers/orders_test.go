package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestOrderHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	orderStore := store.NewOrderStore(sqlxDB)
	customerStore := store.NewCustomerStore(sqlxDB)
	productStore := store.NewProductStore(sqlxDB)
	orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
	h := NewOrderHandler(orderService)

	t.Run("Success", func(t *testing.T) {
		input := models.CreateOrderInput{
			FirstName:     "John",
			LastName:      "Doe",
			Phone:         "1234567890",
			PaymentMethod: "whatsapp",
			Items: []models.CreateOrderItem{
				{ProductID: "550e8400-e29b-41d4-a716-446655440012", Quantity: 2},
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
			WithArgs("550e8400-e29b-41d4-a716-446655440012").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price_reference", "price_source", "stock"}).
				AddRow("550e8400-e29b-41d4-a716-446655440012", "Product 1", 1.0, "tcgplayer", 10))

		// 3. Storage fetch
		mock.ExpectQuery("SELECT ps.product_id as product_id_temp, .* FROM product_storage ps .* WHERE ps.product_id IN \\(\\$1\\)").
			WithArgs("550e8400-e29b-41d4-a716-446655440012").
			WillReturnRows(sqlmock.NewRows([]string{"product_id_temp", "stored_in_id", "name", "quantity"}).AddRow("550e8400-e29b-41d4-a716-446655440012", "550e8400-e29b-41d4-a716-446655440013", "Box 1", 10))

		// 4. Place Order via SP
		mock.ExpectQuery("SELECT order_id, order_number FROM fn_place_order").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "order_number"}).
				AddRow("550e8400-e29b-41d4-a716-446655440010", "EB-20260325-ABCDEF1234567890"))

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var res map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Equal(t, "EB-20260325-ABCDEF1234567890", res["order_number"])
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
		settingsService.ResetCache()
		input := models.CreateOrderInput{
			FirstName: "John", Phone: "12345", PaymentMethod: "WhatsApp",
			Items: []models.CreateOrderItem{{ProductID: "550e8400-e29b-41d4-a716-446655440012", Quantity: 1}},
		}
		body, _ := json.Marshal(input)

		// Code fetches settings and products before PlaceOrder
		mock.ExpectQuery("SELECT .* FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}))
		mock.ExpectQuery("SELECT .* FROM product WHERE id IN").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("550e8400-e29b-41d4-a716-446655440012"))
		mock.ExpectQuery("SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Now PlaceOrder fails (mocking a direct SP call error or any previous error if we had a tx)
		mock.ExpectQuery("SELECT .* FROM fn_place_order").WillReturnError(fmt.Errorf("db error"))

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code) // Service currently returns 400 for its errors in handler
	})

	t.Run("Create Order Error", func(t *testing.T) {
		settingsService.ResetCache()
		input := models.CreateOrderInput{
			FirstName: "John", Phone: "12345", PaymentMethod: "WhatsApp",
			Items: []models.CreateOrderItem{{ProductID: "550e8400-e29b-41d4-a716-446655440012", Quantity: 1}},
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("SELECT .* FROM setting").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}))
		mock.ExpectQuery("SELECT .* FROM product WHERE id IN").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("550e8400-e29b-41d4-a716-446655440012"))
		mock.ExpectQuery("SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT .* FROM fn_place_order").WillReturnError(fmt.Errorf("order error"))

		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestOrderHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	orderStore := store.NewOrderStore(sqlxDB)
	customerStore := store.NewCustomerStore(sqlxDB)
	productStore := store.NewProductStore(sqlxDB)
	orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
	h := NewOrderHandler(orderService)

	t.Run("List Admin", func(t *testing.T) {
		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM view_order_list o").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "order_number", "customer_id", "status", "payment_method", "total_cop", "notes", "created_at", "completed_at", "customer_name", "item_count"}).
			AddRow("550e8400-e29b-41d4-a716-446655440010", "EB-1", "550e8400-e29b-41d4-a716-446655440011", "pending", "whatsapp", 1000.0, nil, time.Now(), nil, "John Doe", 1)

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
	settingsStore := store.NewSettingsStore(sqlxDB)
	settingsService := service.NewSettingsService(settingsStore, nil)
	orderStore := store.NewOrderStore(sqlxDB)
	customerStore := store.NewCustomerStore(sqlxDB)
	productStore := store.NewProductStore(sqlxDB)
	orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
	h := NewOrderHandler(orderService)

	t.Run("Success", func(t *testing.T) {
		orderID := "550e8400-e29b-41d4-a716-446655440010"
		customerID := "550e8400-e29b-41d4-a716-446655440011"
		mock.ExpectQuery("(?i)SELECT .* FROM \"order\"").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow(orderID, customerID))

		mock.ExpectQuery("(?i)SELECT .* FROM customer WHERE id = \\$1").
			WithArgs(customerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone", "id_number", "address", "created_at"}).
				AddRow(customerID, "John", "Doe", "john@example.com", "123", "ID123", "Main St", time.Now()))

		mock.ExpectQuery("(?i)SELECT \\* FROM view_order_item_enriched").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "product_id", "product_name", "product_set", "foil_treatment", "card_treatment", "condition", "unit_price_cop", "quantity", "stored_in_snapshot", "image_url", "stock", "stored_in"}).
				AddRow("oi1", orderID, "550e8400-e29b-41d4-a716-446655440012", "Product 1", "Set 1", "non_foil", "normal", "NM", 100.0, 1, "Box 1", "image.png", 10, []byte("[]")))

		r := chi.NewRouter()
		r.Get("/api/admin/orders/{id}", h.GetDetail)
		req, _ := http.NewRequest("GET", "/api/admin/orders/"+orderID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB Error GetDetail", func(t *testing.T) {
		orderID := "550e8400-e29b-41d4-a716-446655440010"
		mock.ExpectQuery("SELECT .* FROM \"order\"").WillReturnError(fmt.Errorf("db error"))
		r := chi.NewRouter()
		r.Get("/api/admin/orders/{id}", h.GetDetail)
		req, _ := http.NewRequest("GET", "/api/admin/orders/"+orderID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestOrderHandler_Update(t *testing.T) {
	t.Run("Update Status", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		customerID := "550e8400-e29b-41d4-a716-446655440011"
		input := map[string]interface{}{"status": "shipped"}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WithArgs(orderID).WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(orderID, "confirmed"))
		mock.ExpectQuery("SELECT status FROM \"order\" WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("confirmed"))

		mock.ExpectExec("(?i)UPDATE \"order\" SET status = COALESCE").
			WithArgs("shipped", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), orderID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"subtotal", "shipping", "tax"}).AddRow(1000.0, 0.0, 0.0))
		mock.ExpectExec("UPDATE \"order\" SET subtotal_cop = \\$1, total_cop = \\$2 WHERE id = \\$3").
			WithArgs(1000.0, 1000.0, orderID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		// For GetDetail call inside Update
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow(orderID, customerID))
		mock.ExpectQuery("SELECT .* FROM customer WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone"}).AddRow(customerID, "John", "Doe", "123"))
		mock.ExpectQuery("SELECT \\* FROM view_order_item_enriched").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("550e8400-e29b-41d4-a716-446655440020"))

		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+orderID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Tx Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		mock.ExpectBegin().WillReturnError(fmt.Errorf("tx error"))
		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/550e8400-e29b-41d4-a716-446655440010", bytes.NewBuffer([]byte(`{"status":"shipped"}`)))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Inventory Lock (Confirmed Order)", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		input := map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "550e8400-e29b-41d4-a716-446655440020", "quantity": 5},
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT status FROM \"order\" WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("confirmed"))
		mock.ExpectRollback()

		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+orderID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Metadata Lock (Shipped Order)", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		shippingValue := 20000.0
		input := map[string]interface{}{
			"shipping_cop":   &shippingValue,
			"payment_method": "cash",
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT status FROM \"order\" WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("shipped"))
		mock.ExpectRollback()

		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+orderID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Reject skipping confirmation", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		statusValue := "shipped"
		input := map[string]interface{}{
			"status": &statusValue,
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT status FROM \"order\" WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("pending"))
		mock.ExpectRollback()

		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+orderID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Pending to Cancelled cleans up pending stock", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		customerID := "550e8400-e29b-41d4-a716-446655440011"
		statusValue := "cancelled"
		input := map[string]interface{}{
			"status": &statusValue,
		}
		body, _ := json.Marshal(input)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WithArgs(orderID).WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(orderID, "pending"))

		// Initial status check
		mock.ExpectQuery("SELECT status FROM \"order\" WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("pending"))

		// Order info update
		mock.ExpectExec("(?i)UPDATE \"order\" SET status = COALESCE").
			WithArgs("cancelled", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), orderID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		// For GetDetail call inside Update
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow(orderID, customerID))
		mock.ExpectQuery("SELECT .* FROM customer WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone"}).AddRow(customerID, "John", "Doe", "123"))
		mock.ExpectQuery("SELECT \\* FROM view_order_item_enriched").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("550e8400-e29b-41d4-a716-446655440020"))

		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+orderID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Reject update for terminal status", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		statusValue := "shipped"
		input := map[string]interface{}{
			"status": &statusValue,
		}
		body, _ := json.Marshal(input)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT status FROM \"order\" WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("completed"))
		mock.ExpectRollback()

		r := chi.NewRouter()
		r.Put("/api/admin/orders/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+orderID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestOrderHandler_Confirm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		customerID := "550e8400-e29b-41d4-a716-446655440011"
		input := models.ConfirmOrderInput{
			Decrements: []models.StockDecrement{
				{ProductID: "550e8400-e29b-41d4-a716-446655440012", StorageID: "550e8400-e29b-41d4-a716-446655440013", Quantity: 2},
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectExec("SELECT fn_confirm_order").
			WithArgs(orderID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// For GetDetail follow-up
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow(orderID, customerID))
		mock.ExpectQuery("SELECT .* FROM customer WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone"}).AddRow(customerID, "John", "Doe", "123"))
		mock.ExpectQuery("SELECT \\* FROM view_order_item_enriched").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("oi1"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/confirm", h.Confirm)
		req, _ := http.NewRequest("POST", "/api/admin/orders/"+orderID+"/confirm", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Insufficient Stock", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)
		orderID := "550e8400-e29b-41d4-a716-446655440010"
		input := models.ConfirmOrderInput{
			Decrements: []models.StockDecrement{
				{ProductID: "550e8400-e29b-41d4-a716-446655440012", StorageID: "550e8400-e29b-41d4-a716-446655440013", Quantity: 20}, // More than 10
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectExec("SELECT fn_confirm_order").
			WithArgs(orderID, sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("Insufficient stock"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/confirm", h.Confirm)
		req, _ := http.NewRequest("POST", "/api/admin/orders/"+orderID+"/confirm", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Tx Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		mock.ExpectExec("SELECT fn_confirm_order").
			WithArgs(orderID, sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/confirm", h.Confirm)
		req, _ := http.NewRequest("POST", "/api/admin/orders/"+orderID+"/confirm", bytes.NewBuffer([]byte(`{"decrements":[]}`))) // Corrected body key from 'items' to 'decrements'
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrderHandler_RestoreStock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		orderID := "550e8400-e29b-41d4-a716-446655440010"
		customerID := "550e8400-e29b-41d4-a716-446655440011"
		input := map[string]interface{}{
			"increments": []models.StockDecrement{
				{ProductID: "550e8400-e29b-41d4-a716-446655440012", StorageID: "550e8400-e29b-41d4-a716-446655440013", Quantity: 2},
			},
		}
		body, _ := json.Marshal(input)

		mock.ExpectExec("SELECT fn_restore_order_stock").
			WithArgs(orderID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// GetDetail follow-up
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow(orderID, customerID))
		mock.ExpectQuery("SELECT .* FROM customer WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone"}).AddRow(customerID, "John", "Doe", "123"))
		mock.ExpectQuery("SELECT \\* FROM view_order_item_enriched").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("oi1"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/restore", h.RestoreStock)
		req, _ := http.NewRequest("POST", "/api/admin/orders/"+orderID+"/restore", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)
		orderID := "550e8400-e29b-41d4-a716-446655440010"
		body := []byte(`{"increments": []}`)
		mock.ExpectExec("SELECT fn_restore_order_stock").WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Post("/api/admin/orders/{id}/restore", h.RestoreStock)
		req, _ := http.NewRequest("POST", "/api/admin/orders/"+orderID+"/restore", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestOrderHandler_CancelMe(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		settingsStore := store.NewSettingsStore(sqlxDB)
		settingsService := service.NewSettingsService(settingsStore, nil)
		orderStore := store.NewOrderStore(sqlxDB)
		customerStore := store.NewCustomerStore(sqlxDB)
		productStore := store.NewProductStore(sqlxDB)
		orderService := service.NewOrderService(orderStore, productStore, customerStore, settingsService, nil)
		h := NewOrderHandler(orderService)

		userID := "550e8400-e29b-41d4-a716-446655440011"
		orderID := "550e8400-e29b-41d4-a716-446655440010"

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE \"order\" SET status = 'cancelled'").
			WithArgs(orderID, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		// For GetMeDetail (called after success)
		mock.ExpectQuery("SELECT .* FROM \"order\" WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id"}).AddRow(orderID, userID))
		mock.ExpectQuery("SELECT .* FROM customer WHERE id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone"}).AddRow(userID, "John", "Doe", "123"))
		mock.ExpectQuery("SELECT \\* FROM view_order_item_enriched").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("oi1"))

		r := chi.NewRouter()
		r.Post("/api/orders/me/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
			h.CancelMe(w, r.WithContext(ctx))
		})

		req, _ := http.NewRequest("POST", "/api/orders/me/"+orderID+"/cancel", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
