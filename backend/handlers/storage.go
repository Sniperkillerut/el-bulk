package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
)

type StorageHandler struct {
	Service *service.StorageLocationService
}

func NewStorageHandler(s *service.StorageLocationService) *StorageHandler {
	return &StorageHandler{Service: s}
}

// GET /api/admin/storage
func (h *StorageHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering StorageHandler.List")
	locations, err := h.Service.List()
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, locations)
}

// POST /api/admin/storage
func (h *StorageHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering StorageHandler.Create")
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	loc, err := h.Service.Create(input.Name)
	if err != nil {
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	render.Success(w, loc)
}

// PUT /api/admin/storage/:id
func (h *StorageHandler) Update(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering StorageHandler.Update")
	id := chi.URLParam(r, "id")
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.Update(id, input.Name); err != nil {
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	render.Success(w, map[string]string{"id": id, "name": input.Name})
}

// DELETE /api/admin/storage/:id
func (h *StorageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering StorageHandler.Delete")
	id := chi.URLParam(r, "id")
	if err := h.Service.Delete(id); err != nil {
		render.Error(w, "Failed to delete location", http.StatusInternalServerError)
		return
	}
	render.Success(w, map[string]string{"message": "Deleted successfully"})
}
