-- fn_accept_client_request
-- Atomically accepts a client_request:
--   1. Find an existing active bounty matching card identity (via oracle_id or scryfall_id).
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
        -- Find existing generic bounty (oracle_id match preferred, fallback to name)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE is_generic = true
          AND is_active = true
          AND (
            (oracle_id IS NOT NULL AND v_req.oracle_id IS NOT NULL AND oracle_id = v_req.oracle_id)
            OR 
            (lower(trim(name)) = lower(trim(v_req.card_name)) AND tcg = v_req.tcg)
          )
        ORDER BY (oracle_id IS NOT NULL AND v_req.oracle_id IS NOT NULL AND oracle_id = v_req.oracle_id) DESC, created_at DESC
        LIMIT 1;
    ELSE
        -- Find existing specific bounty (scryfall_id match preferred)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE is_generic = false
          AND is_active = true
          AND (
            (scryfall_id IS NOT NULL AND v_req.scryfall_id IS NOT NULL AND v_req.scryfall_id != '' AND scryfall_id = v_req.scryfall_id::UUID)
            OR
            (lower(trim(name)) = lower(trim(v_req.card_name)) AND tcg = v_req.tcg AND (set_name IS NOT DISTINCT FROM v_req.set_name))
          )
        ORDER BY (scryfall_id IS NOT NULL AND v_req.scryfall_id IS NOT NULL AND v_req.scryfall_id != '' AND scryfall_id = v_req.scryfall_id::UUID) DESC, created_at DESC
        LIMIT 1;
    END IF;

    -- Create bounty if not found
    IF v_bounty_id IS NULL THEN
        INSERT INTO bounty (
            name, tcg, set_name, quantity_needed, is_active, is_generic,
            scryfall_id, oracle_id, image_url, set_code, collector_number,
            foil_treatment, card_treatment, frame_effects, language, hide_price, price_source
        ) VALUES (
            trim(v_req.card_name), v_req.tcg, v_req.set_name, v_req.quantity,
            true, v_is_generic,
            NULLIF(v_req.scryfall_id, '')::UUID, v_req.oracle_id, v_req.image_url, v_req.set_code, v_req.collector_number,
            COALESCE(v_req.foil_treatment, 'non_foil'), COALESCE(v_req.card_treatment, 'normal'), v_req.frame_effects,
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
