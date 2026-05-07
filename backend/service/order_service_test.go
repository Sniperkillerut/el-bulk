package service

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderService_GenerateOrderNumber(t *testing.T) {
	s := &OrderService{}

	orderNumber := s.GenerateOrderNumber()

	// Format should be EB-YYYYMMDD-HEX(16 chars)
	// Example: EB-20260325-ABCDEF1234567890

	assert.True(t, strings.HasPrefix(orderNumber, "EB-"), "Should start with EB-")

	today := time.Now().Format("20060102")
	assert.Contains(t, orderNumber, today, "Should contain today's date")

	// Regexp to match the full format
	// EB- followed by 8 digits, followed by -, followed by 16 hex chars
	re := regexp.MustCompile(`^EB-\d{8}-[0-9A-F]{16}$`)
	assert.True(t, re.MatchString(orderNumber), "Should match the expected format ^EB-\\d{8}-[0-9A-F]{16}$, got: %s", orderNumber)

	// Check for randomness (unlikely to have collision in 2 tries with 64 bits)
	anotherOrderNumber := s.GenerateOrderNumber()
	assert.NotEqual(t, orderNumber, anotherOrderNumber, "Sequential order numbers should be different")
}

func TestOrderService_ConfirmOrder(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	orderStore := store.NewOrderStore(sqlxDB)
	productStore := store.NewProductStore(sqlxDB)
	mockAudit := new(MockAuditService)
	// We need 5 args for NewOrderService: s, ps, cs, settings, audit
	s := NewOrderService(orderStore, productStore, nil, nil, mockAudit, NewPDFService(), NewEmailService())

	t.Run("ConfirmOrder Success", func(t *testing.T) {
		orderID := "o-1"
		decrements := []models.StockDecrement{
			{ProductID: "p-1", StorageID: "s1", Quantity: 2},
		}

		// OrderStore.ConfirmOrder calls SELECT fn_confirm_order($1::uuid, $2::jsonb)
		sqlMock.ExpectExec(`SELECT fn_confirm_order`).
			WithArgs(orderID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mockAudit.On("LogAction", mock.Anything, "CONFIRM_ORDER", "order", orderID, mock.Anything).Return()

		err := s.ConfirmOrder(context.Background(), orderID, decrements)
		assert.NoError(t, err)
		mockAudit.AssertExpectations(t)
	})
}

func TestOrderService_RestoreStock(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	orderStore := store.NewOrderStore(sqlxDB)
	productStore := store.NewProductStore(sqlxDB)
	mockAudit := new(MockAuditService)
	s := NewOrderService(orderStore, productStore, nil, nil, mockAudit, NewPDFService(), NewEmailService())

	t.Run("RestoreStock Success", func(t *testing.T) {
		orderID := "o-1"
		increments := []models.StockDecrement{
			{ProductID: "p-1", Quantity: 3},
		}

		// OrderStore.RestoreStock calls SELECT fn_restore_order_stock($1::uuid, $2::jsonb)
		sqlMock.ExpectExec(`SELECT fn_restore_order_stock`).
			WithArgs(orderID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mockAudit.On("LogAction", mock.Anything, "RESTORE_STOCK", "order", orderID, mock.Anything).Return()

		err := s.RestoreStock(context.Background(), orderID, increments)
		assert.NoError(t, err)
		mockAudit.AssertExpectations(t)
	})
}

func TestOrderService_ListOrders(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	orderStore := store.NewOrderStore(sqlxDB)
	s := NewOrderService(orderStore, nil, nil, nil, nil, NewPDFService(), NewEmailService())

	t.Run("ListOrders Success", func(t *testing.T) {
		// GetOrderCount calls SELECT COUNT(*) FROM view_order_list o
		sqlMock.ExpectQuery(`SELECT COUNT`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// ListWithCustomer calls SELECT * FROM view_order_list o ...
		sqlMock.ExpectQuery(`SELECT`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_number", "customer_name"}).
				AddRow("o-1", "EB-1", "John Doe"))

		res, total, err := s.ListOrders(context.Background(), "", []interface{}{}, 1, 20)
		if err != nil {
			t.Fatalf("ListOrders failed: %v", err)
		}
		assert.Equal(t, 1, total)
		if !assert.Len(t, res, 1) {
			t.FailNow()
		}
		assert.Equal(t, "John Doe", res[0].CustomerName)
	})
}
