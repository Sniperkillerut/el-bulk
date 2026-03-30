-- Trigger to sync product stock from product_storage
DROP TRIGGER IF EXISTS trg_sync_product_stock ON product_storage;
CREATE TRIGGER trg_sync_product_stock
AFTER INSERT OR UPDATE OR DELETE ON product_storage
FOR EACH ROW EXECUTE FUNCTION fn_update_product_stock();
