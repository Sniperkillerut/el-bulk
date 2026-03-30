CREATE TABLE IF NOT EXISTS bounty_offer (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bounty_id         UUID REFERENCES bounty(id) ON DELETE CASCADE,
  customer_name     TEXT NOT NULL,
  customer_contact  TEXT NOT NULL,
  condition         TEXT CHECK (condition IN ('NM', 'LP', 'MP', 'HP', 'DMG')),
  status            TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'accepted', 'rejected')),
  notes             TEXT,
  quantity          INTEGER NOT NULL DEFAULT 1,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Trigger for updated_at
CREATE TRIGGER trg_bounty_offer_updated_at
BEFORE UPDATE ON bounty_offer
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
