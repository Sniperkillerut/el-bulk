package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
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
	q := r.URL.Query()
	limitStr := q.Get("limit")
	limit := 0
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	notices, err := h.Service.List(true, limit)
	if err != nil {
		logger.Error("Failed to list notices: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, notices)
}

// GET /api/notices/{slug} - Public detail
func (h *NoticeHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	notice, err := h.Service.GetBySlug(slug)
	if err != nil {
		render.Error(w, "Notice not found", http.StatusNotFound)
		return
	}

	render.Success(w, notice)
}

// GET /api/admin/notices
func (h *NoticeHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	// Admin wants all notices, no limit
	notices, err := h.Service.List(false, 0)
	if err != nil {
		logger.Error("Error listing notices: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, notices)
}

// POST /api/admin/notices
func (h *NoticeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.NoticeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.ContentHTML == "" {
		render.Error(w, "Title and ContentHTML are required", http.StatusBadRequest)
		return
	}

	res, err := h.Service.Create(input)
	if err != nil {
		logger.Error("Error creating notice: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, res)
}

// PUT /api/admin/notices/{id}
func (h *NoticeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		render.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var input models.NoticeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.Service.Update(id, input)
	if err != nil {
		logger.Error("Error updating notice %s: %v", id, err)
		render.Error(w, "Notice not found or database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, res)
}

// DELETE /api/admin/notices/{id}
func (h *NoticeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		render.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(id); err != nil {
		logger.Error("Error deleting notice %s: %v", id, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "Notice deleted"})
}
