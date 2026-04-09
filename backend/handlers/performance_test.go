package handlers

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
)

func BenchmarkSaveProductStorage_Bulk(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	s := store.NewProductStore(sqlxDB)

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

		mock.ExpectExec("INSERT INTO product_storage").
			WillReturnResult(sqlmock.NewResult(0, 10))

		s.SaveStorage(productID, items)
	}
}
