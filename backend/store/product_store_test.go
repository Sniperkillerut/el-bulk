package store

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestProductStore_GetEnrichedByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewProductStore(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		productID := "p1"
		expectedJSON := `{"id":"p1","name":"Test Product","stock":10}`

		mock.ExpectQuery("(?i)SELECT fn_get_product_detail").
			WithArgs(productID).
			WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow([]byte(expectedJSON)))

		product, err := s.GetEnrichedByID(context.Background(), productID)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "p1", product.ID)
		assert.Equal(t, "Test Product", product.Name)
		assert.Equal(t, 10, product.Stock)
	})

	t.Run("Not Found", func(t *testing.T) {
		productID := "nonexistent"
		mock.ExpectQuery("(?i)SELECT fn_get_product_detail").
			WithArgs(productID).
			WillReturnError(sql.ErrNoRows)

		product, err := s.GetEnrichedByID(context.Background(), productID)
		assert.Error(t, err)
		assert.Nil(t, product)
	})
}

func TestProductStore_PopulateStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewProductStore(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		products := []models.Product{
			{ID: "p1"},
			{ID: "p2"},
		}

		rows := sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}).
			AddRow("p1", "s1", "Shelf A", 5).
			AddRow("p1", "s2", "Shelf B", 3).
			AddRow("p2", "s1", "Shelf A", 10)

		mock.ExpectQuery("(?i)SELECT ps.product_id, s.id as stored_in_id").
			WithArgs("p1", "p2").
			WillReturnRows(rows)

		err := s.PopulateStorage(context.Background(), products)
		assert.NoError(t, err)

		assert.Len(t, products[0].StoredIn, 2)
		assert.Equal(t, "Shelf A", products[0].StoredIn[0].Name)
		assert.Len(t, products[1].StoredIn, 1)
		assert.Equal(t, 10, products[1].StoredIn[0].Quantity)
	})
}

func TestProductStore_ListWithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewProductStore(sqlxDB)
	mock.MatchExpectationsInOrder(false)

	t.Run("Basic List", func(t *testing.T) {
		params := ProductFilterParams{Page: 1, PageSize: 20}
		
		// 1. Settings (if used in cache key or similar - wait, ListWithFilters doesn't call settings directly, Service does)
		
		// 2. Count
		mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		// 3. Main Query
		mock.ExpectQuery("(?i)SELECT p\\.\\* FROM product p").
			WithArgs(20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("p1", "Product 1"))

		// 4. Parallel Population (via SelectEnriched)
		mock.ExpectQuery("(?i)SELECT ps.product_id").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))
		mock.ExpectQuery("(?i)SELECT pc.product_id").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))
		mock.ExpectQuery("(?i)SELECT oi.product_id").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))
		mock.ExpectQuery("(?i)SELECT id, product_id").WillReturnRows(sqlmock.NewRows([]string{"id"}))

		products, total, err := s.ListWithFilters(context.Background(), params)
		assert.NoError(t, err)
		assert.Equal(t, 100, total)
		assert.Len(t, products, 1)
		assert.Equal(t, "p1", products[0].ID)
	})
}
func TestProductStore_UpdateProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewProductStore(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		id := "p1"
		input := models.ProductInput{
			Name: "Updated Product",
		}

		mock.ExpectQuery("(?i)UPDATE product SET").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(id, "Updated Product"))

		product, err := s.UpdateProduct(context.Background(), id, input)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "Updated Product", product.Name)
	})
}

func TestProductStore_BulkUpsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewProductStore(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		jsonData := `[{"name":"P1"}]`
		mock.ExpectQuery("(?i)SELECT upserted_id FROM fn_bulk_upsert_product").
			WithArgs(jsonData).
			WillReturnRows(sqlmock.NewRows([]string{"upserted_id"}).AddRow("p1"))

		ids, err := s.BulkUpsert(context.Background(), jsonData)
		assert.NoError(t, err)
		assert.Equal(t, []string{"p1"}, ids)
	})
}

func TestProductStore_GetHotProductIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewProductStore(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("(?i)SELECT product_id FROM order_item").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"product_id"}).AddRow("p1").AddRow("p2"))

		ids, err := s.GetHotProductIDs(context.Background(), 7, 5, []string{"p1", "p2"})
		assert.NoError(t, err)
		assert.Len(t, ids, 2)
		assert.Contains(t, ids, "p1")
		assert.Contains(t, ids, "p2")
	})
}

func TestProductStore_CreateProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewProductStore(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		input := models.ProductInput{
			Name: "New Product",
			TCG: "mtg",
			Category: "singles",
		}

		mock.ExpectQuery("(?i)INSERT INTO product").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("new-p", "New Product"))

		product, err := s.CreateProduct(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "new-p", product.ID)
	})
}
