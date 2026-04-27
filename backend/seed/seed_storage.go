package main

import (
	"fmt"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// StorageMap maps location name → UUID
type StorageMap map[string]string

func seedStorage(db *sqlx.DB) (StorageMap, error) {
	logger.Info("📦 Seeding storage locations...")

	locations := []string{
		"pending",       // system: order flow
		"Showcase A",    // front display cabinet
		"Showcase B",    // side display cabinet
		"Storage Box 1", // back storage area
		"Storage Box 2",
		"Storage Box 3",
		"Binder Vault",    // premium binders
		"Counter Display", // behind the counter
		"Bulk Bin",        // inexpensive bulk cards
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
			return nil, fmt.Errorf("failed to seed storage '%s': %w", loc, err)
		}
		result[loc] = id
	}
	logger.Info("✅ %d storage locations seeded", len(result))
	return result, nil
}
func loadStorage(db *sqlx.DB) (StorageMap, error) {
	result := make(StorageMap)
	rows, err := db.Queryx("SELECT name, id FROM storage_location")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s struct {
			Name string `db:"name"`
			ID   string `db:"id"`
		}
		if err := rows.StructScan(&s); err != nil {
			return nil, err
		}
		result[s.Name] = s.ID
	}
	return result, nil
}
