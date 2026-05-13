-- Migration: Add B-Tree indexes for common sorting columns
-- These allow the database to paginate and sort instantly without full-table scans.
CREATE INDEX IF NOT EXISTS idx_product_name_btree ON product(name);
CREATE INDEX IF NOT EXISTS idx_product_created_at ON product(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_product_stock_sort ON product(stock DESC);
