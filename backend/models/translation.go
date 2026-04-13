package models

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Translation struct {
	Key       string    `json:"key" db:"key"`
	Locale    string    `json:"locale" db:"locale"`
	Value     string    `json:"value" db:"value"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func GetAllTranslations(ctx context.Context, db *sqlx.DB) (map[string]map[string]string, error) {
	query := `SELECT key, locale, value FROM translation`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Nested map: locale -> { key: value }
	translations := make(map[string]map[string]string)
	for rows.Next() {
		var key, locale, value string
		if err := rows.Scan(&key, &locale, &value); err != nil {
			return nil, err
		}
		if _, ok := translations[locale]; !ok {
			translations[locale] = make(map[string]string)
		}
		translations[locale][key] = value
	}
	return translations, nil
}

func GetTranslationsByLocale(ctx context.Context, db *sqlx.DB, locale string) (map[string]string, error) {
	query := `SELECT key, value FROM translation WHERE locale = $1`
	rows, err := db.QueryContext(ctx, query, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	translations := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		translations[key] = value
	}
	return translations, nil
}

func UpsertTranslation(ctx context.Context, db *sqlx.DB, t Translation) error {
	query := `
		INSERT INTO translation (key, locale, value, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (key, locale) DO UPDATE
		SET value = EXCLUDED.value, updated_at = now()
	`
	_, err := db.ExecContext(ctx, query, t.Key, t.Locale, t.Value)
	return err
}

func DeleteTranslation(ctx context.Context, db *sqlx.DB, key, locale string) error {
	query := `DELETE FROM translation WHERE key = $1 AND locale = $2`
	_, err := db.ExecContext(ctx, query, key, locale)
	return err
}

func ListAllTranslationKeys(ctx context.Context, db *sqlx.DB) ([]Translation, error) {
	query := `SELECT key, locale, value, updated_at FROM translation ORDER BY key, locale`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []Translation
	for rows.Next() {
		var t Translation
		if err := rows.Scan(&t.Key, &t.Locale, &t.Value, &t.UpdatedAt); err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}
	return translations, nil
}

func DeleteLocale(ctx context.Context, db *sqlx.DB, locale string) error {
	query := `DELETE FROM translation WHERE locale = $1`
	_, err := db.ExecContext(ctx, query, locale)
	return err
}
