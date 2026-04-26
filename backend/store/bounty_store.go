package store

import (
	"context"
	"encoding/json"

	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type BountyStore struct {
	*BaseStore[models.Bounty]
}

func NewBountyStore(db *sqlx.DB) *BountyStore {
	return &BountyStore{
		BaseStore: NewBaseStore[models.Bounty](db, "bounty"),
	}
}

func (s *BountyStore) ListBounties(ctx context.Context, activeParam string) ([]models.Bounty, error) {
	bounties := []models.Bounty{}
	query := `
		SELECT 
			b.id, b.name, b.tcg, b.set_name, b.condition, b.foil_treatment, b.card_treatment, b.collector_number, b.promo_type, b.language, b.target_price, 
			b.hide_price, b.quantity_needed, b.is_generic, b.image_url, b.price_source, b.price_reference, b.is_active, b.created_at, b.updated_at
		FROM bounty b
		WHERE ($1 = '' OR b.is_active = ($1 = 'true'))
		ORDER BY b.created_at DESC
	`
	err := s.DB.SelectContext(ctx, &bounties, query, activeParam)
	return bounties, err
}

func (s *BountyStore) CreateBounty(ctx context.Context, input models.BountyInput, isActive bool) (*models.Bounty, error) {
	query := `
		INSERT INTO bounty (name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, is_generic, image_url, price_source, price_reference, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, is_generic, image_url, price_source, price_reference, is_active, created_at, updated_at
	`
	var bounty models.Bounty
	err := s.DB.QueryRowxContext(ctx, query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.CardTreatment, input.CollectorNumber, input.PromoType, input.Language,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.IsGeneric, input.ImageURL,
		input.PriceSource, input.PriceReference, isActive,
	).StructScan(&bounty)
	if err != nil {
		return nil, err
	}
	return &bounty, nil
}

func (s *BountyStore) UpdateBounty(ctx context.Context, id string, input models.BountyInput, isActive bool) (*models.Bounty, error) {
	query := `
		UPDATE bounty
		SET name = $1, tcg = $2, set_name = $3, condition = $4, foil_treatment = $5,
		    card_treatment = $6, collector_number = $7, promo_type = $8, language = $9,
		    target_price = $10, hide_price = $11, quantity_needed = $12, is_generic = $13, image_url = $14,
		    price_source = $15, price_reference = $16, is_active = $17, updated_at = now()
		WHERE id = $18
		RETURNING id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, is_generic, image_url, price_source, price_reference, is_active, created_at, updated_at
	`
	var bounty models.Bounty
	err := s.DB.QueryRowxContext(ctx, query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.CardTreatment, input.CollectorNumber, input.PromoType, input.Language,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.IsGeneric, input.ImageURL,
		input.PriceSource, input.PriceReference, isActive, id,
	).StructScan(&bounty)
	if err != nil {
		return nil, err
	}
	return &bounty, nil
}

func (s *BountyStore) DeleteBounty(ctx context.Context, id string) (int64, error) {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM bounty WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── Offers ──────────────────────────────────────────────

func (s *BountyStore) ListOffers(ctx context.Context) ([]models.BountyOffer, error) {
	offers := []models.BountyOffer{}
	query := `
		SELECT 
			o.id, o.bounty_id, o.customer_id, o.quantity, o.condition, o.status, o.notes, o.admin_notes, o.created_at, o.updated_at,
			b.name as bounty_name,
			c.first_name || ' ' || COALESCE(c.last_name, '') as customer_name,
			COALESCE(c.phone, c.email) as customer_contact
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		JOIN customer c ON o.customer_id = c.id
		ORDER BY o.created_at DESC
	`
	err := s.DB.SelectContext(ctx, &offers, query)
	return offers, err
}

func (s *BountyStore) SubmitOffer(ctx context.Context, bountyID, customerName, customerContact string, quantity int, condition, notes, status *string, userID *string) ([]byte, error) {
	var result []byte
	err := s.DB.GetContext(ctx, &result, "SELECT fn_submit_bounty_offer($1, $2, $3, $4, $5, $6, $7, $8)",
		bountyID, customerName, customerContact, quantity, condition, notes, status, userID)
	return result, err
}

func (s *BountyStore) UpdateOfferStatus(ctx context.Context, id, status string) (*models.BountyOffer, error) {
	_, err := s.DB.ExecContext(ctx, `UPDATE bounty_offer SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	if err != nil {
		return nil, err
	}

	var offer models.BountyOffer
	selectQuery := `
		SELECT 
			o.id, o.bounty_id, o.customer_id, o.quantity, o.condition, o.status, o.notes, o.admin_notes, o.created_at, o.updated_at,
			b.name as bounty_name,
			c.first_name || ' ' || COALESCE(c.last_name, '') as customer_name,
			COALESCE(c.phone, c.email) as customer_contact
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		JOIN customer c ON o.customer_id = c.id
		WHERE o.id = $1
	`
	err = s.DB.GetContext(ctx, &offer, selectQuery, id)
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (s *BountyStore) ListMeOffers(ctx context.Context, userID string) ([]models.BountyOffer, error) {
	offers := []models.BountyOffer{}
	query := `
		SELECT 
			o.id, o.bounty_id, o.customer_id, o.quantity, o.condition, o.status, o.notes, o.admin_notes, o.created_at, o.updated_at,
			b.name as bounty_name
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		WHERE o.customer_id = $1
		ORDER BY o.created_at DESC
	`
	err := s.DB.SelectContext(ctx, &offers, query, userID)
	return offers, err
}

func (s *BountyStore) CancelMeOffer(ctx context.Context, id, userID string) (int64, error) {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE bounty_offer 
		SET status = 'cancelled', updated_at = now() 
		WHERE id = $1 AND customer_id = $2 AND status = 'pending'
	`, id, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── Client Requests ─────────────────────────────────────

const requestColumns = `id, customer_id, customer_name, customer_contact, card_name, set_name, details, quantity, tcg, status, cancellation_reason, bounty_id, match_type, scryfall_id, created_at`

func (s *BountyStore) ListRequests(ctx context.Context) ([]models.ClientRequest, error) {
	requests := []models.ClientRequest{}
	query := `SELECT ` + requestColumns + ` FROM client_request ORDER BY created_at DESC`
	err := s.DB.SelectContext(ctx, &requests, query)
	return requests, err
}

func (s *BountyStore) SubmitRequest(ctx context.Context, customerName, customerContact, cardName string, setName, details *string, quantity int, tcg string, userID *string, matchType string, scryfallID *string) ([]byte, error) {
	var result []byte
	err := s.DB.GetContext(ctx, &result, "SELECT fn_submit_client_request($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		customerName, customerContact, cardName, setName, details, quantity, tcg, userID, matchType, scryfallID)
	return result, err
}

func (s *BountyStore) SubmitRequestsBatch(ctx context.Context, input models.ClientRequestBatchInput, userID *string) ([]byte, error) {
	cardsJSON, err := json.Marshal(input.Cards)
	if err != nil {
		return nil, err
	}
	var result []byte
	err = s.DB.GetContext(ctx, &result, "SELECT fn_submit_client_requests_batch($1::TEXT, $2::TEXT, $3::JSONB, $4::UUID)",
		input.CustomerName, input.CustomerContact, cardsJSON, userID)
	return result, err
}

func (s *BountyStore) UpdateRequestStatus(ctx context.Context, id, status string) (*models.ClientRequest, error) {
	query := `
		UPDATE client_request
		SET status = $1
		WHERE id = $2
		RETURNING ` + requestColumns
	var req models.ClientRequest
	err := s.DB.QueryRowxContext(ctx, query, status, id).StructScan(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *BountyStore) ListMeRequests(ctx context.Context, userID string) ([]models.ClientRequest, error) {
	requests := []models.ClientRequest{}
	query := `SELECT ` + requestColumns + ` FROM client_request WHERE customer_id = $1 ORDER BY created_at DESC`
	err := s.DB.SelectContext(ctx, &requests, query, userID)
	return requests, err
}

func (s *BountyStore) CancelMeRequest(ctx context.Context, id, userID, reason string) (int64, error) {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE client_request 
		SET status = 'not_needed', cancellation_reason = $1 
		WHERE id = $2 AND customer_id = $3 AND (status = 'pending' OR status = 'accepted')
	`, reason, id, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── New Bridge Methods ───────────────────────────────────

// AcceptRequest atomically finds/creates a bounty, links the request, and marks it accepted.
func (s *BountyStore) AcceptRequest(ctx context.Context, requestID string) (map[string]interface{}, error) {
	var result []byte
	err := s.DB.GetContext(ctx, &result, "SELECT fn_accept_client_request($1::UUID)", requestID)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FulfillOffer atomically accepts an offer and marks selected requests as solved.
func (s *BountyStore) FulfillOffer(ctx context.Context, offerID string, requestIDs []string) (map[string]interface{}, error) {
	var result []byte
	err := s.DB.GetContext(ctx, &result, "SELECT fn_fulfill_bounty_offer($1::UUID, $2::UUID[])", offerID, pq.Array(requestIDs))
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListRequestsByBounty returns all active requests linked to a specific bounty.
func (s *BountyStore) ListRequestsByBounty(ctx context.Context, bountyID string) ([]models.ClientRequest, error) {
	requests := []models.ClientRequest{}
	query := `SELECT ` + requestColumns + ` FROM client_request WHERE bounty_id = $1 AND status IN ('accepted', 'pending') ORDER BY created_at ASC`
	err := s.DB.SelectContext(ctx, &requests, query, bountyID)
	return requests, err
}
