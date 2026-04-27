package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// JSONB is a map that can be scanned from and to a Postgres JSONB column.
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}

type FoilTreatment string
type CardTreatment string

const (
	FoilNonFoil         FoilTreatment = "non_foil"
	FoilFoil            FoilTreatment = "foil"
	FoilHoloFoil        FoilTreatment = "holo_foil"
	FoilPlatinumFoil    FoilTreatment = "platinum_foil"
	FoilRippleFoil      FoilTreatment = "ripple_foil"
	FoilEtchedFoil      FoilTreatment = "etched_foil"
	FoilGalaxyFoil      FoilTreatment = "galaxy_foil"
	FoilSurgeFoil       FoilTreatment = "surge_foil"
	FoilTexturedFoil    FoilTreatment = "textured_foil"
	FoilStepAndCompleat FoilTreatment = "step_and_compleat"
	FoilOilSlick        FoilTreatment = "oil_slick"
	FoilNeonInk         FoilTreatment = "neon_ink"
	FoilConfettiFoil    FoilTreatment = "confetti_foil"
	FoilDoubleRainbow   FoilTreatment = "double_rainbow"

	TreatmentNormal          CardTreatment = "normal"
	TreatmentFullArt         CardTreatment = "full_art"
	TreatmentExtendedArt     CardTreatment = "extended_art"
	TreatmentBorderless      CardTreatment = "borderless"
	TreatmentShowcase        CardTreatment = "showcase"
	TreatmentEtched          CardTreatment = "etched"
	TreatmentLegacyBorder    CardTreatment = "legacy_border"
	TreatmentTextless        CardTreatment = "textless"
	TreatmentJudgePromo      CardTreatment = "judge_promo"
	TreatmentPromo           CardTreatment = "promo"
	TreatmentAlternateArt    CardTreatment = "alternate_art"
	TreatmentSerialized      CardTreatment = "serialized"
	TreatmentStepAndCompleat CardTreatment = "step_and_compleat"
)

type PriceSource string

const (
	PriceSourceTCGPlayer   PriceSource = "tcgplayer"
	PriceSourceCardmarket  PriceSource = "cardmarket"
	PriceSourceCardKingdom PriceSource = "cardkingdom"
	PriceSourceManual      PriceSource = "manual"
)

// MTGMetadata holds specific data for Magic: The Gathering cards.
// Shared across Product, DeckCard, and external lookup results.
type MTGMetadata struct {
	SetName         *string       `db:"set_name"         json:"set_name,omitempty"`
	SetCode         *string       `db:"set_code"         json:"set_code,omitempty"`
	CollectorNumber *string       `db:"collector_number" json:"collector_number,omitempty"`
	FoilTreatment   FoilTreatment `db:"foil_treatment"   json:"foil_treatment"`
	CardTreatment   CardTreatment `db:"card_treatment"   json:"card_treatment"`

	Language      string   `db:"language"          json:"language"`
	ColorIdentity *string  `db:"color_identity"   json:"color_identity,omitempty"`
	Rarity        *string  `db:"rarity"            json:"rarity,omitempty"`
	CMC           *float64 `db:"cmc"               json:"cmc,omitempty"`
	IsLegendary   bool     `db:"is_legendary"      json:"is_legendary"`
	IsHistoric    bool     `db:"is_historic"       json:"is_historic"`
	IsLand        bool     `db:"is_land"           json:"is_land"`
	IsBasicLand   bool     `db:"is_basic_land"     json:"is_basic_land"`
	ArtVariation  *string  `db:"art_variation"     json:"art_variation,omitempty"`
	OracleText    *string  `db:"oracle_text"       json:"oracle_text,omitempty"`
	Artist        *string  `db:"artist"            json:"artist,omitempty"`
	TypeLine      *string  `db:"type_line"         json:"type_line,omitempty"`
	BorderColor   *string  `db:"border_color"      json:"border_color,omitempty"`
	Frame         *string  `db:"frame"             json:"frame,omitempty"`
	FullArt       bool     `db:"full_art"          json:"full_art"`
	Textless      bool     `db:"textless"          json:"textless"`
	PromoType     *string  `db:"promo_type"         json:"promo_type,omitempty"`
	ScryfallID    *string  `db:"scryfall_id"        json:"scryfall_id,omitempty"`
	Legalities    JSONB    `db:"legalities"         json:"legalities,omitempty"`
}

// Product is the full DB row returned to API clients.
// The Price field is the computed COP price:
//   COALESCE(price_cop_override, price_reference * rate)
// where rate comes from the settings table.
type Product struct {
	ID        string  `db:"id"               json:"id"`
	Name      string  `db:"name"             json:"name"`
	TCG       string  `db:"tcg"              json:"tcg"`
	Category  string  `db:"category"         json:"category"`
	Condition *string `db:"condition"        json:"condition,omitempty"`

	MTGMetadata `db:",inline"`

	// Pricing fields
	PriceReference   *float64    `db:"price_reference"    json:"price_reference,omitempty"`
	PriceSource      PriceSource `db:"price_source"       json:"price_source,omitempty"`
	PriceCOPOverride *float64    `db:"price_cop_override" json:"price_cop_override,omitempty"`
	// Price is the computed COP value injected by the handler (not a DB column)
	Price float64 `db:"-"                  json:"price"`

	Stock        int               `db:"stock"              json:"stock"`
	CostBasisCOP float64           `db:"cost_basis_cop"     json:"-"`
	StoredIn     []StorageLocation `db:"-"                  json:"stored_in,omitempty"`
	Categories   []CustomCategory  `db:"-"                  json:"categories,omitempty"`
	ImageURL     *string           `db:"image_url"          json:"image_url,omitempty"`
	Description  *string           `db:"description"        json:"description,omitempty"`

	CreatedAt *time.Time `db:"created_at"         json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at"         json:"updated_at,omitempty"`

	// Virtual fields
	IsHot bool `json:"is_hot"`
	IsNew bool `json:"is_new"`

	// CartCount is the number of unique customers who have this product in a pending order.
	CartCount int `db:"-"                  json:"cart_count"`

	// DeckCards handles the cards included in a store_exclusives Deck
	DeckCards []DeckCard `db:"-"                  json:"deck_cards,omitempty"`

	// SearchVector is used for database mapping of Postgres tsvector (FTS)
	SearchVector interface{} `db:"search_vector"      json:"-"`
}

// Redact strips sensitive or internal fields for non-admin users.
func (p *Product) Redact(isAdmin bool) {
	if !isAdmin {
		p.PriceReference = nil
		p.PriceSource = ""
		p.PriceCOPOverride = nil
		p.StoredIn = nil
		p.CreatedAt = nil
		p.UpdatedAt = nil

		// Also redact categories if needed
		for i := range p.Categories {
			p.Categories[i].Redact(isAdmin)
		}

		// If there are deck cards, they should also be redacted if they contain sensitive info
		// (Currently they mostly contain metadata which is public)
	}
}

type DeckCard struct {
	ID        string `db:"id"               json:"id"`
	ProductID string `db:"product_id"       json:"product_id"`
	Name      string `db:"name"             json:"name"`
	Quantity  int    `db:"quantity"         json:"quantity"`
	ImageURL  string `db:"image_url"        json:"image_url,omitempty"`

	MTGMetadata `db:",inline"`

	// SearchVector for FTS mapping
	SearchVector interface{} `db:"search_vector"      json:"-"`
}

// ComputePrice calculates the COP price for this product given the exchange rates.
// Priority: price_cop_override > price_reference * rate > 0
func (p *Product) ComputePrice(usdToCOP, eurToCOP, ckToCOP float64) float64 {
	if p.PriceCOPOverride != nil {
		return *p.PriceCOPOverride
	}
	if p.PriceReference != nil {
		switch p.PriceSource {
		case PriceSourceTCGPlayer:
			return *p.PriceReference * usdToCOP
		case PriceSourceCardKingdom:
			return *p.PriceReference * ckToCOP
		case PriceSourceCardmarket:
			return *p.PriceReference * eurToCOP
		}
	}
	return 0
}

type ProductInput struct {
	Name      string  `json:"name"`
	TCG       string  `json:"tcg"`
	Category  string  `json:"category"`
	Condition *string `json:"condition,omitempty"`

	MTGMetadata

	// Pricing
	PriceReference   *float64    `json:"price_reference,omitempty"`
	PriceSource      PriceSource `json:"price_source,omitempty"`
	PriceCOPOverride *float64    `json:"price_cop_override,omitempty"`

	Stock        int               `json:"stock"`
	CostBasis    float64           `json:"cost_basis"`
	CategoryIDs  []string          `json:"category_ids,omitempty"`
	StorageItems []StorageLocation `json:"storage_items,omitempty"`
	ImageURL     *string           `json:"image_url,omitempty"`
	Description  *string           `json:"description,omitempty"`

	DeckCards []DeckCard `json:"deck_cards,omitempty"`
}

var ValidConditions = map[string]bool{
	"NM":  true,
	"LP":  true,
	"MP":  true,
	"HP":  true,
	"DMG": true,
}

func (p *ProductInput) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.TCG == "" {
		return errors.New("tcg is required")
	}
	if p.Category == "" {
		return errors.New("category is required")
	}

	if p.Condition != nil && *p.Condition != "" {
		if !ValidConditions[*p.Condition] {
			return fmt.Errorf("invalid condition: %s. Must be one of NM, LP, MP, HP, DMG", *p.Condition)
		}
	}

	return nil
}

type BulkSearchRequest struct {
	List string `json:"list"`
}

type DeckMatch struct {
	RawLine      string    `json:"raw_line"`
	Quantity     int       `json:"quantity"`
	Matches      []Product `json:"matches"`
	IsMatched    bool      `json:"is_matched"`
	RequestedSet string    `json:"requested_set,omitempty"`
	RequestedCN  string    `json:"requested_cn,omitempty"`
}

type BulkSearchResponse struct {
	Matches []DeckMatch `json:"matches"`
}

type InventoryValuation struct {
	TotalItems        int     `json:"total_items"`
	TotalStock        int     `json:"total_stock"`
	TotalValueCOP     float64 `json:"total_value_cop"`
	TotalCostBasisCOP float64 `json:"total_cost_basis_cop"`
	PotentialProfit   float64 `json:"potential_profit"`
}

// Settings holds admin-configurable global settings and contact info.
type Settings struct {
	USDToCOPRate float64 `json:"usd_to_cop_rate"`
	EURToCOPRate float64 `json:"eur_to_cop_rate"`
	CKToCOPRate  float64 `json:"ck_to_cop_rate"`

	// Contact Info
	ContactAddress     string  `json:"contact_address"`
	ContactPhone       string  `json:"contact_phone"`
	ContactEmail       string  `json:"contact_email"`
	ContactInstagram   string  `json:"contact_instagram"`
	ContactHours       string  `json:"contact_hours"`
	FlatShippingFeeCOP float64 `json:"flat_shipping_fee_cop"`
	LastSetSync        string  `json:"last_set_sync"`
	DefaultThemeID     string  `json:"default_theme_id"`

	// Discovery Algorithms
	HotSalesThreshold int `json:"hot_sales_threshold"`
	HotDaysThreshold  int `json:"hot_days_threshold"`
	NewDaysThreshold  int `json:"new_days_threshold"`

	// Internationalization
	DefaultLocale        string `json:"default_locale"`
	HideLanguageSelector bool   `json:"hide_language_selector"`

	// Bogotá Express & Synergy Scout
	DeliveryPriorityEnabled bool    `json:"delivery_priority_enabled"`
	PriorityShippingFeeCOP  float64 `json:"priority_shipping_fee_cop"`
	SynergyMaxPriceCOP      float64 `json:"synergy_max_price_cop"`
}

// PublicSettings is the safe subset of Settings returned by the public /api/settings endpoint.
// It excludes exchange rates, algorithm thresholds, and operational data like last_set_sync.
type PublicSettings struct {
	ContactAddress          string  `json:"contact_address"`
	ContactPhone            string  `json:"contact_phone"`
	ContactEmail            string  `json:"contact_email"`
	ContactInstagram        string  `json:"contact_instagram"`
	ContactHours            string  `json:"contact_hours"`
	FlatShippingFeeCOP      float64 `json:"flat_shipping_fee_cop"`
	DefaultThemeID          string  `json:"default_theme_id"`
	DefaultLocale           string  `json:"default_locale"`
	HideLanguageSelector    bool    `json:"hide_language_selector"`
	DeliveryPriorityEnabled bool    `json:"delivery_priority_enabled"`
	PriorityShippingFeeCOP  float64 `json:"priority_shipping_fee_cop"`
}

// ToPublic returns the public-safe subset of Settings.
func (s Settings) ToPublic() PublicSettings {
	return PublicSettings{
		ContactAddress:          s.ContactAddress,
		ContactPhone:            s.ContactPhone,
		ContactEmail:            s.ContactEmail,
		ContactInstagram:        s.ContactInstagram,
		ContactHours:            s.ContactHours,
		FlatShippingFeeCOP:      s.FlatShippingFeeCOP,
		DefaultThemeID:          s.DefaultThemeID,
		DefaultLocale:           s.DefaultLocale,
		HideLanguageSelector:    s.HideLanguageSelector,
		DeliveryPriorityEnabled: s.DeliveryPriorityEnabled,
		PriorityShippingFeeCOP:  s.PriorityShippingFeeCOP,
	}
}

type TCG struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	ImageURL  *string   `db:"image_url"  json:"image_url"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ItemCount int       `db:"item_count" json:"item_count"`
}

type TCGSet struct {
	TCG        string  `db:"tcg"         json:"tcg"`
	Code       string  `db:"code"        json:"code"`
	Name       string  `db:"name"        json:"name"`
	ReleasedAt string  `db:"released_at" json:"released_at"`
	SetType    string  `db:"set_type"    json:"set_type"`
	CKName     *string `db:"ck_name"     json:"ck_name,omitempty"`

	// Virtual fields
	IsHot bool `json:"is_hot"`
	IsNew bool `json:"is_new"`
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

type FacetItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type Facets struct {
	Condition   map[string]int `json:"condition"`
	Foil        map[string]int `json:"foil"`
	Treatment   map[string]int `json:"treatment"`
	Rarity      map[string]int `json:"rarity"`
	Language    map[string]int `json:"language"`
	Color       map[string]int `json:"color"`
	Collection  map[string]int `json:"collection"`
	SetName     []FacetItem    `json:"set_name"`
	IsLegendary map[string]int `json:"is_legendary"`
	IsLand      map[string]int `json:"is_land"`
	IsHistoric  map[string]int `json:"is_historic"`
	Format      map[string]int `json:"format"`
}

type Admin struct {
	ID           string    `db:"id"           json:"id"`
	Username     string    `db:"username"     json:"username"`
	Email        string    `db:"email"        json:"email"`
	PasswordHash *string   `db:"password_hash" json:"-"`
	AvatarURL    *string   `db:"avatar_url"    json:"avatar_url,omitempty"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type AuditLog struct {
	ID            string    `db:"id"               json:"id"`
	AdminID       *string   `db:"admin_id"         json:"admin_id,omitempty"`
	AdminUsername string    `db:"admin_username"   json:"admin_username"`
	Action        string    `db:"action"           json:"action"`
	ResourceType  string    `db:"resource_type"    json:"resource_type"`
	ResourceID    string    `db:"resource_id"      json:"resource_id"`
	Details       JSONB     `db:"details"          json:"details"`
	IPAddress     *string   `db:"ip_address"       json:"ip_address,omitempty"`
	CreatedAt     time.Time `db:"created_at"       json:"created_at"`
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
	ProductID string `db:"product_id"   json:"product_id"`
	StorageID string `db:"storage_id" json:"stored_in_id"`
	Quantity  int    `db:"quantity"     json:"quantity"`
}

type StorageLocation struct {
	StorageID string `db:"storage_id" json:"stored_in_id"`
	Name      string `db:"name"         json:"name"`
	Quantity  int    `db:"quantity"     json:"quantity"`
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
	AuthProvider   *string   `db:"auth_provider"    json:"-"`
	AuthProviderID *string   `db:"auth_provider_id" json:"-"`
	AvatarURL      *string   `db:"avatar_url"       json:"avatar_url,omitempty"`
	CreatedAt      time.Time `db:"created_at"       json:"created_at"`

	// Virtual fields
	LinkedProviders []string `db:"-" json:"linked_providers,omitempty"`
}

type CustomerAuth struct {
	ID         string    `db:"id"`
	CustomerID string    `db:"customer_id"`
	Provider   string    `db:"provider"`
	ProviderID string    `db:"provider_id"`
	CreatedAt  time.Time `db:"created_at"`
}

type Order struct {
	ID                string     `db:"id"             json:"id"`
	OrderNumber       string     `db:"order_number"   json:"order_number"`
	CustomerID        string     `db:"customer_id"    json:"customer_id"`
	Status            string     `db:"status"         json:"status"`
	PaymentMethod     string     `db:"payment_method" json:"payment_method"`
	SubtotalCOP       float64    `db:"subtotal_cop"   json:"subtotal_cop"`
	ShippingCOP       float64    `db:"shipping_cop"   json:"shipping_cop"`
	TaxCOP            float64    `db:"tax_cop"        json:"tax_cop"`
	TotalCOP          float64    `db:"total_cop"      json:"total_cop"`
	TrackingNumber    *string    `db:"tracking_number" json:"tracking_number,omitempty"`
	TrackingURL       *string    `db:"tracking_url"    json:"tracking_url,omitempty"`
	IsLocalPickup     bool       `db:"is_local_pickup" json:"is_local_pickup"`
	IsPriority        bool       `db:"is_priority"      json:"is_priority"`
	InventoryRestored bool       `db:"inventory_restored" json:"inventory_restored"`
	Notes             *string    `db:"notes"          json:"notes,omitempty"`
	CreatedAt         time.Time  `db:"created_at"     json:"created_at"`
	ConfirmedAt       *time.Time `db:"confirmed_at"   json:"confirmed_at,omitempty"`
	CompletedAt       *time.Time `db:"completed_at"   json:"completed_at,omitempty"`
}

type OrderItem struct {
	ID               string  `db:"id"                  json:"id"`
	OrderID          string  `db:"order_id"            json:"order_id"`
	ProductID        *string `db:"product_id"          json:"product_id,omitempty"`
	ProductName      string  `db:"product_name"        json:"product_name"`
	ProductSet       *string `db:"product_set"         json:"product_set,omitempty"`
	FoilTreatment    *string `db:"foil_treatment"      json:"foil_treatment,omitempty"`
	CardTreatment    *string `db:"card_treatment"      json:"card_treatment,omitempty"`
	Condition        *string `db:"condition"           json:"condition,omitempty"`
	UnitPriceCOP     float64 `db:"unit_price_cop"      json:"unit_price_cop"`
	Quantity         int     `db:"quantity"            json:"quantity"`
	StoredInSnapshot *string `db:"stored_in_snapshot"  json:"stored_in_snapshot,omitempty"`
}

// OrderDetail is the enriched response for admin order viewing
type OrderDetail struct {
	Order       Order             `json:"order"`
	Customer    Customer          `json:"customer"`
	Items       []OrderItemDetail `json:"items"`
	WhatsAppURL string            `json:"whatsapp_url"`
}

// Redact strips internal fulfilment data for non-admin users.
func (d *OrderDetail) Redact(isAdmin bool) {
	if !isAdmin {
		for i := range d.Items {
			d.Items[i].Redact(isAdmin)
		}
	}
}

type OrderItemDetail struct {
	OrderItem
	ImageURL *string           `json:"image_url,omitempty"`
	Stock    int               `json:"stock"`
	StoredIn []StorageLocation `json:"stored_in"`
}

// Redact strips internal fulfilment data for non-admin users.
func (i *OrderItemDetail) Redact(isAdmin bool) {
	if !isAdmin {
		i.Stock = 0
		i.StoredIn = nil
	}
}

type CreateOrderInput struct {
	FirstName     string            `json:"first_name"`
	LastName      string            `json:"last_name"`
	Email         string            `json:"email"`
	Phone         string            `json:"phone"`
	IDNumber      string            `json:"id_number"`
	Address       string            `json:"address"`
	PaymentMethod string            `json:"payment_method"`
	IsLocalPickup bool              `json:"is_local_pickup"`
	IsPriority    bool              `json:"is_priority"`
	Notes         string            `json:"notes"`
	Items         []CreateOrderItem `json:"items"`
}

type CreateOrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type ConfirmOrderInput struct {
	Decrements []StockDecrement `json:"decrements"`
}

type StockDecrement struct {
	ProductID string `json:"product_id"`
	StorageID string `json:"stored_in_id"`
	Quantity  int    `json:"quantity"`
}

type UpdateOrderInput struct {
	Status         *string  `json:"status"`
	PaymentMethod  *string  `json:"payment_method"`
	ShippingCOP    *float64 `json:"shipping_cop"`
	TrackingNumber *string  `json:"tracking_number"`
	TrackingURL    *string  `json:"tracking_url"`
	Items          []struct {
		ID       string `json:"id"`
		Quantity int    `json:"quantity"`
	} `json:"items"`
	AddedItems []struct {
		ProductID    string  `json:"product_id"`
		Quantity     int     `json:"quantity"`
		UnitPriceCOP float64 `json:"unit_price_cop"`
	} `json:"added_items"`
	DeletedIDs []string `json:"deleted_ids"`
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

type OrderWithItemCount struct {
	Order
	ItemCount int `db:"item_count" json:"item_count"`
}

func (c *CustomCategory) TableName() string { return "custom_category" }

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
	ID         string    `db:"id" json:"id"`
	Email      string    `db:"email" json:"email"`
	CustomerID *string   `db:"customer_id" json:"customer_id,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	// Join fields
	FirstName *string `db:"first_name" json:"first_name,omitempty"`
	LastName  *string `db:"last_name" json:"last_name,omitempty"`
}

type CustomerNote struct {
	ID         string    `db:"id" json:"id"`
	CustomerID string    `db:"customer_id" json:"customer_id"`
	OrderID    *string   `db:"order_id" json:"order_id,omitempty"`
	Content    string    `db:"content" json:"content"`
	AdminID    *string   `db:"admin_id" json:"admin_id,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	// Join fields
	AdminName *string `db:"admin_name" json:"admin_name,omitempty"`
}

type ClientRequest struct {
	ID                 string     `db:"id" json:"id"`
	CustomerID         *string    `db:"customer_id" json:"customer_id,omitempty"`
	CustomerName       string     `db:"customer_name" json:"customer_name"`
	CustomerContact    string     `db:"customer_contact" json:"customer_contact"`
	CardName           string     `db:"card_name" json:"card_name"`
	SetName            *string    `db:"set_name" json:"set_name,omitempty"`
	Details            *string    `db:"details" json:"details,omitempty"`
	Quantity           int        `db:"quantity" json:"quantity"`
	TCG                string     `db:"tcg" json:"tcg"`
	Status             string     `db:"status" json:"status"`
	CancellationReason *string    `db:"cancellation_reason,omitempty" json:"cancellation_reason,omitempty"`
	BountyID           *string    `db:"bounty_id" json:"bounty_id,omitempty"`
	MatchType          string     `db:"match_type" json:"match_type"`
	ScryfallID         *string    `db:"scryfall_id" json:"scryfall_id,omitempty"`
	ImageURL           *string    `db:"image_url" json:"image_url,omitempty"`
	FoilTreatment      *string    `db:"foil_treatment" json:"foil_treatment,omitempty"`
	CardTreatment      *string    `db:"card_treatment" json:"card_treatment,omitempty"`
	SetCode            *string    `db:"set_code" json:"set_code,omitempty"`
	CollectorNumber    *string    `db:"collector_number" json:"collector_number,omitempty"`
	CreatedAt          *time.Time `db:"created_at" json:"created_at,omitempty"`
}

type BountyOffer struct {
	ID         string     `db:"id" json:"id"`
	BountyID   string     `db:"bounty_id" json:"bounty_id"`
	CustomerID string     `db:"customer_id" json:"customer_id"`
	Quantity   int        `db:"quantity" json:"quantity"`
	Condition  *string    `db:"condition" json:"condition,omitempty"`
	Status     string     `db:"status" json:"status"`
	Notes      *string    `db:"notes" json:"notes,omitempty"`
	AdminNotes *string    `db:"admin_notes" json:"admin_notes,omitempty"`
	CreatedAt  *time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt  *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	// Join fields
	BountyName      *string `db:"bounty_name" json:"bounty_name,omitempty"`
	BountyImage     *string `db:"bounty_image" json:"bounty_image,omitempty"`
	BountyFoil      *string `db:"bounty_foil"  json:"bounty_foil,omitempty"`
	ScryfallID      *string `db:"scryfall_id"  json:"scryfall_id,omitempty"`
	TCG             string  `db:"tcg"         json:"tcg"`
	CustomerName    string  `db:"customer_name" json:"customer_name"`
	CustomerContact string  `db:"customer_contact" json:"customer_contact,omitempty"`
}

type ClientRequestInput struct {
	CustomerID      *string `json:"customer_id,omitempty"`
	CustomerName    string  `json:"customer_name"`
	CustomerContact string  `json:"customer_contact"`
	CardName        string  `json:"card_name"`
	SetName         *string `json:"set_name,omitempty"`
	Details         *string `json:"details,omitempty"`
	Quantity        int     `json:"quantity"`
	TCG             string  `json:"tcg"`
	MatchType       string  `json:"match_type"`
	ScryfallID      *string `json:"scryfall_id,omitempty"`
	ImageURL        *string `json:"image_url,omitempty"`
	FoilTreatment   *string `json:"foil_treatment,omitempty"`
	CardTreatment   *string `json:"card_treatment,omitempty"`
	SetCode         *string `json:"set_code,omitempty"`
	CollectorNumber *string `json:"collector_number,omitempty"`
}

type ClientRequestBatchInput struct {
	CustomerName    string `json:"customer_name"`
	CustomerContact string `json:"customer_contact"`
	Cards           []struct {
		CardName        string  `json:"card_name"`
		SetName         *string `json:"set_name,omitempty"`
		Details         *string `json:"details,omitempty"`
		Quantity        int     `json:"quantity"`
		TCG             string  `json:"tcg"`
		ScryfallID      *string `json:"scryfall_id,omitempty"`
		ImageURL        *string `json:"image_url,omitempty"`
		FoilTreatment   *string `json:"foil_treatment,omitempty"`
		CardTreatment   *string `json:"card_treatment,omitempty"`
		SetCode         *string `json:"set_code,omitempty"`
		CollectorNumber *string `json:"collector_number,omitempty"`
	} `json:"cards"`
}

type ClientRequestBatchResponse struct {
	Count      int    `json:"count"`
	CustomerID string `json:"customer_id"`
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

type CancelMeRequestInput struct {
	Reason  string `json:"reason"`
	Details string `json:"details"`
}

type UpdateBountyOfferStatusInput struct {
	Status string `json:"status"` // expected: 'pending', 'accepted', 'rejected', 'fulfilled'
}

type CustomerStats struct {
	Customer
	OrderCount         int     `db:"order_count" json:"order_count"`
	TotalSpend         float64 `db:"total_spend" json:"total_spend"`
	IsSubscriber       bool    `db:"is_subscriber" json:"is_subscriber"`
	LatestNote         *string `db:"latest_note" json:"latest_note"`
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

// ── Themes ─────────────────────────────────────────────

type Theme struct {
	ID       string `db:"id"                   json:"id"`
	Name     string `db:"name"                 json:"name"`
	IsSystem bool   `db:"is_system"            json:"is_system"`

	// Colors
	BgPage    string `db:"bg_page"              json:"bg_page"`
	BgHeader  string `db:"bg_header"            json:"bg_header"`
	BgSurface string `db:"bg_surface"           json:"bg_surface"`

	TextMain      string `db:"text_main"            json:"text_main"`
	TextSecondary string `db:"text_secondary"       json:"text_secondary"`
	TextMuted     string `db:"text_muted"           json:"text_muted"`
	TextOnAccent  string `db:"text_on_accent"       json:"text_on_accent"`

	AccentPrimary      string `db:"accent_primary"       json:"accent_primary"`
	AccentPrimaryHover string `db:"accent_primary_hover" json:"accent_primary_hover"`
	BorderMain         string `db:"border_main"          json:"border_main"`

	StatusNM  string `db:"status_nm"            json:"status_nm"`
	StatusLP  string `db:"status_lp"            json:"status_lp"`
	StatusMP  string `db:"status_mp"            json:"status_mp"`
	StatusHP  string `db:"status_hp"            json:"status_hp"`
	StatusDMG string `db:"status_dmg"           json:"status_dmg"`

	// Specialized Colors
	BgCard       string `db:"bg_card"              json:"bg_card"`
	TextOnHeader string `db:"text_on_header"       json:"text_on_header"`
	BorderFocus  string `db:"border_focus"         json:"border_focus"`

	// Context-Specific Header Colors
	AccentHeader   string `db:"accent_header"        json:"accent_header"`
	StatusHPHeader string `db:"status_hp_header"     json:"status_hp_header"`

	// Interactive Elements
	BtnPrimaryBg     string `db:"btn_primary_bg"       json:"btn_primary_bg"`
	BtnPrimaryText   string `db:"btn_primary_text"     json:"btn_primary_text"`
	BtnSecondaryBg   string `db:"btn_secondary_bg"     json:"btn_secondary_bg"`
	BtnSecondaryText string `db:"btn_secondary_text"   json:"btn_secondary_text"`

	CheckboxBorder  string `db:"checkbox_border"      json:"checkbox_border"`
	CheckboxChecked string `db:"checkbox_checked"     json:"checkbox_checked"`

	// Layout & Geometry
	RadiusBase  string `db:"radius_base"          json:"radius_base"`
	PaddingCard string `db:"padding_card"         json:"padding_card"`
	GapGrid     string `db:"gap_grid"             json:"gap_grid"`

	// Advanced Branding
	BgImageURL      *string `db:"bg_image_url"         json:"bg_image_url"`
	FontHeading     *string `db:"font_heading"         json:"font_heading"`
	FontBody        *string `db:"font_body"            json:"font_body"`
	AccentSecondary *string `db:"accent_secondary"     json:"accent_secondary"`
	AccentRose      *string `db:"accent_rose"          json:"accent_rose"`

	CreatedAt time.Time `db:"created_at"           json:"created_at"`
	UpdatedAt time.Time `db:"updated_at"           json:"updated_at"`
}

type ThemeInput struct {
	Name string `json:"name"`

	// Colors
	BgPage    string `json:"bg_page"`
	BgHeader  string `json:"bg_header"`
	BgSurface string `json:"bg_surface"`

	TextMain      string `json:"text_main"`
	TextSecondary string `json:"text_secondary"`
	TextMuted     string `json:"text_muted"`
	TextOnAccent  string `json:"text_on_accent"`

	AccentPrimary      string `json:"accent_primary"`
	AccentPrimaryHover string `json:"accent_primary_hover"`
	BorderMain         string `json:"border_main"`

	StatusNM  string `json:"status_nm"`
	StatusLP  string `json:"status_lp"`
	StatusMP  string `json:"status_mp"`
	StatusHP  string `json:"status_hp"`
	StatusDMG string `json:"status_dmg"`

	// Specialized Colors
	BgCard       string `json:"bg_card"`
	TextOnHeader string `json:"text_on_header"`
	BorderFocus  string `json:"border_focus"`

	AccentHeader   string `json:"accent_header"`
	StatusHPHeader string `json:"status_hp_header"`

	// Interactive Elements
	BtnPrimaryBg     string `json:"btn_primary_bg"`
	BtnPrimaryText   string `json:"btn_primary_text"`
	BtnSecondaryBg   string `json:"btn_secondary_bg"`
	BtnSecondaryText string `json:"btn_secondary_text"`

	CheckboxBorder  string `json:"checkbox_border"`
	CheckboxChecked string `json:"checkbox_checked"`

	// Layout & Geometry
	RadiusBase  string `json:"radius_base"`
	PaddingCard string `json:"padding_card"`
	GapGrid     string `json:"gap_grid"`

	// Advanced Branding
	BgImageURL      *string `json:"bg_image_url,omitempty"`
	FontHeading     *string `json:"font_heading,omitempty"`
	FontBody        *string `json:"font_body,omitempty"`
	AccentSecondary *string `json:"accent_secondary,omitempty"`
	AccentRose      *string `json:"accent_rose,omitempty"`
}

type BulkMoveStorageRequest struct {
	TargetStorageID string            `json:"target_storage_id"`
	Moves           []MoveStorageItem `json:"moves"`
}

type MoveStorageItem struct {
	ProductID     string `json:"product_id"`
	FromStorageID string `json:"from_storage_id"`
	Quantity      int    `json:"quantity"`
}
