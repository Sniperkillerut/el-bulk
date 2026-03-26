-- Sync product stock from product_storage sum
CREATE OR REPLACE FUNCTION fn_update_product_stock()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    UPDATE product SET stock = COALESCE((SELECT sum(quantity) FROM product_storage WHERE product_id = OLD.product_id), 0) WHERE id = OLD.product_id;
    RETURN OLD;
  END IF;
  
  UPDATE product SET stock = COALESCE((SELECT sum(quantity) FROM product_storage WHERE product_id = NEW.product_id), 0) WHERE id = NEW.product_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
