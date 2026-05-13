-- Client Request Table
CREATE TABLE IF NOT EXISTS client_request (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id       UUID REFERENCES customer(id) ON DELETE SET NULL,
  customer_name     TEXT NOT NULL,
  customer_contact  TEXT NOT NULL,
  card_name         TEXT NOT NULL,
  set_name          TEXT,
  details           TEXT,
  quantity          INTEGER DEFAULT 1,
  tcg               TEXT DEFAULT 'mtg',
  status            TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'solved', 'cancelled', 'not_needed')),
  
  -- Metadata for enrichment
  image_url         TEXT,
  foil_treatment    TEXT,
  card_treatment    TEXT,
  set_code          TEXT,
  collector_number  TEXT,
  
  -- Tracking
  bounty_id         UUID REFERENCES bounty(id) ON DELETE SET NULL,
  match_type        TEXT NOT NULL DEFAULT 'any' CHECK (match_type IN ('any', 'exact')),
  scryfall_id       TEXT,
  frame_effects     JSONB,
  oracle_id         UUID,
  is_prepared       BOOLEAN NOT NULL DEFAULT false,
  cancellation_reason TEXT,
  
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_client_request_status ON client_request(status);
CREATE INDEX IF NOT EXISTS idx_client_request_bounty_id ON client_request(bounty_id);
CREATE INDEX IF NOT EXISTS idx_client_request_oracle_id ON client_request(oracle_id);

-- Trigger to update updated_at
DROP TRIGGER IF EXISTS tr_client_request_updated_at ON client_request;
CREATE TRIGGER tr_client_request_updated_at
    BEFORE UPDATE ON client_request
    FOR EACH ROW
    EXECUTE FUNCTION fn_update_updated_at();
