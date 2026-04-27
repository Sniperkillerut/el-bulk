-- fn_cancel_client_request
-- Atomically cancels a client_request:
--   1. If status is 'accepted', decrement bounty.quantity_needed.
--   2. Set status = 'not_needed' and record reason.
CREATE OR REPLACE FUNCTION fn_cancel_client_request(
    p_request_id UUID,
    p_customer_id UUID,
    p_reason TEXT
) RETURNS VOID AS $$
DECLARE
    v_req client_request%ROWTYPE;
BEGIN
    -- Lock and fetch the request
    SELECT * INTO v_req FROM client_request WHERE id = p_request_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'request not found';
    END IF;

    -- Verify ownership
    IF v_req.customer_id IS NOT NULL AND v_req.customer_id != p_customer_id THEN
        RAISE EXCEPTION 'unauthorized: request belongs to another user';
    END IF;

    -- Only pending or accepted can be cancelled by user
    IF v_req.status != 'pending' AND v_req.status != 'accepted' THEN
        -- If already not_needed, do nothing
        IF v_req.status = 'not_needed' THEN
            RETURN;
        END IF;
        RAISE EXCEPTION 'request cannot be cancelled in current status: %', v_req.status;
    END IF;

    -- If accepted, subtract from associated bounty
    IF v_req.status = 'accepted' AND v_req.bounty_id IS NOT NULL THEN
        UPDATE bounty
        SET quantity_needed = GREATEST(0, quantity_needed - v_req.quantity),
            updated_at = now()
        WHERE id = v_req.bounty_id;
    END IF;

    -- Mark as not_needed
    UPDATE client_request
    SET status = 'not_needed',
        cancellation_reason = p_reason,
        updated_at = now()
    WHERE id = p_request_id;

END;
$$ LANGUAGE plpgsql;
