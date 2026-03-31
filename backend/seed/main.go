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
		"customer_note",
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
	if user == "" { user = "admin" }
	if pass == "" { pass = "elbulk2024!" }

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
	cats := []struct{ Name, Slug string }{
		{"Featured", "featured"},
		{"Hot Items", "hot-items"},
		{"New Arrivals", "new-arrivals"},
		{"Sale", "sale"},
	}
	mapping := make(map[string]string)
	for _, cat := range cats {
		var id string
		db.QueryRow(`INSERT INTO custom_category (name, slug) VALUES ($1, $2) RETURNING id`, cat.Name, cat.Slug).Scan(&id)
		mapping[cat.Slug] = id
	}
	return mapping
}

func seedStorage(db *sqlx.DB) []string {
	locations := []string{"Showcase A", "Storage Box 1", "Binder Vault"}
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
		"contact_email":    "contact@el-bulk.com",
		"contact_phone":    "+57 300 123 4567",
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
		INSERT INTO product (name, tcg, category, set_name, set_code, price_source, price_cop_override, stock, image_url, color_identity)
		VALUES ($1, 'mtg', 'singles', 'Alpha', 'LEA', 'manual', 25000000, 1, 'https://cards.scryfall.io/normal/front/1/9/19911e6e-7c35-4281-b31c-266382f052cc.jpg?1717190810', 'C') RETURNING id
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
				full_art, textless, promo_type, cmc, color_identity
			) VALUES ($1, 'mtg', 'singles', $2, $3, $4, $5, $6, $7, $8, 'manual', $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
			RETURNING id
		`, 
		res.Name, res.SetName, res.SetCode, res.CollectorNumber, cond, f, t, res.Language, 
		price, res.ImageURL, stock, res.Rarity, 
		res.IsLegendary, res.IsHistoric, res.IsLand, res.IsBasicLand, 
		res.ArtVariation, res.OracleText, res.Artist, res.TypeLine, 
		res.BorderColor, res.Frame, res.FullArt, res.Textless, res.PromoType, 
		res.CMC, res.ColorIdentity)

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
	mtgSealed := []struct{ Name, Set, Img string; Price float64 }{
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
	pkmnItems := []struct{ Name, Cat, Set, Img string; Price float64 }{
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
	ygoItems := []struct{ Name, Cat, Set, Img string; Price float64 }{
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
	accs := []struct{ Name, Img string; Price float64 }{
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
		logger.Info("Looking up metadata for %d deck cards...", len(identifiers))
		var results []external.CardLookupResult
		for i := 0; i < len(identifiers); i += 75 {
			end := i + 75
			if end > len(identifiers) { end = len(identifiers) }
			res, err := external.BatchLookupMTGCard(identifiers[i:end])
			if err == nil { results = append(results, res...) }
		}

		for _, r := range results {
			db.Exec(`
				INSERT INTO deck_card (
					product_id, name, set_name, set_code, collector_number, quantity, 
					language, color_identity, cmc, is_legendary, is_historic, is_land, is_basic_land,
					art_variation, oracle_text, artist, type_line, border_color, frame,
					full_art, textless, promo_type, image_url, foil_treatment, card_treatment, rarity
				)
				VALUES ($1, $2, $3, $4, $5, 1, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, 'non_foil', 'normal', $23)
			`, 
			deckID, r.Name, r.SetName, r.SetCode, r.CollectorNumber, 
			r.Language, r.ColorIdentity, r.CMC, r.IsLegendary, r.IsHistoric, r.IsLand, r.IsBasicLand, 
			r.ArtVariation, r.OracleText, r.Artist, r.TypeLine, r.BorderColor, r.Frame, 
			r.FullArt, r.Textless, r.PromoType, r.ImageURL, r.Rarity)
		}
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

	// 7. Bounties
	logger.Info("Seeding bounties...")
	for i := 0; i < 15; i++ {
		db.Exec(`
			INSERT INTO bounty (name, tcg, set_name, condition, quantity_needed, target_price, is_active)
			VALUES ($1, 'mtg', 'Modern Horizons 3', 'NM', $2, $3, true)
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
		if i < 5 { numOrders = rand.Intn(10) + 10 } // Create some VIP customers with 10-20 orders
		for j := 0; j < numOrders; j++ {
			var oID string
			status := "completed"
			if rand.Intn(10) > 8 { status = "pending" }
			
			db.QueryRow(`
				INSERT INTO "order" (order_number, customer_id, status, payment_method, total_cop, created_at)
				VALUES ($1, $2, $3, 'Transfer', $4, $5) RETURNING id
			`, fmt.Sprintf("ORD-%d-%d", i, j), cID, status, 0.0, time.Now().AddDate(0, 0, -rand.Intn(30))).Scan(&oID)

			// Items
			total := 0.0
			for k := 0; k < rand.Intn(3)+1; k++ {
				if len(productIDs) == 0 { break }
				pID := productIDs[rand.Intn(len(productIDs))]
				var pName, pSet string
				var pPrice float64
				db.QueryRow("SELECT name, set_name, price_cop_override FROM product WHERE id = $1", pID).Scan(&pName, &pSet, &pPrice)
				
				qty := rand.Intn(2) + 1
				db.Exec(`
					INSERT INTO order_item (order_id, product_id, product_name, product_set, unit_price_cop, quantity)
					VALUES ($1, $2, $3, $4, $5, $6)
				`, oID, pID, pName, pSet, pPrice, qty)
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
	if len(productIDs) > 0 { p1Id = productIDs[0] }
	p2Id := "PROD-2"
	if len(productIDs) > 1 { p2Id = productIDs[1] }

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
			HTML: "<h2>Battle for the Top!</h2><p>Registration opens this Friday. Limited slots available. Format: Standard.</p>",
		},
		{
			Title: "Weekly Deal: Buy 3 Boosters, Get 1 Free!",
			Slug:  "weekly-deal-boosters",
			Img:   "https://images.unsplash.com/photo-1541560052753-107f96307409?q=80&w=800&auto=format&fit=crop",
			HTML: "<p>This week only, buy any 3 TCG boosters and get the 4th one free. Valid across MTG, Pokémon, and Yu-Gi-Oh.</p>",
		},
		{
			Title: "Trading Corner: Bulk Trade-in Event",
			Slug:  "bulk-trade-in-event",
			Img:   "https://images.unsplash.com/photo-1598214886806-c87b84b7078b?q=80&w=800&auto=format&fit=crop",
			HTML: "<p>Turn your bulk commons and uncommons into store credit! We're running a special trade-in event this Saturday starting at 10 AM.</p>",
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
			
			db.Exec(`
				INSERT INTO bounty_offer (bounty_id, customer_id, quantity, status, admin_notes)
				VALUES ($1, $2, $3, $4, $5)
			`, bID, c.ID, (rand.Intn(3) + 1), status, "Verified collection. Quality looks good.")
		}
	}

	// 5. Seed some guest requests (Unlinked)
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
		if i < 5 { noteChance = 1 } // Every top customer gets a note
		
		if i % noteChance == 0 {
			db.Exec(`
				INSERT INTO customer_note (customer_id, content, admin_id)
				VALUES ($1, $2, $3)
			`, c.ID, interactions[rand.Intn(len(interactions))], adminID)
		}
	}
}
