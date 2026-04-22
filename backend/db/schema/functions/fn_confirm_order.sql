-- Confirm Order
-- Atomically decrements stock from admin locations, removes from pending, and updates status.
CREATE OR REPLACE FUNCTION fn_confirm_order(
    p_order_id UUID,
    decrements jsonb
)
RETURNS VOID AS $$
DECLARE
    dec jsonb;
    v_status TEXT;
    v_item RECORD;
    v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
BEGIN
    SELECT status INTO v_status FROM "order" WHERE id = p_order_id;
    IF v_status = 'confirmed' OR v_status = 'completed' OR v_status = 'cancelled' THEN
        RAISE EXCEPTION 'Order is already processed (status: %)', v_status;
    END IF;

    -- 1. Decrement from 'pending' location (Increases product.stock)
    FOR v_item IN SELECT product_id, quantity FROM order_item WHERE order_id = p_order_id AND product_id IS NOT NULL AND quantity > 0
    LOOP
        UPDATE product_storage 
        SET quantity = GREATEST(0, quantity - v_item.quantity)
        WHERE product_id = v_item.product_id AND storage_id = v_pending_id;
    END LOOP;

    -- 2. Decrement from specified admin locations (Decreases product.stock)
    FOR dec IN SELECT * FROM jsonb_array_elements(decrements)
    LOOP
        UPDATE product_storage 
        SET quantity = GREATEST(0, quantity - (dec->>'quantity')::int)
        WHERE product_id = (dec->>'product_id')::uuid 
          AND storage_id = (dec->>'stored_in_id')::uuid;
    END LOOP;

    UPDATE "order" 
    SET status = 'confirmed',
        confirmed_at = now()
    WHERE id = p_order_id;
END;
$$ LANGUAGE plpgsql;
