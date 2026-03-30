-- Bounty Offer Table
DROP TABLE IF EXISTS bounty_offer CASCADE;
CREATE TABLE IF NOT EXISTS bounty_offer (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bounty_id         UUID REFERENCES bounty(id) ON DELETE CASCADE NOT NULL,
  customer_id       UUID REFERENCES customer(id) ON DELETE CASCADE NOT NULL,
  quantity          INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
  condition         TEXT,
  status            TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'fulfilled')),
  notes             TEXT,
  admin_notes       TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indices
CREATE INDEX IF NOT EXISTS idx_bounty_offer_bounty ON bounty_offer(bounty_id);
CREATE INDEX IF NOT EXISTS idx_bounty_offer_customer ON bounty_offer(customer_id);
CREATE INDEX IF NOT EXISTS idx_bounty_offer_status ON bounty_offer(status);

-- Trigger to update updated_at
DROP TRIGGER IF EXISTS tr_bounty_offer_updated_at ON bounty_offer;
CREATE TRIGGER tr_bounty_offer_updated_at
    BEFORE UPDATE ON bounty_offer
    FOR EACH ROW
    EXECUTE FUNCTION fn_update_updated_at();
