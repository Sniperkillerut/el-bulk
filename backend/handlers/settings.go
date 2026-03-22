package handlers

import (
	"encoding/json"
	"fmt"
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
	s := models.Settings{
		USDToCOPRate: 4200,
		EURToCOPRate: 4600,
		ContactAddress: "Cra. 15 # 76-54, Local 201, Centro Comercial Unilago, Bogotá",
		ContactPhone: "+57 300 000 0000",
		ContactEmail: "contact@el-bulk.co",
		ContactInstagram: "el-bulk",
		ContactHours: "Mon - Sat: 11:00 AM - 7:00 PM",
	}

	if db == nil {
		fmt.Println("⚠️ Database is offline, providing defaults.")
		return s, nil
	}

	rows, err := db.Query("SELECT key, value FROM settings")
	if err != nil {
		fmt.Printf("⚠️ Settings table error: %v (using defaults)\n", err)
		return s, nil // Return defaults instead of error to prevent 500
	}
	defer rows.Close()
	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			continue
		}
		switch key {
		case "usd_to_cop_rate":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				s.USDToCOPRate = f
			}
		case "eur_to_cop_rate":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				s.EURToCOPRate = f
			}
		case "contact_address":
			s.ContactAddress = val
		case "contact_phone":
			s.ContactPhone = val
		case "contact_email":
			s.ContactEmail = val
		case "contact_instagram":
			s.ContactInstagram = val
		case "contact_hours":
			s.ContactHours = val
		}
	}
	return s, rows.Err()
}

// GET /api/admin/settings
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	s, err := loadSettings(h.DB)
	if err != nil {
		fmt.Printf("Settings error: %v\n", err)
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
		USDToCOPRate     *float64 `json:"usd_to_cop_rate"`
		EURToCOPRate     *float64 `json:"eur_to_cop_rate"`
		ContactAddress   *string  `json:"contact_address"`
		ContactPhone     *string  `json:"contact_phone"`
		ContactEmail     *string  `json:"contact_email"`
		ContactInstagram *string  `json:"contact_instagram"`
		ContactHours     *string  `json:"contact_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	upsert := func(key, val string) error {
		_, err := h.DB.Exec("INSERT INTO settings(key,value) VALUES($1,$2) ON CONFLICT(key) DO UPDATE SET value=$2", key, val)
		return err
	}

	if input.USDToCOPRate != nil {
		if err := upsert("usd_to_cop_rate", strconv.FormatFloat(*input.USDToCOPRate, 'f', 4, 64)); err != nil {
			jsonError(w, "failed to update USD rate", http.StatusInternalServerError)
			return
		}
	}
	if input.EURToCOPRate != nil {
		if err := upsert("eur_to_cop_rate", strconv.FormatFloat(*input.EURToCOPRate, 'f', 4, 64)); err != nil {
			jsonError(w, "failed to update EUR rate", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactAddress != nil {
		if err := upsert("contact_address", *input.ContactAddress); err != nil {
			jsonError(w, "failed to update contact_address", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactPhone != nil {
		if err := upsert("contact_phone", *input.ContactPhone); err != nil {
			jsonError(w, "failed to update contact_phone", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactEmail != nil {
		if err := upsert("contact_email", *input.ContactEmail); err != nil {
			jsonError(w, "failed to update contact_email", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactInstagram != nil {
		if err := upsert("contact_instagram", *input.ContactInstagram); err != nil {
			jsonError(w, "failed to update contact_instagram", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactHours != nil {
		if err := upsert("contact_hours", *input.ContactHours); err != nil {
			jsonError(w, "failed to update contact_hours", http.StatusInternalServerError)
			return
		}
	}

	s, _ := loadSettings(h.DB)
	jsonOK(w, s)
}
