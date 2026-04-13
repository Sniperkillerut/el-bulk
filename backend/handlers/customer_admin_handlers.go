package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/crypto"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type CustomerAdminHandler struct {
	DB *sqlx.DB
}

func (h *CustomerAdminHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering CustomerAdminHandler.ListCustomers")
	var customers []models.CustomerStats
	err := h.DB.SelectContext(r.Context(), &customers, `
		SELECT 
			c.*,
			(SELECT COUNT(*) FROM "order" o WHERE o.customer_id = c.id) as order_count,
			(SELECT COALESCE(SUM(total_cop), 0) FROM "order" o WHERE o.customer_id = c.id) as total_spend,
			(SELECT EXISTS(SELECT 1 FROM newsletter_subscriber n WHERE n.customer_id = c.id OR n.email = c.email)) as is_subscriber,
			(SELECT content FROM customer_note n WHERE n.customer_id = c.id ORDER BY created_at DESC LIMIT 1) as latest_note,
			(SELECT COUNT(*) FROM client_request r WHERE r.customer_id = c.id) as request_count,
			(SELECT COUNT(*) FROM client_request r WHERE r.customer_id = c.id AND r.status IN ('pending', 'accepted')) as active_request_count,
			(SELECT COUNT(*) FROM bounty_offer bo WHERE bo.customer_id = c.id) as offer_count,
			(SELECT COUNT(*) FROM bounty_offer bo WHERE bo.customer_id = c.id AND bo.status IN ('pending', 'accepted')) as active_offer_count
		FROM customer c
		ORDER BY created_at DESC
	`)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to list customers: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if customers == nil {
		customers = []models.CustomerStats{}
	}

	// Decrypt sensitive fields for admin view
	for i := range customers {
		customers[i].Phone = crypto.DecryptSafe(customers[i].Phone)
		customers[i].IDNumber = crypto.DecryptSafe(customers[i].IDNumber)
		customers[i].Address = crypto.DecryptSafe(customers[i].Address)
	}

	render.Success(w, customers)
}

func (h *CustomerAdminHandler) GetCustomerDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering CustomerAdminHandler.GetCustomerDetail | ID: %s", id)
	if id == "" {
		render.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var detail models.CustomerDetail
	err := h.DB.GetContext(r.Context(), &detail.Customer, "SELECT * FROM customer WHERE id = $1", id)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Error fetching customer %s: %v", id, err)
		render.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	// Decrypt sensitive fields
	detail.Customer.Phone = crypto.DecryptSafe(detail.Customer.Phone)
	detail.Customer.IDNumber = crypto.DecryptSafe(detail.Customer.IDNumber)
	detail.Customer.Address = crypto.DecryptSafe(detail.Customer.Address)

	// Fetch orders
	err = h.DB.SelectContext(r.Context(), &detail.Orders, "SELECT * FROM \"order\" WHERE customer_id = $1 ORDER BY created_at DESC", id)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Error fetching orders for customer %s: %v", id, err)
		render.Error(w, "Error fetching orders", http.StatusInternalServerError)
		return
	}

	// Fetch notes
	err = h.DB.SelectContext(r.Context(), &detail.Notes, `
		SELECT n.*, a.username as admin_name
		FROM customer_note n
		LEFT JOIN admin a ON n.admin_id = a.id
		WHERE n.customer_id = $1
		ORDER BY n.created_at DESC
	`, id)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Error fetching notes for customer %s: %v", id, err)
		render.Error(w, "Error fetching notes", http.StatusInternalServerError)
		return
	}

	// Fetch subscription status
	_ = h.DB.GetContext(r.Context(), &detail.IsSubscriber, "SELECT EXISTS(SELECT 1 FROM newsletter_subscriber WHERE customer_id = $1 OR email = $2)", id, detail.Email)
	
	// Fetch requests
	err = h.DB.SelectContext(r.Context(), &detail.Requests, `SELECT * FROM client_request WHERE customer_id = $1 OR customer_contact = $2 ORDER BY created_at DESC`, id, detail.Email)
	if err != nil {
		detail.Requests = []models.ClientRequest{}
	}

	// Fetch bounty offers
	err = h.DB.SelectContext(r.Context(), &detail.Offers, `
		SELECT o.*, b.name as bounty_name
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		WHERE o.customer_id = $1
		ORDER BY o.created_at DESC
	`, id)
	if err != nil {
		detail.Offers = []models.BountyOffer{}
	}

	if detail.Orders == nil {
		detail.Orders = []models.Order{}
	}
	if detail.Notes == nil {
		detail.Notes = []models.CustomerNote{}
	}
	if detail.Requests == nil {
		detail.Requests = []models.ClientRequest{}
	}
	if detail.Offers == nil {
		detail.Offers = []models.BountyOffer{}
	}

	render.Success(w, detail)
}

func (h *CustomerAdminHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering CustomerAdminHandler.AddNote | CustomerID: %s", id)
	if id == "" {
		render.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var input struct {
		Content string  `json:"content"`
		OrderID *string `json:"order_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		render.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Get admin ID from context (set by AdminAuth middleware)
	adminID := r.Context().Value(middleware.AdminContextKey)

	_, err := h.DB.ExecContext(r.Context(), `
		INSERT INTO customer_note (customer_id, order_id, content, admin_id)
		VALUES ($1, $2, $3, $4)
	`, id, input.OrderID, input.Content, adminID)

	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to add note for customer %s: %v", id, err)
		render.Error(w, "Failed to add note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, map[string]string{"message": "Note added successfully"})
}
