-- Add email column to admin table
-- Added on 2026-04-13 to support OAuth login

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin' AND column_name='email') THEN
        ALTER TABLE admin ADD COLUMN email TEXT UNIQUE;
    END IF;
END $$;
