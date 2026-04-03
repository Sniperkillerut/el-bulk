package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/el-bulk/backend/utils/logger"
)

type HealthHandler struct {
	DB *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{DB: db}
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

func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "message": "database connection unavailable"})
		return
	}

	err := h.DB.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *HealthHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.DB == nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "message": "database connection unavailable"})
		return
	}
	var stats DBStats

	// 1. Get database size
	err := h.DB.Get(&stats.DatabaseSize, "SELECT pg_size_pretty(pg_database_size(current_database()))")
	if err != nil {
		logger.Error("Stats Error (Size): %v", err)
	}

	// 2. Get cache hit ratio
	// Formula: blks_hit / (blks_hit + blks_read)
	err = h.DB.Get(&stats.CacheHitRatio, `
		SELECT 
			CASE 
				WHEN (blks_hit + blks_read) = 0 THEN 0 
				ELSE ROUND(CAST(blks_hit AS NUMERIC) / (blks_hit + blks_read) * 100, 2)
			END
		FROM pg_stat_database 
		WHERE datname = current_database()`)
	if err != nil {
		logger.Error("Stats Error (Cache): %v", err)
	}

	// 3. Connection info
	err = h.DB.Get(&stats.ActiveConns, "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database()")
	if err != nil {
		stats.ActiveConns = -1
	}
	
	stats.MaxConns = 25 // From db.go SetMaxOpenConns

	// 4. Record counts
	err = h.DB.Get(&stats.TotalProducts, "SELECT COUNT(*) FROM product")
	if err != nil {
		stats.TotalProducts = 0
	}
	stats.TotalSKURecords = stats.TotalProducts

	// 5. Pending items
	_ = h.DB.Get(&stats.PendingOrdersCount, "SELECT COUNT(*) FROM \"order\" WHERE status = 'pending'")
	_ = h.DB.Get(&stats.PendingOffersCount, "SELECT COUNT(*) FROM bounty_offer WHERE status = 'pending'")
	_ = h.DB.Get(&stats.PendingRequestsCount, "SELECT COUNT(*) FROM client_request WHERE status = 'pending'")

	// 6. Query Speed
	start := time.Now()
	var dummy int
	_ = h.DB.Get(&dummy, "SELECT 1")
	stats.QuerySpeedMS = int(time.Since(start).Milliseconds())

	// 7. Translation Progress
	var totalKeys int
	_ = h.DB.Get(&totalKeys, "SELECT COUNT(DISTINCT key) FROM translation")

	if totalKeys > 0 {
		var localeCounts []struct {
			Locale string `db:"locale"`
			Count  int    `db:"count"`
		}
		_ = h.DB.Select(&localeCounts, "SELECT locale, COUNT(*) as count FROM translation GROUP BY locale")

		for _, lc := range localeCounts {
			completion := (float64(lc.Count) / float64(totalKeys)) * 100
			stats.TranslationProgress = append(stats.TranslationProgress, TranslationLocaleStats{
				Locale:      lc.Locale,
				Completion:  completion,
				MissingKeys: totalKeys - lc.Count,
			})
		}
		// If totalKeys > 0 but a locale is missing entirely, we should ideally know about it. 
		// For now, we only report on locales that have at least one translation.
	} else {
		stats.TranslationProgress = []TranslationLocaleStats{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
