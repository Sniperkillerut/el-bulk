package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/el-bulk/backend/external"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://elbulk:elbulk@localhost:5432/elbulk?sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	fmt.Println("Starting Backfill for Bounties...")
	backfillBounties(ctx, db)

	fmt.Println("\nStarting Backfill for Client Requests...")
	backfillRequests(ctx, db)

	fmt.Println("\nBackfill Complete!")
}

func backfillBounties(ctx context.Context, db *sqlx.DB) {
	var bounties []struct {
		ID         string  `db:"id"`
		Name       string  `db:"name"`
		ScryfallID *string `db:"scryfall_id"`
	}

	err := db.SelectContext(ctx, &bounties, "SELECT id, name, scryfall_id FROM bounty WHERE tcg = 'mtg' AND oracle_id IS NULL")
	if err != nil {
		log.Printf("Error selecting bounties: %v", err)
		return
	}

	fmt.Printf("Found %d bounties to backfill\n", len(bounties))

	for _, b := range bounties {
		sid := ""
		if b.ScryfallID != nil {
			sid = *b.ScryfallID
		}

		time.Sleep(100 * time.Millisecond)
		res, err := external.LookupMTGCard(ctx, sid, b.Name, "", "", "")
		if err != nil {
			fmt.Printf("  [!] Failed to resolve bounty %s (%s): %v\n", b.ID, b.Name, err)
			continue
		}

		_, err = db.ExecContext(ctx, "UPDATE bounty SET oracle_id = $1, name = $2, scryfall_id = $3 WHERE id = $4", res.OracleID, res.Name, res.ScryfallID, b.ID)
		if err != nil {
			fmt.Printf("  [!] Failed to update bounty %s: %v\n", b.ID, err)
		} else {
			fmt.Printf("  [✓] Updated bounty %s -> %s (%s)\n", b.ID, res.Name, res.OracleID)
		}
	}
}

func backfillRequests(ctx context.Context, db *sqlx.DB) {
	var requests []struct {
		ID         string  `db:"id"`
		CardName   string  `db:"card_name"`
		ScryfallID *string `db:"scryfall_id"`
	}

	err := db.SelectContext(ctx, &requests, "SELECT id, card_name, scryfall_id FROM client_request WHERE tcg = 'mtg' AND oracle_id IS NULL")
	if err != nil {
		log.Printf("Error selecting requests: %v", err)
		return
	}

	fmt.Printf("Found %d requests to backfill\n", len(requests))

	for _, r := range requests {
		sid := ""
		if r.ScryfallID != nil {
			sid = *r.ScryfallID
		}

		time.Sleep(100 * time.Millisecond)
		res, err := external.LookupMTGCard(ctx, sid, r.CardName, "", "", "")
		if err != nil {
			fmt.Printf("  [!] Failed to resolve request %s (%s): %v\n", r.ID, r.CardName, err)
			continue
		}

		_, err = db.ExecContext(ctx, "UPDATE client_request SET oracle_id = $1, card_name = $2, scryfall_id = $3 WHERE id = $4", res.OracleID, res.Name, res.ScryfallID, r.ID)
		if err != nil {
			fmt.Printf("  [!] Failed to update request %s: %v\n", r.ID, err)
		} else {
			fmt.Printf("  [✓] Updated request %s -> %s (%s)\n", r.ID, res.Name, res.OracleID)
		}
	}
}
