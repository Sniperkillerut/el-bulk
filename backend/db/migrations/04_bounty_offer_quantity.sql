-- Add quantity to bounty_offer (Idempotent)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bounty_offer' AND column_name='quantity') THEN
        ALTER TABLE bounty_offer ADD COLUMN quantity INTEGER NOT NULL DEFAULT 1;
    END IF;
END $$;
