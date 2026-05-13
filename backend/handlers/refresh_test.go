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
	external.ResetScryfallCache()
	external.ResetCardKingdomCache()

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
		rows := sqlmock.NewRows([]string{"id", "tcg", "name", "set_name", "set_code", "collector_number", "foil_treatment", "card_treatment", "price_source", "scryfall_id", "ck_set_name"}).
			AddRow("550e8400-e29b-41d4-a716-446655440000", "mtg", "Black Lotus", "Limited Edition Alpha", "lea", "1", "non_foil", "normal", "tcgplayer", "std_id", "")
		
		mock.ExpectQuery("(?i)SELECT .* FROM product p .*").
			WillReturnRows(rows)

		// Mock SYNC calls after product list
		mock.ExpectExec("(?i)TRUNCATE TABLE external_scryfall").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("(?i)INSERT INTO external_scryfall").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("(?i)SELECT .* FROM external_scryfall").
			WillReturnRows(sqlmock.NewRows([]string{"scryfall_id", "name", "set_code", "collector_number", "price_usd", "price_usd_foil", "price_eur", "image_url"}).
				AddRow("std_id", "Black Lotus", "lea", "1", 50000.0, nil, 45000.0, ""))

		mock.ExpectQuery("(?i)SELECT .* FROM setting").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("usd_to_cop_rate", "4000").
				AddRow("eur_to_cop_rate", "4500").
				AddRow("ck_to_cop_rate", "3800"))

		mock.ExpectExec("(?i)UPDATE product AS p SET.*").
			WithArgs("550e8400-e29b-41d4-a716-446655440000", 50000.0, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 3800.0, 4000.0, 4500.0).
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

	// Mock Scryfall API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/bulk-data" {
			fmt.Fprintf(w, `{"data": [{"type": "default_cards", "download_uri": "%s/download"}]}`, "http://"+r.Host)
		} else {
			fmt.Fprintf(w, `[{"name": "Black Lotus", "set": "lea", "prices": {"usd": "50000.00"}}]`)
		}
	}))
	defer server.Close()
	oldBase := external.ScryfallBase
	external.ScryfallBase = server.URL
	defer func() { external.ScryfallBase = oldBase }()

	h := testRefreshHandler(sqlxDB)
	external.ResetScryfallCache()
	external.ResetCardKingdomCache()

	mock.ExpectQuery("(?i)SELECT .* FROM product p .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_name", "set_code", "collector_number", "foil_treatment", "card_treatment", "price_source", "scryfall_id", "ck_set_name"}).
			AddRow("550e8400-e29b-41d4-a716-446655440000", "mtg", "Black Lotus", "Limited Edition Alpha", "lea", "1", "non_foil", "normal", "tcgplayer", "std_id", ""))

	// Mock SYNC calls after product list
	mock.ExpectExec("(?i)TRUNCATE TABLE external_scryfall").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?i)INSERT INTO external_scryfall").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?i)SELECT .* FROM external_scryfall").
		WillReturnRows(sqlmock.NewRows([]string{"scryfall_id", "name", "set_code", "collector_number", "price_usd", "price_usd_foil", "price_eur", "image_url"}))

	mock.ExpectQuery("(?i)SELECT .* FROM setting").
		WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("usd_to_cop_rate", "4000").
			AddRow("eur_to_cop_rate", "4500").
			AddRow("ck_to_cop_rate", "3800"))

	req, _ := http.NewRequest("POST", "/api/admin/prices/refresh", nil)
	rr := httptest.NewRecorder()
	h.Trigger(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var res map[string]int
	json.NewDecoder(rr.Body).Decode(&res)
	assert.Equal(t, 0, res["updated"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
