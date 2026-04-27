package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

// SettingsHandler manages admin-configurable global setting.
type SettingsHandler struct {
	Service *service.SettingsService
}

func NewSettingsHandler(s *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{Service: s}
}

// GET /api/admin/settings
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	s, err := h.Service.GetSettings(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Settings error: %v", err)
		render.Error(w, "failed to load settings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	render.Success(w, s)
}

// GET /api/settings (Public)
func (h *SettingsHandler) PublicGet(w http.ResponseWriter, r *http.Request) {
	s, err := h.Service.GetSettings(r.Context())
	if err != nil {
		// On error return empty public settings — no sensitive defaults leak
		render.Success(w, models.PublicSettings{})
		return
	}

	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)
	if isAdmin {
		render.Success(w, s)
		return
	}

	render.Success(w, s.ToPublic())
}

// PUT /api/admin/settings
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		USDToCOPRate            *float64 `json:"usd_to_cop_rate"`
		EURToCOPRate            *float64 `json:"eur_to_cop_rate"`
		CKToCOPRate             *float64 `json:"ck_to_cop_rate"`
		ContactAddress          *string  `json:"contact_address"`
		ContactPhone            *string  `json:"contact_phone"`
		ContactEmail            *string  `json:"contact_email"`
		ContactInstagram        *string  `json:"contact_instagram"`
		ContactHours            *string  `json:"contact_hours"`
		FlatShippingFeeCOP      *float64 `json:"flat_shipping_fee_cop"`
		HotSalesThreshold       *int     `json:"hot_sales_threshold"`
		HotDaysThreshold        *int     `json:"hot_days_threshold"`
		NewDaysThreshold        *int     `json:"new_days_threshold"`
		DefaultLocale           *string  `json:"default_locale"`
		HideLanguageSelector    *bool    `json:"hide_language_selector"`
		DefaultThemeID          *string  `json:"default_theme_id"`
		DeliveryPriorityEnabled *bool    `json:"delivery_priority_enabled"`
		PriorityShippingFeeCOP  *float64 `json:"priority_shipping_fee_cop"`
		SynergyMaxPriceCOP      *float64 `json:"synergy_max_price_cop"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if input.USDToCOPRate != nil {
		if err := h.Service.Upsert(r.Context(), "usd_to_cop_rate", strconv.FormatFloat(*input.USDToCOPRate, 'f', 4, 64)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update usd_to_cop_rate: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.EURToCOPRate != nil {
		if err := h.Service.Upsert(r.Context(), "eur_to_cop_rate", strconv.FormatFloat(*input.EURToCOPRate, 'f', 4, 64)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update eur_to_cop_rate: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.CKToCOPRate != nil {
		if err := h.Service.Upsert(r.Context(), "ck_to_cop_rate", strconv.FormatFloat(*input.CKToCOPRate, 'f', 4, 64)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update ck_to_cop_rate: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactAddress != nil {
		if err := h.Service.Upsert(r.Context(), "contact_address", *input.ContactAddress); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update contact_address: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactPhone != nil {
		if err := h.Service.Upsert(r.Context(), "contact_phone", *input.ContactPhone); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update contact_phone: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactEmail != nil {
		if err := h.Service.Upsert(r.Context(), "contact_email", *input.ContactEmail); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update contact_email: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactInstagram != nil {
		if err := h.Service.Upsert(r.Context(), "contact_instagram", *input.ContactInstagram); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update contact_instagram: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.ContactHours != nil {
		if err := h.Service.Upsert(r.Context(), "contact_hours", *input.ContactHours); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update contact_hours: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.FlatShippingFeeCOP != nil {
		if err := h.Service.Upsert(r.Context(), "flat_shipping_fee_cop", strconv.FormatFloat(*input.FlatShippingFeeCOP, 'f', 2, 64)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update flat_shipping_fee_cop: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.HotSalesThreshold != nil {
		if err := h.Service.Upsert(r.Context(), "hot_sales_threshold", strconv.Itoa(*input.HotSalesThreshold)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update hot_sales_threshold: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.HotDaysThreshold != nil {
		if err := h.Service.Upsert(r.Context(), "hot_days_threshold", strconv.Itoa(*input.HotDaysThreshold)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update hot_days_threshold: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.NewDaysThreshold != nil {
		if err := h.Service.Upsert(r.Context(), "new_threshold_days", strconv.Itoa(*input.NewDaysThreshold)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update new_threshold_days: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.DefaultLocale != nil {
		if err := h.Service.Upsert(r.Context(), "default_locale", *input.DefaultLocale); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update default_locale: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.HideLanguageSelector != nil {
		val := "false"
		if *input.HideLanguageSelector {
			val = "true"
		}
		if err := h.Service.Upsert(r.Context(), "hide_language_selector", val); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update hide_language_selector: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.DefaultThemeID != nil {
		if err := h.Service.Upsert(r.Context(), "default_theme_id", *input.DefaultThemeID); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update default_theme_id: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.DeliveryPriorityEnabled != nil {
		val := "false"
		if *input.DeliveryPriorityEnabled {
			val = "true"
		}
		if err := h.Service.Upsert(r.Context(), "delivery_priority_enabled", val); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update delivery_priority_enabled: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.SynergyMaxPriceCOP != nil {
		if err := h.Service.Upsert(r.Context(), "synergy_max_price_cop", strconv.FormatFloat(*input.SynergyMaxPriceCOP, 'f', 2, 64)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update synergy_max_price_cop: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}
	if input.PriorityShippingFeeCOP != nil {
		if err := h.Service.Upsert(r.Context(), "priority_shipping_fee_cop", strconv.FormatFloat(*input.PriorityShippingFeeCOP, 'f', 2, 64)); err != nil {
			logger.ErrorCtx(r.Context(), "Failed to update priority_shipping_fee_cop: %v", err)
			render.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
	}

	s, err := h.Service.GetSettings(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to load settings after update: %v", err)
		render.Error(w, "Update succeeded but failed to reload", http.StatusInternalServerError)
		return
	}
	render.Success(w, s)
}
