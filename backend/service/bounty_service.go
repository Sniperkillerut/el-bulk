package service

import (
	"encoding/json"
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/crypto"
)

type BountyService struct {
	Store *store.BountyStore
}

func NewBountyService(s *store.BountyStore) *BountyService {
	return &BountyService{Store: s}
}

func (s *BountyService) ListBounties(activeParam string) ([]models.Bounty, error) {
	bounties, err := s.Store.ListBounties(activeParam)
	if err != nil {
		return nil, err
	}
	if bounties == nil {
		bounties = []models.Bounty{}
	}
	return bounties, nil
}

func (s *BountyService) CreateBounty(input models.BountyInput) (*models.Bounty, error) {
	if input.Name == "" || input.TCG == "" {
		return nil, fmt.Errorf("name and tcg are required")
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	return s.Store.CreateBounty(input, isActive)
}

func (s *BountyService) UpdateBounty(id string, input models.BountyInput) (*models.Bounty, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	return s.Store.UpdateBounty(id, input, isActive)
}

func (s *BountyService) DeleteBounty(id string) (int64, error) {
	return s.Store.DeleteBounty(id)
}

// ── Offers ──────────────────────────────────────────────

func (s *BountyService) ListOffers() ([]models.BountyOffer, error) {
	offers, err := s.Store.ListOffers()
	if err != nil {
		return nil, err
	}
	if offers == nil {
		offers = []models.BountyOffer{}
	}
	// Decrypt sensitive contact info for admin view
	for i := range offers {
		offers[i].CustomerContact = *crypto.DecryptSafe(&offers[i].CustomerContact)
	}
	return offers, nil
}

func (s *BountyService) SubmitOffer(input models.BountyOfferInput, userID *string) (*models.BountyOffer, error) {
	if input.BountyID == "" || input.CustomerName == "" || input.CustomerContact == "" {
		return nil, fmt.Errorf("BountyID, customer_name, and customer_contact are required")
	}
	if input.Quantity <= 0 {
		input.Quantity = 1
	}

	result, err := s.Store.SubmitOffer(
		input.BountyID, input.CustomerName, input.CustomerContact,
		input.Quantity, input.Condition, input.Notes, input.Status, userID,
	)
	if err != nil {
		return nil, err
	}

	var offer models.BountyOffer
	if err := json.Unmarshal(result, &offer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bounty offer: %w", err)
	}
	return &offer, nil
}

func (s *BountyService) UpdateOfferStatus(id, status string) (*models.BountyOffer, error) {
	offer, err := s.Store.UpdateOfferStatus(id, status)
	if err != nil {
		return nil, err
	}
	// Decrypt sensitive contact info
	offer.CustomerContact = *crypto.DecryptSafe(&offer.CustomerContact)
	return offer, nil
}

func (s *BountyService) ListMeOffers(userID string) ([]models.BountyOffer, error) {
	offers, err := s.Store.ListMeOffers(userID)
	if err != nil {
		return nil, err
	}
	if offers == nil {
		offers = []models.BountyOffer{}
	}
	return offers, nil
}

func (s *BountyService) CancelMeOffer(id, userID string) error {
	rows, err := s.Store.CancelMeOffer(id, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("offer cannot be cancelled")
	}
	return nil
}

// ── Client Requests ─────────────────────────────────────

func (s *BountyService) ListRequests() ([]models.ClientRequest, error) {
	requests, err := s.Store.ListRequests()
	if err != nil {
		return nil, err
	}
	if requests == nil {
		requests = []models.ClientRequest{}
	}
	// Decrypt sensitive contact info for admin view
	for i := range requests {
		requests[i].CustomerContact = *crypto.DecryptSafe(&requests[i].CustomerContact)
	}
	return requests, nil
}

func (s *BountyService) SubmitRequest(input models.ClientRequestInput, userID *string) (*models.ClientRequest, error) {
	if input.CustomerName == "" || input.CustomerContact == "" || input.CardName == "" {
		return nil, fmt.Errorf("customer_name, customer_contact, and card_name are required")
	}

	result, err := s.Store.SubmitRequest(
		input.CustomerName, input.CustomerContact, input.CardName,
		input.SetName, input.Details, userID,
	)
	if err != nil {
		return nil, err
	}

	var req models.ClientRequest
	if err := json.Unmarshal(result, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client request: %w", err)
	}
	return &req, nil
}

func (s *BountyService) UpdateRequestStatus(id, status string) (*models.ClientRequest, error) {
	return s.Store.UpdateRequestStatus(id, status)
}

func (s *BountyService) ListMeRequests(userID string) ([]models.ClientRequest, error) {
	requests, err := s.Store.ListMeRequests(userID)
	if err != nil {
		return nil, err
	}
	if requests == nil {
		requests = []models.ClientRequest{}
	}
	return requests, nil
}

func (s *BountyService) CancelMeRequest(id, userID string) error {
	rows, err := s.Store.CancelMeRequest(id, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("request cannot be cancelled")
	}
	return nil
}
