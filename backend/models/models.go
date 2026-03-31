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
	FoilEtchedFoil     FoilTreatment = "etched_foil"
	FoilGalaxyFoil     FoilTreatment = "galaxy_foil"
	FoilSurgeFoil      FoilTreatment = "surge_foil"
	FoilTexturedFoil   FoilTreatment = "textured_foil"
	FoilStepAndCompleat FoilTreatment = "step_and_compleat"
	FoilOilSlick       FoilTreatment = "oil_slick"
	FoilNeonInk        FoilTreatment = "neon_ink"
	FoilConfettiFoil   FoilTreatment = "confetti_foil"
	FoilDoubleRainbow  FoilTreatment = "double_rainbow"

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
	TreatmentSerialized   CardTreatment = "serialized"
	TreatmentStepAndCompleat CardTreatment = "step_and_compleat"

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
	Artist       *string  `db:"artist"            json:"artist,omitempty"`
	TypeLine     *string  `db:"type_line"         json:"type_line,omitempty"`
	BorderColor  *string  `db:"border_color"      json:"border_color,omitempty"`
	Frame        *string  `db:"frame"             json:"frame,omitempty"`
	FullArt      bool     `db:"full_art"          json:"full_art"`
	Textless     bool     `db:"textless"          json:"textless"`

	CreatedAt       time.Time         `db:"created_at"         json:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"         json:"updated_at"`

	// CartCount is the number of unique customers who have this product in a pending order.
	CartCount       int               `db:"-"                  json:"cart_count"`

	// DeckCards handles the cards included in a store_exclusives Deck
	DeckCards       []DeckCard        `db:"-"                  json:"deck_cards,omitempty"`
}

type DeckCard struct {
	ID              string `db:"id"               json:"id"`
	ProductID       string `db:"product_id"       json:"product_id"`
	Name            string `db:"name"             json:"name"`
	SetCode         string `db:"set_code"         json:"set_code,omitempty"`
	CollectorNumber string `db:"collector_number" json:"collector_number,omitempty"`
	Quantity        int    `db:"quantity"         json:"quantity"`
	TypeLine        string `db:"type_line"        json:"type_line,omitempty"`
	ImageURL        string `db:"image_url"        json:"image_url,omitempty"`
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
	CategoryIDs     []string          `json:"category_ids,omitempty"`
	StorageItems    []StorageLocation `json:"storage_items,omitempty"`
	ImageURL        *string           `json:"image_url,omitempty"`
	Description     *string           `json:"description,omitempty"`
	CollectorNumber *string           `json:"collector_number,omitempty"`
	PromoType       *string           `json:"promo_type,omitempty"`

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
	Artist       *string  `json:"artist,omitempty"`
	TypeLine     *string  `json:"type_line,omitempty"`
	BorderColor  *string  `json:"border_color,omitempty"`
	Frame        *string  `json:"frame,omitempty"`
	FullArt      bool     `json:"full_art"`
	Textless     bool     `json:"textless"`

	DeckCards    []DeckCard `json:"deck_cards,omitempty"`
}

// Settings holds admin-configurable global settings and contact info.
type Settings struct {
	USDToCOPRate float64 `json:"usd_to_cop_rate"`
	EURToCOPRate float64 `json:"eur_to_cop_rate"`

	// Contact Info
	ContactAddress   string   `json:"contact_address"`
	ContactPhone     string   `json:"contact_phone"`
	ContactEmail     string   `json:"contact_email"`
	ContactInstagram string   `json:"contact_instagram"`
	ContactHours     string   `json:"contact_hours"`
}

type TCG struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ItemCount int       `db:"item_count" json:"item_count"`
}

type TCGInput struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type ProductListResponse struct {
	Products    []Product `json:"products"`
	Total       int       `json:"total"`
	Page        int       `json:"page"`
	PageSize    int       `json:"page_size"`
	Facets      Facets    `json:"facets"`
	QueryTimeMS int64     `json:"query_time_ms"`
}

type Facets struct {
	Condition  map[string]int `json:"condition"`
	Foil       map[string]int `json:"foil"`
	Treatment  map[string]int `json:"treatment"`
	Rarity     map[string]int `json:"rarity"`
	Language   map[string]int `json:"language"`
	Color      map[string]int `json:"color"`
	Collection map[string]int `json:"collection"`
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
	StorageID  string `db:"storage_id" json:"stored_in_id"`
	Quantity   int    `db:"quantity"     json:"quantity"`
}

type StorageLocation struct {
	StorageID string `db:"storage_id" json:"stored_in_id"`
	Name       string `db:"name"         json:"name"`
	Quantity   int    `db:"quantity"     json:"quantity"`
}

// ── Orders ──────────────────────────────────────────────

type Customer struct {
	ID             string    `db:"id"               json:"id"`
	FirstName      string    `db:"first_name"       json:"first_name"`
	LastName       string    `db:"last_name"        json:"last_name"`
	Email          *string   `db:"email"            json:"email,omitempty"`
	Phone          *string   `db:"phone"            json:"phone,omitempty"`
	IDNumber       *string   `db:"id_number"        json:"id_number,omitempty"`
	Address        *string   `db:"address"          json:"address,omitempty"`
	AuthProvider   *string   `db:"auth_provider"    json:"auth_provider,omitempty"`
	AuthProviderID *string   `db:"auth_provider_id" json:"auth_provider_id,omitempty"`
	AvatarURL      *string   `db:"avatar_url"       json:"avatar_url,omitempty"`
	CreatedAt      time.Time `db:"created_at"       json:"created_at"`
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
	StorageID string `json:"stored_in_id"`
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
	CustomerName  string `db:"customer_name" json:"customer_name"`
	CustomerPhone string `db:"customer_phone" json:"customer_phone"`
	CustomerEmail string `db:"customer_email" json:"customer_email"`
	ItemCount     int    `db:"item_count"    json:"item_count"`
}

// ── Notices (Blog/News) ─────────────────────────────

type Notice struct {
	ID               string    `db:"id"                 json:"id"`
	Title            string    `db:"title"              json:"title"`
	Slug             string    `db:"slug"               json:"slug"`
	ContentHTML      string    `db:"content_html"       json:"content_html"`
	FeaturedImageURL *string   `db:"featured_image_url" json:"featured_image_url,omitempty"`
	IsPublished      bool      `db:"is_published"       json:"is_published"`
	CreatedAt        time.Time `db:"created_at"         json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"         json:"updated_at"`
}

type NoticeInput struct {
	Title            string  `json:"title"`
	Slug             string  `json:"slug"`
	ContentHTML      string  `json:"content_html"`
	FeaturedImageURL *string `json:"featured_image_url,omitempty"`
	IsPublished      bool    `json:"is_published"`
}

type NewsletterSubscriber struct {
	ID         string     `db:"id" json:"id"`
	Email      string     `db:"email" json:"email"`
	CustomerID *string    `db:"customer_id" json:"customer_id,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	// Join fields
	FirstName  *string    `db:"first_name" json:"first_name,omitempty"`
	LastName   *string    `db:"last_name" json:"last_name,omitempty"`
}

type CustomerNote struct {
	ID         string     `db:"id" json:"id"`
	CustomerID string     `db:"customer_id" json:"customer_id"`
	OrderID    *string    `db:"order_id" json:"order_id,omitempty"`
	Content    string     `db:"content" json:"content"`
	AdminID    *string    `db:"admin_id" json:"admin_id,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	// Join fields
	AdminName  *string    `db:"admin_name" json:"admin_name,omitempty"`
}

type ClientRequest struct {
	ID              string    `db:"id" json:"id"`
	CustomerID      *string   `db:"customer_id" json:"customer_id,omitempty"`
	CustomerName    string    `db:"customer_name" json:"customer_name"`
	CustomerContact string    `db:"customer_contact" json:"customer_contact"`
	CardName        string    `db:"card_name" json:"card_name"`
	SetName         *string   `db:"set_name" json:"set_name,omitempty"`
	Details         *string   `db:"details" json:"details,omitempty"`
	Status          string    `db:"status" json:"status"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type BountyOffer struct {
	ID         string    `db:"id" json:"id"`
	BountyID   string    `db:"bounty_id" json:"bounty_id"`
	CustomerID string    `db:"customer_id" json:"customer_id"`
	Quantity   int       `db:"quantity" json:"quantity"`
	Condition  *string   `db:"condition" json:"condition,omitempty"`
	Status     string    `db:"status" json:"status"`
	Notes      *string   `db:"notes" json:"notes,omitempty"`
	AdminNotes *string   `db:"admin_notes" json:"admin_notes,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	// Join fields
	BountyName      *string `db:"bounty_name" json:"bounty_name,omitempty"`
	CustomerName    string  `db:"customer_name" json:"customer_name"`
	CustomerContact string  `db:"customer_contact" json:"customer_contact"`
}

type ClientRequestInput struct {
	CustomerID      *string `json:"customer_id,omitempty"`
	CustomerName    string  `json:"customer_name"`
	CustomerContact string  `json:"customer_contact"`
	CardName        string  `json:"card_name"`
	SetName         *string `json:"set_name,omitempty"`
	Details         *string `json:"details,omitempty"`
}

type BountyOfferInput struct {
	BountyID        string  `json:"bounty_id"`
	CustomerName    string  `json:"customer_name"`
	CustomerContact string  `json:"customer_contact"`
	Quantity        int     `json:"quantity"`
	Condition       *string `json:"condition,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	AdminNotes      *string `json:"admin_notes,omitempty"`
	Status          *string `json:"status,omitempty"`
}

type UpdateClientRequestStatusInput struct {
	Status string `json:"status"` // expected: 'pending', 'accepted', 'rejected', 'solved'
}

type UpdateBountyOfferStatusInput struct {
	Status string `json:"status"` // expected: 'pending', 'accepted', 'rejected', 'fulfilled'
}

type CustomerStats struct {
	Customer
	OrderCount   int     `db:"order_count" json:"order_count"`
	TotalSpend   float64 `db:"total_spend" json:"total_spend"`
	IsSubscriber bool    `db:"is_subscriber" json:"is_subscriber"`
	LatestNote   *string `db:"latest_note" json:"latest_note"`
	RequestCount       int     `db:"request_count" json:"request_count"`
	ActiveRequestCount int     `db:"active_request_count" json:"active_request_count"`
	OfferCount         int     `db:"offer_count" json:"offer_count"`
	ActiveOfferCount   int     `db:"active_offer_count" json:"active_offer_count"`
}

type CustomerDetail struct {
	Customer
	Orders       []Order         `json:"orders"`
	Notes        []CustomerNote  `json:"notes"`
	Requests     []ClientRequest `json:"requests"`
	Offers       []BountyOffer   `json:"offers"`
	IsSubscriber bool            `json:"is_subscriber"`
}
