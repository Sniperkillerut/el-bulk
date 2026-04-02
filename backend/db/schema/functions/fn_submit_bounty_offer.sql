-- Submit Bounty Offer
-- Atomically handles customer lookup/linking and offer creation.
CREATE OR REPLACE FUNCTION fn_submit_bounty_offer(
    p_bounty_id UUID,
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_quantity INTEGER,
    p_condition TEXT DEFAULT NULL,
    p_notes TEXT DEFAULT NULL,
    p_status TEXT DEFAULT 'pending',
    p_customer_id UUID DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID := p_customer_id;
    v_offer_id UUID;
    v_created_at TIMESTAMPTZ;
    v_bounty_name TEXT;
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
    
    INSERT INTO bounty_offer (bounty_id, customer_id, quantity, condition, notes, status)
    VALUES (p_bounty_id, v_customer_id, p_quantity, p_condition, p_notes, COALESCE(p_status, 'pending'))
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
        'status', COALESCE(p_status, 'pending'),
        'notes', p_notes,
        'created_at', v_created_at,
        'updated_at', v_created_at
    );
END;
$$ LANGUAGE plpgsql;
