-- View: view_product_enriched
-- Encapsulates the complex product retrieval logic including storage and categories
CREATE OR REPLACE VIEW view_product_enriched AS
SELECT p.*,
       (SELECT COALESCE(jsonb_agg(jsonb_build_object(
           'stored_in_id', ps.storage_id,
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
       )), '[]') FROM product_category pc JOIN custom_category cc ON pc.category_id = cc.id WHERE pc.product_id = p.id) as categories_json,
       (SELECT COALESCE(jsonb_agg(jsonb_build_object(
           'id', dc.id,
           'name', dc.name,
           'set_code', dc.set_code,
           'collector_number', dc.collector_number,
           'quantity', dc.quantity,
           'type_line', dc.type_line,
           'image_url', dc.image_url,
           'foil_treatment', dc.foil_treatment,
           'card_treatment', dc.card_treatment,
           'rarity', dc.rarity,
           'art_variation', dc.art_variation
       )), '[]') FROM deck_card dc WHERE dc.product_id = p.id) as deck_cards_json
FROM product p;
