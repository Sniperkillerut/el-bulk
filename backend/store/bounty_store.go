package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type BountyStore struct {
	*BaseStore[models.Bounty]
}

func NewBountyStore(db *sqlx.DB) *BountyStore {
	return &BountyStore{
		BaseStore: NewBaseStore[models.Bounty](db, "bounty"),
	}
}

func (s *BountyStore) ListBounties(activeParam string) ([]models.Bounty, error) {
	var bounties []models.Bounty
	query := `
		SELECT 
			id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, 
			hide_price, quantity_needed, image_url, price_source, price_reference, is_active, created_at, updated_at
		FROM bounty
		WHERE ($1 = '' OR is_active = ($1 = 'true'))
		ORDER BY created_at DESC
	`
	err := s.DB.Select(&bounties, query, activeParam)
	return bounties, err
}

func (s *BountyStore) CreateBounty(input models.BountyInput, isActive bool) (*models.Bounty, error) {
	query := `
		INSERT INTO bounty (name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, image_url, price_source, price_reference, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, image_url, price_source, price_reference, is_active, created_at, updated_at
	`
	var bounty models.Bounty
	err := s.DB.QueryRowx(query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.CardTreatment, input.CollectorNumber, input.PromoType, input.Language,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.ImageURL,
		input.PriceSource, input.PriceReference, isActive,
	).StructScan(&bounty)
	if err != nil {
		return nil, err
	}
	return &bounty, nil
}

func (s *BountyStore) UpdateBounty(id string, input models.BountyInput, isActive bool) (*models.Bounty, error) {
	query := `
		UPDATE bounty
		SET name = $1, tcg = $2, set_name = $3, condition = $4, foil_treatment = $5,
		    card_treatment = $6, collector_number = $7, promo_type = $8, language = $9,
		    target_price = $10, hide_price = $11, quantity_needed = $12, image_url = $13,
		    price_source = $14, price_reference = $15, is_active = $16, updated_at = now()
		WHERE id = $17
		RETURNING id, name, tcg, set_name, condition, foil_treatment, card_treatment, collector_number, promo_type, language, target_price, hide_price, quantity_needed, image_url, price_source, price_reference, is_active, created_at, updated_at
	`
	var bounty models.Bounty
	err := s.DB.QueryRowx(query,
		input.Name, input.TCG, input.SetName, input.Condition, input.FoilTreatment,
		input.CardTreatment, input.CollectorNumber, input.PromoType, input.Language,
		input.TargetPrice, input.HidePrice, input.QuantityNeeded, input.ImageURL,
		input.PriceSource, input.PriceReference, isActive, id,
	).StructScan(&bounty)
	if err != nil {
		return nil, err
	}
	return &bounty, nil
}

func (s *BountyStore) DeleteBounty(id string) (int64, error) {
	res, err := s.DB.Exec(`DELETE FROM bounty WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── Offers ──────────────────────────────────────────────

func (s *BountyStore) ListOffers() ([]models.BountyOffer, error) {
	var offers []models.BountyOffer
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
	err := s.DB.Select(&offers, query)
	return offers, err
}

func (s *BountyStore) SubmitOffer(bountyID, customerName, customerContact string, quantity int, condition, notes, status *string, userID *string) ([]byte, error) {
	var result []byte
	err := s.DB.Get(&result, "SELECT fn_submit_bounty_offer($1, $2, $3, $4, $5, $6, $7, $8)",
		bountyID, customerName, customerContact, quantity, condition, notes, status, userID)
	return result, err
}

func (s *BountyStore) UpdateOfferStatus(id, status string) (*models.BountyOffer, error) {
	_, err := s.DB.Exec(`UPDATE bounty_offer SET status = $1, updated_at = now() WHERE id = $2`, status, id)
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
	err = s.DB.Get(&offer, selectQuery, id)
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (s *BountyStore) ListMeOffers(userID string) ([]models.BountyOffer, error) {
	var offers []models.BountyOffer
	query := `
		SELECT 
			o.id, o.bounty_id, o.customer_id, o.quantity, o.condition, o.status, o.notes, o.admin_notes, o.created_at, o.updated_at,
			b.name as bounty_name
		FROM bounty_offer o
		JOIN bounty b ON o.bounty_id = b.id
		WHERE o.customer_id = $1
		ORDER BY o.created_at DESC
	`
	err := s.DB.Select(&offers, query, userID)
	return offers, err
}

func (s *BountyStore) CancelMeOffer(id, userID string) (int64, error) {
	res, err := s.DB.Exec(`
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

func (s *BountyStore) ListRequests() ([]models.ClientRequest, error) {
	var requests []models.ClientRequest
	query := `
		SELECT id, customer_id, customer_name, customer_contact, card_name, set_name, details, status, created_at
		FROM client_request
		ORDER BY created_at DESC
	`
	err := s.DB.Select(&requests, query)
	return requests, err
}

func (s *BountyStore) SubmitRequest(customerName, customerContact, cardName string, setName, details *string, userID *string) ([]byte, error) {
	var result []byte
	err := s.DB.Get(&result, "SELECT fn_submit_client_request($1, $2, $3, $4, $5, $6)",
		customerName, customerContact, cardName, setName, details, userID)
	return result, err
}

func (s *BountyStore) UpdateRequestStatus(id, status string) (*models.ClientRequest, error) {
	query := `
		UPDATE client_request
		SET status = $1
		WHERE id = $2
		RETURNING id, customer_name, customer_contact, card_name, set_name, details, status, created_at
	`
	var req models.ClientRequest
	err := s.DB.QueryRowx(query, status, id).StructScan(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *BountyStore) ListMeRequests(userID string) ([]models.ClientRequest, error) {
	var requests []models.ClientRequest
	query := `
		SELECT id, customer_id, customer_name, customer_contact, card_name, set_name, details, status, created_at
		FROM client_request
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	err := s.DB.Select(&requests, query, userID)
	return requests, err
}

func (s *BountyStore) CancelMeRequest(id, userID string) (int64, error) {
	res, err := s.DB.Exec(`
		UPDATE client_request 
		SET status = 'cancelled' 
		WHERE id = $1 AND customer_id = $2 AND status = 'pending'
	`, id, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
