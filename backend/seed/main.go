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
	tcgIDs := seedTCGs(database)
	categoryMap := seedCategories(database)
	storageIDs := seedStorage(database)
	seedSettings(database)

	if *mode == "minimal" {
		seedMinimalData(database, tcgIDs, categoryMap, storageIDs)
	} else {
		productIDs := seedFullData(database, tcgIDs, categoryMap, storageIDs)
		seedNotices(database, productIDs)
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
		"tcg",
		"admin",
	}
	for _, t := range tables {
		db.Exec(fmt.Sprintf("DELETE FROM %s", t))
	}
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

func seedMinimalData(db *sqlx.DB, tcgIDs map[string]string, cats map[string]string, storageIDs []string) string {
	// 1 Sample Product
	name := "Black Lotus"
	var pID string
	db.QueryRow(`
		INSERT INTO product (name, tcg, category, set_name, set_code, price_source, price_cop_override, stock, image_url)
		VALUES ($1, 'mtg', 'singles', 'Alpha', 'LEA', 'manual', 25000000, 1, 'https://cards.scryfall.io/normal/front/1/9/19911e6e-7c35-4281-b31c-266382f052cc.jpg?1717190810') RETURNING id
	`, name).Scan(&pID)
	
	db.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 1)`, pID, storageIDs[0])
	db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`, pID, cats["featured"])
	return pID
}

func seedFullData(db *sqlx.DB, tcgIDs map[string]string, cats map[string]string, storageIDs []string) []string {
	// Seed 1 Minimal Product first as a safety baseline
	seedMinimalData(db, tcgIDs, cats, storageIDs)

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
	// Scryfall /cards/collection limit is 75 identifiers
	for i := 0; i < len(identifiers); i += 75 {
		end := i + 75
		if end > len(identifiers) {
			end = len(identifiers)
		}
		chunk := identifiers[i:end]
		res, err := external.BatchLookupMTGCard(chunk)
		if err != nil {
			logger.Error("Batch lookup chunk failed: %v", err)
			continue
		}
		results = append(results, res...)
		time.Sleep(100 * time.Millisecond) // Respect rate limits
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
				image_url, stock, rarity, is_legendary, oracle_text
			) VALUES ($1, 'mtg', 'singles', $2, $3, $4, $5, $6, $7, 'en', 'manual', $8, $9, $10, $11, $12, $13)
			RETURNING id
		`, res.Name, res.SetName, res.SetCode, res.CollectorNumber, cond, f, t, price, res.ImageURL, stock, res.Rarity, res.IsLegendary, res.OracleText)

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

	// 6. Bounties
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
	}

	for _, n := range notices {
		db.Exec(`
			INSERT INTO notice (title, slug, content_html, featured_image_url)
			VALUES ($1, $2, $3, $4)
		`, n.Title, n.Slug, n.HTML, n.Img)
	}
}
