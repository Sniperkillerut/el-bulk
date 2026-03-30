-- Add price_source and price_reference to bounty table (Idempotent)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bounty' AND column_name='price_source') THEN
        ALTER TABLE bounty ADD COLUMN price_source TEXT NOT NULL DEFAULT 'manual';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bounty' AND column_name='price_reference') THEN
        ALTER TABLE bounty ADD COLUMN price_reference NUMERIC(12, 2);
    END IF;
END $$;
