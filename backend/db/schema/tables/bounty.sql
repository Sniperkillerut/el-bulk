CREATE TABLE IF NOT EXISTS bounty (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name              TEXT NOT NULL,
  tcg               TEXT NOT NULL,
  set_name          TEXT,
  condition         TEXT CHECK (condition IN ('NM', 'LP', 'MP', 'HP', 'DMG') OR condition IS NULL),
  foil_treatment    TEXT NOT NULL DEFAULT 'non_foil',
  card_treatment    TEXT NOT NULL DEFAULT 'normal',
  collector_number  TEXT,
  promo_type        TEXT,
  language          TEXT NOT NULL DEFAULT 'en',
  target_price      NUMERIC(12, 2) CHECK (target_price >= 0),
  hide_price        BOOLEAN NOT NULL DEFAULT false,
  quantity_needed   INTEGER NOT NULL DEFAULT 1 CHECK (quantity_needed >= 0),
  image_url         TEXT,
  price_source      TEXT NOT NULL DEFAULT 'manual',
  price_reference   NUMERIC(12, 2),
  is_active         BOOLEAN NOT NULL DEFAULT true,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indices
CREATE INDEX IF NOT EXISTS idx_bounty_tcg ON bounty(tcg);
CREATE INDEX IF NOT EXISTS idx_bounty_search ON bounty USING gin(to_tsvector('english', name || ' ' || COALESCE(set_name, '')));
