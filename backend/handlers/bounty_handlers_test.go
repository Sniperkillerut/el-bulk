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
			AddRow("550e8400-e29b-41d4-a716-446655440030", "Black Lotus", "550e8400-e29b-41d4-a716-446655440003", true, time.Now(), time.Now())
		mock.ExpectQuery("SELECT .* FROM bounty").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/bounties", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Sanitization", func(t *testing.T) {
		now := time.Now()
		targetPrice := 100.0
		priceReference := 90.0
		rows := sqlmock.NewRows([]string{"id", "name", "tcg", "target_price", "hide_price", "price_source", "price_reference", "is_active", "created_at", "updated_at"}).
			AddRow("550e8400-e29b-41d4-a716-446655440030", "Black Lotus", "550e8400-e29b-41d4-a716-446655440003", &targetPrice, true, "source", &priceReference, true, &now, &now)
		mock.ExpectQuery("SELECT .* FROM bounty").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/bounties", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp []map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)

		b := resp[0]
		assert.NotContains(t, b, "price_source")
		assert.NotContains(t, b, "price_reference")
		assert.NotContains(t, b, "created_at")
		assert.NotContains(t, b, "updated_at")
		assert.NotContains(t, b, "target_price") // Since hide_price is true
	})
}

func TestBountyHandler_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		tcgID := "550e8400-e29b-41d4-a716-446655440003"
		bountyID := "550e8400-e29b-41d4-a716-446655440030"
		input := models.BountyInput{
			Name: "Black Lotus",
			TCG:  tcgID,
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("INSERT INTO bounty").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "tcg"}).
				AddRow(bountyID, "Black Lotus", tcgID))

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
		bountyID := "550e8400-e29b-41d4-a716-446655440030"
		input := models.BountyInput{Name: "Updated Lotus"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("UPDATE bounty").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(bountyID, "Updated Lotus"))

		r := chi.NewRouter()
		r.Put("/api/admin/bounties/{id}", h.Update)
		req, _ := http.NewRequest("PUT", "/api/admin/bounties/"+bountyID, bytes.NewBuffer(body))
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
		bountyID := "550e8400-e29b-41d4-a716-446655440030"
		offerID := "550e8400-e29b-41d4-a716-446655440040"
		customerID := "550e8400-e29b-41d4-a716-446655440010"
		input := models.BountyOfferInput{
			BountyID:        bountyID,
			CustomerName:    "John Doe",
			CustomerContact: "12345",
			Quantity:        1,
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("SELECT fn_submit_bounty_offer").
			WillReturnRows(sqlmock.NewRows([]string{"fn_submit_bounty_offer"}).
				AddRow([]byte(`{"id":"` + offerID + `","bounty_id":"` + bountyID + `","customer_id":"` + customerID + `","quantity":1}`)))

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
		requestID := "550e8400-e29b-41d4-a716-446655440050"
		input := models.ClientRequestInput{
			CustomerName:    "John Doe",
			CustomerContact: "12345",
			CardName:        "Black Lotus",
		}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("SELECT fn_submit_client_request").
			WillReturnRows(sqlmock.NewRows([]string{"fn_submit_client_request"}).
				AddRow([]byte(`{"id":"` + requestID + `","customer_name":"John Doe","card_name":"Black Lotus"}`)))

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
		AddRow("550e8400-e29b-41d4-a716-446655440040", "550e8400-e29b-41d4-a716-446655440030", "550e8400-e29b-41d4-a716-446655440010", "12345")
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
		offerID := "550e8400-e29b-41d4-a716-446655440040"
		input := models.UpdateBountyOfferStatusInput{Status: "accepted"}
		body, _ := json.Marshal(input)

		mock.ExpectExec("UPDATE bounty_offer").WithArgs("accepted", offerID).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT .* FROM bounty_offer").WithArgs(offerID).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "customer_contact"}).AddRow(offerID, "accepted", "12345"))

		r := chi.NewRouter()
		r.Put("/api/admin/bounties/offers/{id}/status", h.UpdateOfferStatus)
		req, _ := http.NewRequest("PUT", "/api/admin/bounties/offers/"+offerID+"/status", bytes.NewBuffer(body))
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
		AddRow("550e8400-e29b-41d4-a716-446655440050", "John Doe", "12345", "Black Lotus")
	mock.ExpectQuery("SELECT .* FROM client_request").WillReturnRows(rows)

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
		requestID := "550e8400-e29b-41d4-a716-446655440050"
		input := models.UpdateClientRequestStatusInput{Status: "solved"}
		body, _ := json.Marshal(input)

		mock.ExpectQuery("UPDATE client_request").WithArgs("solved", requestID).WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(requestID, "solved"))

		r := chi.NewRouter()
		r.Put("/api/admin/wanted/requests/{id}/status", h.UpdateRequestStatus)
		req, _ := http.NewRequest("PUT", "/api/admin/wanted/requests/"+requestID+"/status", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestBountyHandler_Delete(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := testBountyHandler(sqlxDB)

	bountyID := "550e8400-e29b-41d4-a716-446655440030"
	mock.ExpectExec("DELETE FROM bounty").WithArgs(bountyID).WillReturnResult(sqlmock.NewResult(1, 1))

	r := chi.NewRouter()
	r.Delete("/api/admin/bounties/{id}", h.Delete)
	req, _ := http.NewRequest("DELETE", "/api/admin/bounties/"+bountyID, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
