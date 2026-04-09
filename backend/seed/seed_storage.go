package main

import (
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// StorageMap maps location name → UUID
type StorageMap map[string]string

func seedStorage(db *sqlx.DB) StorageMap {
	logger.Info("📦 Seeding storage locations...")

	locations := []string{
		"pending",        // system: order flow
		"Showcase A",     // front display cabinet
		"Showcase B",     // side display cabinet
		"Storage Box 1",  // back storage area
		"Storage Box 2",
		"Storage Box 3",
		"Binder Vault",   // premium binders
		"Counter Display",// behind the counter
		"Bulk Bin",       // inexpensive bulk cards
	}

	result := make(StorageMap)
	for _, loc := range locations {
		var id string
		err := db.QueryRow(`
			INSERT INTO storage_location (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, loc).Scan(&id)
		if err != nil {
			logger.Error("Failed to seed storage '%s': %v", loc, err)
			continue
		}
		result[loc] = id
	}
	logger.Info("✅ %d storage locations seeded", len(result))
	return result
}
