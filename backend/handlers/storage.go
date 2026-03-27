package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/models"
)

type StorageHandler struct {
	DB *sqlx.DB
}

func NewStorageHandler(db *sqlx.DB) *StorageHandler {
	return &StorageHandler{DB: db}
}

// GET /api/admin/storage
func (h *StorageHandler) List(w http.ResponseWriter, r *http.Request) {
	var locations []models.StoredIn
	err := h.DB.Select(&locations, `
		SELECT 
			s.id, 
			s.name, 
			COALESCE(SUM(ps.quantity), 0) AS item_count 
		FROM storage_location s 
		LEFT JOIN product_storage ps ON s.id = ps.storage_id 
		GROUP BY s.id, s.name 
		ORDER BY s.name
	`)
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if locations == nil {
		locations = []models.StoredIn{}
	}
	jsonOK(w, locations)
}

// POST /api/admin/storage
func (h *StorageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.StoredIn
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRow("INSERT INTO storage_location (name) VALUES ($1) RETURNING id", input.Name).Scan(&input.ID)
	if err != nil {
		jsonError(w, "Failed to create location or name already exists", http.StatusInternalServerError)
		return
	}
	jsonOK(w, input)
}

// PUT /api/admin/storage/:id
func (h *StorageHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.StoredIn
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("UPDATE storage_location SET name = $1 WHERE id = $2", input.Name, id)
	if err != nil {
		jsonError(w, "Failed to update location", http.StatusInternalServerError)
		return
	}
	input.ID = id
	jsonOK(w, input)
}

// DELETE /api/admin/storage/:id
func (h *StorageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.DB.Exec("DELETE FROM storage_location WHERE id = $1", id)
	if err != nil {
		jsonError(w, "Failed to delete location", http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted successfully"})
}
