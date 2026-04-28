-- Migration: Bounty Request Bridge Model
-- Replaces the old 1:1 bounty.request_id with a proper 1:N bounty_id on client_request
-- so many requests can link to one aggregate bounty.

-- 1. Add bounty_id, match_type, scryfall_id to client_request
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='client_request' AND column_name='bounty_id') THEN
        ALTER TABLE client_request ADD COLUMN bounty_id UUID REFERENCES bounty(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='client_request' AND column_name='match_type') THEN
        ALTER TABLE client_request ADD COLUMN match_type TEXT NOT NULL DEFAULT 'any' CHECK (match_type IN ('any', 'exact'));
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='client_request' AND column_name='scryfall_id') THEN
        ALTER TABLE client_request ADD COLUMN scryfall_id TEXT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_client_request_bounty_id ON client_request(bounty_id);

-- 2. Add is_generic flag to bounty (true = accepts any version of the card)
ALTER TABLE bounty ADD COLUMN IF NOT EXISTS is_generic BOOLEAN NOT NULL DEFAULT false;

-- 3. Remove old request_id from bounty (was a 1:1 reverse link, now replaced by bounty_id on client_request)
ALTER TABLE bounty DROP COLUMN IF EXISTS request_id;
