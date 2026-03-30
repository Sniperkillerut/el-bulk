package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type NewsletterHandler struct {
	DB *sqlx.DB
}

func (h *NewsletterHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if input.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// 1. Check if subscriber already exists
	var count int
	err := h.DB.Get(&count, "SELECT COUNT(*) FROM newsletter_subscriber WHERE email = $1", input.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If already subscribed, return success anyway to avoid email enumeration
	if count > 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Already subscribed"})
		return
	}

	// 2. Check if this email belongs to an existing customer
	var customerID *string
	var foundID string
	err = h.DB.Get(&foundID, "SELECT id FROM customer WHERE email = $1 LIMIT 1", input.Email)
	if err == nil {
		customerID = &foundID
	}

	// 3. Create subscriber
	_, err = h.DB.Exec(`
		INSERT INTO newsletter_subscriber (email, customer_id)
		VALUES ($1, $2)
	`, input.Email, customerID)

	if err != nil {
		http.Error(w, "Failed to subscribe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscribed successfully"})
}

func (h *NewsletterHandler) AdminGetSubscribers(w http.ResponseWriter, r *http.Request) {
	var subscribers []models.NewsletterSubscriber
	err := h.DB.Select(&subscribers, `
		SELECT n.*, c.first_name, c.last_name
		FROM newsletter_subscriber n
		LEFT JOIN customer c ON n.customer_id = c.id
		ORDER BY n.created_at DESC
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if subscribers == nil {
		subscribers = []models.NewsletterSubscriber{}
	}

	json.NewEncoder(w).Encode(subscribers)
}

func (h *NewsletterHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("DELETE FROM newsletter_subscriber WHERE email = $1", input.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Unsubscribed successfully"})
}
