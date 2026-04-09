package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	mode := flag.String("mode", "minimal", "Seeding mode: 'minimal' (configs + 1 product) or 'full' (hundreds of records)")
	flag.Parse()

	database := db.Connect()
	defer database.Close()

	if *mode == "full" {
		logger.Info("🌟 Running FULL seeding mode (this may take a minute)...")
	} else {
		logger.Info("🌱 Running MINIMAL seeding mode...")
	}

	clearTables(database)

	adminID := seedAdmin(database)
	seedTCGs(database)
	categoryMap := seedCategories(database)
	storageIDs := seedStorage(database)
	seedSettings(database)
	seedSets(database)
	seedThemes(database)
	seedTranslations(database)

	if *mode == "minimal" {
		seedMinimalData(database, categoryMap, storageIDs)
	} else {
		productIDs := seedFullData(database, categoryMap, storageIDs)
		seedNotices(database, productIDs)
		seedCRM(database, adminID)
	}

	logger.Info("✅ Seeding complete! (Mode: %s, Admin: %s)", *mode, adminID)
}

func clearTables(db *sqlx.DB) {
	logger.Info("Clearing tables...")
	tables := []string{
		"order_item",
		"\"order\"",
		"customer",
		"product_category",
		"product_storage",
		"product",
		"bounty_offer",
		"bounty",
		"client_request",
		"storage_location",
		"custom_category",
		"notice",
		"newsletter_subscriber",
		"customer_auth",
		"customer_note",
		"deck_card",
		"tcg",
		"admin",
	}
	for _, t := range tables {
		db.Exec(fmt.Sprintf("DELETE FROM %s", t))
	}
	db.Exec("DELETE FROM tcg_set")
}

func seedAdmin(db *sqlx.DB) string {
	user := os.Getenv("ADMIN_USERNAME")
	pass := os.Getenv("ADMIN_PASSWORD")
	if user == "" {
		user = "admin"
	}
	if pass == "" {
		pass = "elbulk2024!"
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	var id string
	db.QueryRow(`
		INSERT INTO admin (username, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`, user, string(hash)).Scan(&id)
	logger.Info("Admin user '%s' created", user)
	return id
}

func seedTCGs(db *sqlx.DB) map[string]string {
	tcgs := []struct{ ID, Name string }{
		{"mtg", "Magic: The Gathering"},
		{"pokemon", "Pokémon"},
		{"yugioh", "Yu-Gi-Oh!"},
		{"lorcana", "Disney Lorcana"},
		{"onepiece", "One Piece"},
	}
	ids := make(map[string]string)
	for _, t := range tcgs {
		db.Exec(`INSERT INTO tcg (id, name) VALUES ($1, $2)`, t.ID, t.Name)
		ids[t.ID] = t.ID
	}
	return ids
}

func seedCategories(db *sqlx.DB) map[string]string {
	cats := []struct {
		Name, Slug, BgColor, TextColor, Icon string
	}{
		{"Featured", "featured", "#ffd700", "#000000", "Star"},
		{"Hot Items", "hot-items", "#ff4500", "#ffffff", "Flame"},
		{"New Arrivals", "new-arrivals", "#1e90ff", "#ffffff", "Zap"},
		{"Sale", "sale", "#32cd32", "#000000", "Tag"},
	}
	mapping := make(map[string]string)
	for _, cat := range cats {
		var id string
		db.QueryRow(`
			INSERT INTO custom_category (name, slug, bg_color, text_color, icon) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id
		`, cat.Name, cat.Slug, cat.BgColor, cat.TextColor, cat.Icon).Scan(&id)
		mapping[cat.Slug] = id
	}
	return mapping
}

func seedStorage(db *sqlx.DB) []string {
	locations := []string{"Showcase A", "Storage Box 1", "Binder Vault", "pending"}
	var ids []string
	for _, loc := range locations {
		var id string
		db.QueryRow(`INSERT INTO storage_location (name) VALUES ($1) RETURNING id`, loc).Scan(&id)
		ids = append(ids, id)
	}
	return ids
}

func seedSettings(db *sqlx.DB) {
	settings := map[string]string{
		"usd_to_cop_rate": "4450",
		"eur_to_cop_rate": "4800",
		"contact_email":   "contact@el-bulk.com",
		"contact_phone":   "+57 300 123 4567",
	}
	for k, v := range settings {
		db.Exec(`INSERT INTO setting (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, k, v)
	}
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
	tx.Exec("INSERT INTO setting (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "last_set_sync", time.Now().Format(time.RFC3339))
	tx.Commit()
	logger.Info("✅ %d sets synchronized", len(sets))
}

func seedMinimalData(db *sqlx.DB, cats map[string]string, storageIDs []string) string {
	// 1 Sample Product
	name := "Black Lotus"
	var pID string
	db.QueryRow(`
		INSERT INTO product (
			name, tcg, category, set_name, set_code, price_source, price_cop_override, 
			stock, image_url, color_identity, oracle_text, legalities
		)
		VALUES ($1, 'mtg', 'singles', 'Alpha', 'LEA', 'manual', 25000000, 1, 
			'https://cards.scryfall.io/normal/front/1/9/19911e6e-7c35-4281-b31c-266382f052cc.jpg?1717190810', 'C',
			'{T}, Sacrifice Black Lotus: Add three mana of any one color.',
			'{"commander": "banned", "legacy": "banned", "vintage": "restricted"}'::jsonb
		) RETURNING id
	`, name).Scan(&pID)

	db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 1)`, pID, storageIDs[0])
	db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats["featured"])
	return pID
}

func seedFullData(db *sqlx.DB, cats map[string]string, storageIDs []string) []string {
	// Seed 1 Minimal Product first as a safety baseline
	seedMinimalData(db, cats, storageIDs)

	// 1. Seed Hundreds of MTG Products (Bulk)
	identifiers := []external.CardIdentifier{
		{Name: "Sheoldred, the Apocalypse", Set: "dmu"},
		{Name: "The One Ring", Set: "ltr"},
		{Name: "Mana Crypt", Set: "2xm"},
		{Name: "Ragavan, Nimble Pilferer", Set: "mh2"},
		{Name: "Orcish Bowmasters", Set: "ltr"},
		{Name: "Sol Ring", Set: "v10"},
		{Name: "Lightning Bolt", Set: "lea"},
		{Name: "Force of Will", Set: "all"},
		{Name: "Brainstorm", Set: "mmq"},
		{Name: "Tarmogoyf", Set: "fut"},
		{Name: "Snapcaster Mage", Set: "isd"},
		{Name: "Liliana of the Veil", Set: "isd"},
		{Name: "Jace, the Mind Sculptor", Set: "wwk"},
		{Name: "Thoughtseize", Set: "lrw"},
		{Name: "Mox Amber", Set: "dom"},
		{Name: "Chrome Mox", Set: "mrd"},
		{Name: "Mox Opal", Set: "som"},
		{Name: "Cavern of Souls", Set: "avr"},
		{Name: "Ancient Tomb", Set: "tmp"},
		{Name: "City of Traitors", Set: "exo"},
		{Name: "Gaea's Cradle", Set: "usg"},
		{Name: "Serra's Sanctum", Set: "usg"},
		{Name: "Tolarian Academy", Set: "usg"},
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
	}

	// Double the list to reach ~100 unique-ish entries
	baseLen := len(identifiers)
	for i := 0; i < 70; i++ {
		identifiers = append(identifiers, identifiers[i%baseLen])
	}

	logger.Info("Fetching bulk metadata for %d cards...", len(identifiers))

	var results []external.CardLookupResult
	// Reduced chunk size for better reliability and added basic retry
	chunkSize := 25
	for i := 0; i < len(identifiers); i += chunkSize {
		end := i + chunkSize
		if end > len(identifiers) {
			end = len(identifiers)
		}
		chunk := identifiers[i:end]

		var res []external.CardLookupResult
		var err error

		// 3 attempts per chunk
		for attempt := 1; attempt <= 3; attempt++ {
			res, err = external.BatchLookupMTGCard(chunk)
			if err == nil {
				break
			}
			logger.Warn("  ⚠️ Attempt %d failed for chunk at index %d: %v", attempt, i, err)
			time.Sleep(2 * time.Second)
		}

		if err != nil {
			logger.Error("  ❌ Batch lookup chunk failed after 3 attempts: %v", err)
			continue
		}
		results = append(results, res...)
		time.Sleep(200 * time.Millisecond) // Respect rate limits
	}

	fTreatments := []models.FoilTreatment{models.FoilNonFoil, models.FoilFoil, models.FoilEtchedFoil, models.FoilSurgeFoil}
	cTreatments := []models.CardTreatment{models.TreatmentNormal, models.TreatmentShowcase, models.TreatmentBorderless, models.TreatmentExtendedArt}
	conditions := []string{"NM", "LP", "MP"}

	var productIDs []string
	for i, res := range results {
		f := fTreatments[i%len(fTreatments)]
		t := cTreatments[i%len(cTreatments)]
		cond := conditions[i%len(conditions)]
		stock := rand.Intn(15) + 1
		price := float64(rand.Intn(500000) + 10000)

		var pID string
		err := db.Get(&pID, `
			INSERT INTO product (
				name, tcg, category, set_name, set_code, collector_number, condition,
				foil_treatment, card_treatment, language, price_source, price_cop_override,
				image_url, stock, rarity, is_legendary, is_historic, is_land, is_basic_land,
				art_variation, oracle_text, artist, type_line, border_color, frame,
				full_art, textless, promo_type, cmc, color_identity, scryfall_id, legalities
			) VALUES ($1, 'mtg', 'singles', $2, $3, $4, $5, $6, $7, $8, 'manual', $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
			RETURNING id
		`,
			res.Name, res.SetName, res.SetCode, res.CollectorNumber, cond, f, t, res.Language,
			price, res.ImageURL, stock, res.Rarity,
			res.IsLegendary, res.IsHistoric, res.IsLand, res.IsBasicLand,
			res.ArtVariation, res.OracleText, res.Artist, res.TypeLine,
			res.BorderColor, res.Frame, res.FullArt, res.Textless, res.PromoType,
			res.CMC, res.ColorIdentity, res.ScryfallID, res.Legalities)

		if err == nil {
			productIDs = append(productIDs, pID)
			// Categories: Every product gets one primary category, some get "Featured" too
			catKeys := []string{"new-arrivals", "hot-items", "sale"}
			db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats[catKeys[i%len(catKeys)]])
			if i%5 == 0 {
				db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats["featured"])
			}
			// Storage
			db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, $3)`, pID, storageIDs[rand.Intn(len(storageIDs))], stock)
		}
	}

	// 2. MTG Sealed
	logger.Info("Seeding MTG Sealed...")
	mtgSealed := []struct {
		Name, Set, Img string
		Price          float64
	}{
		{"Outlaws of Thunder Junction Play Booster Box", "OTJ", "https://m.media-amazon.com/images/I/71D+8+8+L+L._AC_SL1500_.jpg", 650000},
		{"Modern Horizons 3 Collector Booster Box", "MH3", "https://m.media-amazon.com/images/I/81P+P+P+L+L._AC_SL1500_.jpg", 1850000},
		{"Murders at Karlov Manor Commander Deck", "MKC", "https://m.media-amazon.com/images/I/71R+R+R+L+L._AC_SL1500_.jpg", 180000},
	}
	for _, s := range mtgSealed {
		var pID string
		db.QueryRow(`INSERT INTO product (name, tcg, category, set_name, set_code, price_source, price_cop_override, stock, image_url)
				 VALUES ($1, 'mtg', 'sealed', $2, $3, 'manual', $4, 5, $5) RETURNING id`, s.Name, s.Name, s.Set, s.Price, s.Img).Scan(&pID)
		db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 5)`, pID, storageIDs[0])
		db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats["new-arrivals"])
	}

	// 3. Pokémon Singles & Sealed
	logger.Info("Seeding Pokémon...")
	pkmnItems := []struct {
		Name, Cat, Set, Img string
		Price               float64
	}{
		{"Charizard ex", "singles", "151", "https://images.pokemontcg.io/sv3pt5/199_hires.png", 550000},
		{"Pikachu with Grey Felt Hat", "singles", "Promo", "https://images.pokemontcg.io/svnp/85_hires.png", 800000},
		{"Scarlet & Violet 151 Elite Trainer Box", "sealed", "151", "https://m.media-amazon.com/images/I/71S+S+S+L+L._AC_SL1500_.jpg", 280000},
		{"Crown Zenith Special Collection", "sealed", "CRZ", "https://m.media-amazon.com/images/I/81W+W+W+L+L._AC_SL1500_.jpg", 160000},
	}
	for _, p := range pkmnItems {
		var pID string
		db.QueryRow(`INSERT INTO product (name, tcg, category, set_name, set_code, price_source, price_cop_override, stock, image_url)
				 VALUES ($1, 'pokemon', $2, $3, $4, 'manual', $5, 10, $6) RETURNING id`, p.Name, p.Cat, p.Set, p.Set, p.Price, p.Img).Scan(&pID)
		db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 10)`, pID, storageIDs[rand.Intn(len(storageIDs))])
		db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats["hot-items"])
	}

	// 4. Yu-Gi-Oh Singles & Sealed
	logger.Info("Seeding Yu-Gi-Oh...")
	ygoItems := []struct {
		Name, Cat, Set, Img string
		Price               float64
	}{
		{"Blue-Eyes White Dragon (25th Anniversary)", "singles", "LOB", "https://m.media-amazon.com/images/I/51R+R+R+L+L.jpg", 120000},
		{"Dark Magician (Ghost Rare)", "singles", "GFP2", "https://m.media-amazon.com/images/I/51G+G+G+L+L.jpg", 450000},
		{"Legendary Collection: 25th Anniversary Edition", "sealed", "LC01", "https://m.media-amazon.com/images/I/81L+L+L+L+L._AC_SL1500_.jpg", 145000},
	}
	for _, y := range ygoItems {
		var pID string
		db.QueryRow(`INSERT INTO product (name, tcg, category, set_name, set_code, price_source, price_cop_override, stock, image_url)
				 VALUES ($1, 'yugioh', $2, $3, $4, 'manual', $5, 8, $6) RETURNING id`, y.Name, y.Cat, y.Set, y.Set, y.Price, y.Img).Scan(&pID)
		db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 8)`, pID, storageIDs[1])
		db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats["featured"])
	}

	// 5. Accessories
	logger.Info("Seeding Accessories...")
	accs := []struct {
		Name, Img string
		Price     float64
	}{
		{"Dragon Shield Matte - Jet Black", "https://m.media-amazon.com/images/I/71J+J+J+L+L._AC_SL1500_.jpg", 55000},
		{"Ultimate Guard Flip'n'Tray 100+ - XenoSkin Blue", "https://m.media-amazon.com/images/I/71B+B+B+L+L._AC_SL1500_.jpg", 115000},
		{"Gamegenic Squire 100+ Convertible - Red", "https://m.media-amazon.com/images/I/61G+G+G+L+L._AC_SL1500_.jpg", 85000},
	}
	for _, a := range accs {
		var pID string
		db.QueryRow(`INSERT INTO product (name, tcg, category, set_name, set_code, price_source, price_cop_override, stock, image_url)
				 VALUES ($1, 'accessories', 'accessories', 'N/A', 'N/A', 'manual', $2, 20, $3) RETURNING id`, a.Name, a.Price, a.Img).Scan(&pID)
		db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 20)`, pID, storageIDs[2])
		db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats["sale"])
	}

	// 6. Store Exclusives (Decks)
	logger.Info("Seeding Store Exclusives (Decks)...")

	// Helper for seeding deck cards with Scryfall metadata
	seedDeckCards := func(deckID string, identifiers []external.CardIdentifier) {
		logger.Info("Looking up metadata for %d deck cards (Resilient Mode)...", len(identifiers))
		inserted := 0

		for _, identifier := range identifiers {
			// Singular lookup with multiple fallbacks (Exact -> Name/Set -> Fuzzy -> Global)
			res, err := external.LookupMTGCard(identifier.ScryfallID, identifier.Name, identifier.Set, identifier.CollectorNumber, "non_foil")

			if err != nil {
				logger.Error("  ❌ Card NOT found: %s (%s #%s): %v", identifier.Name, identifier.Set, identifier.CollectorNumber, err)
				continue
			}

			r := *res
			_, err = db.Exec(`
				INSERT INTO deck_card (
					product_id, name, set_name, set_code, collector_number, quantity, 
					language, color_identity, cmc, is_legendary, is_historic, is_land, is_basic_land,
					art_variation, oracle_text, artist, type_line, border_color, frame,
					full_art, textless, promo_type, image_url, foil_treatment, card_treatment, rarity,
					scryfall_id, legalities
				)
				VALUES ($1, $2, $3, $4, $5, 1, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, 'non_foil', 'normal', $23, $24, $25)
			`,
				deckID, r.Name, r.SetName, r.SetCode, r.CollectorNumber,
				r.Language, r.ColorIdentity, r.CMC, r.IsLegendary, r.IsHistoric, r.IsLand, r.IsBasicLand,
				r.ArtVariation, r.OracleText, r.Artist, r.TypeLine, r.BorderColor, r.Frame,
				r.FullArt, r.Textless, r.PromoType, r.ImageURL, r.Rarity, r.ScryfallID, r.Legalities)

			if err != nil {
				logger.Error("  ❌ Database error for '%s': %v", r.Name, err)
			} else {
				inserted++
			}
		}
		logger.Info("  ✅ Successfully seeded %d/%d cards for product %s", inserted, len(identifiers), deckID)
	}

	var deckID string
	err := db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description)
		VALUES ($1, 'mtg', 'store_exclusives', 'manual', $2, 5, $3, $4) RETURNING id
	`,
		"Custom Commander Precon: Goblin Swarm",
		150000,
		"https://cards.scryfall.io/art_crop/front/0/e/0e8f6e6e-7c35-4281-b31c-266382f052cc.jpg",
		"A highly synergistic 100-card goblin deck ready to play out of the box. Includes tokens and a deck box.",
	).Scan(&deckID)

	if err == nil {
		db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 5)`, deckID, storageIDs[0])
		db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, deckID, cats["featured"])

		seedDeckCards(deckID, []external.CardIdentifier{
			{Name: "Krenko, Mob Boss", Set: "rvr", CollectorNumber: "114"},
			{Name: "Goblin Chieftain", Set: "jmp", CollectorNumber: "324"},
			{Name: "Goblin Warchief", Set: "dom", CollectorNumber: "130"},
			{Name: "Mountain", Set: "usg", CollectorNumber: "343"},
		})
	}

	// 6b. Premium Dragon Deck (100 Cards)
	logger.Info("Seeding Premium Commander Deck (Dragon Hoard)...")
	var dragonDeckID string
	err = db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description)
		VALUES ($1, 'mtg', 'store_exclusives', 'manual', $2, 3, $3, $4) RETURNING id
	`,
		"Premium Commander: Dragon Hoard",
		450000,
		"https://cards.scryfall.io/art_crop/front/4/8/48002002-0002-48a4-a3ad-0002b8004f1a.jpg",
		"A high-power 100-card Commander deck centered around the Ur-Dragon and Miirym. Includes rare dragons, fetch lands, and dual lands.",
	).Scan(&dragonDeckID)

	if err == nil {
		db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 3)`, dragonDeckID, storageIDs[1])
		db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, dragonDeckID, cats["hot-items"])

		dragonDeckIds := []external.CardIdentifier{
			// Lands (35 Unique)
			{Name: "Ancient Tomb", Set: "tmp", CollectorNumber: "315"},
			{Name: "Cavern of Souls", Set: "avr", CollectorNumber: "226"},
			{Name: "City of Brass", Set: "mma", CollectorNumber: "221"},
			{Name: "Mana Confluence", Set: "jou", CollectorNumber: "163"},
			{Name: "Command Tower", Set: "cm2", CollectorNumber: "242"},
			{Name: "Reflecting Pool", Set: "tpr", CollectorNumber: "241"},
			{Name: "Forbidden Orchard", Set: "chk", CollectorNumber: "276"},
			{Name: "Exotic Orchard", Set: "cn2", CollectorNumber: "219"},
			{Name: "Path of Ancestry", Set: "c17", CollectorNumber: "63"},
			{Name: "Haven of the Spirit Dragon", Set: "dtk", CollectorNumber: "249"},
			{Name: "Crucible of the Spirit Dragon", Set: "frf", CollectorNumber: "167"},
			{Name: "Stomping Ground", Set: "rna", CollectorNumber: "259"},
			{Name: "Steam Vents", Set: "grn", CollectorNumber: "257"},
			{Name: "Watery Grave", Set: "grn", CollectorNumber: "259"},
			{Name: "Blood Crypt", Set: "rna", CollectorNumber: "245"},
			{Name: "Overgrown Tomb", Set: "grn", CollectorNumber: "253"},
			{Name: "Temple Garden", Set: "grn", CollectorNumber: "258"},
			{Name: "Hallowed Fountain", Set: "rna", CollectorNumber: "251"},
			{Name: "Sacred Foundry", Set: "grn", CollectorNumber: "254"},
			{Name: "Breeding Pool", Set: "rna", CollectorNumber: "246"},
			{Name: "Godless Shrine", Set: "rna", CollectorNumber: "248"},
			{Name: "Wooded Foothills", Set: "ons", CollectorNumber: "330"},
			{Name: "Polluted Delta", Set: "ons", CollectorNumber: "335"},
			{Name: "Bloodstained Mire", Set: "ons", CollectorNumber: "313"},
			{Name: "Windswept Heath", Set: "ons", CollectorNumber: "328"},
			{Name: "Flooded Strand", Set: "ons", CollectorNumber: "316"},
			{Name: "Scalding Tarn", Set: "zen", CollectorNumber: "223"},
			{Name: "Marsh Flats", Set: "zen", CollectorNumber: "219"},
			{Name: "Verdant Catacombs", Set: "zen", CollectorNumber: "229"},
			{Name: "Misty Rainforest", Set: "zen", CollectorNumber: "220"},
			{Name: "Arid Mesa", Set: "zen", CollectorNumber: "211"},
			{Name: "Mountain", Set: "clb", CollectorNumber: "453"},
			{Name: "Mountain", Set: "clb", CollectorNumber: "454"},
			{Name: "Island", Set: "clb", CollectorNumber: "451"},
			{Name: "Forest", Set: "clb", CollectorNumber: "457"},

			// Creatures (30 Unique)
			{Name: "The Ur-Dragon", Set: "c17", CollectorNumber: "48"},
			{Name: "Miirym, Sentinel Wyrm", Set: "clb", CollectorNumber: "284"},
			{Name: "Lathliss, Dragon Queen", Set: "m19", CollectorNumber: "149"},
			{Name: "Korvold, Fae-Cursed King", Set: "eld", CollectorNumber: "329"},
			{Name: "Utvara Hellkite", Set: "rtr", CollectorNumber: "110"},
			{Name: "Scourge of Valkas", Set: "m14", CollectorNumber: "151"},
			{Name: "Balefire Dragon", Set: "isd", CollectorNumber: "129"},
			{Name: "Terror of the Peaks", Set: "m21", CollectorNumber: "164"},
			{Name: "Hellkite Tyrant", Set: "gtc", CollectorNumber: "94"},
			{Name: "Old Gnawbone", Set: "afr", CollectorNumber: "197"},
			{Name: "Kyodai, Soul of Kamigawa", Set: "neo", CollectorNumber: "23"},
			{Name: "Dragonlord Dromoka", Set: "dtk", CollectorNumber: "217"},
			{Name: "Dragonlord Silumgar", Set: "dtk", CollectorNumber: "220"},
			{Name: "Dragonlord Kolaghan", Set: "dtk", CollectorNumber: "218"},
			{Name: "Dragonlord Atarka", Set: "dtk", CollectorNumber: "216"},
			{Name: "Dragonlord Ojutai", Set: "dtk", CollectorNumber: "219"},
			{Name: "Silumgar, the Drifting Death", Set: "frf", CollectorNumber: "154"},
			{Name: "Kolaghan, the Storm's Fury", Set: "frf", CollectorNumber: "155"},
			{Name: "Atarka, World Render", Set: "frf", CollectorNumber: "149"},
			{Name: "Ojutai, Soul of Winter", Set: "frf", CollectorNumber: "156"},
			{Name: "Dromoka, the Eternal", Set: "frf", CollectorNumber: "151"},
			{Name: "Tiamat", Set: "afr", CollectorNumber: "235"},
			{Name: "Ancient Silver Dragon", Set: "clb", CollectorNumber: "57"},
			{Name: "Ancient Copper Dragon", Set: "clb", CollectorNumber: "161"},
			{Name: "Ancient Brass Dragon", Set: "clb", CollectorNumber: "111"},
			{Name: "Ancient Gold Dragon", Set: "clb", CollectorNumber: "3"},
			{Name: "Ancient Bronze Dragon", Set: "clb", CollectorNumber: "214"},
			{Name: "Klauth, Unrivaled Ancient", Set: "afc", CollectorNumber: "50"},
			{Name: "Rivaz of the Claw", Set: "dmu", CollectorNumber: "215"},
			{Name: "Dragonborn Looter", Set: "clb", CollectorNumber: "65"},

			// Artifacts/Enchantments (15 Unique)
			{Name: "Sol Ring", Set: "clb", CollectorNumber: "882"},
			{Name: "Arcane Signet", Set: "clb", CollectorNumber: "861"},
			{Name: "Dragon's Hoard", Set: "m19", CollectorNumber: "232"},
			{Name: "Herald's Horn", Set: "c17", CollectorNumber: "53"},
			{Name: "Urza's Incubator", Set: "vma", CollectorNumber: "287"},
			{Name: "Chromatic Lantern", Set: "grn", CollectorNumber: "233"},
			{Name: "The Great Henge", Set: "eld", CollectorNumber: "161"},
			{Name: "Temur Ascendancy", Set: "ktk", CollectorNumber: "207"},
			{Name: "Dragon Tempest", Set: "dtk", CollectorNumber: "136"},
			{Name: "Kindred Discovery", Set: "c17", CollectorNumber: "11"},
			{Name: "Rhythm of the Wild", Set: "rna", CollectorNumber: "201"},
			{Name: "Smothering Tithe", Set: "rna", CollectorNumber: "22"},
			{Name: "Rhystic Study", Set: "pcy", CollectorNumber: "45"},
			{Name: "Sylvan Library", Set: "ema", CollectorNumber: "187"},
			{Name: "Garruk's Uprising", Set: "m21", CollectorNumber: "186"},

			// Instants/Sorceries (20 Unique)
			{Name: "Swords to Plowshares", Set: "clb", CollectorNumber: "707"},
			{Name: "Path to Exile", Set: "2xm", CollectorNumber: "25"},
			{Name: "Cyclonic Rift", Set: "rtr", CollectorNumber: "35"},
			{Name: "Heroic Intervention", Set: "aer", CollectorNumber: "109"},
			{Name: "Teferi's Protection", Set: "c17", CollectorNumber: "8"},
			{Name: "Counterspell", Set: "clb", CollectorNumber: "441"},
			{Name: "Vampiric Tutor", Set: "cmr", CollectorNumber: "156"},
			{Name: "Enlightened Tutor", Set: "ema", CollectorNumber: "9"},
			{Name: "Worldly Tutor", Set: "cc1", CollectorNumber: "6"},
			{Name: "Mystical Tutor", Set: "ema", CollectorNumber: "62"},
			{Name: "Cultivate", Set: "clb", CollectorNumber: "821"},
			{Name: "Kodama's Reach", Set: "clb", CollectorNumber: "835"},
			{Name: "Farseek", Set: "rav", CollectorNumber: "164"},
			{Name: "Three Visits", Set: "cmr", CollectorNumber: "261"},
			{Name: "Nature's Lore", Set: "por", CollectorNumber: "184"},
			{Name: "Blasphemous Act", Set: "clb", CollectorNumber: "788"},
			{Name: "Toxic Deluge", Set: "2xm", CollectorNumber: "110"},
			{Name: "Demonic Tutor", Set: "vma", CollectorNumber: "116"},
			{Name: "Farewell", Set: "neo", CollectorNumber: "13"},
			{Name: "Crux of Fate", Set: "frf", CollectorNumber: "65"},
		}
		seedDeckCards(dragonDeckID, dragonDeckIds)
	}

	// 6c. Store Exclusives: Wooden Tokens
	logger.Info("Seeding Store Exclusives: Wooden Tokens...")
	var tokensID string
	err = db.QueryRow(`
		INSERT INTO product (name, tcg, category, price_source, price_cop_override, stock, image_url, description)
		VALUES ($1, 'mtg', 'store_exclusives', 'manual', $2, 20, $3, $4) RETURNING id
	`,
		"Hand-Crafted Wooden Tokens (Set of 10)",
		45000,
		"https://images.unsplash.com/photo-1598214886806-c87b84b7078b?q=80&w=800&auto=format&fit=crop",
		"Beautifully laser-engraved wooden tokens for tracking various MTG status effects and creatures. Includes 2x Goblin, 2x Zombie, 2x Treasure, and 4x Generic +1/+1 counters.",
	).Scan(&tokensID)

	if err == nil {
		db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 20)`, tokensID, storageIDs[0])
		db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, tokensID, cats["featured"])
	}

	// 7. Bounties
	logger.Info("Seeding bounties...")
	bountyNames := []string{"Orcish Bowmasters", "Sheoldred, the Apocalypse", "The One Ring", "Ragavan, Nimble Pilferer", "Mana Crypt"}
	for _, name := range bountyNames {
		db.Exec(`
			INSERT INTO bounty (name, tcg, set_name, condition, quantity_needed, target_price, is_active, foil_treatment, language)
			VALUES ($1, 'mtg', 'Modern Horizons 3', 'NM', $2, $3, true, 'non_foil', 'en')
		`, name, rand.Intn(5)+2, (rand.Intn(10)+5)*10000)
	}

	for i := 0; i < 10; i++ {
		db.Exec(`
			INSERT INTO bounty (name, tcg, set_name, condition, quantity_needed, target_price, is_active, language)
			VALUES ($1, 'mtg', 'Generic Set', 'NM', $2, $3, true, 'en')
		`, fmt.Sprintf("Bounty Card #%d", i), rand.Intn(10)+1, rand.Intn(100000)+5000)
	}

	// 7. Client Requests
	logger.Info("Seeding client requests...")
	for i := 0; i < 12; i++ {
		status := []string{"pending", "accepted", "rejected", "solved"}[i%4]
		db.Exec(`
			INSERT INTO client_request (customer_name, customer_contact, card_name, set_name, details, status)
			VALUES ($1, $2, $3, 'MH3', 'Looking for this card in foil.', $4)
		`, fmt.Sprintf("Client %d", i), "+57 310 111 2233", fmt.Sprintf("Request Card %d", i), status)
	}

	// 8. Customers & Orders
	logger.Info("Seeding customers and history...")
	for i := 0; i < 50; i++ {
		var cID string
		db.QueryRow(`
			INSERT INTO customer (first_name, last_name, email, phone)
			VALUES ($1, $2, $3, $4) RETURNING id
		`, fmt.Sprintf("User%d", i), "Test", fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("300%07d", i)).Scan(&cID)

		numOrders := rand.Intn(4) + 1
		if i < 5 {
			numOrders = rand.Intn(10) + 10
		} // Create some VIP customers with 10-20 orders
		for j := 0; j < numOrders; j++ {
			var oID string
			status := "completed"
			if rand.Intn(10) > 8 {
				status = "pending"
			}

			orderDate := time.Now().AddDate(0, 0, -rand.Intn(60))
			completedAt := "NULL"
			if status == "completed" {
				completedAt = fmt.Sprintf("'%s'", orderDate.Add(time.Hour*24).Format(time.RFC3339))
			}

			db.QueryRow(fmt.Sprintf(`
				INSERT INTO "order" (order_number, customer_id, status, payment_method, total_cop, created_at, completed_at)
				VALUES ($1, $2, $3, 'Transfer', $4, $5, %s) RETURNING id
			`, completedAt), fmt.Sprintf("ORD-%d-%d", i, j), cID, status, 0.0, orderDate).Scan(&oID)

			// Items
			total := 0.0
			for k := 0; k < rand.Intn(3)+1; k++ {
				if len(productIDs) == 0 {
					break
				}
				pID := productIDs[rand.Intn(len(productIDs))]
				var pName, pSet string
				var pPrice float64
				db.QueryRow("SELECT name, set_name, price_cop_override FROM product WHERE id = $1", pID).Scan(&pName, &pSet, &pPrice)

				qty := rand.Intn(2) + 1
				var pCond, pFoil, pTreat string
				db.QueryRow("SELECT condition, foil_treatment, card_treatment FROM product WHERE id = $1", pID).Scan(&pCond, &pFoil, &pTreat)

				db.Exec(`
					INSERT INTO order_item (order_id, product_id, product_name, product_set, unit_price_cop, quantity, condition, foil_treatment, card_treatment)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`, oID, pID, pName, pSet, pPrice, qty, pCond, pFoil, pTreat)
				total += pPrice * float64(qty)
			}
			db.Exec(`UPDATE "order" SET total_cop = $1 WHERE id = $2`, total, oID)
		}
	}
	return productIDs
}

func seedNotices(db *sqlx.DB, productIDs []string) {
	logger.Info("Seeding notices (blog/news)...")

	p1Id := "PROD-1"
	if len(productIDs) > 0 {
		p1Id = productIDs[0]
	}
	p2Id := "PROD-2"
	if len(productIDs) > 1 {
		p2Id = productIDs[1]
	}

	notices := []struct{ Title, Slug, HTML, Img string }{
		{
			Title: "Modern Horizons 3: The Full Spoiler is Here!",
			Slug:  "mh3-full-spoiler",
			Img:   "https://cards.scryfall.io/art_crop/front/1/2/12128594-82ea-4a4b-9e4a-464817a02796.jpg",
			HTML: `
				<h2>Prepare your wallets!</h2>
				<p>The latest Magic set is finally revealed in its entirety. We're seeing some incredible reprints and brand new powerhouses for Modern and Commander alike.</p>
				<h3>Top Picks</h3>
				<ul>
					<li><b>Eldrazi Support:</b> Massive titans are back.</li>
					<li><b>New Elementals:</b> A new cycle of evoke-like effects.</li>
				</ul>
				<div style="margin: 2rem 0; border: 2px solid var(--gold); padding: 1rem; background: var(--ink-surface);">
					<p style="margin:0; font-size: 0.8rem;"><b>CHECK THIS OUT:</b> We already have <a data-card-id="` + p1Id + `">This Incredible Card</a> in stock!</p>
				</div>
				<p>Stay tuned for our pre-release events next weekend.</p>
				<iframe width="100%" height="315" src="https://www.youtube.com/embed/dQw4w9WgXcQ" frameborder="0" allowfullscreen></iframe>
			`,
		},
		{
			Title: "New Pokémon 151 Restock Incoming",
			Slug:  "pokemon-151-restock",
			Img:   "https://images.pokemontcg.io/sv3pt5/logo.png",
			HTML: `
				<h2>The Kanto starters are back.</h2>
				<p>We've secured a massive shipment of Elite Trainer Boxes and Booster Bundles for the 151 set. Don't miss your chance to complete your master set!</p>
				<p>Available starting this Friday. We'll also have a limited supply of <a data-card-id="` + p2Id + `">special promo cards</a> for the first 10 customers.</p>
				<img src="https://images.pokemontcg.io/sv3pt5/199_hires.png" alt="Charizard" style="width: 200px; display: block; margin: 1rem auto;" />
			`,
		},
		{
			Title: "Tournament Report: Regional Qualifiers",
			Slug:  "regional-qualifiers-report",
			Img:   "https://cards.scryfall.io/art_crop/front/0/e/0e8f6e6e-7c35-4281-b31c-266382f052cc.jpg",
			HTML: `
				<h2>What a weekend!</h2>
				<p>Over 100 players gathered for our RQ last Saturday. The meta was dominated by combo decks, but a rogue aggro pilot took down the whole thing.</p>
				<blockquote>"I didn't expect to win, I just played what I liked." - Winner</blockquote>
				<p>Next big event is the Store Championship in July!</p>
			`,
		},
		{
			Title: "Collector's Corner: Grading your cards",
			Slug:  "grading-guide",
			Img:   "https://cards.scryfall.io/art_crop/front/a/e/ae56ce7c-b31c-266382f052cc.jpg",
			HTML: `
				<h2>Is that 10 worth it?</h2>
				<p>Many customers ask if they should grade their pulls. Here's our quick guide on what to look for: centering, surface, edges, and corners.</p>
				<p>We provide a pre-grading service here at the shop to help you decide!</p>
			`,
		},
		{
			Title: "Weekly Shop Update",
			Slug:  "weekly-update-march-30",
			Img:   "https://cards.scryfall.io/art_crop/front/f/1/f1911e6e-7c35-4281-b31c-266382f052cc.jpg",
			HTML: `
				<h2>New space and better lighting!</h2>
				<p>We've finished redecorating the play area. Come by and see the new displays. We also have a new coffee machine for those long tournament rounds.</p>
				<p>See you all this week!</p>
			`,
		},
		{
			Title: "Upcoming Tournament: Pokémon Regional Qualifier",
			Slug:  "pokemon-reg-qualifier-2026",
			Img:   "https://images.unsplash.com/photo-1613771404721-1f92d799e49f?q=80&w=800&auto=format&fit=crop",
			HTML:  "<h2>Battle for the Top!</h2><p>Registration opens this Friday. Limited slots available. Format: Standard.</p>",
		},
		{
			Title: "Weekly Deal: Buy 3 Boosters, Get 1 Free!",
			Slug:  "weekly-deal-boosters",
			Img:   "https://images.unsplash.com/photo-1541560052753-107f96307409?q=80&w=800&auto=format&fit=crop",
			HTML:  "<p>This week only, buy any 3 TCG boosters and get the 4th one free. Valid across MTG, Pokémon, and Yu-Gi-Oh.</p>",
		},
		{
			Title: "Trading Corner: Bulk Trade-in Event",
			Slug:  "bulk-trade-in-event",
			Img:   "https://images.unsplash.com/photo-1598214886806-c87b84b7078b?q=80&w=800&auto=format&fit=crop",
			HTML:  "<p>Turn your bulk commons and uncommons into store credit! We're running a special trade-in event this Saturday starting at 10 AM.</p>",
		},
	}

	for _, n := range notices {
		_, err := db.Exec(`
			INSERT INTO notice (title, slug, content_html, featured_image_url)
			VALUES ($1, $2, $3, $4)
		`, n.Title, n.Slug, n.HTML, n.Img)
		if err != nil {
			logger.Error("Failed to seed notice '%s': %v", n.Title, err)
		}
	}
}

func seedCRM(db *sqlx.DB, adminID string) {
	logger.Info("Seeding CRM data (subscribers, notes, requests, and offers)...")

	// 1. Fetch some customers to link data to
	var customers []struct {
		ID        string `db:"id"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}
	db.Select(&customers, "SELECT id, first_name, last_name, email FROM customer LIMIT 20")

	// 2. Seed some newsletter subscribers
	for i, c := range customers {
		if i%3 == 0 { // ~33% signup rate
			db.Exec(`
				INSERT INTO newsletter_subscriber (email, customer_id)
				VALUES ($1, $2) ON CONFLICT DO NOTHING
			`, c.Email, c.ID)
		}

		// Seed some OAuth accounts for multi-account testing
		if i < 5 {
			db.Exec(`INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, 'google', $2)`, c.ID, "google-id-"+c.ID)
			if i%2 == 0 {
				db.Exec(`INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, 'facebook', $2)`, c.ID, "facebook-id-"+c.ID)
			}
		}
	}

	// 3. Seed Client Requests (Linked to top customers)
	requestTemplates := []struct{ Card, Details string }{
		{"Black Lotus", "Looking for a budget Unlimited edition."},
		{"Charizard Base Set", "PSA 8 or higher preferred."},
		{"Sheoldred, the Apocalypse", "Need 4 copies for a tournament."},
		{"Mewtwo GX SV107", "Shiny Vault version only."},
		{"Ragavan, Nimble Pilferer", "Borderless art if possible."},
		{"The One Ring", "Special edition for my collection."},
		{"Sol Ring", "Looking for Masterpiece version."},
		{"Pikachu Illustrator", "Just kidding, unless you have it!"},
	}

	for i, c := range customers {
		if i >= 10 { // Only link first 10
			break
		}
		template := requestTemplates[i%len(requestTemplates)]
		status := []string{"pending", "accepted", "solved"}[i%3]

		db.Exec(`
			INSERT INTO client_request (customer_id, customer_name, customer_contact, card_name, details, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, c.ID, c.FirstName+" "+c.LastName, c.Email, template.Card, template.Details, status)
	}

	// 4. Seed Bounty Offers (Linked to top customers)
	var bountyIDs []string
	db.Select(&bountyIDs, "SELECT id FROM bounty LIMIT 10")

	if len(bountyIDs) > 0 {
		for i, c := range customers {
			if i >= 12 { // Link more offers than requests
				break
			}
			bID := bountyIDs[i%len(bountyIDs)]
			status := []string{"pending", "accepted", "fulfilled"}[i%3]

			createdAt := time.Now().AddDate(0, 0, -rand.Intn(45))
			db.Exec(`
				INSERT INTO bounty_offer (bounty_id, customer_id, quantity, status, admin_notes, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, bID, c.ID, (rand.Intn(3) + 1), status, "Verified collection. Quality looks good.", createdAt)
		}
	}

	// 5. Seed extra data for Accounting test (linked Offer + Request)
	if len(bountyIDs) > 0 {
		// Specifically seed a "cancelled out" pair
		// Find the "Orcish Bowmasters" bounty
		var bowmastersID string
		db.Get(&bowmastersID, "SELECT id FROM bounty WHERE name = 'Orcish Bowmasters' LIMIT 1")
		if bowmastersID != "" {
			c := customers[rand.Intn(len(customers))]
			// Outcome
			db.Exec(`
				INSERT INTO bounty_offer (bounty_id, customer_id, quantity, status, admin_notes, created_at)
				VALUES ($1, $2, 1, 'fulfilled', 'Linked transaction test', $3)
			`, bowmastersID, c.ID, time.Now().AddDate(0, 0, -5))

			// Income
			db.Exec(`
				INSERT INTO client_request (customer_name, customer_contact, card_name, status, created_at)
				VALUES ($1, $2, 'Orcish Bowmasters', 'solved', $3)
			`, c.FirstName+" "+c.LastName, c.Email, time.Now().AddDate(0, 0, -5))
		}
	}

	// 6. Seed some guest requests (Unlinked)
	for i := 0; i < 5; i++ {
		db.Exec(`
			INSERT INTO client_request (customer_name, customer_contact, card_name, details, status)
			VALUES ($1, $2, $3, $4, 'pending')
		`, fmt.Sprintf("Guest %d", i), fmt.Sprintf("guest%d@example.com", i), "Some Random Card", "I'm not a registered user yet.")
	}

	// 6. Seed Customer Notes
	interactions := []string{
		"Spoke to customer via WhatsApp. Interested in MH3 boosters.",
		"Requested a custom quote for bulk collection.",
		"Note: Preferred shipping method is via courier.",
		"Verified identity document for high-value order.",
		"Resolved condition complaint on previous order.",
		"Asking for pre-order availability of next set.",
		"Loyal customer. Provided discount code.",
		"Confirmed receipt of package. Very satisfied.",
	}

	for i, c := range customers {
		noteChance := 5
		if i < 5 {
			noteChance = 1
		} // Every top customer gets a note

		if i%noteChance == 0 {
			db.Exec(`
				INSERT INTO customer_note (customer_id, content, admin_id)
				VALUES ($1, $2, $3)
			`, c.ID, interactions[rand.Intn(len(interactions))], adminID)
		}
	}
}

func seedThemes(db *sqlx.DB) {
	logger.Info("🎨 Seeding Theme Palette Pack...")

	ptr := func(s string) *string { return &s }

	themes := []models.Theme{
		{
			ID: "00000000-0000-0000-0000-000000000001", Name: "Cardboard", IsSystem: true,
			BgPage: "#e6dac3", BgHeader: "#1a1f2e", BgSurface: "#fdfbf7", BgCard: "#ffffff",
			TextMain: "#3b3127", TextSecondary: "#5c4e3d", TextMuted: "#5c4e3d", TextOnAccent: "#2c251d", TextOnHeader: "#ffffff",
			AccentPrimary: "#d4af37", AccentPrimaryHover: "#b8961e", BorderMain: "#d4c5ab", BorderFocus: "#3b3127",
			StatusNM: "#2e7d32", StatusLP: "#558b2f", StatusMP: "#ef6c00", StatusHP: "#c62828", StatusDMG: "#455a64",
			AccentHeader: "#fbbf24", StatusHPHeader: "#f87171",
			BtnPrimaryBg: "#1a1f2e", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#3b3127",
			CheckboxBorder: "#8b795c", CheckboxChecked: "#d4af37",
			RadiusBase: "8px", PaddingCard: "12px", GapGrid: "16px",
			BgImageURL: ptr(""), FontHeading: ptr("Inter, sans-serif"), FontBody: ptr("Inter, sans-serif"), AccentSecondary: ptr(""),
		},
		{
			ID: "00000000-0000-0000-0000-000000000002", Name: "Obsidiana", IsSystem: true,
			BgPage: "#0a0a0a", BgHeader: "#121212", BgSurface: "#1a1a1a", BgCard: "#1a1a1a",
			TextMain: "#f8fafc", TextSecondary: "#cbd5e1", TextMuted: "#94a3b8", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#3b82f6", AccentPrimaryHover: "#2563eb", BorderMain: "#475569", BorderFocus: "#3b82f6",
			StatusNM: "#10b981", StatusLP: "#fbbf24", StatusMP: "#f59e0b", StatusHP: "#ef4444", StatusDMG: "#64748b",
			AccentHeader: "#60a5fa", StatusHPHeader: "#f87171",
			BtnPrimaryBg: "#3b82f6", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#f8fafc",
			CheckboxBorder: "#475569", CheckboxChecked: "#3b82f6",
			RadiusBase: "2px", PaddingCard: "14px", GapGrid: "20px",
			BgImageURL: ptr(""), FontHeading: ptr("Inter, sans-serif"), FontBody: ptr("Inter, sans-serif"), AccentSecondary: ptr(""),
		},
		{
			ID: "00000000-0000-0000-0000-000000000003", Name: "Yule", IsSystem: true,
			BgPage: "#052e16", BgHeader: "#991b1b", BgSurface: "#064e3b", BgCard: "#064e3b",
			TextMain: "#f0fdf4", TextSecondary: "#d1fae5", TextMuted: "#86efac", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#fbbf24", AccentPrimaryHover: "#f59e0b", BorderMain: "#1e6e3d", BorderFocus: "#f59e0b",
			StatusNM: "#4ade80", StatusLP: "#fbbf24", StatusMP: "#f97316", StatusHP: "#ef4444", StatusDMG: "#991b1b",
			AccentHeader: "#fbbf24", StatusHPHeader: "#fecaca",
			BtnPrimaryBg: "#991b1b", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#f0fdf4",
			CheckboxBorder: "#166534", CheckboxChecked: "#fbbf24",
			RadiusBase: "12px", PaddingCard: "12px", GapGrid: "16px",
			BgImageURL: ptr(""), FontHeading: ptr("Inter, sans-serif"), FontBody: ptr("Inter, sans-serif"), AccentSecondary: ptr(""),
		},
		{
			ID: "00000000-0000-0000-0000-000000000004", Name: "Spring Egg", IsSystem: true,
			BgPage: "#fffbea", BgHeader: "#f5f3ff", BgSurface: "#ffffff", BgCard: "#ffffff",
			TextMain: "#2e1065", TextSecondary: "#3b0764", TextMuted: "#581c87", TextOnAccent: "#ffffff", TextOnHeader: "#2e1065",
			AccentPrimary: "#8b5cf6", AccentPrimaryHover: "#a78bfa", BorderMain: "#94a3b8", BorderFocus: "#a78bfa",
			StatusNM: "#10b981", StatusLP: "#fbbf24", StatusMP: "#f59e0b", StatusHP: "#ef4444", StatusDMG: "#94a3b8",
			AccentHeader: "#7c3aed", StatusHPHeader: "#b91c1c",
			BtnPrimaryBg: "#8b5cf6", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#2e1065",
			CheckboxBorder: "#a78bfa", CheckboxChecked: "#8b5cf6",
			RadiusBase: "24px", PaddingCard: "16px", GapGrid: "24px",
			BgImageURL: ptr(""), FontHeading: ptr("Inter, sans-serif"), FontBody: ptr("Inter, sans-serif"), AccentSecondary: ptr(""),
		},
		{
			ID: "00000000-0000-0000-0000-000000000005", Name: "Neon Flux", IsSystem: true,
			BgPage: "#020617", BgHeader: "#0f172a", BgSurface: "#020617", BgCard: "#020617",
			TextMain: "#f8fafc", TextSecondary: "#94a3b8", TextMuted: "#64748b", TextOnAccent: "#000000", TextOnHeader: "#ffffff",
			AccentPrimary: "#22c55e", AccentPrimaryHover: "#4ade80", BorderMain: "#334155", BorderFocus: "#4ade80",
			StatusNM: "#22c55e", StatusLP: "#eab308", StatusMP: "#f97316", StatusHP: "#ef4444", StatusDMG: "#64748b",
			AccentHeader: "#4ade80", StatusHPHeader: "#f87171",
			BtnPrimaryBg: "#22c55e", BtnPrimaryText: "#000000", BtnSecondaryBg: "transparent", BtnSecondaryText: "#f8fafc",
			CheckboxBorder: "#334155", CheckboxChecked: "#22c55e",
			RadiusBase: "0px", PaddingCard: "10px", GapGrid: "12px",
			BgImageURL: ptr(""), FontHeading: ptr("Space Mono, monospace"), FontBody: ptr("Space Mono, monospace"), AccentSecondary: ptr(""),
		},
		{
			ID: "00000000-0000-0000-0000-000000000006", Name: "Arena", IsSystem: true,
			BgPage: "#171717", BgHeader: "#262626", BgSurface: "#1c1c1c", BgCard: "#1c1c1c",
			TextMain: "#ffffff", TextSecondary: "#d4d4d4", TextMuted: "#a3a3a3", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#ea580c", AccentPrimaryHover: "#f97316", BorderMain: "#525252", BorderFocus: "#f97316",
			StatusNM: "#22c55e", StatusLP: "#eab308", StatusMP: "#f97316", StatusHP: "#dc2626", StatusDMG: "#404040",
			AccentHeader: "#fb923c", StatusHPHeader: "#f87171",
			BtnPrimaryBg: "#ea580c", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#ffffff",
			CheckboxBorder: "#525252", CheckboxChecked: "#ea580c",
			RadiusBase: "4px", PaddingCard: "14px", GapGrid: "16px",
			BgImageURL: ptr(""), FontHeading: ptr("Inter, sans-serif"), FontBody: ptr("Inter, sans-serif"), AccentSecondary: ptr(""),
		},
		{
			ID: "00000000-0000-0000-0000-000000000007", Name: "Celebrate", IsSystem: true,
			BgPage: "#fdf2f8", BgHeader: "#be185d", BgSurface: "#ffffff", BgCard: "#ffffff",
			TextMain: "#831843", TextSecondary: "#9d174d", TextMuted: "#be185d", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#db2777", AccentPrimaryHover: "#f472b6", BorderMain: "#fce7f3", BorderFocus: "#f472b6",
			StatusNM: "#10b981", StatusLP: "#fbbf24", StatusMP: "#f59e0b", StatusHP: "#ef4444", StatusDMG: "#db2777",
			AccentHeader: "#fdf2f8", StatusHPHeader: "#fce7f3",
			BtnPrimaryBg: "#be185d", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#831843",
			CheckboxBorder: "#f472b6", CheckboxChecked: "#be185d",
			RadiusBase: "16px", PaddingCard: "14px", GapGrid: "20px",
			BgImageURL: ptr(""), FontHeading: ptr("Inter, sans-serif"), FontBody: ptr("Inter, sans-serif"), AccentSecondary: ptr(""),
		},
		{
			ID: "00000000-0000-0000-0000-000000000010", Name: "Strixhaven: Lorehold", IsSystem: true,
			BgPage: "#fafaf9", BgHeader: "#5c1919", BgSurface: "#ffffff", BgCard: "#ffffff",
			TextMain: "#451a03", TextSecondary: "#7c2d12", TextMuted: "#9a3412", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#d97706", AccentPrimaryHover: "#fbbf24", BorderMain: "#f3f4f6", BorderFocus: "#fbbf24",
			StatusNM: "#10b981", StatusLP: "#fbbf24", StatusMP: "#f59e0b", StatusHP: "#ef4444", StatusDMG: "#92400e",
			AccentHeader: "#fef3c7", StatusHPHeader: "#fde68a",
			BtnPrimaryBg: "#7c2d12", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#451a03",
			CheckboxBorder: "#d97706", CheckboxChecked: "#7c2d12",
			RadiusBase: "4px", PaddingCard: "16px", GapGrid: "24px",
			BgImageURL:  ptr("data:image/svg+xml,%3Csvg width='20' height='20' viewBox='0 0 20 20' xmlns='http://www.w3.org/2000/svg'%3E%3Ccircle cx='2' cy='2' r='2' fill='%23a83232' fill-opacity='0.05'/%3E%3C/svg%3E"),
			FontHeading: ptr("Cinzel, serif"), FontBody: ptr("Inter, sans-serif"),
			AccentSecondary: ptr("#e65100"),
		},
		{
			ID: "00000000-0000-0000-0000-000000000011", Name: "Strixhaven: Prismari", IsSystem: true,
			BgPage: "#0f172a", BgHeader: "#7f1d1d", BgSurface: "#1e293b", BgCard: "#1e293b",
			TextMain: "#f8fafc", TextSecondary: "#cbd5e1", TextMuted: "#94a3b8", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#60a5fa", AccentPrimaryHover: "#3b82f6", BorderMain: "#334155", BorderFocus: "#ef4444",
			StatusNM: "#10b981", StatusLP: "#f59e0b", StatusMP: "#f97316", StatusHP: "#fca5a5", StatusDMG: "#475569",
			AccentHeader: "#fef2f2", StatusHPHeader: "#fca5a5",
			BtnPrimaryBg: "linear-gradient(135deg, #3b82f6, #ef4444)", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#f8fafc",
			CheckboxBorder: "#475569", CheckboxChecked: "#3b82f6",
			RadiusBase: "12px", PaddingCard: "14px", GapGrid: "24px",
			BgImageURL:  ptr("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath fill='%23ef4444' fill-opacity='0.1' d='M44.7,-76.4C58.9,-69.2,71.8,-59.1,81.3,-46.3C90.8,-33.5,96.8,-18,95.5,-3C94.2,12,85.6,26.5,75.4,38.8C65.2,51.1,53.4,61.2,40.1,68.6C26.8,76,12,80.7,-2.8,84.9C-17.6,89.1,-35.1,92.8,-49,86.2C-62.9,79.6,-73.2,62.7,-80.6,45.3C-88,27.9,-92.5,10,-91.4,-7.5C-90.3,-25,-83.6,-42,-72.6,-54.6C-61.6,-67.2,-46.3,-75.4,-31.6,-80.5C-16.9,-85.6, -1.2,-87.6,13.4,-84.9C28,-82.2,41.9,-74.8,44.7,-76.4Z' transform='translate(100 100)' /%3E%3C/svg%3E"),
			FontHeading: ptr("Bebas Neue, sans-serif"), FontBody: ptr("Space Mono, monospace"),
			AccentSecondary: ptr("#ef4444"),
		},
		{
			ID: "00000000-0000-0000-0000-000000000021", Name: "Strixhaven: Quandrix", IsSystem: true,
			BgPage: "#f0fdfa", BgHeader: "#064e3b", BgSurface: "#ffffff", BgCard: "#ffffff",
			TextMain: "#042f2e", TextSecondary: "#0f766e", TextMuted: "#115e59", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#14b8a6", AccentPrimaryHover: "#5eead4", BorderMain: "#f3f4f6", BorderFocus: "#5eead4",
			StatusNM: "#10b981", StatusLP: "#fbbf24", StatusMP: "#f59e0b", StatusHP: "#ef4444", StatusDMG: "#115e59",
			AccentHeader: "#ccfbf1", StatusHPHeader: "#99f6e4",
			BtnPrimaryBg: "#064e3b", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#042f2e",
			CheckboxBorder: "#14b8a6", CheckboxChecked: "#064e3b",
			RadiusBase: "12px", PaddingCard: "16px", GapGrid: "20px",
			BgImageURL:  ptr("data:image/svg+xml,%3Csvg width='40' height='40' viewBox='0 0 40 40' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M20 20.5V18H0v-2h20v-2H0v-2h20v-2H0V8h20V6H0V4h20V2H0V0h22v20h18v2H22v18H0v-2h20v-2H0v-2h20v-2H0v-2h20v-2H0v-2h20z' fill='%230f766e' fill-opacity='0.05' fill-rule='evenodd'/%3E%3C/svg%3E"),
			FontHeading: ptr("Space Mono, monospace"), FontBody: ptr("Inter, sans-serif"),
			AccentSecondary: ptr("#10b981"),
		},
		{
			ID: "00000000-0000-0000-0000-000000000013", Name: "Strixhaven: Silverquill", IsSystem: true,
			BgPage: "#fafafa", BgHeader: "#09090b", BgSurface: "#ffffff", BgCard: "#ffffff",
			TextMain: "#09090b", TextSecondary: "#27272a", TextMuted: "#52525b", TextOnAccent: "#ffffff", TextOnHeader: "#ffffff",
			AccentPrimary: "#e2e8f0", AccentPrimaryHover: "#ffffff", BorderMain: "#f3f4f6", BorderFocus: "#09090b",
			StatusNM: "#10b981", StatusLP: "#fbbf24", StatusMP: "#f97316", StatusHP: "#ef4444", StatusDMG: "#27272a",
			AccentHeader: "#e2e8f0", StatusHPHeader: "#cbd5e1",
			BtnPrimaryBg: "#09090b", BtnPrimaryText: "#ffffff", BtnSecondaryBg: "transparent", BtnSecondaryText: "#09090b",
			CheckboxBorder: "#27272a", CheckboxChecked: "#09090b",
			RadiusBase: "0px", PaddingCard: "16px", GapGrid: "20px",
			BgImageURL:  ptr("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23a1a1aa' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E"),
			FontHeading: ptr("Playfair Display, serif"), FontBody: ptr("Inter, sans-serif"),
			AccentSecondary: ptr("#52525b"),
		},
		{
			ID: "00000000-0000-0000-0000-000000000014", Name: "Strixhaven: Witherbloom", IsSystem: true,
			BgPage: "#1c1917", BgHeader: "#0d0401", BgSurface: "#292524", BgCard: "#292524",
			TextMain: "#d9f99d", TextSecondary: "#a3e635", TextMuted: "#84cc16", TextOnAccent: "#0d0401", TextOnHeader: "#d9f99d",
			AccentPrimary: "#84cc16", AccentPrimaryHover: "#65a30d", BorderMain: "#57534e", BorderFocus: "#84cc16",
			StatusNM: "#84cc16", StatusLP: "#fef08a", StatusMP: "#f97316", StatusHP: "#ef4444", StatusDMG: "#57534e",
			AccentHeader: "#bef264", StatusHPHeader: "#f87171",
			BtnPrimaryBg: "#84cc16", BtnPrimaryText: "#0d0401", BtnSecondaryBg: "#292524", BtnSecondaryText: "#84cc16",
			CheckboxBorder: "#57534e", CheckboxChecked: "#84cc16",
			RadiusBase: "24px", PaddingCard: "12px", GapGrid: "16px",
			BgImageURL:  ptr("data:image/svg+xml,%3Csvg width='20' height='20' viewBox='0 0 20 20' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M0 0h20v20H0V0zm10 17L3 9.5 9.5 3 17 9.5 10 17z' fill='%2365a30d' fill-opacity='0.05' fill-rule='evenodd'/%3E%3C/svg%3E"),
			FontHeading: ptr("Cinzel, serif"), FontBody: ptr("Inter, sans-serif"),
			AccentSecondary: ptr("#22c55e"),
		},
	}

	for _, t := range themes {
		_, err := db.NamedExec(`
			INSERT INTO theme (
				id, name, is_system, bg_page, bg_header, bg_surface, bg_card,
				text_main, text_secondary, text_muted, text_on_accent, text_on_header,
				accent_primary, accent_primary_hover, border_main, border_focus,
				status_nm, status_lp, status_mp, status_hp, status_dmg,
				btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
				checkbox_border, checkbox_checked,
				radius_base, padding_card, gap_grid,
				bg_image_url, font_heading, font_body, accent_secondary,
				accent_header, status_hp_header
			) VALUES (
				:id, :name, :is_system, :bg_page, :bg_header, :bg_surface, :bg_card,
				:text_main, :text_secondary, :text_muted, :text_on_accent, :text_on_header,
				:accent_primary, :accent_primary_hover, :border_main, :border_focus,
				:status_nm, :status_lp, :status_mp, :status_hp, :status_dmg,
				:btn_primary_bg, :btn_primary_text, :btn_secondary_bg, :btn_secondary_text,
				:checkbox_border, :checkbox_checked,
				:radius_base, :padding_card, :gap_grid,
				:bg_image_url, :font_heading, :font_body, :accent_secondary,
				:accent_header, :status_hp_header
			) ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				is_system = EXCLUDED.is_system,
				bg_page = EXCLUDED.bg_page,
				bg_header = EXCLUDED.bg_header,
				bg_surface = EXCLUDED.bg_surface,
				bg_card = EXCLUDED.bg_card,
				text_main = EXCLUDED.text_main,
				text_secondary = EXCLUDED.text_secondary,
				text_muted = EXCLUDED.text_muted,
				text_on_accent = EXCLUDED.text_on_accent,
				text_on_header = EXCLUDED.text_on_header,
				accent_primary = EXCLUDED.accent_primary,
				accent_primary_hover = EXCLUDED.accent_primary_hover,
				border_main = EXCLUDED.border_main,
				border_focus = EXCLUDED.border_focus,
				status_nm = EXCLUDED.status_nm,
				status_lp = EXCLUDED.status_lp,
				status_mp = EXCLUDED.status_mp,
				status_hp = EXCLUDED.status_hp,
				status_dmg = EXCLUDED.status_dmg,
				btn_primary_bg = EXCLUDED.btn_primary_bg,
				btn_primary_text = EXCLUDED.btn_primary_text,
				btn_secondary_bg = EXCLUDED.btn_secondary_bg,
				btn_secondary_text = EXCLUDED.btn_secondary_text,
				checkbox_border = EXCLUDED.checkbox_border,
				checkbox_checked = EXCLUDED.checkbox_checked,
				radius_base = EXCLUDED.radius_base,
				padding_card = EXCLUDED.padding_card,
				gap_grid = EXCLUDED.gap_grid,
				bg_image_url = EXCLUDED.bg_image_url,
				font_heading = EXCLUDED.font_heading,
				font_body = EXCLUDED.font_body,
				accent_secondary = EXCLUDED.accent_secondary,
				accent_header = EXCLUDED.accent_header,
				status_hp_header = EXCLUDED.status_hp_header
		`, t)
		if err != nil {
			logger.Error("Failed to seed theme '%s': %v", t.Name, err)
		}
	}
}

func seedTranslations(db *sqlx.DB) {
	logger.Info("🌐 Seeding Storefront Translations...")

	type Transl struct {
		Key    string
		Locale string
		Value  string
	}

	data := []Transl{
		// Common
		{"pages.common.buttons.add", "en", "ADD"},
		{"pages.common.buttons.add", "es", "AGREGAR"},
		{"pages.common.buttons.add_to_cart", "en", "ADD TO CART"},
		{"pages.common.buttons.add_to_cart", "es", "AGREGAR AL CARRITO"},
		{"pages.common.buttons.added_to_cart", "en", "ADDED TO CART"},
		{"pages.common.buttons.added_to_cart", "es", "AGREGADO"},
		{"pages.common.buttons.close", "en", "Close"},
		{"pages.common.buttons.close", "es", "Cerrar"},
		{"pages.common.status.sold_out", "en", "SOLD OUT"},
		{"pages.common.status.sold_out", "es", "AGOTADO"},
		{"pages.common.status.out_of_stock", "en", "OUT OF STOCK"},
		{"pages.common.status.out_of_stock", "es", "SIN STOCK"},
		{"pages.common.status.in_stock", "en", "IN STOCK"},
		{"pages.common.status.in_stock", "es", "EN STOCK"},
		{"pages.common.labels.art_by", "en", "Art by"},
		{"pages.common.labels.art_by", "es", "Arte por"},
		{"pages.common.labels.rarity", "en", "Rarity"},
		{"pages.common.labels.rarity", "es", "Rareza"},
		{"pages.common.labels.identity", "en", "Identity"},
		{"pages.common.labels.identity", "es", "Identidad"},
		{"pages.common.labels.art_var", "en", "Art Var."},
		{"pages.common.labels.art_var", "es", "Var. de Arte"},
		{"pages.common.labels.cmc", "en", "CMC"},
		{"pages.common.labels.cmc", "es", "CMC"},
		{"pages.common.labels.any_edition", "en", "Any Edition"},
		{"pages.common.labels.any_edition", "es", "Cualquier Edición"},
		{"pages.common.labels.normal", "en", "Normal"},
		{"pages.common.labels.normal", "es", "Normal"},
		{"pages.common.labels.any_condition", "en", "ANY"},
		{"pages.common.labels.any_condition", "es", "CUALQUIERA"},
		{"pages.common.labels.hidden", "en", "Hidden"},
		{"pages.common.labels.hidden", "es", "Oculto"},
		{"pages.common.labels.direct_demand", "en", "DIRECT DEMAND"},
		{"pages.common.labels.direct_demand", "es", "DEMANDA DIRECTA"},
		{"pages.common.labels.any", "en", "Any"},
		{"pages.common.labels.any", "es", "Cualquiera"},
		{"pages.common.labels.offer", "en", "Offer"},
		{"pages.common.labels.offer", "es", "Oferta"},
		{"pages.common.labels.ask", "en", "ASK"},
		{"pages.common.labels.ask", "es", "CONSULTAR"},
		{"pages.common.actions.sell", "en", "SELL"},
		{"pages.common.actions.sell", "es", "VENDER"},
		{"pages.common.tooltips.quantity_needed", "en", "Quantity needed"},
		{"pages.common.tooltips.quantity_needed", "es", "Cantidad necesaria"},
		{"pages.common.labels.id_number", "en", "ID NUMBER"},
		{"pages.common.labels.id_number", "es", "CÉDULA / ID"},
		{"pages.common.labels.address", "en", "ADDRESS"},
		{"pages.common.labels.address", "es", "DIRECCIÓN"},
		{"pages.common.currency.cop", "en", "COP"},
		{"pages.common.currency.cop", "es", "COP"},
		{"pages.common.labels.error", "en", "Error"},
		{"pages.common.labels.error", "es", "Error"},
		{"pages.common.actions.cancel", "en", "CANCEL"},
		{"pages.common.actions.cancel", "es", "CANCELAR"},
		{"pages.common.labels.id_number", "en", "ID NUMBER"},
		{"pages.common.labels.id_number", "es", "CÉDULA / ID"},
		{"pages.common.labels.address", "en", "ADDRESS"},
		{"pages.common.labels.address", "es", "DIRECCIÓN"},
		{"pages.common.forms.newsletter.title", "en", "Join the Pack"},
		{"pages.common.forms.newsletter.title", "es", "Únete a la manada"},
		{"pages.common.forms.newsletter.subtitle", "en", "Get restock alerts and shop news."},
		{"pages.common.forms.newsletter.subtitle", "es", "Alertas de stock y novedades."},
		{"pages.common.forms.newsletter.placeholder", "en", "YOUR EMAIL"},
		{"pages.common.forms.newsletter.placeholder", "es", "TU EMAIL"},
		{"pages.common.forms.newsletter.submit", "en", "OK"},
		{"pages.common.forms.newsletter.submit", "es", "OK"},
		{"pages.common.forms.newsletter.success", "en", "THX! CHECK YOUR INBOX SOON."},
		{"pages.common.forms.newsletter.success", "es", "¡GRACIAS! REVISA TU CORREO PRONTO."},
		{"pages.common.forms.newsletter.error", "en", "SOMETHING WENT WRONG."},
		{"pages.common.forms.newsletter.error", "es", "ALGO SALIÓ MAL."},

		// Product
		{"pages.product.details.not_found", "en", "ITEM NOT FOUND"},
		{"pages.product.details.not_found", "es", "ARTÍCULO NO ENCONTRADO"},
		{"pages.product.details.not_found_desc", "en", "This item may have been sold or removed."},
		{"pages.product.details.not_found_desc", "es", "Este artículo puede haber sido vendido o eliminado."},
		{"pages.product.details.view_full_page", "en", "View full page"},
		{"pages.product.details.view_full_page", "es", "Ver página completa"},
		{"pages.product.details.no_info", "en", "No additional information available."},
		{"pages.product.details.no_info", "es", "No hay información adicional disponible."},
		{"pages.product.details.textless", "en", "TEXTLESS"},
		{"pages.product.details.textless", "es", "SIN TEXTO"},
		{"pages.product.details.full_art", "en", "FULL ART"},
		{"pages.product.details.full_art", "es", "ARTE COMPLETO"},
		{"pages.product.status.in_stock", "en", "{count} IN STOCK"},
		{"pages.product.status.in_stock", "es", "{count} EN STOCK"},
		{"pages.product.cart_users_has", "en", "{count} OTHER USER HAS THIS IN THEIR CART"},
		{"pages.product.cart_users_has", "es", "{count} OTRO USUARIO TIENE ESTO EN SU CARRITO"},
		{"pages.product.cart_users_have", "en", "{count} OTHER USERS HAVE THIS IN THEIR CART"},
		{"pages.product.cart_users_have", "es", "{count} OTROS USUARIOS TIENEN ESTO EN SU CARRITO"},
		{"pages.product.status.available", "en", "available"},
		{"pages.product.status.available", "es", "dispon."},
		{"pages.product.details.legalities", "en", "FORMAT LEGALITY"},
		{"pages.product.details.legalities", "es", "LEGALIDAD EN FORMATOS"},
		{"pages.product.details.store_notice", "en", "Complete purchase in-store or verify availability at counter."},
		{"pages.product.details.store_notice", "es", "Completa la compra en tienda o verifica disponibilidad en el mostrador."},

		{"pages.common.status.active_bounty", "en", "Active Bounty"},
		{"pages.common.status.active_bounty", "es", "Bounty Activo"},
		{"pages.common.status.completed_past", "en", "Completed / Past"},
		{"pages.common.status.completed_past", "es", "Completado / Pasado"},
		{"pages.common.status.mission_complete", "en", "MISSION COMPLETE\nCARD DELIVERED"},
		{"pages.common.status.mission_complete", "es", "MISIÓN CUMPLIDA\nCARTA ENTREGADA"},

		// Checkout
		{"pages.checkout.page.title", "en", "CHECKOUT"},
		{"pages.checkout.page.title", "es", "FINALIZAR COMPRA"},
		{"pages.checkout.section.contact", "en", "CONTACT INFORMATION"},
		{"pages.checkout.section.contact", "es", "INFORMACIÓN DE CONTACTO"},
		{"pages.checkout.form.first_name", "en", "FIRST NAME"},
		{"pages.checkout.form.first_name", "es", "NOMBRE"},
		{"pages.checkout.form.last_name", "en", "LAST NAME"},
		{"pages.checkout.form.last_name", "es", "APELLIDO"},
		{"pages.checkout.form.phone", "en", "PHONE / WHATSAPP"},
		{"pages.checkout.form.phone", "es", "TELÉFONO / WHATSAPP"},
		{"pages.checkout.form.email", "en", "EMAIL"},
		{"pages.checkout.form.email", "es", "EMAIL"},
		{"pages.checkout.form.id_number", "en", "ID NUMBER / CÉDULA"},
		{"pages.checkout.form.id_number", "es", "CÉDULA / ID"},
		{"pages.checkout.form.address", "en", "ADDRESS"},
		{"pages.checkout.form.address", "es", "DIRECCIÓN"},
		{"pages.checkout.section.payment", "en", "PAYMENT METHOD"},
		{"pages.checkout.section.payment", "es", "MÉTODO DE PAGO"},
		{"pages.checkout.form.notes", "en", "NOTES (OPTIONAL)"},
		{"pages.checkout.form.notes", "es", "NOTAS (OPCIONAL)"},
		{"pages.checkout.form.notes_placeholder", "en", "Special instructions..."},
		{"pages.checkout.form.notes_placeholder", "es", "Instrucciones especiales..."},
		{"pages.checkout.section.summary", "en", "ORDER SUMMARY"},
		{"pages.checkout.section.summary", "es", "RESUMEN DEL PEDIDO"},
		{"pages.checkout.summary.empty_cart", "en", "Your cart is empty."},
		{"pages.checkout.summary.empty_cart", "es", "Tu carrito está vacío."},
		{"pages.checkout.summary.vaciar", "en", "CLEAR CART"},
		{"pages.checkout.summary.vaciar", "es", "VACIAR CARRITO"},
		{"pages.checkout.summary.item", "en", "ITEM"},
		{"pages.checkout.summary.item", "es", "ARTÍCULO"},
		{"pages.checkout.summary.items", "en", "ITEMS"},
		{"pages.checkout.summary.items", "es", "ARTÍCULOS"},
		{"pages.checkout.buttons.confirm", "en", "CONFIRM ORDER →"},
		{"pages.checkout.buttons.confirm", "es", "CONFIRMAR PEDIDO →"},
		{"pages.checkout.buttons.processing", "en", "PROCESSING..."},
		{"pages.checkout.buttons.processing", "es", "PROCESANDO..."},
		{"pages.checkout.footer.notice", "en", "Upon confirmation, an advisor will contact you to coordinate delivery."},
		{"pages.checkout.footer.notice", "es", "Al confirmar, un asesor se pondrá en contacto contigo para coordinar la entrega."},

		// Order Confirmation
		{"pages.order.confirmation.title", "en", "ORDER RECEIVED!"},
		{"pages.order.confirmation.title", "es", "¡ORDEN RECIBIDA!"},
		{"pages.order.confirmation.order_no", "en", "ORDER NUMBER"},
		{"pages.order.confirmation.order_no", "es", "NÚMERO DE ORDEN"},
		{"pages.order.confirmation.desc", "en", "Your order has been successfully registered. An advisor will contact you to coordinate payment and delivery."},
		{"pages.order.confirmation.desc", "es", "Tu pedido ha sido registrado exitosamente. Un asesor se pondrá en contacto contigo para coordinar el pago y la entrega."},
		{"pages.order.confirmation.back", "en", "← BACK TO STORE"},
		{"pages.order.confirmation.back", "es", "← VOLVER A LA TIENDA"},
		{"pages.order.confirmation.contact", "en", "CONTACT"},
		{"pages.order.confirmation.contact", "es", "CONTACTAR"},

		// Order Details Modal
		{"pages.order.modal.title", "en", "Order {number}"},
		{"pages.order.modal.title", "es", "Pedido {number}"},
		{"pages.order.modal.title_generic", "en", "Order Details"},
		{"pages.order.modal.title_generic", "es", "Detalles del Pedido"},
		{"pages.order.modal.fetching", "en", "Fetching secure details"},
		{"pages.order.modal.fetching", "es", "Obteniendo detalles seguros"},
		{"pages.order.modal.error", "en", "Failed to load order details. Please try again."},
		{"pages.order.modal.error", "es", "Error al cargar los detalles. Intenta de nuevo."},
		{"pages.order.modal.close", "en", "Close Window"},
		{"pages.order.modal.close", "es", "Cerrar Ventana"},
		{"pages.order.modal.status_label", "en", "Order Status"},
		{"pages.order.modal.status_label", "es", "Estado del Pedido"},
		{"pages.order.modal.date_label", "en", "Transaction Date"},
		{"pages.order.modal.date_label", "es", "Fecha de Transacción"},
		{"pages.order.modal.amount_label", "en", "Total Amount"},
		{"pages.order.modal.amount_label", "es", "Monto Total"},
		{"pages.order.modal.summary_label", "en", "Order Summary"},
		{"pages.order.modal.summary_label", "es", "Resumen del Pedido"},
		{"pages.order.modal.unknown_set", "en", "Unknown Set"},
		{"pages.order.modal.unknown_set", "es", "Set Desconocido"},
		{"pages.order.modal.qty_label", "en", "QTY:"},
		{"pages.order.modal.qty_label", "es", "CANT:"},
		{"pages.order.modal.reference", "en", "Order Reference:"},
		{"pages.order.modal.reference", "es", "Referencia de Pedido:"},
		{"pages.order.modal.method", "en", "Method:"},
		{"pages.order.modal.method", "es", "Método:"},
		{"pages.order.modal.issues", "en", "Issues? Contact"},
		{"pages.order.modal.issues", "es", "¿Problemas? Contacta a"},

		// Profile Page
		{"pages.profile.title", "en", "Your Account"},
		{"pages.profile.title", "es", "Tu Cuenta"},
		{"pages.profile.subtitle", "en", "Manage your profile and view your order history."},
		{"pages.profile.subtitle", "es", "Administra tu perfil y revisa tu historial de compras."},
		{"pages.profile.section.settings", "en", "Profile Settings"},
		{"pages.profile.section.settings", "es", "Configuración de Perfil"},
		{"pages.profile.form.full_name", "en", "Full Name"},
		{"pages.profile.form.full_name", "es", "Nombre Completo"},
		{"pages.profile.form.email", "en", "Email Address"},
		{"pages.profile.form.email", "es", "Correo Electrónico"},
		{"pages.profile.form.phone", "en", "Phone Number"},
		{"pages.profile.form.phone", "es", "Número de Teléfono"},
		{"pages.profile.form.id", "en", "ID Number / Cedula"},
		{"pages.profile.form.id", "es", "Documento de Identidad"},
		{"pages.profile.form.shipping", "en", "Shipping Address"},
		{"pages.profile.form.shipping", "es", "Dirección de Envío"},
		{"pages.profile.placeholders.phone", "en", "e.g. +57 300..."},
		{"pages.profile.placeholders.phone", "es", "ej. +57 300..."},
		{"pages.profile.placeholders.id", "en", "Required for shipping in Colombia"},
		{"pages.profile.placeholders.id", "es", "Requerido para envíos en Colombia"},
		{"pages.profile.placeholders.address", "en", "Full address, city, and instructions"},
		{"pages.profile.placeholders.address", "es", "Dirección completa, ciudad e instrucciones"},
		{"pages.profile.actions.save", "en", "Save Profile"},
		{"pages.profile.actions.save", "es", "Guardar Perfil"},
		{"pages.profile.actions.saving", "en", "Saving..."},
		{"pages.profile.actions.saving", "es", "Guardando..."},
		{"pages.profile.messages.success", "en", "Profile updated successfully!"},
		{"pages.profile.messages.success", "es", "¡Perfil actualizado con éxito!"},
		{"pages.profile.messages.error", "en", "Failed to update profile."},
		{"pages.profile.messages.error", "es", "Error al actualizar perfil."},
		{"pages.profile.section.linked", "en", "Linked Accounts"},
		{"pages.profile.section.linked", "es", "Cuentas Vinculadas"},
		{"pages.profile.status.connected", "en", "Connected"},
		{"pages.profile.status.connected", "es", "Conectado"},
		{"pages.profile.actions.link", "en", "Link"},
		{"pages.profile.actions.link", "es", "Vincular"},
		{"pages.profile.section.orders", "en", "Order History"},
		{"pages.profile.section.orders", "es", "Historial de Pedidos"},
		{"pages.profile.orders.empty", "en", "You haven't placed any orders yet."},
		{"pages.profile.orders.empty", "es", "Aún no has realizado pedidos."},
		{"pages.profile.orders.browse", "en", "Browse our inventory →"},
		{"pages.profile.orders.browse", "es", "Explora nuestro inventario →"},
		{"pages.profile.table.order_no", "en", "Order No."},
		{"pages.profile.table.order_no", "es", "No. Pedido"},
		{"pages.profile.table.date", "en", "Date"},
		{"pages.profile.table.date", "es", "Fecha"},
		{"pages.profile.table.status", "en", "Status"},
		{"pages.profile.table.status", "es", "Estado"},
		{"pages.profile.table.items", "en", "Items"},
		{"pages.profile.table.items", "es", "Artículos"},
		{"pages.profile.table.total", "en", "Total"},
		{"pages.profile.table.total", "es", "Total"},

		// Checkout Placeholders
		{"pages.checkout.placeholders.first_name", "en", "Juan"},
		{"pages.checkout.placeholders.first_name", "es", "Juan"},
		{"pages.checkout.placeholders.last_name", "en", "Perez"},
		{"pages.checkout.placeholders.last_name", "es", "Pérez"},
		{"pages.checkout.placeholders.phone", "en", "3001234567"},
		{"pages.checkout.placeholders.phone", "es", "3001234567"},
		{"pages.checkout.placeholders.email", "en", "john@example.com"},
		{"pages.checkout.placeholders.email", "es", "correo@ejemplo.com"},
		{"pages.checkout.placeholders.id", "en", "123456789"},
		{"pages.checkout.placeholders.id", "es", "123456789"},
		{"pages.checkout.placeholders.address", "en", "St # 1-2"},
		{"pages.checkout.placeholders.address", "es", "Cra 1 # 2-3"},

		// Login Page (Themed)
		{"pages.login.loading", "en", "Initializing Auth Matrix..."},
		{"pages.login.loading", "es", "Iniciando Matriz de Autenticación..."},
		{"pages.login.security_v", "en", "Security Protocol v2.5"},
		{"pages.login.security_v", "es", "Protocolo de Seguridad v2.5"},
		{"pages.login.title", "en", "WELCOME BACK"},
		{"pages.login.title", "es", "BIENVENIDO DE NUEVO"},
		{"pages.login.subtitle", "en", "Access the El Bulk Collective"},
		{"pages.login.subtitle", "es", "Accede al Colectivo El Bulk"},
		{"pages.login.google", "en", "SIGN IN WITH GOOGLE"},
		{"pages.login.google", "es", "INICIAR SESIÓN CON GOOGLE"},
		{"pages.login.facebook", "en", "SIGN IN WITH FACEBOOK"},
		{"pages.login.facebook", "es", "INICIAR SESIÓN CON FACEBOOK"},
		{"pages.login.footer.text", "en", "By entering the platform, you agree to our {terms} and {privacy}."},
		{"pages.login.footer.text", "es", "Al ingresar a la plataforma, aceptas nuestros {terms} y {privacy}."},
		{"pages.login.footer.terms", "en", "Terms of Engagement"},
		{"pages.login.footer.terms", "es", "Términos de Participación"},
		{"pages.login.footer.privacy", "en", "Data Protocols"},
		{"pages.login.footer.privacy", "es", "Protocolos de Datos"},
		{"pages.login.return", "en", "Return to Field Ops"},
		{"pages.login.return", "es", "Regresar a Operaciones"},
		{"pages.login.accent.store", "en", "EL BULK"},
		{"pages.login.accent.store", "es", "EL BULK"},
		{"pages.login.accent.logistics", "en", "LOGISTICS"},
		{"pages.login.accent.logistics", "es", "LOGÍSTICA"},

		// ProductGrid UI (Common Search UI)
		{"pages.inventory.grid.filters.title", "en", "FILTERS"},
		{"pages.inventory.grid.filters.title", "es", "FILTROS"},
		{"pages.inventory.grid.filters.strategy.title", "en", "Search Strategy"},
		{"pages.inventory.grid.filters.strategy.title", "es", "Estrategia de Búsqueda"},
		{"pages.inventory.grid.filters.strategy.broad", "en", "BROAD (OR)"},
		{"pages.inventory.grid.filters.strategy.broad", "es", "AMPLIA (O)"},
		{"pages.inventory.grid.filters.strategy.narrow", "en", "NARROW (AND)"},
		{"pages.inventory.grid.filters.strategy.narrow", "es", "ESTRICTA (Y)"},
		{"pages.inventory.grid.filters.strategy.broad_desc", "en", "Broadens results: match ANY selected filter."},
		{"pages.inventory.grid.filters.strategy.broad_desc", "es", "Amplía resultados: coincide con CUALQUIER filtro."},
		{"pages.inventory.grid.filters.strategy.narrow_desc", "en", "Narrows results: match ALL selected filters."},
		{"pages.inventory.grid.filters.strategy.narrow_desc", "es", "Reduce resultados: coincide con TODOS los filtros."},
		{"pages.inventory.grid.filters.keywords", "en", "Keywords"},
		{"pages.inventory.grid.filters.keywords", "es", "Palabras Clave"},
		{"pages.inventory.grid.filters.search_placeholder", "en", "Search cards..."},
		{"pages.inventory.grid.filters.search_placeholder", "es", "Buscar cartas..."},
		{"pages.inventory.grid.filters.availability", "en", "Availability"},
		{"pages.inventory.grid.filters.availability", "es", "Disponibilidad"},
		{"pages.inventory.grid.filters.in_stock", "en", "In Stock Only"},
		{"pages.inventory.grid.filters.in_stock", "es", "Solo en Stock"},
		{"pages.inventory.grid.filters.clear", "en", "Clear All Filters ×"},
		{"pages.inventory.grid.filters.clear", "es", "Limpiar Filtros ×"},
		{"pages.inventory.grid.sort.label", "en", "Sort"},
		{"pages.inventory.grid.sort.label", "es", "Ordenar"},
		{"pages.inventory.grid.sort.newest", "en", "Newest"},
		{"pages.inventory.grid.sort.newest", "es", "Más Nuevo"},
		{"pages.inventory.grid.sort.name", "en", "Name"},
		{"pages.inventory.grid.sort.name", "es", "Nombre"},
		{"pages.inventory.grid.sort.price", "en", "Price"},
		{"pages.inventory.grid.sort.price", "es", "Precio"},
		{"pages.inventory.grid.sort.cmc", "en", "Mana Cost"},
		{"pages.inventory.grid.sort.cmc", "es", "Costo de Maná"},
		{"pages.inventory.grid.sort.rarity", "en", "Rarity"},
		{"pages.inventory.grid.sort.rarity", "es", "Rareza"},
		{"pages.inventory.grid.status.result", "en", "result"},
		{"pages.inventory.grid.status.result", "es", "resultado"},
		{"pages.inventory.grid.status.results", "en", "results"},
		{"pages.inventory.grid.status.results", "es", "resultados"},
		{"pages.inventory.grid.status.no_results", "en", "NO RESULTS FOUND"},
		{"pages.inventory.grid.status.no_results", "es", "SIN RESULTADOS"},
		{"pages.inventory.grid.status.no_results_desc", "en", "Try clearing your filters or check back later."},
		{"pages.inventory.grid.status.no_results_desc", "es", "Intenta limpiar los filtros o vuelve más tarde."},
		{"pages.inventory.grid.pagination.prev", "en", "← Prev"},
		{"pages.inventory.grid.pagination.prev", "es", "← Ant"},
		{"pages.inventory.grid.pagination.next", "en", "Next →"},
		{"pages.inventory.grid.pagination.next", "es", "Siguiente →"},

		// Filter Section Headers
		{"grid.filters.condition", "en", "CONDITION"},
		{"grid.filters.condition", "es", "ESTADO"},
		{"grid.filters.finish", "en", "FINISH"},
		{"grid.filters.finish", "es", "TIPO DE FOIL"},
		{"grid.filters.version", "en", "VERSION"},
		{"grid.filters.version", "es", "VERSIÓN"},
		{"grid.filters.rarity", "en", "RARITY"},
		{"grid.filters.rarity", "es", "RAREZA"},
		{"grid.filters.set", "en", "SET"},
		{"grid.filters.set", "es", "EDICIÓN"},
		{"grid.filters.color", "en", "COLOR"},
		{"grid.filters.color", "es", "COLOR"},
		{"grid.filters.language", "en", "LANGUAGE"},
		{"grid.filters.language", "es", "IDIOMA"},
		{"grid.filters.collections", "en", "COLLECTIONS"},
		{"grid.filters.collections", "es", "COLECCIONES"},

		// Labels: FOIL
		{"pages.product.finish.non_foil", "en", "Non-Foil"},
		{"pages.product.finish.non_foil", "es", "No Foil"},
		{"pages.product.finish.foil", "en", "Foil"},
		{"pages.product.finish.foil", "es", "Foil"},
		{"pages.product.finish.etched_foil", "en", "Etched Foil"},
		{"pages.product.finish.etched_foil", "es", "Foil Etched"},

		// Labels: Treatment
		{"pages.product.version.normal", "en", "Regular"},
		{"pages.product.version.normal", "es", "Regular"},
		{"pages.product.version.borderless", "en", "Borderless"},
		{"pages.product.version.borderless", "es", "Sin Borde"},
		{"pages.product.version.showcase", "en", "Showcase"},
		{"pages.product.version.showcase", "es", "Showcase"},

		// TCG Labels
		{"tcg.mtg", "en", "Magic: The Gathering"},
		{"tcg.mtg", "es", "Magic: The Gathering"},
		{"tcg.pokemon", "en", "Pokémon"},
		{"tcg.pokemon", "es", "Pokémon"},

		// Singles Landing
		{"pages.singles.landing.category", "en", "CATEGORY // SINGLES"},
		{"pages.singles.landing.category", "es", "CATEGORÍA // SINGLES"},
		{"pages.singles.landing.title", "en", "INDIVIDUAL CARDS"},
		{"pages.singles.landing.title", "es", "CARTAS SUELTAS"},
		{"pages.singles.landing.desc", "en", "Browse our collection of hundreds of singles across your favorite TCGs. Pick your game to see the full inventory."},
		{"pages.singles.landing.desc", "es", "Explora nuestra colección de cientos de singles de tus TCGs favoritos. Elige tu juego para ver el inventario completo."},
		{"pages.singles.landing.view_all", "en", "VIEW ALL SINGLES →"},
		{"pages.singles.landing.view_all", "es", "VER TODAS LAS SINGLES →"},
		{"pages.singles.landing.featured", "en", "FEATURED SINGLES"},
		{"pages.singles.landing.featured", "es", "SINGLES DESTACADAS"},
		{"pages.singles.landing.no_featured", "en", "NO FEATURED SINGLES FOUND"},
		{"pages.singles.landing.no_featured", "es", "NO SE ENCONTRARON SINGLES DESTACADAS"},

		// Titles for ProductGrid
		{"pages.singles.title", "en", "{tcg} SINGLES"},
		{"pages.singles.title", "es", "SINGLES DE {tcg}"},
		{"pages.sealed.title", "en", "{tcg} SEALED PRODUCT"},
		{"pages.sealed.title", "es", "PRODUCTO SELLADO DE {tcg}"},
		{"pages.accessories.title", "en", "PROTECTION & ACCESSORIES"},
		{"pages.accessories.title", "es", "ACCESORIOS Y PROTECCIÓN"},
		{"pages.store_exclusives.title", "en", "STORE EXCLUSIVES"},
		{"pages.store_exclusives.title", "es", "EXCLUSIVOS DE LA TIENDA"},
		{"pages.checkout.section.removed", "en", "REMOVED PRODUCTS"},
		{"pages.checkout.section.removed", "es", "PRODUCTOS ELIMINADOS"},
		{"pages.checkout.buttons.re_add", "en", "RE-ADD"},
		{"pages.checkout.buttons.re_add", "es", "REAGREGAR"},
		{"pages.checkout.error.empty", "en", "Your cart is empty. Please add products before continuing."},
		{"pages.checkout.error.empty", "es", "Tu carrito está vacío. Agrega productos antes de continuar."},
		{"pages.checkout.buttons.back", "en", "← BACK TO SHOP"},
		{"pages.checkout.buttons.back", "es", "← VOLVER A LA TIENDA"},

		{"pages.checkout.buttons.back", "en", "← BACK TO SHOP"},
		{"pages.checkout.buttons.back", "es", "← VOLVER A LA TIENDA"},

		// Bounties
		{"pages.bounties.page.title", "en", "WANTED / BOUNTIES"},
		{"pages.bounties.page.title", "es", "BUSCAMOS / BOUNTIES"},
		{"pages.bounties.page.subtitle", "en", "We are actively looking to buy the cards below. If you have them, reach out to us! Can't find what you are looking for? Send us a card request!"},
		{"pages.bounties.page.subtitle", "es", "Estamos buscando comprar las siguientes cartas. ¡Si las tienes, contáctanos! ¿No encuentras lo que buscas? ¡Envíanos una solicitud!"},
		{"pages.bounties.status.showing", "en", "SHOWING {count} ENTRIES"},
		{"pages.bounties.status.showing", "es", "MOSTRANDO {count} ENTRADAS"},
		{"pages.bounties.buttons.request", "en", "REQUEST A CARD"},
		{"pages.bounties.buttons.request", "es", "SOLICITAR CARTA"},
		{"pages.bounties.success.request", "en", "Your request was submitted successfully! We will contact you if we locate the card."},
		{"pages.bounties.success.request", "es", "¡Tu solicitud fue enviada con éxito! Te contactaremos si localizamos la carta."},
		{"pages.bounties.empty.title", "en", "NO ACTIVE BOUNTIES"},
		{"pages.bounties.empty.title", "es", "SIN BOUNTIES ACTIVOS"},
		{"pages.bounties.empty.desc", "en", "We are not currently explicitly searching for any specific cards. But you can still send us a request!"},
		{"pages.bounties.empty.desc", "es", "No estamos buscando ninguna carta específica en este momento. ¡Pero aún puedes enviarnos una solicitud!"},
		{"pages.bounties.buttons.submit_request", "en", "SUBMIT REQUEST"},
		{"pages.bounties.buttons.submit_request", "es", "ENVIAR SOLICITUD"},

		// Singles & Sealed Pages
		{"pages.singles.title", "en", "{tcg} SINGLES"},
		{"pages.singles.title", "es", "SINGLES DE {tcg}"},
		{"pages.singles.subtitle", "en", "Browse individual {tcg} cards by condition, treatment, and foil finish."},
		{"pages.singles.subtitle", "es", "Explora cartas individuales de {tcg} por estado, tratamiento y acabado foil."},
		{"pages.sealed.title", "en", "{tcg} SEALED"},
		{"pages.sealed.title", "es", "PRODUCTO SELLADO DE {tcg}"},
		{"pages.sealed.subtitle", "en", "Booster boxes, bundles, and sealed product for {tcg}."},
		{"pages.sealed.subtitle", "es", "Cajas de sobres, bundles y producto sellado de {tcg}."},

		{"pages.sealed.subtitle", "en", "Booster boxes, bundles, and sealed product for {tcg}."},
		{"pages.sealed.subtitle", "es", "Cajas de sobres, bundles y producto sellado de {tcg}."},

		// Contact Page
		{"pages.contact.page.title", "en", "Contact Us"},
		{"pages.contact.page.title", "es", "Contáctanos"},
		{"pages.contact.section.visit", "en", "VISIT THE BOX"},
		{"pages.contact.section.visit", "es", "VISÍTANOS"},
		{"pages.contact.info.address", "en", "ADDRESS"},
		{"pages.contact.info.address", "es", "DIRECCIÓN"},
		{"pages.contact.info.hours", "en", "HOURS"},
		{"pages.contact.info.hours", "es", "HORARIOS"},
		{"pages.contact.info.quote", "en", "\"Just look for the stack of shoeboxes near the back entrance.\""},
		{"pages.contact.info.quote", "es", "\"Busca la pila de cajas de zapatos cerca de la entrada trasera.\""},
		{"pages.contact.section.digital", "en", "DIGITAL COMMS"},
		{"pages.contact.section.digital", "es", "MEDIOS DIGITALES"},
		{"pages.contact.info.whatsapp", "en", "WHATSAPP / SALES"},
		{"pages.contact.info.whatsapp", "es", "WHATSAPP / VENTAS"},
		{"pages.contact.buttons.sell_bulk", "en", "SELL US YOUR BULK →"},
		{"pages.contact.buttons.sell_bulk", "es", "VÉNDENOS TU BULK →"},
		{"pages.contact.map.placeholder", "en", "MAP INTEGRATION"},
		{"pages.contact.map.placeholder", "es", "INTEGRACIÓN DE MAPA"},

		{"pages.contact.map.placeholder", "en", "MAP INTEGRATION"},
		{"pages.contact.map.placeholder", "es", "INTEGRACIÓN DE MAPA"},

		// Home Page
		{"pages.home.status.empty", "en", "STORE IS EMPTY"},
		{"pages.home.status.empty", "es", "LA TIENDA ESTÁ VACÍA"},
		{"pages.home.status.empty_desc", "en", "No collections have been populated yet."},
		{"pages.home.status.empty_desc", "es", "No se han publicado colecciones todavía."},
		{"pages.home.buttons.view_all", "en", "VIEW ALL →"},
		{"pages.home.buttons.view_all", "es", "VER TODO →"},
		{"pages.home.status.no_items", "en", "No items assigned to this collection yet."},
		{"pages.home.status.no_items", "es", "No hay artículos asignados a esta colección todavía."},
		{"pages.home.tags.urgently_needed", "en", "Urgently Needed"},
		{"pages.home.tags.urgently_needed", "es", "Urgente"},

		// Nav
		{"pages.nav.main.singles", "en", "Singles"},
		{"pages.nav.main.singles", "es", "Singles"},
		{"pages.nav.main.sealed", "en", "Sealed"},
		{"pages.nav.main.sealed", "es", "Sellado"},
		{"pages.nav.main.accessories", "en", "Accessories"},
		{"pages.nav.main.accessories", "es", "Accesorios"},
		{"pages.nav.main.store_exclusives", "en", "Store Exclusives"},
		{"pages.nav.main.store_exclusives", "es", "Exclusivos"},
		{"pages.nav.main.notices", "en", "Notices"},
		{"pages.nav.main.notices", "es", "Noticias"},
		{"pages.nav.main.contact", "en", "Contact"},
		{"pages.nav.main.contact", "es", "Contacto"},
		{"pages.nav.main.wanted", "en", "Wanted Cards"},
		{"pages.nav.main.wanted", "es", "Buscamos"},
		{"pages.nav.main.bulk", "en", "Sell Your Bulk"},
		{"pages.nav.main.bulk", "es", "Vende tu Bulk"},
		{"pages.nav.main.tcg_store", "en", "TCG STORE"},
		{"pages.nav.main.tcg_store", "es", "TIENDA TCG"},
		{"pages.nav.mobile.inventory_title", "en", "Singles Inventory"},
		{"pages.nav.mobile.inventory_title", "es", "Inventario de Singles"},

		// Cookies
		{"pages.cookies.title", "en", "COOKIE CONSENT"},
		{"pages.cookies.title", "es", "CONSENTIMIENTO DE COOKIES"},
		{"pages.cookies.description", "en", "We use cookies to improve your experience. You can choose to accept all or customize your preferences."},
		{"pages.cookies.description", "es", "Usamos cookies para mejorar tu experiencia. Puedes elegir aceptar todas o personalizar tus preferencias."},
		{"pages.cookies.customize", "en", "CUSTOMIZE"},
		{"pages.cookies.customize", "es", "PERSONALIZAR"},
		{"pages.order.status.pending", "en", "Pending"},
		{"pages.order.status.pending", "es", "Pendiente"},
		{"pages.order.status.confirmed", "en", "Confirmed"},
		{"pages.order.status.confirmed", "es", "Confirmado"},
		{"pages.order.status.completed", "en", "Completed"},
		{"pages.order.status.completed", "es", "Completado"},
		{"pages.order.status.cancelled", "en", "Cancelled"},
		{"pages.order.status.cancelled", "es", "Cancelado"},
		{"pages.order.status.shipped", "en", "Shipped"},
		{"pages.order.status.shipped", "es", "Enviado"},
		{"pages.order.status.ready_for_pickup", "en", "Ready for Pickup"},
		{"pages.order.status.ready_for_pickup", "es", "Listo para Recoger"},

		{"pages.cookies.always_required", "en", "Always Required"},
		{"pages.cookies.always_required", "es", "Siempre Requerido"},
		{"pages.cookies.preferences", "en", "COOKIE PREFERENCES"},
		{"pages.cookies.preferences", "es", "PREFERENCIAS de COOKIES"},
		{"pages.cookies.essential_title", "en", "NECESSARY"},
		{"pages.cookies.essential_title", "es", "NECESARIAS"},
		{"pages.cookies.essential_desc", "en", "Essential for the site to function (Auth, Theme, Cart)."},
		{"pages.cookies.essential_desc", "es", "Esenciales para que el sitio funcione (Autenticación, Tema, Carrito)."},
		{"pages.cookies.analytics_title", "en", "ANALYTICS"},
		{"pages.cookies.analytics_title", "es", "ANALÍTICA"},
		{"pages.cookies.analytics_desc", "en", "Help us understand how people use the store to improve user experience."},
		{"pages.cookies.analytics_desc", "es", "Nos ayudan a entender cómo la gente usa la tienda para mejorar la experiencia."},
		{"pages.cookies.marketing_title", "en", "MARKETING"},
		{"pages.cookies.marketing_title", "es", "MARKETING"},
		{"pages.cookies.marketing_desc", "en", "Used for personalized advertising and social features."},
		{"pages.cookies.marketing_desc", "es", "Se usan para publicidad personalizada y funciones sociales."},
		{"pages.cookies.save", "en", "SAVE MY CHOICES"},
		{"pages.cookies.save", "es", "GUARDAR MIS PREFERENCIAS"},
		{"pages.common.back", "en", "BACK"},
		{"pages.common.back", "es", "ATRÁS"},
		{"pages.nav.mobile.sealed_title", "en", "Sealed Product"},
		{"pages.nav.mobile.sealed_title", "es", "Producto Sellado"},
		{"pages.nav.tooltips.foil_enable", "en", "Enable foil effects"},
		{"pages.nav.tooltips.foil_enable", "es", "Activar efectos foil"},
		{"pages.nav.tooltips.foil_disable", "en", "Disable foil effects"},
		{"pages.nav.tooltips.foil_disable", "es", "Desactivar efectos foil"},
		{"pages.nav.tooltips.change_lang", "en", "Change language"},
		{"pages.nav.tooltips.change_lang", "es", "Cambiar idioma"},
		{"pages.nav.user.login", "en", "Login"},
		{"pages.nav.user.login", "es", "Ingresar"},
		{"pages.nav.user.profile", "en", "My Profile"},
		{"pages.nav.user.profile", "es", "Mi Perfil"},
		{"pages.nav.user.logout", "en", "Logout"},
		{"pages.nav.user.logout", "es", "Cerrar Sesión"},

		// Cart Drawer
		{"pages.cart.drawer.title", "en", "YOUR CART"},
		{"pages.cart.drawer.title", "es", "TU CARRITO"},
		{"pages.cart.drawer.empty_msg", "en", "Your cart is empty. Go find some cards."},
		{"pages.cart.drawer.empty_msg", "es", "Tu carrito está vacío. ¡Ve por unas cartas!"},
		{"pages.cart.drawer.clear", "en", "CLEAR"},
		{"pages.cart.drawer.clear", "es", "LIMPIAR"},
		{"pages.cart.drawer.confirm_clear", "en", "Are you sure you want to clear your cart?"},
		{"pages.cart.drawer.confirm_clear", "es", "¿Estás seguro de que deseas vaciar tu carrito?"},
		{"pages.cart.drawer.remove_tooltip", "en", "Remove"},
		{"pages.cart.drawer.remove_tooltip", "es", "Eliminar"},
		{"pages.cart.drawer.delete_perm_tooltip", "en", "Permanently delete"},
		{"pages.cart.drawer.delete_perm_tooltip", "es", "Borrar permanentemente"},
		{"pages.cart.drawer.total", "en", "TOTAL"},
		{"pages.cart.drawer.total", "es", "TOTAL"},
		{"pages.cart.drawer.checkout_btn", "en", "PROCEED TO CHECKOUT →"},
		{"pages.cart.drawer.checkout_btn", "es", "PROCEDER AL PAGO →"},
		{"pages.cart.drawer.checkout_notice", "en", "Review your order and complete shipping details."},
		{"pages.cart.drawer.checkout_notice", "es", "Revisa tu orden y completa los datos de envío."},
		{"pages.cart.drawer.removed_items", "en", "Removed Items"},
		{"pages.cart.drawer.removed_items", "es", "Artículos Eliminados"},
		{"pages.cart.drawer.restore", "en", "Restore"},
		{"pages.cart.drawer.restore", "es", "Restaurar"},

		// Home
		{"pages.home.hero.subtitle", "en", "YOUR LOCAL TCG SHOP"},
		{"pages.home.hero.subtitle", "es", "TU TIENDA TCG LOCAL"},
		{"pages.home.hero.title_refined", "en", "EL BULK / THE CARDS THEY OVERLOOKED."},
		{"pages.home.hero.title_refined", "es", "EL BULK / LAS CARTAS QUE ELLOS IGNORAN."},
		{"pages.home.hero.description", "en", "The shoebox where we keep all the good stuff. Singles, sealed product, and accessories."},
		{"pages.home.hero.description", "es", "La caja de zapatos donde guardamos lo bueno. Singles, sellado y accesorios."},
		{"pages.home.hero.description_refined", "en", "The essential pieces others treat as trash. We curate the common, the uncommon, and the impossible-to-find cards that complete your strategy."},
		{"pages.home.hero.description_refined", "es", "Las piezas esenciales que otros tratan como basura. Curamos las comunes, infrecuentes y cartas imposibles de encontrar que completan tu estrategia."},
		{"pages.home.hero.search_placeholder", "en", "Find the cards everyone else missed..."},
		{"pages.home.hero.search_placeholder", "es", "Encuentra las cartas que todos los demás pasaron por alto..."},
		{"pages.home.hero.browse_singles", "en", "BROWSE SINGLES"},
		{"pages.home.hero.browse_singles", "es", "VER SINGLES"},
		{"pages.home.hero.sell_bulk", "en", "SELL YOUR BULK →"},
		{"pages.home.hero.sell_bulk", "es", "VENDE TU BULK →"},
		{"pages.home.sections.view_all", "en", "VIEW ALL →"},
		{"pages.home.sections.view_all", "es", "VER TODO →"},
		{"pages.home.sections.empty", "en", "STORE IS EMPTY"},
		{"pages.home.sections.empty", "es", "TIENDA VACÍA"},
		{"pages.home.sections.got_bulk", "en", "GOT BULK?"},
		{"pages.home.sections.got_bulk", "es", "¿TIENES BULK?"},
		{"pages.home.sections.bulk_cta_text", "en", "We buy bulk commons and uncommons, bulk rares, and junk rare lots. Box it up and bring it in, get cash. No appointment needed."},
		{"pages.home.sections.bulk_cta_text", "es", "Compramos tus comunes, infrecuentes y raras. Empácalas y tráelas por efectivo. Sin cita previa."},
		{"pages.home.sections.see_bulk_prices", "en", "SEE BULK PRICES"},
		{"pages.home.sections.see_bulk_prices", "es", "VER PRECIOS DE BULK"},

		{"pages.layout.footer.slogan", "en", "We buy bulk. We sell singles. We love cardboard."},
		{"pages.layout.footer.slogan", "es", "Compramos bulk. Vendemos singles. Amamos el cartón."},
		{"pages.home.bounties.urgently_needed", "en", "Urgently Needed"},
		{"pages.home.bounties.urgently_needed", "es", "Urgente"},

		// Bulk Page
		{"pages.bulk.page.title", "en", "BRING YOUR BULK"},
		{"pages.bulk.page.title", "es", "TRAE TU BULK"},
		{"pages.bulk.page.subtitle", "en", "Got a shoebox of old cards gathering dust? We'll take them off your hands and put cash in your pocket. No appointment needed — just walk in."},
		{"pages.bulk.page.subtitle", "es", "¿Tienes una caja de cartas viejas llenándose de polvo? Nosotros las recibimos y te damos efectivo. Sin cita — solo ven."},
		{"pages.bulk.intro.we_buy", "en", "WE BUY"},
		{"pages.bulk.intro.we_buy", "es", "COMPRAMOS"},
		{"pages.bulk.tiers.common.label", "en", "BULK COMMONS & UNCOMMONS"},
		{"pages.bulk.tiers.common.label", "es", "BULK COMUNES E INFRECUENTES"},
		{"pages.bulk.tiers.common.price", "en", "$20.000 COP per 1,000"},
		{"pages.bulk.tiers.common.price", "es", "$20.000 COP por 1.000"},
		{"pages.bulk.tiers.common.desc", "en", "Any condition, any set. Sorted or unsorted. We take it all."},
		{"pages.bulk.tiers.common.desc", "es", "Cualquier estado, cualquier set. Clasificadas o no. Lo recibimos todo."},
		{"pages.bulk.tiers.rare.label", "en", "BULK RARES & MYTHICS"},
		{"pages.bulk.tiers.rare.label", "es", "BULK RARAS Y MÍTICAS"},
		{"pages.bulk.tiers.rare.price", "en", "$1.000 COP per card"},
		{"pages.bulk.tiers.rare.price", "es", "$1.000 COP por carta"},
		{"pages.bulk.tiers.rare.desc", "en", "NM/LP only. Bulk rares from Standard and below."},
		{"pages.bulk.tiers.rare.desc", "es", "Solo NM/LP. Raras bulk de Standard y anteriores."},
		{"pages.bulk.tiers.junk.label", "en", "JUNK RARE LOTS"},
		{"pages.bulk.tiers.junk.label", "es", "LOTES DE RARAS JUNK"},
		{"pages.bulk.tiers.junk.price", "en", "$12.000 COP per 100"},
		{"pages.bulk.tiers.junk.price", "es", "$12.000 COP por 100"},
		{"pages.bulk.tiers.junk.desc", "en", "MP-DMG rares and mythics, or commons/uncommons in poor condition."},
		{"pages.bulk.tiers.junk.desc", "es", "Raras y míticas MP-DMG, o comunes/infrecuentes en mal estado."},
		{"pages.bulk.tiers.foil.label", "en", "FOIL COMMONS & UNCOMMONS"},
		{"pages.bulk.tiers.foil.label", "es", "BULK COMUNES E INFRECUENTES FOIL"},
		{"pages.bulk.tiers.foil.price", "en", "$40.000 COP per 500"},
		{"pages.bulk.tiers.foil.price", "es", "$40.000 COP por 500"},
		{"pages.bulk.tiers.foil.desc", "en", "Any condition. Foil bulk sorted separately."},
		{"pages.bulk.tiers.foil.desc", "es", "Cualquier estado. El bulk foil se clasifica por separado."},
		{"pages.bulk.accepts.0", "en", "Magic: The Gathering (all sets, all formats)"},
		{"pages.bulk.accepts.0", "es", "Magic: The Gathering (todos los sets, todos los formatos)"},
		{"pages.bulk.accepts.1", "en", "Pokémon TCG (English only)"},
		{"pages.bulk.accepts.1", "es", "Pokémon TCG (Solo inglés)"},
		{"pages.bulk.accepts.2", "en", "Disney Lorcana"},
		{"pages.bulk.accepts.2", "es", "Disney Lorcana"},
		{"pages.bulk.accepts.3", "en", "One Piece TCG"},
		{"pages.bulk.accepts.3", "es", "One Piece TCG"},
		{"pages.bulk.accepts.4", "en", "Basic lands (we pay $4.000 COP per 200)"},
		{"pages.bulk.accepts.4", "es", "Tierras básicas (pagamos $4.000 COP por 200)"},
		{"pages.bulk.accepts.5", "en", "Tokens and emblems ($0 value but we take them)"},
		{"pages.bulk.accepts.5", "es", "Tokens y emblemas (valor $0 pero los recibimos)"},
		{"pages.bulk.sections.prices", "en", "CURRENT BULK PRICES"},
		{"pages.bulk.sections.prices", "es", "PRECIOS ACTUALES"},
		{"pages.bulk.sections.accept", "en", "WE GLADLY ACCEPT"},
		{"pages.bulk.sections.accept", "es", "ACEPTAMOS"},
		{"pages.bulk.sections.how_it_works", "en", "HOW IT WORKS"},
		{"pages.bulk.sections.how_it_works", "es", "¿CÓMO FUNCIONA?"},
		{"pages.bulk.cta.questions", "en", "HAVE QUESTIONS?"},
		{"pages.bulk.cta.questions", "es", "¿TIENES DUDAS?"},
		{"pages.bulk.cta.email_btn", "en", "EMAIL US ABOUT YOUR BULK"},
		{"pages.bulk.cta.email_btn", "es", "ESCRÍBENOS POR TU BULK"},
		{"pages.bulk.sections.price_disclaimer", "en", "⚠ Prices updated regularly. Large lots (1,000+ cards) may receive bonus offers. Prices are in-store cash — store credit offers up to 25% more."},
		{"pages.bulk.sections.price_disclaimer", "es", "⚠ Precios actualizados regularmente. Lotes grandes (más de 1,000 cartas) pueden recibir bonificaciones. Los precios son en efectivo — las ofertas de crédito de la tienda son hasta un 25% más."},
		{"pages.bulk.how.1.title", "en", "BRING YOUR CARDS"},
		{"pages.bulk.how.1.title", "es", "TRAE TUS CARTAS"},
		{"pages.bulk.how.1.desc", "en", "Walk in with your bulk. Sorted or unsorted, boxed or bagged — doesn't matter."},
		{"pages.bulk.how.1.desc", "es", "Ven con tu bulk. Ordenado o sin ordenar, en cajas o bolsas — no importa."},
		{"pages.bulk.how.2.title", "en", "WE COUNT & GRADE"},
		{"pages.bulk.how.2.title", "es", "CONTAMOS Y CLASIFICAMOS"},
		{"pages.bulk.how.2.desc", "en", "We do a quick count of your cards. For large lots we may ask for a day to sort."},
		{"pages.bulk.how.2.desc", "es", "Hacemos un conteo rápido de tus cartas. Para lotes grandes, podemos pedir un día para clasificar."},
		{"pages.bulk.how.3.title", "en", "GET AN OFFER"},
		{"pages.bulk.how.3.title", "es", "RECIBE UNA OFERTA"},
		{"pages.bulk.how.3.desc", "en", "We give you a cash offer on the spot. No pressure to accept."},
		{"pages.bulk.how.3.desc", "es", "Te damos una oferta en efectivo al instante. Sin presión para aceptar."},
		{"pages.bulk.how.4.title", "en", "TAKE THE CASH"},
		{"pages.bulk.how.4.title", "es", "LLÉVATE EL EFECTIVO"},
		{"pages.bulk.how.4.desc", "en", "Accept and walk out with cash (or store credit for more). Simple."},
		{"pages.bulk.how.4.desc", "es", "Acepta y retírate con efectivo (o más crédito de la tienda). Simple."},

		// Notices Section
		{"pages.sealed.landing.category", "en", "CATEGORY // SEALED"},
		{"pages.sealed.landing.category", "es", "CATEGORÍA // SELLADO"},
		{"pages.sealed.landing.title.main", "en", "BOXES &"},
		{"pages.sealed.landing.title.main", "es", "CAJAS Y"},
		{"pages.sealed.landing.title.accent", "en", "PACKS"},
		{"pages.sealed.landing.title.accent", "es", "SOBRES"},
		{"pages.sealed.landing.desc", "en", "From booster boxes to collector packs, explore every set in our vault. Choose your game to see the current inventory."},
		{"pages.sealed.landing.desc", "es", "Desde booster boxes hasta sobres de coleccionista, explora cada set en nuestra bóveda. Elige tu juego para ver el inventario actual."},
		{"pages.sealed.landing.view_all", "en", "VIEW ALL SEALED →"},
		{"pages.sealed.landing.view_all", "es", "VER TODO SELLADO →"},
		{"pages.sealed.landing.no_featured", "en", "NO FEATURED SEALED FOUND"},
		{"pages.sealed.landing.no_featured", "es", "NO SE ENCONTRARON PRODUCTOS SELLADOS DESTACADOS"},
		{"pages.sealed.landing.featured.main", "en", "FEATURED"},
		{"pages.sealed.landing.featured.main", "es", "SELLADO"},
		{"pages.sealed.landing.featured.accent", "en", "SEALED"},
		{"pages.sealed.landing.featured.accent", "es", "DESTACADO"},

		{"pages.accessories.title", "en", "ACCESSORIES"},
		{"pages.accessories.title", "es", "ACCESORIOS"},
		{"pages.accessories.subtitle", "en", "Sleeves, binders, deck boxes, playmats and more — for all TCGs."},
		{"pages.accessories.subtitle", "es", "Protectores, carpetas, deck boxes, playmats y más — para todos los TCGs."},

		{"components.client_request_modal.title", "en", "Request a card"},
		{"components.client_request_modal.title", "es", "Solicitar una carta"},
		{"components.client_request_modal.desc", "en", "Can't find what you need? Tell us the details and we'll start the hunt!"},
		{"components.client_request_modal.desc", "es", "¿No encuentras lo que buscas? ¡Danos los detalles y empezaremos la búsqueda!"},
		{"components.client_request_modal.form.name_label", "en", "Your Name *"},
		{"components.client_request_modal.form.name_label", "es", "Tu Nombre *"},
		{"components.client_request_modal.form.name_placeholder", "en", "John Doe"},
		{"components.client_request_modal.form.name_placeholder", "es", "John Doe"},
		{"components.client_request_modal.form.contact_label", "en", "Contact Info *"},
		{"components.client_request_modal.form.contact_label", "es", "Información de Contacto *"},
		{"components.client_request_modal.form.contact_placeholder", "en", "Phone or Instagram"},
		{"components.language_selector.names", "en", "🇺🇸 English"},
		{"components.language_selector.names", "es", "🇪🇸 Español"},
		{"components.client_request_modal.form.contact_placeholder", "es", "Teléfono o Instagram"},
		{"components.client_request_modal.form.card_label", "en", "Card Name *"},
		{"components.client_request_modal.form.card_label", "es", "Nombre de la Carta *"},
		{"components.client_request_modal.form.card_placeholder", "en", "e.g. Sheoldred, the Apocalypse"},
		{"components.client_request_modal.form.card_placeholder", "es", "ej. Sheoldred, the Apocalypse"},
		{"components.client_request_modal.form.set_label", "en", "Specific Set (Optional)"},
		{"components.client_request_modal.form.set_label", "es", "Set Específico (Opcional)"},
		{"components.client_request_modal.form.set_placeholder", "en", "e.g. Dominaria United"},
		{"components.client_request_modal.form.set_placeholder", "es", "ej. Dominaria United"},
		{"components.client_request_modal.form.details_label", "en", "Additional Details"},
		{"components.client_request_modal.form.details_label", "es", "Detalles Adicionales"},
		{"components.client_request_modal.form.details_placeholder", "en", "Condition, foil, language, etc..."},
		{"components.client_request_modal.form.details_placeholder", "es", "Estado, foil, idioma, etc..."},
		{"components.client_request_modal.form.submit_btn", "en", "SUBMIT MISSION"},
		{"components.client_request_modal.form.submit_btn", "es", "ENVIAR MISIÓN"},
		{"components.client_request_modal.login_prompt", "en", "to automatically fill your info and track your request status."},
		{"components.client_request_modal.login_prompt", "es", "para autocompletar tu info y seguir el estado de tu solicitud."},

		{"components.deck_contents.title", "en", "Deck Contents"},
		{"components.deck_contents.title", "es", "Contenido del mazo"},
		{"components.deck_contents.expand_all", "en", "Expand All"},
		{"components.deck_contents.expand_all", "es", "Expandir todo"},
		{"components.deck_contents.collapse_all", "en", "Collapse All"},
		{"components.deck_contents.collapse_all", "es", "Colapsar todo"},

		// Notices Section
		{"pages.notices.list.header_main", "en", "NOTICES &"},
		{"pages.notices.list.header_main", "es", "AVISOS Y"},
		{"pages.notices.list.header_accent", "en", "UPDATES"},
		{"pages.notices.list.header_accent", "es", "NOVEDADES"},
		{"pages.notices.list.subtitle", "en", "Latest news, spoilers, and reviews from the shop."},
		{"pages.notices.list.subtitle", "es", "Últimas noticias, spoilers y reseñas de la tienda."},
		{"pages.notices.list.total_count", "en", "{count} POSTS TOTAL"},
		{"pages.notices.list.total_count", "es", "{count} PUBLICACIONES EN TOTAL"},
		{"pages.notices.list.empty_title", "en", "NO NOTICES POSTED YET"},
		{"pages.notices.list.empty_title", "es", "AÚN NO HAY AVISOS PUBLICADOS"},
		{"pages.notices.list.empty_subtitle", "en", "Check back soon for news and updates."},
		{"pages.notices.list.empty_subtitle", "es", "Vuelve pronto para ver noticias y actualizaciones."},
		{"pages.notices.list.open_btn", "en", "OPEN NOTICE →"},
		{"pages.notices.list.open_btn", "es", "ABRIR AVISO →"},

		{"pages.notices.detail.back_btn", "en", "← BACK TO ALL NOTICES"},
		{"pages.notices.detail.back_btn", "es", "← VOLVER A TODOS LOS AVISOS"},
		{"pages.notices.detail.published_on", "en", "Published on {date}"},
		{"pages.notices.detail.published_on", "es", "Publicado el {date}"},
		{"pages.notices.detail.all_notices_btn", "en", "← ALL NOTICES"},
		{"pages.notices.detail.all_notices_btn", "es", "← TODOS LOS AVISOS"},
		{"pages.notices.detail.share_label", "en", "SHARE:"},
		{"pages.notices.detail.share_label", "es", "COMPARTIR:"},
		{"pages.notices.detail.sidebar.title", "en", "EL BULK SHOP"},
		{"pages.notices.detail.sidebar.title", "es", "TIENDA EL BULK"},
		{"pages.notices.detail.sidebar.desc", "en", "The shoebox where we keep all the good stuff. Selling singles, sealed product, and accessories. Visit us in person or shop online!"},
		{"pages.notices.detail.sidebar.desc", "es", "La caja de zapatos donde guardamos lo bueno. Vendemos singles, producto sellado y accesorios. ¡Visítanos en persona o compra online!"},
		{"pages.notices.detail.sidebar.shop_btn", "en", "SHOP SINGLES"},
		{"pages.notices.detail.sidebar.shop_btn", "es", "COMPRAR SINGLES"},
		{"pages.notices.detail.sidebar.newsletter_title", "en", "Newsletter"},
		{"pages.notices.detail.sidebar.newsletter_title", "es", "Boletín Informativo"},
		{"pages.notices.detail.sidebar.newsletter_desc", "en", "Stay updated with our latest news and spoilers."},
		{"pages.notices.detail.sidebar.newsletter_desc", "es", "Mantente al día con nuestras últimas noticias y spoilers."},

		{"pages.notices.section.title", "en", "NOTICES / NEWS"},
		{"pages.notices.section.title", "es", "AVISOS / NOTICIAS"},
		{"pages.notices.actions.view_all", "en", "VIEW ALL →"},
		{"pages.notices.actions.view_all", "es", "VER TODO →"},
		{"pages.notices.actions.read_more", "en", "READ MORE"},
		{"pages.notices.actions.read_more", "es", "LEER MÁS"},
		{"pages.notices.actions.show_more", "en", "SHOW MORE PREVIOUS POSTS"},
		{"pages.notices.actions.show_more", "es", "MOSTRAR MÁS PUBLICACIONES"},

		// Search Bar
		{"components.search.placeholder", "en", "Search for cards, sets, or products..."},
		{"components.search.placeholder", "es", "Buscar cartas, sets o productos..."},
		{"components.search.status.manual_price", "en", "MANUAL PRICE"},
		{"components.search.status.manual_price", "es", "PRECIO MANUAL"},
		{"components.search.status.no_set", "en", "No Set"},
		{"components.search.status.no_set", "es", "Sin Set"},
		{"components.search.status.in_stock", "en", "{count} IN STOCK"},
		{"components.search.status.in_stock", "es", "{count} EN STOCK"},
		{"components.search.status.out_of_stock", "en", "OUT OF STOCK"},
		{"components.search.status.out_of_stock", "es", "AGOTADO"},
		{"components.search.actions.add", "en", "+ ADD"},
		{"components.search.actions.add", "es", "+ AGREGAR"},
		{"components.search.actions.sold_out", "en", "SOLD OUT"},
		{"components.search.actions.sold_out", "es", "AGOTADO"},
		{"components.search.status.no_results", "en", "No products found for \"{query}\""},
		{"components.search.status.no_results", "es", "No se encontraron productos para \"{query}\""},
		{"components.search.status.top_results", "en", "Showing top results"},
		{"components.search.status.top_results", "es", "Mostrando resultados principales"},

		// Checkout
		{"pages.checkout.page.title", "en", "FINALIZE PURCHASE"},
		{"pages.checkout.page.title", "es", "FINALIZAR COMPRA"},
		{"pages.checkout.form.contact_info", "en", "CONTACT INFORMATION"},
		{"pages.checkout.form.contact_info", "es", "INFORMACIÓN DE CONTACTO"},
		{"pages.checkout.form.first_name", "en", "FIRST NAME"},
		{"pages.checkout.form.first_name", "es", "NOMBRE"},
		{"pages.checkout.form.last_name", "en", "LAST NAME"},
		{"pages.checkout.form.last_name", "es", "APELLIDO"},
		{"pages.checkout.form.phone", "en", "PHONE / WHATSAPP"},
		{"pages.checkout.form.phone", "es", "TELÉFONO / WHATSAPP"},
		{"pages.checkout.form.email", "en", "EMAIL"},
		{"pages.checkout.form.email", "es", "EMAIL"},
		{"pages.checkout.form.id_number", "en", "ID NUMBER"},
		{"pages.checkout.form.id_number", "es", "CÉDULA / ID"},
		{"pages.checkout.form.address", "en", "ADDRESS"},
		{"pages.checkout.form.address", "es", "DIRECCIÓN"},
		{"pages.checkout.form.payment_method", "en", "PAYMENT METHOD"},
		{"pages.checkout.form.payment_method", "es", "MÉTODO DE PAGO"},
		{"pages.checkout.form.notes", "en", "NOTES (OPTIONAL)"},
		{"pages.checkout.form.notes", "es", "NOTAS (OPCIONAL)"},
		{"pages.checkout.summary.title", "en", "ORDER SUMMARY"},
		{"pages.checkout.summary.title", "es", "RESUMEN DEL PEDIDO"},
		{"pages.checkout.summary.confirm_btn", "en", "CONFIRM ORDER →"},
		{"pages.checkout.summary.confirm_btn", "es", "CONFIRMAR PEDIDO →"},
		{"pages.checkout.summary.processing", "en", "PROCESSING..."},
		{"pages.checkout.summary.processing", "es", "PROCESANDO..."},
		{"pages.checkout.summary.footer_note", "en", "After confirmation, an advisor will contact you to coordinate delivery."},
		{"pages.checkout.summary.footer_note", "es", "Al confirmar, un asesor se pondrá en contacto contigo para coordinar la entrega."},

		// Admin Notices
		{"pages.admin.notices.title", "en", "NOTICES (BLOG/NEWS)"},
		{"pages.admin.notices.title", "es", "AVISOS (BLOG/NOTICIAS)"},
		{"pages.admin.notices.subtitle", "en", "Manage your shop updates and news posts."},
		{"pages.admin.notices.subtitle", "es", "Administra las actualizaciones y noticias de la tienda."},

		{"pages.admin.tcg_registry.title", "en", "Registered Systems"},
		{"pages.admin.tcg_registry.title", "es", "Sistemas Registrados"},
		{"pages.admin.tcg_registry.subtitle", "en", "Enable or disable game systems from the warehouse"},
		{"pages.admin.tcg_registry.subtitle", "es", "Activa o desactiva sistemas de juego del almacén"},
		{"pages.admin.tcg_registry.loading", "en", "UNPACKING TCG REGISTRY..."},
		{"pages.admin.tcg_registry.loading", "es", "DESEMPAQUETANDO REGISTRO TCG..."},
		{"pages.admin.tcg_registry.buttons.register_new", "en", "＋ REGISTER NEW TCG"},
		{"pages.admin.tcg_registry.buttons.register_new", "es", "＋ REGISTRAR NUEVO TCG"},
		{"pages.admin.tcg_registry.form.slug_label", "en", "Internal Slug (URL ID)"},
		{"pages.admin.tcg_registry.form.slug_label", "es", "Slug Interno (ID de URL)"},
		{"pages.admin.tcg_registry.form.display_name", "en", "Display Name"},
		{"pages.admin.tcg_registry.form.display_name", "es", "Nombre a Mostrar"},
		{"pages.admin.tcg_registry.buttons.confirm_registration", "en", "CONFIRM SYSTEM REGISTRATION"},
		{"pages.admin.tcg_registry.buttons.confirm_registration", "es", "CONFIRMAR REGISTRO DE SISTEMA"},
		{"pages.admin.tcg_registry.buttons.save", "en", "SAVE"},
		{"pages.admin.tcg_registry.buttons.save", "es", "GUARDAR"},
		{"pages.admin.tcg_registry.buttons.cancel", "en", "CANCEL"},
		{"pages.admin.tcg_registry.buttons.cancel", "es", "CANCELAR"},
		{"pages.admin.tcg_registry.no_systems", "en", "No Systems Registered"},
		{"pages.admin.tcg_registry.no_systems", "es", "No hay sistemas registrados"},

		{"components.admin.product.storage.quick_add", "en", "Quick Add Location"},
		{"components.admin.product.storage.quick_add", "es", "Agregar ubicación rápida"},
		{"components.admin.product.storage.title", "en", "STORAGE LOCATIONS"},
		{"components.admin.product.storage.title", "es", "UBICACIONES DE ALMACENAMIENTO"},
		{"components.admin.product.storage.empty", "en", "No storage assignments yet."},
		{"components.admin.product.storage.empty", "es", "Aún no hay asignaciones de almacenamiento."},
		{"components.admin.product.storage.select_placeholder", "en", "-- Select Location --"},
		{"components.admin.product.storage.select_placeholder", "es", "-- Seleccionar Ubicación --"},
		{"pages.admin.notices.create_btn", "en", "CREATE NEW NOTICE"},
		{"pages.admin.notices.create_btn", "es", "CREAR NUEVO AVISO"},
		{"pages.admin.notices.confirm_delete", "en", "Are you sure you want to delete this notice?"},
		{"pages.admin.notices.confirm_delete", "es", "¿Estás seguro de que deseas eliminar este aviso?"},
		{"pages.admin.notices.error_delete", "en", "Failed to delete notice"},
		{"pages.admin.notices.error_delete", "es", "Error al eliminar el aviso"},
		{"pages.admin.notices.status.published", "en", "PUBLISHED"},
		{"pages.admin.notices.status.published", "es", "PUBLICADO"},
		{"pages.admin.notices.status.draft", "en", "DRAFT"},
		{"pages.admin.notices.status.draft", "es", "BORRADOR"},
		{"pages.admin.notices.actions.view", "en", "VIEW"},
		{"pages.admin.notices.actions.view", "es", "VER"},
		{"pages.admin.notices.actions.edit", "en", "EDIT"},
		{"pages.admin.notices.actions.edit", "es", "EDITAR"},
		{"pages.admin.notices.actions.delete", "en", "DELETE"},
		{"pages.admin.notices.actions.delete", "es", "ELIMINAR"},
		{"pages.admin.notices.table.date", "en", "Date"},
		{"pages.admin.notices.table.date", "es", "Fecha"},
		{"pages.admin.notices.table.title", "en", "Title"},
		{"pages.admin.notices.table.title", "es", "Título"},
		{"pages.admin.notices.table.slug", "en", "Slug"},
		{"pages.admin.notices.table.slug", "es", "Slug"},
		{"pages.admin.notices.table.status", "en", "Status"},
		{"pages.admin.notices.table.status", "es", "Estado"},
		{"pages.admin.notices.table.actions", "en", "Actions"},
		{"pages.admin.notices.table.actions", "es", "Acciones"},
		{"pages.admin.notices.table.empty", "en", "No notices found. Create your first one!"},
		{"pages.admin.notices.table.empty", "es", "No se encontraron noticias. ¡Crea la primera!"},

		{"pages.admin.notices.editor.loading", "en", "Unpacking Notice..."},
		{"pages.admin.notices.editor.loading", "es", "Desempaquetando Aviso..."},
		{"pages.admin.notices.editor.title_edit", "en", "EDIT NOTICE"},
		{"pages.admin.notices.editor.title_edit", "es", "EDITAR AVISO"},
		{"pages.admin.notices.editor.title_new", "en", "NEW NOTICE"},
		{"pages.admin.notices.editor.title_new", "es", "NUEVO AVISO"},
		{"pages.admin.notices.editor.subtitle", "en", "Compose your shop update using raw HTML."},
		{"pages.admin.notices.editor.subtitle", "es", "Redacta la actualización de tu tienda usando HTML puro."},
		{"pages.admin.notices.editor.error_save", "en", "Failed to save notice"},
		{"pages.admin.notices.editor.error_save", "es", "Error al guardar el aviso"},
		{"pages.admin.notices.editor.form.title_label", "en", "Post Title"},
		{"pages.admin.notices.editor.form.title_label", "es", "Título de la Publicación"},
		{"pages.admin.notices.editor.form.title_placeholder", "en", "E.G. NEW SINGLES JUST ARRIVED!"},
		{"pages.admin.notices.editor.form.title_placeholder", "es", "EJ. ¡LLEGARON NUEVAS SINGLES!"},
		{"pages.admin.notices.editor.form.content_label", "en", "HTML Content"},
		{"pages.admin.notices.editor.form.content_label", "es", "Contenido HTML"},
		{"pages.admin.notices.editor.form.content_placeholder", "en", "<h2>Subheading</h2><p>Write your content here...</p>"},
		{"pages.admin.notices.editor.form.content_placeholder", "es", "<h2>Subtítulo</h2><p>Escribe tu contenido aquí...</p>"},
		{"pages.admin.notices.editor.sidebar.settings_title", "en", "Publish Settings"},
		{"pages.admin.notices.editor.sidebar.settings_title", "es", "Ajustes de Publicación"},
		{"pages.admin.notices.editor.sidebar.slug_label", "en", "URL Slug"},
		{"pages.admin.notices.editor.sidebar.slug_label", "es", "Slug de URL"},
		{"pages.admin.notices.editor.sidebar.sync_btn", "en", "SYNC WITH TITLE"},
		{"pages.admin.notices.editor.sidebar.sync_btn", "es", "SINCRONIZAR CON TÍTULO"},
		{"pages.admin.notices.editor.sidebar.slug_placeholder", "en", "post-url-slug"},
		{"pages.admin.notices.editor.sidebar.slug_placeholder", "es", "slug-de-la-publicacion"},
		{"pages.admin.notices.editor.sidebar.image_label", "en", "Featured Image URL"},
		{"pages.admin.notices.editor.sidebar.image_label", "es", "URL de Imagen Destacada"},
		{"pages.admin.notices.editor.sidebar.image_placeholder", "en", "https://..."},
		{"pages.admin.notices.editor.sidebar.image_placeholder", "es", "https://..."},
		{"pages.admin.notices.editor.sidebar.published_label", "en", "Published / Public"},
		{"pages.admin.notices.editor.sidebar.published_label", "es", "Publicado / Público"},
		{"pages.admin.notices.editor.actions.update", "en", "UPDATE POST"},
		{"pages.admin.notices.editor.actions.update", "es", "ACTUALIZAR POST"},
		{"pages.admin.notices.editor.actions.publish", "en", "PUBLISH POST"},
		{"pages.admin.notices.editor.actions.publish", "es", "PUBLICAR POST"},
		{"pages.admin.notices.editor.sidebar.help_title", "en", "HTML Help"},
		{"pages.admin.notices.editor.sidebar.help_title", "es", "Ayuda HTML"},
		{"pages.admin.notices.editor.sidebar.help_desc", "en", "You can use standard HTML tags like {tags}."},
		{"pages.admin.notices.editor.sidebar.help_desc", "es", "Puedes usar etiquetas HTML estándar como {tags}."},
		{"pages.admin.notices.editor.sidebar.card_preview_label", "en", "To embed a card preview, use:"},
		{"pages.admin.notices.editor.sidebar.card_preview_label", "es", "Para incrustar una vista previa de carta, usa:"},

		{"pages.admin.clients.details.not_found", "en", "CLIENT NOT FOUND."},
		{"pages.admin.clients.details.not_found", "es", "CLIENTE NO ENCONTRADO."},
		{"pages.admin.clients.details.subscriber_badge", "en", "NEWSLETTER SUBSCRIBER"},
		{"pages.admin.clients.details.subscriber_badge", "es", "SUSCRIPTOR AL NEWSLETTER"},
		{"pages.admin.clients.details.contact_profile", "en", "Contact Profile"},
		{"pages.admin.clients.details.contact_profile", "es", "Perfil de contacto"},
		{"pages.admin.clients.details.email_label", "en", "Email address"},
		{"pages.admin.clients.details.email_label", "es", "Correo electrónico"},
		{"pages.admin.clients.details.phone_label", "en", "Phone number"},
		{"pages.admin.clients.details.phone_label", "es", "Número de teléfono"},
		{"pages.admin.clients.details.account_details", "en", "Account Details"},
		{"pages.admin.clients.details.account_details", "es", "Detalles de cuenta"},
		{"pages.admin.clients.details.id_label", "en", "Identity Document"},
		{"pages.admin.clients.details.id_label", "es", "Documento de identidad"},
		{"pages.admin.clients.details.address_label", "en", "Delivery Address"},
		{"pages.admin.clients.details.address_label", "es", "Dirección de entrega"},
		{"pages.admin.clients.details.general_interaction", "en", "General Interaction / No specific order"},
		{"pages.admin.clients.details.general_interaction", "es", "Interacción general / Sin pedido específico"},
		{"pages.admin.clients.details.add_note_btn", "en", "ADD NOTE"},
		{"pages.admin.clients.details.add_note_btn", "es", "AGREGAR NOTA"},
		{"pages.admin.clients.details.empty_notes", "en", "No entries recorded in the journal."},
		{"pages.admin.clients.details.empty_notes", "es", "No hay entradas registradas en el diario."},
		{"pages.admin.clients.details.order_history", "en", "ORDER HISTORY"},
		{"pages.admin.clients.details.order_history", "es", "HISTORIAL DE PEDIDOS"},
		{"pages.admin.clients.details.no_orders", "en", "NO ORDERS FOUND."},
		{"pages.admin.clients.details.no_orders", "es", "NO SE ENCONTRARON PEDIDOS."},
		{"pages.admin.clients.details.journal_title", "en", "JOURNAL OF INTERACTIONS"},
		{"pages.admin.clients.details.journal_title", "es", "DIARIO DE INTERACCIONES"},
		{"pages.admin.clients.details.note_placeholder", "en", "ADD A NEW NOTE OR COMMENT..."},
		{"pages.admin.clients.details.note_placeholder", "es", "AGREGAR UNA NUEVA NOTA O COMENTARIO..."},
		{"pages.admin.clients.details.order_option", "en", "About Order {number}"},
		{"pages.admin.clients.details.order_option", "es", "Sobre el pedido {number}"},
		{"pages.admin.clients.details.reference_label", "en", "REFERENCE"},
		{"pages.admin.clients.details.reference_label", "es", "REFERENCIA"},
		{"pages.admin.clients.details.summary_title", "en", "VALUED CLIENT SUMMARY"},
		{"pages.admin.clients.details.summary_title", "es", "RESUMEN DEL CLIENTE"},
		{"pages.admin.clients.details.lifetime_label", "en", "Lifetime"},
		{"pages.admin.clients.details.lifetime_label", "es", "Total Histórico"},
		{"pages.admin.clients.details.purchased_label", "en", "Purchased"},
		{"pages.admin.clients.details.purchased_label", "es", "Comprado"},
		{"pages.admin.clients.details.back_link", "en", "← BACK TO CLIENTS"},
		{"pages.admin.clients.details.back_link", "es", "← VOLVER A CLIENTES"},
		{"pages.admin.clients.details.active_group", "en", "Active"},
		{"pages.admin.clients.details.active_group", "es", "Activo"},
		{"pages.admin.clients.details.past_group", "en", "Past"},
		{"pages.admin.clients.details.past_group", "es", "Pasado"},
		{"pages.admin.clients.details.no_active_requests", "en", "NO ACTIVE REQUESTS."},
		{"pages.admin.clients.details.no_active_requests", "es", "SIN SOLICITUDES ACTIVAS."},
		{"pages.admin.clients.details.no_pending_offers", "en", "NO PENDING OFFERS."},
		{"pages.admin.clients.details.no_pending_offers", "es", "SIN OFERTAS PENDIENTES."},
		{"pages.admin.clients.details.requests_title", "en", "CLIENT REQUESTS"},
		{"pages.admin.clients.details.requests_title", "es", "SOLICITUDES DE CLIENTES"},
		{"pages.admin.clients.details.offers_title", "en", "BOUNTY OFFERS"},
		{"pages.admin.clients.details.offers_title", "es", "OFERTAS DE BOUNTIES"},

		{"pages.admin.bounties.focus_mode", "en", "Focus Mode: Hide Solved History"},
		{"pages.admin.bounties.focus_mode", "es", "Modo Enfoque: Ocultar resueltos"},
		{"pages.admin.bounties.offering_card", "en", "Offering Card:"},
		{"pages.admin.bounties.offering_card", "es", "Ofreciendo carta:"},
		{"pages.admin.bounties.resolve_btn", "en", "RESOLVE OFFER"},
		{"pages.admin.bounties.resolve_btn", "es", "RESOLVER OFERTA"},
		{"pages.admin.bounties.revert_btn", "en", "REVERT TO PENDING"},
		{"pages.admin.bounties.revert_btn", "es", "REVERTIR A PENDIENTE"},

		// Admin Sidebar
		{"components.admin.sidebar.nav.inventory", "en", "INVENTORY"},
		{"components.admin.sidebar.nav.inventory", "es", "INVENTARIO"},
		{"components.admin.sidebar.nav.tcg_registry", "en", "TCG REGISTRY"},
		{"components.admin.sidebar.nav.tcg_registry", "es", "REGISTRO TCG"},
		{"components.admin.sidebar.nav.orders", "en", "ORDERS"},
		{"components.admin.sidebar.nav.orders", "es", "PEDIDOS"},
		{"components.admin.sidebar.nav.bounties", "en", "WANTED / BOUNTIES"},
		{"components.admin.sidebar.nav.bounties", "es", "BUSCADOS / BOUNTIES"},
		{"components.admin.sidebar.nav.clients", "en", "CLIENTS"},
		{"components.admin.sidebar.nav.clients", "es", "CLIENTES"},
		{"components.admin.sidebar.nav.subscribers", "en", "SUBSCRIBERS"},
		{"components.admin.sidebar.nav.subscribers", "es", "SUSCRIPTORES"},
		{"components.admin.sidebar.nav.notices", "en", "NOTICES"},
		{"components.admin.sidebar.nav.notices", "es", "AVISOS"},
		{"components.admin.sidebar.nav.themes", "en", "THEMES & SKINS"},
		{"components.admin.sidebar.nav.themes", "es", "TEMAS Y APARIENCIA"},
		{"components.admin.sidebar.nav.translations", "en", "TRANSLATIONS"},
		{"components.admin.sidebar.nav.translations", "es", "TRADUCCIONES"},
		{"components.admin.sidebar.nav.settings", "en", "GLOBAL SETTINGS"},
		{"components.admin.sidebar.nav.settings", "es", "AJUSTES GLOBALES"},
		{"components.admin.sidebar.section.ops", "en", "Core Operations"},
		{"components.admin.sidebar.section.ops", "es", "Operaciones Core"},
		{"components.admin.sidebar.section.design", "en", "Design & Language"},
		{"components.admin.sidebar.section.design", "es", "Diseño e Idioma"},
		{"components.admin.sidebar.section.system", "en", "System Actions"},
		{"components.admin.sidebar.section.system", "es", "Acciones del Sistema"},
		{"components.admin.sidebar.health.title", "en", "System Health"},
		{"components.admin.sidebar.health.title", "es", "Salud del Sistema"},
		{"components.admin.sidebar.health.refresh", "en", "Refresh Stats"},
		{"components.admin.sidebar.health.refresh", "es", "Refrescar Estadísticas"},
		{"components.admin.sidebar.health.sync", "en", "Synchronizing core..."},
		{"components.admin.sidebar.health.sync", "es", "Sincronizando núcleo..."},
		{"components.admin.sidebar.health.db_size", "en", "DB Size"},
		{"components.admin.sidebar.health.db_size", "es", "Tamaño DB"},
		{"components.admin.sidebar.health.latency", "en", "Latency"},
		{"components.admin.sidebar.health.latency", "es", "Latencia"},
		{"components.admin.sidebar.health.clients", "en", "Active Clients"},
		{"components.admin.sidebar.health.clients", "es", "Clientes Activos"},
		{"components.admin.sidebar.health.cache", "en", "Cache"},
		{"components.admin.sidebar.health.cache", "es", "Caché"},
		{"components.admin.sidebar.auth.logout", "en", "LOG OUT SESSION"},
		{"components.admin.sidebar.auth.logout", "es", "CERRAR SESIÓN"},
		{"components.admin.sidebar.status.secure", "en", "Secure Link Active"},
		{"components.admin.sidebar.status.secure", "es", "Enlace Seguro Activo"},
		{"pages.admin.bounties.offers.waiting_clients", "en", "WAITING CLIENTS"},
		{"pages.admin.bounties.offers.waiting_clients", "es", "CLIENTES ESPERANDO"},
		{"pages.admin.bounties.offers.select_clients", "en", "Select Clients to Fulfill (Max {qty})"},
		{"pages.admin.bounties.offers.select_clients", "es", "Seleccionar clientes para cumplir (Máx {qty})"},
		{"pages.admin.bounties.offers.over_limit", "en", "⚠️ OVER QUANTITY LIMIT"},
		{"pages.admin.bounties.offers.over_limit", "es", "⚠️ LÍMITE DE CANTIDAD EXCEDIDO"},
		{"pages.admin.bounties.requests.complete_status", "en", "COMPLETE"},
		{"pages.admin.bounties.requests.complete_status", "es", "COMPLETO"},
		{"pages.admin.bounties.title", "en", "WANTED / BOUNTIES"},
		{"pages.admin.bounties.title", "es", "BUSCADOS / BOUNTIES"},
		{"pages.admin.bounties.subtitle", "en", "Cards We Want to Buy // Client Requests"},
		{"pages.admin.bounties.subtitle", "es", "Cartas que queremos comprar // Solicitudes de clientes"},
		{"pages.admin.bounties.add_btn", "en", "ADD NEW BOUNTY"},
		{"pages.admin.bounties.add_btn", "es", "AGREGAR NUEVO BOUNTY"},
		{"pages.admin.bounties.tabs.bounties", "en", "WANTED LIST"},
		{"pages.admin.bounties.tabs.bounties", "es", "LISTA DE BUSCADOS"},
		{"pages.admin.bounties.tabs.offers", "en", "OFFERS VERIFICATION"},
		{"pages.admin.bounties.tabs.offers", "es", "VERIFICACIÓN DE OFERTAS"},
		{"pages.admin.bounties.tabs.requests", "en", "CLIENT REQUESTS"},
		{"pages.admin.bounties.tabs.requests", "es", "SOLICITUDES DE CLIENTES"},
		{"pages.admin.bounties.tabs.pending_suffix", "en", "PENDING"},
		{"pages.admin.bounties.tabs.pending_suffix", "es", "PENDIENTE"},
		{"pages.admin.bounties.loading", "en", "Synchronizing Logistics..."},
		{"pages.admin.bounties.loading", "es", "Sincronizando logística..."},
		{"pages.admin.bounties.table.card", "en", "Card"},
		{"pages.admin.bounties.table.card", "es", "Carta"},
		{"pages.admin.bounties.table.set_info", "en", "Set / Info"},
		{"pages.admin.bounties.table.set_info", "es", "Set / Info"},
		{"pages.admin.bounties.table.condition", "en", "Cond."},
		{"pages.admin.bounties.table.condition", "es", "Est."},
		{"pages.admin.bounties.table.target_price", "en", "Target Price"},
		{"pages.admin.bounties.table.target_price", "es", "Precio Objetivo"},
		{"pages.admin.bounties.table.qty", "en", "Qty"},
		{"pages.admin.bounties.table.qty", "es", "Cant."},
		{"pages.admin.bounties.table.status", "en", "Status"},
		{"pages.admin.bounties.table.status", "es", "Estado"},
		{"pages.admin.bounties.table.actions", "en", "Actions"},
		{"pages.admin.bounties.table.actions", "es", "Acciones"},
		{"pages.admin.bounties.offers.seller", "en", "Seller:"},
		{"pages.admin.bounties.offers.seller", "es", "Vendedor:"},
		{"pages.admin.bounties.offers.condition", "en", "Condition:"},
		{"pages.admin.bounties.offers.condition", "es", "Estado:"},
		{"pages.admin.bounties.offers.quantity", "en", "Quantity:"},
		{"pages.admin.bounties.offers.quantity", "es", "Cantidad:"},
		{"pages.admin.bounties.offers.waiting_clients", "en", "WAITING CLIENTS"},
		{"pages.admin.bounties.offers.waiting_clients", "es", "CLIENTES ESPERANDO"},
		{"pages.admin.bounties.offers.submitted_on", "en", "Submitted on:"},
		{"pages.admin.bounties.offers.submitted_on", "es", "Enviado el:"},
		{"pages.admin.bounties.offers.select_clients", "en", "Select Clients to Fulfill (Max {qty})"},
		{"pages.admin.bounties.offers.select_clients", "es", "Selecciona clientes para completar (Máx {qty})"},
		{"pages.admin.bounties.offers.over_limit", "en", "⚠️ OVER QUANTITY LIMIT"},
		{"pages.admin.bounties.offers.over_limit", "es", "⚠️ EXCESO DE CANTIDAD"},
		{"pages.admin.bounties.requests.complete_status", "en", "COMPLETE"},
		{"pages.admin.bounties.requests.complete_status", "es", "COMPLETO"},
		{"pages.admin.bounties.requests.client_label", "en", "Client:"},
		{"pages.admin.bounties.requests.client_label", "es", "Cliente:"},
		{"pages.admin.bounties.requests.requested_date", "en", "Requested:"},
		{"pages.admin.bounties.requests.requested_date", "es", "Solicitado:"},
		{"pages.admin.bounties.requests.no_details", "en", "No additional details provided."},
		{"pages.admin.bounties.requests.no_details", "es", "No se proporcionaron detalles adicionales."},
		{"pages.admin.bounties.requests.accept_btn", "en", "ACCEPT & ADD BOUNTY"},
		{"pages.admin.bounties.requests.accept_btn", "es", "ACEPTAR Y AGREGAR BOUNTY"},
		{"pages.admin.bounties.requests.solve_btn", "en", "MARK AS SOLVED"},
		{"pages.admin.bounties.requests.solve_btn", "es", "MARCAR COMO RESUELTO"},
		{"pages.admin.bounties.requests.reject_btn", "en", "REJECT"},
		{"pages.admin.bounties.requests.reject_btn", "es", "RECHAZAR"},
		{"pages.admin.bounties.requests.mission_complete", "en", "MISSION COMPLETE\nCARD DELIVERED"},
		{"pages.admin.bounties.requests.mission_complete", "es", "MISIÓN COMPLETADA\nCARTA ENTREGADA"},

		// Bounty Modal (Slugs: components.admin.bounty_modal)
		{"components.admin.bounty_modal.title_edit", "en", "EDIT WANTED CARD"},
		{"components.admin.bounty_modal.title_edit", "es", "EDITAR CARTA BUSCADA"},
		{"components.admin.bounty_modal.title_add", "en", "ADD WANTED CARD"},
		{"components.admin.bounty_modal.title_add", "es", "AGREGAR CARTA BUSCADA"},
		{"components.admin.bounty_modal.tcg_label", "en", "TCG SYSTEM"},
		{"components.admin.bounty_modal.tcg_label", "es", "SISTEMA TCG"},
		{"components.admin.bounty_modal.condition_label", "en", "CONDITION"},
		{"components.admin.bounty_modal.condition_label", "es", "ESTADO"},
		{"components.admin.bounty_modal.card_name", "en", "CARD NAME"},
		{"components.admin.bounty_modal.card_name", "es", "NOMBRE DE LA CARTA"},
		{"components.admin.bounty_modal.set_name", "en", "SET NAME / INFO"},
		{"components.admin.bounty_modal.set_name", "es", "NOMBRE DEL SET / INFO"},
		{"components.admin.bounty_modal.image_url", "en", "IMAGE URL"},
		{"components.admin.bounty_modal.image_url", "es", "URL DE LA IMAGEN"},
		{"components.admin.bounty_modal.pricing_title", "en", "PRICING"},
		{"components.admin.bounty_modal.pricing_title", "es", "PRECIOS"},
		{"components.admin.bounty_modal.price_source", "en", "PRICE SOURCE *"},
		{"components.admin.bounty_modal.price_source", "es", "FUENTE DE PRECIO *"},
		{"components.admin.bounty_modal.source_manual", "en", "Manual Override (COP)"},
		{"components.admin.bounty_modal.source_manual", "es", "Anulación Manual (COP)"},
		{"components.admin.bounty_modal.source_tcgplayer", "en", "External: TCGPlayer (USD)"},
		{"components.admin.bounty_modal.source_tcgplayer", "es", "Externo: TCGPlayer (USD)"},
		{"components.admin.bounty_modal.source_cardmarket", "en", "External: Cardmarket (EUR)"},
		{"components.admin.bounty_modal.source_cardmarket", "es", "Externo: Cardmarket (EUR)"},
		{"components.admin.bounty_modal.ref_price_label", "en", "REFERENCE PRICE ({source}) *"},
		{"components.admin.bounty_modal.ref_price_label", "es", "PRECIO DE REFERENCIA ({source}) *"},
		{"components.admin.bounty_modal.price_cop_label", "en", "PRICE (COP) *"},
		{"components.admin.bounty_modal.price_cop_label", "es", "PRECIO (COP) *"},

		// Product Edit Modal (Slugs: components.admin.product_modal)
		{"components.admin.product_modal.title_edit", "en", "EDIT PRODUCT"},
		{"components.admin.product_modal.title_edit", "es", "EDITAR PRODUCTO"},
		{"components.admin.product_modal.title_new", "en", "NEW PRODUCT"},
		{"components.admin.product_modal.title_new", "es", "NUEVO PRODUCTO"},
		{"components.admin.product_modal.product_id", "en", "PRODUCT ID: {id}"},
		{"components.admin.product_modal.product_id", "es", "ID PRODUCTO: {id}"},
		{"components.admin.product_modal.tcg_system_label", "en", "TCG SYSTEM"},
		{"components.admin.product_modal.tcg_system_label", "es", "SISTEMA TCG"},
		{"components.admin.product_modal.category_label", "en", "CATEGORY"},
		{"components.admin.product_modal.category_label", "es", "CATEGORÍA"},
		{"components.admin.product_modal.condition_label", "en", "CONDITION"},
		{"components.admin.product_modal.condition_label", "es", "ESTADO"},
		{"components.admin.product_modal.product_name_label", "en", "PRODUCT NAME"},
		{"components.admin.product_modal.product_name_label", "es", "NOMBRE DEL PRODUCTO"},
		{"components.admin.product_modal.tab_deck", "en", "DECK BUILDER"},
		{"components.admin.product_modal.tab_deck", "es", "CONSTRUCTOR DE MAZO"},
		{"components.admin.product_modal.tab_variant", "en", "VARIANT & IDENTITY"},
		{"components.admin.product_modal.tab_variant", "es", "VARIANTE E IDENTIDAD"},
		{"components.admin.product_modal.tab_pricing", "en", "PRICING & STOCK"},
		{"components.admin.product_modal.tab_pricing", "es", "PRECIO Y STOCK"},
		{"components.admin.product_modal.image_preview_label", "en", "IMAGE PREVIEW"},
		{"components.admin.product_modal.image_preview_label", "es", "VISTA PREVIA"},
		{"components.admin.product_modal.saving", "en", "SAVING..."},
		{"components.admin.product_modal.saving", "es", "GUARDANDO..."},
		{"components.admin.product_modal.save_btn", "en", "SAVE PRODUCT"},
		{"components.admin.product_modal.save_btn", "es", "GUARDAR PRODUCTO"},
		{"components.admin.product_modal.save_and_new_btn", "en", "SAVE & ADD NEW"},
		{"components.admin.product_modal.save_and_new_btn", "es", "GUARDAR Y NUEVO"},
		{"components.admin.product_modal.cancel_btn", "en", "CANCEL"},
		{"components.admin.product_modal.cancel_btn", "es", "CANCELAR"},
		{"components.admin.product_modal.error_required", "en", "Name, TCG, and Category are required."},
		{"components.admin.product_modal.error_required", "es", "Nombre, TCG y Categoría son requeridos."},
		{"components.admin.product_modal.error_save", "en", "Failed to save product."},
		{"components.admin.product_modal.error_save", "es", "Error al guardar el producto."},
		{"components.admin.product_modal.error_no_prints", "en", "No printings found for that search."},
		{"components.admin.product_modal.error_no_prints", "es", "No se encontraron impresiones para esa búsqueda."},
		{"components.admin.product_modal.error_no_match", "en", "Could not identify a matching print."},
		{"components.admin.product_modal.error_no_match", "es", "No se pudo identificar una impresión coincidente."},
		{"components.admin.product_modal.error_fetch", "en", "Scryfall fetch failed"},
		{"components.admin.product_modal.error_fetch", "es", "La descarga de Scryfall falló"},

		// Variant Tab (Slugs: components.admin.product_modal.variant)
		{"components.admin.product_modal.variant.set_info_title", "en", "SET INFORMATION"},
		{"components.admin.product_modal.variant.set_info_title", "es", "INFORMACIÓN DEL SET"},
		{"components.admin.product_modal.variant.set_name_label", "en", "SET NAME"},
		{"components.admin.product_modal.variant.set_name_label", "es", "NOMBRE DEL SET"},
		{"components.admin.product_modal.variant.set_code_label", "en", "CODE"},
		{"components.admin.product_modal.variant.set_code_label", "es", "CÓDIGO"},
		{"components.admin.product_modal.variant.collector_number_label", "en", "COLLECTOR #"},
		{"components.admin.product_modal.variant.collector_number_label", "es", "NÚMERO #"},
		{"components.admin.product_modal.variant.language_label", "en", "CARD LANGUAGE"},
		{"components.admin.product_modal.variant.language_label", "es", "IDIOMA DE LA CARTA"},
		{"components.admin.product_modal.variant.treatment_title", "en", "TREATMENT & FINISH"},
		{"components.admin.product_modal.variant.treatment_title", "es", "TRATAMIENTO Y ACABADO"},
		{"components.admin.product_modal.variant.foil_label", "en", "SHINE (FOIL)"},
		{"components.admin.product_modal.variant.foil_label", "es", "BRILLO (FOIL)"},
		{"components.admin.product_modal.variant.card_treatment_label", "en", "CARD TREATMENT"},
		{"components.admin.product_modal.variant.card_treatment_label", "es", "TRATAMIENTO DE CARTA"},
		{"components.admin.product_modal.variant.promo_type_label", "en", "PROMO TYPE"},
		{"components.admin.product_modal.variant.promo_type_label", "es", "TIPO DE PROMO"},
		{"components.admin.product_modal.variant.art_variation_label", "en", "ART VARIATION"},
		{"components.admin.product_modal.variant.art_variation_label", "es", "VARIACIÓN DE ARTE"},

		{"components.admin.product_modal.variant.optimized_hint", "en", "VARIANT OPTIONS ARE CURRENTLY OPTIMIZED FOR MTG SINGLES."},
		{"components.admin.product_modal.variant.optimized_hint", "es", "LAS OPCIONES DE VARIANTE ESTÁN OPTIMIZADAS PARA MTG SINGLES."},
		{"components.admin.product_modal.variant.general_hint", "en", "General product details are managed in the main header and Pricing tab."},
		{"components.admin.product_modal.variant.general_hint", "es", "Los detalles generales del producto se gestionan en el encabezado y la pestaña de Precios."},
		{"components.admin.product_modal.variant.metadata_title", "en", "MTG Metadata"},
		{"components.admin.product_modal.variant.metadata_title", "es", "Metadatos MTG"},
		{"components.admin.product_modal.variant.rarity_label", "en", "Rarity"},
		{"components.admin.product_modal.variant.rarity_label", "es", "Rareza"},
		{"components.admin.product_modal.variant.colors_label", "en", "Colors"},
		{"components.admin.product_modal.variant.colors_label", "es", "Colores"},
		{"components.admin.product_modal.variant.cmc_label", "en", "CMC"},
		{"components.admin.product_modal.variant.cmc_label", "es", "CMC"},
		{"components.admin.product_modal.variant.legendary_label", "en", "LEGENDARY"},
		{"components.admin.product_modal.variant.legendary_label", "es", "LEGENDARIA"},
		{"components.admin.product_modal.variant.historic_label", "en", "HISTORIC"},
		{"components.admin.product_modal.variant.historic_label", "es", "HISTÓRICA"},
		{"components.admin.product_modal.variant.land_label", "en", "LAND"},
		{"components.admin.product_modal.variant.land_label", "es", "TIERRA"},
		{"components.admin.product_modal.variant.basic_label", "en", "BASIC"},
		{"components.admin.product_modal.variant.basic_label", "es", "BÁSICA"},
		{"components.admin.product_modal.variant.full_art_label", "en", "FULL ART"},
		{"components.admin.product_modal.variant.full_art_label", "es", "ARTE COMPLETO"},
		{"components.admin.product_modal.variant.textless_label", "en", "TEXTLESS"},
		{"components.admin.product_modal.variant.textless_label", "es", "SIN TEXTO"},
		{"components.admin.product_modal.variant.type_line_label", "en", "Type Line"},
		{"components.admin.product_modal.variant.type_line_label", "es", "Línea de Tipo"},
		{"components.admin.product_modal.variant.artist_label", "en", "Artist"},
		{"components.admin.product_modal.variant.artist_label", "es", "Artista"},
		{"components.admin.product_modal.variant.oracle_text_label", "en", "Oracle Text"},
		{"components.admin.product_modal.variant.oracle_text_label", "es", "Texto Oracle"},

		// Pricing Tab (Slugs: components.admin.product_modal.pricing)
		{"components.admin.product_modal.pricing.stock_title", "en", "STOCK MANAGEMENT"},
		{"components.admin.product_modal.pricing.stock_title", "es", "GESTIÓN DE STOCK"},
		{"components.admin.product_modal.pricing.total_stock_label", "en", "TOTAL STOCK"},
		{"components.admin.product_modal.pricing.total_stock_label", "es", "STOCK TOTAL"},
		{"components.admin.product_modal.pricing.add_location_btn", "en", "ASSIGN NEW LOCATION"},
		{"components.admin.product_modal.pricing.add_location_btn", "es", "ASIGNAR NUEVA UBICACIÓN"},
		{"components.admin.product_modal.pricing.description_label", "en", "DESCRIPTION / NOTES (INTERNAL)"},
		{"components.admin.product_modal.pricing.description_label", "es", "DESCRIPCIÓN / NOTAS (INTERNO)"},
		{"components.admin.product_modal.pricing.description_placeholder", "en", "Optional card notes..."},
		{"components.admin.product_modal.pricing.description_placeholder", "es", "Notas opcionales de la carta..."},
		{"components.admin.product_modal.pricing.collections_label", "en", "COLLECTIONS / TAGS"},
		{"components.admin.product_modal.pricing.collections_label", "es", "COLECCIONES / ETIQUETAS"},

		{"components.admin.product_modal.pricing.title", "en", "PRICING"},
		{"components.admin.product_modal.pricing.title", "es", "PRECIOS"},
		{"components.admin.product_modal.pricing.source_label", "en", "PRICE SOURCE *"},
		{"components.admin.product_modal.pricing.source_label", "es", "FUENTE DE PRECIO *"},
		{"components.admin.product_modal.pricing.source_manual", "en", "Manual Override (COP)"},
		{"components.admin.product_modal.pricing.source_manual", "es", "Anulación Manual (COP)"},
		{"components.admin.product_modal.pricing.source_tcgplayer", "en", "External: TCGPlayer (USD)"},
		{"components.admin.product_modal.pricing.source_tcgplayer", "es", "Externo: TCGPlayer (USD)"},
		{"components.admin.product_modal.pricing.source_cardmarket", "en", "External: Cardmarket (EUR)"},
		{"components.admin.product_modal.pricing.source_cardmarket", "es", "Externo: Cardmarket (EUR)"},
		{"components.admin.product_modal.pricing.price_cop_label", "en", "PRICE (COP) *"},
		{"components.admin.product_modal.pricing.price_cop_label", "es", "PRECIO (COP) *"},
		{"components.admin.product_modal.pricing.ref_price_label", "en", "REFERENCE PRICE ({currency}) *"},
		{"components.admin.product_modal.pricing.ref_price_label", "es", "PRECIO DE REFERENCIA ({currency}) *"},
		{"components.admin.product_modal.pricing.basic_title", "en", "Basic Details"},
		{"components.admin.product_modal.pricing.basic_title", "es", "Detalles Básicos"},
		{"components.admin.product_modal.pricing.image_url_label", "en", "Image URL"},
		{"components.admin.product_modal.pricing.image_url_label", "es", "URL de la Imagen"},
		{"components.admin.product_modal.pricing.no_collections", "en", "No custom collections defined."},
		{"components.admin.product_modal.pricing.no_collections", "es", "No hay colecciones personalizadas."},

		// Scryfall Populate (Slugs: components.admin.product_modal.scryfall)
		{"components.admin.product_modal.scryfall.search_hint", "en", "Enter Name or SET + CN and click Populate to fetch from Scryfall"},
		{"components.admin.product_modal.scryfall.search_hint", "es", "Ingrese Nombre o SET + CN y haga clic en Cargar para obtener de Scryfall"},
		{"components.admin.product_modal.scryfall.populate_btn", "en", "POPULATE DATA"},
		{"components.admin.product_modal.scryfall.populate_btn", "es", "CARGAR DATOS"},
		{"components.admin.product_modal.scryfall.looking_up", "en", "LOOKING UP..."},
		{"components.admin.product_modal.scryfall.looking_up", "es", "BUSCANDO..."},
		{"components.admin.product_modal.scryfall.no_image", "en", "NO IMAGE"},
		{"components.admin.product_modal.scryfall.no_image", "es", "SIN IMAGEN"},

		{"components.admin.product_modal.deck.select_printing_label", "en", "Select Specific Printing to Add"},
		{"components.admin.product_modal.deck.select_printing_label", "es", "Seleccionar Impresión Específica para Agregar"},
		{"components.admin.product_modal.deck.add_btn", "en", "+ ADD"},
		{"components.admin.product_modal.deck.add_btn", "es", "+ AGREGAR"},
		{"components.admin.product_modal.deck.custom_name_label", "en", "Card Name (Custom)"},
		{"components.admin.product_modal.deck.custom_name_label", "es", "Nombre de Carta (Personalizado)"},
		{"components.admin.product_modal.deck.custom_name_placeholder", "en", "Type a custom card name..."},
		{"components.admin.product_modal.deck.custom_name_placeholder", "es", "Escriba un nombre de carta personalizado..."},
		{"components.admin.product_modal.deck.add_custom_btn", "en", "Add Custom Card"},
		{"components.admin.product_modal.deck.add_custom_btn", "es", "Agregar Carta Personalizada"},
		{"components.admin.product_modal.deck.list_title", "en", "Current Deck List"},
		{"components.admin.product_modal.deck.list_title", "es", "Lista de Mazo Actual"},
		{"components.admin.product_modal.deck.card_count", "en", "{count} CARDS"},
		{"components.admin.product_modal.deck.card_count", "es", "{count} CARTAS"},
		{"components.admin.product_modal.deck.empty_msg", "en", "Deck is currently empty."},
		{"components.admin.product_modal.deck.empty_msg", "es", "El mazo está actualmente vacío."},
		{"components.admin.product_modal.deck.edit_tooltip", "en", "Edit/Repopulate"},
		{"components.admin.product_modal.deck.edit_tooltip", "es", "Editar/Recargar"},
		{"components.admin.product_modal.deck.remove_tooltip", "en", "Remove"},
		{"components.admin.product_modal.deck.remove_tooltip", "es", "Eliminar"},
		{"components.admin.bounty_modal.quantity_label", "en", "QUANTITY NEEDED"},
		{"components.admin.bounty_modal.quantity_label", "es", "CANTIDAD NECESARIA"},
		{"components.admin.bounty_modal.hide_price_label", "en", "Hide target price from public view"},
		{"components.admin.bounty_modal.hide_price_label", "es", "Ocultar precio objetivo del público"},
		{"components.admin.bounty_modal.hide_price_desc", "en", "If checked, users will see \"Contact for Price\" instead of your target COP value."},
		{"components.admin.bounty_modal.hide_price_desc", "es", "Si se marca, los usuarios verán \"Contactar para precio\" en lugar de su valor COP objetivo."},
		{"components.admin.bounty_modal.no_image", "en", "NO IMAGE"},
		{"components.admin.bounty_modal.no_image", "es", "SIN IMAGEN"},
		{"components.admin.bounty_modal.name_placeholder", "en", "Entity Name"},
		{"components.admin.bounty_modal.name_placeholder", "es", "Nombre de la entidad"},
		{"components.admin.bounty_modal.set_placeholder", "en", "Optional Set Name"},
		{"components.admin.bounty_modal.set_placeholder", "es", "Nombre opcional del set"},
		{"components.admin.bounty_modal.image_placeholder", "en", "https://..."},
		{"components.admin.bounty_modal.image_placeholder", "es", "https://..."},
		{"components.admin.bounty_modal.saving", "en", "SAVING..."},
		{"components.admin.bounty_modal.saving", "es", "GUARDANDO..."},
		{"components.admin.bounty_modal.save_btn", "en", "💾 SAVE BOUNTY"},
		{"components.admin.bounty_modal.save_btn", "es", "💾 GUARDAR BOUNTY"},
		{"components.admin.bounty_modal.saving", "en", "SAVING..."},
		{"components.admin.bounty_modal.saving", "es", "GUARDANDO..."},
		{"components.admin.bounty_modal.save_btn", "en", "💾 SAVE BOUNTY"},
		{"components.admin.bounty_modal.save_btn", "es", "💾 GUARDAR BOUNTY"},
		{"components.admin.bounty_modal.error_required", "en", "Name and TCG are required."},
		{"components.admin.bounty_modal.error_required", "es", "El nombre y el TCG son obligatorios."},
		{"components.admin.bounty_modal.error_save", "en", "Failed to save bounty."},
		{"components.admin.bounty_modal.error_save", "es", "Error al guardar el bounty."},

		// Resolve Offer Modal (Slugs: components.admin.resolve_modal)
		{"components.admin.resolve_modal.title", "en", "Resolve Offer"},
		{"components.admin.resolve_modal.title", "es", "Resolver Oferta"},
		{"components.admin.resolve_modal.offer_details", "en", "Offer details:"},
		{"components.admin.resolve_modal.offer_details", "es", "Detalles de la oferta:"},
		{"components.admin.resolve_modal.no_notes", "en", "No notes provided"},
		{"components.admin.resolve_modal.no_notes", "es", "No se proporcionaron notas"},
		{"components.admin.resolve_modal.action_title", "en", "ACTION UPON ACCEPTANCE"},
		{"components.admin.resolve_modal.action_title", "es", "ACCIÓN AL ACEPTAR"},
		{"components.admin.resolve_modal.action_inventory", "en", "Add to Inventory"},
		{"components.admin.resolve_modal.action_inventory", "es", "Agregar al Inventario"},
		{"components.admin.resolve_modal.action_inventory_desc", "en", "Accept the card and add it directly to open stock for sale."},
		{"components.admin.resolve_modal.action_inventory_desc", "es", "Acepte la carta y agréguela directamente al stock abierto para la venta."},
		{"components.admin.resolve_modal.action_fulfill_selected", "en", "Fulfill {count} Selected Requests"},
		{"components.admin.resolve_modal.action_fulfill_selected", "es", "Cumplir {count} Solicitudes Seleccionadas"},
		{"components.admin.resolve_modal.action_fulfill_matching", "en", "Fulfill Matching Requests"},
		{"components.admin.resolve_modal.action_fulfill_matching", "es", "Cumplir Solicitudes Coincidentes"},
		{"components.admin.resolve_modal.action_fulfill_selected_desc", "en", "Accept the card and notify the {count} clients you selected."},
		{"components.admin.resolve_modal.action_fulfill_selected_desc", "es", "Acepte la carta y notifique a los {count} clientes que seleccionó."},
		{"components.admin.resolve_modal.action_fulfill_matching_desc", "en", "Accept the card and notify ALL clients waiting for it."},
		{"components.admin.resolve_modal.action_fulfill_matching_desc", "es", "Acepte la carta y notifique a TODOS los clientes que la esperan."},
		{"components.admin.resolve_modal.matching_selected", "en", "{count} of {total} matching requests selected."},
		{"components.admin.resolve_modal.matching_selected", "es", "{count} de {total} solicitudes coincidentes seleccionadas."},
		{"components.admin.resolve_modal.matching_found", "en", "{total} matching requests found."},
		{"components.admin.resolve_modal.matching_found", "es", "{total} solicitudes coincidentes encontradas."},
		{"components.admin.resolve_modal.reject_btn", "en", "REJECT OFFER"},
		{"components.admin.resolve_modal.reject_btn", "es", "RECHAZAR OFERTA"},
		{"components.admin.resolve_modal.accept_btn", "en", "ACCEPT OFFER"},
		{"components.admin.resolve_modal.accept_btn", "es", "ACEPTAR OFERTA"},

		// Admin Inventory Page (Slugs: pages.admin.inventory)
		{"pages.admin.inventory.title", "en", "INVENTORY MANAGEMENT"},
		{"pages.admin.inventory.title", "es", "GESTIÓN DE INVENTARIO"},
		{"pages.admin.inventory.subtitle", "en", "Store Dashboard // Operations Active"},
		{"pages.admin.inventory.subtitle", "es", "Panel de Control // Operaciones Activas"},
		{"pages.admin.inventory.sync_sets_btn", "en", "SYNC SETS"},
		{"pages.admin.inventory.sync_sets_btn", "es", "SINCRONIZAR SETS"},
		{"pages.admin.inventory.syncing", "en", "SYNCING..."},
		{"pages.admin.inventory.syncing", "es", "SINCRONIZANDO..."},
		{"pages.admin.inventory.last_sync", "en", "LAST SYNC: {date}"},
		{"pages.admin.inventory.last_sync", "es", "ÚLTIMA SINCRONIZACIÓN: {date}"},
		{"pages.admin.inventory.import_csv_btn", "en", "IMPORT CSV"},
		{"pages.admin.inventory.import_csv_btn", "es", "IMPORTAR CSV"},
		{"pages.admin.inventory.add_product_btn", "en", "ADD NEW PRODUCT"},
		{"pages.admin.inventory.add_product_btn", "es", "AGREGAR PRODUCTO"},
		{"pages.admin.inventory.search_label", "en", "Product Search"},
		{"pages.admin.inventory.search_label", "es", "Búsqueda de Productos"},
		{"pages.admin.inventory.search_placeholder", "en", "Search by name, set, code..."},
		{"pages.admin.inventory.search_placeholder", "es", "Buscar por nombre, set, código..."},
		{"pages.admin.inventory.tcg_filter_label", "en", "TCG Filter"},
		{"pages.admin.inventory.tcg_filter_label", "es", "Filtro TCG"},
		{"pages.admin.inventory.category_label", "en", "Category"},
		{"pages.admin.inventory.category_label", "es", "Categoría"},
		{"pages.admin.inventory.storage_label", "en", "Physical Location"},
		{"pages.admin.inventory.storage_label", "es", "Ubicación Física"},
		{"pages.admin.inventory.manage_locations_tooltip", "en", "Manage Locations"},
		{"pages.admin.inventory.manage_locations_tooltip", "es", "Gestionar Ubicaciones"},
		{"components.admin.inventory.manage_locations_tooltip", "en", "Manage Locations"},
		{"components.admin.inventory.manage_locations_tooltip", "es", "Gestionar Ubicaciones"},
		{"pages.admin.inventory.manage_collections_tooltip", "en", "Manage Collections"},
		{"pages.admin.inventory.manage_collections_tooltip", "es", "Gestionar Colecciones"},
		{"pages.admin.inventory.count_label", "en", "INVENTORY COUNT"},
		{"pages.admin.inventory.count_label", "es", "CONTEO DE INVENTARIO"},
		{"pages.admin.inventory.response_time_label", "en", "RESPONSE TIME"},
		{"pages.admin.inventory.response_time_label", "es", "TIEMPO DE RESPUESTA"},
		{"pages.admin.inventory.confirm_delete", "en", "Are you sure you want to delete {name}?"},
		{"pages.admin.inventory.confirm_delete", "es", "¿Estás seguro de que quieres eliminar {name}?"},
		{"pages.admin.inventory.error_delete", "en", "Failed to delete product."},
		{"pages.admin.inventory.error_delete", "es", "Error al eliminar el producto."},
		{"pages.admin.inventory.sync_success", "en", "Successfully synced {count} sets!"},
		{"pages.admin.inventory.sync_success", "es", "¡Se sincronizaron correctamente {count} sets!"},
		{"pages.admin.inventory.sync_error", "en", "Failed to sync sets."},
		{"pages.admin.inventory.sync_error", "es", "Error al sincronizar los sets."},
		{"pages.admin.inventory.confirm_delete_storage_items", "en", "Location \"{name}\" contains {count} items. Deleting it will clear these assignments. Continue?"},
		{"pages.admin.inventory.confirm_delete_storage_items", "es", "La ubicación \"{name}\" contiene {count} artículos. Al eliminarla se borrarán estas asignaciones. ¿Continuar?"},
		{"pages.admin.inventory.confirm_delete_storage", "en", "Delete storage location \"{name}\"?"},
		{"pages.admin.inventory.confirm_delete_storage", "es", "¿Eliminar ubicación de almacenamiento \"{name}\"?"},
		{"pages.admin.inventory.confirm_delete_category", "en", "Delete collection \"{name}\"?"},
		{"pages.admin.inventory.confirm_delete_category", "es", "¿Eliminar colección \"{name}\"?"},

		// Common Dates (Slugs: pages.common.dates)
		{"pages.common.dates.just_now", "en", "Just now"},
		{"pages.common.dates.just_now", "es", "Justo ahora"},
		{"pages.common.dates.mins_ago", "en", "{mins}m ago"},
		{"pages.common.dates.mins_ago", "es", "hace {mins}m"},
		{"pages.common.dates.hours_ago", "en", "{hours}h ago"},
		{"pages.common.dates.hours_ago", "es", "hace {hours}h"},

		// Product Table (Slugs: pages.admin.inventory.table)
		{"pages.admin.inventory.unassigned", "en", "unassigned"},
		{"pages.admin.inventory.unassigned", "es", "sin asignar"},
		{"pages.admin.inventory.delete_product_tooltip", "en", "Delete Product"},
		{"pages.admin.inventory.delete_product_tooltip", "es", "Eliminar Producto"},
		{"pages.admin.inventory.scanning_catalog", "en", "Scanning Catalog..."},
		{"pages.admin.inventory.scanning_catalog", "es", "Escaneando Catálogo..."},
		{"pages.admin.inventory.table.product", "en", "PRODUCT"},
		{"pages.admin.inventory.table.product", "es", "PRODUCTO"},
		{"pages.admin.inventory.table.type", "en", "TYPE"},
		{"pages.admin.inventory.table.type", "es", "TIPO"},
		{"pages.admin.inventory.table.rarity", "en", "RARITY"},
		{"pages.admin.inventory.table.rarity", "es", "RAREZA"},
		{"pages.admin.inventory.table.price", "en", "PRICE"},
		{"pages.admin.inventory.table.price", "es", "PRECIO"},
		{"pages.admin.inventory.table.stock", "en", "STOCK"},
		{"pages.admin.inventory.table.stock", "es", "STOCK"},
		{"pages.admin.inventory.table.updated", "en", "UPDATED"},
		{"pages.admin.inventory.table.updated", "es", "ACTUALIZADO"},
		{"pages.admin.inventory.table.cmd", "en", "CMD"},
		{"pages.admin.inventory.table.cmd", "es", "CMD"},
		{"pages.admin.inventory.no_products", "en", "NO PRODUCTS FOUND"},
		{"pages.admin.inventory.no_products", "es", "NO SE ENCONTRARON PRODUCTOS"},
		{"pages.admin.inventory.no_products_hint", "en", "Try adjusting your scanner filters"},
		{"pages.admin.inventory.no_products_hint", "es", "Intente ajustar sus filtros de escaneo"},

		// Profile (Slugs: profile)
		{"pages.profile.title", "en", "MY PROFILE"},
		{"pages.profile.title", "es", "MI PERFIL"},
		{"pages.profile.linked_accounts", "en", "LINKED ACCOUNTS"},
		{"pages.profile.linked_accounts", "es", "CUENTAS VINCULADAS"},
		{"pages.profile.order_history", "en", "ORDER HISTORY"},
		{"pages.profile.order_history", "es", "HISTORIAL DE PEDIDOS"},
		{"pages.profile.personal_info", "en", "PERSONAL INFORMATION"},
		{"pages.profile.personal_info", "es", "INFORMACIÓN PERSONAL"},
		{"pages.profile.no_orders", "en", "You haven't placed any orders yet."},
		{"pages.profile.no_orders", "es", "Aún no has realizado pedidos."},

		// Deck Importer (Slugs: pages.deck_importer)
		{"nav.deck_importer_cta", "en", "Try our deck importer"},
		{"nav.deck_importer_cta", "es", "Prueba nuestro importador de mazos"},
		{"pages.cart.drawer.try_importer", "en", "USE DECK IMPORTER →"},
		{"pages.cart.drawer.try_importer", "es", "USAR IMPORTADOR DE MAZOS →"},
		{"pages.cart.drawer.importer_nudge", "en", "Missing some cards? Try the deck importer"},
		{"pages.cart.drawer.importer_nudge", "es", "¿Te faltan algunas cartas? Prueba el importador de mazos"},
		{"pages.deck_importer.title", "en", "DECK IMPORTER"},
		{"pages.deck_importer.title", "es", "IMPORTADOR DE MAZOS"},
		{"pages.deck_importer.subtitle", "en", "Paste your card list below (e.g. \"4 Birds of Paradise\") and we'll match it against our current stock."},
		{"pages.deck_importer.subtitle", "es", "Pega tu lista de cartas abajo (ej. \"4 Birds of Paradise\") y la compararemos con nuestro stock actual."},
		{"pages.deck_importer.list_label", "en", "CARD LIST"},
		{"pages.deck_importer.list_label", "es", "LISTA DE CARTAS"},
		{"pages.deck_importer.buttons.analyze", "en", "ANALYZE LIST →"},
		{"pages.deck_importer.buttons.analyze", "es", "ANALIZAR LISTA →"},
		{"pages.deck_importer.buttons.analyzing", "en", "ANALYZING..."},
		{"pages.deck_importer.buttons.analyzing", "es", "ANALIZANDO..."},
		{"pages.deck_importer.results_title", "en", "MATCHES"},
		{"pages.deck_importer.results_title", "es", "COINCIDENCIAS"},
		{"pages.deck_importer.buttons.add_all", "en", "ADD ALL FOUND"},
		{"pages.deck_importer.buttons.add_all", "es", "AGREGAR TODO LO ENCONTRADO"},
		{"pages.deck_importer.no_results", "en", "Paste a list to see available cards."},
		{"pages.deck_importer.no_results", "es", "Pega una lista para ver las cartas disponibles."},
		{"pages.deck_importer.errors.analyze_failed", "en", "Failed to analyze list. Please check the format."},
		{"pages.deck_importer.errors.analyze_failed", "es", "Error al analizar la lista. Por favor verifica el formato."},
		{"pages.deck_importer.messages.added_count", "en", "Added products from {count} cards to your cart."},
		{"pages.deck_importer.messages.added_count", "es", "Se agregaron productos de {count} cartas a tu carrito."},
		{"pages.deck_importer.matches.different_version", "en", "DIFFERENT VERSION"},
		{"pages.deck_importer.matches.different_version", "es", "VERSIÓN DIFERENTE"},
		{"pages.deck_importer.matches.matched", "en", "MATCHED"},
		{"pages.deck_importer.matches.matched", "es", "ENCONTRADO"},
		{"pages.deck_importer.matches.no_stock", "en", "NO STOCK"},
		{"pages.deck_importer.matches.no_stock", "es", "SIN STOCK"},

		// Product Badges & Conditions
		{"pages.product.condition.nm", "en", "Near Mint"},
		{"pages.product.condition.nm", "es", "Near Mint"},
		{"pages.product.condition.lp", "en", "Lightly Played"},
		{"pages.product.condition.lp", "es", "Lightly Played"},
		{"pages.product.condition.mp", "en", "Moderately Played"},
		{"pages.product.condition.mp", "es", "Moderately Played"},
		{"pages.product.condition.hp", "en", "Heavily Played"},
		{"pages.product.condition.hp", "es", "Heavily Played"},
		{"pages.product.condition.dmg", "en", "Damaged"},
		{"pages.product.condition.dmg", "es", "Dañado"},
		{"pages.product.badges.hot", "en", "HOT"},
		{"pages.product.badges.hot", "es", "CALIENTE"},
		{"pages.product.badges.new", "en", "NEW"},
		{"pages.product.badges.new", "es", "NUEVO"},

		// Common Labels & Pagination
		{"pages.common.labels.home", "en", "HOME"},
		{"pages.common.labels.home", "es", "INICIO"},
		{"pages.common.labels.collection", "en", "COLLECTION"},
		{"pages.common.labels.collection", "es", "COLECCIÓN"},
		{"pages.common.labels.product", "en", "PRODUCT"},
		{"pages.common.labels.product", "es", "PRODUCTO"},
		{"pages.common.labels.products", "en", "PRODUCTS"},
		{"pages.common.labels.products", "es", "PRODUCTOS"},
		{"pages.common.pagination.prev", "en", "← PREV"},
		{"pages.common.pagination.prev", "es", "← ANTERIOR"},
		{"pages.common.pagination.next", "en", "NEXT →"},
		{"pages.common.pagination.next", "es", "SIGUIENTE →"},

		// Theme Selector
		{"components.theme_selector.title", "en", "Select Theme"},
		{"components.theme_selector.title", "es", "Seleccionar Tema"},

		// Checkout Delivery & Payment
		{"pages.checkout.delivery.shipping", "en", "SHIPPING"},
		{"pages.checkout.delivery.shipping", "es", "ENVÍO"},
		{"pages.checkout.delivery.pickup", "en", "LOCAL PICKUP"},
		{"pages.checkout.delivery.pickup", "es", "RECOGIDA LOCAL"},
		{"pages.checkout.delivery.shipping_desc", "en", "Reliable delivery to your address"},
		{"pages.checkout.delivery.shipping_desc", "es", "Envío confiable a tu dirección"},
		{"pages.checkout.delivery.pickup_desc", "en", "Pick up at our store in Bogotá"},
		{"pages.checkout.delivery.pickup_desc", "es", "Recoge en nuestra tienda en Bogotá"},
		{"pages.checkout.section.contact", "en", "CONTACT INFORMATION"},
		{"pages.checkout.section.contact", "es", "INFORMACIÓN DE CONTACTO"},
		{"pages.checkout.section.delivery", "en", "DELIVERY METHOD"},
		{"pages.checkout.section.delivery", "es", "MÉTODO DE ENTREGA"},
		{"pages.checkout.section.summary", "en", "ORDER SUMMARY"},
		{"pages.checkout.section.summary", "es", "RESUMEN DEL PEDIDO"},
		{"pages.checkout.summary.item", "en", "ITEM"},
		{"pages.checkout.summary.item", "es", "ARTÍCULO"},
		{"pages.checkout.summary.items", "en", "ITEMS"},
		{"pages.checkout.summary.items", "es", "ARTÍCULOS"},
		{"pages.checkout.summary.subtotal", "en", "SUBTOTAL"},
		{"pages.checkout.summary.subtotal", "es", "SUBTOTAL"},
		{"pages.checkout.summary.shipping", "en", "SHIPPING"},
		{"pages.checkout.summary.shipping", "es", "ENVÍO"},
		{"pages.checkout.summary.free", "en", "FREE"},
		{"pages.checkout.summary.free", "es", "GRATIS"},
		{"pages.checkout.total", "en", "TOTAL"},
		{"pages.checkout.total", "es", "TOTAL"},
		{"pages.checkout.footer.notice", "en", "Upon confirmation, an advisor will contact you to coordinate delivery."},
		{"pages.checkout.footer.notice", "es", "Al confirmar, un asesor te contactará para coordinar la entrega."},
		{"pages.checkout.payment_methods.cash", "en", "Cash"},
		{"pages.checkout.payment_methods.cash", "es", "Efectivo"},
		{"pages.checkout.payment_methods.transfer", "en", "Bank Transfer"},
		{"pages.checkout.payment_methods.transfer", "es", "Transferencia Bancaria"},
		{"pages.checkout.payment_methods.nequi", "en", "Nequi"},
		{"pages.checkout.payment_methods.nequi", "es", "Nequi"},
		{"pages.checkout.payment_methods.daviplata", "en", "Daviplata"},
		{"pages.checkout.payment_methods.daviplata", "es", "Daviplata"},

		// Bounty Offer Modal
		{"components.bounty_offer.title", "en", "Sell Us Your Card"},
		{"components.bounty_offer.title", "es", "Véndenos Tu Carta"},
		{"components.bounty_offer.success_title", "en", "Offer Received!"},
		{"components.bounty_offer.success_title", "es", "¡Oferta Recibida!"},
		{"components.bounty_offer.success_desc", "en", "We'll review it and contact you soon."},
		{"components.bounty_offer.success_desc", "es", "La revisaremos y te contactaremos pronto."},
		{"components.bounty_offer.login_prompt", "en", "to automatically fill your info and track your offers."},
		{"components.bounty_offer.login_prompt", "es", "para autocompletar tu info y seguir tus ofertas."},
		{"components.bounty_offer.pay_upto", "en", "We pay up to:"},
		{"components.bounty_offer.pay_upto", "es", "Pagamos hasta:"},
		{"components.bounty_offer.labels.name", "en", "Your Name"},
		{"components.bounty_offer.labels.name", "es", "Tu Nombre"},
		{"components.bounty_offer.labels.contact", "en", "Contact Info"},
		{"components.bounty_offer.labels.contact", "es", "Info de Contacto"},
		{"components.bounty_offer.labels.condition", "en", "Condition"},
		{"components.bounty_offer.labels.condition", "es", "Estado"},
		{"components.bounty_offer.labels.quantity", "en", "Quantity"},
		{"components.bounty_offer.labels.quantity", "es", "Cantidad"},
		{"components.bounty_offer.labels.notes", "en", "Additional Notes (Optional)"},
		{"components.bounty_offer.labels.notes", "es", "Notas Adicionales (Opcional)"},
		{"components.bounty_offer.placeholders.notes", "en", "Any details about the card..."},
		{"components.bounty_offer.placeholders.notes", "es", "Cualquier detalle sobre la carta..."},
		{"components.bounty_offer.placeholders.name", "en", "John Doe"},
		{"components.bounty_offer.placeholders.name", "es", "Juan Pérez"},
		{"components.bounty_offer.placeholders.contact", "en", "Phone number or Instagram handle"},
		{"components.bounty_offer.placeholders.contact", "es", "Teléfono o usuario de Instagram"},
		{"components.bounty_offer.buttons.submit", "en", "SUBMIT OFFER"},
		{"components.bounty_offer.buttons.submit", "es", "ENVIAR OFERTA"},

		// Store Exclusives
		{"pages.store_exclusives.title", "en", "STORE EXCLUSIVES"},
		{"pages.store_exclusives.title", "es", "EXCLUSIVOS DE LA TIENDA"},
		{"pages.store_exclusives.subtitle", "en", "Custom Commander decks, proxy kits, and other premium items crafted in-house."},
		{"pages.store_exclusives.subtitle", "es", "Mazos Commander personalizados, kits de proxies y otros artículos premium creados internamente."},

		// Collection Vault Landing
		{"pages.collection.vault.bg_text_1", "en", "MY VAULT"},
		{"pages.collection.vault.bg_text_1", "es", "MI BÓVEDA"},
		{"pages.collection.vault.bg_text_2", "en", "COLLECTION"},
		{"pages.collection.vault.bg_text_2", "es", "COLECCIÓN"},
		{"pages.collection.vault.title_part1", "en", "THE"},
		{"pages.collection.vault.title_part1", "es", "LA"},
		{"pages.collection.vault.title_part2", "en", "VAULT"},
		{"pages.collection.vault.title_part2", "es", "BÓVEDA"},
		{"pages.collection.vault.subject_prefix", "en", "Secure Digital Archive // Subject:"},
		{"pages.collection.vault.subject_prefix", "es", "Archivo Digital Seguro // Sujeto:"},
		{"pages.collection.vault.default_subject", "en", "COLLECTOR"},
		{"pages.collection.vault.default_subject", "es", "COLECCIONISTA"},
		{"pages.collection.vault.scanning", "en", "SCANNING INVENTORY..."},
		{"pages.collection.vault.scanning", "es", "ESCANEANDO INVENTARIO..."},
		{"pages.collection.vault.desc", "en", "Your personal TCG archive is currently being indexed. Soon you will be able to manage your physical collection and track market valuations here."},
		{"pages.collection.vault.desc", "es", "Tu archivo personal de TCG está siendo indexado. Pronto podrás gestionar tu colección física y seguir valoraciones de mercado aquí."},
		{"pages.collection.vault.return", "en", "Return to Command Center"},
		{"pages.collection.vault.return", "es", "Volver al Centro de Mando"},

		// Collection Grid
		{"pages.collection.not_found", "en", "COLLECTION NOT FOUND"},
		{"pages.collection.not_found", "es", "COLECCIÓN NO ENCONTRADA"},
		{"pages.collection.not_found_desc", "en", "The requested collection does not exist or has been removed."},
		{"pages.collection.not_found_desc", "es", "La colección solicitada no existe o ha sido eliminada."},
		{"pages.collection.empty", "en", "EMPTY COLLECTION"},
		{"pages.collection.empty", "es", "COLECCIÓN VACÍA"},
		{"pages.collection.empty_desc", "en", "No products found in this collection."},
		{"pages.collection.empty_desc", "es", "No se encontraron productos en esta colección."},

		// 404 Page
		{"pages.404.title", "en", "LINK SEVERED"},
		{"pages.404.title", "es", "ENLACE CORTADO"},
		{"pages.404.message", "en", "The requested coordinates lead to a void in the archive. Error Code: [PAGE_NOT_FOUND]."},
		{"pages.404.message", "es", "Las coordenadas solicitadas llevan a un vacío en el archivo. Código de Error: [PÁGINA_NO_ENCONTRADA]."},
		{"pages.404.back_home", "en", "← Return to Command Center"},
		{"pages.404.back_home", "es", "← Volver al Centro de Mando"},
	}

	for _, t := range data {
		_, err := db.Exec(`
			INSERT INTO translation (key, locale, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (key, locale) DO UPDATE SET value = EXCLUDED.value
		`, t.Key, t.Locale, t.Value)
		if err != nil {
			logger.Error("Failed to seed translation [key: %s, locale: %s]: %v", t.Key, t.Locale, err)
		}
	}
	logger.Info("✅ %d translation records seeded", len(data))
}
