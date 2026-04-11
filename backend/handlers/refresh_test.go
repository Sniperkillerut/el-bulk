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

	oldBase := external.ScryfallBase
	external.ScryfallBase = server.URL
	defer func() { external.ScryfallBase = oldBase }()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "tcg", "name", "set_code", "foil_treatment", "price_source"}).
			AddRow("550e8400-e29b-41d4-a716-446655440000", "mtg", "Black Lotus", "lea", "non_foil", "tcgplayer").
			AddRow("550e8400-e29b-41d4-a716-446655440001", "mtg", "Mox Pearl", "lea", "non_foil", "tcgplayer")
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM product").
			WillReturnRows(rows)

		// Mock bulk update (chunked) - expectations must match the 7 fields updated per row
		mock.ExpectExec("UPDATE product AS p SET").
			WithArgs(
				"550e8400-e29b-41d4-a716-446655440000", 50000.0, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				"550e8400-e29b-41d4-a716-446655440001", 10000.0, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 2))

		updated, errs := svc.RunPriceRefresh()

		assert.Equal(t, 2, updated)
		assert.Equal(t, 0, errs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NoProducts", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM product").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_code", "foil_treatment", "price_source"}))

		updated, errs := svc.RunPriceRefresh()

		assert.Equal(t, 0, updated)
		assert.Equal(t, 0, errs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM product").
			WillReturnError(fmt.Errorf("db error"))

		updated, errs := svc.RunPriceRefresh()

		assert.Equal(t, 0, updated)
		assert.Equal(t, 1, errs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("HTTP Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM product").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_code", "foil_treatment", "price_source"}).AddRow("1", "mtg", "N1", "S1", "non_foil", "tcgplayer"))

		oldBase := external.ScryfallBase
		external.ScryfallBase = "http://invalid-url-123.com"
		defer func() { external.ScryfallBase = oldBase }()

		updated, errs := svc.RunPriceRefresh()
		assert.Equal(t, 0, updated)
		assert.Equal(t, 1, errs)
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

	mock.ExpectQuery("SELECT id, tcg, name, set_code, foil_treatment, price_source FROM product").
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
