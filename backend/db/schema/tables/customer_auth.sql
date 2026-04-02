-- Customer Auth Table (for multiple OAuth providers)
CREATE TABLE IF NOT EXISTS customer_auth (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id      UUID NOT NULL REFERENCES customer(id) ON DELETE CASCADE,
  provider         TEXT NOT NULL, -- 'google', 'facebook', etc.
  provider_id      TEXT NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, provider_id)
);
