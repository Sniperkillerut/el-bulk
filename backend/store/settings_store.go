package store

import (
	"context"
	"github.com/el-bulk/backend/utils/cache"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"time"
)

const settingsCacheKey = "all_settings"

type SettingsStore struct {
	DB    *sqlx.DB
	cache *cache.Cache[map[string]string]
}

func NewSettingsStore(db *sqlx.DB) *SettingsStore {
	return &SettingsStore{
		DB:    db,
		cache: cache.New[map[string]string](),
	}
}

func (s *SettingsStore) GetAll(ctx context.Context) (map[string]string, error) {
	// Try cache first
	if cached, found := s.cache.Get(settingsCacheKey); found {
		return cached, nil
	}

	start := time.Now()
	query := "SELECT key, value FROM setting"
	logger.TraceCtx(ctx, "[DB] Executing GetAll Settings (Cache Miss): %s", query)
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

	// Store in cache for 5 minutes
	s.cache.Set(settingsCacheKey, settings, 5*time.Minute)

	logger.DebugCtx(ctx, "[DB] GetAll Settings took %v", time.Since(start))
	return settings, rows.Err()
}

func (s *SettingsStore) InvalidateCache() {
	s.cache.Delete(settingsCacheKey)
}

func (s *SettingsStore) Upsert(ctx context.Context, key, value string) error {
	query := "INSERT INTO setting(key, value) VALUES($1, $2) ON CONFLICT(key) DO UPDATE SET value = $2"
	logger.TraceCtx(ctx, "[DB] Executing Upsert Setting: %s | Key: %s", query, key)
	_, err := s.DB.ExecContext(ctx, query, key, value)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] Upsert Setting failed for key %s: %v", key, err)
	} else {
		// Invalidate cache on change
		s.cache.Delete(settingsCacheKey)
	}
	return err
}
