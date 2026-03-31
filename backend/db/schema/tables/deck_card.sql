-- Deck Card Table
CREATE TABLE IF NOT EXISTS deck_card (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id        UUID REFERENCES product(id) ON DELETE CASCADE,
  name              TEXT NOT NULL,
  set_code          TEXT,
  collector_number  TEXT,
  quantity          INTEGER NOT NULL DEFAULT 1,
  image_url         TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_deck_card_product ON deck_card(product_id);
