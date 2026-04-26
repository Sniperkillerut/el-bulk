ALTER TABLE client_request ADD COLUMN cancellation_reason TEXT;
ALTER TABLE client_request DROP CONSTRAINT client_request_status_check;
ALTER TABLE client_request ADD CONSTRAINT client_request_status_check CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'rejected'::text, 'solved'::text, 'cancelled'::text, 'not_needed'::text]));
