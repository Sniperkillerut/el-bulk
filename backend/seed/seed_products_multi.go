package main

import (
	"fmt"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// seedMultiTCGProducts seeds Pokémon, Yu-Gi-Oh!, Lorcana and One Piece products.
func seedMultiTCGProducts(db *sqlx.DB, cats CategoryMap, storage StorageMap) ([]string, error) {
	logger.Info("🌍 Seeding multi-TCG products (Pokémon, Yu-Gi-Oh!, Lorcana, One Piece)...")

	type Item struct {
		Name      string
		TCG       string
		Category  string   // 'singles' or 'sealed'
		SetName   string
		SetCode   string
		Condition string
		Price     float64
		Stock     int
		ImageURL  string
		StorageLoc string
		CatSlug   string
	}

	items := []Item{
		// ── Pokémon Singles ──────────────────────────────────────────────────
		{
			"Charizard ex (Full Art)", "pokemon", "singles", "Scarlet & Violet 151",
			"sv3pt5", "NM", 720000, 3,
			"https://images.pokemontcg.io/sv3pt5/199_hires.png", "Showcase A", "featured",
		},
		{
			"Pikachu with Grey Felt Hat (World Champions Pack)", "pokemon", "singles", "Promo",
			"svnp", "NM", 950000, 1,
			"https://images.pokemontcg.io/svnp/85_hires.png", "Binder Vault", "featured",
		},
		{
			"Mewtwo VMAX (Secret Rare)", "pokemon", "singles", "Celebrations",
			"cel25", "LP", 280000, 2,
			"https://images.pokemontcg.io/cel25/114_hires.png", "Binder Vault", "hot-items",
		},
		{
			"Lugia VSTAR (Secret Rare)", "pokemon", "singles", "Silver Tempest",
			"swsh12", "NM", 420000, 2,
			"https://images.pokemontcg.io/swsh12/211_hires.png", "Showcase B", "hot-items",
		},
		{
			"Iono (Full Art SAR)", "pokemon", "singles", "Paradox Rift",
			"sv4", "NM", 320000, 4,
			"https://images.pokemontcg.io/sv4/269_hires.png", "Binder Vault", "new-arrivals",
		},
		{
			"Umbreon VMAX Alt Art", "pokemon", "singles", "Evolving Skies",
			"swsh7", "LP", 580000, 1,
			"https://images.pokemontcg.io/swsh7/215_hires.png", "Binder Vault", "featured",
		},
		{
			"Rayquaza VMAX Alt Art", "pokemon", "singles", "Evolving Skies",
			"swsh7", "NM", 850000, 1,
			"https://images.pokemontcg.io/swsh7/218_hires.png", "Binder Vault", "featured",
		},
		{
			"Pikachu ex (Stellar Crown)", "pokemon", "singles", "Stellar Crown",
			"sv7", "NM", 95000, 8,
			"https://images.pokemontcg.io/sv7/62_hires.png", "Showcase B", "new-arrivals",
		},

		// ── Pokémon Sealed ───────────────────────────────────────────────────
		{
			"Scarlet & Violet 151 Elite Trainer Box", "pokemon", "sealed", "Scarlet & Violet 151",
			"sv3pt5", "", 295000, 5,
			"https://images.pokemontcg.io/sv3pt5/logo.png", "Storage Box 1", "new-arrivals",
		},
		{
			"Paradox Rift Booster Box", "pokemon", "sealed", "Paradox Rift",
			"sv4", "", 680000, 3,
			"https://images.pokemontcg.io/sv4/logo.png", "Storage Box 2", "hot-items",
		},
		{
			"Stellar Crown Booster Box", "pokemon", "sealed", "Stellar Crown",
			"sv7", "", 650000, 4,
			"https://images.pokemontcg.io/sv7/logo.png", "Storage Box 1", "new-arrivals",
		},
		{
			"Temporal Forces Elite Trainer Box", "pokemon", "sealed", "Temporal Forces",
			"sv5", "", 285000, 6,
			"https://images.pokemontcg.io/sv5/logo.png", "Storage Box 2", "featured",
		},

		// ── Yu-Gi-Oh! Singles ───────────────────────────────────────────────
		{
			"Blue-Eyes White Dragon (25th Anniversary Secret Rare)", "yugioh", "singles", "LOB",
			"LOB-EN001", "NM", 185000, 3,
			"https://images.ygoprodeck.com/images/cards/89631139.jpg", "Showcase A", "featured",
		},
		{
			"Dark Magician (Ghost Rare — 25th Anniversary)", "yugioh", "singles", "GFP2",
			"GFP2-EN178", "LP", 480000, 2,
			"https://images.ygoprodeck.com/images/cards/46986414.jpg", "Binder Vault", "hot-items",
		},
		{
			"Ash Blossom & Joyous Spring (Secret Rare)", "yugioh", "singles", "MACR",
			"MACR-EN036", "NM", 65000, 8,
			"https://images.ygoprodeck.com/images/cards/14558127.jpg", "Showcase B", "tournament-ready",
		},
		{
			"Nibiru, the Primal Being (Secret Rare)", "yugioh", "singles", "IGAS",
			"IGAS-EN022", "NM", 45000, 10,
			"https://images.ygoprodeck.com/images/cards/27204311.jpg", "Storage Box 1", "tournament-ready",
		},
		{
			"Mathmech Circular (Collector Rare)", "yugioh", "singles", "DIFO",
			"DIFO-EN022", "NM", 35000, 6,
			"https://images.ygoprodeck.com/images/cards/36521307.jpg", "Storage Box 2", "budget-builds",
		},
		{
			"Exodia the Forbidden One (25th Secret Rare)", "yugioh", "singles", "MZMI",
			"MZMI-EN007", "NM", 125000, 2,
			"https://images.ygoprodeck.com/images/cards/33396948.jpg", "Binder Vault", "featured",
		},

		// ── Yu-Gi-Oh! Sealed ─────────────────────────────────────────────────
		{
			"Maze of Millennia Collector Box", "yugioh", "sealed", "MZMI",
			"MZMI", "", 325000, 3,
			"https://images.ygoprodeck.com/images/sets/MZMI.jpg", "Storage Box 1", "featured",
		},
		{
			"Wild Survivors Booster Box", "yugioh", "sealed", "WISU",
			"WISU", "", 280000, 4,
			"https://images.ygoprodeck.com/images/sets/WISU.jpg", "Storage Box 2", "new-arrivals",
		},

		// ── Disney Lorcana Singles ────────────────────────────────────────────
		{
			"Mickey Mouse - Brave Little Tailor (Enchanted)", "lorcana", "singles", "The First Chapter",
			"TFC", "NM", 1850000, 1,
			"https://lorcana-api.com/images/thefirstchapter/card/high/171_mickeyMouse_braveLittleTailor.webp", "Binder Vault", "featured",
		},
		{
			"Elsa - Spirit of Winter (Legendary)", "lorcana", "singles", "The First Chapter",
			"TFC", "NM", 350000, 2,
			"https://lorcana-api.com/images/thefirstchapter/card/high/039_elsa_spiritOfWinter.webp", "Showcase A", "featured",
		},
		{
			"Moana - Of Motunui (Super Rare)", "lorcana", "singles", "Rise of the Floodborn",
			"ROF", "LP", 120000, 4,
			"https://lorcana-api.com/images/riseofthefloodborn/card/high/076_moana_ofMotunui.webp", "Showcase B", "hot-items",
		},
		{
			"Sisu - Divine Water Dragon (Legendary)", "lorcana", "singles", "Into the Inklands",
			"ITI", "NM", 280000, 2,
			"https://lorcana-api.com/images/intotheinklands/card/high/030_sisu_divineWaterDragon.webp", "Showcase A", "staff-picks",
		},

		// ── Lorcana Sealed ────────────────────────────────────────────────────
		{
			"Ursula's Return Booster Box", "lorcana", "sealed", "Ursula's Return",
			"UR", "", 580000, 5,
			"https://lorcana-api.com/images/ursulasreturn/booster_box.webp", "Storage Box 1", "new-arrivals",
		},
		{
			"The First Chapter Starter Deck - Amber/Amethyst", "lorcana", "sealed", "The First Chapter",
			"TFC", "", 95000, 8,
			"https://lorcana-api.com/images/thefirstchapter/starter_amber.webp", "Storage Box 2", "budget-builds",
		},

		// ── One Piece Singles ──────────────────────────────────────────────────
		{
			"Monkey D. Luffy Gear 5 (Parallel Rare)", "onepiece", "singles", "OP-01",
			"OP01-060", "NM", 280000, 3,
			"https://en.onepiece-cardgame.com/images/cardlist/card/OP01-060.png", "Showcase A", "featured",
		},
		{
			"Boa Hancock (Leader) (Secret Super Rare)", "onepiece", "singles", "OP-05",
			"OP05-004", "NM", 180000, 2,
			"https://en.onepiece-cardgame.com/images/cardlist/card/OP05-004.png", "Showcase B", "hot-items",
		},
		{
			"Roronoa Zoro (Alt Art)", "onepiece", "singles", "OP-03",
			"OP03-001", "LP", 95000, 4,
			"https://en.onepiece-cardgame.com/images/cardlist/card/OP03-001.png", "Binder Vault", "staff-picks",
		},
		{
			"Nami (Parallel Rare)", "onepiece", "singles", "OP-02",
			"OP02-016", "NM", 65000, 6,
			"https://en.onepiece-cardgame.com/images/cardlist/card/OP02-016.png", "Storage Box 1", "budget-builds",
		},

		// ── One Piece Sealed ─────────────────────────────────────────────────
		{
			"OP-07 500 Years in the Future Booster Box", "onepiece", "sealed", "OP-07",
			"OP07", "", 420000, 4,
			"https://en.onepiece-cardgame.com/images/cardlist/booster/OP07.png", "Storage Box 2", "new-arrivals",
		},
		{
			"OP-09 The Four Emperors Booster Box", "onepiece", "sealed", "OP-09",
			"OP09", "", 480000, 3,
			"https://en.onepiece-cardgame.com/images/cardlist/booster/OP09.png", "Storage Box 1", "hot-items",
		},
	}

	var ids []string
	for i, item := range items {
		createdAt := daysAgo(randInt(2, 60))
		costBasis := item.Price * 0.60

		var cond *string
		if item.Condition != "" {
			cond = &item.Condition
		}

		var pID string
		err := db.QueryRow(`
			INSERT INTO product (
				name, tcg, category, set_name, set_code,
				condition, price_source, price_cop_override,
				stock, image_url, cost_basis_cop, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, 'manual', $7, $8, $9, $10, $11)
			RETURNING id
		`,
			item.Name, item.TCG, item.Category, item.SetName, item.SetCode,
			cond, item.Price, item.Stock, item.ImageURL, costBasis, createdAt,
		).Scan(&pID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert '%s' (%s): %w", item.Name, item.TCG, err)
		}
		ids = append(ids, pID)

		// Storage
		if sid, ok := storage[item.StorageLoc]; ok {
			db.Exec(`
				INSERT INTO product_storage (product_id, storage_id, quantity)
				VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
			`, pID, sid, item.Stock)
		}

		// Category
		if catID, ok := cats[item.CatSlug]; ok {
			db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
		}

		// Every 5th also gets "sale" category for variety
		if i%5 == 0 {
			if catID, ok := cats["sale"]; ok {
				db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
			}
		}
	}

	logger.Info("✅ %d multi-TCG products seeded", len(ids))
	return ids, nil
}
