package store

import (
	"github.com/jmoiron/sqlx"
)

type SettingsStore struct {
	DB *sqlx.DB
}

func NewSettingsStore(db *sqlx.DB) *SettingsStore {
	return &SettingsStore{DB: db}
}

func (s *SettingsStore) GetAll() (map[string]string, error) {
	rows, err := s.DB.Query("SELECT key, value FROM setting")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			return nil, err
		}
		settings[key] = val
	}
	return settings, rows.Err()
}

func (s *SettingsStore) Upsert(key, value string) error {
	_, err := s.DB.Exec("INSERT INTO setting(key, value) VALUES($1, $2) ON CONFLICT(key) DO UPDATE SET value = $2", key, value)
	return err
}
