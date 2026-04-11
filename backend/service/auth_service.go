package service

import (
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

func (s *AuthService) FindLinkedCustomerID(provider, providerID string) (string, error) {
	return s.Store.FindLinkedCustomerID(provider, providerID)
}

func (s *AuthService) CustomerExists(id string) (bool, error) {
	return s.Store.CustomerExists(id)
}

func (s *AuthService) LinkProvider(customerID, provider, providerID string) error {
	return s.Store.LinkProvider(customerID, provider, providerID)
}

func (s *AuthService) LinkProviderIfNotExists(customerID, provider, providerID string) error {
	return s.Store.LinkProviderIfNotExists(customerID, provider, providerID)
}

func (s *AuthService) GetCustomerByID(id string) (*models.Customer, error) {
	return s.Store.GetCustomerByID(id)
}

func (s *AuthService) GetCustomerByEmail(email string) (*models.Customer, error) {
	return s.Store.GetCustomerByEmail(email)
}

func (s *AuthService) CreateCustomerWithAuth(firstName, lastName, email, avatarURL, provider, providerID string) (*models.Customer, error) {
	return s.Store.CreateCustomerWithAuth(firstName, lastName, email, avatarURL, provider, providerID)
}

func (s *AuthService) CleanOrphanAuth(provider, providerID string) {
	s.Store.CleanOrphanAuth(provider, providerID)
}

func (s *AuthService) GetLinkedProviders(customerID string) ([]string, error) {
	return s.Store.GetLinkedProviders(customerID)
}

func (s *AuthService) GetMe(userID string) (*models.Customer, error) {
	customer, err := s.Store.GetCustomerByID(userID)
	if err != nil {
		return nil, err
	}
	// Fetch linked providers
	providers, err := s.Store.GetLinkedProviders(userID)
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

func (s *AuthService) UpdateProfile(userID, phone, idNumber, address string) error {
	encPhone, _ := crypto.Encrypt(phone)
	encIDNumber, _ := crypto.Encrypt(idNumber)
	encAddress, _ := crypto.Encrypt(address)
	return s.Store.UpdateCustomerProfile(userID, encPhone, encIDNumber, encAddress)
}
