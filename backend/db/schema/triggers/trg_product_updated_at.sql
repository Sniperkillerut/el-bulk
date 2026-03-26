-- Trigger to update updated_at on product
CREATE TRIGGER trg_product_updated_at
BEFORE UPDATE ON product
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
