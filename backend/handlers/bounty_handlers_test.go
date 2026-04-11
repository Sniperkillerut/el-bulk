package handlers

import (
	"bytes"
	"encoding/json"
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

func TestBountyHandler_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "tcg", "is_active", "created_at", "updated_at"}).
			AddRow("b1", "Black Lotus", "mtg", true, time.Now(), time.Now())
		mock.ExpectQuery("SELECT .* FROM bounty").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/bounties", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestBountyHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		input := models.BountyInput{
			Name: "Black Lotus",
			TCG:  "mtg",
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("INSERT INTO bounty").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "tcg"}).
				AddRow("b1", "Black Lotus", "mtg"))

		req, _ := http.NewRequest("POST", "/api/admin/bounties", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		input := models.BountyInput{Name: ""} // Missing TCG and Name
		body, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", "/api/admin/bounties", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestBountyHandler_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		input := models.BountyInput{Name: "Updated Lotus"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("UPDATE bounty").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow("b1", "Updated Lotus"))

		r := chi.NewRouter()
		r.Put("/api/admin/bounties/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/bounties/b1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestBountyHandler_SubmitOffer(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		input := models.BountyOfferInput{
			BountyID:        "b1",
			CustomerName:    "John Doe",
			CustomerContact: "12345",
			Quantity:        1,
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("SELECT fn_submit_bounty_offer").
			WillReturnRows(sqlmock.NewRows([]string{"fn_submit_bounty_offer"}).
				AddRow([]byte(`{"id":"o1","bounty_id":"b1","customer_id":"c1","quantity":1}`)))

		req, _ := http.NewRequest("POST", "/api/bounties/offer", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.SubmitOffer(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestBountyHandler_CreateRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		input := models.ClientRequestInput{
			CustomerName:    "John Doe",
			CustomerContact: "12345",
			CardName:        "Black Lotus",
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("SELECT fn_submit_client_request").
			WillReturnRows(sqlmock.NewRows([]string{"fn_submit_client_request"}).
				AddRow([]byte(`{"id":"r1","customer_name":"John Doe","card_name":"Black Lotus"}`)))

		req, _ := http.NewRequest("POST", "/api/wanted/request", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.CreateRequest(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestBountyHandler_ListOffers(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "bounty_id", "customer_id", "customer_contact"}).
		AddRow("o1", "b1", "c1", "12345")
	mock.ExpectQuery("SELECT .* FROM bounty_offer").WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/admin/bounties/offers", nil)
	rr := httptest.NewRecorder()
	h.ListOffers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBountyHandler_UpdateOfferStatus(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		input := models.UpdateBountyOfferStatusInput{Status: "accepted"}
		body, _ := json.Marshal(input)

		mock.ExpectExec("UPDATE bounty_offer").WithArgs("accepted", "o1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT .* FROM bounty_offer").WithArgs("o1").WillReturnRows(sqlmock.NewRows([]string{"id", "status", "customer_contact"}).AddRow("o1", "accepted", "12345"))

		r := chi.NewRouter()
		r.Put("/api/admin/bounties/offers/{id}/status", h.UpdateOfferStatus)
		req, _ := http.NewRequest("PUT", "/api/admin/bounties/offers/o1/status", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestBountyHandler_ListRequests(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "customer_name", "customer_contact", "card_name"}).
		AddRow("r1", "John Doe", "12345", "Black Lotus")
	mock.ExpectQuery("SELECT id, customer_id, customer_name, customer_contact, card_name, set_name, details, status, created_at FROM client_request").WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/admin/wanted/requests", nil)
	rr := httptest.NewRecorder()
	h.ListRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBountyHandler_UpdateRequestStatus(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		input := models.UpdateClientRequestStatusInput{Status: "solved"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("UPDATE client_request").WithArgs("solved", "r1").WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("r1", "solved"))

		r := chi.NewRouter()
		r.Put("/api/admin/wanted/requests/{id}/status", h.UpdateRequestStatus)
		req, _ := http.NewRequest("PUT", "/api/admin/wanted/requests/r1/status", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestBountyHandler_Delete(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	mock.ExpectExec("DELETE FROM bounty").WithArgs("b1").WillReturnResult(sqlmock.NewResult(1, 1))

	r := chi.NewRouter()
	r.Delete("/api/admin/bounties/{id}", h.Delete)
	req, _ := http.NewRequest("DELETE", "/api/admin/bounties/b1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
