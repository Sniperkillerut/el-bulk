package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/crypto"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type BountyHandler struct {
	DB *sqlx.DB
}

func NewBountyHandler(db *sqlx.DB) *BountyHandler {
	return &BountyHandler{DB: db}
}

// === BOUNTIES ===

func (h *BountyHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.List | URL: %s", r.URL.String())
	var bounties []models.Bounty
	query := `
		SELECT 
			id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, 
			hide_price, quantity_needed, image_url, price_source, price_reference, is_active, created_at, updated_at
		FROM bounty
		WHERE ($1 = '' OR is_active = ($1 = 'true'))
		ORDER BY created_at DESC
	`
	activeParam := r.URL.Query().Get("active")
	err := h.DB.Select(&bounties, query, activeParam)
	if err != nil {
		logger.Error("Failed to list bounties: %v", err)
		render.Error(w, "Failed to fetch bounties", http.StatusInternalServerError)
		return
	}
	// Default to empty array instead of null
	if bounties == nil {
		bounties = []models.Bounty{}
	}

	render.Success(w, bounties)
}

func (h *BountyHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.Create")
	var input models.BountyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Failed to decode bounty input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.TCG == "" {
		render.Error(w, "name and tcg are required", http.StatusBadRequest)
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	query := `
		INSERT INTO bounty (name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, image_url, price_source, price_reference, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, image_url, price_source, price_reference, is_active, created_at, updated_at
	`
	var bounty models.Bounty
	err := h.DB.QueryRowx(query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.CardTreatment, input.CollectorNumber, input.PromoType, input.Language,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.ImageURL,
		input.PriceSource, input.PriceReference, isActive,
	).StructScan(&bounty)

	if err != nil {
		logger.Error("Failed to create bounty: %v", err)
		render.Error(w, "Failed to create bounty", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, bounty)
}

func (h *BountyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.Trace("Entering BountyHandler.Update | ID: %s", id)
	var input models.BountyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Failed to decode bounty update for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	query := `
		UPDATE bounty
		SET name = $1, tcg = $2, set_name = $3, condition = $4, foil_treatment = $5,
		    card_treatment = $6, collector_number = $7, promo_type = $8, language = $9,
		    target_price = $10, hide_price = $11, quantity_needed = $12, image_url = $13,
		    price_source = $14, price_reference = $15, is_active = $16, updated_at = now()
		WHERE id = $17
		RETURNING id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, image_url, price_source, price_reference, is_active, created_at, updated_at
	`
	var bounty models.Bounty
	err := h.DB.QueryRowx(query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.CardTreatment, input.CollectorNumber, input.PromoType, input.Language,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.ImageURL,
		input.PriceSource, input.PriceReference, isActive, id,
	).StructScan(&bounty)

	if err != nil {
		if err == sql.ErrNoRows {
			render.Error(w, "Bounty not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to update bounty: %v", err)
		render.Error(w, "Failed to update bounty", http.StatusInternalServerError)
		return
	}

	render.Success(w, bounty)
}

func (h *BountyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.Trace("Entering BountyHandler.Delete | ID: %s", id)
	res, err := h.DB.Exec(`DELETE FROM bounty WHERE id = $1`, id)
	if err != nil {
		logger.Error("Failed to delete bounty: %v", err)
		render.Error(w, "Failed to delete bounty", http.StatusInternalServerError)
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		render.Error(w, "Bounty not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === BOUNTY OFFERS ===

func (h *BountyHandler) ListOffers(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.ListOffers")
	var offers []models.BountyOffer
	query := `
		SELECT 
			o.id, o.bounty_id, o.customer_id, o.quantity, o.condition, o.status, o.notes, o.admin_notes, o.created_at, o.updated_at,
			b.name as bounty_name,
			c.first_name || ' ' || COALESCE(c.last_name, '') as customer_name,
			COALESCE(c.phone, c.email) as customer_contact
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		JOIN customer c ON o.customer_id = c.id
		ORDER BY o.created_at DESC
	`
	err := h.DB.Select(&offers, query)
	if err != nil {
		logger.Error("Failed to list bounty offers: %v", err)
		render.Error(w, "Failed to fetch bounty offers", http.StatusInternalServerError)
		return
	}
	if offers == nil {
		offers = []models.BountyOffer{}
	}

	// Decrypt sensitive contact info for admin view (if it was encrypted)
	for i := range offers {
		offers[i].CustomerContact = *crypto.DecryptSafe(&offers[i].CustomerContact)
	}

	render.Success(w, offers)
}

func (h *BountyHandler) SubmitOffer(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.SubmitOffer")
	var input models.BountyOfferInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Failed to decode offer input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.BountyID == "" || input.CustomerName == "" || input.CustomerContact == "" {
		render.Error(w, "BountyID, customer_name, and customer_contact are required", http.StatusBadRequest)
		return
	}

	if input.Quantity <= 0 {
		input.Quantity = 1
	}

	// Try to get authenticated user ID from context
	var userID *string
	if val, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
		userID = &val
	}

	var result []byte
	err := h.DB.Get(&result, "SELECT fn_submit_bounty_offer($1, $2, $3, $4, $5, $6, $7, $8)",
		input.BountyID, input.CustomerName, input.CustomerContact, input.Quantity, input.Condition, input.Notes, input.Status, userID)

	if err != nil {
		logger.Error("Failed to call fn_submit_bounty_offer: %v", err)
		render.Error(w, "Failed to submit offer", http.StatusInternalServerError)
		return
	}

	var offer models.BountyOffer
	if err := json.Unmarshal(result, &offer); err != nil {
		logger.Error("Failed to unmarshal bounty offer: %v", err)
		render.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, offer)
}

func (h *BountyHandler) UpdateOfferStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.Trace("Entering BountyHandler.UpdateOfferStatus | ID: %s", id)
	var input models.UpdateBountyOfferStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Failed to decode offer status update for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE bounty_offer
		SET status = $1, updated_at = now()
		WHERE id = $2
	`
	_, err := h.DB.Exec(query, input.Status, id)
	if err != nil {
		logger.Error("Failed to update offer status: %v", err)
		render.Error(w, "Failed to update offer status", http.StatusInternalServerError)
		return
	}

	var offer models.BountyOffer
	selectQuery := `
		SELECT 
			o.id, o.bounty_id, o.customer_id, o.quantity, o.condition, o.status, o.notes, o.admin_notes, o.created_at, o.updated_at,
			b.name as bounty_name,
			c.first_name || ' ' || COALESCE(c.last_name, '') as customer_name,
			COALESCE(c.phone, c.email) as customer_contact
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		JOIN customer c ON o.customer_id = c.id
		WHERE o.id = $1
	`
	err = h.DB.Get(&offer, selectQuery, id)

	if err != nil {
		logger.Error("Failed to fetch updated offer: %v", err)
		render.Error(w, "Failed to fetch updated offer", http.StatusInternalServerError)
		return
	}

	// Decrypt sensitive contact info
	offer.CustomerContact = *crypto.DecryptSafe(&offer.CustomerContact)

	render.Success(w, offer)
}

// === CLIENT REQUESTS ===

func (h *BountyHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.ListRequests")
	var requests []models.ClientRequest
	query := `
		SELECT id, customer_id, customer_name, customer_contact, card_name, set_name, details, status, created_at
		FROM client_request
		ORDER BY created_at DESC
	`
	err := h.DB.Select(&requests, query)
	if err != nil {
		logger.Error("Failed to list client requests: %v", err)
		render.Error(w, "Failed to fetch client requests", http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []models.ClientRequest{}
	}

	// Decrypt sensitive contact info for admin view
	for i := range requests {
		requests[i].CustomerContact = *crypto.DecryptSafe(&requests[i].CustomerContact)
	}

	render.Success(w, requests)
}

func (h *BountyHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.CreateRequest")
	var input models.ClientRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Failed to decode client request input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.CustomerName == "" || input.CustomerContact == "" || input.CardName == "" {
		render.Error(w, "customer_name, customer_contact, and card_name are required", http.StatusBadRequest)
		return
	}

	// Optional user ID from context
	var userID *string
	if val, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
		userID = &val
	}

	var result []byte
	err := h.DB.Get(&result, "SELECT fn_submit_client_request($1, $2, $3, $4, $5, $6)",
		input.CustomerName, input.CustomerContact, input.CardName, input.SetName, input.Details, userID)

	if err != nil {
		logger.Error("Failed to call fn_submit_client_request: %v", err)
		render.Error(w, "Failed to submit request", http.StatusInternalServerError)
		return
	}

	var req models.ClientRequest
	if err := json.Unmarshal(result, &req); err != nil {
		logger.Error("Failed to unmarshal client request: %v", err)
		render.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, req)
}

func (h *BountyHandler) UpdateRequestStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.Trace("Entering BountyHandler.UpdateRequestStatus | ID: %s", id)
	var input models.UpdateClientRequestStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Failed to decode client request status update for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
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
			render.Error(w, "Request not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to update client request status: %v", err)
		render.Error(w, "Failed to update request status", http.StatusInternalServerError)
		return
	}

	render.Success(w, req)
}

// GET /api/bounties/offers/me — list bounty offers for the current customer
func (h *BountyHandler) ListMeOffers(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.ListMeOffers")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	var offers []models.BountyOffer
	query := `
		SELECT 
			o.id, o.bounty_id, o.customer_id, o.quantity, o.condition, o.status, o.notes, o.admin_notes, o.created_at, o.updated_at,
			b.name as bounty_name
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		WHERE o.customer_id = $1
		ORDER BY o.created_at DESC
	`
	err := h.DB.Select(&offers, query, userID)
	if err != nil {
		logger.Error("Failed to list user bounty offers for %s: %v", userID, err)
		render.Error(w, "Failed to fetch bounty offers", http.StatusInternalServerError)
		return
	}
	if offers == nil {
		offers = []models.BountyOffer{}
	}

	render.Success(w, offers)
}

// DELETE /api/bounties/offers/me/{id} — cancel a pending bounty offer
func (h *BountyHandler) CancelMeOffer(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.CancelMeOffer")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	res, err := h.DB.Exec(`
		UPDATE bounty_offer 
		SET status = 'cancelled', updated_at = now() 
		WHERE id = $1 AND customer_id = $2 AND status = 'pending'
	`, id, userID)

	if err != nil {
		logger.Error("Failed to cancel user bounty offer %s for user %s: %v", id, userID, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		render.Error(w, "Offer cannot be cancelled. It may not exist, belong to you, or is already being processed.", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/client-requests/me — list requests for the current customer
func (h *BountyHandler) ListMeRequests(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.ListMeRequests")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	var requests []models.ClientRequest
	query := `
		SELECT id, customer_id, customer_name, customer_contact, card_name, set_name, details, status, created_at
		FROM client_request
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	err := h.DB.Select(&requests, query, userID)
	if err != nil {
		logger.Error("Failed to list user client requests for %s: %v", userID, err)
		render.Error(w, "Failed to fetch client requests", http.StatusInternalServerError)
		return
	}
	if requests == nil {
		requests = []models.ClientRequest{}
	}

	render.Success(w, requests)
}

// DELETE /api/client-requests/me/{id} — cancel a pending client request
func (h *BountyHandler) CancelMeRequest(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering BountyHandler.CancelMeRequest")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	res, err := h.DB.Exec(`
		UPDATE client_request 
		SET status = 'cancelled' 
		WHERE id = $1 AND customer_id = $2 AND status = 'pending'
	`, id, userID)

	if err != nil {
		logger.Error("Failed to cancel user client request %s for user %s: %v", id, userID, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		render.Error(w, "Request cannot be cancelled. It may not exist, belong to you, or is already being processed.", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
