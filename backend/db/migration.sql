CREATE TABLE IF NOT EXISTS stored_in (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS product_stored_in (
  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
  stored_in_id UUID REFERENCES stored_in(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  PRIMARY KEY (product_id, stored_in_id)
);

CREATE OR REPLACE FUNCTION update_product_stock()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    UPDATE products SET stock = COALESCE((SELECT sum(quantity) FROM product_stored_in WHERE product_id = OLD.product_id), 0) WHERE id = OLD.product_id;
    RETURN OLD;
  END IF;
  
  UPDATE products SET stock = COALESCE((SELECT sum(quantity) FROM product_stored_in WHERE product_id = NEW.product_id), 0) WHERE id = NEW.product_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_product_stock ON product_stored_in;

CREATE TRIGGER trg_sync_product_stock
AFTER INSERT OR UPDATE OR DELETE ON product_stored_in
FOR EACH ROW EXECUTE FUNCTION update_product_stock();
