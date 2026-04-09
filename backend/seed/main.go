package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func main() {
	// ── CLI flags ────────────────────────────────────────────────────────────
	mode := flag.String("mode", "minimal",
		"Seeding mode: 'minimal' (config + 1 product) or 'full' (hundreds of records)")
	env := flag.String("env", "development",
		"Deployment environment: 'development' or 'production'")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	database := db.Connect()
	defer database.Close()

	// ── Banner ───────────────────────────────────────────────────────────────
	if *mode == "full" {
		logger.Info("🌟 El Bulk Seed — FULL mode (%s env) starting...", *env)
		logger.Info("   This fetches live Scryfall data and may take 2-5 minutes.")
	} else {
		logger.Info("🌱 El Bulk Seed — MINIMAL mode (%s env) starting...", *env)
	}

	// ── Clear existing data ───────────────────────────────────────────────────
	clearTables(database)

	// ── Configuration seed (runs in ALL modes) ───────────────────────────────
	logger.Info("--- Phase 1: Configuration ---")
	adminID := seedAdmin(database)
	seedTCGs(database)
	cats := seedCategories(database)
	storage := seedStorage(database)
	seedSettings(database)
	seedThemes(database)
	seedTranslations(database)
	seedNotices(database) // Blog posts are config-adjacent, always seed
	seedSets(database)    // Sync MTG sets (network required, non-fatal)

	if *mode == "minimal" {
		seedMinimalProduct(database, cats, storage)
		logger.Info("✅ Minimal seeding complete! Admin: %s", adminID)
		return
	}

	// ── Full data seed ────────────────────────────────────────────────────────
	logger.Info("--- Phase 2: Products & Inventory ---")
	var allProductIDs []string

	mtgSingleIDs := seedMTGSingles(database, cats, storage)
	allProductIDs = append(allProductIDs, mtgSingleIDs...)

	mtgSealedIDs := seedMTGSealed(database, cats, storage)
	allProductIDs = append(allProductIDs, mtgSealedIDs...)

	multiIDs := seedMultiTCGProducts(database, cats, storage)
	allProductIDs = append(allProductIDs, multiIDs...)

	accIDs := seedAccessories(database, cats, storage)
	allProductIDs = append(allProductIDs, accIDs...)

	exclusiveIDs := seedStoreExclusives(database, cats, storage)
	allProductIDs = append(allProductIDs, exclusiveIDs...)

	logger.Info("--- Phase 3: CRM & Commerce ---")
	bountyIDs := seedBounties(database)
	customers := seedCustomers(database)
	seedOrders(database, customers, allProductIDs)
	seedCRM(database, adminID, customers, bountyIDs)

	// ── Final summary ─────────────────────────────────────────────────────────
	var productCount, customerCount, orderCount, bountyCount int
	database.Get(&productCount, "SELECT COUNT(*) FROM product")
	database.Get(&customerCount, "SELECT COUNT(*) FROM customer")
	database.Get(&orderCount, `SELECT COUNT(*) FROM "order"`)
	database.Get(&bountyCount, "SELECT COUNT(*) FROM bounty")

	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("✅  El Bulk Seed Complete!")
	logger.Info("    Products  : %d", productCount)
	logger.Info("    Customers : %d", customerCount)
	logger.Info("    Orders    : %d", orderCount)
	logger.Info("    Bounties  : %d", bountyCount)
	logger.Info("    Admin     : %s", adminID)
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// clearTables truncates all data tables respecting FK order.
func clearTables(db *sqlx.DB) {
	logger.Info("🗑️  Clearing existing data...")
	tables := []string{
		"deck_card",
		"order_item",
		`"order"`,
		"bounty_offer",
		"bounty",
		"client_request",
		"customer_note",
		"newsletter_subscriber",
		"customer_auth",
		"customer",
		"product_category",
		"product_storage",
		"product",
		"notice",
		"custom_category",
		"storage_location",
		"admin",
	}
	for _, t := range tables {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", t)); err != nil {
			logger.Warn("  Could not clear %s: %v", t, err)
		}
	}
	// tcg and tcg_set are left/overwritten with ON CONFLICT DO UPDATE
	logger.Info("  ✅ Tables cleared")
}

// seedMinimalProduct inserts one reference product for production/minimal mode.
func seedMinimalProduct(database *sqlx.DB, cats CategoryMap, storage StorageMap) {
	logger.Info("🌱 Inserting reference product (Black Lotus)...")
	var pID string
	legalities := `{"commander":"banned","legacy":"banned","vintage":"restricted","oldschool":"restricted"}`
	if err := database.QueryRow(`
		INSERT INTO product (
			name, tcg, category, set_name, set_code,
			price_source, price_cop_override, stock,
			image_url, color_identity, oracle_text, legalities,
			rarity, cost_basis_cop, created_at
		) VALUES (
			'Black Lotus','mtg','singles','Limited Edition Alpha','lea',
			'manual',25000000,1,
			'https://cards.scryfall.io/normal/front/1/9/19911e6e-7c35-4281-b31c-266382f052cc.jpg?1717190810',
			'C','{T}, Sacrifice Black Lotus: Add three mana of any one color.',
			$1::jsonb,
			'rare',15000000,now()
		) RETURNING id
	`, legalities).Scan(&pID); err != nil {
		logger.Error("Failed to insert reference product: %v", err)
		return
	}
	if sid, ok := storage["Showcase A"]; ok {
		database.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 1) ON CONFLICT DO NOTHING`, pID, sid)
	}
	if catID, ok := cats["featured"]; ok {
		database.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
	}
	logger.Info("  ✅ Reference product created: %s", pID)
}
