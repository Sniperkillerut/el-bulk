package store

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type HealthStore struct {
	DB *sqlx.DB
}

func NewHealthStore(db *sqlx.DB) *HealthStore {
	return &HealthStore{DB: db}
}

func (s *HealthStore) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *HealthStore) GetDatabaseSize(ctx context.Context) (string, error) {
	var size string
	err := s.DB.GetContext(ctx, &size, "SELECT pg_size_pretty(pg_database_size(current_database()))")
	return size, err
}

func (s *HealthStore) GetCacheHitRatio(ctx context.Context) (float64, error) {
	var ratio float64
	err := s.DB.GetContext(ctx, &ratio, `
		SELECT 
			CASE 
				WHEN (blks_hit + blks_read) = 0 THEN 0 
				ELSE ROUND(CAST(blks_hit AS NUMERIC) / (blks_hit + blks_read) * 100, 2)
			END
		FROM pg_stat_database 
		WHERE datname = current_database()`)
	return ratio, err
}

func (s *HealthStore) GetActiveConnections(ctx context.Context) (int, error) {
	var count int
	err := s.DB.GetContext(ctx, &count, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database()")
	return count, err
}

func (s *HealthStore) GetProductCount(ctx context.Context) (int, error) {
	var count int
	err := s.DB.GetContext(ctx, &count, "SELECT COUNT(*) FROM product")
	return count, err
}

func (s *HealthStore) GetPendingOrdersCount(ctx context.Context) (int, error) {
	var count int
	err := s.DB.GetContext(ctx, &count, "SELECT COUNT(*) FROM \"order\" WHERE status = 'pending'")
	return count, err
}

func (s *HealthStore) GetPendingOffersCount(ctx context.Context) (int, error) {
	var count int
	err := s.DB.GetContext(ctx, &count, "SELECT COUNT(*) FROM bounty_offer WHERE status = 'pending'")
	return count, err
}

func (s *HealthStore) GetPendingRequestsCount(ctx context.Context) (int, error) {
	var count int
	err := s.DB.GetContext(ctx, &count, "SELECT COUNT(*) FROM client_request WHERE status = 'pending'")
	return count, err
}

type TranslationLocaleCount struct {
	Locale string `db:"locale"`
	Count  int    `db:"count"`
}

func (s *HealthStore) GetTotalTranslationKeys(ctx context.Context) (int, error) {
	var count int
	err := s.DB.GetContext(ctx, &count, "SELECT COUNT(DISTINCT key) FROM translation")
	return count, err
}

func (s *HealthStore) GetTranslationLocaleCounts(ctx context.Context) ([]TranslationLocaleCount, error) {
	var counts []TranslationLocaleCount
	err := s.DB.SelectContext(ctx, &counts, "SELECT locale, COUNT(*) as count FROM translation GROUP BY locale")
	return counts, err
}

func (s *HealthStore) PingQuery(ctx context.Context) error {
	var dummy int
	return s.DB.GetContext(ctx, &dummy, "SELECT 1")
}

func (s *HealthStore) GetMaxOpenConns() int {
	return s.DB.Stats().MaxOpenConnections
}
