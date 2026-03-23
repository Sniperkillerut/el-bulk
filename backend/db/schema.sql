-- El Bulk TCG Store Schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Admin-configurable global settings (key/value)
CREATE TABLE settings (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings (key, value) VALUES
  ('usd_to_cop_rate', '4200'),   -- TCGPlayer prices
  ('eur_to_cop_rate', '4600'),   -- Cardmarket prices
  ('contact_address', 'Calle Falsa 123, Bogotá, Colombia'),
  ('contact_whatsapp', '+57 321 456 7890'),
  ('contact_email', 'contacto@elbulk.com'),
  ('contact_instagram', 'elbulk_tcg'),
  ('contact_hours', 'Mon-Sat: 10am - 8pm');

CREATE TABLE products (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name             TEXT NOT NULL,
  tcg              TEXT NOT NULL,
  category         TEXT NOT NULL CHECK (category IN ('singles', 'sealed', 'accessories')),
  set_name         TEXT,
  set_code         TEXT,
  condition        TEXT CHECK (condition IN ('NM', 'LP', 'MP', 'HP', 'DMG')),
  foil_treatment   TEXT NOT NULL DEFAULT 'non_foil',
  card_treatment   TEXT NOT NULL DEFAULT 'normal',

  -- Pricing ------------------------------------------------------------------
  -- price_reference: raw price fetched from external source (USD or EUR)
  -- price_source:    where the reference price came from
  --                  'tcgplayer'  → USD, use usd_to_cop_rate
  --                  'cardmarket' → EUR, use eur_to_cop_rate
  --                  'manual'     → price_cop_override is the final COP price
  -- price_cop_override: explicit COP price set by admin; overrides computed value
  -- Final COP shown to customers = COALESCE(price_cop_override,
  --                                         price_reference * rate)
  price_reference   NUMERIC(12, 4) CHECK (price_reference >= 0),
  price_source      TEXT NOT NULL DEFAULT 'manual'
                    CHECK (price_source IN ('tcgplayer', 'cardmarket', 'manual')),
  price_cop_override NUMERIC(12, 2) CHECK (price_cop_override >= 0),
  -- -------------------------------------------------------------------------

  stock            INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
  image_url        TEXT,
  description      TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_tcg      ON products(tcg);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_search   ON products USING gin(to_tsvector('english', name || ' ' || COALESCE(set_name, '')));

-- Custom Categories (Collections) --------------------------------------------
CREATE TABLE custom_categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Default "Featured" category
INSERT INTO custom_categories (name, slug) VALUES ('Featured', 'featured');

CREATE TABLE product_categories (
  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
  category_id UUID REFERENCES custom_categories(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);

CREATE INDEX idx_pc_category_id ON product_categories(category_id);
-- ----------------------------------------------------------------------------

CREATE TABLE admins (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username      TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Auto-update updated_at
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

-- Storage Locations ----------------------------------------------------------

CREATE TABLE stored_in (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE product_stored_in (
  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
  stored_in_id UUID REFERENCES stored_in(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  PRIMARY KEY (product_id, stored_in_id)
);

CREATE INDEX idx_ps_stored_in_id ON product_stored_in(stored_in_id) WHERE quantity > 0;

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

