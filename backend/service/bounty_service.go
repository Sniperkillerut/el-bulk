package service

import (
	"context"
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

func (s *BountyService) ListBounties(ctx context.Context, activeParam string, isAdmin bool) ([]models.Bounty, error) {
	bounties, err := s.Store.ListBounties(ctx, activeParam)
	if err != nil {
		return nil, err
	}
	if bounties == nil {
		bounties = []models.Bounty{}
	}

	for i := range bounties {
		bounties[i].Redact(isAdmin)
	}

	return bounties, nil
}

func (s *BountyService) CreateBounty(ctx context.Context, input models.BountyInput) (*models.Bounty, error) {
	if input.Name == "" || input.TCG == "" {
		return nil, fmt.Errorf("name and tcg are required")
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	return s.Store.CreateBounty(ctx, input, isActive)
}

func (s *BountyService) UpdateBounty(ctx context.Context, id string, input models.BountyInput) (*models.Bounty, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	return s.Store.UpdateBounty(ctx, id, input, isActive)
}

func (s *BountyService) DeleteBounty(ctx context.Context, id string) (int64, error) {
	return s.Store.DeleteBounty(ctx, id)
}

// ── Offers ──────────────────────────────────────────────

func (s *BountyService) ListOffers(ctx context.Context) ([]models.BountyOffer, error) {
	offers, err := s.Store.ListOffers(ctx)
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

func (s *BountyService) SubmitOffer(ctx context.Context, input models.BountyOfferInput, userID *string) (*models.BountyOffer, error) {
	if input.BountyID == "" || input.CustomerName == "" || input.CustomerContact == "" {
		return nil, fmt.Errorf("BountyID, customer_name, and customer_contact are required")
	}
	if input.Quantity <= 0 {
		input.Quantity = 1
	}

	result, err := s.Store.SubmitOffer(
		ctx, input.BountyID, input.CustomerName, input.CustomerContact,
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

func (s *BountyService) UpdateOfferStatus(ctx context.Context, id, status string) (*models.BountyOffer, error) {
	offer, err := s.Store.UpdateOfferStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}
	// Decrypt sensitive contact info
	offer.CustomerContact = *crypto.DecryptSafe(&offer.CustomerContact)
	return offer, nil
}

func (s *BountyService) ListMeOffers(ctx context.Context, userID string) ([]models.BountyOffer, error) {
	offers, err := s.Store.ListMeOffers(ctx, userID)
	if err != nil {
		return nil, err
	}
	if offers == nil {
		offers = []models.BountyOffer{}
	}
	// Sanitize: user shouldn't see AdminNotes on their own offers unless intended
	// In this system, AdminNotes are internal.
	for i := range offers {
		offers[i].AdminNotes = nil
		// CustomerContact is already their own, but let's be consistent.
		// It's encrypted in DB, likely already decrypted or raw here.
		// ListMeOffers from store seems to return it.
	}
	return offers, nil
}

func (s *BountyService) CancelMeOffer(ctx context.Context, id, userID string) error {
	rows, err := s.Store.CancelMeOffer(ctx, id, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("offer cannot be cancelled")
	}
	return nil
}

// ── Client Requests ─────────────────────────────────────

func (s *BountyService) ListRequests(ctx context.Context) ([]models.ClientRequest, error) {
	requests, err := s.Store.ListRequests(ctx)
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

func (s *BountyService) SubmitRequest(ctx context.Context, input models.ClientRequestInput, userID *string) (*models.ClientRequest, error) {
	if input.CustomerName == "" || input.CustomerContact == "" || input.CardName == "" {
		return nil, fmt.Errorf("customer_name, customer_contact, and card_name are required")
	}

	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	if input.TCG == "" {
		input.TCG = "mtg"
	}

	result, err := s.Store.SubmitRequest(
		ctx, input.CustomerName, input.CustomerContact, input.CardName,
		input.SetName, input.Details, input.Quantity, input.TCG, userID,
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

func (s *BountyService) SubmitRequestsBatch(ctx context.Context, input models.ClientRequestBatchInput, userID *string) (*models.ClientRequestBatchResponse, error) {
	if input.CustomerName == "" || input.CustomerContact == "" || len(input.Cards) == 0 {
		return nil, fmt.Errorf("customer_name, customer_contact, and at least one card are required")
	}

	result, err := s.Store.SubmitRequestsBatch(ctx, input, userID)
	if err != nil {
		return nil, err
	}

	var resp models.ClientRequestBatchResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch response: %w", err)
	}
	return &resp, nil
}

func (s *BountyService) UpdateRequestStatus(ctx context.Context, id, status string) (*models.ClientRequest, error) {
	req, err := s.Store.UpdateRequestStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}
	// Decrypt sensitive contact info for admin view
	req.CustomerContact = *crypto.DecryptSafe(&req.CustomerContact)
	return req, nil
}

func (s *BountyService) ListMeRequests(ctx context.Context, userID string) ([]models.ClientRequest, error) {
	requests, err := s.Store.ListMeRequests(ctx, userID)
	if err != nil {
		return nil, err
	}
	if requests == nil {
		requests = []models.ClientRequest{}
	}
	// Sanitize: CreatedAt is already a pointer now, ensure it's handled if we want to hide it
	// For "Me" endpoints, timestamps are usually fine, but we'll be consistent if needed.
	return requests, nil
}

func (s *BountyService) CancelMeRequest(ctx context.Context, id, userID, reason string) error {
	rows, err := s.Store.CancelMeRequest(ctx, id, userID, reason)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("request cannot be cancelled (must be pending or accepted)")
	}
	return nil
}
