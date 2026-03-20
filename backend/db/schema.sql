-- El Bulk TCG Store Schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE foil_treatment_type AS ENUM (
  'non_foil', 'foil', 'holo_foil', 'platinum_foil', 'ripple_foil', 'etched_foil', 'galaxy_foil'
);

CREATE TYPE card_treatment_type AS ENUM (
  'normal', 'full_art', 'extended_art', 'borderless', 'showcase',
  'legacy_border', 'textless', 'judge_promo', 'promo', 'alternate_art'
);

-- Admin-configurable global settings (key/value)
CREATE TABLE settings (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings (key, value) VALUES
  ('usd_to_cop_rate', '4200'),   -- TCGPlayer prices
  ('eur_to_cop_rate', '4600');   -- Cardmarket prices

CREATE TABLE products (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name             TEXT NOT NULL,
  tcg              TEXT NOT NULL,
  category         TEXT NOT NULL CHECK (category IN ('singles', 'sealed', 'accessories')),
  set_name         TEXT,
  set_code         TEXT,
  condition        TEXT CHECK (condition IN ('NM', 'LP', 'MP', 'HP', 'DMG')),
  foil_treatment   foil_treatment_type NOT NULL DEFAULT 'non_foil',
  card_treatment   card_treatment_type NOT NULL DEFAULT 'normal',

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
  featured         BOOLEAN NOT NULL DEFAULT FALSE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_tcg      ON products(tcg);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_featured ON products(featured);
CREATE INDEX idx_products_search   ON products USING gin(to_tsvector('english', name || ' ' || COALESCE(set_name, '')));

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
