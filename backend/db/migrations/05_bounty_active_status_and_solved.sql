-- Add is_active to bounty and update client_request status (Idempotent)
DO $$ 
BEGIN 
    -- 1. Add is_active to bounty
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bounty' AND column_name='is_active') THEN
        ALTER TABLE bounty ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
    END IF;

    -- 2. Update client_request status check constraint
    -- Note: We drop the old constraint and add the new one
    IF EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name = 'client_request' AND column_name = 'status' AND constraint_name = 'client_request_status_check') THEN
        ALTER TABLE client_request DROP CONSTRAINT client_request_status_check;
    END IF;
    
    ALTER TABLE client_request ADD CONSTRAINT client_request_status_check CHECK (status IN ('pending', 'accepted', 'rejected', 'solved'));
END $$;
