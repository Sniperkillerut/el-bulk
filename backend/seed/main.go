package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
	clear := flag.Bool("clear", false, "Clear existing tables before seeding")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	database := db.Connect()
	defer database.Close()

	// Check for Cloud Run task index
	taskIDStr := os.Getenv("CLOUD_RUN_TASK_INDEX")
	if taskIDStr != "" {
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil {
			logger.Error("❌ Invalid CLOUD_RUN_TASK_INDEX: %v", err)
			os.Exit(1)
		}
		if err := runTask(database, taskID, *mode, *env, *clear); err != nil {
			logger.Error("❌ Task %d failed: %v", taskID, err)
			os.Exit(1)
		}
		logger.Info("🏅 Task %d completed successfully!", taskID)
		return
	}

	// Default behavior (local/Legacy Mode)
	if err := runSeed(database, *mode, *env, *clear); err != nil {
		logger.Error("❌ Seeding failed: %v", err)
		os.Exit(1)
	}

	logger.Info("🏆  El Bulk Seed successfully completed!")
}

// runTask executes a specific seeding step based on the task index.
func runTask(database *sqlx.DB, index int, mode, env string, clear bool) error {
	logger.Info("🚀 Running Seeding Task #%d (Mode: %s, Env: %s)", index, mode, env)

	// Dependency mapping & Wait logic
	if index > 0 {
		if err := waitForDependencies(database, index); err != nil {
			return err
		}
	}

	switch index {
	case 0:
		if clear {
			return clearTables(database)
		}
		logger.Info("⏩ Task 0: Skipping clearTables (clear=false)")
		return nil
	case 1:
		_, err := seedAdmin(database)
		return err
	case 2:
		return seedTCGs(database)
	case 3:
		_, err := seedCategories(database)
		return err
	case 4:
		_, err := seedStorage(database)
		return err
	case 5:
		return seedSettings(database)
	case 6:
		return seedThemes(database)
	case 7:
		return seedTranslations(database)
	case 8:
		return seedNotices(database)
	case 9:
		seedSets(database)
		return nil
	case 10:
		if mode != "minimal" {
			logger.Info("⏩ Task 10: Skipping Minimal Product (mode != minimal)")
			return nil
		}
		cats, stor, _, err := loadSeedResources(database)
		if err != nil {
			return err
		}
		return seedMinimalProduct(database, cats, stor)
	case 11:
		if mode != "full" {
			logger.Info("⏩ Task 11: Skipping MTG Singles (mode != full)")
			return nil
		}
		cats, stor, _, err := loadSeedResources(database)
		if err != nil {
			return err
		}
		_, err = seedMTGSingles(database, cats, stor)
		return err
	case 12:
		if mode != "full" {
			logger.Info("⏩ Task 12: Skipping MTG Sealed (mode != full)")
			return nil
		}
		cats, stor, _, err := loadSeedResources(database)
		if err != nil {
			return err
		}
		_, err = seedMTGSealed(database, cats, stor)
		return err
	case 13:
		if mode != "full" {
			logger.Info("⏩ Task 13: Skipping Multi-TCG (mode != full)")
			return nil
		}
		cats, stor, _, err := loadSeedResources(database)
		if err != nil {
			return err
		}
		_, err = seedMultiTCGProducts(database, cats, stor)
		return err
	case 14:
		if mode != "full" {
			logger.Info("⏩ Task 14: Skipping Accessories (mode != full)")
			return nil
		}
		cats, stor, _, err := loadSeedResources(database)
		if err != nil {
			return err
		}
		_, err = seedAccessories(database, cats, stor)
		return err
	case 15:
		if mode != "full" {
			logger.Info("⏩ Task 15: Skipping Store Exclusives (mode != full)")
			return nil
		}
		cats, stor, _, err := loadSeedResources(database)
		if err != nil {
			return err
		}
		_, err = seedStoreExclusives(database, cats, stor)
		return err
	case 16:
		if mode != "full" {
			logger.Info("⏩ Task 16: Skipping Bounties (mode != full)")
			return nil
		}
		_, err := seedBounties(database)
		return err
	case 17:
		if mode != "full" {
			logger.Info("⏩ Task 17: Skipping Customers (mode != full)")
			return nil
		}
		_, err := seedCustomers(database)
		return err
	case 18:
		if mode != "full" {
			logger.Info("⏩ Task 18: Skipping Orders (mode != full)")
			return nil
		}
		cats, _, _, err := loadSeedResources(database) // Loading cats to check if available
		if err != nil {
			return err
		}
		_ = cats
		// For orders we need both customers and products IDs.
		// Since we split tasks, we load them from DB.
		customers, err := loadSeededCustomers(database)
		if err != nil {
			return err
		}
		productIDs, err := loadSeededProductIDs(database)
		if err != nil {
			return err
		}
		return seedOrders(database, customers, productIDs)
	case 19:
		if mode != "full" {
			logger.Info("⏩ Task 19: Skipping CRM (mode != full)")
			return nil
		}
		_, _, adminID, err := loadSeedResources(database)
		if err != nil {
			return err
		}
		customers, err := loadSeededCustomers(database)
		if err != nil {
			return err
		}
		bountyIDs, err := loadSeededBountyIDs(database)
		if err != nil {
			return err
		}
		return seedCRM(database, adminID, customers, bountyIDs)
	default:
		logger.Warn("⚠️ Unknown Task Index %d — doing nothing", index)
		return nil
	}
}

// waitForDependencies polls the database to ensure previous phases are ready.
func waitForDependencies(database *sqlx.DB, index int) error {
	maxWait := 60 * time.Second
	start := time.Now()

	for {
		if time.Since(start) > maxWait {
			return fmt.Errorf("timeout waiting for dependencies for task %d", index)
		}

		ready := false
		if index >= 1 && index <= 9 {
			// Configuration tasks depend on nothing specific other than DB connection
			ready = true
		} else if index >= 10 && index <= 15 {
			// Product tasks need Categories and Storage
			var catCount, storCount int
			database.Get(&catCount, "SELECT COUNT(*) FROM custom_category")
			database.Get(&storCount, "SELECT COUNT(*) FROM storage_location")
			if catCount >= 5 && storCount >= 5 {
				ready = true
			}
		} else if index >= 16 && index <= 19 {
			// CRM/Orders need Admin and potentially Products
			var adminCount int
			database.Get(&adminCount, "SELECT COUNT(*) FROM admin")
			if adminCount > 0 {
				ready = true
			}
		}

		if ready {
			return nil
		}

		logger.Info("⏳ Task %d: Waiting for prerequisites...", index)
		time.Sleep(5 * time.Second)
	}
}

// loadSeedResources fetches existing categories, storage locations and admin ID.
func loadSeedResources(db *sqlx.DB) (CategoryMap, StorageMap, string, error) {
	cats, err := loadCategories(db)
	if err != nil {
		return nil, nil, "", fmt.Errorf("loadCategories: %w", err)
	}
	stor, err := loadStorage(db)
	if err != nil {
		return nil, nil, "", fmt.Errorf("loadStorage: %w", err)
	}
	var adminID string
	err = db.Get(&adminID, "SELECT id FROM admin LIMIT 1")
	if err != nil {
		return nil, nil, "", fmt.Errorf("get adminID: %w", err)
	}
	return cats, stor, adminID, nil
}

// Additional helpers for task-based data retrieval ───────────────────────────

func loadSeededCustomers(db *sqlx.DB) ([]CustomerSeeded, error) {
	var result []CustomerSeeded
	err := db.Select(&result, "SELECT id, first_name, last_name, email, phone FROM customer")
	return result, err
}

func loadSeededProductIDs(db *sqlx.DB) ([]string, error) {
	var result []string
	err := db.Select(&result, "SELECT id FROM product")
	return result, err
}

func loadSeededBountyIDs(db *sqlx.DB) ([]string, error) {
	var result []string
	err := db.Select(&result, "SELECT id FROM bounty")
	return result, err
}

func runSeed(database *sqlx.DB, mode, env string, clear bool) error {
	// ── Banner ───────────────────────────────────────────────────────────────
	if mode == "full" {
		logger.Info("🌟 El Bulk Seed — FULL mode (%s env) starting...", env)
		logger.Info("   This fetches live Scryfall data and may take 2-5 minutes.")
	} else {
		logger.Info("🌱 El Bulk Seed — MINIMAL mode (%s env) starting...", env)
	}

	// ── Clear existing data ───────────────────────────────────────────────────
	if clear {
		if err := clearTables(database); err != nil {
			return fmt.Errorf("failed to clear tables: %w", err)
		}
	} else {
		logger.Info("⏩  Skipping table clearing (run with --clear=true to wipe data)")
	}

	// ── Configuration seed (runs in ALL modes) ───────────────────────────────
	logger.Info("--- Phase 1: Configuration ---")
	adminID, err := seedAdmin(database)
	if err != nil {
		return fmt.Errorf("seedAdmin failed: %w", err)
	}

	if err := seedTCGs(database); err != nil {
		return fmt.Errorf("seedTCGs failed: %w", err)
	}

	cats, err := seedCategories(database)
	if err != nil {
		return fmt.Errorf("seedCategories failed: %w", err)
	}

	storage, err := seedStorage(database)
	if err != nil {
		return fmt.Errorf("seedStorage failed: %w", err)
	}

	if err := seedSettings(database); err != nil {
		return fmt.Errorf("seedSettings failed: %w", err)
	}

	if err := seedThemes(database); err != nil {
		return fmt.Errorf("seedThemes failed: %w", err)
	}

	if err := seedTranslations(database); err != nil {
		return fmt.Errorf("seedTranslations failed: %w", err)
	}

	if err := seedNotices(database); err != nil {
		return fmt.Errorf("seedNotices failed: %w", err)
	}

	seedSets(database) // Sync MTG sets (network required, non-fatal)

	if mode == "minimal" {
		if err := seedMinimalProduct(database, cats, storage); err != nil {
			return fmt.Errorf("seedMinimalProduct failed: %w", err)
		}
		logger.Info("✅ Minimal seeding complete! Admin: %s", adminID)
		return nil
	}

	// ── Full data seed ────────────────────────────────────────────────────────
	logger.Info("--- Phase 2: Products & Inventory ---")
	var allProductIDs []string

	mtgSingleIDs, err := seedMTGSingles(database, cats, storage)
	if err != nil {
		return fmt.Errorf("seed MTG singles failed: %w", err)
	}
	allProductIDs = append(allProductIDs, mtgSingleIDs...)

	mtgSealedIDs, err := seedMTGSealed(database, cats, storage)
	if err != nil {
		return fmt.Errorf("seed MTG sealed failed: %w", err)
	}
	allProductIDs = append(allProductIDs, mtgSealedIDs...)

	multiIDs, err := seedMultiTCGProducts(database, cats, storage)
	if err != nil {
		return fmt.Errorf("seed multi-TCG products failed: %w", err)
	}
	allProductIDs = append(allProductIDs, multiIDs...)

	accIDs, err := seedAccessories(database, cats, storage)
	if err != nil {
		return fmt.Errorf("seed accessories failed: %w", err)
	}
	allProductIDs = append(allProductIDs, accIDs...)

	exclusiveIDs, err := seedStoreExclusives(database, cats, storage)
	if err != nil {
		return fmt.Errorf("seed store exclusives failed: %w", err)
	}
	allProductIDs = append(allProductIDs, exclusiveIDs...)

	logger.Info("--- Phase 3: CRM & Commerce ---")
	bountyIDs, err := seedBounties(database)
	if err != nil {
		return fmt.Errorf("seed bounties failed: %w", err)
	}

	customers, err := seedCustomers(database)
	if err != nil {
		return fmt.Errorf("seed customers failed: %w", err)
	}

	if err := seedOrders(database, customers, allProductIDs); err != nil {
		return fmt.Errorf("seed orders failed: %w", err)
	}

	if err := seedCRM(database, adminID, customers, bountyIDs); err != nil {
		return fmt.Errorf("seed CRM failed: %w", err)
	}

	// ── Final summary ─────────────────────────────────────────────────────────
	var productCount, customerCount, orderCount, bountyCount int
	database.Get(&productCount, "SELECT COUNT(*) FROM product")
	database.Get(&customerCount, "SELECT COUNT(*) FROM customer")
	database.Get(&orderCount, `SELECT COUNT(*) FROM "order"`)
	database.Get(&bountyCount, "SELECT COUNT(*) FROM bounty")

	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("✅  El Bulk Seed Summary:")
	logger.Info("    Products  : %d", productCount)
	logger.Info("    Customers : %d", customerCount)
	logger.Info("    Orders    : %d", orderCount)
	logger.Info("    Bounties  : %d", bountyCount)
	logger.Info("    Admin     : %s", adminID)
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// clearTables truncates all data tables respecting FK order.
func clearTables(db *sqlx.DB) error {
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
			return fmt.Errorf("could not clear %s: %w", t, err)
		}
	}
	// tcg and tcg_set are left/overwritten with ON CONFLICT DO UPDATE
	logger.Info("  ✅ Tables cleared")
	return nil
}

// seedMinimalProduct inserts one reference product for production/minimal mode.
func seedMinimalProduct(database *sqlx.DB, cats CategoryMap, storage StorageMap) error {
	logger.Info("🌱 Inserting reference product (Black Lotus)...")
	var pID string
	legalities := `{"commander":"banned","legacy":"banned","vintage":"restricted","oldschool":"restricted"}`
	err := database.QueryRow(`
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
	`, legalities).Scan(&pID)
	if err != nil {
		return fmt.Errorf("failed to insert reference product: %w", err)
	}

	if sid, ok := storage["Showcase A"]; ok {
		if _, err := database.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, 1) ON CONFLICT DO NOTHING`, pID, sid); err != nil {
			return fmt.Errorf("failed to insert reference product storage: %w", err)
		}
	}
	if catID, ok := cats["featured"]; ok {
		if _, err := database.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID); err != nil {
			return fmt.Errorf("failed to insert reference product category: %w", err)
		}
	}
	logger.Info("  ✅ Reference product created: %s", pID)
	return nil
}
