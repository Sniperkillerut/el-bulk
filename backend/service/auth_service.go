package service

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/crypto"
)

type AuthService struct {
	Store *store.AuthStore
}

func NewAuthService(s *store.AuthStore) *AuthService {
	return &AuthService{Store: s}
}

func (s *AuthService) FindLinkedCustomerID(ctx context.Context, provider, providerID string) (string, error) {
	return s.Store.FindLinkedCustomerID(ctx, provider, providerID)
}

func (s *AuthService) CustomerExists(ctx context.Context, id string) (bool, error) {
	return s.Store.CustomerExists(ctx, id)
}

func (s *AuthService) LinkProvider(ctx context.Context, customerID, provider, providerID string) error {
	return s.Store.LinkProvider(ctx, customerID, provider, providerID)
}

func (s *AuthService) LinkProviderIfNotExists(ctx context.Context, customerID, provider, providerID string) error {
	return s.Store.LinkProviderIfNotExists(ctx, customerID, provider, providerID)
}

func (s *AuthService) GetCustomerByID(ctx context.Context, id string) (*models.Customer, error) {
	return s.Store.GetCustomerByID(ctx, id)
}

func (s *AuthService) GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error) {
	return s.Store.GetCustomerByEmail(ctx, email)
}

func (s *AuthService) CreateCustomerWithAuth(ctx context.Context, firstName, lastName, email, avatarURL, provider, providerID string) (*models.Customer, error) {
	return s.Store.CreateCustomerWithAuth(ctx, firstName, lastName, email, avatarURL, provider, providerID)
}

func (s *AuthService) CleanOrphanAuth(ctx context.Context, provider, providerID string) {
	s.Store.CleanOrphanAuth(ctx, provider, providerID)
}

func (s *AuthService) GetLinkedProviders(ctx context.Context, customerID string) ([]string, error) {
	return s.Store.GetLinkedProviders(ctx, customerID)
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*models.Customer, error) {
	customer, err := s.Store.GetCustomerByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Fetch linked providers
	providers, err := s.Store.GetLinkedProviders(ctx, userID)
	if err != nil {
		providers = []string{}
	}
	customer.LinkedProviders = providers

	// Decrypt sensitive fields
	customer.Phone = crypto.DecryptSafe(customer.Phone)
	customer.IDNumber = crypto.DecryptSafe(customer.IDNumber)
	customer.Address = crypto.DecryptSafe(customer.Address)

	return customer, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID, phone, idNumber, address string) error {
	encPhone, _ := crypto.Encrypt(phone)
	encIDNumber, _ := crypto.Encrypt(idNumber)
	encAddress, _ := crypto.Encrypt(address)
	return s.Store.UpdateCustomerProfile(ctx, userID, encPhone, encIDNumber, encAddress)
}
