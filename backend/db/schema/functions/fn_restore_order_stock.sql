-- Restore Order Stock
-- Adds quantities back to specified physical storage locations for a cancelled order.
CREATE OR REPLACE FUNCTION fn_restore_order_stock(
    p_order_id UUID,
    increments jsonb
)
RETURNS VOID AS $$
DECLARE
    inc jsonb;
    v_status TEXT;
BEGIN
    SELECT status INTO v_status FROM "order" WHERE id = p_order_id;
    IF v_status != 'cancelled' THEN
        RAISE EXCEPTION 'Only cancelled orders can be restored to stock (current status: %)', v_status;
    END IF;

    -- Iterate through increments and add back to physical locations
    FOR inc IN SELECT * FROM jsonb_array_elements(increments)
    LOOP
        -- upsert in case the location record was deleted (though unlikely)
        INSERT INTO product_storage (product_id, storage_id, quantity)
        VALUES (
            (inc->>'product_id')::uuid, 
            (inc->>'stored_in_id')::uuid, 
            (inc->>'quantity')::int
        )
        ON CONFLICT (product_id, storage_id) 
        DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
