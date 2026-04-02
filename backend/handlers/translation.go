package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/models"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type TranslationHandler struct {
	DB *sqlx.DB
}

func NewTranslationHandler(db *sqlx.DB) *TranslationHandler {
	return &TranslationHandler{DB: db}
}

// List returns all translations grouped by locale: { "en": { "key": "value" }, "es": { ... } }
func (h *TranslationHandler) List(w http.ResponseWriter, r *http.Request) {
	locale := r.URL.Query().Get("locale")
	if locale != "" {
		translations, err := models.GetTranslationsByLocale(h.DB, locale)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(translations)
		return
	}

	translations, err := models.GetAllTranslations(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(translations)
}

// AdminList returns all translations as a flat list for the admin table
func (h *TranslationHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	translations, err := models.ListAllTranslationKeys(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(translations)
}

// Update upserts a translation
func (h *TranslationHandler) Update(w http.ResponseWriter, r *http.Request) {
	var t models.Translation
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Key from URL takes precedence if provided
	key := chi.URLParam(r, "key")
	if key != "" {
		t.Key = key
	}

	if t.Key == "" || t.Locale == "" || t.Value == "" {
		http.Error(w, "Key, locale, and value are required", http.StatusBadRequest)
		return
	}

	if err := models.UpsertTranslation(h.DB, t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Delete removes a translation
func (h *TranslationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	locale := r.URL.Query().Get("locale")

	if key == "" || locale == "" {
		http.Error(w, "Key and locale are required", http.StatusBadRequest)
		return
	}

	if err := models.DeleteTranslation(h.DB, key, locale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
