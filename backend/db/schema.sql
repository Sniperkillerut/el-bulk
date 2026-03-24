-- El Bulk TCG Store Schema
-- Consolidated on 2026-03-24

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Admin-configurable global settings (key/value)
CREATE TABLE settings (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings (key, value) VALUES
  ('usd_to_cop_rate', '4200'),
  ('eur_to_cop_rate', '4600'),
  ('contact_address', 'Cra. 15 # 76-54, Local 201, Centro Comercial Unilago, Bogotá'),
  ('contact_phone', '+57 300 000 0000'),
  ('contact_email', 'contact@el-bulk.co'),
  ('contact_instagram', 'el-bulk'),
  ('contact_hours', 'Mon - Sat: 11:00 AM - 7:00 PM')
ON CONFLICT (key) DO NOTHING;

CREATE TABLE products (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name              TEXT NOT NULL,
  tcg               TEXT NOT NULL,
  category          TEXT NOT NULL CHECK (category IN ('singles', 'sealed', 'accessories')),
  set_name          TEXT,
  set_code          TEXT,
  collector_number  TEXT,
  condition         TEXT CHECK (condition IN ('NM', 'LP', 'MP', 'HP', 'DMG')),
  foil_treatment    TEXT NOT NULL DEFAULT 'non_foil',
  card_treatment    TEXT NOT NULL DEFAULT 'normal',
  promo_type        TEXT,

  -- Pricing
  price_reference    NUMERIC(12, 4) CHECK (price_reference >= 0),
  price_source       TEXT NOT NULL DEFAULT 'manual'
                     CHECK (price_source IN ('tcgplayer', 'cardmarket', 'manual')),
  price_cop_override NUMERIC(12, 2) CHECK (price_cop_override >= 0),

  stock             INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
  image_url         TEXT,
  description       TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_tcg      ON products(tcg);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_search   ON products USING gin(to_tsvector('english', name || ' ' || COALESCE(set_name, '')));
CREATE INDEX idx_products_trgm     ON products USING gin (name gin_trgm_ops);

-- Custom Categories (Collections)
CREATE TABLE custom_categories (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT UNIQUE NOT NULL,
  slug       TEXT UNIQUE NOT NULL,
  is_active  BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Default categories
INSERT INTO custom_categories (name, slug) VALUES 
  ('Featured', 'featured'),
  ('Sale', 'sale'),
  ('New Arrivals', 'new-arrivals'),
  ('Hot Items', 'hot-items')
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE product_categories (
  product_id  UUID REFERENCES products(id) ON DELETE CASCADE,
  category_id UUID REFERENCES custom_categories(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);

CREATE INDEX idx_pc_category_id ON product_categories(category_id);

-- Admins
CREATE TABLE admins (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username      TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Storage Locations
CREATE TABLE stored_in (
  id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE product_stored_in (
  product_id   UUID REFERENCES products(id) ON DELETE CASCADE,
  stored_in_id UUID REFERENCES stored_in(id) ON DELETE CASCADE,
  quantity     INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  PRIMARY KEY (product_id, stored_in_id)
);

CREATE INDEX idx_ps_stored_in_id ON product_stored_in(stored_in_id) WHERE quantity > 0;

-- Checkout / Orders
CREATE TABLE customers (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  first_name      TEXT NOT NULL,
  last_name       TEXT NOT NULL,
  email           TEXT,
  phone           TEXT NOT NULL,
  id_number       TEXT,
  address         TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_number   TEXT UNIQUE NOT NULL,
  customer_id    UUID REFERENCES customers(id) ON DELETE RESTRICT,
  status         TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'confirmed', 'cancelled')),
  payment_method TEXT NOT NULL,
  total_cop      NUMERIC(14, 2) NOT NULL,
  notes          TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id       UUID REFERENCES orders(id) ON DELETE CASCADE,
  product_id     UUID REFERENCES products(id) ON DELETE SET NULL,
  product_name   TEXT NOT NULL,
  product_set    TEXT,
  foil_treatment TEXT,
  card_treatment TEXT,
  condition      TEXT,
  unit_price_cop NUMERIC(14, 2) NOT NULL,
  quantity       INTEGER NOT NULL,
  stored_in_snapshot JSONB
);

-- Triggers & Functions
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_settings_updated_at
BEFORE UPDATE ON settings
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Sync product stock from product_stored_in sum
CREATE OR REPLACE FUNCTION update_product_stock()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    UPDATE products SET stock = COALESCE((SELECT sum(quantity) FROM product_stored_in WHERE product_id = OLD.product_id), 0) WHERE id = OLD.product_id;
    RETURN OLD;
  END IF;
  
  UPDATE products SET stock = COALESCE((SELECT sum(quantity) FROM product_stored_in WHERE product_id = NEW.product_id), 0) WHERE id = NEW.product_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sync_product_stock
AFTER INSERT OR UPDATE OR DELETE ON product_stored_in
FOR EACH ROW EXECUTE FUNCTION update_product_stock();
