CREATE TABLE IF NOT EXISTS client_request (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_name     TEXT NOT NULL,
  customer_contact  TEXT NOT NULL,
  card_name         TEXT NOT NULL,
  set_name          TEXT,
  details           TEXT,
  status            TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'solved')),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_client_request_status ON client_request(status);
