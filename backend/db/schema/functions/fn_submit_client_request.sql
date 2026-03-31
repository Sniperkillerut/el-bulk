-- Submit Client Request
-- Atomically handles customer lookup/linking and request creation.
CREATE OR REPLACE FUNCTION fn_submit_client_request(
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_card_name TEXT,
    p_set_name TEXT DEFAULT NULL,
    p_details TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID;
    v_request_id UUID;
    v_created_at TIMESTAMPTZ;
BEGIN
    -- Try to find an existing customer by email or phone
    SELECT id INTO v_customer_id 
    FROM customer 
    WHERE email = p_customer_contact OR phone = p_customer_contact 
    LIMIT 1;

    -- Insert the request (customer_id will be NULL if no match found, which is fine for anonymous requests)
    INSERT INTO client_request (customer_id, customer_name, customer_contact, card_name, set_name, details, status)
    VALUES (v_customer_id, p_customer_name, p_customer_contact, p_card_name, p_set_name, p_details, 'pending')
    RETURNING id, created_at INTO v_request_id, v_created_at;
    
    RETURN jsonb_build_object(
        'id', v_request_id,
        'customer_id', v_customer_id,
        'customer_name', p_customer_name,
        'customer_contact', p_customer_contact,
        'card_name', p_card_name,
        'set_name', p_set_name,
        'details', p_details,
        'status', 'pending',
        'created_at', v_created_at
    );
END;
$$ LANGUAGE plpgsql;
