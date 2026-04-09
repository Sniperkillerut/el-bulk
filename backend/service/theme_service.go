package service

import (
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type ThemeService struct {
	Store *store.ThemeStore
}

func NewThemeService(s *store.ThemeStore) *ThemeService {
	return &ThemeService{Store: s}
}

func (s *ThemeService) List() ([]models.Theme, error) {
	return s.Store.List()
}

func (s *ThemeService) Create(input models.ThemeInput) (*models.Theme, error) {
	return s.Store.Create(input)
}

func (s *ThemeService) Update(id string, input models.ThemeInput) (*models.Theme, error) {
	return s.Store.Update(id, input)
}

func (s *ThemeService) Delete(id string) error {
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
