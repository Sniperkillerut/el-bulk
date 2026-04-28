package external

import "github.com/el-bulk/backend/models"

// PriceKey is the standard key for in-memory pricing lookups.
type PriceKey struct {
	Name      string `db:"name"`
	SetCode   string `db:"set_code"`
	Collector string `db:"collector_number"`
}

// CardMetadata contains the essential data for pricing and display.
type CardMetadata struct {
	ScryfallID        string   `db:"scryfall_id"`
	TCGPlayerUSD      *float64 `db:"price_usd"`
	TCGPlayerUSDFoil  *float64 `db:"price_usd_foil"`
	CardmarketEUR     *float64 `db:"price_eur"`
	CardmarketEURFoil *float64 `db:"price_eur_foil"`
	ImageURL          string   `db:"image_url"`
	OracleText        string   `db:"oracle_text"`
	TypeLine          string       `db:"type_line"`
	Legalities        models.JSONB `db:"legalities"`
	CardKingdomID     string       `db:"card_kingdom_id"`
	CardKingdomFoilID string   `db:"card_kingdom_foil_id"`
}

// CardLookupResult represents the output of a single card resolution.
type CardLookupResult struct {
	Name             string   `json:"name"`
	ScryfallID       string   `json:"scryfall_id,omitempty"`
	OracleID         string   `json:"oracle_id,omitempty"`
	ImageURL         string   `json:"image_url"`
	PriceTCGPlayer   *float64 `json:"price_tcgplayer,omitempty"`   // USD
	PriceCardmarket  *float64 `json:"price_cardmarket,omitempty"`  // EUR
	PriceCardKingdom *float64 `json:"price_cardkingdom,omitempty"` // USD (CK Sell Price)

	models.MTGMetadata
}

// ToCardMetadata converts a lookup result into the curated metadata format.
func (c *CardLookupResult) ToCardMetadata() CardMetadata {
	m := CardMetadata{
		ScryfallID: c.ScryfallID,
		ImageURL:   c.ImageURL,
	}

	// Simple mapping for now - in a real resolution we'd have both prices
	// but here we just fill what we have from the lookup.
	if c.FoilTreatment != models.FoilNonFoil {
		m.TCGPlayerUSDFoil = c.PriceTCGPlayer
		m.CardmarketEURFoil = c.PriceCardmarket
	} else {
		m.TCGPlayerUSD = c.PriceTCGPlayer
		m.CardmarketEUR = c.PriceCardmarket
	}

	if c.SetName != nil {
		m.OracleText = *c.SetName // Placeholder if oracle text is missing
	}

	return m
}

// ResolvedPrices is the unified output structure for pricing operations.
type ResolvedPrices struct {
	ScryfallID     string
	TCGPlayerUSD   *float64
	CardmarketEUR  *float64
	CardKingdomUSD *float64
	Metadata       *CardMetadata
}

// PokemonCard is a minimal representation for Pokémon search results.
type PokemonCard struct {
	Name string `json:"name"`
}
