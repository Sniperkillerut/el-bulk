-- View: view_product_enriched
-- Encapsulates the complex product retrieval logic including storage and categories
CREATE OR REPLACE VIEW view_product_enriched AS
SELECT p.*,
       (SELECT COALESCE(jsonb_agg(jsonb_build_object(
           'storage_id', ps.storage_id,
           'name', sl.name,
           'quantity', ps.quantity
       )), '[]') FROM product_storage ps JOIN storage_location sl ON ps.storage_id = sl.id WHERE ps.product_id = p.id AND ps.quantity > 0) as stored_in_json,
       (SELECT COALESCE(jsonb_agg(jsonb_build_object(
           'id', cc.id,
           'name', cc.name,
           'slug', cc.slug,
           'show_badge', cc.show_badge,
           'is_active', cc.is_active,
           'searchable', cc.searchable
       )), '[]') FROM product_category pc JOIN custom_category cc ON pc.category_id = cc.id WHERE pc.product_id = p.id) as categories_json
FROM product p;
