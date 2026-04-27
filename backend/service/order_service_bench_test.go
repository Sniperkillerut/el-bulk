package service

import (
	"context"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"testing"
)

func BenchmarkUpdateOrder_AddedItems(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	orderStore := store.NewOrderStore(sqlxDB)
	orderService := NewOrderService(orderStore, nil, nil, nil, nil)

	ctx := context.Background()
	orderID := "test-order-id"

	addedItems := make([]struct {
		ProductID    string  `json:"product_id"`
		Quantity     int     `json:"quantity"`
		UnitPriceCOP float64 `json:"unit_price_cop"`
	}, 100)
	for i := 0; i < 100; i++ {
		addedItems[i] = struct {
			ProductID    string  `json:"product_id"`
			Quantity     int     `json:"quantity"`
			UnitPriceCOP float64 `json:"unit_price_cop"`
		}{
			ProductID:    "test-product-id",
			Quantity:     1,
			UnitPriceCOP: 1000,
		}
	}

	input := models.UpdateOrderInput{
		AddedItems: addedItems,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(`SELECT \* FROM "order" WHERE id = \$1`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(orderID, "pending"))

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT status FROM "order" WHERE id = \$1`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("pending"))

		// Generate the mock expectations for the new IN query
		inQueryArgs := make([]driver.Value, 100)
		for j := 0; j < 100; j++ {
			inQueryArgs[j] = "test-product-id"
		}

		mock.ExpectQuery(`SELECT id, name, set_name, foil_treatment, card_treatment, condition, stock FROM product WHERE id IN \(\?.*\)`).
			WithArgs(inQueryArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "set_name", "foil_treatment", "card_treatment", "condition", "stock"}).
				AddRow("test-product-id", "Test Product", nil, "Normal", "Normal", nil, 1000))

		var orderItemArgs []driver.Value
		var storageArgs []driver.Value
		for j := 0; j < 100; j++ {
			orderItemArgs = append(orderItemArgs, orderID, "test-product-id", "Test Product", nil, "Normal", "Normal", nil, 1, 1000.0)
			storageArgs = append(storageArgs, "test-product-id", 1)
		}

		mock.ExpectExec(`INSERT INTO order_item \(order_id, product_id, product_name, product_set, foil_treatment, card_treatment, condition, quantity, unit_price_cop\) VALUES `).
			WithArgs(orderItemArgs...).
			WillReturnResult(sqlmock.NewResult(1, 100))

		mock.ExpectExec(`INSERT INTO product_storage \(product_id, storage_id, quantity\) VALUES .* ON CONFLICT`).
			WithArgs(storageArgs...).
			WillReturnResult(sqlmock.NewResult(1, 100))

		mock.ExpectQuery(`SELECT.*COALESCE\(.*`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"subtotal", "shipping", "tax"}).AddRow(1000, 0, 0))

		mock.ExpectExec(`UPDATE "order" SET subtotal_cop = \$1, total_cop = \$2 WHERE id = \$3`).
			WithArgs(1000.0, 1000.0, orderID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err = orderService.UpdateOrder(ctx, orderID, input)
	}
}
