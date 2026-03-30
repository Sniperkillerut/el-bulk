-- Function to decrement bounty quantity when an offer is fulfilled
CREATE OR REPLACE FUNCTION fn_fulfill_bounty_offer()
RETURNS TRIGGER AS $$
BEGIN
    -- If status changed to 'fulfilled'
    IF NEW.status = 'fulfilled' AND OLD.status != 'fulfilled' THEN
        UPDATE bounty
        SET quantity_needed = GREATEST(0, quantity_needed - NEW.quantity),
            is_active = CASE WHEN quantity_needed - NEW.quantity <= 0 THEN false ELSE is_active END,
            updated_at = now()
        WHERE id = NEW.bounty_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
