CREATE TABLE bounty (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name              TEXT NOT NULL,
  tcg               TEXT NOT NULL,
  set_name          TEXT,
  condition         TEXT CHECK (condition IN ('NM', 'LP', 'MP', 'HP', 'DMG') OR condition IS NULL),
  foil_treatment    TEXT NOT NULL DEFAULT 'non_foil',
  target_price      NUMERIC(12, 2) CHECK (target_price >= 0),
  hide_price        BOOLEAN NOT NULL DEFAULT false,
  quantity_needed   INTEGER NOT NULL DEFAULT 1 CHECK (quantity_needed >= 0),
  image_url         TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indices
CREATE INDEX idx_bounty_tcg ON bounty(tcg);
CREATE INDEX idx_bounty_search ON bounty USING gin(to_tsvector('english', name || ' ' || COALESCE(set_name, '')));
