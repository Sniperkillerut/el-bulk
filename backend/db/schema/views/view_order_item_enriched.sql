-- View: view_order_item_enriched
-- Enriches order items with product imagery, current stock, and storage locations
CREATE OR REPLACE VIEW view_order_item_enriched AS
SELECT oi.*, 
       p.image_url, 
       COALESCE(p.stock, 0) as stock,
       COALESCE((
           SELECT jsonb_agg(jsonb_build_object(
               'stored_in_id', ps.storage_id,
               'name', sl.name,
               'quantity', ps.quantity
           ))
           FROM product_storage ps
           JOIN storage_location sl ON ps.storage_id = sl.id
           WHERE ps.product_id = oi.product_id AND ps.quantity > 0
       ), '[]'::jsonb) as stored_in
FROM order_item oi
LEFT JOIN product p ON oi.product_id = p.id;
