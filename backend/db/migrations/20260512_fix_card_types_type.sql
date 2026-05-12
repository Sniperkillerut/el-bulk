-- Fix: Convert card_types from TEXT[] to JSONB to match models.StringArray scanner
-- Must handle default value casting and dependent views
DROP VIEW IF EXISTS view_product_enriched CASCADE;

ALTER TABLE product ALTER COLUMN card_types DROP DEFAULT;
ALTER TABLE product ALTER COLUMN card_types TYPE JSONB USING to_jsonb(card_types);
ALTER TABLE product ALTER COLUMN card_types SET DEFAULT '[]'::jsonb;

ALTER TABLE deck_card ALTER COLUMN card_types DROP DEFAULT;
ALTER TABLE deck_card ALTER COLUMN card_types TYPE JSONB USING to_jsonb(card_types);
ALTER TABLE deck_card ALTER COLUMN card_types SET DEFAULT '[]'::jsonb;

ALTER TABLE client_request ALTER COLUMN card_types DROP DEFAULT;
ALTER TABLE client_request ALTER COLUMN card_types TYPE JSONB USING to_jsonb(card_types);
ALTER TABLE client_request ALTER COLUMN card_types SET DEFAULT '[]'::jsonb;

-- Recreate view_product_enriched
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
           'searchable', cc.searchable,
           'bg_color', cc.bg_color,
           'text_color', cc.text_color,
           'icon', cc.icon
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
           'art_variation', dc.art_variation,
           'scryfall_id', dc.scryfall_id,
           'frame_effects', dc.frame_effects
       )), '[]') FROM deck_card dc WHERE dc.product_id = p.id) as deck_cards_json
FROM product p;
