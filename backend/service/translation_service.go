package service

import (
	"context"
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type TranslationService struct {
	Store *store.TranslationStore
}

func NewTranslationService(s *store.TranslationStore) *TranslationService {
	return &TranslationService{Store: s}
}

func (s *TranslationService) GetByLocale(ctx context.Context, locale string) (map[string]string, error) {
	return s.Store.GetByLocale(ctx, locale)
}

func (s *TranslationService) GetAll(ctx context.Context) (map[string]map[string]string, error) {
	return s.Store.GetAll(ctx)
}

func (s *TranslationService) ListKeys(ctx context.Context) ([]models.Translation, error) {
	return s.Store.ListKeys(ctx)
}

func (s *TranslationService) Upsert(ctx context.Context, t models.Translation) error {
	if t.Key == "" || t.Locale == "" || t.Value == "" {
		return fmt.Errorf("key, locale, and value are required")
	}
	return s.Store.Upsert(ctx, t)
}

func (s *TranslationService) Delete(ctx context.Context, key, locale string) error {
	return s.Store.Delete(ctx, key, locale)
}

func (s *TranslationService) DeleteLocale(ctx context.Context, locale string) error {
	if locale == "en" || locale == "es" {
		return fmt.Errorf("cannot delete system protected locales (en/es)")
	}
	return s.Store.DeleteLocale(ctx, locale)
}
