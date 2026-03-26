package external

// CardLookupResult holds card metadata returned by external TCG APIs.
type CardLookupResult struct {
	Name            string   `json:"name"`
	ImageURL        string   `json:"image_url"`
	SetName         string   `json:"set_name"`
	SetCode         string   `json:"set_code"`
	CollectorNumber string   `json:"collector_number,omitempty"`
	// Prices from Scryfall (MTG only). Both may be nil if unavailable.
	PriceTCGPlayer  *float64 `json:"price_tcgplayer,omitempty"`  // USD
	PriceCardmarket *float64 `json:"price_cardmarket,omitempty"` // EUR

	// MTG Metadata
	Language     string   `json:"language"`
	ColorIdentity *string  `json:"color_identity,omitempty"`
	Rarity       *string  `json:"rarity,omitempty"`
	CMC          *float64 `json:"cmc,omitempty"`
	IsLegendary  bool     `json:"is_legendary"`
	IsHistoric   bool     `json:"is_historic"`
	IsLand       bool     `json:"is_land"`
	IsBasicLand  bool     `json:"is_basic_land"`
	ArtVariation *string  `json:"art_variation,omitempty"`
	OracleText   *string  `json:"oracle_text,omitempty"`
	Artist       *string  `json:"artist,omitempty"`
	TypeLine     *string  `json:"type_line,omitempty"`
	BorderColor  *string  `json:"border_color,omitempty"`
	Frame        *string  `json:"frame,omitempty"`
	FullArt      bool     `json:"full_art"`
	Textless     bool     `json:"textless"`
}

