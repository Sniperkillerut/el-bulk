package models

import (
	"time"
)

type FoilTreatment string
type CardTreatment string

const (
	FoilNonFoil      FoilTreatment = "non_foil"
	FoilFoil         FoilTreatment = "foil"
	FoilHoloFoil     FoilTreatment = "holo_foil"
	FoilPlatinumFoil FoilTreatment = "platinum_foil"
	FoilRippleFoil   FoilTreatment = "ripple_foil"
	FoilEtchedFoil   FoilTreatment = "etched_foil"
	FoilGalaxyFoil   FoilTreatment = "galaxy_foil"

	TreatmentNormal       CardTreatment = "normal"
	TreatmentFullArt      CardTreatment = "full_art"
	TreatmentExtendedArt  CardTreatment = "extended_art"
	TreatmentBorderless   CardTreatment = "borderless"
	TreatmentShowcase     CardTreatment = "showcase"
	TreatmentLegacyBorder CardTreatment = "legacy_border"
	TreatmentTextless     CardTreatment = "textless"
	TreatmentJudgePromo   CardTreatment = "judge_promo"
	TreatmentPromo        CardTreatment = "promo"
	TreatmentAlternateArt CardTreatment = "alternate_art"

	// PriceSource values
	PriceSourceTCGPlayer   = "tcgplayer"   // USD via Scryfall, use usd_to_cop_rate
	PriceSourceCardmarket  = "cardmarket"  // EUR via Scryfall, use eur_to_cop_rate
	PriceSourceManual      = "manual"      // price_cop_override is authoritative
)

// Product is the full DB row returned to API clients.
// The Price field is the computed COP price:
//   COALESCE(price_cop_override, price_reference * rate)
// where rate comes from the settings table.
type Product struct {
	ID              string        `db:"id"               json:"id"`
	Name            string        `db:"name"             json:"name"`
	TCG             string        `db:"tcg"              json:"tcg"`
	Category        string        `db:"category"         json:"category"`
	SetName         *string       `db:"set_name"         json:"set_name,omitempty"`
	SetCode         *string       `db:"set_code"         json:"set_code,omitempty"`
	Condition       *string       `db:"condition"        json:"condition,omitempty"`
	FoilTreatment   FoilTreatment `db:"foil_treatment"   json:"foil_treatment"`
	CardTreatment   CardTreatment `db:"card_treatment"   json:"card_treatment"`

	// Pricing fields
	PriceReference   *float64 `db:"price_reference"    json:"price_reference,omitempty"`
	PriceSource      string   `db:"price_source"       json:"price_source"`
	PriceCOPOverride *float64 `db:"price_cop_override" json:"price_cop_override,omitempty"`
	// Price is the computed COP value injected by the handler (not a DB column)
	Price            float64  `db:"-"                  json:"price"`

	Stock           int               `db:"stock"              json:"stock"`
	StoredIn        []StorageLocation `db:"-"                  json:"stored_in,omitempty"`
	Categories      []CustomCategory  `db:"-"                  json:"categories,omitempty"`
	ImageURL        *string           `db:"image_url"          json:"image_url,omitempty"`
	Description     *string           `db:"description"        json:"description,omitempty"`
	CollectorNumber *string           `db:"collector_number"   json:"collector_number,omitempty"`
	PromoType       *string           `db:"promo_type"         json:"promo_type,omitempty"`
	CreatedAt       time.Time         `db:"created_at"         json:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"         json:"updated_at"`
}

// ComputePrice calculates the COP price for this product given the exchange rates.
// Priority: price_cop_override > price_reference * rate > 0
func (p *Product) ComputePrice(usdToCOP, eurToCOP float64) float64 {
	if p.PriceCOPOverride != nil {
		return *p.PriceCOPOverride
	}
	if p.PriceReference != nil {
		switch p.PriceSource {
		case PriceSourceTCGPlayer:
			return *p.PriceReference * usdToCOP
		case PriceSourceCardmarket:
			return *p.PriceReference * eurToCOP
		}
	}
	return 0
}

type ProductInput struct {
	Name            string        `json:"name"`
	TCG             string        `json:"tcg"`
	Category        string        `json:"category"`
	SetName         *string       `json:"set_name,omitempty"`
	SetCode         *string       `json:"set_code,omitempty"`
	Condition       *string       `json:"condition,omitempty"`
	FoilTreatment   FoilTreatment `json:"foil_treatment"`
	CardTreatment   CardTreatment `json:"card_treatment"`

	// Pricing
	PriceReference   *float64 `json:"price_reference,omitempty"`
	PriceSource      string   `json:"price_source,omitempty"`
	PriceCOPOverride *float64 `json:"price_cop_override,omitempty"`

	Stock           int      `json:"stock"`
	CategoryIDs     []string `json:"category_ids,omitempty"`
	ImageURL        *string  `json:"image_url,omitempty"`
	Description     *string  `json:"description,omitempty"`
	CollectorNumber *string  `json:"collector_number,omitempty"`
	PromoType       *string  `json:"promo_type,omitempty"`
}

// Settings holds admin-configurable global settings and contact info.
type Settings struct {
	USDToCOPRate float64 `json:"usd_to_cop_rate"`
	EURToCOPRate float64 `json:"eur_to_cop_rate"`

	// Contact Info
	ContactAddress   string `json:"contact_address"`
	ContactPhone     string `json:"contact_phone"`
	ContactEmail     string `json:"contact_email"`
	ContactInstagram string `json:"contact_instagram"`
	ContactHours     string `json:"contact_hours"`
}

type ProductListResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type Admin struct {
	ID           string    `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StoredIn struct {
	ID        string `db:"id"         json:"id"`
	Name      string `db:"name"       json:"name"`
	ItemCount int    `db:"item_count" json:"item_count"`
}

type ProductStorage struct {
	ProductID  string `db:"product_id"   json:"product_id"`
	StoredInID string `db:"stored_in_id" json:"stored_in_id"`
	Quantity   int    `db:"quantity"     json:"quantity"`
}

type StorageLocation struct {
	StoredInID string `db:"stored_in_id" json:"stored_in_id"`
	Name       string `db:"name"         json:"name"`
	Quantity   int    `db:"quantity"     json:"quantity"`
}
