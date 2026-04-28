-- Submit Client Requests Batch
-- Atomically handles customer lookup/linking and multiple request creations.
CREATE OR REPLACE FUNCTION fn_submit_client_requests_batch(
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_cards JSONB, -- Array of objects: {card_name, set_name, details, quantity, tcg, scryfall_id, oracle_id, ...}
    p_customer_id UUID DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID := p_customer_id;
    v_card JSONB;
    v_count INT := 0;
    v_first_name TEXT;
    v_last_name TEXT;
BEGIN
    -- Only lookup/create if no explicit ID provided
    IF v_customer_id IS NULL THEN
        -- Try to find an existing customer by email or phone
        SELECT id INTO v_customer_id 
        FROM customer 
        WHERE email = p_customer_contact OR phone = p_customer_contact 
        LIMIT 1;

        -- If no customer found, create one to ensure data integrity
        IF v_customer_id IS NULL THEN
            v_first_name := split_part(p_customer_name, ' ', 1);
            v_last_name := trim(substring(p_customer_name from char_length(v_first_name) + 1));
            
            IF v_last_name = '' THEN
                v_last_name := '-';
            END IF;

            INSERT INTO customer (
                first_name, 
                last_name, 
                email, 
                phone
            ) VALUES (
                v_first_name, 
                v_last_name, 
                CASE WHEN p_customer_contact LIKE '%@%' THEN p_customer_contact ELSE NULL END,
                CASE WHEN p_customer_contact NOT LIKE '%@%' THEN p_customer_contact ELSE NULL END
            ) RETURNING id INTO v_customer_id;
        END IF;
    END IF;

    -- Iterate over cards and insert
    FOR v_card IN SELECT * FROM jsonb_array_elements(p_cards) LOOP
        INSERT INTO client_request (
            customer_id, 
            customer_name, 
            customer_contact, 
            card_name, 
            set_name, 
            details, 
            quantity,
            tcg,
            status,
            match_type,
            scryfall_id,
            oracle_id,
            image_url,
            foil_treatment,
            card_treatment,
            set_code,
            collector_number
        ) VALUES (
            v_customer_id, 
            p_customer_name, 
            p_customer_contact, 
            trim(v_card->>'card_name'), 
            v_card->>'set_name', 
            v_card->>'details', 
            COALESCE((v_card->>'quantity')::INT, 1),
            COALESCE(lower(trim(v_card->>'tcg')), 'mtg'),
            'pending',
            COALESCE(v_card->>'match_type', 'any'),
            v_card->>'scryfall_id',
            (v_card->>'oracle_id')::UUID,
            v_card->>'image_url',
            v_card->>'foil_treatment',
            v_card->>'card_treatment',
            v_card->>'set_code',
            v_card->>'collector_number'
        );
        v_count := v_count + 1;
    END LOOP;

    RETURN jsonb_build_object(
        'count', v_count,
        'customer_id', v_customer_id
    );
END;
$$ LANGUAGE plpgsql;
