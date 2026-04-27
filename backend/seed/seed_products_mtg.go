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
		{Name: "Black Lotus", Set: "lea"},
		{Name: "Ancestral Recall", Set: "lea"},
		{Name: "Time Walk", Set: "lea"},
		{Name: "Mox Pearl", Set: "lea"},
		{Name: "Mox Emerald", Set: "lea"},

		// ── Dual Lands (Revised) ─────────────────────────────────
		{Name: "Volcanic Island", Set: "rev"},
		{Name: "Underground Sea", Set: "rev"},
		{Name: "Tropical Island", Set: "rev"},
		{Name: "Tundra", Set: "rev"},
		{Name: "Badlands", Set: "rev"},
		{Name: "Taiga", Set: "rev"},
		{Name: "Savannah", Set: "rev"},
		{Name: "Scrubland", Set: "rev"},
		{Name: "Bayou", Set: "rev"},
		{Name: "Plateau", Set: "rev"},

		// ── Modern Staples ───────────────────────────────────────
		{Name: "Ragavan, Nimble Pilferer", Set: "mh2"},
		{Name: "Orcish Bowmasters", Set: "ltr"},
		{Name: "The One Ring", Set: "ltr"},
		{Name: "Sheoldred, the Apocalypse", Set: "dmu"},
		{Name: "Tarmogoyf", Set: "fut"},
		{Name: "Snapcaster Mage", Set: "isd"},
		{Name: "Liliana of the Veil", Set: "isd"},
		{Name: "Thoughtseize", Set: "ths"},
		{Name: "Force of Will", Set: "all"},
		{Name: "Brainstorm", Set: "mmq"},
		{Name: "Dark Ritual", Set: "lea"},

		// ── Commander Staples ─────────────────────────────────────
		{Name: "Sol Ring", Set: "v10"},
		{Name: "Mana Crypt", Set: "2xm"},
		{Name: "Rhystic Study", Set: "pcy"},
		{Name: "Smothering Tithe", Set: "rna"},
		{Name: "Cyclonic Rift", Set: "rtr"},
		{Name: "Dockside Extortionist", Set: "c19"},
		{Name: "Thassa's Oracle", Set: "thb"},
		{Name: "Demonic Tutor", Set: "vma"},
		{Name: "Vampiric Tutor", Set: "cmr"},

		// ── Fetch Lands ───────────────────────────────────────────
		{Name: "Scalding Tarn", Set: "zen"},
		{Name: "Misty Rainforest", Set: "zen"},
		{Name: "Verdant Catacombs", Set: "zen"},
		{Name: "Arid Mesa", Set: "zen"},
		{Name: "Marsh Flats", Set: "zen"},

		// ── Shock Lands ───────────────────────────────────────────
		{Name: "Steam Vents", Set: "grn"},
		{Name: "Hallowed Fountain", Set: "rna"},
		{Name: "Stomping Ground", Set: "rna"},
		{Name: "Breeding Pool", Set: "rna"},
		{Name: "Blood Crypt", Set: "rna"},

		// ── Artifacts / Moxen ────────────────────────────────────
		{Name: "Mox Opal", Set: "som"},
		{Name: "Chrome Mox", Set: "mrd"},
		{Name: "Mox Amber", Set: "dom"},
		{Name: "Grim Monolith", Set: "ulg"},
		{Name: "Lion's Eye Diamond", Set: "mir"},

		// ── Legendary & Historic ─────────────────────────────────
		{Name: "Jace, the Mind Sculptor", Set: "wwk"},
		{Name: "Liliana, the Last Hope", Set: "emn"},
		{Name: "Wrenn and Six", Set: "mh1"},
		{Name: "Teferi, Time Raveler", Set: "war"},
		{Name: "Urza, Lord High Artificer", Set: "mh1"},

		// ── Special Lands ─────────────────────────────────────────
		{Name: "Gaea's Cradle", Set: "usg"},
		{Name: "Serra's Sanctum", Set: "usg"},
		{Name: "Tolarian Academy", Set: "usg"},
		{Name: "Ancient Tomb", Set: "tmp"},
		{Name: "Cavern of Souls", Set: "avr"},

		// ── Standard Staples (Recent) ───────────────────────────
		{Name: "Atraxa, Praetors' Voice", Set: "one"},
		{Name: "Phyrexian Obliterator", Set: "one"},
		{Name: "Elesh Norn, Mother of Machines", Set: "one"},
		{Name: "The Wandering Emperor", Set: "neo"},
		{Name: "Kaito Shizuki", Set: "neo"},
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
	for i, res := range results {
		foil := foilTreatments[i%len(foilTreatments)]
		treat := cardTreatments[i%len(cardTreatments)]
		cond := conditions[i%len(conditions)]
		lang := languages[i%len(languages)]
		ps := priceSources[i%len(priceSources)]
		stock := randInt(1, 12)
		createdAt := daysAgo(randInt(1, 90))
		costBasis := float64(randInt(5, 50)) * 1000.0

		// Determine pricing
		var priceRef *float64
		var priceCOPOverride *float64
		if ps.source == "manual" {
			v := float64(randInt(10, 800)) * 1000.0
			priceCOPOverride = &v
		} else {
			ref := ps.ref * (0.8 + rand.Float64()*0.6) // ±20% variance
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
				created_at
			) VALUES (
				$1, 'mtg', 'singles', $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14,
				$15, $16, $17, $18, $19,
				$20, $21, $22, $23, $24, $25,
				$26, $27, $28, $29, $30, $31, $32,
				$33
			) RETURNING id
		`,
			res.Name, res.SetName, res.SetCode, res.CollectorNumber, cond,
			foil, treat, lang, ps.source, priceRef,
			priceCOPOverride, res.ImageURL, stock, costBasis,
			res.Rarity, res.IsLegendary, res.IsHistoric, res.IsLand, res.IsBasicLand,
			res.ArtVariation, res.OracleText, res.Artist, res.TypeLine, res.BorderColor, res.Frame,
			res.FullArt, res.Textless, res.PromoType, res.CMC, res.ColorIdentity, res.ScryfallID, res.Legalities,
			createdAt,
		)

		if err != nil {
			return nil, fmt.Errorf("insert failed for '%s': %w", res.Name, err)
		}
		productIDs = append(productIDs, pID)

		// Distribute stock across 1-2 storage locations
		loc1 := storageLocs[i%len(storageLocs)]
		qty1 := randInt(1, stock)
		if sid, ok := storage[loc1]; ok {
			db.Exec(`
				INSERT INTO product_storage (product_id, storage_id, quantity)
				VALUES ($1, $2, $3)
				ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity
			`, pID, sid, qty1)
		}
		// Some products split across 2 locations
		if i%3 == 0 && stock > qty1 {
			loc2 := storageLocs[(i+2)%len(storageLocs)]
			if loc2 != loc1 {
				if sid, ok := storage[loc2]; ok {
					db.Exec(`
						INSERT INTO product_storage (product_id, storage_id, quantity)
						VALUES ($1, $2, $3)
						ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity
					`, pID, sid, stock-qty1)
				}
			}
		}

		// Assign 1 primary category, some get 2
		primaryCat := catKeys[i%len(catKeys)]
		if catID, ok := cats[primaryCat]; ok {
			db.Exec(`
				INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, pID, catID)
		}
		// Every 4th card also gets "featured"
		if i%4 == 0 {
			if catID, ok := cats["featured"]; ok {
				db.Exec(`
					INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)
					ON CONFLICT DO NOTHING
				`, pID, catID)
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
