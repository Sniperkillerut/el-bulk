package main

import (
	"fmt"
	"log"
	"os"

	"github.com/el-bulk/backend/db"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	database := db.Connect()
	defer database.Close()

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

	// Seed MTG singles
	mtgSingles := []struct {
		Name          string
		SetName       string
		SetCode       string
		Condition     string
		FoilTreatment string
		CardTreatment string
		Price         float64
		Stock         int
		Featured      bool
		Description   string
	}{
		{"Lightning Bolt", "Magic 2011", "M11", "NM", "non_foil", "normal", 4.99, 12, true, "Classic direct damage spell. Goes in almost every red deck."},
		{"Counterspell", "Masters 25", "A25", "NM", "non_foil", "normal", 1.99, 8, false, "The definitive counter spell."},
		{"Thoughtseize", "Theros", "THS", "LP", "non_foil", "normal", 12.99, 3, true, "Premium hand disruption for any format."},
		{"Ragavan, Nimble Pilferer", "Modern Horizons 2", "MH2", "NM", "foil", "normal", 45.00, 1, true, "One of the most powerful one-drops in Modern."},
		{"Wrenn and Six", "Modern Horizons", "MH1", "NM", "non_foil", "borderless", 38.50, 2, true, "Top tier planeswalker for Modern and Legacy."},
		{"Solitude", "Modern Horizons 2", "MH2", "NM", "non_foil", "full_art", 22.00, 4, false, "White's premier removal spell."},
		{"Fury", "Modern Horizons 2", "MH2", "LP", "foil", "showcase", 18.00, 2, false, "Red evoke elemental for clearing boards."},
		{"Force of Will", "Alliances", "ALL", "MP", "non_foil", "normal", 89.99, 1, true, "The iconic free counterspell of Legacy."},
	}

	for _, c := range mtgSingles {
		_, err = database.Exec(`
			INSERT INTO products (name, tcg, category, set_name, set_code, condition, foil_treatment, card_treatment, price_cop_override, price_source, stock, featured, description)
			VALUES ($1, 'mtg', 'singles', $2, $3, $4, $5, $6, $7, 'manual', $8, $9, $10)
		`, c.Name, c.SetName, c.SetCode, c.Condition, c.FoilTreatment, c.CardTreatment, c.Price, c.Stock, c.Featured, c.Description)
		if err != nil {
			log.Printf("Warning: failed to insert %s: %v", c.Name, err)
		}
	}
	fmt.Println("✓ MTG singles seeded")

	// Seed MTG Sealed
	mtgSealed := []struct {
		Name        string
		SetName     string
		SetCode     string
		Price       float64
		Stock       int
		Featured    bool
		Description string
	}{
		{"Duskmourn: House of Horror Booster Box", "Duskmourn", "DSK", 149.99, 5, true, "36 Play Boosters from the horror-themed set."},
		{"Bloomburrow Collector Booster Box", "Bloomburrow", "BLB", 229.99, 3, true, "12 Collector Boosters with premium foils and alternate arts."},
		{"Modern Horizons 3 Play Booster Box", "Modern Horizons 3", "MH3", 299.99, 2, true, "36 Play Boosters — high power level set for Modern."},
		{"Outlaws of Thunder Junction Bundle", "Thunder Junction", "OTJ", 44.99, 8, false, "9 Play Boosters + promo card + basic lands."},
		{"Foundations Starter Kit", "Foundations", "FDN", 12.99, 15, false, "Two ready-to-play decks for new players."},
	}

	for _, s := range mtgSealed {
		_, err = database.Exec(`
			INSERT INTO products (name, tcg, category, set_name, set_code, foil_treatment, card_treatment, price_cop_override, price_source, stock, featured, description)
			VALUES ($1, 'mtg', 'sealed', $2, $3, 'non_foil', 'normal', $4, 'manual', $5, $6, $7)
		`, s.Name, s.SetName, s.SetCode, s.Price, s.Stock, s.Featured, s.Description)
		if err != nil {
			log.Printf("Warning: failed to insert %s: %v", s.Name, err)
		}
	}
	fmt.Println("✓ MTG sealed seeded")

	// Seed Accessories
	accessories := []struct {
		Name        string
		Price       float64
		Stock       int
		Description string
	}{
		{"Ultra Pro Dragon Shield Matte Sleeves (100ct)", 12.99, 20, "Matte finish, perfect shuffle feel. Compatible with standard card size."},
		{"BCW 800 Count Long Box", 8.99, 15, "Heavy-duty cardboard long box for storing your collection."},
		{"Ultimate Guard Quadrow Playmat", 34.99, 5, "3mm rubber base, playmat with quad-row design."},
		{"KMC Perfect Size Inner Sleeves (100ct)", 4.99, 30, "Double-sleeving inner sleeves for extra protection."},
		{"Gamegenic Watchtower 100+ Deck Box", 18.99, 10, "Holds 100 double-sleeved cards. Magnetic closure."},
		{"Ultra Pro 9-Pocket Pages (25ct)", 6.49, 25, "Standard binder pages for card organization."},
	}

	for _, a := range accessories {
		_, err = database.Exec(`
			INSERT INTO products (name, tcg, category, foil_treatment, card_treatment, price_cop_override, price_source, stock, description)
			VALUES ($1, 'accessories', 'accessories', 'non_foil', 'normal', $2, 'manual', $3, $4)
		`, a.Name, a.Price, a.Stock, a.Description)
		if err != nil {
			log.Printf("Warning: failed to insert %s: %v", a.Name, err)
		}
	}
	fmt.Println("✓ Accessories seeded")

	fmt.Println("\n✅ Seed complete!")
}
