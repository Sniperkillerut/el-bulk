-- Place Order
-- Handles customer upsert, order creation, and item population in one transaction.
CREATE OR REPLACE FUNCTION fn_place_order(
    customer_data jsonb,
    order_items_data jsonb,
    order_meta jsonb
)
RETURNS TABLE(order_id UUID, order_number TEXT) AS $$
DECLARE
    v_customer_id UUID;
    v_order_id UUID;
    v_order_num TEXT;
BEGIN
    -- Upsert Customer logic based on whether ID is provided
    IF customer_data->>'id' IS NOT NULL AND customer_data->>'id' != '' THEN
        UPDATE customer SET
            first_name = customer_data->>'first_name',
            last_name = customer_data->>'last_name',
            email = customer_data->>'email',
            phone = customer_data->>'phone',
            id_number = customer_data->>'id_number',
            address = customer_data->>'address'
        WHERE id = (customer_data->>'id')::uuid
        RETURNING id INTO v_customer_id;
    END IF;

    IF v_customer_id IS NULL THEN
        INSERT INTO customer (first_name, last_name, email, phone, id_number, address)
        VALUES (
            customer_data->>'first_name',
            customer_data->>'last_name',
            customer_data->>'email',
            customer_data->>'phone',
            customer_data->>'id_number',
            customer_data->>'address'
        )
        ON CONFLICT (phone) DO UPDATE SET
            first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            email = EXCLUDED.email,
            id_number = EXCLUDED.id_number,
            address = EXCLUDED.address
        RETURNING id INTO v_customer_id;
    END IF;

    -- Create Order
    INSERT INTO "order" (order_number, customer_id, status, payment_method, total_cop, notes)
    VALUES (
        order_meta->>'order_number',
        v_customer_id,
        'pending',
        order_meta->>'payment_method',
        (order_meta->>'total_cop')::numeric,
        order_meta->>'notes'
    )
    RETURNING id, "order".order_number INTO v_order_id, v_order_num;

    -- Insert Order Items
    INSERT INTO order_item (
        order_id, product_id, product_name, product_set, 
        foil_treatment, card_treatment, condition, unit_price_cop, quantity, stored_in_snapshot
    )
    SELECT 
        v_order_id,
        (oi->>'product_id')::uuid,
        oi->>'product_name',
        oi->>'product_set',
        oi->>'foil_treatment',
        oi->>'card_treatment',
        oi->>'condition',
        (oi->>'unit_price_cop')::numeric,
        (oi->>'quantity')::int,
        (oi->'stored_in_snapshot')
    FROM jsonb_array_elements(order_items_data) AS oi;

    order_id := v_order_id;
    order_number := v_order_num;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;
