package main

import (
	"time"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedTCGs(db *sqlx.DB) {
	logger.Info("🎮 Seeding TCG systems...")

	type TCG struct {
		ID       string
		Name     string
		ImageURL string
		IsActive bool
	}

	tcgs := []TCG{
		{"mtg", "Magic: The Gathering", "/tcgs/mtg_banner.png", true},
		{"pokemon", "Pokémon", "/tcgs/pokemon_banner.png", true},
		{"yugioh", "Yu-Gi-Oh!", "/tcgs/yugioh_banner.png", true},
		{"lorcana", "Disney Lorcana", "/tcgs/lorcana_banner.png", true},
		{"onepiece", "One Piece", "/tcgs/one_piece_banner.png", true},
		{"starwars", "Star Wars Unlimited", "/tcgs/starwars_banner.png", true},
		{"weiss", "Weiss Schwarz", "/tcgs/weiss_banner.png", false}, // inactive - test toggle
	}

	for _, t := range tcgs {
		_, err := db.Exec(`
			INSERT INTO tcg (id, name, image_url, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				image_url = EXCLUDED.image_url,
				is_active = EXCLUDED.is_active
		`, t.ID, t.Name, t.ImageURL, t.IsActive)
		if err != nil {
			logger.Error("Failed to seed TCG '%s': %v", t.ID, err)
		}
	}
	logger.Info("✅ %d TCGs seeded", len(tcgs))
}

func seedSets(db *sqlx.DB) {
	logger.Info("🔭 Syncing MTG Sets from Scryfall...")
	sets, err := external.FetchSets()
	if err != nil {
		logger.Warn("⚠️ Failed to fetch sets for seeding: %v", err)
		return
	}

	tx, _ := db.Beginx()
	for _, s := range sets {
		tx.Exec(`
			INSERT INTO tcg_set (tcg, code, name, released_at, set_type)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tcg, code) DO UPDATE SET
				name = EXCLUDED.name,
				released_at = EXCLUDED.released_at,
				set_type = EXCLUDED.set_type
		`, "mtg", s.Code, s.Name, s.ReleasedAt, s.SetType)
	}
	tx.Exec(`
		INSERT INTO setting (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
	`, "last_set_sync", time.Now().Format(time.RFC3339))
	tx.Commit()
	logger.Info("✅ %d MTG sets synchronized", len(sets))
}
