-- Trigger to release pending stock when an order is cancelled
CREATE OR REPLACE FUNCTION fn_release_pending_stock()
RETURNS TRIGGER AS $$
DECLARE
   v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
   v_item RECORD;
BEGIN
   -- Only process if status changed TO cancelled, and it wasn't already completed/cancelled
   IF NEW.status = 'cancelled' AND OLD.status NOT IN ('cancelled', 'completed') THEN
      FOR v_item IN SELECT product_id, quantity FROM order_item WHERE order_id = NEW.id AND product_id IS NOT NULL AND quantity > 0 LOOP
         UPDATE product_storage 
         SET quantity = GREATEST(0, quantity - v_item.quantity)
         WHERE product_id = v_item.product_id AND storage_id = v_pending_id;
      END LOOP;
   END IF;
   
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_order_cancel ON "order";
CREATE TRIGGER trg_order_cancel
AFTER UPDATE OF status ON "order"
FOR EACH ROW
WHEN (NEW.status = 'cancelled' AND OLD.status != 'cancelled')
EXECUTE FUNCTION fn_release_pending_stock();
