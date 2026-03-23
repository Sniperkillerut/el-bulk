-- Migration to Add Custom Categories and Migrate 'featured' column

-- 1. Create tables
CREATE TABLE IF NOT EXISTS custom_categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS product_categories (
  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
  category_id UUID REFERENCES custom_categories(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_pc_category_id ON product_categories(category_id);

-- 2. Create the default "Featured" category
INSERT INTO custom_categories (name, slug) 
VALUES ('Featured', 'featured') 
ON CONFLICT (slug) DO NOTHING;

-- 3. Migrate existing 'featured = true' products
INSERT INTO product_categories (product_id, category_id)
SELECT id, (SELECT id FROM custom_categories WHERE slug = 'featured')
FROM products
WHERE featured = TRUE
ON CONFLICT DO NOTHING;

-- 4. Drop the old 'featured' column and its index
DROP INDEX IF EXISTS idx_products_featured;
ALTER TABLE products DROP COLUMN IF EXISTS featured;
