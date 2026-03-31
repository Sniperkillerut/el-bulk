-- Submit Bounty Offer
-- Atomically handles customer lookup/linking and offer creation.
CREATE OR REPLACE FUNCTION fn_submit_bounty_offer(
    p_bounty_id UUID,
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_quantity INTEGER,
    p_condition TEXT DEFAULT NULL,
    p_notes TEXT DEFAULT NULL,
    p_status TEXT DEFAULT 'pending'
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID;
    v_offer_id UUID;
    v_created_at TIMESTAMPTZ;
    v_bounty_name TEXT;
BEGIN
    -- Try to find an existing customer by email or phone
    SELECT id INTO v_customer_id 
    FROM customer 
    WHERE email = p_customer_contact OR phone = p_customer_contact 
    LIMIT 1;

    -- If no customer found, we could choose to create one or reject. 
    -- For now, we allow NULL customer_id as it might be managed by admin later,
    -- but usually, we want to link it.
    
    INSERT INTO bounty_offer (bounty_id, customer_id, quantity, condition, notes, status)
    VALUES (p_bounty_id, v_customer_id, p_quantity, p_condition, p_notes, p_status)
    RETURNING id, created_at INTO v_offer_id, v_created_at;

    -- Get bounty name for return object
    SELECT name INTO v_bounty_name FROM bounty WHERE id = p_bounty_id;
    
    RETURN jsonb_build_object(
        'id', v_offer_id,
        'bounty_id', p_bounty_id,
        'customer_id', v_customer_id,
        'customer_name', p_customer_name,
        'customer_contact', p_customer_contact,
        'bounty_name', v_bounty_name,
        'quantity', p_quantity,
        'condition', p_condition,
        'status', p_status,
        'notes', p_notes,
        'created_at', v_created_at,
        'updated_at', v_created_at
    );
END;
$$ LANGUAGE plpgsql;
