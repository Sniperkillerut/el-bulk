package service

import (
	"strconv"
	"sync"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

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

func (s *SettingsService) GetSettings() (models.Settings, error) {
	s.mu.RLock()
	if time.Since(s.cacheTime) < s.cacheDuration {
		val := s.cache
		s.mu.RUnlock()
		return val, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check after acquiring lock
	if time.Since(s.cacheTime) < s.cacheDuration {
		return s.cache, nil
	}

	settings := models.Settings{
		USDToCOPRate:       4200,
		EURToCOPRate:       4600,
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

	raw, err := s.Store.GetAll()
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

func (s *SettingsService) Upsert(key, value string) error {
	err := s.Store.Upsert(key, value)
	if err != nil {
		return err
	}
	
	s.mu.Lock()
	s.cacheTime = time.Time{} // Invalidate cache
	s.mu.Unlock()
	
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
