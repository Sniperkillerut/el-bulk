-- Phase 3: Add strategic indexes for sorting and filtering

-- 1. Index for Price sorting (Materialized COP price)
CREATE INDEX IF NOT EXISTS idx_product_price_cop ON product(price_cop ASC);

-- 2. Index for Newest sorting
CREATE INDEX IF NOT EXISTS idx_product_created_at ON product(created_at DESC);

-- 3. Index for Alphabetical sorting (using C collation for consistent speed)
CREATE INDEX IF NOT EXISTS idx_product_name_sort ON product(name COLLATE "C" ASC);

-- 4. Index for In-stock filtering
CREATE INDEX IF NOT EXISTS idx_product_stock_filter ON product(stock) WHERE stock > 0;

-- 5. Index for CMC and Rarity sorting/filtering
CREATE INDEX IF NOT EXISTS idx_product_cmc ON product(cmc);
CREATE INDEX IF NOT EXISTS idx_product_rarity ON product(rarity);
