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
	ID         string
	Price      *float64
	Legalities models.JSONB
	OracleText string
	ScryfallID string
	TypeLine   string
	ImageURL   string
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

		query := `
			UPDATE product AS p SET 
				price_reference = COALESCE(v.price_reference, p.price_reference),
				price_source = COALESCE(v.price_source, p.price_source),
				legalities = COALESCE(v.legalities, p.legalities),
				oracle_text = COALESCE(NULLIF(v.oracle_text, ''), p.oracle_text),
				scryfall_id = COALESCE(v.scryfall_id, p.scryfall_id),
				type_line = COALESCE(NULLIF(v.type_line, ''), p.type_line),
				image_url = COALESCE(NULLIF(v.image_url, ''), p.image_url),
				updated_at = now()
			FROM (VALUES 
		`
		placeholders := make([]string, len(chunk))
		args := make([]interface{}, len(chunk)*8)

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

		query += strings.Join(placeholders, ", ")
		query += ") AS v(id, price_reference, legalities, oracle_text, scryfall_id, type_line, image_url, price_source) WHERE p.id = v.id"

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

// BuildPriceUpdates resolves prices from Scryfall and CardKingdom price maps.
func BuildPriceUpdates(rows []RefreshRow, scryPriceMap map[external.PriceKey]external.CardMetadata, ckPriceMap map[string]*float64) ([]MetadataUpdate, int) {
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
		// Hierarchical lookup:
		// 1. Exact match (Name + Set + Collector + Foil)
		// 2. Set fallback (Name + Set + Foil)
		// 3. Global fallback (Name + Foil)
		
		scryMeta, hasScry := scryPriceMap[external.PriceKey{
			Name: name, SetCode: setCode, 
			Collector: strings.TrimSpace(p.CollectorNumber), 
			Foil: foil,
		}]
		
		if !hasScry {
			scryMeta, hasScry = scryPriceMap[external.PriceKey{Name: name, SetCode: setCode, Collector: "", Foil: foil}]
		}
		if !hasScry {
			scryMeta, hasScry = scryPriceMap[external.PriceKey{Name: name, SetCode: "", Collector: "", Foil: foil}]
		}

		if !hasScry && p.PriceSource != "cardkingdom" {
			logger.Warn("[price-refresh] no Scryfall metadata found for %q set=%s foil=%s", p.Name, setCode, foil)
			errs++
			continue
		}

		var refPrice *float64
		switch p.PriceSource {
		case "tcgplayer":
			refPrice = scryMeta.TCGPlayerUSD
		case "cardmarket":
			refPrice = scryMeta.CardmarketEUR
		case "cardkingdom":
			// 1. Try matching by CardKingdom ID if we have Scryfall metadata
			if hasScry && scryMeta.CardKingdomID != "" {
				if cp, ok := ckPriceMap["ckid:"+scryMeta.CardKingdomID]; ok {
					refPrice = cp
				}
			}

			// 2. Fallback to Name + Edition + Variation matching via the shared CK matcher
			if refPrice == nil {
				setName := ""
				if p.SetName != nil {
					setName = *p.SetName
				}
				isFoil := p.FoilTreatment != "non_foil"

				// Prefer the curated ck_name from DB; fall back to normalized Scryfall set name
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

				refPrice = external.LookupCKPrice(p.Name, ckEdition, variation, isFoil, ckPriceMap)
			}
		}

		if refPrice == nil {
			logger.Warn("[price-refresh] no price found for %q source=%s set=%s foil=%s", p.Name, p.PriceSource, setCode, foil)
			errs++
			continue
		}

		update := MetadataUpdate{
			ID:          p.ID,
			Price:       refPrice,
			PriceSource: p.PriceSource,
		}

		// Update metadata only if we have Scryfall info
		if hasScry {
			update.Legalities = scryMeta.Legalities
			update.OracleText = scryMeta.OracleText
			update.ScryfallID = scryMeta.ScryfallID
			update.TypeLine = scryMeta.TypeLine
			update.ImageURL = scryMeta.ImageURL
		}

		updates = append(updates, update)
	}

	return updates, errs
}
