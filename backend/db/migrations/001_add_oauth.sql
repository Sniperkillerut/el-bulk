-- Remove the NOT NULL constraint from phone
ALTER TABLE customer ALTER COLUMN phone DROP NOT NULL;

-- Add new columns safely
ALTER TABLE customer ADD COLUMN IF NOT EXISTS auth_provider TEXT;
ALTER TABLE customer ADD COLUMN IF NOT EXISTS auth_provider_id TEXT;
ALTER TABLE customer ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- Add unique constraint on auth_provider_id if it doesn't already exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'customer_auth_provider_id_key'
    ) THEN
        ALTER TABLE customer ADD CONSTRAINT customer_auth_provider_id_key UNIQUE (auth_provider_id);
    END IF;
END $$;
