-- Complete Order
-- Atomically decrements stock and updates status.
CREATE OR REPLACE FUNCTION fn_complete_order(
    p_order_id UUID,
    decrements jsonb
)
RETURNS VOID AS $$
DECLARE
    dec jsonb;
    v_status TEXT;
BEGIN
    SELECT status INTO v_status FROM "order" WHERE id = p_order_id;
    IF v_status = 'completed' THEN
        RAISE EXCEPTION 'Order is already completed';
    END IF;

    FOR dec IN SELECT * FROM jsonb_array_elements(decrements)
    LOOP
        UPDATE product_storage 
        SET quantity = quantity - (dec->>'quantity')::int 
        WHERE product_id = (dec->>'product_id')::uuid 
          AND storage_id = (dec->>'stored_in_id')::uuid;
          
        -- trigger trg_sync_product_stock handles product.stock total
    END LOOP;

    UPDATE "order" 
    SET status = 'completed', completed_at = now() 
    WHERE id = p_order_id;
END;
$$ LANGUAGE plpgsql;
