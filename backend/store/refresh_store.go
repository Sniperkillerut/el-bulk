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
}

func (s *RefreshStore) ListRefreshableProducts(ctx context.Context) ([]RefreshRow, error) {
	var rows []RefreshRow
	err := s.DB.SelectContext(ctx, &rows, `
		SELECT id, tcg, name, set_name, set_code, collector_number, foil_treatment, card_treatment, price_source, scryfall_id
		FROM product
		WHERE price_source IN ('tcgplayer', 'cardmarket', 'cardkingdom')
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
				oracle_text = COALESCE(v.oracle_text, p.oracle_text),
				scryfall_id = COALESCE(v.scryfall_id, p.scryfall_id),
				type_line = COALESCE(v.type_line, p.type_line),
				image_url = COALESCE(v.image_url, p.image_url),
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

			// 2. Fallback to Name + Edition + Variation matching
			if refPrice == nil {
				setName := ""
				if p.SetName != nil {
					setName = *p.SetName
				}
				isFoil := p.FoilTreatment != "non_foil"

				nameKeyPrefix := strings.ToLower(p.Name) + "|"
				foilSuffix := "|non_foil"
				if isFoil {
					foilSuffix = "|foil"
				}
				targetEdition := strings.ToLower(setName)
				targetCollector := strings.ToLower(strings.TrimSpace(p.CollectorNumber))

				var bestMatch *float64

				for k, cp := range ckPriceMap {
					if strings.HasPrefix(k, nameKeyPrefix) && strings.HasSuffix(k, foilSuffix) {
						parts := strings.Split(k, "|")
						if len(parts) >= 3 {
							ckEdition := parts[1]
							ckVariation := parts[2]
							
							if strings.Contains(ckVariation, "art card") || strings.Contains(ckVariation, "token") {
								continue
							}

							editionMatches := targetEdition != "" && (ckEdition == targetEdition || strings.Contains(ckEdition, targetEdition) || strings.Contains(targetEdition, ckEdition))
							collectorMatches := targetCollector != "" && (ckVariation == targetCollector || strings.Contains(ckVariation, targetCollector))

							if editionMatches {
								if collectorMatches {
									refPrice = cp
									break
								}
								if bestMatch == nil || (cp != nil && *cp > *bestMatch) {
									bestMatch = cp
								}
							} else if collectorMatches {
								if bestMatch == nil || (cp != nil && *cp > *bestMatch) {
									bestMatch = cp
								}
							}
						}
					}
				}

				if refPrice == nil && bestMatch != nil {
					refPrice = bestMatch
				}
				
				// Final pass: if still no price, take the highest available for the card name
				if refPrice == nil {
					for k, cp := range ckPriceMap {
						if strings.HasPrefix(k, nameKeyPrefix) && strings.HasSuffix(k, foilSuffix) {
							if bestMatch == nil || (cp != nil && *cp > *bestMatch) {
								bestMatch = cp
							}
						}
					}
					refPrice = bestMatch
				}
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
