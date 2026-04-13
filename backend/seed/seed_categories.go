package main

import (
	"fmt"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// CategoryMap maps slug → UUID
type CategoryMap map[string]string

func seedCategories(db *sqlx.DB) (CategoryMap, error) {
	logger.Info("🏷️  Seeding custom categories...")

	type Cat struct {
		Name       string
		Slug       string
		BgColor    string
		TextColor  string
		Icon       string
		ShowBadge  bool
		Searchable bool
		IsActive   bool
	}

	cats := []Cat{
		{"Featured", "featured", "#d4af37", "#000000", "Star", true, true, true},
		{"Hot Items", "hot-items", "#ff4500", "#ffffff", "Flame", true, true, true},
		{"New Arrivals", "new-arrivals", "#1e90ff", "#ffffff", "Zap", true, true, true},
		{"Sale", "sale", "#32cd32", "#000000", "Tag", true, true, true},
		{"Staff Picks", "staff-picks", "#8b5cf6", "#ffffff", "Award", true, true, true},
		{"Commander Staples", "commander-staples", "#1a1f2e", "#d4af37", "Shield", true, true, true},
		{"Budget Builds", "budget-builds", "#059669", "#ffffff", "DollarSign", true, true, true},
		{"Tournament Ready", "tournament-ready", "#dc2626", "#ffffff", "Trophy", true, true, true},
	}

	result := make(CategoryMap)
	for _, cat := range cats {
		var id string
		err := db.QueryRow(`
			INSERT INTO custom_category (name, slug, bg_color, text_color, icon, show_badge, searchable, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name, bg_color = EXCLUDED.bg_color,
				text_color = EXCLUDED.text_color, icon = EXCLUDED.icon,
				show_badge = EXCLUDED.show_badge, searchable = EXCLUDED.searchable,
				is_active = EXCLUDED.is_active
			RETURNING id
		`, cat.Name, cat.Slug, cat.BgColor, cat.TextColor, cat.Icon, cat.ShowBadge, cat.Searchable, cat.IsActive).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to seed category '%s': %w", cat.Name, err)
		}
		result[cat.Slug] = id
	}
	logger.Info("✅ %d categories seeded", len(result))
	return result, nil
}
func loadCategories(db *sqlx.DB) (CategoryMap, error) {
	result := make(CategoryMap)
	rows, err := db.Queryx("SELECT slug, id FROM custom_category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c struct {
			Slug string `db:"slug"`
			ID   string `db:"id"`
		}
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
		result[c.Slug] = c.ID
	}
	return result, nil
}
