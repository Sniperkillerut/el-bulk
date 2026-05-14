package store

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/cache"
	"github.com/jmoiron/sqlx"
	"time"
)

const (
	translationAllCacheKey = "all_translations"
	translationLocalePrefix = "translation_locale_"
)

type TranslationStore struct {
	DB    *sqlx.DB
	cache *cache.Cache[interface{}]
}

func NewTranslationStore(db *sqlx.DB) *TranslationStore {
	return &TranslationStore{
		DB:    db,
		cache: cache.New[interface{}](),
	}
}

func (s *TranslationStore) GetByLocale(ctx context.Context, locale string) (map[string]string, error) {
	cacheKey := translationLocalePrefix + locale
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(map[string]string), nil
	}

	translations, err := models.GetTranslationsByLocale(ctx, s.DB, locale)
	if err == nil {
		s.cache.Set(cacheKey, translations, 5*time.Minute)
	}
	return translations, err
}

func (s *TranslationStore) GetAll(ctx context.Context) (map[string]map[string]string, error) {
	if cached, found := s.cache.Get(translationAllCacheKey); found {
		return cached.(map[string]map[string]string), nil
	}

	translations, err := models.GetAllTranslations(ctx, s.DB)
	if err == nil {
		s.cache.Set(translationAllCacheKey, translations, 5*time.Minute)
	}
	return translations, err
}

func (s *TranslationStore) ListKeys(ctx context.Context) ([]models.Translation, error) {
	return models.ListAllTranslationKeys(ctx, s.DB)
}

func (s *TranslationStore) Upsert(ctx context.Context, t models.Translation) error {
	err := models.UpsertTranslation(ctx, s.DB, t)
	if err == nil {
		s.cache.Flush()
	}
	return err
}

func (s *TranslationStore) Delete(ctx context.Context, key, locale string) error {
	err := models.DeleteTranslation(ctx, s.DB, key, locale)
	if err == nil {
		s.cache.Flush()
	}
	return err
}

func (s *TranslationStore) DeleteLocale(ctx context.Context, locale string) error {
	err := models.DeleteLocale(ctx, s.DB, locale)
	if err == nil {
		s.cache.Flush()
	}
	return err
}
