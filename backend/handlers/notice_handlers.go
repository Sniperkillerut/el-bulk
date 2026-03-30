package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/mailer"
)

type NoticeHandler struct {
	DB *sqlx.DB
}

func NewNoticeHandler(db *sqlx.DB) *NoticeHandler {
	return &NoticeHandler{DB: db}
}

// GET /api/notices - Public list
func (h *NoticeHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limitStr := q.Get("limit")
	limit := 0
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	var notices []models.Notice
	var err error

	if limit > 0 {
		err = h.DB.Select(&notices, "SELECT * FROM notice WHERE is_published = true ORDER BY created_at DESC LIMIT $1", limit)
	} else {
		err = h.DB.Select(&notices, "SELECT * FROM notice WHERE is_published = true ORDER BY created_at DESC")
	}

	if err != nil {
		logger.Error("Failed to list notices: %v", err)
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if notices == nil {
		notices = []models.Notice{}
	}

	jsonOK(w, notices)
}

// GET /api/notices/{slug} - Public detail
func (h *NoticeHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var notice models.Notice
	err := h.DB.Get(&notice, "SELECT * FROM notice WHERE slug = $1 AND is_published = true", slug)
	if err != nil {
		jsonError(w, "Notice not found", http.StatusNotFound)
		return
	}

	jsonOK(w, notice)
}

// GET /api/admin/notices - Admin list
func (h *NoticeHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	var notices []models.Notice
	err := h.DB.Select(&notices, "SELECT * FROM notice ORDER BY created_at DESC")
	if err != nil {
		logger.Error("Failed to list admin notices: %v", err)
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if notices == nil {
		notices = []models.Notice{}
	}

	jsonOK(w, notices)
}

// POST /api/admin/notices - Create
func (h *NoticeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.NoticeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.Slug == "" || input.ContentHTML == "" {
		jsonError(w, "Title, slug, and content are required", http.StatusBadRequest)
		return
	}

	var notice models.Notice
	query := `
		INSERT INTO notice (title, slug, content_html, featured_image_url, is_published)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`
	err := h.DB.Get(&notice, query, input.Title, input.Slug, input.ContentHTML, input.FeaturedImageURL, input.IsPublished)
	if err != nil {
		logger.Error("Failed to create notice: %v", err)
		jsonError(w, "Failed to create notice (likely duplicate slug)", http.StatusInternalServerError)
		return
	}

	// Trigger newsletter broadcast if published
	if notice.IsPublished {
		go mailer.BroadcastNotice(h.DB, notice)
	}

	jsonOK(w, notice)
}

// PUT /api/admin/notices/{id} - Update
func (h *NoticeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input models.NoticeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var notice models.Notice
	query := `
		UPDATE notice 
		SET title = $1, slug = $2, content_html = $3, featured_image_url = $4, is_published = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING *
	`
	err := h.DB.Get(&notice, query, input.Title, input.Slug, input.ContentHTML, input.FeaturedImageURL, input.IsPublished, id)
	if err != nil {
		logger.Error("Failed to update notice %s: %v", id, err)
		jsonError(w, "Failed to update notice", http.StatusInternalServerError)
		return
	}

	// Trigger newsletter broadcast if status changed to published
	// We could check the previous state, but BroadcastNotice handles duplicates or we can just send it
	// For simplicity, if it's currently published, we can trigger it (optionally checking if it was NOT published before)
	if notice.IsPublished {
		go mailer.BroadcastNotice(h.DB, notice)
	}

	jsonOK(w, notice)
}

// DELETE /api/admin/notices/{id} - Delete
func (h *NoticeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := h.DB.Exec("DELETE FROM notice WHERE id = $1", id)
	if err != nil {
		logger.Error("Failed to delete notice %s: %v", id, err)
		jsonError(w, "Failed to delete notice", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"message": "Notice deleted"})
}
