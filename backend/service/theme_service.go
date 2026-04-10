package service

import (
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type ThemeService struct {
	Store *store.ThemeStore
}

func NewThemeService(s *store.ThemeStore) *ThemeService {
	return &ThemeService{Store: s}
}

func (s *ThemeService) List() ([]models.Theme, error) {
	logger.Trace("Entering ThemeService.List")
	return s.Store.List()
}

func (s *ThemeService) Create(input models.ThemeInput) (*models.Theme, error) {
	logger.Trace("Entering ThemeService.Create | Name: %s", input.Name)
	return s.Store.Create(input)
}

func (s *ThemeService) Update(id string, input models.ThemeInput) (*models.Theme, error) {
	logger.Trace("Entering ThemeService.Update | ID: %s", id)
	return s.Store.Update(id, input)
}

func (s *ThemeService) Delete(id string) error {
	logger.Trace("Entering ThemeService.Delete | ID: %s", id)
	// Check if system theme
	isSystem, err := s.Store.IsSystemTheme(id)
	if err != nil {
		return err
	}
	if isSystem {
		return fmt.Errorf("cannot delete system theme")
	}

	return s.Store.BaseStore.Delete(id)
}
