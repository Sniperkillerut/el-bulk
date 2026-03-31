package handlers

import (
"github.com/el-bulk/backend/utils/render"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/external"
	"time"
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
	err := h.DB.Select(&tcgs, `
		SELECT t.*, COUNT(p.id) as item_count 
		FROM tcg t 
		LEFT JOIN product p ON t.id = p.tcg 
		GROUP BY t.id 
		ORDER BY t.name
	`)
	if err != nil {
		logger.Error("Error listing TCGs for admin: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if tcgs == nil {
		tcgs = []models.TCG{}
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

	var tcg models.TCG
	err := h.DB.QueryRowx(`
		INSERT INTO tcg (id, name, is_active)
		VALUES ($1, $2, $3)
		RETURNING *
	`, input.ID, input.Name, true).StructScan(&tcg)

	if err != nil {
		logger.Error("Error creating TCG: %v", err)
		render.Error(w, "Failed to create TCG (ID may already exist)", http.StatusInternalServerError)
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

	var tcg models.TCG
	err := h.DB.QueryRowx(`
		UPDATE tcg
		SET name = $1, is_active = $2
		WHERE id = $3
		RETURNING *
	`, input.Name, input.IsActive, id).StructScan(&tcg)

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

	// Check if any products exist for this TCG
	var productCount int
	err := h.DB.Get(&productCount, "SELECT COUNT(*) FROM product WHERE tcg = $1", id)
	logger.Info("[TCG_DELETE] Checking products for %s: count=%d", id, productCount)
	if err != nil {
		render.Error(w, "Database error checking products", http.StatusInternalServerError)
		return
	}

	if productCount > 0 {
		render.Error(w, "Cannot delete TCG with existing products. Delete products first.", http.StatusConflict)
		return
	}

	result, err := h.DB.Exec("DELETE FROM tcg WHERE id = $1", id)
	if err != nil {
		logger.Error("Error deleting TCG %s: %v", id, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		render.Error(w, "TCG not found", http.StatusNotFound)
		return
	}

	render.Success(w, map[string]string{"message": "TCG deleted successfully"})
}

// POST /api/admin/tcgs/sync-sets
func (h *TCGHandler) SyncSets(w http.ResponseWriter, r *http.Request) {
	sets, err := external.FetchSets()
	if err != nil {
		logger.Error("Error fetching MTG sets from Scryfall: %v", err)
		render.Error(w, "Failed to reach Scryfall", http.StatusBadGateway)
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		render.Error(w, "Database transaction error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, s := range sets {
		_, err := tx.Exec(`
			INSERT INTO tcg_set (tcg, code, name, released_at, set_type)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tcg, code) DO UPDATE SET
				name = EXCLUDED.name,
				released_at = EXCLUDED.released_at,
				set_type = EXCLUDED.set_type
		`, "mtg", s.Code, s.Name, s.ReleasedAt, s.SetType)
		if err != nil {
			logger.Error("Error upserting set %s: %v", s.Code, err)
			continue
		}
	}

	_, _ = tx.Exec("INSERT INTO setting (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "last_set_sync", time.Now().Format(time.RFC3339))

	if err := tx.Commit(); err != nil {
		render.Error(w, "Failed to commit sets sync", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]interface{}{
		"count": len(sets),
		"last_sync": time.Now().Format(time.RFC3339),
	})
}
