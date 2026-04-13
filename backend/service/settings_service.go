package service

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type SettingsProvider interface {
	GetSettings(ctx context.Context) (models.Settings, error)
	Upsert(ctx context.Context, key, value string) error
	InvalidateCache()
}

type SettingsService struct {
	Store *store.SettingsStore
	
	cache         models.Settings
	cacheTime     time.Time
	cacheDuration time.Duration
	mu            sync.RWMutex
}

func NewSettingsService(s *store.SettingsStore) *SettingsService {
	return &SettingsService{
		Store:         s,
		cacheDuration: 60 * time.Second,
	}
}

func (s *SettingsService) GetSettings(ctx context.Context) (models.Settings, error) {
	s.mu.RLock()
	if time.Since(s.cacheTime) < s.cacheDuration {
		val := s.cache
		s.mu.RUnlock()
		logger.DebugCtx(ctx, "Settings cache hit")
		return val, nil
	}
	s.mu.RUnlock()
	logger.TraceCtx(ctx, "Settings cache miss or expired, reloading from DB")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check after acquiring lock
	if time.Since(s.cacheTime) < s.cacheDuration {
		return s.cache, nil
	}

	settings := models.Settings{
		USDToCOPRate:       4200,
		EURToCOPRate:       4600,
		CKToCOPRate:        4000,
		ContactAddress:     "Cra. 15 # 76-54, Local 201, Centro Comercial Unilago, Bogotá",
		ContactPhone:       "+57 300 000 0000",
		ContactEmail:       "contact@el-bulk.co",
		ContactInstagram:   "el-bulk",
		ContactHours:       "Mon - Sat: 11:00 AM - 7:00 PM",
		FlatShippingFeeCOP: 20000,
		HotSalesThreshold:  3,
		HotDaysThreshold:   7,
		NewDaysThreshold:   10,
		DefaultLocale:      "en",
		HideLanguageSelector: false,
	}

	raw, err := s.Store.GetAll(ctx)
	if err != nil {
		return settings, err
	}

	for key, val := range raw {
		switch key {
		case "usd_to_cop_rate":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				settings.USDToCOPRate = f
			}
		case "eur_to_cop_rate":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				settings.EURToCOPRate = f
			}
		case "ck_to_cop_rate":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				settings.CKToCOPRate = f
			}
		case "contact_address":
			settings.ContactAddress = val
		case "contact_phone":
			settings.ContactPhone = val
		case "contact_email":
			settings.ContactEmail = val
		case "contact_instagram":
			settings.ContactInstagram = val
		case "contact_hours":
			settings.ContactHours = val
		case "last_set_sync":
			settings.LastSetSync = val
		case "default_theme_id":
			settings.DefaultThemeID = val
		case "flat_shipping_fee_cop":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				settings.FlatShippingFeeCOP = f
			}
		case "hot_sales_threshold":
			if i, err := strconv.Atoi(val); err == nil {
				settings.HotSalesThreshold = i
			}
		case "hot_days_threshold":
			if i, err := strconv.Atoi(val); err == nil {
				settings.HotDaysThreshold = i
			}
		case "new_threshold_days":
			if i, err := strconv.Atoi(val); err == nil {
				settings.NewDaysThreshold = i
			}
		case "default_locale":
			settings.DefaultLocale = val
		case "hide_language_selector":
			settings.HideLanguageSelector = val == "true"
		}
	}

	s.cache = settings
	s.cacheTime = time.Now()

	return settings, nil
}

func (s *SettingsService) Upsert(ctx context.Context, key, value string) error {
	logger.TraceCtx(ctx, "Entering SettingsService.Upsert | Key: %s", key)
	err := s.Store.Upsert(ctx, key, value)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to upsert setting %s: %v", key, err)
		return err
	}
	
	s.mu.Lock()
	s.cacheTime = time.Time{} // Invalidate cache
	s.mu.Unlock()
	logger.DebugCtx(ctx, "Settings cache invalidated after upsert of %s", key)
	
	return nil
}

func (s *SettingsService) InvalidateCache() {
	s.mu.Lock()
	s.cacheTime = time.Time{}
	s.mu.Unlock()
}

// ResetCache clears the internal settings cache (primarily for unit tests).
func (s *SettingsService) ResetCache() {
s.mu.Lock()
s.cacheTime = time.Now().Add(-2 * time.Minute) // Force expiration
s.mu.Unlock()
}
