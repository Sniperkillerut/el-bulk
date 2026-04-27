package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

const CardKingdomPricelistURL = "https://api.cardkingdom.com/api/v2/pricelist"

type CardKingdomProduct struct {
	ID          int    `json:"id"`
	SKU         string `json:"sku"`
	ScryfallID  string `json:"scryfall_id"`
	Name        string `json:"name"`
	Variation   string `json:"variation"`
	Edition     string `json:"edition"`
	IsFoil      string `json:"is_foil"`
	PriceRetail string `json:"price_retail"`
	QtyRetail   int    `json:"qty_retail"`
}

type CardKingdomResponse struct {
	Meta struct {
		CreatedAt   string `json:"created_at"`
		LastUpdated string `json:"last_updated"`
	} `json:"meta"`
	Data []CardKingdomProduct `json:"data"`
}

var ckClient = &http.Client{Timeout: 60 * time.Second}

var (
	ckCache      map[string]*float64
	ckNameIndex  map[string][]string // name -> list of composite keys
	ckCacheMutex sync.RWMutex
	ckCacheTime  time.Time
)

const CacheDuration = 1 * time.Hour

// BuildCardKingdomPriceMap loads the CK pricelist from the external_cardkingdom table.
// Uses an in-memory cache valid for 1 hour.
func BuildCardKingdomPriceMap(ctx context.Context, db *sqlx.DB) (map[string]*float64, error) {
	ckCacheMutex.RLock()
	if ckCache != nil && time.Since(ckCacheTime) < CacheDuration {
		logger.DebugCtx(ctx, "Using cached CardKingdom pricelist (loaded %s ago)", time.Since(ckCacheTime).Round(time.Minute))
		cacheCopy := ckCache
		ckCacheMutex.RUnlock()
		return cacheCopy, nil
	}
	ckCacheMutex.RUnlock()

	ckCacheMutex.Lock()
	defer ckCacheMutex.Unlock()

	// Double-check after acquiring lock
	if ckCache != nil && time.Since(ckCacheTime) < CacheDuration {
		return ckCache, nil
	}

	logger.InfoCtx(ctx, "Loading CardKingdom pricelist from database...")

	type row struct {
		ID          int      `db:"ck_id"`
		ScryfallID  *string  `db:"scryfall_id"`
		Name        string   `db:"name"`
		Edition     string   `db:"edition"`
		Variation   string   `db:"variation"`
		IsFoil      bool     `db:"is_foil"`
		PriceRetail *float64 `db:"price_retail"`
	}

	var rows []row
	err := db.SelectContext(ctx, &rows, "SELECT ck_id, scryfall_id, name, edition, variation, is_foil, price_retail FROM external_cardkingdom")
	if err != nil {
		return nil, fmt.Errorf("failed to query external_cardkingdom: %w", err)
	}

	priceMap := make(map[string]*float64, len(rows)*3)
	nameIndex := make(map[string][]string, len(rows))
	for _, p := range rows {
		if p.PriceRetail == nil {
			continue
		}
		priceCopy := *p.PriceRetail

		// Primary index: scryfall_id + foil
		if p.ScryfallID != nil && *p.ScryfallID != "" {
			foilSuffix := "non_foil"
			if p.IsFoil {
				foilSuffix = "foil"
			}
			scryKey := "scry:" + *p.ScryfallID + ":" + foilSuffix
			if existing, ok := priceMap[scryKey]; !ok || priceCopy < *existing {
				priceMap[scryKey] = &priceCopy
			}
		}

		// Secondary index: CK integer ID
		priceMap[fmt.Sprintf("ckid:%d", p.ID)] = &priceCopy

		// Fallback index
		key := generateCKKey(p.Name, p.Edition, p.Variation, p.IsFoil)
		priceMap[key] = &priceCopy

		// Populate NameIndex
		nameKey := strings.ToLower(strings.TrimSpace(p.Name))
		nameIndex[nameKey] = append(nameIndex[nameKey], key)
	}

	ckCache = priceMap
	ckNameIndex = nameIndex
	ckCacheTime = time.Now()

	logger.InfoCtx(ctx, "Loaded %d CardKingdom products from database", len(rows))
	return priceMap, nil
}

func generateCKKey(name, edition, variation string, isFoil bool) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// Do NOT apply NormalizeCKEdition here — that function translates Scryfall
	// set names to CK names. Applying it to CK's own edition names corrupts
	// them (e.g. "Ice Age" → "ice age edition"). Just lowercase as-is.
	edition = strings.ToLower(strings.TrimSpace(edition))
	variation = strings.ToLower(strings.TrimSpace(variation))

	foilSuffix := "non_foil"
	if isFoil {
		foilSuffix = "foil"
	}

	return fmt.Sprintf("%s|%s|%s|%s", name, edition, variation, foilSuffix)
}

// NormalizeCKEdition maps common Scryfall/TCGPlayer set names to CardKingdom's edition names.
// Used as a fallback when ck_name is not seeded in the DB.
func NormalizeCKEdition(edition string) string {
	e := strings.ToLower(edition)

	// Common mappings (Scryfall name → CK name)
	switch e {
	case "revised":
		return "revised edition"
	case "unlimited":
		return "unlimited edition"
	case "antiquities":
		return "antiquities edition"
	case "arabian nights":
		return "arabian nights edition"
	case "legends":
		return "legends edition"
	case "the dark":
		return "the dark edition"
	case "ice age":
		return "ice age"
	case "homelands":
		return "homelands"
	case "fourth edition", "4th edition":
		return "4th edition"
	case "fifth edition", "5th edition":
		return "5th edition"
	case "sixth edition", "6th edition":
		return "6th edition"
	case "seventh edition", "7th edition":
		return "7th edition"
	case "eighth edition", "8th edition":
		return "8th edition"
	case "ninth edition", "9th edition":
		return "9th edition"
	case "tenth edition", "10th edition":
		return "10th edition"
	case "classic sixth edition":
		return "6th edition"
	case "media promos":
		return "promotional"
	case "pro tour promos":
		return "promotional"
	case "judge gift cards", "judge gift program":
		return "promotional"
	case "magic player rewards":
		return "promotional"
	case "modern horizons 1":
		return "modern horizons"
	case "modern horizons 2":
		return "modern horizons 2"
	case "time spiral remastered":
		return "time spiral remastered"
	}

	return e
}

func MapFoilTreatmentToCKVariation(foil models.FoilTreatment, treatment models.CardTreatment) string {
	var parts []string

	switch treatment {
	case models.TreatmentBorderless:
		parts = append(parts, "borderless")
	case models.TreatmentExtendedArt:
		parts = append(parts, "extended art")
	case models.TreatmentShowcase:
		parts = append(parts, "showcase")
	case models.TreatmentEtched:
		parts = append(parts, "etched")
	case models.TreatmentFullArt:
		parts = append(parts, "full art")
	case models.TreatmentAlternateArt:
		parts = append(parts, "alternate art")
	case models.TreatmentLegacyBorder, "retro":
		parts = append(parts, "retro")
	}

	switch foil {
	case models.FoilEtchedFoil:
		parts = append(parts, "etched")
	case models.FoilSurgeFoil:
		parts = append(parts, "surge foil")
	case models.FoilGalaxyFoil:
		parts = append(parts, "galaxy foil")
	case models.FoilRippleFoil:
		parts = append(parts, "ripple foil")
	case models.FoilConfettiFoil:
		parts = append(parts, "confetti foil")
	case models.FoilDoubleRainbow:
		parts = append(parts, "double rainbow foil")
	case models.FoilTexturedFoil:
		parts = append(parts, "textured")
	}

	return strings.Join(parts, " ")
}

// SyncCardKingdomToDB streams the CK pricelist JSON into the external_cardkingdom table.
func SyncCardKingdomToDB(ctx context.Context, db *sqlx.DB, r io.Reader) error {
	logger.InfoCtx(ctx, "Syncing CardKingdom pricelist to database...")

	if r == nil {
		resp, err := http.Get(CardKingdomPricelistURL)
		if err != nil {
			return fmt.Errorf("failed to download cardkingdom pricelist: %w", err)
		}
		defer resp.Body.Close()
		r = resp.Body
	}

	// 1. Clear existing data (Full refresh pattern)
	if _, err := db.ExecContext(ctx, "TRUNCATE TABLE external_cardkingdom"); err != nil {
		return fmt.Errorf("failed to truncate external_cardkingdom: %w", err)
	}

	decoder := json.NewDecoder(r)

	// Navigate to the "data" array
	for {
		t, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("failed to find data array: %w", err)
		}
		if s, ok := t.(string); ok && s == "data" {
			break
		}
	}

	// Read opening '['
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("failed to read data array opening: %w", err)
	}

	const batchSize = 1000
	var batch []CardKingdomProduct

	for decoder.More() {
		var p CardKingdomProduct
		if err := decoder.Decode(&p); err != nil {
			logger.WarnCtx(ctx, "Skipping malformed CK product: %v", err)
			continue
		}
		batch = append(batch, p)

		if len(batch) >= batchSize {
			if err := flushCKBatch(ctx, db, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := flushCKBatch(ctx, db, batch); err != nil {
			return err
		}
	}

	logger.InfoCtx(ctx, "Successfully synced CardKingdom pricelist to database")
	return nil
}

func flushCKBatch(ctx context.Context, db *sqlx.DB, batch []CardKingdomProduct) error {
	query := `
		INSERT INTO external_cardkingdom (ck_id, scryfall_id, name, edition, variation, is_foil, price_retail)
		VALUES (:id, :scryfall_id, :name, :edition, :variation, :is_foil, :price_retail)
	`
	// Map is_foil string to boolean
	type row struct {
		ID          int      `db:"id"`
		ScryfallID  *string  `db:"scryfall_id"`
		Name        string   `db:"name"`
		Edition     string   `db:"edition"`
		Variation   string   `db:"variation"`
		IsFoil      bool     `db:"is_foil"`
		PriceRetail *float64 `db:"price_retail"`
	}

	rows := make([]row, len(batch))
	for i, p := range batch {
		var sid *string
		if p.ScryfallID != "" {
			sid = &p.ScryfallID
		}
		var price *float64
		if p.PriceRetail != "" {
			if v, err := strconv.ParseFloat(p.PriceRetail, 64); err == nil {
				price = &v
			}
		}

		rows[i] = row{
			ID:          p.ID,
			ScryfallID:  sid,
			Name:        p.Name,
			Edition:     p.Edition,
			Variation:   p.Variation,
			IsFoil:      p.IsFoil == "true",
			PriceRetail: price,
		}
	}

	_, err := db.NamedExecContext(ctx, query, rows)
	return err
}
