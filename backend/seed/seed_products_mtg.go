package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// seedMTGSingles fetches real card data from Scryfall and inserts
// into the product table with exhaustive field coverage.
func seedMTGSingles(db *sqlx.DB, cats CategoryMap, storage StorageMap) ([]string, error) {
	logger.Info("🔮 Seeding MTG Singles via Scryfall...")

	// Curated list: Modern/Legacy staples, Alpha duals, Commander bombs, new Standard
	identifiers := []external.CardIdentifier{
		// ── Alpha/Beta/Unlimited Power Cards ─────────────────────
		{Name: "Black Lotus", SetCode: "lea"},
		{Name: "Ancestral Recall", SetCode: "lea"},
		{Name: "Time Walk", SetCode: "lea"},
		{Name: "Mox Pearl", SetCode: "lea"},
		{Name: "Mox Emerald", SetCode: "lea"},

		// ── Dual Lands (Revised) ─────────────────────────────────
		{Name: "Volcanic Island", SetCode: "rev"},
		{Name: "Underground Sea", SetCode: "rev"},
		{Name: "Tropical Island", SetCode: "rev"},
		{Name: "Tundra", SetCode: "rev"},
		{Name: "Badlands", SetCode: "rev"},
		{Name: "Taiga", SetCode: "rev"},
		{Name: "Savannah", SetCode: "rev"},
		{Name: "Scrubland", SetCode: "rev"},
		{Name: "Bayou", SetCode: "rev"},
		{Name: "Plateau", SetCode: "rev"},

		// ── Modern Staples ───────────────────────────────────────
		{Name: "Ragavan, Nimble Pilferer", SetCode: "mh2"},
		{Name: "Orcish Bowmasters", SetCode: "ltr"},
		{Name: "The One Ring", SetCode: "ltr"},
		{Name: "Sheoldred, the Apocalypse", SetCode: "dmu"},
		{Name: "Tarmogoyf", SetCode: "fut"},
		{Name: "Snapcaster Mage", SetCode: "isd"},
		{Name: "Liliana of the Veil", SetCode: "isd"},
		{Name: "Thoughtseize", SetCode: "ths"},
		{Name: "Force of Will", SetCode: "all"},
		{Name: "Brainstorm", SetCode: "mmq"},
		{Name: "Dark Ritual", SetCode: "lea"},

		// ── Commander Staples ─────────────────────────────────────
		{Name: "Sol Ring", SetCode: "v10"},
		{Name: "Mana Crypt", SetCode: "2xm"},
		{Name: "Rhystic Study", SetCode: "pcy"},
		{Name: "Smothering Tithe", SetCode: "rna"},
		{Name: "Cyclonic Rift", SetCode: "rtr"},
		{Name: "Dockside Extortionist", SetCode: "c19"},
		{Name: "Thassa's Oracle", SetCode: "thb"},
		{Name: "Demonic Tutor", SetCode: "vma"},
		{Name: "Vampiric Tutor", SetCode: "cmr"},

		// ── Fetch Lands ───────────────────────────────────────────
		{Name: "Scalding Tarn", SetCode: "zen"},
		{Name: "Misty Rainforest", SetCode: "zen"},
		{Name: "Verdant Catacombs", SetCode: "zen"},
		{Name: "Arid Mesa", SetCode: "zen"},
		{Name: "Marsh Flats", SetCode: "zen"},

		// ── Shock Lands ───────────────────────────────────────────
		{Name: "Steam Vents", SetCode: "grn"},
		{Name: "Hallowed Fountain", SetCode: "rna"},
		{Name: "Stomping Ground", SetCode: "rna"},
		{Name: "Breeding Pool", SetCode: "rna"},
		{Name: "Blood Crypt", SetCode: "rna"},

		// ── Artifacts / Moxen ────────────────────────────────────
		{Name: "Mox Opal", SetCode: "som"},
		{Name: "Chrome Mox", SetCode: "mrd"},
		{Name: "Mox Amber", SetCode: "dom"},
		{Name: "Grim Monolith", SetCode: "ulg"},
		{Name: "Lion's Eye Diamond", SetCode: "mir"},

		// ── Legendary & Historic ─────────────────────────────────
		{Name: "Jace, the Mind Sculptor", SetCode: "wwk"},
		{Name: "Liliana, the Last Hope", SetCode: "emn"},
		{Name: "Wrenn and Six", SetCode: "mh1"},
		{Name: "Teferi, Time Raveler", SetCode: "war"},
		{Name: "Urza, Lord High Artificer", SetCode: "mh1"},

		// ── Special Lands ─────────────────────────────────────────
		{Name: "Gaea's Cradle", SetCode: "usg"},
		{Name: "Serra's Sanctum", SetCode: "usg"},
		{Name: "Tolarian Academy", SetCode: "usg"},
		{Name: "Ancient Tomb", SetCode: "tmp"},
		{Name: "Cavern of Souls", SetCode: "avr"},

		// ── Standard Staples (Recent) ───────────────────────────
		{Name: "Atraxa, Praetors' Voice", SetCode: "one"},
		{Name: "Phyrexian Obliterator", SetCode: "one"},
		{Name: "Elesh Norn, Mother of Machines", SetCode: "one"},
		{Name: "The Wandering Emperor", SetCode: "neo"},
		{Name: "Kaito Shizuki", SetCode: "neo"},

		// ── Full Art & Textless ──────────────────────────────────
		{Name: "Mountain", SetCode: "unf", CollectorNumber: "242"},
		{Name: "Forest", SetCode: "ust", CollectorNumber: "216"},
		{Name: "Cryptic Command", SetCode: "p09", CollectorNumber: "1"},
		{Name: "Lightning Bolt", SetCode: "p11", CollectorNumber: "1"},
		{Name: "Hallowed Fountain", SetCode: "exp", CollectorNumber: "16"},
	}

	// All treatments, conditions, foils, languages, price sources to rotate through
	foilTreatments := []models.FoilTreatment{
		models.FoilNonFoil, models.FoilFoil, models.FoilEtchedFoil,
		models.FoilSurgeFoil, models.FoilTexturedFoil, models.FoilGalaxyFoil,
	}
	cardTreatments := []models.CardTreatment{
		models.TreatmentNormal, models.TreatmentShowcase,
		models.TreatmentBorderless, models.TreatmentExtendedArt,
		models.TreatmentFullArt,
	}
	conditions := []string{"NM", "LP", "MP", "HP", "DMG"}
	languages := []string{"en", "en", "en", "es", "ja", "pt"} // weighted toward English
	priceSources := []struct {
		source string
		ref    float64 // USD or EUR reference
	}{
		{"manual", 0},        // will use price_cop_override
		{"tcgplayer", 12.50}, // USD
		{"tcgplayer", 45.00},
		{"cardmarket", 8.75}, // EUR
		{"manual", 0},
	}

	storageLocs := []string{
		"Showcase A", "Showcase B", "Binder Vault", "Storage Box 1", "Storage Box 2",
	}

	catKeys := []string{"featured", "hot-items", "new-arrivals", "sale", "commander-staples", "tournament-ready", "staff-picks"}

	logger.Info("Fetching bulk Scryfall metadata for %d cards...", len(identifiers))
	var results []external.CardLookupResult

	chunkSize := 20
	for i := 0; i < len(identifiers); i += chunkSize {
		end := i + chunkSize
		if end > len(identifiers) {
			end = len(identifiers)
		}
		chunk := identifiers[i:end]

		var res []external.CardLookupResult
		var err error
		for attempt := 1; attempt <= 3; attempt++ {
			res, err = external.BatchLookupMTGCard(context.Background(), chunk)
			if err == nil {
				break
			}
			logger.Warn("  ⚠️ Attempt %d failed for chunk at index %d: %v", attempt, i, err)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			logger.Error("  ❌ Batch failed after 3 attempts: %v", err)
			continue
		}
		results = append(results, res...)
		time.Sleep(150 * time.Millisecond) // rate limit
	}

	var productIDs []string
	logger.Info("🌱 Generating 1000+ variations from %d unique cards...", len(results))
	
	variantCount := 0
	for _, res := range results {
		// Create 8-15 variations for every unique card to hit the 1000+ target
		numVariants := randInt(8, 15)
		
		for v := 0; v < numVariants; v++ {
			variantCount++
			foil := foilTreatments[randInt(0, len(foilTreatments)-1)]
			treat := cardTreatments[randInt(0, len(cardTreatments)-1)]
			cond := conditions[randInt(0, len(conditions)-1)]
			lang := languages[randInt(0, len(languages)-1)]
			ps := priceSources[randInt(0, len(priceSources)-1)]
			
			stock := randInt(0, 15) // Include some out of stock for filter testing
			createdAt := daysAgo(randInt(1, 120))
			costBasis := float64(randInt(5, 50)) * 1000.0
	
			// Determine pricing
			var priceRef *float64
			var priceCOPOverride *float64
			if ps.source == "manual" {
				val := float64(randInt(5, 1200)) * 1000.0
				priceCOPOverride = &val
			} else {
				ref := ps.ref * (0.7 + rand.Float64()*0.8) // High variance for price sorting test
				priceRef = &ref
			}
	
			var pID string
			err := db.Get(&pID, `
				INSERT INTO product (
					name, tcg, category, set_name, set_code, collector_number, condition,
					foil_treatment, card_treatment, language, price_source, price_reference,
					price_cop_override, image_url, stock, cost_basis_cop,
					rarity, is_legendary, is_historic, is_land, is_basic_land,
					art_variation, oracle_text, artist, type_line, border_color, frame,
					full_art, textless, promo_type, cmc, color_identity, scryfall_id, legalities,
					card_types, created_at
				) VALUES (
					$1, 'mtg', 'singles', $2, $3, $4, $5, $6, $7, $8, $9, $10,
					$11, $12, $13, $14,
					$15, $16, $17, $18, $19,
					$20, $21, $22, $23, $24, $25,
					$26, $27, $28, $29, $30, $31, $32,
					$33, $34
				) RETURNING id
			`,
				res.Name, res.SetName, res.SetCode, res.CollectorNumber, cond,
				foil, treat, lang, ps.source, priceRef,
				priceCOPOverride, res.ImageURL, stock, costBasis,
				res.Rarity, res.IsLegendary, res.IsHistoric, res.IsLand, res.IsBasicLand,
				res.ArtVariation, res.OracleText, res.Artist, res.TypeLine, res.BorderColor, res.Frame,
				res.FullArt, res.Textless, res.PromoType, res.CMC, res.ColorIdentity, res.ScryfallID, res.Legalities,
				res.CardTypes, createdAt,
			)
	
			if err != nil {
				logger.Warn("  ⚠️ Variation insert failed for '%s': %v", res.Name, err)
				continue
			}
			productIDs = append(productIDs, pID)
	
			// Distribute stock across storage locations
			if stock > 0 {
				loc1 := storageLocs[randInt(0, len(storageLocs)-1)]
				if sid, ok := storage[loc1]; ok {
					db.Exec(`
						INSERT INTO product_storage (product_id, storage_id, quantity)
						VALUES ($1, $2, $3)
						ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity
					`, pID, sid, stock)
				}
			}
	
			// Assign categories
			primaryCat := catKeys[randInt(0, len(catKeys)-1)]
			if catID, ok := cats[primaryCat]; ok {
				db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
			}
			if variantCount%5 == 0 {
				if catID, ok := cats["featured"]; ok {
					db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
				}
			}
		}
	}

	logger.Info("✅ %d MTG singles seeded", len(productIDs))
	return productIDs, nil
}

// seedMTGSealed inserts sealed MTG product with real set data.
func seedMTGSealed(db *sqlx.DB, cats CategoryMap, storage StorageMap) ([]string, error) {
	logger.Info("📦 Seeding MTG Sealed Products...")

	type SealedItem struct {
		Name        string
		SetName     string
		SetCode     string
		Price       float64
		Stock       int
		ImageURL    string
		Description string
		CreatedAt   time.Time
	}

	items := []SealedItem{
		{
			"Outlaws of Thunder Junction Play Booster Box",
			"Outlaws of Thunder Junction", "OTJ", 720000, 6,
			"https://cards.scryfall.io/art_crop/front/4/d/4d1c73c0-8c5b-4af6-a2df-bad1e20c18d7.jpg",
			"36 Play Boosters. Designed for both draft and enjoyment. Includes cards from the Breaking News series.",
			daysAgo(5),
		},
		{
			"Modern Horizons 3 Collector Booster Box",
			"Modern Horizons 3", "MH3", 1950000, 3,
			"https://cards.scryfall.io/art_crop/front/0/3/0337a311-2c41-4438-b70e-6b87ba04cc3c.jpg",
			"12 Collector Boosters packed with exclusive treatments and serialized cards.",
			daysAgo(15),
		},
		{
			"Murders at Karlov Manor Commander Deck - Blame Game",
			"Murders at Karlov Manor Commander", "MKC", 185000, 4,
			"https://cards.scryfall.io/art_crop/front/2/2/22da6cc8-13c1-4ae3-9023-b29b9db5d8fc.jpg",
			"100-card Commander deck based on the Murders at Karlov Manor set.",
			daysAgo(10),
		},
		{
			"The Lord of the Rings: Tales of Middle-earth Bundle",
			"The Lord of the Rings: Tales of Middle-earth", "LTR", 280000, 5,
			"https://cards.scryfall.io/art_crop/front/b/3/b3a2c3cb-fe8d-4c5e-9e00-1e45c4d3acf2.jpg",
			"8 Play Boosters + 1 Collector Booster + scene card special promo.",
			daysAgo(25),
		},
		{
			"Bloomburrow Play Booster Box",
			"Bloomburrow", "BLB", 680000, 8,
			"https://cards.scryfall.io/art_crop/front/3/f/3f5ac5af-10a4-4afe-a76c-5e9eb5fdc06b.jpg",
			"36 Play Boosters from the beloved critter-themed set.",
			daysAgo(3),
		},
		{
			"Foundations Play Booster Box",
			"Magic: The Gathering Foundations", "FDN", 590000, 10,
			"https://cards.scryfall.io/art_crop/front/7/4/74e07aec-c6f6-4de1-823d-cae08f1b4c9f.jpg",
			"The perfect entry point for new and returning players. 36 Play Boosters.",
			daysAgo(7),
		},
		{
			"Innistrad: Midnight Hunt Draft Booster Box",
			"Innistrad: Midnight Hunt", "MID", 380000, 2,
			"https://cards.scryfall.io/art_crop/front/c/1/c180af88-6a94-443b-8a15-1c52e5b0ac08.jpg",
			"36 Draft Boosters from the spooky Innistrad set.",
			daysAgo(45),
		},
	}

	var ids []string
	storeLocs := []string{"Showcase A", "Storage Box 1", "Storage Box 2"}
	sealedCats := []string{"new-arrivals", "featured", "hot-items"}

	for i, item := range items {
		var pID string
		err := db.QueryRow(`
			INSERT INTO product (
				name, tcg, category, set_name, set_code, price_source, price_cop_override,
				stock, image_url, description, cost_basis_cop, created_at
			) VALUES ($1, 'mtg', 'sealed', $2, $3, 'manual', $4, $5, $6, $7, $8, $9)
			RETURNING id
		`,
			item.Name, item.SetName, item.SetCode, item.Price,
			item.Stock, item.ImageURL, item.Description,
			item.Price*0.65, // ~35% margin
			item.CreatedAt,
		).Scan(&pID)
		if err != nil {
			return nil, fmt.Errorf("failed to seed MTG sealed '%s': %w", item.Name, err)
		}
		ids = append(ids, pID)

		loc := storeLocs[i%len(storeLocs)]
		if sid, ok := storage[loc]; ok {
			db.Exec(`
				INSERT INTO product_storage (product_id, storage_id, quantity)
				VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
			`, pID, sid, item.Stock)
		}
		cat := sealedCats[i%len(sealedCats)]
		if catID, ok := cats[cat]; ok {
			db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
		}
	}

	logger.Info("✅ %d MTG sealed products seeded", len(ids))
	return ids, nil
}
