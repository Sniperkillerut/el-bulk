package service

import (
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

func (s *TranslationService) GetByLocale(locale string) (map[string]string, error) {
	return s.Store.GetByLocale(locale)
}

func (s *TranslationService) GetAll() (map[string]map[string]string, error) {
	return s.Store.GetAll()
}

func (s *TranslationService) ListKeys() ([]models.Translation, error) {
	return s.Store.ListKeys()
}

func (s *TranslationService) Upsert(t models.Translation) error {
	if t.Key == "" || t.Locale == "" || t.Value == "" {
		return fmt.Errorf("key, locale, and value are required")
	}
	return s.Store.Upsert(t)
}

func (s *TranslationService) Delete(key, locale string) error {
	return s.Store.Delete(key, locale)
}

func (s *TranslationService) DeleteLocale(locale string) error {
	if locale == "en" || locale == "es" {
		return fmt.Errorf("cannot delete system protected locales (en/es)")
	}
	return s.Store.DeleteLocale(locale)
}
