package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/crypto"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/el_bulk?sslmode=disable"
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if len(encryptionKey) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be exactly 32 bytes (got %d)", len(encryptionKey))
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Starting PII migration...")

	// 1. Migrate Customers
	var customers []models.Customer
	err = db.Select(&customers, "SELECT * FROM customer")
	if err != nil {
		log.Fatalf("Failed to fetch customers: %v", err)
	}

	for _, c := range customers {
		needsUpdate := false
		
		// Only encrypt if it looks like plain text (not base64 or doesn't decrypt successfully)
		if c.Phone != nil && *c.Phone != "" {
			_, err := crypto.Decrypt(*c.Phone)
			if err != nil {
				// Failed to decrypt, assume it's plain text and encrypt it
				enc, _ := crypto.Encrypt(*c.Phone)
				c.Phone = &enc
				needsUpdate = true
			}
		}

		if c.IDNumber != nil && *c.IDNumber != "" {
			_, err := crypto.Decrypt(*c.IDNumber)
			if err != nil {
				enc, _ := crypto.Encrypt(*c.IDNumber)
				c.IDNumber = &enc
				needsUpdate = true
			}
		}

		if c.Address != nil && *c.Address != "" {
			_, err := crypto.Decrypt(*c.Address)
			if err != nil {
				enc, _ := crypto.Encrypt(*c.Address)
				c.Address = &enc
				needsUpdate = true
			}
		}

		if needsUpdate {
			_, err = db.Exec(`
				UPDATE customer 
				SET phone = $1, id_number = $2, address = $3
				WHERE id = $4
			`, c.Phone, c.IDNumber, c.Address, c.ID)
			if err != nil {
				log.Printf("Failed to update customer %s: %v", c.ID, err)
			} else {
				fmt.Printf("Updated customer %s\n", c.ID)
			}
		}
	}

	fmt.Println("PII migration completed.")
}
