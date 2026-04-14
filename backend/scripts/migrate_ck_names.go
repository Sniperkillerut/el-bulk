package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/el-bulk/backend/external"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/elbulk?sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Fetch all sets for MTG
	var sets []struct {
		Code string `db:"code"`
		Name string `db:"name"`
	}
	err = db.SelectContext(ctx, &sets, "SELECT code, name FROM tcg_set WHERE tcg = 'mtg' AND ck_name IS NULL")
	if err != nil {
		log.Fatalf("Failed to fetch sets: %v", err)
	}

	fmt.Printf("Migrating %d sets...\n", len(sets))

	for _, s := range sets {
		ckName := external.NormalizeCKEdition(s.Name)
		_, err = db.ExecContext(ctx, "UPDATE tcg_set SET ck_name = $1 WHERE tcg = 'mtg' AND code = $2", ckName, s.Code)
		if err != nil {
			fmt.Printf("  Error updating %s (%s): %v\n", s.Name, s.Code, err)
		} else {
			fmt.Printf("  Updated %s -> %s\n", s.Name, ckName)
		}
	}

	log.Println("Migration complete.")
}
