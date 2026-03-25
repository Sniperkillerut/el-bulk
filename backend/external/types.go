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

	// MTG Metadata
	Language     string   `json:"language"`
	Color        *string  `json:"color,omitempty"`
	Rarity       *string  `json:"rarity,omitempty"`
	CMC          *float64 `json:"cmc,omitempty"`
	IsLegendary  bool     `json:"is_legendary"`
	IsHistoric   bool     `json:"is_historic"`
	IsLand       bool     `json:"is_land"`
	IsBasicLand  bool     `json:"is_basic_land"`
	ArtVariation *string  `json:"art_variation,omitempty"`
}

