package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/jmoiron/sqlx"
)

type AccountingHandler struct {
	DB *sqlx.DB
}

func NewAccountingHandler(db *sqlx.DB) *AccountingHandler {
	return &AccountingHandler{DB: db}
}

func (h *AccountingHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var args []interface{}

	// Base queries
	// Orders: Income (One row per product)
	orderQuery := `
		SELECT 
			o.completed_at as date, 
			'Order Item' as type, 
			o.order_number as ref, 
			oi.product_name || ' (x' || oi.quantity || ') (Order ' || o.order_number || ')' as detail, 
			(oi.unit_price_cop * oi.quantity) as income, 
			0 as outcome, 
			'' as notes 
		FROM "order" o
		JOIN order_item oi ON o.id = oi.order_id
		JOIN customer c ON o.customer_id = c.id
		WHERE o.status = 'completed'
	`
	if startDate != "" {
		orderQuery += fmt.Sprintf(" AND o.completed_at >= $%d", len(args)+1)
		args = append(args, startDate)
	}
	if endDate != "" {
		orderQuery += fmt.Sprintf(" AND o.completed_at <= $%d", len(args)+1)
		args = append(args, endDate)
	}

	// Bounty Offers: Outcome (Expense)
	offerQuery := `
		SELECT 
			o.created_at as date, 
			'Bounty Offer' as type, 
			b.name as ref, 
			b.name || ' (from ' || c.first_name || ' ' || COALESCE(c.last_name, '') || ')' as detail, 
			0 as income, 
			COALESCE(b.target_price, 0) as outcome, 
			'pending confirmation' as notes 
		FROM bounty_offer o 
		JOIN bounty b ON o.bounty_id = b.id 
		JOIN customer c ON o.customer_id = c.id 
		WHERE o.status IN ('accepted', 'fulfilled')
	`
	if startDate != "" {
		offerQuery += fmt.Sprintf(" AND o.created_at >= $%d", len(args)+1)
		args = append(args, startDate)
	}
	if endDate != "" {
		offerQuery += fmt.Sprintf(" AND o.created_at <= $%d", len(args)+1)
		args = append(args, endDate)
	}

	// Client Requests: Income
	requestQuery := `
		SELECT 
			r.created_at as date, 
			'Client Request' as type, 
			r.card_name as ref, 
			r.card_name || ' (for ' || r.customer_name || ')' as detail, 
			COALESCE(b.target_price, 0) as income, 
			0 as outcome, 
			'pending confirmation' as notes 
		FROM client_request r 
		LEFT JOIN bounty b ON LOWER(r.card_name) = LOWER(b.name) 
		WHERE r.status = 'solved'
	`
	if startDate != "" {
		requestQuery += fmt.Sprintf(" AND r.created_at >= $%d", len(args)+1)
		args = append(args, startDate)
	}
	if endDate != "" {
		requestQuery += fmt.Sprintf(" AND r.created_at <= $%d", len(args)+1)
		args = append(args, endDate)
	}

	// Combine all queries with UNION ALL
	fullQuery := fmt.Sprintf(`
		SELECT * FROM (
			%s
			UNION ALL
			%s
			UNION ALL
			%s
		) AS combined
		ORDER BY date DESC
	`, orderQuery, offerQuery, requestQuery)

	type Row struct {
		Date    time.Time `db:"date"`
		Type    string    `db:"type"`
		Ref     string    `db:"ref"`
		Detail  string    `db:"detail"`
		Income  float64   `db:"income"`
		Outcome float64   `db:"outcome"`
		Notes   string    `db:"notes"`
	}

	var rows []Row
	err := h.DB.Select(&rows, fullQuery, args...)
	if err != nil {
		logger.Error("Failed to fetch accounting data: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Set headers for CSV download
	filename := fmt.Sprintf("accounting_export_%s.csv", time.Now().Format("2006-01-02"))
	if startDate != "" && endDate != "" {
		filename = fmt.Sprintf("accounting_%s_to_%s.csv", startDate, endDate)
	} else if startDate != "" {
		filename = fmt.Sprintf("accounting_from_%s.csv", startDate)
	} else if endDate != "" {
		filename = fmt.Sprintf("accounting_to_%s.csv", endDate)
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Date", "Type", "Reference", "Description", "Income (COP)", "Outcome (COP)", "Notes"})

	// Write rows
	for _, row := range rows {
		writer.Write([]string{
			row.Date.Format("2006-01-02 15:04:05"),
			row.Type,
			row.Ref,
			row.Detail,
			fmt.Sprintf("%.0f", row.Income),
			fmt.Sprintf("%.0f", row.Outcome),
			row.Notes,
		})
	}
}
