package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
)

type TranslationHandler struct {
	Service *service.TranslationService
}

func NewTranslationHandler(s *service.TranslationService) *TranslationHandler {
	return &TranslationHandler{Service: s}
}

// List returns all translations grouped by locale: { "en": { "key": "value" }, "es": { ... } }
func (h *TranslationHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering TranslationHandler.List")
	locale := r.URL.Query().Get("locale")
	if locale != "" {
		translations, err := h.Service.GetByLocale(r.Context(), locale)
		if err != nil {
			logger.ErrorCtx(r.Context(), "Failed to get translations for locale %s: %v", locale, err)
			render.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		render.Success(w, translations)
		return
	}

	translations, err := h.Service.GetAll(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to get all translations: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, translations)
}

// AdminList returns all translations as a flat list for the admin table
func (h *TranslationHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering TranslationHandler.AdminList")
	translations, err := h.Service.ListKeys(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list translation keys: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, translations)
}

// Update upserts a translation
func (h *TranslationHandler) Update(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering TranslationHandler.Update")
	var t models.Translation
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	key := chi.URLParam(r, "key")
	if key != "" {
		t.Key = key
	}

	if err := h.Service.Upsert(r.Context(), t); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to upsert translation %s/%s: %v", t.Locale, t.Key, err)
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.Success(w, map[string]string{"status": "success"})
}

// Delete removes a translation
func (h *TranslationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering TranslationHandler.Delete")
	key := chi.URLParam(r, "key")
	locale := r.URL.Query().Get("locale")

	if key == "" || locale == "" {
		render.Error(w, "Key and locale are required", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), key, locale); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to delete translation %s/%s: %v", locale, key, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"status": "success"})
}

// DeleteLocale removes all translations for a full locale
func (h *TranslationHandler) DeleteLocale(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering TranslationHandler.DeleteLocale")
	locale := chi.URLParam(r, "locale")

	if locale == "" {
		render.Error(w, "Locale is required", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteLocale(r.Context(), locale); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to delete locale %s: %v", locale, err)
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.Success(w, map[string]string{"status": "success"})
}
