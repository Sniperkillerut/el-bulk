package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_List_FiltersInactiveTCG(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	// 1. Total count query (buildFilters)
	mock.ExpectQuery("(?i)SELECT COUNT\\(\\*\\) FROM product p").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 2. Main list query (Note: sqlx expands p.* into explicit columns)
	mock.ExpectQuery("(?i)SELECT .* FROM product p").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "tcg"}).AddRow("1", "Black Lotus", "mtg"))

	// 3. Facets query (getFacets calls buildFilters 7 times for different dimensions)
	for i := 0; i < 7; i++ {
		mock.ExpectQuery("(?i)SELECT .* COUNT\\(\\*\\) FROM product p").
			WillReturnRows(sqlmock.NewRows([]string{"val", "count"}).AddRow("NM", 1))
	}

	// 4. Mock enrichment queries (populatePrices, populateStorage, populateCategories)
	mock.ExpectQuery("(?i)SELECT .* FROM product_prices").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("(?i)SELECT .* FROM product_storage").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("(?i)SELECT .* FROM product_category").WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name"}))

	req, _ := http.NewRequest("GET", "/api/products", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestProductHandler_GetByID_Public_FiltersInactiveTCG(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &ProductHandler{DB: sqlxDB}

	// Mocking a scenario where product is not found due to join/active-check
	mock.ExpectQuery("(?i)SELECT .* FROM product p").
		WithArgs("123").
		WillReturnError(http.ErrNoLocation)

	r := chi.NewRouter()
	r.Get("/api/products/{id}", h.GetByID)

	req, _ := http.NewRequest("GET", "/api/products/123", nil)
	rr := httptest.NewRecorder()
	
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
