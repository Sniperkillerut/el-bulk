package models

import "time"

// CustomCategory represents a user-defined collection for grouping products.
type CustomCategory struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Slug      string    `db:"slug"       json:"slug"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ItemCount int       `db:"item_count" json:"item_count"` // Computed field
}

// CustomCategoryInput is used for creating/updating custom categories.
type CustomCategoryInput struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IsActive *bool  `json:"is_active"` // Use pointer to distinguish between false and missing
}

// ProductCategory maps a product to a custom category in memory.
type ProductCategory struct {
	ProductID  string `db:"product_id"  json:"product_id"`
	CategoryID string `db:"category_id" json:"category_id"`
}
