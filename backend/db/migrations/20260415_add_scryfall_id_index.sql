-- Add index on scryfall_id to ensure fast lookup for unified pricing logic.
-- Use COALESCE in the search to match both existing IDs and potentially 'null' strings
-- if the DB uses a string placeholder.
CREATE INDEX IF NOT EXISTS idx_product_scryfall_id ON product (scryfall_id);
