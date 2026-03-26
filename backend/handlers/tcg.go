package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
)

type TCGHandler struct {
	DB *sqlx.DB
}

func NewTCGHandler(db *sqlx.DB) *TCGHandler {
	return &TCGHandler{DB: db}
}

// GET /api/admin/tcgs
func (h *TCGHandler) List(w http.ResponseWriter, r *http.Request) {
	var tcgs []models.TCG
	err := h.DB.Select(&tcgs, "SELECT * FROM tcg ORDER BY name")
	if err != nil {
		logger.Error("Error listing TCGs for admin: %v", err)
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if tcgs == nil {
		tcgs = []models.TCG{}
	}
	jsonOK(w, tcgs)
}

// POST /api/admin/tcgs
func (h *TCGHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.TCGInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.ID == "" || input.Name == "" {
		jsonError(w, "ID (slug) and Name are required", http.StatusBadRequest)
		return
	}

	var tcg models.TCG
	err := h.DB.QueryRowx(`
		INSERT INTO tcg (id, name, is_active)
		VALUES ($1, $2, $3)
		RETURNING *
	`, input.ID, input.Name, true).StructScan(&tcg)

	if err != nil {
		logger.Error("Error creating TCG: %v", err)
		jsonError(w, "Failed to create TCG (ID may already exist)", http.StatusInternalServerError)
		return
	}

	jsonOK(w, tcg)
}

// PUT /api/admin/tcgs/{id}
func (h *TCGHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.TCGInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var tcg models.TCG
	err := h.DB.QueryRowx(`
		UPDATE tcg
		SET name = $1, is_active = $2
		WHERE id = $3
		RETURNING *
	`, input.Name, input.IsActive, id).StructScan(&tcg)

	if err != nil {
		logger.Error("Error updating TCG %s: %v", id, err)
		jsonError(w, "TCG not found or update failed", http.StatusNotFound)
		return
	}

	jsonOK(w, tcg)
}

// DELETE /api/admin/tcgs/{id}
func (h *TCGHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Check if any products exist for this TCG
	var productCount int
	err := h.DB.Get(&productCount, "SELECT COUNT(*) FROM product WHERE tcg = $1", id)
	if err != nil {
		jsonError(w, "Database error checking products", http.StatusInternalServerError)
		return
	}

	if productCount > 0 {
		jsonError(w, "Cannot delete TCG with existing products. Delete products first.", http.StatusConflict)
		return
	}

	result, err := h.DB.Exec("DELETE FROM tcg WHERE id = $1", id)
	if err != nil {
		logger.Error("Error deleting TCG %s: %v", id, err)
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, "TCG not found", http.StatusNotFound)
		return
	}

	jsonOK(w, map[string]string{"message": "TCG deleted successfully"})
}
