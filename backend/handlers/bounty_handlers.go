package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
)

type BountyHandler struct {
	Service *service.BountyService
}

func NewBountyHandler(s *service.BountyService) *BountyHandler {
	return &BountyHandler{Service: s}
}

// === BOUNTIES ===

func (h *BountyHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.List | URL: %s", r.URL.String())
	activeParam := r.URL.Query().Get("active")
	isAdmin := middleware.IsAdmin(r.Context())
	bounties, err := h.Service.ListBounties(r.Context(), activeParam, isAdmin)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list bounties: %v", err)
		render.Error(w, "Failed to fetch bounties", http.StatusInternalServerError)
		return
	}
	render.Success(w, bounties)
}

func (h *BountyHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.Create")
	var input models.BountyInput
	if err := decodeJSON(r, &input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode bounty input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	bounty, err := h.Service.CreateBounty(r.Context(), input)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to create bounty: %v", err)
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, bounty)
}

func (h *BountyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering BountyHandler.Update | ID: %s", id)
	var input models.BountyInput
	if err := decodeJSON(r, &input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode bounty update for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	bounty, err := h.Service.UpdateBounty(r.Context(), id, input)
	if err != nil {
		if err == sql.ErrNoRows {
			render.Error(w, "Bounty not found", http.StatusNotFound)
			return
		}
		logger.ErrorCtx(r.Context(), "Failed to update bounty: %v", err)
		render.Error(w, "Failed to update bounty", http.StatusInternalServerError)
		return
	}

	render.Success(w, bounty)
}

func (h *BountyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering BountyHandler.Delete | ID: %s", id)
	count, err := h.Service.DeleteBounty(r.Context(), id)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to delete bounty: %v", err)
		render.Error(w, "Failed to delete bounty", http.StatusInternalServerError)
		return
	}

	if count == 0 {
		render.Error(w, "Bounty not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === BOUNTY OFFERS ===

func (h *BountyHandler) ListOffers(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.ListOffers")
	offers, err := h.Service.ListOffers(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list bounty offers: %v", err)
		render.Error(w, "Failed to fetch bounty offers", http.StatusInternalServerError)
		return
	}
	render.Success(w, offers)
}

func (h *BountyHandler) SubmitOffer(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.SubmitOffer")
	var input models.BountyOfferInput
	if err := decodeJSON(r, &input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode offer input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var userID *string
	if val, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
		userID = &val
	}

	offer, err := h.Service.SubmitOffer(r.Context(), input, userID)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to submit offer: %v", err)
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, offer)
}

func (h *BountyHandler) UpdateOfferStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering BountyHandler.UpdateOfferStatus | ID: %s", id)
	var input models.UpdateBountyOfferStatusInput
	if err := decodeJSON(r, &input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode offer status update for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{"pending": true, "accepted": true, "rejected": true, "fulfilled": true}
	if !validStatuses[input.Status] {
		render.Error(w, "Invalid status value. Must be one of: pending, accepted, rejected, fulfilled", http.StatusBadRequest)
		return
	}

	offer, err := h.Service.UpdateOfferStatus(r.Context(), id, input.Status)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to update offer status: %v", err)
		render.Error(w, "Failed to update offer status", http.StatusInternalServerError)
		return
	}

	render.Success(w, offer)
}

// === CLIENT REQUESTS ===

func (h *BountyHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.ListRequests")
	requests, err := h.Service.ListRequests(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list client requests: %v", err)
		render.Error(w, "Failed to fetch client requests", http.StatusInternalServerError)
		return
	}
	render.Success(w, requests)
}

func (h *BountyHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.CreateRequest")
	var input models.ClientRequestInput
	if err := decodeJSON(r, &input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode client request input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var userID *string
	if val, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
		userID = &val
	}

	req, err := h.Service.SubmitRequest(r.Context(), input, userID)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to submit request: %v", err)
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, req)
}

func (h *BountyHandler) UpdateRequestStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering BountyHandler.UpdateRequestStatus | ID: %s", id)
	var input models.UpdateClientRequestStatusInput
	if err := decodeJSON(r, &input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode client request status update for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{"pending": true, "accepted": true, "rejected": true, "solved": true}
	if !validStatuses[input.Status] {
		render.Error(w, "Invalid status value. Must be one of: pending, accepted, rejected, solved", http.StatusBadRequest)
		return
	}

	req, err := h.Service.UpdateRequestStatus(r.Context(), id, input.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			render.Error(w, "Request not found", http.StatusNotFound)
			return
		}
		logger.ErrorCtx(r.Context(), "Failed to update client request status: %v", err)
		render.Error(w, "Failed to update request status", http.StatusInternalServerError)
		return
	}

	render.Success(w, req)
}

// === USER-FACING (Me) ===

func (h *BountyHandler) ListMeOffers(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.ListMeOffers")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	offers, err := h.Service.ListMeOffers(r.Context(), userID)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list user bounty offers for %s: %v", userID, err)
		render.Error(w, "Failed to fetch bounty offers", http.StatusInternalServerError)
		return
	}
	render.Success(w, offers)
}

func (h *BountyHandler) CancelMeOffer(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.CancelMeOffer")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.Service.CancelMeOffer(r.Context(), id, userID); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to cancel user bounty offer %s for user %s: %v", id, userID, err)
		render.Error(w, "Offer cannot be cancelled. It may not exist, belong to you, or is already being processed.", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BountyHandler) ListMeRequests(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.ListMeRequests")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	requests, err := h.Service.ListMeRequests(r.Context(), userID)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list user client requests for %s: %v", userID, err)
		render.Error(w, "Failed to fetch client requests", http.StatusInternalServerError)
		return
	}
	render.Success(w, requests)
}

func (h *BountyHandler) CancelMeRequest(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering BountyHandler.CancelMeRequest")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.Service.CancelMeRequest(r.Context(), id, userID); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to cancel user client request %s for user %s: %v", id, userID, err)
		render.Error(w, "Request cannot be cancelled. It may not exist, belong to you, or is already being processed.", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helper to decode JSON bodies
func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
