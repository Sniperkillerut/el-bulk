package models

import "time"

// CustomCategory represents a user-defined collection for grouping products.
type CustomCategory struct {
	ID         string     `db:"id"          json:"id"`
	Name       string     `db:"name"        json:"name"`
	Slug       string     `db:"slug"        json:"slug"`
	IsActive   bool       `db:"is_active"   json:"is_active"`
	ShowBadge  bool       `db:"show_badge"  json:"show_badge"`
	Searchable bool       `db:"searchable"  json:"searchable"`
	BgColor    *string    `db:"bg_color"    json:"bg_color"`
	TextColor  *string    `db:"text_color"  json:"text_color"`
	Icon       *string    `db:"icon"        json:"icon"`
	CreatedAt  *time.Time `db:"created_at"  json:"created_at,omitempty"`
	ItemCount  int        `db:"item_count"  json:"item_count,omitempty"` // Computed field
	IsHot      bool       `json:"is_hot,omitempty"`
	IsNew      bool       `json:"is_new,omitempty"`
}

// Redact strips sensitive or internal fields for non-admin users.
func (c *CustomCategory) Redact(isAdmin bool) {
	if !isAdmin {
		c.CreatedAt = nil
		c.ItemCount = 0
		c.IsHot = false
		c.IsNew = false
	}
}

// CustomCategoryInput is used for creating/updating custom categories.
type CustomCategoryInput struct {
	ID         *string `json:"id,omitempty"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	IsActive   *bool   `json:"is_active"`
	ShowBadge  *bool   `json:"show_badge"`
	Searchable *bool   `json:"searchable"`
	BgColor    *string `json:"bg_color"`
	TextColor  *string `json:"text_color"`
	Icon       *string `json:"icon"`
}

// ProductCategory maps a product to a custom category in memory.
type ProductCategory struct {
	ProductID  string `db:"product_id"  json:"product_id"`
	CategoryID string `db:"category_id" json:"category_id"`
}
