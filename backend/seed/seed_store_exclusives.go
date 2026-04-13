package main

import (
	"fmt"
	"time"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// seedStoreExclusives seeds custom Commander decks, tokens, proxy kits and bundles.
func seedStoreExclusives(db *sqlx.DB, cats CategoryMap, storage StorageMap) ([]string, error) {
	logger.Info("⭐ Seeding Store Exclusives...")

	// Helper: insert deck card via Scryfall lookup
	seedDeckCards := func(deckID string, identifiers []external.CardIdentifier) {
		inserted := 0
		for _, identifier := range identifiers {
			res, err := external.LookupMTGCard(
				identifier.ScryfallID, identifier.Name, identifier.Set,
				identifier.CollectorNumber, "non_foil",
			)
			if err != nil {
				logger.Error("  ❌ Card NOT found: %s (%s #%s): %v",
					identifier.Name, identifier.Set, identifier.CollectorNumber, err)
				continue
			}
			r := *res
			qty := 1
			_ = qty

			_, err = db.Exec(`
				INSERT INTO deck_card (
					product_id, name, set_name, set_code, collector_number, quantity,
					language, color_identity, cmc, is_legendary, is_historic, is_land, is_basic_land,
					art_variation, oracle_text, artist, type_line, border_color, frame,
					full_art, textless, promo_type, image_url, foil_treatment, card_treatment, rarity,
					scryfall_id, legalities
				) VALUES ($1,$2,$3,$4,$5,1,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,'non_foil','normal',$23,$24,$25)
			`,
				deckID, r.Name, r.SetName, r.SetCode, r.CollectorNumber,
				r.Language, r.ColorIdentity, r.CMC, r.IsLegendary, r.IsHistoric, r.IsLand, r.IsBasicLand,
				r.ArtVariation, r.OracleText, r.Artist, r.TypeLine, r.BorderColor, r.Frame,
				r.FullArt, r.Textless, r.PromoType, r.ImageURL, r.Rarity, r.ScryfallID, r.Legalities)

			if err != nil {
				logger.Error("  ❌ DB error for '%s': %v", r.Name, err)
			} else {
				inserted++
			}
			time.Sleep(80 * time.Millisecond)
		}
		logger.Info("  ✅ Seeded %d/%d deck cards for %s", inserted, len(identifiers), deckID)
	}

	var allIDs []string

	// ── 1. Goblin Swarm Commander ──────────────────────────────────────────
	logger.Info("  Building: Custom Commander — Goblin Swarm...")
	var goblinID string
	if err := db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description, cost_basis_cop, created_at)
		VALUES ($1,'mtg','store_exclusives','manual',$2,$3,$4,$5,$6,$7) RETURNING id
	`,
		"Custom Commander Precon: Goblin Swarm",
		165000, 5,
		"https://cards.scryfall.io/art_crop/front/0/e/0e386888-57f5-4eb6-88e8-5679bb0173ae.jpg",
		"A highly synergistic 100-card goblin deck ready to play out of the box. Includes Krenko as commander, tokens, and a foil-wrapped deck box.",
		90000,
		daysAgo(20),
	).Scan(&goblinID); err != nil {
		return nil, fmt.Errorf("failed to seed goblin deck: %w", err)
	}
	allIDs = append(allIDs, goblinID)
	db.Exec(`INSERT INTO product_storage (product_id,storage_id,quantity) VALUES ($1,$2,5) ON CONFLICT DO NOTHING`, goblinID, storage["Showcase A"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, goblinID, cats["featured"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, goblinID, cats["staff-picks"])

	seedDeckCards(goblinID, []external.CardIdentifier{
		{Name: "Krenko, Mob Boss", Set: "rvr", CollectorNumber: "114"},
		{Name: "Goblin Chieftain", Set: "jmp", CollectorNumber: "324"},
		{Name: "Goblin Warchief", Set: "dom", CollectorNumber: "130"},
		{Name: "Goblin Guide", Set: "zen", CollectorNumber: "136"},
		{Name: "Siege-Gang Commander", Set: "dom", CollectorNumber: "144"},
		{Name: "Skirk Prospector", Set: "ons", CollectorNumber: "222"},
		{Name: "Goblin Matron", Set: "7ed", CollectorNumber: "190"},
		{Name: "Goblin Recruiter", Set: "vis", CollectorNumber: "93"},
		{Name: "Reckless Bushwhacker", Set: "ogw", CollectorNumber: "115"},
		{Name: "Lightning Bolt", Set: "m11", CollectorNumber: "149"},
		{Name: "Pyroclasm", Set: "m11", CollectorNumber: "155"},
		{Name: "Goblin Bombardment", Set: "tsx", CollectorNumber: "99"},
		{Name: "Phyrexian Altar", Set: "inv", CollectorNumber: "308"},
		{Name: "Sol Ring", Set: "clb", CollectorNumber: "882"},
		{Name: "Arcane Signet", Set: "clb", CollectorNumber: "861"},
		{Name: "Swiftfoot Boots", Set: "cmr", CollectorNumber: "474"},
		{Name: "Lightning Greaves", Set: "cmr", CollectorNumber: "470"},
		{Name: "Mountain", Set: "usg", CollectorNumber: "343"},
		{Name: "Mountain", Set: "usg", CollectorNumber: "344"},
		{Name: "Mountain", Set: "usg", CollectorNumber: "345"},
		{Name: "Ancient Tomb", Set: "tmp", CollectorNumber: "315"},
		{Name: "Cavern of Souls", Set: "avr", CollectorNumber: "226"},
	})

	// ── 2. Dragon Hoard Commander (Premium) ───────────────────────────────
	logger.Info("  Building: Premium Commander — Dragon Hoard...")
	var dragonID string
	if err := db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description, cost_basis_cop, created_at)
		VALUES ($1,'mtg','store_exclusives','manual',$2,$3,$4,$5,$6,$7) RETURNING id
	`,
		"Premium Commander: Dragon Hoard",
		480000, 3,
		"https://cards.scryfall.io/art_crop/front/4/8/48002002-0002-48a4-a3ad-0002b8004f1a.jpg",
		"A high-power 100-card Commander deck centered around The Ur-Dragon. Includes dual lands, fetch lands, all five Dragonlords, and a Dragon's Hoard.",
		280000,
		daysAgo(10),
	).Scan(&dragonID); err != nil {
		return nil, fmt.Errorf("failed to seed dragon deck: %w", err)
	}
	allIDs = append(allIDs, dragonID)
	db.Exec(`INSERT INTO product_storage (product_id,storage_id,quantity) VALUES ($1,$2,3) ON CONFLICT DO NOTHING`, dragonID, storage["Showcase A"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, dragonID, cats["featured"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, dragonID, cats["hot-items"])

	seedDeckCards(dragonID, []external.CardIdentifier{
		// Commander
		{Name: "The Ur-Dragon", Set: "c17", CollectorNumber: "48"},
		// Dragons
		{Name: "Miirym, Sentinel Wyrm", Set: "clb", CollectorNumber: "284"},
		{Name: "Lathliss, Dragon Queen", Set: "m19", CollectorNumber: "149"},
		{Name: "Terror of the Peaks", Set: "m21", CollectorNumber: "164"},
		{Name: "Balefire Dragon", Set: "isd", CollectorNumber: "129"},
		{Name: "Dragonlord Dromoka", Set: "dtk", CollectorNumber: "217"},
		{Name: "Dragonlord Silumgar", Set: "dtk", CollectorNumber: "220"},
		{Name: "Dragonlord Kolaghan", Set: "dtk", CollectorNumber: "218"},
		{Name: "Dragonlord Atarka", Set: "dtk", CollectorNumber: "216"},
		{Name: "Dragonlord Ojutai", Set: "dtk", CollectorNumber: "219"},
		{Name: "Tiamat", Set: "afr", CollectorNumber: "235"},
		// Support
		{Name: "Sol Ring", Set: "clb", CollectorNumber: "882"},
		{Name: "Arcane Signet", Set: "clb", CollectorNumber: "861"},
		{Name: "Dragon's Hoard", Set: "m19", CollectorNumber: "232"},
		{Name: "Chromatic Lantern", Set: "grn", CollectorNumber: "233"},
		{Name: "The Great Henge", Set: "eld", CollectorNumber: "161"},
		{Name: "Rhystic Study", Set: "pcy", CollectorNumber: "45"},
		{Name: "Smothering Tithe", Set: "rna", CollectorNumber: "22"},
		{Name: "Cyclonic Rift", Set: "rtr", CollectorNumber: "35"},
		{Name: "Swords to Plowshares", Set: "clb", CollectorNumber: "707"},
		// Lands
		{Name: "Command Tower", Set: "cm2", CollectorNumber: "242"},
		{Name: "Path of Ancestry", Set: "c17", CollectorNumber: "63"},
		{Name: "Stomping Ground", Set: "rna", CollectorNumber: "259"},
		{Name: "Steam Vents", Set: "grn", CollectorNumber: "257"},
		{Name: "Breeding Pool", Set: "rna", CollectorNumber: "246"},
		{Name: "Blood Crypt", Set: "rna", CollectorNumber: "245"},
		{Name: "Wooded Foothills", Set: "ons", CollectorNumber: "330"},
		{Name: "Polluted Delta", Set: "ons", CollectorNumber: "335"},
		{Name: "Scalding Tarn", Set: "zen", CollectorNumber: "223"},
		{Name: "Ancient Tomb", Set: "tmp", CollectorNumber: "315"},
	})

	// ── 3. Elf Tribal Commander ────────────────────────────────────────────
	logger.Info("  Building: Custom Commander — Elf Tribal...")
	var elfID string
	if err := db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description, cost_basis_cop, created_at)
		VALUES ($1,'mtg','store_exclusives','manual',$2,$3,$4,$5,$6,$7) RETURNING id
	`,
		"Custom Commander: Elf Tribal (Selvala Focus)",
		220000, 4,
		"https://cards.scryfall.io/art_crop/front/5/b/5b2b89a0-4117-4ea4-8d0c-3a8c08c7dd44.jpg",
		"100-card Elf tribal deck helmed by Selvala, Heart of the Wilds. Generates massive mana and overruns opponents with elf synergies.",
		130000,
		daysAgo(8),
	).Scan(&elfID); err != nil {
		return nil, fmt.Errorf("failed to seed elf deck: %w", err)
	}
	allIDs = append(allIDs, elfID)
	db.Exec(`INSERT INTO product_storage (product_id,storage_id,quantity) VALUES ($1,$2,4) ON CONFLICT DO NOTHING`, elfID, storage["Showcase B"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, elfID, cats["featured"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, elfID, cats["commander-staples"])

	seedDeckCards(elfID, []external.CardIdentifier{
		{Name: "Selvala, Heart of the Wilds", Set: "cns", CollectorNumber: "181"},
		{Name: "Llanowar Elves", Set: "dom", CollectorNumber: "168"},
		{Name: "Fyndhorn Elves", Set: "ice", CollectorNumber: "259"},
		{Name: "Priest of Titania", Set: "usg", CollectorNumber: "263"},
		{Name: "Elvish Archdruid", Set: "m10", CollectorNumber: "168"},
		{Name: "Wirewood Symbiote", Set: "lgn", CollectorNumber: "141"},
		{Name: "Glimpse of Nature", Set: "chk", CollectorNumber: "212"},
		{Name: "Green Sun's Zenith", Set: "mbs", CollectorNumber: "117"},
		{Name: "Craterhoof Behemoth", Set: "avr", CollectorNumber: "175"},
		{Name: "Ezuri, Renegade Leader", Set: "som", CollectorNumber: "118"},
		{Name: "Heritage Druid", Set: "mor", CollectorNumber: "120"},
		{Name: "Nettle Sentinel", Set: "mor", CollectorNumber: "131"},
		{Name: "Chord of Calling", Set: "m15", CollectorNumber: "173"},
		{Name: "Sylvan Library", Set: "ema", CollectorNumber: "187"},
		{Name: "Gaea's Cradle", Set: "usg", CollectorNumber: "321"},
		{Name: "Nykthos, Shrine to Nyx", Set: "ths", CollectorNumber: "223"},
		{Name: "Sol Ring", Set: "clb", CollectorNumber: "882"},
		{Name: "Forest", Set: "clb", CollectorNumber: "457"},
		{Name: "Forest", Set: "clb", CollectorNumber: "458"},
	})

	// ── 4. Wooden Token Set (Non-Scryfall) ────────────────────────────────
	logger.Info("  Building: Store Exclusive — Wooden Tokens...")
	var tokensID string
	if err := db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description, cost_basis_cop, created_at)
		VALUES ($1,'mtg','store_exclusives','manual',$2,$3,$4,$5,$6,$7) RETURNING id
	`,
		"Hand-Crafted Wooden Tokens (Set of 10)",
		48000, 20,
		"https://images.unsplash.com/photo-1598214886806-c87b84b7078b?q=80&w=800&auto=format&fit=crop",
		"Beautifully laser-engraved wooden tokens for MTG. Includes 2x Goblin, 2x Zombie, 2x Treasure, 2x +1/+1 Counter, 1x Soldier, 1x Spirit token.",
		22000,
		daysAgo(30),
	).Scan(&tokensID); err != nil {
		return nil, fmt.Errorf("failed to seed wooden tokens: %w", err)
	}
	allIDs = append(allIDs, tokensID)
	db.Exec(`INSERT INTO product_storage (product_id,storage_id,quantity) VALUES ($1,$2,20) ON CONFLICT DO NOTHING`, tokensID, storage["Showcase A"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, tokensID, cats["featured"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, tokensID, cats["new-arrivals"])

	// ── 5. Proxy Art Kit ───────────────────────────────────────────────────
	logger.Info("  Building: Store Exclusive — Proxy Art Kit...")
	var proxyID string
	if err := db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description, cost_basis_cop, created_at)
		VALUES ($1,'mtg','store_exclusives','manual',$2,$3,$4,$5,$6,$7) RETURNING id
	`,
		"Premium Proxy Art Kit - Power 9 Replica (Set of 9)",
		95000, 10,
		"https://images.unsplash.com/photo-1541560052753-107f96307409?q=80&w=800&auto=format&fit=crop",
		"High-quality proxies for casual and display use of the Power 9. High-resolution custom card art printed on premium cardstock. NOT for tournament use.",
		40000,
		daysAgo(15),
	).Scan(&proxyID); err != nil {
		return nil, fmt.Errorf("failed to seed proxy art kit: %w", err)
	}
	allIDs = append(allIDs, proxyID)
	db.Exec(`INSERT INTO product_storage (product_id,storage_id,quantity) VALUES ($1,$2,10) ON CONFLICT DO NOTHING`, proxyID, storage["Showcase B"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, proxyID, cats["hot-items"])

	// ── 6. Starter Bundle ─────────────────────────────────────────────────
	logger.Info("  Building: Store Exclusive — Starter Bundle...")
	var bundleID string
	if err := db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description, cost_basis_cop, created_at)
		VALUES ($1,'mtg','store_exclusives','manual',$2,$3,$4,$5,$6,$7) RETURNING id
	`,
		"MTG Starter Bundle: Deck Box + Sleeves + Dice",
		85000, 15,
		"https://images.unsplash.com/photo-1613771404721-1f92d799e49f?q=80&w=800&auto=format&fit=crop",
		"Everything you need to start playing Magic: 1 deck box (100+), 100 sleeves, 1 six-sided die, and 1 20-sided life counter die. Handpicked by the El Bulk team.",
		48000,
		daysAgo(18),
	).Scan(&bundleID); err != nil {
		return nil, fmt.Errorf("failed to seed starter bundle: %w", err)
	}
	allIDs = append(allIDs, bundleID)
	db.Exec(`INSERT INTO product_storage (product_id,storage_id,quantity) VALUES ($1,$2,15) ON CONFLICT DO NOTHING`, bundleID, storage["Counter Display"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, bundleID, cats["budget-builds"])
	db.Exec(`INSERT INTO product_category (product_id,category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, bundleID, cats["new-arrivals"])

	logger.Info("✅ %d store exclusives seeded", len(allIDs))
	return allIDs, nil
}
