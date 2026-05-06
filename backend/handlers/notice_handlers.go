package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/httputil"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
)

type NoticeHandler struct {
	Service *service.NoticeService
}

func NewNoticeHandler(s *service.NoticeService) *NoticeHandler {
	return &NoticeHandler{Service: s}
}

// GET /api/notices - Public list
func (h *NoticeHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering NoticeHandler.List | URL: %s", r.URL.String())
	q := r.URL.Query()
	limitStr := q.Get("limit")
	limit := 0
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	notices, err := h.Service.List(r.Context(), true, limit)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list notices: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, notices)
}

// GET /api/notices/{slug} - Public detail
func (h *NoticeHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	logger.TraceCtx(r.Context(), "Entering NoticeHandler.GetBySlug | Slug: %s", slug)

	notice, err := h.Service.GetBySlug(r.Context(), slug)
	if err != nil {
		render.Error(w, "Notice not found", http.StatusNotFound)
		return
	}

	render.Success(w, notice)
}

// GET /api/admin/notices
func (h *NoticeHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering NoticeHandler.AdminList")
	// Admin wants all notices, no limit
	notices, err := h.Service.List(r.Context(), false, 0)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Error listing notices: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, notices)
}

// POST /api/admin/notices
func (h *NoticeHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering NoticeHandler.Create")
	var input models.NoticeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode notice input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.ContentHTML == "" {
		render.Error(w, "Title and ContentHTML are required", http.StatusBadRequest)
		return
	}

	res, err := h.Service.Create(r.Context(), input)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Error creating notice: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, res)
}

// PUT /api/admin/notices/{id}
func (h *NoticeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering NoticeHandler.Update | ID: %s", id)
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Valid UUID is required", http.StatusBadRequest)
		return
	}

	var input models.NoticeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode notice update for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.Service.Update(r.Context(), id, input)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Error updating notice %s: %v", id, err)
		render.Error(w, "Notice not found or database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, res)
}

// DELETE /api/admin/notices/{id}
func (h *NoticeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering NoticeHandler.Delete | ID: %s", id)
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Valid UUID is required", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		logger.ErrorCtx(r.Context(), "Error deleting notice %s: %v", id, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "Notice deleted"})
}
