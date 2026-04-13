package store

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type AuthStore struct {
	DB *sqlx.DB
}

func NewAuthStore(db *sqlx.DB) *AuthStore {
	return &AuthStore{DB: db}
}

func (s *AuthStore) FindLinkedCustomerID(ctx context.Context, provider, providerID string) (string, error) {
	var id string
	err := s.DB.GetContext(ctx, &id, "SELECT customer_id FROM customer_auth WHERE provider = $1 AND provider_id = $2", provider, providerID)
	return id, err
}

func (s *AuthStore) CustomerExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := s.DB.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM customer WHERE id = $1)", id)
	return exists, err
}

func (s *AuthStore) LinkProvider(ctx context.Context, customerID, provider, providerID string) error {
	_, err := s.DB.ExecContext(ctx, "INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, $2, $3)", customerID, provider, providerID)
	return err
}

func (s *AuthStore) LinkProviderIfNotExists(ctx context.Context, customerID, provider, providerID string) error {
	_, err := s.DB.ExecContext(ctx, "INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", customerID, provider, providerID)
	return err
}

func (s *AuthStore) GetCustomerByID(ctx context.Context, id string) (*models.Customer, error) {
	var c models.Customer
	err := s.DB.GetContext(ctx, &c, "SELECT * FROM customer WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *AuthStore) GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error) {
	var c models.Customer
	err := s.DB.GetContext(ctx, &c, "SELECT * FROM customer WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *AuthStore) CreateCustomerWithAuth(ctx context.Context, firstName, lastName, email, avatarURL, provider, providerID string) (*models.Customer, error) {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var customer models.Customer
	err = tx.GetContext(ctx, &customer, `
		INSERT INTO customer (first_name, last_name, email, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, firstName, lastName, email, avatarURL)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO customer_auth (customer_id, provider, provider_id) VALUES ($1, $2, $3)", customer.ID, provider, providerID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &customer, nil
}

func (s *AuthStore) CleanOrphanAuth(ctx context.Context, provider, providerID string) {
	s.DB.ExecContext(ctx, "DELETE FROM customer_auth WHERE provider = $1 AND provider_id = $2", provider, providerID)
}

func (s *AuthStore) GetLinkedProviders(ctx context.Context, customerID string) ([]string, error) {
	var providers []string
	err := s.DB.SelectContext(ctx, &providers, "SELECT provider FROM customer_auth WHERE customer_id = $1", customerID)
	return providers, err
}

func (s *AuthStore) UpdateCustomerProfile(ctx context.Context, userID, encPhone, encIDNumber, encAddress string) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE customer SET phone = $1, id_number = $2, address = $3
		WHERE id = $4
	`, encPhone, encIDNumber, encAddress, userID)
	return err
}
