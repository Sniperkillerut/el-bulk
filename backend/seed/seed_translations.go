package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedTranslations(db *sqlx.DB) error {
	logger.Info("🌐 Seeding Storefront Translations...")

	// Find the translations directory with multiple fallbacks for different environments
	searchPaths := []string{
		filepath.Join("seed", "data", "translations"),         // Production (from Dockerfile)
		filepath.Join("backend", "seed", "data", "translations"), // Local (from project root)
		filepath.Join("data", "translations"),                 // Alternative production
	}

	basePath := ""
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			basePath = p
			break
		}
	}

	if basePath == "" {
		// Final fallback for debugging
		return fmt.Errorf("failed to find translations directory in any of: %v", searchPaths)
	}

	files, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read translations directory: %w", err)
	}

	type TranslEntry struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	totalSeeded := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		locale := strings.TrimSuffix(file.Name(), ".json")
		filePath := filepath.Join(basePath, file.Name())

		jsonData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", file.Name(), err)
		}

		var entries []TranslEntry
		if err := json.Unmarshal(jsonData, &entries); err != nil {
			return fmt.Errorf("failed to unmarshal translations from %s: %w", file.Name(), err)
		}

		for _, t := range entries {
			_, err := db.Exec(`
				INSERT INTO translation (key, locale, value)
				VALUES ($1, $2, $3)
				ON CONFLICT (key, locale) DO UPDATE SET value = EXCLUDED.value
			`, t.Key, locale, t.Value)
			if err != nil {
				return fmt.Errorf("failed to seed translation [key: %s, locale: %s]: %w", t.Key, locale, err)
			}
			totalSeeded++
		}
		logger.Info("  ✅ Loaded %d translations for locale: %s", len(entries), locale)
	}

	logger.Info("✅ Total %d translation records seeded across all locales", totalSeeded)
	return nil
}
