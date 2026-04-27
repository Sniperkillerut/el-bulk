package main

import (
	"fmt"
	"time"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedSettings(db *sqlx.DB) error {
	logger.Info("⚙️  Seeding settings...")

	settings := map[string]string{
		// Exchange rates
		"usd_to_cop_rate": "4450",
		"eur_to_cop_rate": "4800",

		// Contact info
		"contact_address":   "Cra. 15 # 76-54, Local 201, Centro Comercial Unilago, Bogotá",
		"contact_phone":     "+57 300 123 4567",
		"contact_email":     "admin@elbulk.com",
		"contact_instagram": "elbulktcg",
		"contact_hours":     "Mon - Sat: 11:00 AM - 7:00 PM",

		// Shipping
		"flat_shipping_fee_cop": "15000",

		// Theme
		"default_theme_id": "00000000-0000-0000-0000-000000000001",

		// Discovery algorithms
		"hot_sales_threshold": "5",
		"hot_days_threshold":  "30",
		"new_days_threshold":  "14",

		// Internationalization
		"default_locale":         "es",
		"hide_language_selector": "false",

		// Last set sync placeholder (updated after real sync)
		"last_set_sync": time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
	}

	for k, v := range settings {
		_, err := db.Exec(`
			INSERT INTO setting (key, value)
			VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
		`, k, v)
		if err != nil {
			return fmt.Errorf("failed to set '%s': %w", k, err)
		}
	}
	logger.Info("✅ %d settings seeded", len(settings))
	return nil
}
