package service

import (
	"context"
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

func (s *ThemeService) List(ctx context.Context) ([]models.Theme, error) {
	logger.TraceCtx(ctx, "Entering ThemeService.List")
	return s.Store.List(ctx)
}

func (s *ThemeService) Create(ctx context.Context, input models.ThemeInput) (*models.Theme, error) {
	logger.TraceCtx(ctx, "Entering ThemeService.Create | Name: %s", input.Name)
	return s.Store.Create(ctx, input)
}

func (s *ThemeService) Update(ctx context.Context, id string, input models.ThemeInput) (*models.Theme, error) {
	logger.TraceCtx(ctx, "Entering ThemeService.Update | ID: %s", id)
	return s.Store.Update(ctx, id, input)
}

func (s *ThemeService) Delete(ctx context.Context, id string) error {
	logger.TraceCtx(ctx, "Entering ThemeService.Delete | ID: %s", id)
	// Check if system theme
	isSystem, err := s.Store.IsSystemTheme(ctx, id)
	if err != nil {
		return err
	}
	if isSystem {
		return fmt.Errorf("cannot delete system theme")
	}

	return s.Store.BaseStore.Delete(ctx, id)
}
