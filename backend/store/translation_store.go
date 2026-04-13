package store

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type TranslationStore struct {
	DB *sqlx.DB
}

func NewTranslationStore(db *sqlx.DB) *TranslationStore {
	return &TranslationStore{DB: db}
}

func (s *TranslationStore) GetByLocale(ctx context.Context, locale string) (map[string]string, error) {
	return models.GetTranslationsByLocale(ctx, s.DB, locale)
}

func (s *TranslationStore) GetAll(ctx context.Context) (map[string]map[string]string, error) {
	return models.GetAllTranslations(ctx, s.DB)
}

func (s *TranslationStore) ListKeys(ctx context.Context) ([]models.Translation, error) {
	return models.ListAllTranslationKeys(ctx, s.DB)
}

func (s *TranslationStore) Upsert(ctx context.Context, t models.Translation) error {
	return models.UpsertTranslation(ctx, s.DB, t)
}

func (s *TranslationStore) Delete(ctx context.Context, key, locale string) error {
	return models.DeleteTranslation(ctx, s.DB, key, locale)
}

func (s *TranslationStore) DeleteLocale(ctx context.Context, locale string) error {
	return models.DeleteLocale(ctx, s.DB, locale)
}
