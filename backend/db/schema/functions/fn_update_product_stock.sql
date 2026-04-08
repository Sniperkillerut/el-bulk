-- Sync product stock from product_storage sum (excluding 'pending', and subtracting 'pending')
CREATE OR REPLACE FUNCTION fn_update_product_stock()
RETURNS TRIGGER AS $$
DECLARE
  v_pid UUID;
  v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
BEGIN
  IF TG_OP = 'DELETE' THEN
    v_pid := OLD.product_id;
  ELSE
    v_pid := NEW.product_id;
  END IF;

  UPDATE product 
  SET stock = GREATEST(0, COALESCE((SELECT sum(quantity) FROM product_storage WHERE product_id = v_pid AND storage_id != v_pending_id), 0) -
                          COALESCE((SELECT quantity FROM product_storage WHERE product_id = v_pid AND storage_id = v_pending_id), 0))
  WHERE id = v_pid;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
