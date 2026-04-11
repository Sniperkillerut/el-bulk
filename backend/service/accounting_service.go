package service

import (
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type AccountingService struct {
	Store    *store.AccountingStore
	Settings *SettingsService
}

func NewAccountingService(s *store.AccountingStore, settings *SettingsService) *AccountingService {
	return &AccountingService{Store: s, Settings: settings}
}

func (s *AccountingService) GetInventoryValuation() (*models.InventoryValuation, error) {
	settings, err := s.Settings.GetSettings()
	if err != nil {
		return nil, err
	}
	return s.Store.GetInventoryValuation(settings.USDToCOPRate, settings.EURToCOPRate)
}
