package main

import (
	"os"
	"time"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	database := db.Connect()
	defer database.Close()

	// Clear products table
	logger.Info("Clearing products table...")
	_, err := database.Exec(`DELETE FROM product`)
	if err != nil {
		logger.Error("Failed to clear products: %v", err)
		os.Exit(1)
	}

	// Create admin user
	adminUser := os.Getenv("ADMIN_USERNAME")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminUser == "" {
		adminUser = "admin"
	}
	if adminPass == "" {
		adminPass = "elbulk2024!"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password: %v", err)
		os.Exit(1)
	}

	_, err = database.Exec(`
		INSERT INTO admin (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash
	`, adminUser, string(hash))
	if err != nil {
		logger.Error("Failed to create admin: %v", err)
		os.Exit(1)
	}
	logger.Info("Admin user '%s' created/updated", adminUser)

	// -------------------------------------------------------------------------
	// PRODUCT SEEDING
	// -------------------------------------------------------------------------

	type seedProduct struct {
		Name          string
		TCG           string
		Category      string
		SetName       string
		SetCode       string
		Condition     string
		PriceCOP      float64
		Stock         int
		CategorySlugs []string
		Description   string
		ImageURL      string
		Foil          models.FoilTreatment
		Treatment     models.CardTreatment
		Language      string
		FullArt       bool
		Textless      bool
		BorderColor   string
		Frame         string
	}

	products := []seedProduct{
		// Sheoldred, the Apocalypse
		{"Sheoldred, the Apocalypse", "mtg", "singles", "Dominaria United", "DMU", "NM", 280000, 10, []string{"featured", "hot-items"}, "The Praetor of the Apocalypse.", "", models.FoilNonFoil, models.TreatmentNormal, "en", false, false, "", ""},
		{"Sheoldred, the Apocalypse", "mtg", "singles", "Dominaria United", "DMU", "NM", 350000, 2, []string{"hot-items"}, "Showcase variant.", "", models.FoilFoil, models.TreatmentShowcase, "en", false, false, "", ""},
		
		// The One Ring
		{"The One Ring", "mtg", "singles", "The Lord of the Rings: Tales of Middle-earth", "LTR", "NM", 450000, 5, []string{"hot-items"}, "Main set version.", "", models.FoilNonFoil, models.TreatmentNormal, "en", false, false, "", ""},
		{"The One Ring", "mtg", "singles", "Tales of Middle-earth Bundle", "LTR", "NM", 550000, 2, []string{"featured"}, "Bundle alternate art.", "", models.FoilFoil, models.TreatmentAlternateArt, "en", false, false, "", ""},
		{"The One Ring", "mtg", "singles", "LTR Extension", "LTR", "NM", 850000, 1, []string{"hot-items"}, "Borderless Poster variant.", "", models.FoilFoil, models.TreatmentBorderless, "en", true, false, "", ""},

		// Mana Crypt
		{"Mana Crypt", "mtg", "singles", "Special Guests", "SPG", "NM", 780000, 1, []string{"featured", "hot-items"}, "Special Guest version.", "", models.FoilNonFoil, models.TreatmentNormal, "en", false, false, "", ""},
		{"Mana Crypt", "mtg", "singles", "Double Masters", "2XM", "NM", 950000, 1, []string{"hot-items"}, "Borderless variant.", "", models.FoilFoil, models.TreatmentBorderless, "en", true, false, "", ""},

		// Specialty
		{"Lightning Bolt", "mtg", "singles", "Judge Gift Cards", "PPRO", "NM", 250000, 2, []string{"featured"}, "Judge Promo Bolt.", "", models.FoilFoil, models.TreatmentPromo, "en", false, false, "", ""},
		{"Force of Will", "mtg", "singles", "Judge Gift Cards", "PPRO", "NM", 1200000, 1, []string{"hot-items"}, "Judge Promo Force.", "", models.FoilFoil, models.TreatmentPromo, "en", false, false, "", ""},
		{"Cryptic Command", "mtg", "singles", "Player Rewards 2009", "P09", "NM", 150000, 2, []string{"featured"}, "Textless Player Reward.", "", models.FoilFoil, models.TreatmentTextless, "en", true, true, "", ""},
		{"Damnation", "mtg", "singles", "Player Rewards 2008", "P08", "NM", 180000, 1, []string{"hot-items"}, "Textless Player Reward.", "", models.FoilFoil, models.TreatmentTextless, "en", true, true, "", ""},

		// Languages
		{"Snapcaster Mage", "mtg", "singles", "Innistrad", "ISD", "NM", 120000, 4, []string{"featured"}, "Japanese variant.", "", models.FoilFoil, models.TreatmentNormal, "jp", false, false, "", ""},
		{"Thoughtseize", "mtg", "singles", "Theros", "THS", "LP", 55000, 8, []string{"sale"}, "Spanish variant.", "", models.FoilNonFoil, models.TreatmentNormal, "es", false, false, "", ""},
		{"Tarmogoyf", "mtg", "singles", "Future Sight", "FUT", "NM", 85000, 2, []string{"featured"}, "German variant.", "", models.FoilNonFoil, models.TreatmentNormal, "de", false, false, "", ""},

		// Retro
		{"Thalia, Guardian of Thraben", "mtg", "singles", "Modern Horizons 2", "MH2", "NM", 45000, 12, []string{"sale"}, "Retro Frame variant.", "", models.FoilNonFoil, models.TreatmentLegacyBorder, "en", false, false, "black", "old"},
		{"Counterspell", "mtg", "singles", "Alpha", "LEA", "NM", 4500000, 1, []string{"hot-items"}, "Original Old Frame.", "", models.FoilNonFoil, models.TreatmentNormal, "en", false, false, "black", "old"},

		// Exotic
		{"Sol Ring", "mtg", "singles", "Warhammer 40,000", "40K", "NM", 125000, 3, []string{"featured"}, "Surge Foil Sol Ring.", "", models.FoilSurgeFoil, models.TreatmentNormal, "en", false, false, "black", "2015"},
		{"Elesh Norn, Mother of Machines", "mtg", "singles", "Phyrexia: All Will Be One", "ONE", "NM", 850000, 1, []string{"featured", "hot-items"}, "Oil Slick Raised Foil.", "", models.FoilOilSlick, models.TreatmentBorderless, "en", true, false, "black", "2015"},
		{"Mondrak, Glory Dominus", "mtg", "singles", "Phyrexia: All Will Be One", "ONE", "NM", 450000, 1, []string{"featured"}, "Step-and-Compleat Foil.", "", models.FoilStepAndCompleat, models.TreatmentShowcase, "en", false, false, "black", "2015"},
		{"Ragavan, Nimble Pilferer", "mtg", "singles", "Multiverse Legends", "MUL", "NM", 9500000, 1, []string{"hot-items"}, "Serialized 001/500.", "", models.FoilDoubleRainbow, models.TreatmentSerialized, "en", false, false, "black", "2015"},
		{"Brazen Borrower", "mtg", "singles", "Modern Horizons 2", "MH2", "NM", 85000, 5, []string{"sale"}, "Etched Foil Showcase.", "", models.FoilEtchedFoil, models.TreatmentShowcase, "en", false, false, "black", "2015"},
		{"Liliana of the Veil", "mtg", "singles", "Double Masters 2022", "2X2", "NM", 1500000, 1, []string{"hot-items"}, "Textured Foil variant.", "", models.FoilTexturedFoil, models.TreatmentBorderless, "en", true, false, "black", "2015"},
		{"Hidetsugu, Devouring Chaos", "mtg", "singles", "Kamigawa: Neon Dynasty", "NEO", "NM", 5500000, 1, []string{"hot-items"}, "Neon Ink Red Foil.", "", models.FoilNeonInk, models.TreatmentNormal, "en", false, false, "black", "2015"},
		{"Smothering Tithe", "mtg", "singles", "Wilds of Eldraine Anime", "WOT", "NM", 650000, 1, []string{"featured"}, "Confetti Foil Anime Art.", "", models.FoilConfettiFoil, models.TreatmentAlternateArt, "en", false, false, "black", "2015"},
		{"The Meathook Massacre", "mtg", "singles", "Innistrad: Midnight Hunt", "MID", "NM", 185000, 2, []string{"hot-items"}, "Extended Art variant.", "", models.FoilNonFoil, models.TreatmentExtendedArt, "en", false, false, "black", "2015"},
		{"Comet, Stellar Pup", "mtg", "singles", "Unfinity", "UNF", "NM", 120000, 4, []string{"featured"}, "Galaxy Foil variant.", "", models.FoilGalaxyFoil, models.TreatmentNormal, "en", false, false, "black", "2015"},

		// SEALED
		{"MTG Duskmourn Box", "mtg", "sealed", "Duskmourn", "DSK", "NM", 680000, 5, []string{"featured", "new-arrivals"}, "36 Play Boosters.", "https://media.wizards.com/2024/wpn/Duskmourn_PlayBoosterBox.png", models.FoilNonFoil, models.TreatmentNormal, "en", false, false, "black", "2015"},

		// ACCESSORIES
		{"Dragon Shield Pink", "accessories", "accessories", "N/A", "N/A", "NM", 55000, 25, []string{"sale"}, "100 Matte sleeves.", "https://m.media-amazon.com/images/I/71Q+B+B+L+L._AC_SL1500_.jpg", models.FoilNonFoil, models.TreatmentNormal, "en", false, false, "black", "N/A"},
	}

	// Seed TCGs
	tcgsToSeed := []struct{ ID, Name string }{
		{"mtg", "Magic: The Gathering"},
		{"pokemon", "Pokémon"},
		{"lorcana", "Disney Lorcana"},
		{"onepiece", "One Piece"},
		{"yugioh", "Yu-Gi-Oh!"},
		{"starwars", "Star Wars Unlimited"},
		{"weiss", "Weiss Schwarz"},
	}
	for _, t := range tcgsToSeed {
		database.Exec(`INSERT INTO tcg (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`, t.ID, t.Name)
	}

	// Seed Categories
	categoryMap := make(map[string]string)
	categoriesToSeed := []struct{ Name, Slug string }{
		{"Featured", "featured"},
		{"Sale", "sale"},
		{"New Arrivals", "new-arrivals"},
		{"Hot Items", "hot-items"},
	}
	for _, cat := range categoriesToSeed {
		var catID string
		database.QueryRow(`INSERT INTO custom_category (name, slug) VALUES ($1, $2) ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name RETURNING id`, cat.Name, cat.Slug).Scan(&catID)
		categoryMap[cat.Slug] = catID
	}

	// Seed Storage
	var binderID, boxA_ID string
	database.QueryRow(`INSERT INTO storage_location (name) VALUES ('Binder 1') ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`).Scan(&binderID)
	database.QueryRow(`INSERT INTO storage_location (name) VALUES ('Box A') ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`).Scan(&boxA_ID)

	for _, p := range products {
		logger.Info("Seeding %s (%s)...", p.Name, p.TCG)

		desc := p.Description
		img := p.ImageURL
		lang := p.Language
		color, rarity := "", ""
		var cmc float64 = 0
		isLegendary, isHistoric, isLand, isBasicLand := false, false, false, false
		artVariation, oracleText, artist, typeLine := "", "", "", ""
		borderColor, frame := p.BorderColor, p.Frame
		fullArt, textless := p.FullArt, p.Textless
		setCode, setName := p.SetCode, p.SetName

		if p.TCG == "mtg" {
			meta, err := external.LookupMTGCard(p.Name, p.SetCode, "", string(p.Foil))
			if err == nil {
				if desc == "" && meta.OracleText != nil { desc = *meta.OracleText }
				if img == "" { img = meta.ImageURL }
				if lang == "" { lang = meta.Language }
				if meta.ColorIdentity != nil { color = *meta.ColorIdentity }
				if meta.Rarity != nil { rarity = *meta.Rarity }
				if meta.CMC != nil { cmc = *meta.CMC }
				isLegendary = meta.IsLegendary
				isHistoric = meta.IsHistoric
				isLand = meta.IsLand
				isBasicLand = meta.IsBasicLand
				if meta.ArtVariation != nil { artVariation = *meta.ArtVariation }
				if setCode == "" || setCode == "N/A" { setCode = meta.SetCode }
				if setName == "" || setName == "N/A" { setName = meta.SetName }
				if meta.OracleText != nil { oracleText = *meta.OracleText }
				if meta.Artist != nil { artist = *meta.Artist }
				if meta.TypeLine != nil { typeLine = *meta.TypeLine }
				if meta.BorderColor != nil { borderColor = *meta.BorderColor }
				if meta.Frame != nil { frame = *meta.Frame }
				fullArt = meta.FullArt
				textless = meta.Textless
			}
		}

		var newProductID string
		err = database.QueryRow(`
			INSERT INTO product (
				name, tcg, category, set_name, set_code, condition,
				foil_treatment, card_treatment,
				price_cop_override, price_source, stock, description, image_url,
				language, color_identity, rarity, cmc, is_legendary, is_historic, is_land, is_basic_land, art_variation,
				oracle_text, artist, type_line, border_color, frame, full_art, textless
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'manual', 0, $10, $11, 
			          $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27) RETURNING id
		`, p.Name, p.TCG, p.Category, setName, setCode, p.Condition,
			p.Foil, p.Treatment,
			p.PriceCOP, desc, img,
			lang, color, rarity, cmc, isLegendary, isHistoric, isLand, isBasicLand, artVariation,
			oracleText, artist, typeLine, borderColor, frame, fullArt, textless).Scan(&newProductID)
		
		if err != nil {
			logger.Warn("failed to insert %s: %v", p.Name, err)
			continue
		}

		for _, slug := range p.CategorySlugs {
			if id, ok := categoryMap[slug]; ok {
				database.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, newProductID, id)
			}
		}

		if p.Stock > 0 {
			database.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, $3)`, newProductID, binderID, p.Stock/2)
			database.Exec(`INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, $3)`, newProductID, boxA_ID, p.Stock-(p.Stock/2))
		}
		time.Sleep(100 * time.Millisecond) // Respect rate limits
	}

	logger.Info("Seed complete!")
}
