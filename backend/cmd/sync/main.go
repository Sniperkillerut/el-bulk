package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/el-bulk/backend/db"
	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func main() {
	ctx := context.Background()
	logger.SetLevel(logger.INFO)
	if os.Getenv("LOG_FORMAT") == "json" {
		logger.SetJSON(true)
	}

	logger.Info("Starting streaming price sync...")

	database, err := db.ConnectResilient()
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	// 1. Sync Scryfall
	if err := syncScryfall(ctx, database); err != nil {
		logger.Error("Scryfall sync failed: %v", err)
	}

	// 2. Sync Card Kingdom
	if err := syncCardKingdom(ctx, database); err != nil {
		logger.Error("CardKingdom sync failed: %v", err)
	}

	logger.Info("Price sync completed.")
}

func syncScryfall(ctx context.Context, db *sqlx.DB) error {
	// Step 1: download stream (DownloadBulkData handles URL discovery internally)
	body, err := external.DownloadBulkData(ctx, nil)
	if err != nil {
		return err
	}
	defer body.Close()

	// Step 2: Sync to DB
	return external.SyncScryfallToDB(ctx, db, body)
}

func syncCardKingdom(ctx context.Context, db *sqlx.DB) error {
	// Step 1: download stream
	resp, err := http.Get(external.CardKingdomPricelistURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CK status %d", resp.StatusCode)
	}

	// Step 2: Sync to DB
	return external.SyncCardKingdomToDB(ctx, db, resp.Body)
}
