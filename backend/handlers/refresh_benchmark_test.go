package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func BenchmarkUpdatePrices_NPlus1(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	numProducts := 100
	ids := make([]string, numProducts)
	prices := make([]float64, numProducts)
	for i := 0; i < numProducts; i++ {
		ids[i] = fmt.Sprintf("prod-%d", i)
		prices[i] = float64(i) + 0.99
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < numProducts; j++ {
			mock.ExpectExec("UPDATE product SET price_reference=\\$1 WHERE id=\\$2").
				WithArgs(prices[j], ids[j]).
				WillReturnResult(sqlmock.NewResult(0, 1))

			_, _ = sqlxDB.Exec("UPDATE product SET price_reference=$1 WHERE id=$2", prices[j], ids[j])
		}
	}
}

func BenchmarkUpdatePrices_Bulk(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	numProducts := 100
	ids := make([]string, numProducts)
	prices := make([]float64, numProducts)
	for i := 0; i < numProducts; i++ {
		ids[i] = fmt.Sprintf("prod-%d", i)
		prices[i] = float64(i) + 0.99
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := "UPDATE product AS p SET price_reference = v.price_reference FROM (VALUES "
		placeholders := make([]string, numProducts)
		args := make([]interface{}, numProducts*2)
		for j := 0; j < numProducts; j++ {
			placeholders[j] = fmt.Sprintf("($%d::uuid, $%d::double precision)", j*2+1, j*2+2)
			args[j*2] = ids[j]
			args[j*2+1] = prices[j]
		}
		query += strings.Join(placeholders, ", ")
		query += ") AS v(id, price_reference) WHERE p.id = v.id"

		mock.ExpectExec("UPDATE product AS p SET price_reference = v.price_reference").
			WillReturnResult(sqlmock.NewResult(0, int64(numProducts)))

		_, _ = sqlxDB.Exec(query, args...)
	}
}
