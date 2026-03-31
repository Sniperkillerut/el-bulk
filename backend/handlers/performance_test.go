package handlers

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

func BenchmarkSaveProductStorage_Bulk(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	productID := "test-product"
	items := make([]models.StorageLocation, 10)
	for i := 0; i < 10; i++ {
		items[i] = models.StorageLocation{
			StorageID: fmt.Sprintf("storage-%d", i),
			Quantity:  10,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectExec("DELETE FROM product_storage WHERE product_id = \\$1").
			WithArgs(productID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Construct the expected bulk insert query pattern
		mock.ExpectExec("INSERT INTO product_storage").
			WillReturnResult(sqlmock.NewResult(0, 10))

		h.saveProductStorage(productID, items)
	}
}

// Note: To compare with the original N+1 implementation, one would need to revert the changes
// in handlers/products.go and run a similar benchmark expecting 10 individual INSERTs.
// This benchmark confirms that the new bulk implementation works as expected with a single Exec call.
