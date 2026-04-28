-- Add missing columns to bounty that were present in queries but missing from schema/migrations

DO $$ 
BEGIN
    ALTER TABLE bounty ADD COLUMN IF NOT EXISTS scryfall_id UUID;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

DO $$ 
BEGIN
    ALTER TABLE bounty ADD COLUMN IF NOT EXISTS set_code TEXT;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

DO $$ 
BEGIN
    ALTER TABLE bounty ADD COLUMN IF NOT EXISTS image_url TEXT;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

-- Also ensure oracle_id exists just in case the previous migration was skipped
DO $$ 
BEGIN
    ALTER TABLE bounty ADD COLUMN IF NOT EXISTS oracle_id UUID;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_bounty_scryfall_id ON bounty(scryfall_id);
