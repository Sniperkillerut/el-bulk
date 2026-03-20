package external

// CardLookupResult holds card metadata returned by external TCG APIs.
type CardLookupResult struct {
	ImageURL        string   `json:"image_url"`
	SetName         string   `json:"set_name"`
	SetCode         string   `json:"set_code"`
	CollectorNumber string   `json:"collector_number,omitempty"`
	// Prices from Scryfall (MTG only). Both may be nil if unavailable.
	PriceTCGPlayer  *float64 `json:"price_tcgplayer,omitempty"`  // USD
	PriceCardmarket *float64 `json:"price_cardmarket,omitempty"` // EUR
}

