-- Performance Optimization Indexes
-- Created: 2026-05-14

-- 1. Composite Catalog Index
-- Speeds up the primary browse query (TCG + Category + In Stock)
CREATE INDEX IF NOT EXISTS idx_product_catalog_lookup 
ON product (tcg, category, stock) 
WHERE stock > 0;

-- 2. JSONB GIN Indexes
-- Speeds up sidebar facet filtering (Format, Frame Effects, Card Types)
CREATE INDEX IF NOT EXISTS idx_product_legalities_gin ON product USING GIN (legalities);
CREATE INDEX IF NOT EXISTS idx_product_frame_effects_gin ON product USING GIN (frame_effects);
CREATE INDEX IF NOT EXISTS idx_product_card_types_gin ON product USING GIN (card_types);

-- 3. Storage Lookup Index
-- Speeds up product-to-storage enrichment
CREATE INDEX IF NOT EXISTS idx_product_storage_lookup 
ON product_storage (product_id, storage_id, quantity) 
WHERE quantity > 0;

-- 4. Category Mapping Index
-- Speeds up collection-based filtering
CREATE INDEX IF NOT EXISTS idx_product_category_lookup
ON product_category (product_id, category_id);
