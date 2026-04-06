package handlers

import (
	"fmt"
	"github.com/el-bulk/backend/utils/render"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/utils/logger"
)

// RefreshHandler handles on-demand and scheduled price refreshes.
type RefreshHandler struct {
	DB *sqlx.DB
}

func NewRefreshHandler(db *sqlx.DB) *RefreshHandler {
	return &RefreshHandler{DB: db}
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
		FROM product
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

	var priceMap map[external.PriceKey]external.CardPrices
	if len(mtgRows) > 0 {
		var err error
		priceMap, err = external.BuildPriceMap()
		if err != nil {
			logger.Error("[price-refresh] scryfall bulk download failed: %v", err)
			return 0, len(mtgRows)
		}
	}

	type priceUpdate struct {
		ID    string
		Price float64
	}
	updates := []priceUpdate{}

	for _, p := range mtgRows {
		setCode := ""
		if p.SetCode != nil {
			setCode = strings.ToLower(*p.SetCode)
		}
		foil := strings.ToLower(p.FoilTreatment)
		name := strings.ToLower(p.Name)

		// Try specific set first, fall back to any set
		prices, ok := priceMap[external.PriceKey{Name: name, SetCode: setCode, Foil: foil}]
		if !ok {
			prices, ok = priceMap[external.PriceKey{Name: name, SetCode: "", Foil: foil}]
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
			refPrice = prices.TCGPlayerUSD
		case "cardmarket":
			// Scryfall's 'eur' already encapsulates Trend -> 1d -> 7d -> Avg fallback
			refPrice = prices.CardmarketEUR
		}

		if refPrice == nil {
			logger.Warn("[price-refresh] source price nil for %q (source: %s)", p.Name, p.PriceSource)
			errs++
			continue
		}

		updates = append(updates, priceUpdate{ID: p.ID, Price: *refPrice})
	}

	// Execute updates in chunks to avoid PostgreSQL parameter limit (65,535)
	chunkSize := 1000
	for i := 0; i < len(updates); i += chunkSize {
		end := i + chunkSize
		if end > len(updates) {
			end = len(updates)
		}
		chunk := updates[i:end]

		query := "UPDATE product AS p SET price_reference = v.price_reference FROM (VALUES "
		placeholders := make([]string, len(chunk))
		args := make([]interface{}, len(chunk)*2)

		for j, u := range chunk {
			placeholders[j] = fmt.Sprintf("($%d::uuid, $%d::double precision)", j*2+1, j*2+2)
			args[j*2] = u.ID
			args[j*2+1] = u.Price
		}

		query += strings.Join(placeholders, ", ")
		query += ") AS v(id, price_reference) WHERE p.id = v.id"

		res, err := db.Exec(query, args...)
		if err != nil {
			logger.Error("[price-refresh] Bulk DB update failed for chunk %d-%d: %v", i, end, err)
			errs += len(chunk)
		} else {
			count, _ := res.RowsAffected()
			updated += int(count)
		}
	}

	logger.Info("[price-refresh] complete: %d updated, %d errors", updated, errs)
	return updated, errs
}

// POST /api/admin/prices/refresh
func (h *RefreshHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	updated, errs := RunPriceRefresh(h.DB)
	render.Success(w, map[string]int{"updated": updated, "errors": errs})
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
