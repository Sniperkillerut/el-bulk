package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/external"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestRefreshHandler_RunPriceRefresh(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// 1. Mock Scryfall Bulk Data Index
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bulk-data" {
			w.WriteHeader(http.StatusOK)
			// Construct a valid scryfallBulkMeta JSON
			fmt.Fprintf(w, `{
				"data": [
					{
						"type": "default_cards",
						"download_uri": "%s/download/default_cards.json",
						"updated_at": "2023-10-27T11:00:00Z"
					}
				]
			}`, "http://"+r.Host)
			return
		}
		if r.URL.Path == "/download/default_cards.json" {
			w.WriteHeader(http.StatusOK)
			// Stream a JSON array of cards
			fmt.Fprintf(w, `[
				{
					"name": "Black Lotus",
					"set": "lea",
					"prices": {
						"usd": "50000.00",
						"usd_foil": null,
						"eur": "45000.00"
					}
				},
				{
					"name": "Mox Pearl",
					"set": "lea",
					"prices": {
						"usd": "10000.00",
						"usd_foil": null,
						"eur": "9000.00"
					}
				}
			]`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Override ScryfallBase
	oldBase := external.ScryfallBase
	external.ScryfallBase = server.URL
	defer func() { external.ScryfallBase = oldBase }()

	t.Run("Success", func(t *testing.T) {
		// Mock initial query for products needing refresh
		rows := sqlmock.NewRows([]string{"id", "tcg", "name", "set_code", "foil_treatment", "price_source"}).
			AddRow("1", "mtg", "Black Lotus", "lea", "non_foil", "tcgplayer").
			AddRow("2", "mtg", "Mox Pearl", "lea", "non_foil", "tcgplayer")
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM products").
			WillReturnRows(rows)

		// Mock updates
		mock.ExpectExec("UPDATE products SET price_reference=\\$1 WHERE id=\\$2").
			WithArgs(50000.0, "1").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("UPDATE products SET price_reference=\\$1 WHERE id=\\$2").
			WithArgs(10000.0, "2").
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, errs := RunPriceRefresh(sqlxDB)

		assert.Equal(t, 2, updated)
		assert.Equal(t, 0, errs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NoProducts", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM products").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_code", "foil_treatment", "price_source"}))

		updated, errs := RunPriceRefresh(sqlxDB)

		assert.Equal(t, 0, updated)
		assert.Equal(t, 0, errs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM products").
			WillReturnError(fmt.Errorf("db error"))

		updated, errs := RunPriceRefresh(sqlxDB)

		assert.Equal(t, 0, updated)
		assert.Equal(t, 1, errs)
	})

	t.Run("HTTP Error", func(t *testing.T) {
		// Mock query success but HTTP fail
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM products").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_code", "foil_treatment", "price_source"}).AddRow("1", "mtg", "N1", "S1", "non_foil", "tcgplayer"))
		
		oldBase := external.ScryfallBase
		external.ScryfallBase = "http://invalid-url-123.com"
		defer func() { external.ScryfallBase = oldBase }()

		updated, errs := RunPriceRefresh(sqlxDB)
		assert.Equal(t, 0, updated)
		assert.Equal(t, 1, errs)
	})
}

func TestRefreshHandler_Trigger(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	h := NewRefreshHandler(sqlxDB)

	// Since RunPriceRefresh calls ScryfallBase, we need to mock it or it will fail
	// Or we can just mock the empty product list to avoid the network call
	mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM products").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_code", "foil_treatment", "price_source"}))

	req, _ := http.NewRequest("POST", "/api/admin/prices/refresh", nil)
	rr := httptest.NewRecorder()
	h.Trigger(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var res map[string]int
	json.NewDecoder(rr.Body).Decode(&res)
	assert.Equal(t, 0, res["updated"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
