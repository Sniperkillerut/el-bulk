-- Migration: Add missing metadata columns to match Go models
-- Tables affected: product, client_request, deck_card

-- 1. Product table
ALTER TABLE product ADD COLUMN IF NOT EXISTS oracle_id UUID;
ALTER TABLE product ADD COLUMN IF NOT EXISTS is_prepared BOOLEAN NOT NULL DEFAULT false;

-- 2. Client Request table
-- oracle_id already exists in some environments but adding is_prepared
ALTER TABLE client_request ADD COLUMN IF NOT EXISTS is_prepared BOOLEAN NOT NULL DEFAULT false;

-- 3. Deck Card table
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS oracle_id UUID;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS is_prepared BOOLEAN NOT NULL DEFAULT false;

-- 4. Update view_product_enriched to include the new columns
DROP VIEW IF EXISTS view_product_enriched CASCADE;
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
           'oracle_id', dc.oracle_id,
           'is_prepared', dc.is_prepared,
           'frame_effects', dc.frame_effects
       )), '[]') FROM deck_card dc WHERE dc.product_id = p.id) as deck_cards_json
FROM product p;
