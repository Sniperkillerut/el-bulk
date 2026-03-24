package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/models"
	"golang.org/x/crypto/bcrypt"
)

type ScryfallResponse struct {
	OracleText string `json:"oracle_text"`
	FlavorText string `json:"flavor_text"`
	ImageUris  struct {
		Normal string `json:"normal"`
	} `json:"image_uris"`
	CardFaces []struct {
		OracleText string `json:"oracle_text"`
		ImageUris  struct {
			Normal string `json:"normal"`
		} `json:"image_uris"`
	} `json:"card_faces"`
}

func fetchScryfallData(name string) (string, string) {
	apiURL := fmt.Sprintf("https://api.scryfall.com/cards/named?exact=%s", url.QueryEscape(name))
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ""
	}

	var data ScryfallResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", ""
	}

	time.Sleep(100 * time.Millisecond) // Be polite to Scryfall

	img := data.ImageUris.Normal
	if img == "" && len(data.CardFaces) > 0 {
		img = data.CardFaces[0].ImageUris.Normal
	}

	desc := data.OracleText
	if desc == "" && len(data.CardFaces) > 0 {
		desc = data.CardFaces[0].OracleText
	}
	if data.FlavorText != "" {
		if desc != "" {
			desc += "\n\n"
		}
		desc += `_"` + data.FlavorText + `"_`
	}

	return img, desc
}

func main() {
	database := db.Connect()
	defer database.Close()

	// Clear products table
	fmt.Println("Clearing products table...")
	_, err := database.Exec(`DELETE FROM products`)
	if err != nil {
		log.Fatalf("Failed to clear products: %v", err)
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
		log.Fatalf("Failed to hash password: %v", err)
	}

	_, err = database.Exec(`
		INSERT INTO admins (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash
	`, adminUser, string(hash))
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}
	fmt.Printf("✓ Admin user '%s' created/updated\n", adminUser)

	// -------------------------------------------------------------------------
	// PRODUCT SEEDING (diverse and verified images)
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
	}

	products := []seedProduct{
		// MTG Singles
		{"Snapcaster Mage", "mtg", "singles", "Innistrad", "ISD", "NM", 85000, 2, []string{"featured"}, "Gives a spell flashback.", "https://cards.scryfall.io/normal/front/7/e/7e41765e-43fe-461d-baeb-bb30d14d0096.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Mana Crypt", "mtg", "singles", "Special Guests", "SPG", "NM", 780000, 1, []string{"featured", "hot-items"}, "Zero-cost mana acceleration.", "https://cards.scryfall.io/normal/front/c/c/cc823a07-5bd0-4597-990e-2c2621e1066a.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Lightning Bolt", "mtg", "singles", "Magic 2011", "M11", "NM", 25000, 12, []string{"sale"}, "Deals 3 damage.", "https://cards.scryfall.io/normal/front/f/2/f29ba16f-c8fb-42fe-aabf-074d4ee3032d.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Thoughtseize", "mtg", "singles", "Theros", "THS", "LP", 75000, 3, []string{"featured"}, "Target opponent discards a card.", "https://cards.scryfall.io/normal/front/1/2/12bb1514-6b83-42be-a178-574d6c4832be.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Sheoldred, the Apocalypse", "mtg", "singles", "Dominaria United", "DMU", "NM", 280000, 4, []string{"featured", "hot-items"}, "Whenever you draw a card, you gain 2 life.", "https://cards.scryfall.io/normal/front/d/6/d67be074-cdd4-41d9-ac89-0a0456c4e4b2.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"The One Ring", "mtg", "singles", "The Lord of the Rings: Tales of Middle-earth", "LTR", "NM", 450000, 1, []string{"hot-items"}, "Indestructible. Protection from everything.", "https://cards.scryfall.io/normal/front/d/7/d772c30c-8c71-4ac9-93b2-4f70fa447e16.jpg", models.FoilNonFoil, models.TreatmentNormal},

		// One Piece
		{"Monkey D. Luffy", "onepiece", "singles", "Romance Dawn", "OP01", "NM", 550000, 1, []string{"featured", "hot-items"}, "Parallel rare Luffy Leader.", "https://limitlesstcg.s3.us-central-1.amazonaws.com/one-piece/OP01/OP01-001_p1.png", models.FoilNonFoil, models.TreatmentAlternateArt},
		{"Roronoa Zoro", "onepiece", "singles", "Romance Dawn", "OP01", "NM", 250000, 2, []string{"featured"}, "Zoro Parallel Rare.", "https://limitlesstcg.s3.us-central-1.amazonaws.com/one-piece/OP01/OP01-025_p1.png", models.FoilNonFoil, models.TreatmentAlternateArt},
		{"Trafalgar Law", "onepiece", "singles", "Romance Dawn", "OP01", "NM", 180000, 2, []string{"new-arrivals"}, "Law Parallel Rare.", "https://limitlesstcg.s3.us-central-1.amazonaws.com/one-piece/OP01/OP01-047_p1.png", models.FoilNonFoil, models.TreatmentAlternateArt},

		// Yu-Gi-Oh!
		{"Dark Magician", "yugioh", "singles", "Metal Raiders", "MRD", "NM", 45000, 5, []string{"featured"}, "The ultimate wizard.", "https://images.ygoprodeck.com/images/cards/46986414.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Blue-Eyes White Dragon", "yugioh", "singles", "Legend of Blue Eyes White Dragon", "LOB", "NM", 120000, 2, []string{"featured", "hot-items"}, "This legendary dragon is a powerful engine of destruction.", "https://images.ygoprodeck.com/images/cards/89631139.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Ash Blossom & Joyous Spring", "yugioh", "singles", "Duel Devastator", "DUDE", "NM", 35000, 8, []string{"sale"}, "Handtrap essential.", "https://images.ygoprodeck.com/images/cards/14558127.jpg", models.FoilFoil, models.TreatmentNormal},

		// Lorcana
		{"Mickey Mouse - Tailor", "lorcana", "singles", "The First Chapter", "1ST", "NM", 65000, 4, []string{"featured"}, "Classic Mickey.", "https://lorcania.com/images/cards/1/mickey-mouse-brave-little-tailor.png", models.FoilNonFoil, models.TreatmentNormal},
		{"Elsa - Spirit of Winter", "lorcana", "singles", "The First Chapter", "1ST", "NM", 150000, 2, []string{"featured", "hot-items"}, "Powerful Elsa.", "https://lorcania.com/images/cards/1/elsa-spirit-of-winter.png", models.FoilNonFoil, models.TreatmentNormal},
		{"Maleficent - Monstrous Dragon", "lorcana", "singles", "The First Chapter", "1ST", "NM", 120000, 3, []string{"sale"}, "Maleficent Dragon.", "https://lorcania.com/images/cards/1/maleficent-monstrous-dragon.png", models.FoilNonFoil, models.TreatmentNormal},

		// Pokemon
		{"Charizard ex", "pokemon", "singles", "Obsidian Flames", "OBF", "NM", 350000, 2, []string{"featured", "hot-items"}, "Terastal Charizard.", "https://images.pokemontcg.io/sv3/125_hires.png", models.FoilFoil, models.TreatmentFullArt},
		{"Mew VMAX", "pokemon", "singles", "Fusion Strike", "FST", "NM", 180000, 3, []string{"featured"}, "Mew VMAX Alternate Art.", "https://images.pokemontcg.io/swsh8/269_hires.png", models.FoilFoil, models.TreatmentAlternateArt},
		{"Lugia VSTAR", "pokemon", "singles", "Silver Tempest", "SIT", "NM", 140000, 4, []string{"new-arrivals"}, "Lugia VSTAR Gold.", "https://images.pokemontcg.io/swsh12/211_hires.png", models.FoilFoil, models.TreatmentFullArt},

		// SEALED PRODUCT
		{"MTG Duskmourn Box", "mtg", "sealed", "Duskmourn", "DSK", "NM", 680000, 5, []string{"featured", "new-arrivals"}, "36 Play Boosters.", "https://media.wizards.com/2024/wpn/Duskmourn_PlayBoosterBox.png", models.FoilNonFoil, models.TreatmentNormal},
		{"Yu-Gi-Oh! Rarity Box", "yugioh", "sealed", "Rarity Collection", "RA01", "NM", 450000, 8, []string{"sale"}, "Luxury foil reprints.", "https://images.ygoprodeck.com/images/cards/25th_anniversary_rarity_collection.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Lorcana: Inklands Trove", "lorcana", "sealed", "Into the Inklands", "INK", "NM", 280000, 4, []string{"sale"}, "Collector chest.", "https://m.media-amazon.com/images/I/71M+B+B+L+L._AC_SL1500_.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Pokemon Prismatic ETB", "pokemon", "sealed", "Prismatic Evolutions", "PRV", "NM", 280000, 10, []string{"new-arrivals"}, "Elite Trainer Box.", "https://products.pokemoncenter.com/img/products/290-85614/290-85614_1.jpg", models.FoilNonFoil, models.TreatmentNormal},

		// ACCESSORIES
		{"Dragon Shield Pink", "accessories", "accessories", "N/A", "N/A", "NM", 55000, 25, []string{"sale"}, "100 Matte sleeves.", "https://m.media-amazon.com/images/I/71Q+B+B+L+L._AC_SL1500_.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Lorcana: Hades Sleeves", "accessories", "accessories", "N/A", "N/A", "NM", 65000, 15, []string{"sale"}, "Character art sleeves.", "https://m.media-amazon.com/images/I/71Y+B+B+L+L._AC_SL1500_.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Pokemon: Pikachu Playmat", "accessories", "accessories", "N/A", "N/A", "NM", 110000, 5, []string{"featured"}, "Official playmat.", "https://m.media-amazon.com/images/I/71Z+B+B+L+L._AC_SL1500_.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Satin Tower: Jet", "accessories", "accessories", "N/A", "N/A", "NM", 65000, 14, []string{"sale"}, "Hard-shell storage.", "https://m.media-amazon.com/images/I/61V+B+B+L+L._AC_SL1500_.jpg", models.FoilNonFoil, models.TreatmentNormal},

		// --- BATCH 2: ICONIC CARDS ---
		
		// MTG
		{"Black Lotus", "mtg", "singles", "Unlimited Edition", "2ED", "HP", 99999999.99, 1, []string{"hot-items"}, "The most iconic card.", "https://cards.scryfall.io/normal/front/b/d/bd8fa327-dd41-4737-8f19-2cf5eb1f7cdd.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Sol Ring", "mtg", "singles", "Commander 2021", "C21", "NM", 5000, 50, []string{"sale"}, "Mana rock staple.", "https://cards.scryfall.io/normal/front/4/d/4df6b871-0c46-451a-8294-f25492158866.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Oko, Thief of Crowns", "mtg", "singles", "Throne of Eldraine", "ELD", "NM", 45000, 3, []string{"featured"}, "Elk master.", "https://cards.scryfall.io/normal/front/3/4/3462a3d0-5552-49fa-9eb7-100960c55891.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Ragavan, Nimble Pilferer", "mtg", "singles", "Modern Horizons 2", "MH2", "NM", 180000, 2, []string{"hot-items"}, "The monkey king.", "https://cards.scryfall.io/normal/front/a/9/a9738cda-adb1-47fb-9f4c-ecd930228c4d.jpg", models.FoilNonFoil, models.TreatmentNormal},

		// One Piece
		{"Nami", "onepiece", "singles", "Romance Dawn", "OP01", "NM", 450000, 1, []string{"featured", "hot-items"}, "Parallel Rare Nami.", "https://limitlesstcg.s3.us-central-1.amazonaws.com/one-piece/OP01/OP01-016_p1.png", models.FoilNonFoil, models.TreatmentAlternateArt},
		{"Nico Robin", "onepiece", "singles", "Romance Dawn", "OP01", "NM", 150000, 2, []string{"new-arrivals"}, "Robin Parallel Rare.", "https://limitlesstcg.s3.us-central-1.amazonaws.com/one-piece/OP01/OP01-017_p1.png", models.FoilNonFoil, models.TreatmentAlternateArt},
		{"Boa Hancock", "onepiece", "singles", "Romance Dawn", "OP01", "NM", 220000, 2, []string{"featured"}, "Hancock Parallel Rare.", "https://limitlesstcg.s3.us-central-1.amazonaws.com/one-piece/OP01/OP01-078_p1.png", models.FoilNonFoil, models.TreatmentAlternateArt},

		// Yu-Gi-Oh!
		{"Exodia the Forbidden One", "yugioh", "singles", "Legend of Blue Eyes White Dragon", "LOB", "NM", 250000, 1, []string{"hot-items"}, "If you have all 5 pieces, you win.", "https://images.ygoprodeck.com/images/cards/33396948.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Monster Reborn", "yugioh", "singles", "Legend of Blue Eyes White Dragon", "LOB", "NM", 15000, 20, []string{"sale"}, "Special summon from GY.", "https://images.ygoprodeck.com/images/cards/83764718.jpg", models.FoilNonFoil, models.TreatmentNormal},
		{"Pot of Greed", "yugioh", "singles", "Legend of Blue Eyes White Dragon", "LOB", "NM", 25000, 15, []string{"sale"}, "Draw 2 cards.", "https://images.ygoprodeck.com/images/cards/55144522.jpg", models.FoilNonFoil, models.TreatmentNormal},

		// Lorcana
		{"Beast - Tragic Hero", "lorcana", "singles", "Rise of the Floodborn", "ROF", "NM", 180000, 2, []string{"hot-items"}, "Meta staple Beast.", "https://lorcania.com/images/cards/2/beast-tragic-hero.png", models.FoilNonFoil, models.TreatmentNormal},
		{"Cinderella - Stouthearted", "lorcana", "singles", "Rise of the Floodborn", "ROF", "NM", 120000, 3, []string{"featured"}, "Knight Cinderella.", "https://lorcania.com/images/cards/2/cinderella-stouthearted.png", models.FoilNonFoil, models.TreatmentNormal},

		// Pokemon
		{"Rayquaza VMAX", "pokemon", "singles", "Evolving Skies", "EVS", "NM", 850000, 1, []string{"hot-items"}, "Alternate Art Rayquaza.", "https://images.pokemontcg.io/swsh7/218_hires.png", models.FoilFoil, models.TreatmentAlternateArt},
		{"Giratina V", "pokemon", "singles", "Lost Origin", "LOR", "NM", 750000, 1, []string{"hot-items"}, "Alternate Art Giratina.", "https://images.pokemontcg.io/swsh11/186_hires.png", models.FoilFoil, models.TreatmentAlternateArt},
		{"Umbreon VMAX", "pokemon", "singles", "Evolving Skies", "EVS", "NM", 1500000, 1, []string{"hot-items", "featured"}, "Moonbreon Alt Art.", "https://images.pokemontcg.io/swsh7/215_hires.png", models.FoilFoil, models.TreatmentAlternateArt},
	}

	fmt.Println("Seeding custom categories...")
	categoryMap := make(map[string]string)
	categoriesToSeed := []struct{ Name, Slug string }{
		{"Featured", "featured"},
		{"Sale", "sale"},
		{"New Arrivals", "new-arrivals"},
		{"Hot Items", "hot-items"},
	}
	for _, cat := range categoriesToSeed {
		var catID string
		err = database.QueryRow(`
			INSERT INTO custom_categories (name, slug) 
			VALUES ($1, $2) 
			ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name 
			RETURNING id
		`, cat.Name, cat.Slug).Scan(&catID)
		if err != nil {
			log.Fatalf("Failed to seed category %s: %v", cat.Slug, err)
		}
		categoryMap[cat.Slug] = catID
		fmt.Printf("✓ Category '%s' ready\n", cat.Slug)
	}

	fmt.Println("Seeding storage locations...")
	var binderID, boxA_ID string
	err = database.QueryRow(`INSERT INTO stored_in (name) VALUES ('Binder 1') ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`).Scan(&binderID)
	if err != nil { log.Fatalf("Failed to seed Binder 1: %v", err) }
	
	err = database.QueryRow(`INSERT INTO stored_in (name) VALUES ('Box A') ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`).Scan(&boxA_ID)
	if err != nil { log.Fatalf("Failed to seed Box A: %v", err) }

	for _, p := range products {
		fmt.Printf("Seeding %s (%s)...\n", p.Name, p.TCG)
		var newProductID string
		err = database.QueryRow(`
			INSERT INTO products (
				name, tcg, category, set_name, set_code, condition,
				foil_treatment, card_treatment,
				price_cop_override, price_source, stock, description, image_url
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'manual', 0, $10, $11) RETURNING id
		`, p.Name, p.TCG, p.Category, p.SetName, p.SetCode, p.Condition,
			p.Foil, p.Treatment,
			p.PriceCOP, p.Description, p.ImageURL).Scan(&newProductID)
		
		if err != nil {
			log.Printf("Warning: failed to insert %s: %v", p.Name, err)
			continue
		}

		// Link categories
		for _, slug := range p.CategorySlugs {
			if id, ok := categoryMap[slug]; ok {
				database.Exec(`INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, newProductID, id)
			}
		}

		if p.Stock > 0 {
			qtyBinder := p.Stock / 2
			qtyBox := p.Stock - qtyBinder
			
			if qtyBinder > 0 {
				database.Exec(`INSERT INTO product_stored_in (product_id, stored_in_id, quantity) VALUES ($1, $2, $3)`, newProductID, binderID, qtyBinder)
			}
			if qtyBox > 0 {
				database.Exec(`INSERT INTO product_stored_in (product_id, stored_in_id, quantity) VALUES ($1, $2, $3)`, newProductID, boxA_ID, qtyBox)
			}
		}

		// ensure slightly different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("\n✅ Seed complete!")
}
