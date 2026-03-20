package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/models"
)

// SettingsHandler manages admin-configurable global settings.
type SettingsHandler struct {
	DB *sqlx.DB
}

func NewSettingsHandler(db *sqlx.DB) *SettingsHandler {
	return &SettingsHandler{DB: db}
}

// loadSettings reads the current exchange rates from the settings table.
func loadSettings(db *sqlx.DB) (models.Settings, error) {
	rows, err := db.Query("SELECT key, value FROM settings WHERE key IN ('usd_to_cop_rate','eur_to_cop_rate')")
	if err != nil {
		return models.Settings{}, err
	}
	defer rows.Close()

	s := models.Settings{USDToCOPRate: 4200, EURToCOPRate: 4600}
	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			continue
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			continue
		}
		switch key {
		case "usd_to_cop_rate":
			s.USDToCOPRate = f
		case "eur_to_cop_rate":
			s.EURToCOPRate = f
		}
	}
	return s, rows.Err()
}

// GET /api/admin/settings
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	s, err := loadSettings(h.DB)
	if err != nil {
		jsonError(w, "failed to load settings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, s)
}

// PUT /api/admin/settings
// Body: { "usd_to_cop_rate": 4350.0, "eur_to_cop_rate": 4750.0 }
// Omit a field to leave it unchanged.
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		USDToCOPRate *float64 `json:"usd_to_cop_rate"`
		EURToCOPRate *float64 `json:"eur_to_cop_rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if input.USDToCOPRate != nil {
		if _, err := h.DB.Exec(
			"INSERT INTO settings(key,value) VALUES('usd_to_cop_rate',$1) ON CONFLICT(key) DO UPDATE SET value=$1",
			strconv.FormatFloat(*input.USDToCOPRate, 'f', 4, 64),
		); err != nil {
			jsonError(w, "failed to update USD rate: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if input.EURToCOPRate != nil {
		if _, err := h.DB.Exec(
			"INSERT INTO settings(key,value) VALUES('eur_to_cop_rate',$1) ON CONFLICT(key) DO UPDATE SET value=$1",
			strconv.FormatFloat(*input.EURToCOPRate, 'f', 4, 64),
		); err != nil {
			jsonError(w, "failed to update EUR rate: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	s, _ := loadSettings(h.DB)
	jsonOK(w, s)
}
