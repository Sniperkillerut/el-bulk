-- Migration: Add oracle_id for robust aggregation
-- Enables perfect matching for generic requests by using Scryfall's unique oracle_id.

ALTER TABLE bounty ADD COLUMN IF NOT EXISTS oracle_id UUID;
ALTER TABLE client_request ADD COLUMN IF NOT EXISTS oracle_id UUID;

CREATE INDEX IF NOT EXISTS idx_bounty_oracle_id ON bounty(oracle_id);
CREATE INDEX IF NOT EXISTS idx_client_request_oracle_id ON client_request(oracle_id);
