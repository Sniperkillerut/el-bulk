package models

import (
	"time"
)

type Bounty struct {
	ID              string        `db:"id"               json:"id"`
	Name            string        `db:"name"             json:"name"`
	TCG             string        `db:"tcg"              json:"tcg"`
	SetName         *string       `db:"set_name"         json:"set_name,omitempty"`
	Condition       *string       `db:"condition"        json:"condition,omitempty"`
	FoilTreatment   FoilTreatment `db:"foil_treatment"   json:"foil_treatment"`
	CardTreatment   CardTreatment `db:"card_treatment"   json:"card_treatment"`
	CollectorNumber *string       `db:"collector_number" json:"collector_number,omitempty"`
	PromoType       *string       `db:"promo_type"       json:"promo_type,omitempty"`
	Language        string        `db:"language"         json:"language"`
	TargetPrice     *float64      `db:"target_price"     json:"target_price,omitempty"`
	HidePrice       bool          `db:"hide_price"       json:"hide_price"`
	QuantityNeeded  int           `db:"quantity_needed"  json:"quantity_needed"`
	IsGeneric       bool          `db:"is_generic"       json:"is_generic"`
	ImageURL        *string       `db:"image_url"        json:"image_url,omitempty"`
	PriceSource     string        `db:"price_source"     json:"price_source,omitempty"`
	PriceReference  *float64      `db:"price_reference"  json:"price_reference,omitempty"`
	IsActive        bool          `db:"is_active"        json:"is_active"`
	CreatedAt       *time.Time    `db:"created_at"       json:"created_at,omitempty"`
	UpdatedAt       *time.Time    `db:"updated_at"       json:"updated_at,omitempty"`
}

// Redact strips sensitive or internal fields for non-admin users.
func (b *Bounty) Redact(isAdmin bool) {
	if !isAdmin {
		b.PriceSource = ""
		b.PriceReference = nil
		b.CreatedAt = nil
		b.UpdatedAt = nil
		if b.HidePrice {
			b.TargetPrice = nil
		}
	}
}

type BountyInput struct {
	Name            string        `json:"name"`
	TCG             string        `json:"tcg"`
	SetName         *string       `json:"set_name,omitempty"`
	Condition       *string       `json:"condition,omitempty"`
	FoilTreatment   FoilTreatment `json:"foil_treatment"`
	CardTreatment   CardTreatment `json:"card_treatment"`
	CollectorNumber *string       `json:"collector_number,omitempty"`
	PromoType       *string       `json:"promo_type,omitempty"`
	Language        string        `json:"language"`
	TargetPrice     *float64      `json:"target_price,omitempty"`
	HidePrice       bool          `json:"hide_price"`
	QuantityNeeded  int           `json:"quantity_needed"`
	IsGeneric       bool          `json:"is_generic"`
	ImageURL        *string       `json:"image_url,omitempty"`
	PriceSource     string        `json:"price_source"`
	PriceReference  *float64      `json:"price_reference,omitempty"`
	IsActive        *bool         `json:"is_active,omitempty"`
}

type BountyWithOffers struct {
	Bounty
	Offers []BountyOffer `json:"offers"`
}

type FulfillOfferInput struct {
	RequestIDs []string `json:"request_ids"`
}
