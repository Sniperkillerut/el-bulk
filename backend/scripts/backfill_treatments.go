package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/external"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ScryfallCollectionRequest struct {
	Identifiers []Identifier `json:"identifiers"`
}

type Identifier struct {
	ID string `json:"id"`
}

type ScryfallCollectionResponse struct {
	Data []ScryfallCard `json:"data"`
}

type ScryfallCard struct {
	ID           string   `json:"id"`
	FrameEffects []string `json:"frame_effects"`
}

func main() {
	syncFlag := flag.Bool("sync", false, "Sync Scryfall bulk data to local cache before backfilling")
	flag.Parse()

	sqlxDB, err := db.ConnectResilient()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer sqlxDB.Close()

	ctx := context.Background()

	// 1. Optional Sync
	if *syncFlag {
		fmt.Println("Syncing Scryfall bulk data to local cache...")
		if err := external.SyncScryfallToDB(ctx, sqlxDB, nil); err != nil {
			log.Fatalf("Sync failed: %v", err)
		}
		fmt.Println("Sync completed.")
	}

	// 2. Gather all unique scryfall_ids that need backfilling
	ids, err := getIDsToBackfill(ctx, sqlxDB)
	if err != nil {
		log.Fatalf("Error gathering IDs: %v", err)
	}

	if len(ids) == 0 {
		fmt.Println("No cards need backfilling.")
		return
	}

	fmt.Printf("Found %d unique Scryfall IDs to backfill.\n", len(ids))

	// 3. Batch process in groups of 75 (Scryfall limit)
	const batchSize = 75
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batch := ids[i:end]
		fmt.Printf("Processing batch %d-%d of %d...\n", i+1, end, len(ids))

		if err := processBatch(ctx, sqlxDB, batch); err != nil {
			log.Printf("Error processing batch: %v", err)
		}
	}

	fmt.Println("Backfill completed successfully.")
}

func getIDsToBackfill(ctx context.Context, db *sqlx.DB) ([]string, error) {
	query := `
		SELECT DISTINCT scryfall_id FROM (
			SELECT CAST(scryfall_id AS TEXT) FROM product WHERE scryfall_id IS NOT NULL AND frame_effects IS NULL
			UNION
			SELECT CAST(scryfall_id AS TEXT) FROM deck_card WHERE scryfall_id IS NOT NULL AND frame_effects IS NULL
			UNION
			SELECT CAST(scryfall_id AS TEXT) FROM client_request WHERE scryfall_id IS NOT NULL AND frame_effects IS NULL
			UNION
			SELECT CAST(scryfall_id AS TEXT) FROM bounty WHERE scryfall_id IS NOT NULL AND frame_effects IS NULL
		) combined
	`
	var ids []string
	err := db.SelectContext(ctx, &ids, query)
	return ids, err
}

func processBatch(ctx context.Context, db *sqlx.DB, ids []string) error {
	// 1. Try to get from local cache first
	cards, missingIDs, err := fetchFromCache(ctx, db, ids)
	if err != nil {
		log.Printf("Cache lookup error: %v", err)
		missingIDs = ids // fallback to full API lookup
	}

	// 2. Fetch missing from Scryfall API
	if len(missingIDs) > 0 {
		fmt.Printf("  (%d cards missing from cache, fetching from API...)\n", len(missingIDs))
		apiCards, err := fetchFromAPI(missingIDs)
		if err != nil {
			return err
		}
		cards = append(cards, apiCards...)
		
		// Update cache with these new findings
		updateCache(db, apiCards)
		
		// Respect rate limit
		time.Sleep(100 * time.Millisecond)
	}

	// 3. Update all tables
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, card := range cards {
		feJSON, _ := json.Marshal(card.FrameEffects)
		if card.FrameEffects == nil {
			feJSON = []byte("[]")
		}

		// Update product
		_, _ = tx.ExecContext(ctx, "UPDATE product SET frame_effects = $1 WHERE scryfall_id = $2", string(feJSON), card.ID)
		// Update deck_card
		_, _ = tx.ExecContext(ctx, "UPDATE deck_card SET frame_effects = $1 WHERE scryfall_id = $2", string(feJSON), card.ID)
		// Update bounty
		_, _ = tx.ExecContext(ctx, "UPDATE bounty SET frame_effects = $1 WHERE scryfall_id = $2", string(feJSON), card.ID)
		// Update client_request (cast string to uuid)
		_, _ = tx.ExecContext(ctx, "UPDATE client_request SET frame_effects = $1 WHERE scryfall_id = $2", string(feJSON), card.ID)
	}

	return tx.Commit()
}

func fetchFromCache(ctx context.Context, db *sqlx.DB, ids []string) ([]ScryfallCard, []string, error) {
	query, args, err := sqlx.In("SELECT scryfall_id, frame_effects FROM external_scryfall WHERE scryfall_id IN (?) AND frame_effects IS NOT NULL", ids)
	if err != nil {
		return nil, ids, err
	}
	query = db.Rebind(query)

	type row struct {
		ID           string `db:"scryfall_id"`
		FrameEffects string `db:"frame_effects"`
	}
	var rows []row
	if err := db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, ids, err
	}

	foundMap := make(map[string]bool)
	var cards []ScryfallCard
	for _, r := range rows {
		var fe []string
		_ = json.Unmarshal([]byte(r.FrameEffects), &fe)
		cards = append(cards, ScryfallCard{ID: r.ID, FrameEffects: fe})
		foundMap[r.ID] = true
	}

	var missing []string
	for _, id := range ids {
		if !foundMap[id] {
			missing = append(missing, id)
		}
	}

	return cards, missing, nil
}

func fetchFromAPI(ids []string) ([]ScryfallCard, error) {
	req := ScryfallCollectionRequest{
		Identifiers: make([]Identifier, len(ids)),
	}
	for i, id := range ids {
		req.Identifiers[i] = Identifier{ID: id}
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post("https://api.scryfall.com/cards/collection", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scryfall status %d", resp.StatusCode)
	}

	var result ScryfallCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func updateCache(db *sqlx.DB, cards []ScryfallCard) {
	for _, card := range cards {
		feJSON, _ := json.Marshal(card.FrameEffects)
		if card.FrameEffects == nil {
			feJSON = []byte("[]")
		}
		_, _ = db.Exec("UPDATE external_scryfall SET frame_effects = $1 WHERE scryfall_id = $2", string(feJSON), card.ID)
	}
}
