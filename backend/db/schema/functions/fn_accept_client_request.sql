-- fn_accept_client_request
-- Atomically accepts a client_request:
--   1. Find an existing active bounty matching card identity (exact or generic).
--   2. If none, create a new bounty.
--   3. Link the request to the bounty, increment bounty.quantity_needed, set request status = 'accepted'.
CREATE OR REPLACE FUNCTION fn_accept_client_request(
    p_request_id UUID
) RETURNS JSONB AS $$
DECLARE
    v_req         client_request%ROWTYPE;
    v_bounty_id   UUID;
    v_is_generic  BOOLEAN;
BEGIN
    -- Lock and fetch the request
    SELECT * INTO v_req FROM client_request WHERE id = p_request_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'request not found: %', p_request_id;
    END IF;

    v_is_generic := (v_req.match_type = 'any');

    IF v_is_generic THEN
        -- Find existing generic bounty (name + tcg match, is_generic = true)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE lower(name) = lower(v_req.card_name)
          AND tcg = v_req.tcg
          AND is_generic = true
          AND is_active = true
        LIMIT 1;
    ELSE
        -- Find existing specific bounty (name + set_name + tcg, is_generic = false)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE lower(name) = lower(v_req.card_name)
          AND tcg = v_req.tcg
          AND is_generic = false
          AND is_active = true
          AND (set_name IS NOT DISTINCT FROM v_req.set_name)
        LIMIT 1;
    END IF;

    -- Create bounty if not found
    IF v_bounty_id IS NULL THEN
        INSERT INTO bounty (
            name, tcg, set_name, quantity_needed, is_active, is_generic,
            scryfall_id, image_url, set_code, collector_number,
            foil_treatment, card_treatment, language, hide_price, price_source
        ) VALUES (
            v_req.card_name, v_req.tcg, v_req.set_name, v_req.quantity,
            true, v_is_generic,
            v_req.scryfall_id, v_req.image_url, v_req.set_code, v_req.collector_number,
            COALESCE(v_req.foil_treatment, 'non_foil'), COALESCE(v_req.card_treatment, 'normal'), 
            'en', false, 'tcgplayer'
        )
        RETURNING id INTO v_bounty_id;
    ELSE
        -- Increment quantity on existing bounty
        UPDATE bounty
        SET quantity_needed = quantity_needed + v_req.quantity,
            updated_at = now()
        WHERE id = v_bounty_id;
    END IF;

    -- Link request to bounty and mark as accepted
    UPDATE client_request
    SET bounty_id = v_bounty_id,
        status = 'accepted'
    WHERE id = p_request_id;

    RETURN jsonb_build_object(
        'bounty_id', v_bounty_id,
        'request_id', p_request_id,
        'is_generic', v_is_generic
    );
END;
$$ LANGUAGE plpgsql;
