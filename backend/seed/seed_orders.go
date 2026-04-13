package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedOrders(db *sqlx.DB, customers []CustomerSeeded, allProductIDs []string) error {
	logger.Info("🛒 Seeding orders (all statuses, all payment methods, varied datetimes)...")

	if len(allProductIDs) == 0 || len(customers) == 0 {
		logger.Warn("⚠️ No products or customers available — skipping order seeding")
		return nil
	}

	// Cache product info to avoid repeated queries
	type ProductInfo struct {
		ID       string
		Name     string
		Set      string
		Price    float64
		Foil     string
		Treat    string
		Cond     string
	}

	productCache := make(map[string]ProductInfo)
	for _, pID := range allProductIDs {
		var p struct {
			ID              string   `db:"id"`
			Name            string   `db:"name"`
			SetName         *string  `db:"set_name"`
			PriceCOPOverride *float64 `db:"price_cop_override"`
			FoilTreatment   string   `db:"foil_treatment"`
			CardTreatment   string   `db:"card_treatment"`
			Condition       *string  `db:"condition"`
		}
		if err := db.Get(&p, `SELECT id, name, set_name, price_cop_override, foil_treatment, card_treatment, condition FROM product WHERE id = $1`, pID); err == nil {
			setName := ""
			if p.SetName != nil {
				setName = *p.SetName
			}
			price := float64(randInt(10, 300)) * 1000
			if p.PriceCOPOverride != nil {
				price = *p.PriceCOPOverride
			}
			cond := ""
			if p.Condition != nil {
				cond = *p.Condition
			}
			productCache[pID] = ProductInfo{
				ID: pID, Name: p.Name, Set: setName,
				Price: price, Foil: p.FoilTreatment, Treat: p.CardTreatment, Cond: cond,
			}
		}
	}

	type OrderSpec struct {
		Status        string
		PaymentMethod string
		IsLocalPickup bool
		Notes         string
		DaysAgoMin    int
		DaysAgoMax    int
		HasTracking   bool
	}

	// Define spread of order types
	orderTemplates := []OrderSpec{
		// Completed orders (most common)
		{"completed", "Nequi", false, "", 60, 90, false},
		{"completed", "Transfer", false, "Gracias por la compra!", 45, 75, false},
		{"completed", "Cash", true, "Cliente recogió en tienda.", 30, 60, false},
		{"completed", "Daviplata", false, "", 20, 50, false},
		{"completed", "Nequi", false, "", 15, 45, false},
		{"completed", "Transfer", true, "Local pickup. Pagó en efectivo.", 10, 30, false},
		// Shipped orders
		{"shipped", "Transfer", false, "", 5, 20, true},
		{"shipped", "Transfer", false, "Pedido especial — prioridad alta.", 3, 15, true},
		{"shipped", "Nequi", false, "", 7, 18, true},
		{"shipped", "Daviplata", false, "", 4, 12, true},
		// Ready for pickup
		{"ready_for_pickup", "Cash", true, "El cliente confirmará hora.", 1, 5, false},
		{"ready_for_pickup", "Nequi", true, "", 2, 7, false},
		// Confirmed (paid, processing)
		{"confirmed", "Transfer", false, "", 1, 4, false},
		{"confirmed", "Nequi", false, "Esperando envío.", 1, 3, false},
		{"confirmed", "Cash", true, "", 2, 5, false},
		// Pending orders
		{"pending", "Transfer", false, "", 0, 2, false},
		{"pending", "Nequi", false, "Cliente dice que pagó ayer.", 0, 1, false},
		{"pending", "Cash", true, "", 0, 2, false},
		// Cancelled orders (with inventory_restored)
		{"cancelled", "Transfer", false, "Cliente canceló por cambio de planes.", 10, 40, false},
		{"cancelled", "Nequi", false, "No recibimos el pago a tiempo.", 15, 50, false},
	}

	orderCounter := 1
	var shippingCarriers = []struct{ num, url string }{
		{"TCC-7812341", "https://tcc.com.co/tracking?n=TCC-7812341"},
		{"COORDINADORA-99123", "https://coordinadora.com/track/99123"},
		{"SERVIENTREGA-456781", "https://www.servientrega.com/tracking/456781"},
		{"INTERRAPIDISIMO-892312", "https://interrapidisimo.com/rastreo?guia=892312"},
	}

	for i, c := range customers {
		// VIP customers (first 5) get 8-15 orders; regular 2-5; guests 1-3
		numOrders := 2
		if i < 5 {
			numOrders = randInt(8, 15)
		} else if i < 15 {
			numOrders = randInt(2, 5)
		} else {
			numOrders = randInt(1, 3)
		}

		for j := 0; j < numOrders; j++ {
			tmpl := orderTemplates[(i+j)%len(orderTemplates)]
			orderNum := fmt.Sprintf("EB-%06d", orderCounter)
			orderCounter++

			daysBack := randInt(tmpl.DaysAgoMin, tmpl.DaysAgoMax)
			createdAt := daysAgo(daysBack)

			var confirmedAt *time.Time
			var completedAt *time.Time
			if tmpl.Status == "confirmed" || tmpl.Status == "completed" ||
				tmpl.Status == "shipped" || tmpl.Status == "ready_for_pickup" {
				t := createdAt.Add(time.Duration(randInt(6, 48)) * time.Hour)
				confirmedAt = &t
			}
			if tmpl.Status == "completed" {
				t := confirmedAt.Add(time.Duration(randInt(24, 120)) * time.Hour)
				completedAt = &t
			}

			shippingCOP := 0.0
			if !tmpl.IsLocalPickup {
				shippingCOP = 15000.0
			}

			var trackNumber, trackURL *string
			if tmpl.HasTracking {
				carrier := shippingCarriers[j%len(shippingCarriers)]
				trackNumber = &carrier.num
				trackURL = &carrier.url
			}

			inventoryRestored := tmpl.Status == "cancelled"

			var notes *string
			if tmpl.Notes != "" {
				notes = &tmpl.Notes
			}

			var oID string
			err := db.QueryRow(`
				INSERT INTO "order" (
					order_number, customer_id, status, payment_method,
					subtotal_cop, shipping_cop, tax_cop, total_cop,
					tracking_number, tracking_url,
					is_local_pickup, inventory_restored, notes,
					created_at, confirmed_at, completed_at
				) VALUES ($1,$2,$3,$4, $5,$6,$7,$8, $9,$10, $11,$12,$13, $14,$15,$16)
				RETURNING id
			`,
				orderNum, c.ID, tmpl.Status, tmpl.PaymentMethod,
				0.0, shippingCOP, 0.0, 0.0,
				trackNumber, trackURL,
				tmpl.IsLocalPickup, inventoryRestored, notes,
				createdAt, confirmedAt, completedAt,
			).Scan(&oID)
			if err != nil {
				return fmt.Errorf("failed to create order %s: %w", orderNum, err)
			}

			// Seed 1-4 order items
			numItems := randInt(1, 4)
			subtotal := 0.0
			for k := 0; k < numItems; k++ {
				pIdx := (i*17 + j*7 + k*3) % len(allProductIDs)
				pID := allProductIDs[pIdx]
				prod, ok := productCache[pID]
				if !ok {
					continue
				}

				qty := randInt(1, 2)
				unitPrice := prod.Price

				// Build stored_in_snapshot JSONB
				snapshot := []map[string]interface{}{
					{"name": "Showcase A", "quantity": qty},
				}
				snapshotJSON, _ := json.Marshal(snapshot)

				var foilTreat, cardTreat, cond *string
				if prod.Foil != "" {
					foilTreat = &prod.Foil
				}
				if prod.Treat != "" {
					cardTreat = &prod.Treat
				}
				if prod.Cond != "" {
					cond = &prod.Cond
				}

				db.Exec(`
					INSERT INTO order_item (
						order_id, product_id, product_name, product_set,
						foil_treatment, card_treatment, condition,
						unit_price_cop, quantity, stored_in_snapshot
					) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
				`, oID, pID, prod.Name, prod.Set,
					foilTreat, cardTreat, cond,
					unitPrice, qty, snapshotJSON)

				subtotal += unitPrice * float64(qty)
			}

			taxCOP := 0.0 // Colombia: no GST/VAT on used cards typically
			totalCOP := subtotal + shippingCOP + taxCOP

			db.Exec(`
				UPDATE "order"
				SET subtotal_cop=$1, tax_cop=$2, total_cop=$3
				WHERE id=$4
			`, subtotal, taxCOP, totalCOP, oID)
		}
	}

	logger.Info("✅ Orders seeded (counter reached %d)", orderCounter-1)
	return nil
}
