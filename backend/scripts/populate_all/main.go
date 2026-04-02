package main

import (
	"fmt"
	"os"
	"time"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/utils/logger"
)

func main() {
	database := db.Connect()
	defer database.Close()

	logger.Info("🚀 Starting bulk metadata population for MTG...")

	// Fetch all MTG products
	var products []struct {
		ID            string `db:"id"`
		Name          string `db:"name"`
		SetCode       string `db:"set_code"`
		FoilTreatment string `db:"foil_treatment"`
	}

	err := database.Select(&products, "SELECT id, name, COALESCE(set_code, '') as set_code, foil_treatment FROM product WHERE tcg = 'mtg'")
	if err != nil {
		logger.Error("Failed to fetch products: %v", err)
		os.Exit(1)
	}

	logger.Info("Found %d MTG products to process.", len(products))

	updated := 0
	errors := 0

	for i, p := range products {
		fmt.Printf("[%d/%d] Processing %s...\n", i+1, len(products), p.Name)

		// Lookup full metadata
		meta, err := external.LookupMTGCard(p.Name, p.SetCode, "", p.FoilTreatment)
		if err != nil {
			logger.Warn("  ❌ Lookup failed for %s: %v", p.Name, err)
			errors++
			continue
		}

		// Update database
		_, err = database.Exec(`
			UPDATE product SET
				image_url = $1,
				set_name = $2,
				set_code = $3,
				collector_number = $4,
				language = $5,
				color_identity = $6,
				rarity = $7,
				cmc = $8,
				is_legendary = $9,
				is_historic = $10,
				is_land = $11,
				is_basic_land = $12,
				art_variation = $13,
				oracle_text = $14,
				artist = $15,
				type_line = $16,
				border_color = $17,
				frame = $18,
				full_art = $19,
				textless = $20,
				promo_type = $21,
				updated_at = NOW()
			WHERE id = $22
		`, 
		meta.ImageURL, 
		meta.SetName,
		meta.SetCode,
		meta.CollectorNumber,
		meta.Language,
		meta.ColorIdentity,
		meta.Rarity,
		meta.CMC,
		meta.IsLegendary,
		meta.IsHistoric,
		meta.IsLand,
		meta.IsBasicLand,
		meta.ArtVariation,
		meta.OracleText,
		meta.Artist,
		meta.TypeLine,
		meta.BorderColor,
		meta.Frame,
		meta.FullArt,
		meta.Textless,
		meta.PromoType,
		p.ID)

		if err != nil {
			logger.Error("  ❌ Update failed for %s: %v", p.Name, err)
			errors++
			continue
		}

		updated++
		// Respect rate limits softly
		time.Sleep(100 * time.Millisecond)
	}

	logger.Info("✅ Bulk population complete!")
	logger.Info("   Updated: %d", updated)
	logger.Info("   Errors:  %d", errors)
}
