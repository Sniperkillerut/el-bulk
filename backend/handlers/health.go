package handlers

import (
	"encoding/json"
	"net/http"

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
	DatabaseSize    string  `json:"database_size"`
	CacheHitRatio   float64 `json:"cache_hit_ratio"`
	ActiveConns     int     `json:"active_connections"`
	MaxConns        int     `json:"max_connections"`
	TotalProducts   int     `json:"total_products"`
}

func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.DB.Ping()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *HealthHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
