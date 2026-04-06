package models

import "time"

// CustomCategory represents a user-defined collection for grouping products.
type CustomCategory struct {
	ID         string    `db:"id"          json:"id"`
	Name       string    `db:"name"        json:"name"`
	Slug       string    `db:"slug"        json:"slug"`
	IsActive   bool      `db:"is_active"   json:"is_active"`
	ShowBadge  bool      `db:"show_badge"  json:"show_badge"`
	Searchable bool      `db:"searchable"  json:"searchable"`
	BgColor    *string   `db:"bg_color"    json:"bg_color"`
	TextColor  *string   `db:"text_color"  json:"text_color"`
	Icon       *string   `db:"icon"        json:"icon"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
	ItemCount  int       `db:"item_count"  json:"item_count"` // Computed field
	IsHot      bool      `json:"is_hot"`
	IsNew      bool      `json:"is_new"`
}

// CustomCategoryInput is used for creating/updating custom categories.
type CustomCategoryInput struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	IsActive   *bool  `json:"is_active"`
	ShowBadge  *bool  `json:"show_badge"`
	Searchable *bool  `json:"searchable"`
	BgColor    *string `json:"bg_color"`
	TextColor  *string `json:"text_color"`
	Icon       *string `json:"icon"`
}

// ProductCategory maps a product to a custom category in memory.
type ProductCategory struct {
	ProductID  string `db:"product_id"  json:"product_id"`
	CategoryID string `db:"category_id" json:"category_id"`
}
