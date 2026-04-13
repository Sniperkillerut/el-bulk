-- Add avatar_url to Admin table if missing
-- This is a separate migration because the previous one was already marked as applied.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin' AND column_name='avatar_url') THEN
        ALTER TABLE admin ADD COLUMN avatar_url TEXT;
    END IF;
END $$;
