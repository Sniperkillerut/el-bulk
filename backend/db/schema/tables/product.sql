-- Product Table
CREATE TABLE IF NOT EXISTS product (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name              TEXT NOT NULL,
  tcg               TEXT NOT NULL,
  category          TEXT NOT NULL CHECK (category IN ('singles', 'sealed', 'accessories', 'store_exclusives')),
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
  cost_basis_cop    NUMERIC(12, 2) NOT NULL DEFAULT 0,

  -- MTG Metadata
  language          TEXT NOT NULL DEFAULT 'en',
  color_identity    TEXT, -- comma-separated (W,U,B,R,G)
  rarity            TEXT,
  cmc               NUMERIC(5, 1),
  is_legendary      BOOLEAN NOT NULL DEFAULT false,
  is_historic       BOOLEAN NOT NULL DEFAULT false,
  is_land           BOOLEAN NOT NULL DEFAULT false,
  is_basic_land     BOOLEAN NOT NULL DEFAULT false,
  art_variation     TEXT,
  oracle_text       TEXT,
  artist            TEXT,
  type_line         TEXT,
  border_color      TEXT,
  frame             TEXT,
  full_art          BOOLEAN NOT NULL DEFAULT false,
  textless          BOOLEAN NOT NULL DEFAULT false,
  scryfall_id       UUID,
  legalities        JSONB, -- { "commander": "legal", "modern": "banned", ... }

  -- Generated Full-Text Search Vector
  search_vector tsvector GENERATED ALWAYS AS (
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
  ) STORED,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indices
CREATE INDEX IF NOT EXISTS idx_product_tcg      ON product(tcg);
CREATE INDEX IF NOT EXISTS idx_product_category ON product(category);
CREATE INDEX IF NOT EXISTS idx_product_search_vector ON product USING gin(search_vector);
CREATE INDEX IF NOT EXISTS idx_product_trgm          ON product USING gin (name gin_trgm_ops);

-- Facet Indices for Performance
CREATE INDEX IF NOT EXISTS idx_product_condition      ON product(condition);
CREATE INDEX IF NOT EXISTS idx_product_foil           ON product(foil_treatment);
CREATE INDEX IF NOT EXISTS idx_product_treatment      ON product(card_treatment);
CREATE INDEX IF NOT EXISTS idx_product_rarity         ON product(rarity);
CREATE INDEX IF NOT EXISTS idx_product_language       ON product(language);
CREATE INDEX IF NOT EXISTS idx_product_color_identity ON product(color_identity);
