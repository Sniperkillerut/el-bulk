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
    v_restored BOOLEAN;
BEGIN
    SELECT status, inventory_restored INTO v_status, v_restored FROM "order" WHERE id = p_order_id;
    
    IF v_restored THEN
        RAISE EXCEPTION 'Inventory already restored for this order';
    END IF;

    IF v_status != 'cancelled' THEN
        RAISE EXCEPTION 'Only cancelled orders can be restored to stock (current status: %)', v_status;
    END IF;

    -- 1. Validate totals (every product in order must be fully restored)
    FOR v_check IN 
        SELECT 
            oi.product_id, 
            oi.product_name,
            oi.quantity as order_qty,
            COALESCE(inc.inc_qty, 0) as restored_qty
        FROM order_item oi
        LEFT JOIN (
            SELECT (elem->>'product_id')::uuid as pid, SUM((elem->>'quantity')::int) as inc_qty
            FROM jsonb_array_elements(increments) AS elem
            GROUP BY (elem->>'product_id')::uuid
        ) inc ON oi.product_id = inc.pid
        WHERE oi.order_id = p_order_id 
          AND oi.product_id IS NOT NULL 
          AND oi.quantity > 0
    LOOP
        IF v_check.restored_qty != v_check.order_qty THEN
            RAISE EXCEPTION 'Debes restaurar la cantidad total (%) para el producto % (asignado: %)', 
                v_check.order_qty, v_check.product_name, v_check.restored_qty;
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
