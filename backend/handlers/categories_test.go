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

func TestCategoriesHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &CategoriesHandler{DB: sqlxDB}

	t.Run("Admin List", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "name", "slug", "is_active", "show_badge", "searchable", "created_at", "item_count"}).
			AddRow("c1", "Cat 1", "cat-1", true, true, true, now, 5)

		mock.ExpectQuery("(?i)SELECT .* FROM custom_category c.*LEFT JOIN product_categories pc").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/admin/categories", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var res []models.CustomCategory
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Len(t, res, 1)
	})
	t.Run("Public List", func(t *testing.T) {
		now := time.Now()
		mock.ExpectQuery("(?i)SELECT .* FROM custom_category c.*LEFT JOIN product_categories pc.*WHERE .* HAVING COUNT.*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "is_active", "show_badge", "searchable", "created_at", "item_count"}).
				AddRow("c1", "Cat 1", "cat-1", true, true, true, now, 5))

		req, _ := http.NewRequest("GET", "/api/categories", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestCategoriesHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &CategoriesHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.CustomCategoryInput{Name: "New Category"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("INSERT INTO custom_categories").
			WithArgs("New Category", "new-category", true, true, true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug"}).AddRow("c2", "New Category", "new-category"))

		req, _ := http.NewRequest("POST", "/api/admin/categories", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		input := models.CustomCategoryInput{Name: "New Category"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("INSERT INTO custom_categories").WillReturnError(fmt.Errorf("db error"))

		req, _ := http.NewRequest("POST", "/api/admin/categories", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestCategoriesHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &CategoriesHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		input := models.CustomCategoryInput{Name: "Updated Cat"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("UPDATE custom_categories SET name = \\$1, slug = \\$2 WHERE id = \\$3").
			WithArgs("Updated Cat", "updated-cat", "c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "Updated Cat"))

		r := chi.NewRouter()
		r.Put("/api/admin/categories/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/categories/c1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DB General Error", func(t *testing.T) {
		input := models.CustomCategoryInput{Name: "Unknown"}
		body, _ := json.Marshal(input)
		mock.ExpectQuery("UPDATE custom_categories").WillReturnError(fmt.Errorf("db error"))

		r := chi.NewRouter()
		r.Put("/api/admin/categories/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/categories/unknown", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestCategoriesHandler_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := &CategoriesHandler{DB: sqlxDB}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM custom_category").
			WithArgs("c1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := chi.NewRouter()
		r.Delete("/api/admin/categories/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/categories/c1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM custom_category").WillReturnError(fmt.Errorf("db error"))
		r := chi.NewRouter()
		r.Delete("/api/admin/categories/{id}", h.Delete)
		req, _ := http.NewRequest("DELETE", "/api/admin/categories/c1", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
