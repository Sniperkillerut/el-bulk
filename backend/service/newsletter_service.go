package service

import (
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type NewsletterService struct {
	Store *store.NewsletterStore
}

func NewNewsletterService(s *store.NewsletterStore) *NewsletterService {
	return &NewsletterService{Store: s}
}

func (s *NewsletterService) Subscribe(email string) (string, error) {
	count, err := s.Store.CountByEmail(email)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "Already subscribed", nil
	}

	// Link to existing customer if email matches
	customerID, _ := s.Store.FindCustomerIDByEmail(email)

	err = s.Store.Subscribe(email, customerID)
	if err != nil {
		return "", err
	}
	return "Subscribed successfully", nil
}

func (s *NewsletterService) Unsubscribe(email string) error {
	return s.Store.Unsubscribe(email)
}

func (s *NewsletterService) ListAll() ([]models.NewsletterSubscriber, error) {
	subs, err := s.Store.ListAll()
	if err != nil {
		return nil, err
	}
	if subs == nil {
		subs = []models.NewsletterSubscriber{}
	}
	return subs, nil
}
