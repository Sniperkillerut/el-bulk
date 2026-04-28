DO $$
BEGIN
    BEGIN
        ALTER TABLE client_request ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;
    EXCEPTION
        WHEN duplicate_column THEN NULL;
    END;

    BEGIN
        ALTER TABLE client_request DROP CONSTRAINT IF EXISTS client_request_status_check;
    EXCEPTION
        WHEN undefined_object THEN NULL;
    END;

    ALTER TABLE client_request ADD CONSTRAINT client_request_status_check CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'rejected'::text, 'solved'::text, 'cancelled'::text, 'not_needed'::text]));
END $$;
