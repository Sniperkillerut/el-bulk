package external

import (
	"github.com/el-bulk/backend/models"
)

// CardLookupResult holds card metadata returned by external TCG APIs.
type CardLookupResult struct {
	Name            string   `json:"name"`
	ImageURL        string   `json:"image_url"`
	PriceTCGPlayer  *float64 `json:"price_tcgplayer,omitempty"`  // USD
	PriceCardmarket *float64 `json:"price_cardmarket,omitempty"` // EUR

	models.MTGMetadata
}

