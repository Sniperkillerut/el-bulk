package store

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type NewsletterStore struct {
	DB *sqlx.DB
}

func NewNewsletterStore(db *sqlx.DB) *NewsletterStore {
	return &NewsletterStore{DB: db}
}

func (s *NewsletterStore) CountByEmail(ctx context.Context, email string) (int, error) {
	var count int
	err := s.DB.GetContext(ctx, &count, "SELECT COUNT(*) FROM newsletter_subscriber WHERE email = $1", email)
	return count, err
}

func (s *NewsletterStore) FindCustomerIDByEmail(ctx context.Context, email string) (*string, error) {
	var foundID string
	err := s.DB.GetContext(ctx, &foundID, "SELECT id FROM customer WHERE email = $1 LIMIT 1", email)
	if err != nil {
		return nil, err
	}
	return &foundID, nil
}

func (s *NewsletterStore) Subscribe(ctx context.Context, email string, customerID *string) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO newsletter_subscriber (email, customer_id) VALUES ($1, $2)`, email, customerID)
	return err
}

func (s *NewsletterStore) Unsubscribe(ctx context.Context, email string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM newsletter_subscriber WHERE email = $1", email)
	return err
}

func (s *NewsletterStore) ListAll(ctx context.Context) ([]models.NewsletterSubscriber, error) {
	var subscribers []models.NewsletterSubscriber
	err := s.DB.SelectContext(ctx, &subscribers, `
		SELECT n.*, c.first_name, c.last_name
		FROM newsletter_subscriber n
		LEFT JOIN customer c ON n.customer_id = c.id
		ORDER BY n.created_at DESC
	`)
	return subscribers, err
}
