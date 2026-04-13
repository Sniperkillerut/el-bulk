package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

type NewsletterHandler struct {
	Service *service.NewsletterService
}

func NewNewsletterHandler(s *service.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{Service: s}
}

func (h *NewsletterHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering NewsletterHandler.Subscribe")
	var input struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if input.Email == "" {
		render.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	msg, err := h.Service.Subscribe(r.Context(), input.Email)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Newsletter subscribe failed: %v", err)
		render.Error(w, "Failed to subscribe", http.StatusInternalServerError)
		return
	}

	if msg == "Already subscribed" {
		render.Success(w, map[string]string{"message": msg})
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, map[string]string{"message": msg})
}

func (h *NewsletterHandler) AdminGetSubscribers(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering NewsletterHandler.AdminGetSubscribers")
	subscribers, err := h.Service.ListAll(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list subscribers: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, subscribers)
}

func (h *NewsletterHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering NewsletterHandler.Unsubscribe")
	var input struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Service.Unsubscribe(r.Context(), input.Email); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to unsubscribe %s: %v", input.Email, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "Unsubscribed successfully"})
}
