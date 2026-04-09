package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"strings"
)

type TCGHandler struct {
	Service *service.TCGService
}

func NewTCGHandler(s *service.TCGService) *TCGHandler {
	return &TCGHandler{Service: s}
}

// GET /api/admin/tcgs
func (h *TCGHandler) List(w http.ResponseWriter, r *http.Request) {
	tcgs, err := h.Service.List(true) // For now returns all with counts
	if err != nil {
		logger.Error("Error listing TCGs for admin: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, tcgs)
}

// POST /api/admin/tcgs
func (h *TCGHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.TCGInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.ID == "" || input.Name == "" {
		render.Error(w, "ID (slug) and Name are required", http.StatusBadRequest)
		return
	}

	tcg, err := h.Service.Create(input)
	if err != nil {
		logger.Error("Error creating TCG: %v", err)
		
		// Check for PostgreSQL unique constraint violation
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			render.Error(w, fmt.Sprintf("TCG with ID '%s' already exists", input.ID), http.StatusConflict)
			return
		}

		render.Error(w, "Failed to create TCG: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render.Success(w, tcg)
}

// PUT /api/admin/tcgs/{id}
func (h *TCGHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.TCGInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tcg, err := h.Service.Update(id, input)
	if err != nil {
		logger.Error("Error updating TCG %s: %v", id, err)
		render.Error(w, "TCG not found or update failed", http.StatusNotFound)
		return
	}

	render.Success(w, tcg)
}

// DELETE /api/admin/tcgs/{id}
func (h *TCGHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.Info("[TCG_DELETE] 📥 Received DELETE request for ID: %s", id)

	if id == "" {
		render.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	err := h.Service.Delete(id)
	if err != nil {
		logger.Error("Error deleting TCG %s: %v", id, err)
		if strings.Contains(err.Error(), "existing products") {
			render.Error(w, err.Error(), http.StatusConflict)
			return
		}
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "TCG deleted successfully"})
}

// POST /api/admin/tcgs/{id}/sync-sets
func (h *TCGHandler) SyncSets(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	count, err := h.Service.SyncSets(id)
	if err != nil {
		logger.Error("Error syncing TCG %s: %v", id, err)
		render.Error(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]interface{}{
		"message":    "Sync completed",
		"sets_count": count,
	})
}
