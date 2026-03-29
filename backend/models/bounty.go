package models

import (
	"time"
)

type Bounty struct {
	ID             string        `db:"id"               json:"id"`
	Name           string        `db:"name"             json:"name"`
	TCG            string        `db:"tcg"              json:"tcg"`
	SetName        *string       `db:"set_name"         json:"set_name,omitempty"`
	Condition      *string       `db:"condition"        json:"condition,omitempty"`
	FoilTreatment  FoilTreatment `db:"foil_treatment"   json:"foil_treatment"`
	CardTreatment  CardTreatment `db:"card_treatment"   json:"card_treatment"`
	CollectorNumber *string      `db:"collector_number" json:"collector_number,omitempty"`
	PromoType      *string       `db:"promo_type"       json:"promo_type,omitempty"`
	Language       string        `db:"language"         json:"language"`
	TargetPrice    *float64      `db:"target_price"     json:"target_price,omitempty"`
	HidePrice      bool          `db:"hide_price"       json:"hide_price"`
	QuantityNeeded int           `db:"quantity_needed"  json:"quantity_needed"`
	ImageURL       *string       `db:"image_url"        json:"image_url,omitempty"`
	CreatedAt      time.Time     `db:"created_at"       json:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"       json:"updated_at"`
}

type BountyInput struct {
	Name           string        `json:"name"`
	TCG            string        `json:"tcg"`
	SetName        *string       `json:"set_name,omitempty"`
	Condition      *string       `json:"condition,omitempty"`
	FoilTreatment  FoilTreatment `json:"foil_treatment"`
	CardTreatment  CardTreatment `json:"card_treatment"`
	CollectorNumber *string      `json:"collector_number,omitempty"`
	PromoType      *string       `json:"promo_type,omitempty"`
	Language       string        `json:"language"`
	TargetPrice    *float64      `json:"target_price,omitempty"`
	HidePrice      bool          `json:"hide_price"`
	QuantityNeeded int           `json:"quantity_needed"`
	ImageURL       *string       `json:"image_url,omitempty"`
}

type ClientRequest struct {
	ID              string    `db:"id"               json:"id"`
	CustomerName    string    `db:"customer_name"    json:"customer_name"`
	CustomerContact string    `db:"customer_contact" json:"customer_contact"`
	CardName        string    `db:"card_name"        json:"card_name"`
	SetName         *string   `db:"set_name"         json:"set_name,omitempty"`
	Details         *string   `db:"details"          json:"details,omitempty"`
	Status          string    `db:"status"           json:"status"`
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
}

type ClientRequestInput struct {
	CustomerName    string  `json:"customer_name"`
	CustomerContact string  `json:"customer_contact"`
	CardName        string  `json:"card_name"`
	SetName         *string `json:"set_name,omitempty"`
	Details         *string `json:"details,omitempty"`
}

type UpdateClientRequestStatusInput struct {
	Status string `json:"status"` // expected: 'pending', 'accepted', 'rejected'
}

type BountyOffer struct {
	ID              string    `db:"id"               json:"id"`
	BountyID        string    `db:"bounty_id"        json:"bounty_id"`
	CustomerName    string    `db:"customer_name"    json:"customer_name"`
	CustomerContact string    `db:"customer_contact" json:"customer_contact"`
	Condition       *string   `db:"condition"        json:"condition,omitempty"`
	Status          string    `db:"status"           json:"status"`
	Notes           *string   `db:"notes"            json:"notes,omitempty"`
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"       json:"updated_at"`
}

type BountyOfferInput struct {
	BountyID        string  `json:"bounty_id"`
	CustomerName    string  `json:"customer_name"`
	CustomerContact string  `json:"customer_contact"`
	Condition       *string `json:"condition,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

type UpdateBountyOfferStatusInput struct {
	Status string `json:"status"` // expected: 'pending', 'reviewed', 'accepted', 'rejected'
}

type BountyWithOffers struct {
	Bounty
	Offers []BountyOffer `json:"offers"`
}
