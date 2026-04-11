package service

import (
	"time"

	"github.com/el-bulk/backend/store"
)

type HealthService struct {
	Store *store.HealthStore
}

func NewHealthService(s *store.HealthStore) *HealthService {
	return &HealthService{Store: s}
}

type DBStats struct {
	DatabaseSize           string                   `json:"database_size"`
	CacheHitRatio          float64                  `json:"cache_hit_ratio"`
	ActiveConns            int                      `json:"active_connections"`
	MaxConns               int                      `json:"max_connections"`
	TotalProducts          int                      `json:"total_products"`
	TotalSKURecords        int                      `json:"total_sku_records"`
	QuerySpeedMS           int                      `json:"query_speed_ms"`
	PendingOrdersCount     int                      `json:"pending_orders_count"`
	PendingOffersCount     int                      `json:"pending_offers_count"`
	PendingRequestsCount   int                      `json:"pending_requests_count"`
	TranslationProgress    []TranslationLocaleStats `json:"translation_progress"`
}

type TranslationLocaleStats struct {
	Locale      string  `json:"locale"`
	Completion  float64 `json:"completion"`
	MissingKeys int     `json:"missing_keys"`
}

func (s *HealthService) Ping() error {
	return s.Store.Ping()
}

func (s *HealthService) IsAvailable() bool {
	return s.Store != nil && s.Store.DB != nil
}

func (s *HealthService) GetStats() (*DBStats, error) {
	stats := &DBStats{}

	size, err := s.Store.GetDatabaseSize()
	if err == nil {
		stats.DatabaseSize = size
	}

	ratio, err := s.Store.GetCacheHitRatio()
	if err == nil {
		stats.CacheHitRatio = ratio
	}

	conns, err := s.Store.GetActiveConnections()
	if err == nil {
		stats.ActiveConns = conns
	} else {
		stats.ActiveConns = -1
	}

	stats.MaxConns = s.Store.GetMaxOpenConns()

	products, err := s.Store.GetProductCount()
	if err == nil {
		stats.TotalProducts = products
	}
	stats.TotalSKURecords = stats.TotalProducts

	pending, _ := s.Store.GetPendingOrdersCount()
	stats.PendingOrdersCount = pending

	offers, _ := s.Store.GetPendingOffersCount()
	stats.PendingOffersCount = offers

	requests, _ := s.Store.GetPendingRequestsCount()
	stats.PendingRequestsCount = requests

	// Query speed
	start := time.Now()
	_ = s.Store.PingQuery()
	stats.QuerySpeedMS = int(time.Since(start).Milliseconds())

	// Translation progress
	totalKeys, _ := s.Store.GetTotalTranslationKeys()
	if totalKeys > 0 {
		localeCounts, err := s.Store.GetTranslationLocaleCounts()
		if err == nil {
			for _, lc := range localeCounts {
				completion := (float64(lc.Count) / float64(totalKeys)) * 100
				stats.TranslationProgress = append(stats.TranslationProgress, TranslationLocaleStats{
					Locale:      lc.Locale,
					Completion:  completion,
					MissingKeys: totalKeys - lc.Count,
				})
			}
		}
	}
	if stats.TranslationProgress == nil {
		stats.TranslationProgress = []TranslationLocaleStats{}
	}

	return stats, nil
}
