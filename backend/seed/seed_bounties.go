package main

import (
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedBounties(db *sqlx.DB) []string {
	logger.Info("🎯 Seeding bounties (all fields exercised)...")

	type Bounty struct {
		Name            string
		TCG             string
		SetName         string
		Condition       string
		FoilTreatment   string
		CardTreatment   string
		CollectorNumber string
		PromoType       string
		Language        string
		TargetPrice     float64
		HidePrice       bool
		QuantityNeeded  int
		ImageURL        string
		PriceSource     string
		PriceReference  *float64
		IsActive        bool
	}

	ref := func(f float64) *float64 { return &f }

	bounties := []Bounty{
		// ── MTG — Modern staples (manual price) ─────────────────────────────
		{
			Name: "Ragavan, Nimble Pilferer", TCG: "mtg",
			SetName: "Modern Horizons 2", Condition: "NM",
			FoilTreatment: "non_foil", CardTreatment: "normal",
			CollectorNumber: "138", Language: "en",
			TargetPrice: 250000, HidePrice: false, QuantityNeeded: 3,
			ImageURL:    "https://cards.scryfall.io/normal/front/a/9/a9a54bf8-4f06-4e44-b831-a501a014e784.jpg",
			PriceSource: "manual", IsActive: true,
		},
		{
			Name: "Orcish Bowmasters", TCG: "mtg",
			SetName: "The Lord of the Rings: Tales of Middle-earth", Condition: "NM",
			FoilTreatment: "foil", CardTreatment: "borderless",
			Language: "en",
			TargetPrice: 180000, HidePrice: true, QuantityNeeded: 4,
			ImageURL:    "https://cards.scryfall.io/normal/front/b/c/bc6e7e22-5c93-4cf3-a5e4-3dccaccfa16b.jpg",
			PriceSource: "tcgplayer", PriceReference: ref(38.50), IsActive: true,
		},
		{
			Name: "The One Ring", TCG: "mtg",
			SetName: "The Lord of the Rings: Tales of Middle-earth", Condition: "LP",
			FoilTreatment: "non_foil", CardTreatment: "borderless",
			Language: "en",
			TargetPrice: 380000, HidePrice: false, QuantityNeeded: 2,
			ImageURL:    "https://cards.scryfall.io/normal/front/d/5/d5806e68-1054-47ec-9a7e-b8f7b27bb1b4.jpg",
			PriceSource: "tcgplayer", PriceReference: ref(82.00), IsActive: true,
		},
		// ── MTG — Commander staples (cardmarket price) ───────────────────────
		{
			Name: "Mana Crypt", TCG: "mtg",
			SetName: "Double Masters 2022", Condition: "NM",
			FoilTreatment: "etched_foil", CardTreatment: "normal",
			CollectorNumber: "400", Language: "en",
			TargetPrice: 0, HidePrice: true, QuantityNeeded: 2,
			ImageURL:    "https://cards.scryfall.io/normal/front/4/d/4d960186-4559-4af1-b5a6-0e72b45c5b23.jpg",
			PriceSource: "cardmarket", PriceReference: ref(145.00), IsActive: true,
		},
		{
			Name: "Sheoldred, the Apocalypse", TCG: "mtg",
			SetName: "Dominaria United", Condition: "NM",
			FoilTreatment: "surge_foil", CardTreatment: "showcase",
			Language: "en",
			TargetPrice: 420000, HidePrice: false, QuantityNeeded: 4,
			ImageURL:    "https://cards.scryfall.io/normal/front/a/7/a7534174-85ef-487e-8b25-f02f69c03e80.jpg",
			PriceSource: "tcgplayer", PriceReference: ref(95.00), IsActive: true,
		},
		{
			Name: "Jace, the Mind Sculptor", TCG: "mtg",
			SetName: "Worldwake", Condition: "MP",
			FoilTreatment: "non_foil", CardTreatment: "normal",
			Language: "en",
			TargetPrice: 180000, HidePrice: false, QuantityNeeded: 2,
			ImageURL:    "https://cards.scryfall.io/normal/front/1/a/1a4dfe0c-7080-46b1-948a-8c3ca9c2de08.jpg",
			PriceSource: "manual", IsActive: true,
		},
		// ── MTG — Promo/Special ─────────────────────────────────────────────
		{
			Name: "Lightning Bolt (30th Anniversary Promo)", TCG: "mtg",
			SetName: "30th Anniversary Promos", Condition: "NM",
			FoilTreatment: "foil", CardTreatment: "promo",
			PromoType: "release", Language: "en",
			TargetPrice: 55000, HidePrice: false, QuantityNeeded: 6,
			ImageURL:    "https://cards.scryfall.io/normal/front/f/2/f22fbf43-84fb-4985-af6c-f15de4fcf4f5.jpg",
			PriceSource: "manual", IsActive: true,
		},
		{
			Name: "Liliana of the Veil (Judge Promo)", TCG: "mtg",
			SetName: "Judge Gift Cards 2019", Condition: "NM",
			FoilTreatment: "foil", CardTreatment: "judge_promo",
			PromoType: "judge", Language: "en",
			TargetPrice: 0, HidePrice: true, QuantityNeeded: 1,
			ImageURL:    "https://cards.scryfall.io/normal/front/a/1/a1aa8fb5-a3c4-42c4-84d0-64f9a2c23e00.jpg",
			PriceSource: "cardmarket", PriceReference: ref(210.00), IsActive: true,
		},
		// ── Pokémon ─────────────────────────────────────────────────────────
		{
			Name: "Charizard ex SAR #199", TCG: "pokemon",
			SetName: "Scarlet & Violet 151", Condition: "NM",
			FoilTreatment: "holo_foil", CardTreatment: "alternate_art",
			Language: "en",
			TargetPrice: 550000, HidePrice: false, QuantityNeeded: 2,
			ImageURL:    "https://images.pokemontcg.io/sv3pt5/199_hires.png",
			PriceSource: "manual", IsActive: true,
		},
		{
			Name: "Rayquaza VMAX Alt Art #218", TCG: "pokemon",
			SetName: "Evolving Skies", Condition: "NM",
			FoilTreatment: "holo_foil", CardTreatment: "alternate_art",
			Language: "en",
			TargetPrice: 750000, HidePrice: false, QuantityNeeded: 1,
			ImageURL:    "https://images.pokemontcg.io/swsh7/218_hires.png",
			PriceSource: "manual", IsActive: true,
		},
		// ── Yu-Gi-Oh! ────────────────────────────────────────────────────────
		{
			Name: "Blue-Eyes White Dragon (LOB 1st Edition)", TCG: "yugioh",
			SetName: "Legend of Blue Eyes White Dragon", Condition: "LP",
			FoilTreatment: "non_foil", CardTreatment: "normal",
			Language: "en",
			TargetPrice: 0, HidePrice: true, QuantityNeeded: 1,
			ImageURL:    "https://images.ygoprodeck.com/images/cards/89631139.jpg",
			PriceSource: "manual", IsActive: true,
		},
		{
			Name: "Ash Blossom & Joyous Spring (Ghost Rare)", TCG: "yugioh",
			SetName: "2023 Tin of the Pharaoh's Gods", Condition: "NM",
			FoilTreatment: "platinum_foil", CardTreatment: "normal",
			Language: "en",
			TargetPrice: 80000, HidePrice: false, QuantityNeeded: 8,
			ImageURL:    "https://images.ygoprodeck.com/images/cards/14558127.jpg",
			PriceSource: "manual", IsActive: true,
		},
		// ── Lorcana ─────────────────────────────────────────────────────────
		{
			Name: "Mickey Mouse - Brave Little Tailor (Enchanted)", TCG: "lorcana",
			SetName: "The First Chapter", Condition: "NM",
			FoilTreatment: "foil", CardTreatment: "alternate_art",
			Language: "en",
			TargetPrice: 0, HidePrice: true, QuantityNeeded: 1,
			PriceSource: "manual", IsActive: true,
		},
		// ── Inactive bounty (for testing toggle) ─────────────────────────────
		{
			Name: "Black Lotus (Unlimited)", TCG: "mtg",
			SetName: "Unlimited Edition", Condition: "HP",
			FoilTreatment: "non_foil", CardTreatment: "normal",
			Language: "en",
			TargetPrice: 0, HidePrice: true, QuantityNeeded: 1,
			ImageURL:    "https://cards.scryfall.io/normal/front/1/9/19911e6e-7c35-4281-b31c-266382f052cc.jpg",
			PriceSource: "manual", IsActive: false, // INACTIVE – testing
		},
		// ── Japanese language testing ─────────────────────────────────────────
		{
			Name: "Force of Will (Japanese)", TCG: "mtg",
			SetName: "Alliances", Condition: "NM",
			FoilTreatment: "non_foil", CardTreatment: "normal",
			Language: "ja",
			TargetPrice: 145000, HidePrice: false, QuantityNeeded: 4,
			PriceSource: "cardmarket", PriceReference: ref(28.00), IsActive: true,
		},
	}

	var ids []string
	for _, b := range bounties {
		var id string
		err := db.QueryRow(`
			INSERT INTO bounty (
				name, tcg, set_name, condition,
				foil_treatment, card_treatment, collector_number, promo_type,
				language, target_price, hide_price, quantity_needed,
				image_url, price_source, price_reference, is_active
			) VALUES (
				$1,$2,$3,$4,
				$5,$6,$7,$8,
				$9,$10,$11,$12,
				$13,$14,$15,$16
			) RETURNING id
		`,
			b.Name, b.TCG, b.SetName, nilIfEmpty(b.Condition),
			b.FoilTreatment, b.CardTreatment, nilIfEmpty(b.CollectorNumber), nilIfEmpty(b.PromoType),
			b.Language, b.TargetPrice, b.HidePrice, b.QuantityNeeded,
			nilIfEmpty(b.ImageURL), b.PriceSource, b.PriceReference, b.IsActive,
		).Scan(&id)
		if err != nil {
			logger.Error("Failed to seed bounty '%s': %v", b.Name, err)
			continue
		}
		ids = append(ids, id)
	}

	logger.Info("✅ %d bounties seeded", len(ids))
	return ids
}

// nilIfEmpty converts an empty string to nil for nullable TEXT columns.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
