package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/models"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type CustomerAdminHandler struct {
	DB *sqlx.DB
}

func (h *CustomerAdminHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	var customers []models.CustomerStats
	err := h.DB.Select(&customers, `
		SELECT 
			c.*,
			(SELECT COUNT(*) FROM "order" o WHERE o.customer_id = c.id) as order_count,
			(SELECT COALESCE(SUM(total_cop), 0) FROM "order" o WHERE o.customer_id = c.id) as total_spend,
			(SELECT EXISTS(SELECT 1 FROM newsletter_subscriber n WHERE n.customer_id = c.id OR n.email = c.email)) as is_subscriber
		FROM customer c
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(customers)
}

func (h *CustomerAdminHandler) GetCustomerDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var detail models.CustomerDetail
	err := h.DB.Get(&detail.Customer, "SELECT * FROM customer WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	// Fetch orders
	err = h.DB.Select(&detail.Orders, "SELECT * FROM \"order\" WHERE customer_id = $1 ORDER BY created_at DESC", id)
	if err != nil {
		http.Error(w, "Error fetching orders", http.StatusInternalServerError)
		return
	}

	// Fetch notes
	err = h.DB.Select(&detail.Notes, `
		SELECT n.*, a.username as admin_name
		FROM customer_note n
		LEFT JOIN admin a ON n.admin_id = a.id
		WHERE n.customer_id = $1
		ORDER BY n.created_at DESC
	`, id)
	if err != nil {
		http.Error(w, "Error fetching notes", http.StatusInternalServerError)
		return
	}

	// Fetch subscription status
	err = h.DB.Get(&detail.IsSubscriber, "SELECT EXISTS(SELECT 1 FROM newsletter_subscriber WHERE customer_id = $1 OR email = $2)", id, detail.Email)
	
	json.NewEncoder(w).Encode(detail)
}

func (h *CustomerAdminHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var input struct {
		Content string  `json:"content"`
		OrderID *string `json:"order_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Get admin ID from context or session (assuming it's there via middleware)
	// For now, we'll try to find an admin or leave as null if not strictly enforced
	adminID := r.Context().Value("admin_id")

	_, err := h.DB.Exec(`
		INSERT INTO customer_note (customer_id, order_id, content, admin_id)
		VALUES ($1, $2, $3, $4)
	`, id, input.OrderID, input.Content, adminID)

	if err != nil {
		http.Error(w, "Failed to add note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Note added successfully"})
}
