package store

import (
	"fmt"
	"strings"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type RefreshStore struct {
	DB *sqlx.DB
}

func NewRefreshStore(db *sqlx.DB) *RefreshStore {
	return &RefreshStore{DB: db}
}

// RefreshRow is the minimal product data needed for a price refresh.
type RefreshRow struct {
	ID            string  `db:"id"`
	TCG           string  `db:"tcg"`
	Name          string  `db:"name"`
	SetCode       *string `db:"set_code"`
	FoilTreatment string  `db:"foil_treatment"`
	PriceSource   string  `db:"price_source"`
}

func (s *RefreshStore) ListRefreshableProducts() ([]RefreshRow, error) {
	var rows []RefreshRow
	err := s.DB.Select(&rows, `
		SELECT id, tcg, name, set_code, foil_treatment, price_source
		FROM product
		WHERE price_source IN ('tcgplayer', 'cardmarket')
	`)
	return rows, err
}

type MetadataUpdate struct {
	ID         string
	Price      *float64
	Legalities models.JSONB
	OracleText string
	ScryfallID string
	TypeLine   string
	ImageURL   string
}

func (s *RefreshStore) BulkUpdateMetadata(updates []MetadataUpdate) (int, int) {
	var totalUpdated, totalErrors int

	chunkSize := 1000
	for i := 0; i < len(updates); i += chunkSize {
		end := i + chunkSize
		if end > len(updates) {
			end = len(updates)
		}
		chunk := updates[i:end]

		query := `
			UPDATE product AS p SET 
				price_reference = COALESCE(v.price_reference, p.price_reference),
				legalities = COALESCE(v.legalities, p.legalities),
				oracle_text = COALESCE(v.oracle_text, p.oracle_text),
				scryfall_id = COALESCE(v.scryfall_id, p.scryfall_id),
				type_line = COALESCE(v.type_line, p.type_line),
				image_url = COALESCE(v.image_url, p.image_url),
				updated_at = now()
			FROM (VALUES 
		`
		placeholders := make([]string, len(chunk))
		args := make([]interface{}, len(chunk)*7)

		for j, u := range chunk {
			base := j * 7
			placeholders[j] = fmt.Sprintf("($%d::uuid, $%d::numeric, $%d::jsonb, $%d::text, $%d::uuid, $%d::text, $%d::text)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7)

			args[base] = u.ID
			args[base+1] = u.Price
			args[base+2] = u.Legalities
			args[base+3] = u.OracleText
			if u.ScryfallID != "" {
				args[base+4] = u.ScryfallID
			} else {
				args[base+4] = nil
			}
			args[base+5] = u.TypeLine
			args[base+6] = u.ImageURL
		}

		query += strings.Join(placeholders, ", ")
		query += ") AS v(id, price_reference, legalities, oracle_text, scryfall_id, type_line, image_url) WHERE p.id = v.id"

		res, err := s.DB.Exec(query, args...)
		if err != nil {
			logger.Error("[price-refresh] Bulk DB update failed for chunk %d-%d: %v", i, end, err)
			totalErrors += len(chunk)
		} else {
			count, _ := res.RowsAffected()
			totalUpdated += int(count)
		}
	}

	return totalUpdated, totalErrors
}

// BuildPriceUpdates resolves prices from a Scryfall price map for the given products.
func BuildPriceUpdates(rows []RefreshRow, priceMap map[external.PriceKey]external.CardMetadata) ([]MetadataUpdate, int) {
	var updates []MetadataUpdate
	errs := 0

	for _, p := range rows {
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
			refPrice = prices.TCGPlayerUSD
		case "cardmarket":
			refPrice = prices.CardmarketEUR
		}

		updates = append(updates, MetadataUpdate{
			ID:         p.ID,
			Price:      refPrice,
			Legalities: prices.Legalities,
			OracleText: prices.OracleText,
			ScryfallID: prices.ScryfallID,
			TypeLine:   prices.TypeLine,
			ImageURL:   prices.ImageURL,
		})
	}

	return updates, errs
}
