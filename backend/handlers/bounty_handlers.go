package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
)

type BountyHandler struct {
	DB *sqlx.DB
}

func NewBountyHandler(db *sqlx.DB) *BountyHandler {
	return &BountyHandler{DB: db}
}

// === BOUNTIES ===

func (h *BountyHandler) List(w http.ResponseWriter, r *http.Request) {
	var bounties []models.Bounty
	query := `
		SELECT 
			id, name, tcg, set_name, condition, foil_treatment, target_price, 
			hide_price, quantity_needed, image_url, created_at, updated_at
		FROM bounty
		ORDER BY created_at DESC
	`
	err := h.DB.Select(&bounties, query)
	if err != nil {
		logger.Error("Failed to list bounties: %v", err)
		jsonError(w, "Failed to fetch bounties", http.StatusInternalServerError)
		return
	}
	// Default to empty array instead of null
	if bounties == nil {
		bounties = []models.Bounty{}
	}

	jsonOK(w, bounties)
}

func (h *BountyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.BountyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.TCG == "" {
		jsonError(w, "name and tcg are required", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO bounty (name, tcg, set_name, condition, foil_treatment, target_price, hide_price, quantity_needed, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, tcg, set_name, condition, foil_treatment, target_price, hide_price, quantity_needed, image_url, created_at, updated_at
	`
	var bounty models.Bounty
	err := h.DB.QueryRowx(query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.ImageURL,
	).StructScan(&bounty)

	if err != nil {
		logger.Error("Failed to create bounty: %v", err)
		jsonError(w, "Failed to create bounty", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, bounty)
}

func (h *BountyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.BountyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE bounty
		SET name = $1, tcg = $2, set_name = $3, condition = $4, foil_treatment = $5,
		    target_price = $6, hide_price = $7, quantity_needed = $8, image_url = $9, updated_at = now()
		WHERE id = $10
		RETURNING id, name, tcg, set_name, condition, foil_treatment, target_price, hide_price, quantity_needed, image_url, created_at, updated_at
	`
	var bounty models.Bounty
	err := h.DB.QueryRowx(query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.ImageURL, id,
	).StructScan(&bounty)

	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, "Bounty not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to update bounty: %v", err)
		jsonError(w, "Failed to update bounty", http.StatusInternalServerError)
		return
	}

	jsonOK(w, bounty)
}

func (h *BountyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.DB.Exec(`DELETE FROM bounty WHERE id = $1`, id)
	if err != nil {
		logger.Error("Failed to delete bounty: %v", err)
		jsonError(w, "Failed to delete bounty", http.StatusInternalServerError)
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		jsonError(w, "Bounty not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === CLIENT REQUESTS ===

func (h *BountyHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	var requests []models.ClientRequest
	query := `
		SELECT id, customer_name, customer_contact, card_name, set_name, details, status, created_at
		FROM client_request
		ORDER BY created_at DESC
	`
	err := h.DB.Select(&requests, query)
	if err != nil {
		logger.Error("Failed to list client requests: %v", err)
		jsonError(w, "Failed to fetch client requests", http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []models.ClientRequest{}
	}

	jsonOK(w, requests)
}

func (h *BountyHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var input models.ClientRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.CustomerName == "" || input.CustomerContact == "" || input.CardName == "" {
		jsonError(w, "customer_name, customer_contact, and card_name are required", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO client_request (customer_name, customer_contact, card_name, set_name, details, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, customer_name, customer_contact, card_name, set_name, details, status, created_at
	`
	var req models.ClientRequest
	err := h.DB.QueryRowx(query,
		input.CustomerName, input.CustomerContact, input.CardName, input.SetName, input.Details,
	).StructScan(&req)

	if err != nil {
		logger.Error("Failed to create client request: %v", err)
		jsonError(w, "Failed to submit request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, req)
}

func (h *BountyHandler) UpdateRequestStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.UpdateClientRequestStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE client_request
		SET status = $1
		WHERE id = $2
		RETURNING id, customer_name, customer_contact, card_name, set_name, details, status, created_at
	`
	var req models.ClientRequest
	err := h.DB.QueryRowx(query, input.Status, id).StructScan(&req)

	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, "Request not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to update client request status: %v", err)
		jsonError(w, "Failed to update request status", http.StatusInternalServerError)
		return
	}

	jsonOK(w, req)
}
