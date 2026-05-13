package store

import (
	"context"
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
	ID              string  `db:"id"`
	TCG             string  `db:"tcg"`
	Name            string  `db:"name"`
	SetName         *string `db:"set_name"`
	SetCode         *string `db:"set_code"`
	CollectorNumber string  `db:"collector_number"`
	FoilTreatment   string  `db:"foil_treatment"`
	CardTreatment   string  `db:"card_treatment"`
	PriceSource     string  `db:"price_source"`
	ScryfallID      string  `db:"scryfall_id"`
	CKSetName       *string `db:"ck_set_name"`
}

func (s *RefreshStore) ListRefreshableProducts(ctx context.Context) ([]RefreshRow, error) {
	var rows []RefreshRow
	err := s.DB.SelectContext(ctx, &rows, `
		SELECT 
			p.id, p.tcg, p.name, p.set_name, p.set_code, p.collector_number, 
			p.foil_treatment, p.card_treatment, p.price_source, p.scryfall_id,
			s.ck_name as ck_set_name
		FROM product p
		LEFT JOIN tcg_set s ON p.tcg = s.tcg AND p.set_code = s.code
		WHERE p.price_source IN ('tcgplayer', 'cardmarket', 'cardkingdom')
	`)
	return rows, err
}

type MetadataUpdate struct {
	ID          string
	Price       *float64
	Legalities  models.JSONB
	OracleText  string
	ScryfallID  string
	TypeLine    string
	ImageURL    string
	PriceSource string
}

func (s *RefreshStore) BulkUpdateMetadata(ctx context.Context, updates []MetadataUpdate, usdRate, eurRate, ckRate float64) (int, int) {
	var totalUpdated, totalErrors int

	chunkSize := 1000
	for i := 0; i < len(updates); i += chunkSize {
		end := i + chunkSize
		if end > len(updates) {
			end = len(updates)
		}
		chunk := updates[i:end]

		placeholders := make([]string, len(chunk))
		args := make([]interface{}, len(chunk)*8)
		var query string

		for j, u := range chunk {
			base := j * 8
			placeholders[j] = fmt.Sprintf("($%d::uuid, $%d::numeric, $%d::jsonb, $%d::text, $%d::uuid, $%d::text, $%d::text, $%d::text)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)

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
			args[base+7] = u.PriceSource
		}

		totalArgs := len(chunk) * 8
		query = fmt.Sprintf(`
			UPDATE product AS p SET 
				price_reference = COALESCE(v.price_reference, p.price_reference),
				price_source = COALESCE(v.price_source, p.price_source),
				price_cop = COALESCE(p.price_cop_override, 
					CASE COALESCE(v.price_source, p.price_source)
						WHEN 'cardkingdom' THEN COALESCE(v.price_reference, p.price_reference) * $%d
						WHEN 'tcgplayer'   THEN COALESCE(v.price_reference, p.price_reference) * $%d
						WHEN 'cardmarket'  THEN COALESCE(v.price_reference, p.price_reference) * $%d
						ELSE COALESCE(v.price_reference, p.price_reference)
					END
				),
				legalities = COALESCE(v.legalities, p.legalities),
				oracle_text = COALESCE(NULLIF(v.oracle_text, ''), p.oracle_text),
				scryfall_id = COALESCE(v.scryfall_id, p.scryfall_id),
				type_line = COALESCE(NULLIF(v.type_line, ''), p.type_line),
				image_url = COALESCE(NULLIF(v.image_url, ''), p.image_url),
				updated_at = now()
			FROM (VALUES 
		`, totalArgs+1, totalArgs+2, totalArgs+3)

		query += strings.Join(placeholders, ", ")
		query += ") AS v(id, price_reference, legalities, oracle_text, scryfall_id, type_line, image_url, price_source) WHERE p.id = v.id"

		// Append rates to the args slice
		args = append(args, ckRate, usdRate, eurRate)

		res, err := s.DB.ExecContext(ctx, query, args...)
		if err != nil {
			logger.ErrorCtx(ctx, "[price-refresh] Bulk DB update failed for chunk %d-%d: %v", i, end, err)
			totalErrors += len(chunk)
		} else {
			count, _ := res.RowsAffected()
			totalUpdated += int(count)
		}
	}

	return totalUpdated, totalErrors
}

func (s *RefreshStore) RecalculateAllPrices(ctx context.Context, usdRate, eurRate, ckRate float64) error {
	query := `
		UPDATE product SET 
			price_cop = COALESCE(price_cop_override, 
				CASE price_source 
					WHEN 'cardkingdom' THEN price_reference * $1
					WHEN 'tcgplayer'   THEN price_reference * $2
					WHEN 'cardmarket'  THEN price_reference * $3
					ELSE price_reference 
				END
			),
			updated_at = now()
	`
	_, err := s.DB.ExecContext(ctx, query, ckRate, usdRate, eurRate)
	return err
}

// BuildPriceUpdates resolves prices from Scryfall and CardKingdom price maps.
func BuildPriceUpdates(rows []RefreshRow, scryPriceMap map[external.PriceKey]external.CardMetadata, scryIdMap map[string]external.CardMetadata, ckPriceMap map[string]*float64) ([]MetadataUpdate, int) {
	var updates []MetadataUpdate
	errs := 0

	for _, p := range rows {
		// Extract CK-specific metadata for the matcher
		setName := ""
		if p.SetName != nil {
			setName = *p.SetName
		}
		ckEdition := ""
		if p.CKSetName != nil && *p.CKSetName != "" {
			ckEdition = *p.CKSetName
		} else {
			ckEdition = external.NormalizeCKEdition(setName)
		}
		variation := external.MapFoilTreatmentToCKVariation(
			models.FoilTreatment(p.FoilTreatment),
			models.CardTreatment(p.CardTreatment),
		)

		pSetCode := ""
		if p.SetCode != nil {
			pSetCode = *p.SetCode
		}

		// Unified Resolve: implements ID > Set|CN > Set|Foil hierarchy
		pResult := external.ResolveMTGPrice(
			p.ScryfallID, p.Name, pSetCode, p.CollectorNumber, p.FoilTreatment,
			p.CardTreatment, ckEdition, variation,
			scryPriceMap, scryIdMap, ckPriceMap,
		)

		var refPrice *float64
		switch p.PriceSource {
		case "tcgplayer":
			refPrice = pResult.TCGPlayerUSD
		case "cardmarket":
			refPrice = pResult.CardmarketEUR
		case "cardkingdom":
			refPrice = pResult.CardKingdomUSD
		}

		if refPrice == nil {
			logger.Warn("[price-refresh] no price found for %q source=%s scryfallID=%s", p.Name, p.PriceSource, p.ScryfallID)
			errs++
			continue
		}

		update := MetadataUpdate{
			ID:          p.ID,
			Price:       refPrice,
			PriceSource: p.PriceSource,
		}

		// Update metadata only if we have Scryfall info
		if pResult.Metadata != nil {
			update.Legalities = pResult.Metadata.Legalities
			update.OracleText = pResult.Metadata.OracleText
			update.ScryfallID = pResult.Metadata.ScryfallID
			update.TypeLine = pResult.Metadata.TypeLine
			update.ImageURL = pResult.Metadata.ImageURL
		}

		updates = append(updates, update)
	}

	return updates, errs
}
