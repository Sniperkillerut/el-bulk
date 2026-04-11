package store

import (
	"github.com/jmoiron/sqlx"
)

type HealthStore struct {
	DB *sqlx.DB
}

func NewHealthStore(db *sqlx.DB) *HealthStore {
	return &HealthStore{DB: db}
}

func (s *HealthStore) Ping() error {
	return s.DB.Ping()
}

func (s *HealthStore) GetDatabaseSize() (string, error) {
	var size string
	err := s.DB.Get(&size, "SELECT pg_size_pretty(pg_database_size(current_database()))")
	return size, err
}

func (s *HealthStore) GetCacheHitRatio() (float64, error) {
	var ratio float64
	err := s.DB.Get(&ratio, `
		SELECT 
			CASE 
				WHEN (blks_hit + blks_read) = 0 THEN 0 
				ELSE ROUND(CAST(blks_hit AS NUMERIC) / (blks_hit + blks_read) * 100, 2)
			END
		FROM pg_stat_database 
		WHERE datname = current_database()`)
	return ratio, err
}

func (s *HealthStore) GetActiveConnections() (int, error) {
	var count int
	err := s.DB.Get(&count, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database()")
	return count, err
}

func (s *HealthStore) GetProductCount() (int, error) {
	var count int
	err := s.DB.Get(&count, "SELECT COUNT(*) FROM product")
	return count, err
}

func (s *HealthStore) GetPendingOrdersCount() (int, error) {
	var count int
	err := s.DB.Get(&count, "SELECT COUNT(*) FROM \"order\" WHERE status = 'pending'")
	return count, err
}

func (s *HealthStore) GetPendingOffersCount() (int, error) {
	var count int
	err := s.DB.Get(&count, "SELECT COUNT(*) FROM bounty_offer WHERE status = 'pending'")
	return count, err
}

func (s *HealthStore) GetPendingRequestsCount() (int, error) {
	var count int
	err := s.DB.Get(&count, "SELECT COUNT(*) FROM client_request WHERE status = 'pending'")
	return count, err
}

type TranslationLocaleCount struct {
	Locale string `db:"locale"`
	Count  int    `db:"count"`
}

func (s *HealthStore) GetTotalTranslationKeys() (int, error) {
	var count int
	err := s.DB.Get(&count, "SELECT COUNT(DISTINCT key) FROM translation")
	return count, err
}

func (s *HealthStore) GetTranslationLocaleCounts() ([]TranslationLocaleCount, error) {
	var counts []TranslationLocaleCount
	err := s.DB.Select(&counts, "SELECT locale, COUNT(*) as count FROM translation GROUP BY locale")
	return counts, err
}

func (s *HealthStore) PingQuery() error {
	var dummy int
	return s.DB.Get(&dummy, "SELECT 1")
}

func (s *HealthStore) GetMaxOpenConns() int {
	return s.DB.Stats().MaxOpenConnections
}
