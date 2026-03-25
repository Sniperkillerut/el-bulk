package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/el-bulk/backend/utils/logger"
)

// RefreshHandler handles on-demand and scheduled price refreshes.
type RefreshHandler struct {
	DB *sqlx.DB
}

func NewRefreshHandler(db *sqlx.DB) *RefreshHandler {
	return &RefreshHandler{DB: db}
}

// ─── Scryfall bulk data structures ──────────────────────────────────────────

// scryfallBulkMeta is the top-level response from GET /bulk-data.
type scryfallBulkMeta struct {
	Data []struct {
		Type        string `json:"type"`
		DownloadURI string `json:"download_uri"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"data"`
}

// scryfallBulkCard is the price-relevant subset of each card in the bulk file.
type scryfallBulkCard struct {
	Name   string `json:"name"`
	Set    string `json:"set"`    // set code, e.g. "m11"
	Prices struct {
		USD       *string `json:"usd"`
		USDFoil   *string `json:"usd_foil"`
		USDEtched *string `json:"usd_etched"`
		EUR       *string `json:"eur"`
		EURFoil   *string `json:"eur_foil"`
	} `json:"prices"`
}

// priceKey uniquely identifies a card+foil combination for the in-memory map.
type priceKey struct {
	name    string // lowercase
	setCode string // lowercase; empty = any set
	foil    string // foil_treatment value
}

// cardPrices holds extracted USD/EUR prices for one priceKey.
type cardPrices struct {
	tcgPlayerUSD *float64
	cardmarketEUR *float64
}

func parseF(s *string) *float64 {
	if s == nil || *s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// ─── Bulk download + price map ───────────────────────────────────────────────

// buildScryfallPriceMap downloads Scryfall's "default_cards" bulk file and
// builds a lookup map keyed by (name, setCode, foilTreatment).
// The download is streamed so memory usage stays bounded.
func buildScryfallPriceMap() (map[priceKey]cardPrices, error) {
	client := &http.Client{Timeout: 5 * time.Minute} // bulk file can be 600MB

	// Step 1: discover today's bulk-data download URL
	metaReq, _ := http.NewRequest(http.MethodGet, "https://api.scryfall.com/bulk-data", nil)
	metaReq.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
	metaReq.Header.Set("Accept", "application/json")

	metaResp, err := client.Do(metaReq)
	if err != nil {
		return nil, fmt.Errorf("fetching bulk-data index: %w", err)
	}
	defer metaResp.Body.Close()

	var meta scryfallBulkMeta
	if err := json.NewDecoder(metaResp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decoding bulk-data index: %w", err)
	}

	downloadURL := ""
	for _, item := range meta.Data {
		if item.Type == "default_cards" {
			downloadURL = item.DownloadURI
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("default_cards bulk file not found in scryfall response")
	}

	logger.Info("[price-refresh] downloading scryfall bulk data from %s", downloadURL)

	// Step 2: stream the bulk card JSON array
	dlReq, _ := http.NewRequest(http.MethodGet, downloadURL, nil)
	dlReq.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")

	dlResp, err := client.Do(dlReq)
	if err != nil {
		return nil, fmt.Errorf("downloading bulk data: %w", err)
	}
	defer dlResp.Body.Close()

	// Step 3: stream-decode the JSON array without loading 600MB into memory at once
	priceMap := make(map[priceKey]cardPrices, 300_000)
	decoder := json.NewDecoder(dlResp.Body)

	// Read opening '['
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("reading bulk JSON opening token: %w", err)
	}

	cardCount := 0
	for decoder.More() {
		var card scryfallBulkCard
		if err := decoder.Decode(&card); err != nil {
			// Skip malformed cards rather than failing the whole run
			continue
		}
		cardCount++

		name := strings.ToLower(card.Name)
		set := strings.ToLower(card.Set)

		// Register entries for each foil variant this print has prices for
		variants := []struct {
			foil string
			usd  *string
			eur  *string
		}{
			{"non_foil", card.Prices.USD, card.Prices.EUR},
			{"foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"holo_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"ripple_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"galaxy_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"platinum_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"etched_foil", card.Prices.USDEtched, card.Prices.EURFoil},
		}

		for _, v := range variants {
			tcg := parseF(v.usd)
			cm := parseF(v.eur)
			if tcg == nil && cm == nil {
				continue
			}
			// Index by specific set
			priceMap[priceKey{name: name, setCode: set, foil: v.foil}] = cardPrices{tcgPlayerUSD: tcg, cardmarketEUR: cm}
			// Also index by name+foil only (no set), so products missing set_code still match;
			// later entries overwrite earlier ones which is fine (any printing is better than none)
			priceMap[priceKey{name: name, setCode: "", foil: v.foil}] = cardPrices{tcgPlayerUSD: tcg, cardmarketEUR: cm}
		}
	}

	logger.Info("[price-refresh] bulk data loaded: %d card printings indexed", cardCount)
	return priceMap, nil
}

// ─── Refresh logic ───────────────────────────────────────────────────────────

// refreshRow is the minimal product data we need for a price refresh.
type refreshRow struct {
	ID            string  `db:"id"`
	TCG           string  `db:"tcg"`
	Name          string  `db:"name"`
	SetCode       *string `db:"set_code"`
	FoilTreatment string  `db:"foil_treatment"`
	PriceSource   string  `db:"price_source"`
}

// RunPriceRefresh fetches the Scryfall bulk data once and updates
// price_reference for all non-manual MTG products in one pass.
func RunPriceRefresh(db *sqlx.DB) (updated int, errs int) {
	rows := []refreshRow{}
	if err := db.Select(&rows, `
		SELECT id, tcg, name, set_code, foil_treatment, price_source
		FROM products
		WHERE price_source IN ('tcgplayer', 'cardmarket')
	`); err != nil {
		logger.Error("[price-refresh] failed to query products: %v", err)
		return 0, 1
	}

	if len(rows) == 0 {
		logger.Info("[price-refresh] no products with external price source, skipping")
		return 0, 0
	}

	// Separate MTG products (use Scryfall) from non-MTG (not yet supported)
	var mtgRows []refreshRow
	for _, r := range rows {
		if r.TCG == "mtg" {
			mtgRows = append(mtgRows, r)
		}
	}

	var priceMap map[priceKey]cardPrices
	if len(mtgRows) > 0 {
		var err error
		priceMap, err = buildScryfallPriceMap()
		if err != nil {
			logger.Error("[price-refresh] scryfall bulk download failed: %v", err)
			return 0, len(mtgRows)
		}
	}

	for _, p := range mtgRows {
		setCode := ""
		if p.SetCode != nil {
			setCode = strings.ToLower(*p.SetCode)
		}
		foil := strings.ToLower(p.FoilTreatment)
		name := strings.ToLower(p.Name)

		// Try specific set first, fall back to any set
		prices, ok := priceMap[priceKey{name: name, setCode: setCode, foil: foil}]
		if !ok {
			prices, ok = priceMap[priceKey{name: name, setCode: "", foil: foil}]
		}
		if !ok {
			logger.Warn("[price-refresh] no price found for %q set=%s foil=%s", p.Name, setCode, foil)
			errs++
			continue
		}

		var refPrice *float64
		switch p.PriceSource {
		case "tcgplayer":
			// We use TCGplayer Market Price (standard Scryfall 'usd' fields)
			refPrice = prices.tcgPlayerUSD
		case "cardmarket":
			// Scryfall's 'eur' already encapsulates Trend -> 1d -> 7d -> Avg fallback
			refPrice = prices.cardmarketEUR
		}

		if refPrice == nil {
			logger.Warn("[price-refresh] source price nil for %q (source: %s)", p.Name, p.PriceSource)
			errs++
			continue
		}

		if _, err := db.Exec("UPDATE products SET price_reference=$1 WHERE id=$2", *refPrice, p.ID); err != nil {
			logger.Error("[price-refresh] DB update failed for %s: %v", p.ID, err)
			errs++
			continue
		}
		updated++
	}

	logger.Info("[price-refresh] complete: %d updated, %d errors", updated, errs)
	return updated, errs
}

// POST /api/admin/prices/refresh
func (h *RefreshHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	updated, errs := RunPriceRefresh(h.DB)
	jsonOK(w, map[string]int{"updated": updated, "errors": errs})
}

// StartMidnightScheduler launches a goroutine that runs RunPriceRefresh
// once per day at midnight (server local time).
func StartMidnightScheduler(db *sqlx.DB) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			sleepDur := time.Until(next)
			logger.Info("⏰ Next price refresh in %s (at %s)",
				sleepDur.Round(time.Minute), next.Format("2006-01-02 15:04"))
			time.Sleep(sleepDur)

			logger.Info("[price-refresh] Starting scheduled midnight refresh...")
			updated, errs := RunPriceRefresh(db)
			logger.Info("[price-refresh] Done: %d updated, %d errors", updated, errs)
		}
	}()
}

