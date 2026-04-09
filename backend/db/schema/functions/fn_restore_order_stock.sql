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
    v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
    v_check RECORD;
BEGIN
    SELECT status INTO v_status FROM "order" WHERE id = p_order_id;
    IF v_status != 'cancelled' THEN
        RAISE EXCEPTION 'Only cancelled orders can be restored to stock (current status: %)', v_status;
    END IF;

    -- 1. Validate totals (sum per product must not exceed order_item quantity)
    FOR v_check IN 
        SELECT (elem->>'product_id')::uuid as pid, SUM((elem->>'quantity')::int) as total_qty
        FROM jsonb_array_elements(increments) elem
        GROUP BY (elem->>'product_id')::uuid
    LOOP
        IF (SELECT COALESCE(SUM(quantity), 0) FROM order_item WHERE order_id = p_order_id AND product_id = v_check.pid) < v_check.total_qty THEN
            RAISE EXCEPTION 'Restoring more than original quantity for product %', v_check.pid;
        END IF;
    END LOOP;

    -- 2. Iterate through increments and add back to physical locations, and decrement from pending
    FOR inc IN SELECT * FROM jsonb_array_elements(increments)
    LOOP
        -- Physical storage update (increment)
        INSERT INTO product_storage (product_id, storage_id, quantity)
        VALUES (
            (inc->>'product_id')::uuid, 
            (inc->>'stored_in_id')::uuid, 
            (inc->>'quantity')::int
        )
        ON CONFLICT (product_id, storage_id) 
        DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity;

        -- Pending storage update (decrement)
        UPDATE product_storage 
        SET quantity = GREATEST(0, quantity - (inc->>'quantity')::int)
        WHERE product_id = (inc->>'product_id')::uuid 
          AND storage_id = v_pending_id;
    END LOOP;

    -- 3. Mark the order as inventory restored
    UPDATE "order" SET inventory_restored = TRUE WHERE id = p_order_id;
END;
$$ LANGUAGE plpgsql;
