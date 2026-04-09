-- Add generated tsvector column for weighted full-text search
ALTER TABLE product ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
  setweight(to_tsvector('english', name), 'A') ||
  setweight(to_tsvector('english', COALESCE(set_name, '')), 'B') ||
  setweight(to_tsvector('english', COALESCE(oracle_text, '')), 'C') ||
  setweight(to_tsvector('english', 
    COALESCE(set_code, '') || ' ' || 
    COALESCE(collector_number, '') || ' ' || 
    COALESCE(artist, '') || ' ' || 
    COALESCE(type_line, '') || ' ' || 
    COALESCE(promo_type, '')
  ), 'D')
) STORED;

-- Replace old expression index with a direct column index for significantly faster lookups
DROP INDEX IF EXISTS idx_product_search;
CREATE INDEX IF NOT EXISTS idx_product_search_vector ON product USING gin(search_vector);

-- Re-build the trigram index specifically for the name column to support fuzzy matching
DROP INDEX IF EXISTS idx_product_trgm;
CREATE INDEX IF NOT EXISTS idx_product_trgm ON product USING gin (name gin_trgm_ops);
