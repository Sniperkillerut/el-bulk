-- Submit Client Request
-- Atomically handles customer lookup/linking and request creation.
CREATE OR REPLACE FUNCTION fn_submit_client_request(
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_card_name TEXT,
    p_set_name TEXT DEFAULT NULL,
    p_details TEXT DEFAULT NULL,
    p_quantity INTEGER DEFAULT 1,
    p_tcg TEXT DEFAULT 'mtg',
    p_customer_id UUID DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID := p_customer_id;
    v_request_id UUID;
    v_created_at TIMESTAMPTZ;
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
            v_last_name := '-'; -- Last name is REQUIRED in our schema, so use a hyphen as placeholder if only one name given
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

    -- Insert the request (customer_id will be NULL if no match found, which is fine for anonymous requests)
    INSERT INTO client_request (customer_id, customer_name, customer_contact, card_name, set_name, details, quantity, tcg, status)
    VALUES (v_customer_id, p_customer_name, p_customer_contact, p_card_name, p_set_name, p_details, p_quantity, p_tcg, 'pending')
    RETURNING id, created_at INTO v_request_id, v_created_at;
    
    RETURN jsonb_build_object(
        'id', v_request_id,
        'customer_id', v_customer_id,
        'customer_name', p_customer_name,
        'customer_contact', p_customer_contact,
        'card_name', p_card_name,
        'set_name', p_set_name,
        'details', p_details,
        'quantity', p_quantity,
        'tcg', p_tcg,
        'status', 'pending',
        'created_at', v_created_at
    );
END;
$$ LANGUAGE plpgsql;
