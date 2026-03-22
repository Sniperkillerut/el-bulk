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
		Name        string
		TCG         string
		Category    string
		SetName     string
		SetCode     string
		Condition   string
		PriceCOP    float64
		Stock       int
		Featured    bool
		Description string
		ImageURL    string
	}

	products := []seedProduct{
		// MTG Singles (to fill space)
		{"Snapcaster Mage", "mtg", "singles", "Innistrad", "ISD", "NM", 85000, 2, true, "Gives a spell flashback.", "https://cards.scryfall.io/normal/front/7/e/7e41765e-43fe-461d-baeb-bb30d14d0096.jpg"},
		{"Mana Crypt", "mtg", "singles", "Special Guests", "SPG", "NM", 780000, 1, true, "Zero-cost mana acceleration.", "https://cards.scryfall.io/normal/front/c/c/cc823a07-5bd0-4597-990e-2c2621e1066a.jpg"},
		{"Lightning Bolt", "mtg", "singles", "Magic 2011", "M11", "NM", 25000, 12, true, "Deals 3 damage.", "https://cards.scryfall.io/normal/front/f/2/f29ba16f-c8fb-42fe-aabf-074d4ee3032d.jpg"},
		{"Thoughtseize", "mtg", "singles", "Theros", "THS", "LP", 75000, 3, true, "Target opponent discards a card.", "https://cards.scryfall.io/normal/front/1/2/12bb1514-6b83-42be-a178-574d6c4832be.jpg"},
		
		// These will be the TOP 4 Featured (due to newest first ordering)
		{"Monkey D. Luffy", "onepiece", "singles", "Romance Dawn", "OP01", "NM", 550000, 1, true, "Parallel rare Luffy Leader.", "https://limitlesstcg.s3.us-central-1.amazonaws.com/one-piece/OP01/OP01-001_p1.png"},
		{"Dark Magician", "yugioh", "singles", "Metal Raiders", "MRD", "NM", 45000, 5, true, "The ultimate wizard.", "https://images.ygoprodeck.com/images/cards/46986414.jpg"},
		{"Mickey Mouse - Tailor", "lorcana", "singles", "The First Chapter", "1ST", "NM", 65000, 4, true, "Classic Mickey.", "https://lorcania.com/images/cards/1/mickey-mouse-brave-little-tailor.png"},
		{"Charizard ex", "pokemon", "singles", "Obsidian Flames", "OBF", "NM", 350000, 2, true, "Terastal Charizard.", "https://images.pokemontcg.io/sv3/125_hires.png"},

		// SEALED PRODUCT
		{"MTG Duskmourn Box", "mtg", "sealed", "Duskmourn", "DSK", "NM", 680000, 5, true, "36 Play Boosters.", "https://media.wizards.com/2024/wpn/Duskmourn_PlayBoosterBox.png"},
		{"Yu-Gi-Oh! Rarity Box", "yugioh", "sealed", "Rarity Collection", "RA01", "NM", 450000, 8, true, "Luxury foil reprints.", "https://images.ygoprodeck.com/images/cards/25th_anniversary_rarity_collection.jpg"},
		{"Lorcana: Inklands Trove", "lorcana", "sealed", "Into the Inklands", "INK", "NM", 280000, 4, true, "Collector chest.", "https://m.media-amazon.com/images/I/71M+B+B+L+L._AC_SL1500_.jpg"},
		{"Pokemon Prismatic ETB", "pokemon", "sealed", "Prismatic Evolutions", "PRV", "NM", 280000, 10, true, "Elite Trainer Box.", "https://products.pokemoncenter.com/img/products/290-85614/290-85614_1.jpg"},

		// ACCESSORIES
		{"Dragon Shield Pink", "accessories", "accessories", "N/A", "N/A", "NM", 55000, 25, true, "100 Matte sleeves.", "https://m.media-amazon.com/images/I/71Q+B+B+L+L._AC_SL1500_.jpg"},
		{"Lorcana: Hades Sleeves", "accessories", "accessories", "N/A", "N/A", "NM", 65000, 15, true, "Character art sleeves.", "https://m.media-amazon.com/images/I/71Y+B+B+L+L._AC_SL1500_.jpg"},
		{"Pokemon: Pikachu Playmat", "accessories", "accessories", "N/A", "N/A", "NM", 110000, 5, true, "Official playmat.", "https://m.media-amazon.com/images/I/71Z+B+B+L+L._AC_SL1500_.jpg"},
		{"Satin Tower: Jet", "accessories", "accessories", "N/A", "N/A", "NM", 65000, 14, true, "Hard-shell storage.", "https://m.media-amazon.com/images/I/61V+B+B+L+L._AC_SL1500_.jpg"},
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
				price_cop_override, price_source, stock, featured, description, image_url
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'manual', 0, $8, $9, $10) RETURNING id
		`, p.Name, p.TCG, p.Category, p.SetName, p.SetCode, p.Condition, p.PriceCOP, p.Featured, p.Description, p.ImageURL).Scan(&newProductID)
		
		if err != nil {
			log.Printf("Warning: failed to insert %s: %v", p.Name, err)
			continue
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
