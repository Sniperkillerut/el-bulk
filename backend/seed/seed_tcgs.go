package main

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

//go:embed seed_ck_names.sql
var ckNamesSQL string

func seedTCGs(db *sqlx.DB) error {
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
			return fmt.Errorf("failed to seed TCG '%s': %w", t.ID, err)
		}
	}
	logger.Info("✅ %d TCGs seeded", len(tcgs))
	return nil
}

func seedSets(db *sqlx.DB) error {
	logger.Info("🔭 Syncing MTG Sets from Scryfall...")
	sets, err := external.FetchSets(context.Background())
	if err != nil {
		logger.Warn("⚠️ Failed to fetch sets for seeding (non-fatal): %v", err)
		return nil
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction for sets: %w", err)
	}

	for _, s := range sets {
		// Use "best guess" logic for initial CK name
		ckGuess := external.NormalizeCKEdition(s.Name)

		_, err := tx.Exec(`
			INSERT INTO tcg_set (tcg, code, name, released_at, set_type, ck_name)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (tcg, code) DO UPDATE SET
				name = EXCLUDED.name,
				released_at = EXCLUDED.released_at,
				set_type = EXCLUDED.set_type,
				-- Preserve manual overrides, otherwise use EXCLUDED
				ck_name = COALESCE(tcg_set.ck_name, EXCLUDED.ck_name)
		`, "mtg", s.Code, s.Name, s.ReleasedAt, s.SetType, ckGuess)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to seed set %s: %w", s.Code, err)
		}
	}

	_, err = tx.Exec(`
		INSERT INTO setting (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
	`, "last_set_sync", time.Now().Format(time.RFC3339))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update last_set_sync: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit sets: %w", err)
	}

	logger.Info("✅ %d MTG sets synchronized", len(sets))
	return nil
}

func seedCKMappings(db *sqlx.DB) error {
	logger.Info("🧹 Applying manual Card Kingdom set name overrides...")
	if ckNamesSQL == "" {
		logger.Warn("⚠️ No CK name overrides found in embedded SQL")
		return nil
	}

	_, err := db.Exec(ckNamesSQL)
	if err != nil {
		return fmt.Errorf("failed to apply CK name overrides: %w", err)
	}

	logger.Info("✅ Card Kingdom set name mappings applied")
	return nil
}
