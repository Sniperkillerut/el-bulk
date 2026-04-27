package handlers

import (
	"context"
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
	svc := testRefreshService(sqlxDB)

	// 1. Mock Scryfall Bulk Data Index
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bulk-data" {
			w.WriteHeader(http.StatusOK)
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
			fmt.Fprintf(w, `[
				{
					"name": "Black Lotus",
					"set": "lea",
					"prices": {
						"usd": "50000.00",
						"usd_foil": null,
						"eur": "45000.00"
					}
				}
			]`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oldBase := external.ScryfallBase
	external.ScryfallBase = server.URL
	defer func() { external.ScryfallBase = oldBase }()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "tcg", "name", "set_name", "set_code", "foil_treatment", "card_treatment", "price_source"}).
			AddRow("550e8400-e29b-41d4-a716-446655440000", "mtg", "Black Lotus", "Limited Edition Alpha", "lea", "non_foil", "normal", "tcgplayer")
		
		mock.ExpectQuery("(?i)SELECT id, tcg, name, set_name, set_code, foil_treatment, card_treatment, price_source FROM product").
			WillReturnRows(rows)

		// Mock SYNC calls after product list
		mock.ExpectExec("(?i)INSERT INTO external_scryfall").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("(?i)INSERT INTO external_cardkingdom").WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec("UPDATE product AS p SET").
			WithArgs("550e8400-e29b-41d4-a716-446655440000", 50000.0, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, errs := svc.RunPriceRefresh(context.Background(), "")

		assert.Equal(t, 1, updated)
		assert.Equal(t, 0, errs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRefreshHandler_Trigger(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	h := testRefreshHandler(sqlxDB)

	mock.ExpectQuery("(?i)SELECT id, tcg, name, set_name, set_code, foil_treatment, card_treatment, price_source FROM product").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_name", "set_code", "foil_treatment", "card_treatment", "price_source"}))

	// Mock SYNC calls after product list
	mock.ExpectExec("(?i)INSERT INTO external_scryfall").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?i)INSERT INTO external_cardkingdom").WillReturnResult(sqlmock.NewResult(0, 0))

	req, _ := http.NewRequest("POST", "/api/admin/prices/refresh", nil)
	rr := httptest.NewRecorder()
	h.Trigger(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var res map[string]int
	json.NewDecoder(rr.Body).Decode(&res)
	assert.Equal(t, 0, res["updated"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
