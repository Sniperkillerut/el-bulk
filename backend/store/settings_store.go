package store

import (
	"context"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"time"
)

type SettingsStore struct {
	DB *sqlx.DB
}

func NewSettingsStore(db *sqlx.DB) *SettingsStore {
	return &SettingsStore{DB: db}
}

func (s *SettingsStore) GetAll(ctx context.Context) (map[string]string, error) {
	start := time.Now()
	query := "SELECT key, value FROM setting"
	logger.TraceCtx(ctx, "[DB] Executing GetAll Settings: %s", query)
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetAll Settings failed: %v", err)
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
	logger.DebugCtx(ctx, "[DB] GetAll Settings took %v", time.Since(start))
	return settings, rows.Err()
}

func (s *SettingsStore) Upsert(ctx context.Context, key, value string) error {
	query := "INSERT INTO setting(key, value) VALUES($1, $2) ON CONFLICT(key) DO UPDATE SET value = $2"
	logger.TraceCtx(ctx, "[DB] Executing Upsert Setting: %s | Key: %s", query, key)
	_, err := s.DB.ExecContext(ctx, query, key, value)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] Upsert Setting failed for key %s: %v", key, err)
	}
	return err
}
