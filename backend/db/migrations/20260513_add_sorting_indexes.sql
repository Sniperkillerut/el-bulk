-- Optimization for sorting in listing pages
CREATE INDEX IF NOT EXISTS idx_product_created_at_id ON product(created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_product_price_cop_id ON product(price_cop DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_product_stock_id ON product(stock DESC, id DESC);
