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
	
	// MTG Metadata
	Language      string   `db:"language"          json:"language"`
	ColorIdentity *string  `db:"color_identity"   json:"color_identity,omitempty"`
	Rarity        *string  `db:"rarity"            json:"rarity,omitempty"`
	CMC           *float64 `db:"cmc"               json:"cmc,omitempty"`
	IsLegendary  bool     `db:"is_legendary"      json:"is_legendary"`
	IsHistoric   bool     `db:"is_historic"       json:"is_historic"`
	IsLand       bool     `db:"is_land"           json:"is_land"`
	IsBasicLand  bool     `db:"is_basic_land"     json:"is_basic_land"`
	ArtVariation *string  `db:"art_variation"     json:"art_variation,omitempty"`
	OracleText   *string  `db:"oracle_text"       json:"oracle_text,omitempty"`
	FlavorText   *string  `db:"flavor_text"       json:"flavor_text,omitempty"`
	Artist       *string  `db:"artist"            json:"artist,omitempty"`
	TypeLine     *string  `db:"type_line"         json:"type_line,omitempty"`
	BorderColor  *string  `db:"border_color"      json:"border_color,omitempty"`
	Frame        *string  `db:"frame"             json:"frame,omitempty"`
	FullArt      bool     `db:"full_art"          json:"full_art"`
	Textless     bool     `db:"textless"          json:"textless"`

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

	// MTG Metadata
	Language      string   `json:"language,omitempty"`
	ColorIdentity *string  `json:"color_identity,omitempty"`
	Rarity        *string  `json:"rarity,omitempty"`
	CMC           *float64 `json:"cmc,omitempty"`
	IsLegendary  bool     `json:"is_legendary"`
	IsHistoric   bool     `json:"is_historic"`
	IsLand       bool     `json:"is_land"`
	IsBasicLand  bool     `json:"is_basic_land"`
	ArtVariation *string  `json:"art_variation,omitempty"`
	OracleText   *string  `json:"oracle_text,omitempty"`
	FlavorText   *string  `json:"flavor_text,omitempty"`
	Artist       *string  `json:"artist,omitempty"`
	TypeLine     *string  `json:"type_line,omitempty"`
	BorderColor  *string  `json:"border_color,omitempty"`
	Frame        *string  `json:"frame,omitempty"`
	FullArt      bool     `json:"full_art"`
	Textless     bool     `json:"textless"`
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

// ── Orders ──────────────────────────────────────────────

type Customer struct {
	ID        string    `db:"id"         json:"id"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name"  json:"last_name"`
	Email     *string   `db:"email"      json:"email,omitempty"`
	Phone     string    `db:"phone"      json:"phone"`
	IDNumber  *string   `db:"id_number"  json:"id_number,omitempty"`
	Address   *string   `db:"address"    json:"address,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Order struct {
	ID            string     `db:"id"             json:"id"`
	OrderNumber   string     `db:"order_number"   json:"order_number"`
	CustomerID    string     `db:"customer_id"    json:"customer_id"`
	Status        string     `db:"status"         json:"status"`
	PaymentMethod string     `db:"payment_method" json:"payment_method"`
	TotalCOP      float64    `db:"total_cop"      json:"total_cop"`
	Notes         *string    `db:"notes"          json:"notes,omitempty"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
	CompletedAt   *time.Time `db:"completed_at"   json:"completed_at,omitempty"`
}

type OrderItem struct {
	ID                string   `db:"id"                  json:"id"`
	OrderID           string   `db:"order_id"            json:"order_id"`
	ProductID         *string  `db:"product_id"          json:"product_id,omitempty"`
	ProductName       string   `db:"product_name"        json:"product_name"`
	ProductSet        *string  `db:"product_set"         json:"product_set,omitempty"`
	FoilTreatment     *string  `db:"foil_treatment"      json:"foil_treatment,omitempty"`
	CardTreatment     *string  `db:"card_treatment"      json:"card_treatment,omitempty"`
	Condition         *string  `db:"condition"           json:"condition,omitempty"`
	UnitPriceCOP      float64  `db:"unit_price_cop"      json:"unit_price_cop"`
	Quantity          int      `db:"quantity"            json:"quantity"`
	StoredInSnapshot  *string  `db:"stored_in_snapshot"  json:"stored_in_snapshot,omitempty"`
}

// OrderDetail is the enriched response for admin order viewing
type OrderDetail struct {
	Order    Order             `json:"order"`
	Customer Customer          `json:"customer"`
	Items    []OrderItemDetail `json:"items"`
}

type OrderItemDetail struct {
	OrderItem
	ImageURL  *string           `json:"image_url,omitempty"`
	Stock     int               `json:"stock"`
	StoredIn  []StorageLocation `json:"stored_in"`
}

type CreateOrderInput struct {
	FirstName     string             `json:"first_name"`
	LastName      string             `json:"last_name"`
	Email         string             `json:"email"`
	Phone         string             `json:"phone"`
	IDNumber      string             `json:"id_number"`
	Address       string             `json:"address"`
	PaymentMethod string             `json:"payment_method"`
	Notes         string             `json:"notes"`
	Items         []CreateOrderItem  `json:"items"`
}

type CreateOrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CompleteOrderInput struct {
	Decrements []StockDecrement `json:"decrements"`
}

type StockDecrement struct {
	ProductID  string `json:"product_id"`
	StoredInID string `json:"stored_in_id"`
	Quantity   int    `json:"quantity"`
}

type OrderListResponse struct {
	Orders   []OrderWithCustomer `json:"orders"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type OrderWithCustomer struct {
	Order
	CustomerName string `db:"customer_name" json:"customer_name"`
	ItemCount    int    `db:"item_count"    json:"item_count"`
}
