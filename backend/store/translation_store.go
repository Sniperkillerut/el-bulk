package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type TranslationStore struct {
	DB *sqlx.DB
}

func NewTranslationStore(db *sqlx.DB) *TranslationStore {
	return &TranslationStore{DB: db}
}

func (s *TranslationStore) GetByLocale(locale string) (map[string]string, error) {
	return models.GetTranslationsByLocale(s.DB, locale)
}

func (s *TranslationStore) GetAll() (map[string]map[string]string, error) {
	return models.GetAllTranslations(s.DB)
}

func (s *TranslationStore) ListKeys() ([]models.Translation, error) {
	return models.ListAllTranslationKeys(s.DB)
}

func (s *TranslationStore) Upsert(t models.Translation) error {
	return models.UpsertTranslation(s.DB, t)
}

func (s *TranslationStore) Delete(key, locale string) error {
	return models.DeleteTranslation(s.DB, key, locale)
}

func (s *TranslationStore) DeleteLocale(locale string) error {
	return models.DeleteLocale(s.DB, locale)
}
