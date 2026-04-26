-- fn_fulfill_bounty_offer
-- Atomically fulfills a bounty offer against selected client requests.
--   p_offer_id: the BountyOffer being accepted.
--   p_request_ids: UUID[] of ClientRequests to mark as solved.
CREATE OR REPLACE FUNCTION fn_fulfill_bounty_offer(
    p_offer_id    UUID,
    p_request_ids UUID[]
) RETURNS JSONB AS $$
DECLARE
    v_offer      bounty_offer%ROWTYPE;
    v_fulfilled  INT;
    v_new_qty    INT;
BEGIN
    SELECT * INTO v_offer FROM bounty_offer WHERE id = p_offer_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'offer not found: %', p_offer_id;
    END IF;

    -- Mark offer as accepted
    UPDATE bounty_offer
    SET status = 'accepted', updated_at = now()
    WHERE id = p_offer_id;

    -- Mark selected requests as solved (only those linked to this bounty)
    UPDATE client_request
    SET status = 'solved'
    WHERE id = ANY(p_request_ids)
      AND bounty_id = v_offer.bounty_id;

    GET DIAGNOSTICS v_fulfilled = ROW_COUNT;

    -- Decrement bounty quantity, deactivate if it hits 0
    UPDATE bounty
    SET quantity_needed = GREATEST(0, quantity_needed - v_fulfilled),
        is_active = (quantity_needed - v_fulfilled) > 0,
        updated_at = now()
    WHERE id = v_offer.bounty_id
    RETURNING quantity_needed INTO v_new_qty;

    RETURN jsonb_build_object(
        'offer_id',        p_offer_id,
        'bounty_id',       v_offer.bounty_id,
        'fulfilled',       v_fulfilled,
        'bounty_qty_left', v_new_qty
    );
END;
$$ LANGUAGE plpgsql;
