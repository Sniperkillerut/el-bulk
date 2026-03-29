ALTER TABLE bounty ADD COLUMN IF NOT EXISTS card_treatment TEXT NOT NULL DEFAULT 'normal';
ALTER TABLE bounty ADD COLUMN IF NOT EXISTS collector_number TEXT;
ALTER TABLE bounty ADD COLUMN IF NOT EXISTS promo_type TEXT;
ALTER TABLE bounty ADD COLUMN IF NOT EXISTS language TEXT NOT NULL DEFAULT 'en';

CREATE TABLE IF NOT EXISTS bounty_offer (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bounty_id         UUID REFERENCES bounty(id) ON DELETE CASCADE,
  customer_name     TEXT NOT NULL,
  customer_contact  TEXT NOT NULL,
  condition         TEXT CHECK (condition IN ('NM', 'LP', 'MP', 'HP', 'DMG') OR condition IS NULL),
  status            TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'accepted', 'rejected')),
  notes             TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
